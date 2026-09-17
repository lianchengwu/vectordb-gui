package storage

import (
	"os"
	"path/filepath"
	"runtime"
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

func TestStorageCorruptedFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tcvectordb-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "connections.json")
	if err := os.WriteFile(filePath, []byte("{corrupted json"), 0600); err != nil {
		t.Fatalf("failed to write corrupted file: %v", err)
	}

	store := NewStorageWithPath(filePath)
	cfg := ConnectionConfig{
		ID:   "conn-1",
		Name: "Test",
	}

	// Save should return error if file is corrupted
	if err := store.Save(cfg); err == nil {
		t.Fatalf("expected error saving to corrupted storage, got nil")
	}

	// List should return error if file is corrupted
	if _, err := store.List(); err == nil {
		t.Fatalf("expected error listing corrupted storage, got nil")
	}

	// Delete should return error if file is corrupted
	if err := store.Delete("conn-1"); err == nil {
		t.Fatalf("expected error deleting from corrupted storage, got nil")
	}
}

func TestStorageFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping POSIX file permission checks on Windows")
	}
	tempDir, err := os.MkdirTemp("", "tcvectordb-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	subDir := filepath.Join(tempDir, "subdir")
	filePath := filepath.Join(subDir, "connections.json")
	store := NewStorageWithPath(filePath)

	cfg := ConnectionConfig{
		ID:   "conn-1",
		Name: "Test",
	}
	if err := store.Save(cfg); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("expected file permission 0600, got %o", perm)
	}

	dirInfo, err := os.Stat(subDir)
	if err != nil {
		t.Fatalf("failed to stat dir: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0700 {
		t.Fatalf("expected dir permission 0700, got %o", perm)
	}
}
