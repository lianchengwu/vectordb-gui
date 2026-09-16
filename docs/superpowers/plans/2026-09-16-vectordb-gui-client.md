# Tencent Cloud VectorDB GUI Client Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Develop a cross-platform desktop GUI client using Wails v3, Go, and Vue 3 to connect to Tencent Cloud VectorDB, view database collections, inspect schemas, and browse documents/vectors with pagination and filtering.

**Architecture:** A layered desktop architecture with Wails v3 as application shell. The Go backend directly interacts with Tencent Cloud VectorDB REST API using Go standard library `net/http`, stores connection configs locally in JSON, and exposes RPC services (`ConnectionService`, `VectorDBService`) to the Vue 3 + Tailwind CSS frontend.

**Tech Stack:** Wails v3 (`v3.0.0-beta.22`), Go 1.27, Vue 3, Vite, Tailwind CSS, TypeScript, Lucide Icons.

---

### File Structure Map

```
vectordb-1/
├── Taskfile.yml                        # Wails build and dev task runner
├── main.go                             # Wails v3 app entrypoint & service registration
├── go.mod                              # Go module definition
├── backend/
│   ├── storage/
│   │   ├── storage.go                  # Local JSON connection config persistence
│   │   └── storage_test.go             # Unit tests for storage CRUD
│   ├── client/
│   │   ├── types.go                    # Tencent VectorDB REST request/response types
│   │   ├── client.go                   # HTTP REST client implementation (auth, CRUD API)
│   │   └── client_test.go              # Unit tests with httptest server
│   └── service/
│       ├── connection.go               # ConnectionService (exposed to Wails)
│       ├── vectordb.go                 # VectorDBService (exposed to Wails)
│       └── service_test.go             # Unit tests for services
├── frontend/
│   ├── package.json                    # Node dependencies (vue, lucide-vue-next, tailwindcss)
│   ├── vite.config.ts                  # Vite build configuration
│   ├── tailwind.config.js              # Tailwind styling config
│   ├── postcss.config.js               # PostCSS config
│   ├── index.html                      # HTML entrypoint
│   └── src/
│       ├── main.ts                     # Vue entrypoint
│       ├── style.css                   # Global styles & Tailwind directives
│       ├── App.vue                     # Main application layout
│       ├── types/
│       │   └── index.ts                # TypeScript interfaces for UI models
│       ├── components/
│       │   ├── Sidebar.vue             # Left sidebar (connections + database tree)
│       │   ├── ConnectionModal.vue     # Modal for adding/editing connection config
│       │   ├── CollectionTree.vue      # Database & Collection tree navigator
│       │   ├── MainView.vue            # Right main content area with tabs
│       │   ├── DataExplorer.vue        # Document table, pagination, filter
│       │   ├── SchemaViewer.vue        # Schema fields & index metadata
│       │   └── JsonDetailModal.vue     # Drawer/Modal to view full document & vector
│       └── stores/
│           ├── connection.ts           # Reactive connection state
│           └── vectordb.ts             # Reactive database/collection/document state
```

---

### Task 1: Scaffold Wails v3 Project with Vue 3 & Tailwind CSS

**Files:**
- Create: `main.go`
- Create: `Taskfile.yml`
- Create: `go.mod`
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tailwind.config.js`
- Create: `frontend/postcss.config.js`
- Create: `frontend/src/style.css`
- Create: `frontend/src/main.ts`
- Create: `frontend/src/App.vue`
- Create: `frontend/index.html`

- [ ] **Step 1: Initialize Wails v3 project structure using template**

```bash
wails3 init -n tcvectordb-gui -t vue -q -d .
```

- [ ] **Step 2: Install frontend dependencies including Tailwind CSS and Lucide icons**

```bash
cd frontend && pnpm add -D tailwindcss@3.4.17 postcss autoprefixer && pnpm add lucide-vue-next
```

- [ ] **Step 3: Configure Tailwind CSS and PostCSS**

Create `frontend/tailwind.config.js`:
```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

Create `frontend/postcss.config.js`:
```javascript
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
```

