package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type ConnectionConfig struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Username string `json:"username"`
	APIKey   string `json:"apiKey"`
	Timeout  int    `json:"timeout"` // in seconds
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
	_ = os.MkdirAll(dir, 0755)
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

func (s *Storage) Save(config ConnectionConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conns, err := s.listUnlocked()
	if err != nil {
		conns = []ConnectionConfig{}
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
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(conns, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}
