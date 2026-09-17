package service

import (
	"context"
	"fmt"
	"time"
	"vectordb-1/backend/client"
	"vectordb-1/backend/storage"
)

type ConnectionService struct {
	storage *storage.Storage
}

func NewConnectionService() *ConnectionService {
	return &ConnectionService{
		storage: storage.DefaultStorage(),
	}
}

func NewConnectionServiceWithStorage(s *storage.Storage) *ConnectionService {
	return &ConnectionService{storage: s}
}

func (s *ConnectionService) ListConnections() ([]storage.ConnectionConfig, error) {
	return s.storage.List()
}

func (s *ConnectionService) SaveConnection(cfg storage.ConnectionConfig) error {
	if cfg.ID == "" {
		cfg.ID = fmt.Sprintf("conn_%d", time.Now().UnixNano())
	}
	if cfg.Username == "" {
		cfg.Username = "root"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10
	}
	return s.storage.Save(cfg)
}

func (s *ConnectionService) DeleteConnection(id string) error {
	return s.storage.Delete(id)
}

type TestConnectionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (s *ConnectionService) TestConnection(cfg storage.ConnectionConfig) (*TestConnectionResult, error) {
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	cli, err := client.NewClientWithConfig(cfg, timeout)
	if err != nil {
		return &TestConnectionResult{Success: false, Message: err.Error()}, nil
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ok, msg, err := cli.Ping(ctx)
	if err != nil {
		return &TestConnectionResult{Success: false, Message: err.Error()}, nil
	}
	return &TestConnectionResult{Success: ok, Message: msg}, nil
}

func (s *ConnectionService) FetchDatabases(cfg storage.ConnectionConfig) ([]client.DatabaseDetail, error) {
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	cli, err := client.NewClientWithConfig(cfg, timeout)
	if err != nil {
		return nil, fmt.Errorf("创建客户端连接失败: %w", err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return cli.ListDatabasesDetailed(ctx)
}
