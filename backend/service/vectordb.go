package service

import (
	"context"
	"fmt"
	"reflect"
	"strings"
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

func (s *VectorDBService) ListDatabasesDetailed(connID string) ([]client.DatabaseDetail, error) {
	cli, err := s.getClient(connID)
	if err != nil {
		return nil, err
	}

	var configuredDbs []string
	if s.storage != nil {
		if cfg, _ := s.storage.Get(connID); cfg != nil {
			for _, d := range cfg.Databases {
				trimmed := strings.TrimSpace(d)
				if trimmed != "" {
					configuredDbs = append(configuredDbs, trimmed)
				}
			}
		}
	}

	timeout := s.getTimeout(connID, 15*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	allDbs, err := cli.ListDatabasesDetailed(ctx)
	if err != nil {
		// If cluster-wide list fails (e.g. sub-account without list permission)
		// but specific visible databases are configured, return configured databases
		if len(configuredDbs) > 0 {
			fallback := make([]client.DatabaseDetail, 0, len(configuredDbs))
			for _, name := range configuredDbs {
				fallback = append(fallback, client.DatabaseDetail{
					Name:   name,
					DbType: "base",
				})
			}
			return fallback, nil
		}
		return nil, err
	}

	if len(configuredDbs) == 0 {
		return allDbs, nil
	}

	dbMap := make(map[string]client.DatabaseDetail)
	for _, d := range allDbs {
		dbMap[d.Name] = d
	}

	var filtered []client.DatabaseDetail
	for _, name := range configuredDbs {
		if detail, ok := dbMap[name]; ok {
			filtered = append(filtered, detail)
		} else {
			filtered = append(filtered, client.DatabaseDetail{
				Name:   name,
				DbType: "base",
			})
		}
	}
	return filtered, nil
}

func (s *VectorDBService) ListDatabases(connID string) ([]string, error) {
	details, err := s.ListDatabasesDetailed(connID)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(details))
	for _, d := range details {
		names = append(names, d.Name)
	}
	return names, nil
}

func (s *VectorDBService) ListCollections(connID, database, dbType string) ([]string, error) {
	cli, err := s.getClient(connID)
	if err != nil {
		return nil, err
	}
	timeout := s.getTimeout(connID, 15*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return cli.ListCollections(ctx, database, dbType)
}

func (s *VectorDBService) DescribeCollection(connID, database, collection, dbType string) (*client.CollectionMeta, error) {
	cli, err := s.getClient(connID)
	if err != nil {
		return nil, err
	}
	timeout := s.getTimeout(connID, 15*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return cli.DescribeCollection(ctx, database, collection, dbType)
}

func (s *VectorDBService) DropCollection(connID, database, collection, dbType string) error {
	cli, err := s.getClient(connID)
	if err != nil {
		return err
	}
	timeout := s.getTimeout(connID, 15*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return cli.DropCollection(ctx, database, collection, dbType)
}

type QueryDocumentsParams struct {
	ConnectionID string `json:"connectionId"`
	Database     string `json:"database"`
	Collection   string `json:"collection"`
	DbType       string `json:"dbType,omitempty"`
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
	return cli.QueryDocuments(ctx, params.Database, params.Collection, params.DbType, params.Limit, params.Offset, params.Filter)
}

type UpdateDocumentParams struct {
	ConnectionID string                     `json:"connectionId"`
	Database     string                     `json:"database"`
	Collection   string                     `json:"collection"`
	DbType       string                     `json:"dbType,omitempty"`
	Query        client.UpdateDocumentQuery `json:"query"`
	Update       map[string]interface{}     `json:"update"`
}

func (s *VectorDBService) UpdateDocument(params UpdateDocumentParams) error {
	cli, err := s.getClient(params.ConnectionID)
	if err != nil {
		return err
	}
	timeout := s.getTimeout(params.ConnectionID, 30*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return cli.UpdateDocument(ctx, params.Database, params.Collection, params.DbType, params.Query, params.Update)
}
