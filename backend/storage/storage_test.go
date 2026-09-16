package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConnectionStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tcvectordb-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "connections.json")
	store := NewStorageWithPath(filePath)

	// Test Empty List
	conns, err := store.List()
	if err != nil {
		t.Fatalf("unexpected error listing empty storage: %v", err)
	}
	if len(conns) != 0 {
		t.Fatalf("expected 0 connections, got %d", len(conns))
	}

	// Test Save
	cfg := ConnectionConfig{
		ID:       "conn-1",
		Name:     "Test Cluster",
		URL:      "http://10.0.0.1:80",
		Username: "root",
		APIKey:   "secret-key",
		Timeout:  10,
	}
	err = store.Save(cfg)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Test Retrieve
	conns, err = store.List()
	if err != nil {
		t.Fatalf("failed to list configs: %v", err)
	}
	// Test Update
	cfg.Name = "Updated Cluster"
	err = store.Save(cfg)
	if err != nil {
		t.Fatalf("failed to update config: %v", err)
	}
	conns, err = store.List()
	if err != nil || len(conns) != 1 || conns[0].Name != "Updated Cluster" {
		t.Fatalf("expected updated name, got %+v", conns)
	}

	if len(conns) != 1 || conns[0].ID != "conn-1" {
		t.Fatalf("expected conn-1, got %+v", conns)
	}

	// Test Delete
	err = store.Delete("conn-1")
	if err != nil {
		t.Fatalf("failed to delete config: %v", err)
	}
	conns, _ = store.List()
	if len(conns) != 0 {
		t.Fatalf("expected 0 connections after delete, got %d", len(conns))
	}
}

func TestDefaultStorage(t *testing.T) {
	store := DefaultStorage()
	if store == nil || store.filePath == "" {
		t.Fatalf("expected non-nil default storage with filePath")
	}
}
