package service

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
	"vectordb-1/backend/client"
	"vectordb-1/backend/storage"
)

type cachedClient struct {
	client *client.Client
	config storage.ConnectionConfig
}

type VectorDBService struct {
	storage *storage.Storage
	mu      sync.RWMutex
	clients map[string]cachedClient
}

func NewVectorDBService() *VectorDBService {
	return NewVectorDBServiceWithStorage(storage.DefaultStorage())
}

func NewVectorDBServiceWithStorage(s *storage.Storage) *VectorDBService {
	return &VectorDBService{
		storage: s,
		clients: make(map[string]cachedClient),
	}
}

// InvalidateCache removes the cached client for a given connection ID,
// or purges all cached clients if connID is empty.
func (s *VectorDBService) InvalidateCache(connID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if connID == "" {
		for _, entry := range s.clients {
			if entry.client != nil {
				_ = entry.client.Close()
			}
		}
		s.clients = make(map[string]cachedClient)
		return
	}
	if entry, ok := s.clients[connID]; ok {
		if entry.client != nil {
			_ = entry.client.Close()
		}
		delete(s.clients, connID)
	}
}

func (s *VectorDBService) getClient(connID string) (*client.Client, error) {
	conns, err := s.storage.List()
	if err != nil {
		return nil, fmt.Errorf("failed to load connections: %w", err)
	}

	var target *storage.ConnectionConfig
	for _, c := range conns {
		if c.ID == connID {
			target = &c
			break
		}
	}
	if target == nil {
		s.mu.Lock()
		delete(s.clients, connID)
		s.mu.Unlock()
		return nil, fmt.Errorf("connection config %s not found", connID)
	}

	s.mu.RLock()
	entry, ok := s.clients[connID]
	s.mu.RUnlock()

	if ok && reflect.DeepEqual(entry.config, *target) {
		return entry.client, nil
	}

	clientTimeout := 60 * time.Second
	if target.Timeout > 60 {
		clientTimeout = time.Duration(target.Timeout) * time.Second
	}
	newCli, err := client.NewClientWithConfig(*target, clientTimeout)
	if err != nil {
		return nil, fmt.Errorf("创建客户端连接失败: %w", err)
	}
	s.mu.Lock()
	s.clients[connID] = cachedClient{
		client: newCli,
		config: *target,
	}
	s.mu.Unlock()

	return newCli, nil
}

func (s *VectorDBService) getTimeout(connID string, defaultDuration time.Duration) time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if entry, ok := s.clients[connID]; ok && entry.config.Timeout > 0 {
		dur := time.Duration(entry.config.Timeout) * time.Second
		if dur > defaultDuration {
			return dur
		}
	}
	return defaultDuration
}

func (s *VectorDBService) ListDatabases(connID string) ([]string, error) {
	cli, err := s.getClient(connID)
	if err != nil {
		return nil, err
	}
	timeout := s.getTimeout(connID, 15*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return cli.ListDatabases(ctx)
}

func (s *VectorDBService) ListCollections(connID, database string) ([]string, error) {
	cli, err := s.getClient(connID)
	if err != nil {
		return nil, err
	}
	timeout := s.getTimeout(connID, 15*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return cli.ListCollections(ctx, database)
}

func (s *VectorDBService) DescribeCollection(connID, database, collection string) (*client.CollectionMeta, error) {
	cli, err := s.getClient(connID)
	if err != nil {
		return nil, err
	}
	timeout := s.getTimeout(connID, 15*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return cli.DescribeCollection(ctx, database, collection)
}

type QueryDocumentsParams struct {
	ConnectionID string `json:"connectionId"`
	Database     string `json:"database"`
	Collection   string `json:"collection"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
	Filter       string `json:"filter"`
}

func (s *VectorDBService) QueryDocuments(params QueryDocumentsParams) (*client.QueryDocumentResponse, error) {
	cli, err := s.getClient(params.ConnectionID)
	if err != nil {
		return nil, err
	}
	timeout := s.getTimeout(params.ConnectionID, 30*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return cli.QueryDocuments(ctx, params.Database, params.Collection, params.Limit, params.Offset, params.Filter)
}
