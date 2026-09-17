package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"vectordb-1/backend/client"
	"vectordb-1/backend/storage"
)

func setupMockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/database/list":
			resp := client.ListDatabasesResponse{
				ResponseHeader: client.ResponseHeader{Code: 0, Message: "Success"},
				Databases: []client.DatabaseEntry{
					{Name: "db_test1"},
					{Name: "db_test2"},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "/collection/list":
			resp := client.ListCollectionsResponse{
				ResponseHeader: client.ResponseHeader{Code: 0, Message: "Success"},
				Collections: []client.CollectionEntry{
					{Collection: "coll_1"},
					{Collection: "coll_2"},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "/collection/describe":
			resp := client.DescribeCollectionResponse{
				ResponseHeader: client.ResponseHeader{Code: 0, Message: "Success"},
				Collection: client.CollectionMeta{
					Database:    "db_test1",
					Collection:  "coll_1",
					ReplicaNum:  1,
					ShardNum:    2,
					Description: "Test collection",
					Fields: []client.FieldMeta{
						{FieldName: "id", FieldType: "string", PrimaryKey: true},
						{FieldName: "vector", FieldType: "vector", FieldUsage: "vector"},
					},
					Indexes: []client.IndexColumn{
						{FieldName: "vector", IndexType: "HNSW", MetricType: "COSINE"},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "/document/query":
			resp := client.QueryDocumentResponse{
				ResponseHeader: client.ResponseHeader{Code: 0, Message: "Success"},
				Count:          1,
				Documents: []map[string]interface{}{
					{"id": "doc_1", "text": "sample document", "vector": []interface{}{0.1, 0.2}},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestConnectionService_CRUD(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "connservice-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := storage.NewStorageWithPath(filepath.Join(tempDir, "conns.json"))
	svc := NewConnectionServiceWithStorage(store)

	// 1. Initially empty
	list, err := svc.ListConnections()
	if err != nil || len(list) != 0 {
		t.Fatalf("expected empty list, got %v, err: %v", list, err)
	}

	// 2. Save with explicit ID
	cfg := storage.ConnectionConfig{
		ID:       "conn-1",
		Name:     "Local VDB",
		URL:      "http://localhost:80",
		Username: "root",
		APIKey:   "secret-key",
		Timeout:  15,
	}
	if err := svc.SaveConnection(cfg); err != nil {
		t.Fatalf("SaveConnection failed: %v", err)
	}

	list, err = svc.ListConnections()
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 connection, got %d, err: %v", len(list), err)
	}
	if list[0].Name != "Local VDB" {
		t.Fatalf("expected name 'Local VDB', got '%s'", list[0].Name)
	}

	// 3. Save with empty ID and defaults
	cfgNoID := storage.ConnectionConfig{
		Name: "Remote VDB",
		URL:  "http://10.0.0.1:80",
	}
	if err := svc.SaveConnection(cfgNoID); err != nil {
		t.Fatalf("SaveConnection with empty ID failed: %v", err)
	}

	list, err = svc.ListConnections()
	if err != nil || len(list) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(list))
	}
	for _, c := range list {
		if c.Name == "Remote VDB" {
			if c.ID == "" {
				t.Fatalf("expected auto-generated ID for Remote VDB")
			}
			if c.Username != "root" {
				t.Fatalf("expected default username 'root', got '%s'", c.Username)
			}
			if c.Timeout != 10 {
				t.Fatalf("expected default timeout 10, got %d", c.Timeout)
			}
		}
	}

	// 4. Delete
	if err := svc.DeleteConnection("conn-1"); err != nil {
		t.Fatalf("DeleteConnection failed: %v", err)
	}
	list, err = svc.ListConnections()
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 connection after delete, got %d", len(list))
	}
}

func TestConnectionService_TestConnection(t *testing.T) {
	server := setupMockServer(t)
	defer server.Close()

	svc := NewConnectionService()

	// Test success
	res, err := svc.TestConnection(storage.ConnectionConfig{
		URL:      server.URL,
		Username: "root",
		APIKey:   "valid-key",
		Timeout:  5,
	})
	if err != nil {
		t.Fatalf("TestConnection returned unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success true, got false with message: %s", res.Message)
	}

	// Test failure with non-existent server
	resFail, err := svc.TestConnection(storage.ConnectionConfig{
		URL:      "http://127.0.0.1:54321", // unreachable port
		Username: "root",
		APIKey:   "key",
		Timeout:  1,
	})
	if err != nil {
		t.Fatalf("expected error handled in TestConnectionResult, got err: %v", err)
	}
	if resFail.Success {
		t.Fatalf("expected success false for unreachable server")
	}
	if resFail.Message == "" {
		t.Fatalf("expected non-empty error message on failure")
	}
}

func TestVectorDBService_Operations(t *testing.T) {
	server := setupMockServer(t)
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "vdb-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := storage.NewStorageWithPath(filepath.Join(tempDir, "conns.json"))
	cfg := storage.ConnectionConfig{
		ID:       "test-conn",
		Name:     "Mock VDB",
		URL:      server.URL,
		Username: "root",
		APIKey:   "mock-key",
		Timeout:  5,
	}
	if err := store.Save(cfg); err != nil {
		t.Fatalf("store.Save failed: %v", err)
	}

	vdbSvc := NewVectorDBServiceWithStorage(store)

	// 1. ListDatabases
	dbs, err := vdbSvc.ListDatabases("test-conn")
	if err != nil {
		t.Fatalf("ListDatabases failed: %v", err)
	}
	if len(dbs) != 2 || dbs[0] != "db_test1" || dbs[1] != "db_test2" {
		t.Fatalf("unexpected databases: %v", dbs)
	}

	// 2. ListCollections
	colls, err := vdbSvc.ListCollections("test-conn", "db_test1", "base")
	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}
	if len(colls) != 2 || colls[0] != "coll_1" || colls[1] != "coll_2" {
		t.Fatalf("unexpected collections: %v", colls)
	}

	// 3. DescribeCollection
	meta, err := vdbSvc.DescribeCollection("test-conn", "db_test1", "coll_1", "base")
	if err != nil {
		t.Fatalf("DescribeCollection failed: %v", err)
	}
	if meta.Collection != "coll_1" || meta.Database != "db_test1" || len(meta.Fields) != 2 {
		t.Fatalf("unexpected meta: %+v", meta)
	}

	// 4. QueryDocuments
	res, err := vdbSvc.QueryDocuments(QueryDocumentsParams{
		ConnectionID: "test-conn",
		Database:     "db_test1",
		Collection:   "coll_1",
		Limit:        10,
		Offset:       0,
	})
	if err != nil {
		t.Fatalf("QueryDocuments failed: %v", err)
	}
	if res.Count != 1 || len(res.Documents) != 1 {
		t.Fatalf("unexpected query result: %+v", res)
	}

	// 5. Non-existent connection ID
	_, err = vdbSvc.ListDatabases("unknown-conn")
	if err == nil {
		t.Fatalf("expected error for unknown connection ID")
	}
}

func TestVectorDBService_ClientCachingAndInvalidation(t *testing.T) {
	server1 := setupMockServer(t)
	defer server1.Close()

	var server2Hit bool
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/database/list" {
			server2Hit = true
			resp := client.ListDatabasesResponse{
				ResponseHeader: client.ResponseHeader{Code: 0, Message: "Success"},
				Databases: []client.DatabaseEntry{
					{Name: "db_from_server2"},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server2.Close()

	tempDir, err := os.MkdirTemp("", "vdb-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := storage.NewStorageWithPath(filepath.Join(tempDir, "conns.json"))
	cfg := storage.ConnectionConfig{
		ID:       "c1",
		Name:     "Server 1",
		URL:      server1.URL,
		Username: "root",
		APIKey:   "key1",
		Timeout:  5,
	}
	_ = store.Save(cfg)

	vdbSvc := NewVectorDBServiceWithStorage(store)

	// First call populates cache
	dbs, err := vdbSvc.ListDatabases("c1")
	if err != nil {
		t.Fatalf("ListDatabases failed: %v", err)
	}
	if len(dbs) != 2 || dbs[0] != "db_test1" {
		t.Fatalf("unexpected dbs from server1: %v", dbs)
	}

	// Update config to point to server2
	cfg.URL = server2.URL
	_ = store.Save(cfg)

	// Second call should detect config change and use new client pointing to server2
	dbs2, err := vdbSvc.ListDatabases("c1")
	if err != nil {
		t.Fatalf("ListDatabases after update failed: %v", err)
	}
	if !server2Hit {
		t.Fatalf("expected server 2 to be queried after config change")
	}
	if len(dbs2) != 1 || dbs2[0] != "db_from_server2" {
		t.Fatalf("expected dbs from server 2, got %v", dbs2)
	}

	// Invalidate cache explicitly
	vdbSvc.InvalidateCache("c1")
	vdbSvc.InvalidateCache("") // clear all

	// Delete connection
	_ = store.Delete("c1")
	_, err = vdbSvc.ListDatabases("c1")
	if err == nil {
		t.Fatalf("expected error after connection deleted from store")
	}
}

func TestVectorDBService_Concurrency(t *testing.T) {
	server := setupMockServer(t)
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "vdb-race-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := storage.NewStorageWithPath(filepath.Join(tempDir, "conns.json"))
	cfg := storage.ConnectionConfig{
		ID:       "race-conn",
		Name:     "Race VDB",
		URL:      server.URL,
		Username: "root",
		APIKey:   "k",
		Timeout:  5,
	}
	_ = store.Save(cfg)

	vdbSvc := NewVectorDBServiceWithStorage(store)

	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%3 == 0 {
				_, _ = vdbSvc.ListDatabases("race-conn")
			} else if idx%3 == 1 {
				_, _ = vdbSvc.ListCollections("race-conn", "db_test1", "base")
			} else {
				vdbSvc.InvalidateCache("race-conn")
			}
		}(i)
	}
	wg.Wait()
}
