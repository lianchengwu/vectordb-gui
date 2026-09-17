package client

import (
	"bufio"
	"context"
	"encoding/base64"
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

type dialerContextFunc func(ctx context.Context, network, addr string) (net.Conn, error)

type customProxyDialer struct {
	dialContext dialerContextFunc
}

func (d *customProxyDialer) Dial(network, addr string) (net.Conn, error) {
	return d.dialContext(context.Background(), network, addr)
}

func (d *customProxyDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d.dialContext(ctx, network, addr)
}

// BuildChainedHTTPClient builds an http.Client supporting arbitrary user-defined multi-hop proxy chains:
// e.g. Client ➔ [HTTP Proxy] ➔ [SOCKS5 Proxy] ➔ [SSH Jump Host] ➔ ... ➔ Target VectorDB.
func BuildChainedHTTPClient(timeout time.Duration, chain []storage.NetworkHop) (*http.Client, []*ssh.Client, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	baseDialer := &net.Dialer{
		Timeout:   15 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	currentDialer := baseDialer.DialContext
	var sshClients []*ssh.Client

	for i, hop := range chain {
		if !hop.Enabled || strings.TrimSpace(hop.Host) == "" {
			continue
		}

		port := hop.Port
		switch strings.ToLower(hop.Type) {
		case "socks5":
			if port <= 0 {
				port = 1080
			}
			socksAddr := net.JoinHostPort(strings.TrimSpace(hop.Host), strconv.Itoa(port))
			var auth *proxy.Auth
			if hop.Username != "" {
				auth = &proxy.Auth{User: hop.Username, Password: hop.Password}
			}

			underlyingDialer := currentDialer
			socksDialer, err := proxy.SOCKS5("tcp", socksAddr, auth, &customProxyDialer{dialContext: underlyingDialer})
			if err != nil {
				return nil, sshClients, fmt.Errorf("第 %d 节点 (SOCKS5 %s) 初始化失败: %w", i+1, socksAddr, err)
			}

			currentDialer = func(ctx context.Context, network, addr string) (net.Conn, error) {
				if cd, ok := socksDialer.(proxy.ContextDialer); ok {
					return cd.DialContext(ctx, network, addr)
				}
				return socksDialer.Dial(network, addr)
			}

		case "http", "https":
			if port <= 0 {
				port = 8080
			}
			proxyAddr := net.JoinHostPort(strings.TrimSpace(hop.Host), strconv.Itoa(port))
			prevDialer := currentDialer
			username := hop.Username
			password := hop.Password

			currentDialer = func(ctx context.Context, network, addr string) (net.Conn, error) {
				conn, err := prevDialer(ctx, "tcp", proxyAddr)
				if err != nil {
					return nil, fmt.Errorf("连接第 %d 节点 (HTTP 代理 %s) 失败: %w", i+1, proxyAddr, err)
				}

				req := &http.Request{
					Method: http.MethodConnect,
					URL:    &url.URL{Opaque: addr},
					Host:   addr,
					Header: make(http.Header),
				}
				if username != "" {
					auth := username + ":" + password
					basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
					req.Header.Set("Proxy-Authorization", basicAuth)
				}

				if err := req.Write(conn); err != nil {
					conn.Close()
					return nil, fmt.Errorf("向第 %d 节点 (HTTP 代理 %s) 发送握手失败: %w", i+1, proxyAddr, err)
				}

				resp, err := http.ReadResponse(bufio.NewReader(conn), req)
				if err != nil {
					conn.Close()
					return nil, fmt.Errorf("读取第 %d 节点 (HTTP 代理 %s) 响应失败: %w", i+1, proxyAddr, err)
				}
				if resp.StatusCode != http.StatusOK {
					conn.Close()
					return nil, fmt.Errorf("第 %d 节点 (HTTP 代理 %s) 拒绝连接目标 (%s): %s", i+1, proxyAddr, addr, resp.Status)
				}
				return conn, nil
			}

		case "ssh":
			if port <= 0 {
				port = 22
			}
			user := strings.TrimSpace(hop.Username)
			if user == "" {
				user = "root"
			}

			var authMethods []ssh.AuthMethod
			if hop.AuthType == "key" && strings.TrimSpace(hop.PrivateKey) != "" {
				keyBytes := []byte(hop.PrivateKey)
				if _, err := os.Stat(strings.TrimSpace(hop.PrivateKey)); err == nil {
					if fileData, err := os.ReadFile(strings.TrimSpace(hop.PrivateKey)); err == nil {
						keyBytes = fileData
					}
				}

				var signer ssh.Signer
				var err error
				if hop.Passphrase != "" {
					signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(hop.Passphrase))
				} else {
					signer, err = ssh.ParsePrivateKey(keyBytes)
				}
				if err != nil {
					return nil, sshClients, fmt.Errorf("第 %d 节点 (SSH) 私钥解析失败: %w", i+1, err)
				}
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			} else if hop.Password != "" {
				authMethods = append(authMethods, ssh.Password(hop.Password))
			}

			clientConfig := &ssh.ClientConfig{
				User:            user,
				Auth:            authMethods,
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				Timeout:         15 * time.Second,
			}

			sshHostPort := net.JoinHostPort(strings.TrimSpace(hop.Host), strconv.Itoa(port))

			dialCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			netConn, err := currentDialer(dialCtx, "tcp", sshHostPort)
			cancel()
			if err != nil {
				return nil, sshClients, fmt.Errorf("连接第 %d 节点 (SSH 跳板机 %s) 失败: %w", i+1, sshHostPort, err)
			}

			clientConn, chans, reqs, err := ssh.NewClientConn(netConn, sshHostPort, clientConfig)
			if err != nil {
				netConn.Close()
				return nil, sshClients, fmt.Errorf("第 %d 节点 (SSH %s) 握手鉴权失败: %w", i+1, sshHostPort, err)
			}

			client := ssh.NewClient(clientConn, chans, reqs)
			sshClients = append(sshClients, client)

			currentDialer = func(ctx context.Context, network, addr string) (net.Conn, error) {
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

		default:
			return nil, sshClients, fmt.Errorf("不支持的网络节点类型: %s", hop.Type)
		}
	}

	transport.DialContext = currentDialer

	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	return httpClient, sshClients, nil
}

// BuildHTTPClient maintains backward compatibility with legacy dual-field config.
func BuildHTTPClient(timeout time.Duration, proxyCfg storage.ProxyConfig, sshCfg storage.SSHTunnelConfig) (*http.Client, *ssh.Client, error) {
	var chain []storage.NetworkHop
	if proxyCfg.Enabled && strings.TrimSpace(proxyCfg.Host) != "" {
		chain = append(chain, storage.NetworkHop{
			Enabled:  true,
			Type:     proxyCfg.Type,
			Host:     proxyCfg.Host,
			Port:     proxyCfg.Port,
			Username: proxyCfg.Username,
			Password: proxyCfg.Password,
		})
	}
	if sshCfg.Enabled && strings.TrimSpace(sshCfg.Host) != "" {
		chain = append(chain, storage.NetworkHop{
			Enabled:    true,
			Type:       "ssh",
			Host:       sshCfg.Host,
			Port:       sshCfg.Port,
			Username:   sshCfg.User,
			AuthType:   sshCfg.AuthType,
			Password:   sshCfg.Password,
			PrivateKey: sshCfg.PrivateKey,
			Passphrase: sshCfg.Passphrase,
		})
	}

	httpClient, sshClients, err := BuildChainedHTTPClient(timeout, chain)
	var singleSSH *ssh.Client
	if len(sshClients) > 0 {
		singleSSH = sshClients[0]
	}
	return httpClient, singleSSH, err
}