Add directives to `frontend/src/style.css`:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;

html, body, #app {
  height: 100%;
  margin: 0;
  padding: 0;
  overflow: hidden;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  user-select: none;
}
```

- [ ] **Step 4: Verify frontend builds successfully**

Run:
```bash
cd frontend && pnpm run build
```
Expected: `frontend/dist/` is generated without errors.

- [ ] **Step 5: Commit scaffold**

```bash
git add .
git commit -m "chore: scaffold wails v3 project with vue 3 and tailwind css"
```

---

### Task 2: Local Storage for Connection Configurations

**Files:**
- Create: `backend/storage/storage.go`
- Create: `backend/storage/storage_test.go`

- [ ] **Step 1: Write failing test for storage CRUD**

Create `backend/storage/storage_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./backend/storage/... -v`
Expected: FAIL compilation ("undefined: NewStorageWithPath", "undefined: ConnectionConfig").

- [ ] **Step 3: Implement Storage and ConnectionConfig**

Create `backend/storage/storage.go`:
```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./backend/storage/... -v`
Expected: PASS

- [ ] **Step 5: Commit storage package**

```bash
git add backend/storage
git commit -m "feat(backend): add connection storage with local JSON persistence"
```

---

### Task 3: Implement Tencent VectorDB HTTP REST Client

**Files:**
- Create: `backend/client/types.go`
- Create: `backend/client/client.go`
- Create: `backend/client/client_test.go`

- [ ] **Step 1: Write types for Tencent VectorDB REST API**

Create `backend/client/types.go`:
```go
package client

type ResponseHeader struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

type ListDatabasesResponse struct {
	ResponseHeader
	Databases []struct {
		Database string `json:"database"`
	} `json:"databases"`
}

type ListCollectionsRequest struct {
	Database string `json:"database"`
}

type ListCollectionsResponse struct {
	ResponseHeader
	Collections []struct {
		Collection string `json:"collection"`
	} `json:"collections"`
}

