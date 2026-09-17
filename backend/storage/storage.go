package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type ProxyConfig struct {
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"` // "http", "socks5"
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type SSHTunnelConfig struct {
	Enabled    bool   `json:"enabled"`
	Host       string `json:"host"`
	Port       int    `json:"port"` // default 22
	User       string `json:"user"` // e.g. root
	AuthType   string `json:"authType"` // "password", "key"
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"` // file path or PEM key text
	Passphrase string `json:"passphrase,omitempty"`
}

type NetworkHop struct {
	ID         string `json:"id"`
	Enabled    bool   `json:"enabled"`
	Type       string `json:"type"` // "http", "socks5", "ssh"
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username,omitempty"` // proxy username or ssh user
	Password   string `json:"password,omitempty"`
	AuthType   string `json:"authType,omitempty"` // for ssh: "password", "key"
	PrivateKey string `json:"privateKey,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
}

type ConnectionConfig struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	URL        string          `json:"url"`
	Username   string          `json:"username"`
	APIKey     string          `json:"apiKey"`
	Timeout    int             `json:"timeout"` // in seconds
	Databases  []string        `json:"databases,omitempty"`
	ProxyChain []NetworkHop    `json:"proxyChain,omitempty"`
	Proxy      ProxyConfig     `json:"proxy,omitempty"`
	SSHTunnel  SSHTunnelConfig `json:"sshTunnel,omitempty"`
}

type Storage struct {
	filePath string
	mu       sync.RWMutex
}

func DefaultStorage() *Storage {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".tcvectordb")
	_ = os.MkdirAll(dir, 0700)
	return &Storage{
		filePath: filepath.Join(dir, "connections.json"),
	}
}

func NewStorageWithPath(path string) *Storage {
	return &Storage{
		filePath: path,
	}
}

func (s *Storage) List() ([]ConnectionConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.listUnlocked()
}

func (s *Storage) Get(id string) (*ConnectionConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conns, err := s.listUnlocked()
	if err != nil {
		return nil, err
	}
	for _, c := range conns {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil
}

func (s *Storage) Save(config ConnectionConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conns, err := s.listUnlocked()
	if err != nil {
		return err
	}

	found := false
	for i, c := range conns {
		if c.ID == config.ID {
			conns[i] = config
			found = true
			break
		}
	}
	if !found {
		conns = append(conns, config)
	}

	return s.saveUnlocked(conns)
}

func (s *Storage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conns, err := s.listUnlocked()
	if err != nil {
		return err
	}

	filtered := make([]ConnectionConfig, 0, len(conns))
	for _, c := range conns {
		if c.ID != id {
			filtered = append(filtered, c)
		}
	}

	return s.saveUnlocked(filtered)
}

func (s *Storage) listUnlocked() ([]ConnectionConfig, error) {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return []ConnectionConfig{}, nil
	}
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return []ConnectionConfig{}, nil
	}
	var conns []ConnectionConfig
	if err := json.Unmarshal(data, &conns); err != nil {
		return nil, err
	}
	return conns, nil
}

func (s *Storage) saveUnlocked(conns []ConnectionConfig) error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(conns, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0600)
}
