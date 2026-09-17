package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
	"vectordb-1/backend/storage"
)

// BuildHTTPClient creates an http.Client configured with optional HTTP/SOCKS5 proxy or SSH tunnel.
// If an SSH tunnel is established, it returns the *ssh.Client which the caller must close when done.
func BuildHTTPClient(timeout time.Duration, proxyCfg storage.ProxyConfig, sshCfg storage.SSHTunnelConfig) (*http.Client, *ssh.Client, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	var sshClient *ssh.Client

	// 1. SSH Tunnel takes precedence if enabled
	if sshCfg.Enabled && sshCfg.Host != "" {
		port := sshCfg.Port
		if port <= 0 {
			port = 22
		}
		user := sshCfg.User
		if user == "" {
			user = "root"
		}

		var authMethods []ssh.AuthMethod
		if sshCfg.AuthType == "key" && sshCfg.PrivateKey != "" {
			keyBytes := []byte(sshCfg.PrivateKey)
			// Check if PrivateKey is a file path
			if _, err := os.Stat(strings.TrimSpace(sshCfg.PrivateKey)); err == nil {
				if fileData, err := os.ReadFile(strings.TrimSpace(sshCfg.PrivateKey)); err == nil {
					keyBytes = fileData
				}
			}

			var signer ssh.Signer
			var err error
			if sshCfg.Passphrase != "" {
				signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(sshCfg.Passphrase))
			} else {
				signer, err = ssh.ParsePrivateKey(keyBytes)
			}
			if err != nil {
				return nil, nil, fmt.Errorf("解析 SSH 私钥失败: %w", err)
			}
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		} else if sshCfg.Password != "" {
			authMethods = append(authMethods, ssh.Password(sshCfg.Password))
		}

		clientConfig := &ssh.ClientConfig{
			User:            user,
			Auth:            authMethods,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         15 * time.Second,
		}

		sshAddr := net.JoinHostPort(sshCfg.Host, strconv.Itoa(port))
		client, err := ssh.Dial("tcp", sshAddr, clientConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("连接 SSH 跳板机 (%s) 失败: %w", sshAddr, err)
		}
		sshClient = client

		// Route all HTTP requests through the SSH tunnel
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			type dialResult struct {
				conn net.Conn
				err  error
			}
			done := make(chan dialResult, 1)

			go func() {
				conn, err := client.Dial(network, addr)
				done <- dialResult{conn: conn, err: err}
			}()

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case res := <-done:
				return res.conn, res.err
			}
		}

	} else if proxyCfg.Enabled && proxyCfg.Host != "" {
		port := proxyCfg.Port
		if port <= 0 {
			if proxyCfg.Type == "socks5" {
				port = 1080
			} else {
				port = 8080
			}
		}

		switch strings.ToLower(proxyCfg.Type) {
		case "socks5":
			var auth *proxy.Auth
			if proxyCfg.Username != "" {
				auth = &proxy.Auth{
					User:     proxyCfg.Username,
					Password: proxyCfg.Password,
				}
			}
			socksAddr := net.JoinHostPort(proxyCfg.Host, strconv.Itoa(port))
			dialer, err := proxy.SOCKS5("tcp", socksAddr, auth, proxy.Direct)
			if err != nil {
				return nil, nil, fmt.Errorf("创建 SOCKS5 代理客户端失败: %w", err)
			}

			if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
				transport.DialContext = contextDialer.DialContext
			} else {
				transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				}
			}

		case "http", "https":
			proxyURL := &url.URL{
				Scheme: strings.ToLower(proxyCfg.Type),
				Host:   net.JoinHostPort(proxyCfg.Host, strconv.Itoa(port)),
			}
			if proxyCfg.Username != "" {
				proxyURL.User = url.UserPassword(proxyCfg.Username, proxyCfg.Password)
			}
			transport.Proxy = http.ProxyURL(proxyURL)

		default:
			return nil, nil, fmt.Errorf("不支持的代理类型: %s", proxyCfg.Type)
		}
	} else {
		// Default standard direct dialer
		transport.Proxy = http.ProxyFromEnvironment
		transport.DialContext = (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext
	}

	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	return httpClient, sshClient, nil
}