type DescribeCollectionRequest struct {
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

type FieldMeta struct {
	FieldName   string `json:"fieldName"`
	FieldType   string `json:"fieldType"`
	FieldUsage  string `json:"fieldUsage"`
	PrimaryKey  bool   `json:"primaryKey"`
	Description string `json:"description"`
}

type IndexMeta struct {
	FieldName  string                 `json:"fieldName"`
	IndexType  string                 `json:"indexType"`
	MetricType string                 `json:"metricType"`
	Params     map[string]interface{} `json:"params"`
}

type CollectionMeta struct {
	Database    string      `json:"database"`
	Collection  string      `json:"collection"`
	ReplicaNum  int         `json:"replicaNum"`
	ShardNum    int         `json:"shardNum"`
	Description string      `json:"description"`
	Fields      []FieldMeta `json:"fields"`
	Indexes     []IndexMeta `json:"indexes"`
}

type DescribeCollectionResponse struct {
	ResponseHeader
	Collection CollectionMeta `json:"collection"`
}

type QueryDocumentRequest struct {
	Database   string `json:"database"`
	Collection string `json:"collection"`
	Query      struct {
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
		Filter string `json:"filter,omitempty"`
	} `json:"query"`
}

type QueryDocumentResponse struct {
	ResponseHeader
	Count     int                      `json:"count"`
	Documents []map[string]interface{} `json:"documents"`
}
```

- [ ] **Step 2: Write failing unit tests with mock HTTP server**

Create `backend/client/client_test.go`:
```go
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVectorDBClient(t *testing.T) {
	mux := http.NewServeMux()

	// Mock /database/list
	mux.HandleFunc("/database/list", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer account=root&api_key=test-key" {
			http.Error(w, `{"code":401,"msg":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		resp := ListDatabasesResponse{
			ResponseHeader: ResponseHeader{Code: 0, Message: "Success"},
		}
		resp.Databases = []struct {
			Database string `json:"database"`
		}{
			{Database: "db_test"},
			{Database: "db_prod"},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Mock /collection/list
	mux.HandleFunc("/collection/list", func(w http.ResponseWriter, r *http.Request) {
		var req ListCollectionsRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Database != "db_test" {
			http.Error(w, `{"code":404,"msg":"db not found"}`, http.StatusNotFound)
			return
		}
		resp := ListCollectionsResponse{
			ResponseHeader: ResponseHeader{Code: 0, Message: "Success"},
		}
		resp.Collections = []struct {
			Collection string `json:"collection"`
		}{
			{Collection: "coll_articles"},
		}
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cli := NewClient(server.URL, "root", "test-key", 5*time.Second)

	// Test ListDatabases
	dbs, err := cli.ListDatabases(context.Background())
	if err != nil {
		t.Fatalf("ListDatabases failed: %v", err)
	}
	if len(dbs) != 2 || dbs[0] != "db_test" {
		t.Fatalf("unexpected databases: %+v", dbs)
	}

	// Test ListCollections
	colls, err := cli.ListCollections(context.Background(), "db_test")
	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}
	if len(colls) != 1 || colls[0] != "coll_articles" {
		t.Fatalf("unexpected collections: %+v", colls)
	}

	// Test Ping
	ok, msg, err := cli.Ping(context.Background())
	if err != nil || !ok {
		t.Fatalf("Ping failed: ok=%v, msg=%s, err=%v", ok, msg, err)
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./backend/client/... -v`
Expected: FAIL ("undefined: NewClient").

- [ ] **Step 4: Implement Client methods**

Create `backend/client/client.go`:
```go
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	username   string
	apiKey     string
	httpClient *http.Client
}

func NewClient(rawURL, username, apiKey string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	u := strings.TrimRight(rawURL, "/")
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "http://" + u
	}
	return &Client{
		baseURL:  u,
		username: username,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, path string, reqBody, respBody interface{}) error {
	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	fullURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	authHeader := fmt.Sprintf("Bearer account=%s&api_key=%s", c.username, c.apiKey)
	req.Header.Set("Authorization", authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var header ResponseHeader
		if jsonErr := json.Unmarshal(respData, &header); jsonErr == nil && header.Message != "" {
			return fmt.Errorf("vectordb error (%d): %s", header.Code, header.Message)
		}
		return fmt.Errorf("http error %d: %s", resp.StatusCode, string(respData))
	}

	if respBody != nil {
		if err := json.Unmarshal(respData, respBody); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}
	return nil
}

func (c *Client) Ping(ctx context.Context) (bool, string, error) {
	dbs, err := c.ListDatabases(ctx)
	if err != nil {
		return false, err.Error(), err
	}
	return true, fmt.Sprintf("Connected successfully. Found %d databases.", len(dbs)), nil
}

func (c *Client) ListDatabases(ctx context.Context) ([]string, error) {
	var resp ListDatabasesResponse
	if err := c.doRequest(ctx, "/database/list", map[string]interface{}{}, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	result := make([]string, 0, len(resp.Databases))
	for _, db := range resp.Databases {
		result = append(result, db.Database)
	}
	return result, nil
}

func (c *Client) ListCollections(ctx context.Context, database string) ([]string, error) {
	req := ListCollectionsRequest{Database: database}
	var resp ListCollectionsResponse
	if err := c.doRequest(ctx, "/collection/list", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	result := make([]string, 0, len(resp.Collections))
	for _, coll := range resp.Collections {
		result = append(result, coll.Collection)
	}
	return result, nil
}

func (c *Client) DescribeCollection(ctx context.Context, database, collection string) (*CollectionMeta, error) {
	req := DescribeCollectionRequest{Database: database, Collection: collection}
	var resp DescribeCollectionResponse
	if err := c.doRequest(ctx, "/collection/describe", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	return &resp.Collection, nil
}

func (c *Client) QueryDocuments(ctx context.Context, database, collection string, limit, offset int, filter string) (*QueryDocumentResponse, error) {
	req := QueryDocumentRequest{
		Database:   database,
		Collection: collection,
	}
	if limit <= 0 {
		limit = 20
	}
	req.Query.Limit = limit
	req.Query.Offset = offset
	req.Query.Filter = filter

	var resp QueryDocumentResponse
	if err := c.doRequest(ctx, "/document/query", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("api error (%d): %s", resp.Code, resp.Message)
	}
	return &resp, nil
}
```

- [ ] **Step 5: Run tests and verify**

Run: `go test ./backend/client/... -v`
Expected: PASS

- [ ] **Step 6: Commit client package**

```bash
git add backend/client
git commit -m "feat(backend): implement tencent vectordb REST client with unit tests"
```

---

### Task 4: Implement Wails Services (ConnectionService & VectorDBService)

**Files:**
- Create: `backend/service/connection.go`
- Create: `backend/service/vectordb.go`
- Create: `backend/service/service_test.go`

- [ ] **Step 1: Write failing unit test for services**

Create `backend/service/service_test.go`:
```go
package service

import (
	"os"
	"path/filepath"
	"testing"
	"vectordb-1/backend/storage"
)

func TestConnectionService(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "tcservice-test-*")
	defer os.RemoveAll(tempDir)

	store := storage.NewStorageWithPath(filepath.Join(tempDir, "conns.json"))
	svc := NewConnectionServiceWithStorage(store)

	list, err := svc.ListConnections()
	if err != nil || len(list) != 0 {
		t.Fatalf("expected empty list, got %v, err %v", list, err)
	}

	cfg := storage.ConnectionConfig{
		ID:       "id-1",
		Name:     "Local",
		URL:      "http://localhost:80",
		Username: "root",
		APIKey:   "k",
		Timeout:  10,
	}
	err = svc.SaveConnection(cfg)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	list, _ = svc.ListConnections()
	if len(list) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list))
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run: `go test ./backend/service/... -v`
Expected: FAIL ("undefined: NewConnectionServiceWithStorage").

- [ ] **Step 3: Implement ConnectionService and VectorDBService**

Create `backend/service/connection.go`:
```go
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
	cli := client.NewClient(cfg.URL, cfg.Username, cfg.APIKey, timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ok, msg, err := cli.Ping(ctx)
	if err != nil {
		return &TestConnectionResult{Success: false, Message: err.Error()}, nil
	}
	return &TestConnectionResult{Success: ok, Message: msg}, nil
}
```

Create `backend/service/vectordb.go`:
```go
package service

import (
	"context"
	"fmt"
	"sync"
	"time"
	"vectordb-1/backend/client"
	"vectordb-1/backend/storage"
)

type VectorDBService struct {
	storage *storage.Storage
	mu      sync.RWMutex
	clients map[string]*client.Client
}

func NewVectorDBService() *VectorDBService {
	return &VectorDBService{
		storage: storage.DefaultStorage(),
		clients: make(map[string]*client.Client),
	}
}

func (s *VectorDBService) getClient(connID string) (*client.Client, error) {
	s.mu.RLock()
	cli, ok := s.clients[connID]
	s.mu.RUnlock()
	if ok {
		return cli, nil
	}

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
		return nil, fmt.Errorf("connection config %s not found", connID)
	}

	timeout := time.Duration(target.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	newCli := client.NewClient(target.URL, target.Username, target.APIKey, timeout)

	s.mu.Lock()
	s.clients[connID] = newCli
	s.mu.Unlock()

	return newCli, nil
}

func (s *VectorDBService) ListDatabases(connID string) ([]string, error) {
	cli, err := s.getClient(connID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return cli.ListDatabases(ctx)
}

func (s *VectorDBService) ListCollections(connID, database string) ([]string, error) {
	cli, err := s.getClient(connID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return cli.ListCollections(ctx, database)
}

func (s *VectorDBService) DescribeCollection(connID, database, collection string) (*client.CollectionMeta, error) {
	cli, err := s.getClient(connID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return cli.QueryDocuments(ctx, params.Database, params.Collection, params.Limit, params.Offset, params.Filter)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./backend/service/... -v`
Expected: PASS

- [ ] **Step 5: Commit service package**

```bash
git add backend/service
git commit -m "feat(backend): implement connection and vectordb Wails services"
```

---

### Task 5: Register Services in `main.go` and Generate Bindings

**Files:**
- Modify: `main.go`
- Generate: `frontend/src/bindings/`

- [ ] **Step 1: Update `main.go` to register services**

Update `main.go`:
```go
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"vectordb-1/backend/service"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	connSvc := service.NewConnectionService()
	vdbSvc := service.NewVectorDBService()

	app := application.New(application.Options{
		Name:        "Tencent Cloud VectorDB Client",
		Description: "GUI Client for Tencent Cloud VectorDB",
		Services: []application.Service{
			application.NewService(connSvc),
			application.NewService(vdbSvc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Tencent Cloud VectorDB Client",
		Width:  1280,
		Height: 800,
		MinWidth: 960,
		MinHeight: 600,
		URL:    "/",
	})

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 2: Generate Wails bindings**

Run:
```bash
wails3 generate bindings
```
Expected: `frontend/bindings/` or `frontend/src/bindings/` generated with TypeScript client methods.

- [ ] **Step 3: Commit main.go and generated bindings**

```bash
git add main.go frontend/bindings frontend/src/bindings || true
git commit -m "feat(app): register backend services and generate wails bindings"
```

---

### Task 6: Frontend State Management and Connection Components

**Files:**
- Create: `frontend/src/types/index.ts`
- Create: `frontend/src/stores/connection.ts`
- Create: `frontend/src/components/ConnectionModal.vue`
- Create: `frontend/src/components/Sidebar.vue`

- [ ] **Step 1: Define TypeScript models for frontend**

Create `frontend/src/types/index.ts`:
```typescript
export interface ConnectionConfig {
  id: string;
  name: string;
  url: string;
  username: string;
  apiKey: string;
  timeout: number;
}

export interface FieldMeta {
  fieldName: string;
  fieldType: string;
  fieldUsage?: string;
  primaryKey?: boolean;
  description?: string;
}

export interface IndexMeta {
  fieldName: string;
  indexType: string;
  metricType?: string;
  params?: Record<string, any>;
}

export interface CollectionMeta {
  database: string;
  collection: string;
  replicaNum: number;
  shardNum: number;
  description: string;
  fields: FieldMeta[];
  indexes: IndexMeta[];
}

export interface QueryResult {
  code: number;
  msg: string;
  count: number;
  documents: Record<string, any>[];
}
```

- [ ] **Step 2: Implement Connection Store**

Create `frontend/src/stores/connection.ts` using Vue 3 reactive state to manage saved connections, current active connection ID, and modal visibility.

- [ ] **Step 3: Implement ConnectionModal Component**

Create `frontend/src/components/ConnectionModal.vue` with:
- Inputs: Connection Name, URL (with hint `http://ip:port`), Username (default `root`), API Key (masked input), Timeout (seconds).
- Actions: "Test Connection" (calls `ConnectionService.TestConnection`), "Save", "Cancel".
- Feedback: Success (green alert) or Error (red alert with message).

- [ ] **Step 4: Implement Sidebar Component**

Create `frontend/src/components/Sidebar.vue` with:
- Connection dropdown selector
- "+" New Connection button and Edit/Delete current connection actions
- Container for database and collection tree

- [ ] **Step 5: Verify frontend compilation**

Run: `cd frontend && pnpm run build`
Expected: PASS

- [ ] **Step 6: Commit connection UI**

```bash
git add frontend/src
git commit -m "feat(frontend): add connection store, modal and sidebar layout"
```

---

### Task 7: Database & Collection Tree Navigation

**Files:**
- Create: `frontend/src/stores/vectordb.ts`
- Create: `frontend/src/components/CollectionTree.vue`

- [ ] **Step 1: Implement VectorDB Store**

Create `frontend/src/stores/vectordb.ts` with state for:
- `databases`: list of database names
- `collectionsByDb`: map of `database -> collection[]`
- `activeDatabase`: currently selected database
- `activeCollection`: currently selected collection
- `activeCollectionMeta`: `CollectionMeta | null`
- `loading`: boolean
- Methods to fetch databases, fetch collections for a database, and select a collection.

- [ ] **Step 2: Implement CollectionTree Component**

Create `frontend/src/components/CollectionTree.vue`:
- Search input to filter collections by name
- Collapsible database nodes with folder icons
- Collection leaf nodes with database table icons
- Active selection styling
- Empty state when no databases or collections exist

- [ ] **Step 3: Test tree integration in frontend build**

Run: `cd frontend && pnpm run build`
Expected: PASS

- [ ] **Step 4: Commit collection tree**

```bash
git add frontend/src
git commit -m "feat(frontend): add database and collection tree navigation"
```

---

### Task 8: Data Explorer, Schema Viewer, and Detail Modal

**Files:**
- Create: `frontend/src/components/DataExplorer.vue`
- Create: `frontend/src/components/SchemaViewer.vue`
- Create: `frontend/src/components/JsonDetailModal.vue`
- Create: `frontend/src/components/MainView.vue`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: Implement JsonDetailModal Component**

Create `frontend/src/components/JsonDetailModal.vue`:
- Modal/Drawer that displays the full raw JSON of a selected document.
- Formatted JSON display with indentation.
- "Copy JSON" button with toast notification.

- [ ] **Step 2: Implement SchemaViewer Component**

Create `frontend/src/components/SchemaViewer.vue`:
- Overview section: Database, Collection, Shards, Replicas, Description.
- Fields Table: Field Name, Data Type, Primary Key indicator (badge), Description.
- Indexes Table: Target Field, Index Type (vector/filter), Metric Type (COSINE/L2/IP), Index Parameters.

- [ ] **Step 3: Implement DataExplorer Component**

Create `frontend/src/components/DataExplorer.vue`:
- Search & Filter bar: Input for filter expression with "Run Query" button and "Clear" button.
- Pagination bar: Limit dropdown (10, 20, 50, 100), Previous Page, Next Page, Current Offset / Total indicator.
- Dynamic Data Table:
  - Header generated dynamically from document fields.
  - Formatted cells: Strings and numbers rendered plainly; vector arrays formatted as `[0.123, -0.456, ... (N dims)]` with a clickable badge.
  - Action column: "View JSON" button to open `JsonDetailModal`.
- Loading spinner and empty state when 0 documents match.

- [ ] **Step 4: Implement MainView and Assemble App.vue**

Create `frontend/src/components/MainView.vue`:
- Top header with active collection breadcrumbs and tab switcher: `[Data Explorer]` / `[Schema Definition]`.
- Displays `DataExplorer` or `SchemaViewer` based on active tab.
- Placeholder state when no collection is selected ("Select a collection on the left to start").

Update `frontend/src/App.vue` to combine `Sidebar` on the left and `MainView` on the right.

- [ ] **Step 5: Verify frontend builds without errors**

Run: `cd frontend && pnpm run build`
Expected: PASS

- [ ] **Step 6: Commit complete frontend**

```bash
git add frontend/src
git commit -m "feat(frontend): add data explorer, schema viewer, and json detail inspector"
```

---

### Task 9: End-to-End Build and Verification

**Files:**
- Verify: `go.mod`, `main.go`, `frontend/`

- [ ] **Step 1: Run all Go unit tests**

Run: `go test ./backend/... -v`
Expected: All tests pass.

- [ ] **Step 2: Compile frontend bundle**

Run: `cd frontend && pnpm run build`
Expected: Builds cleanly to `frontend/dist/`.

- [ ] **Step 3: Build application executable**

Run: `go build -o bin/tcvectordb-gui .`
Expected: Binary `bin/tcvectordb-gui` generated cleanly.

- [ ] **Step 4: Smoke test executable**

Run: `./bin/tcvectordb-gui --help || timeout 2 ./bin/tcvectordb-gui || true`
Expected: Binary executes without immediate runtime panic.

- [ ] **Step 5: Commit final application bundle**

```bash
git add .
git commit -m "chore: verify end-to-end build and smoke test"
```
