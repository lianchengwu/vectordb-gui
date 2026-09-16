# Tencent Cloud VectorDB GUI Client Design Specification

## 1. Overview
A lightweight, modern desktop GUI client built with Wails v3 and Vue 3 to connect to Tencent Cloud VectorDB. The primary purpose is to allow developers and operators to easily manage database connections, browse database and collection structures (schema and indexes), and inspect stored documents and high-dimensional vectors.

## 2. Technical Stack
- **Desktop Application Framework**: Wails v3 (`v3.0.0-beta.22`)
- **Backend**: Go (Go 1.27)
  - Networking: Standard library `net/http` based REST client (zero heavy external gRPC/protobuf dependencies, avoiding version drift)
  - Config Persistence: Local JSON configuration stored at `~/.tcvectordb/connections.json`
- **Frontend**:
  - Framework: Vue 3 (Composition API, `<script setup>`)
  - Build Tool: Vite
  - Styling: Tailwind CSS
  - Icons: Lucide Vue Next

## 3. Architecture & Backend Design

### 3.1 Directory Structure
```
vectordb-1/
├── build/                 # Wails app build assets (icons, manifests)
├── backend/
│   ├── client/
│   │   ├── client.go      # VectorDBClient HTTP REST implementation
│   │   └── types.go       # VectorDB REST API request/response structures
│   ├── service/
│   │   ├── connection.go  # ConnectionService (CRUD connection configs & ping)
│   │   └── vectordb.go    # VectorDBService (Databases, Collections, Documents)
│   └── storage/
│       └── storage.go     # Local file storage for connections
├── frontend/              # Vue 3 SPA
│   ├── src/
│   │   ├── components/    # Reusable UI components (Modals, Tables, TreeView)
│   │   ├── views/         # Workspace view, ConnectionManager view
│   │   ├── stores/        # Reactive state management
│   │   ├── App.vue
│   │   └── main.js
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
├── main.go                # Wails v3 application entrypoint
├── wails.json             # Wails project configuration
└── go.mod
```

### 3.2 Data Contracts & Models

#### Connection Configuration
```go
type ConnectionConfig struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    URL      string `json:"url"`      // e.g., http://10.0.0.1:80 or public endpoint
    Username string `json:"username"` // default "root"
    APIKey   string `json:"apiKey"`   // Account key from Tencent Cloud console
    Timeout  int    `json:"timeout"`  // In seconds, default 10
}
```

#### Collection Metadata & Schema
```go
type FieldSchema struct {
    Name        string `json:"name"`
    Type        string `json:"type"`        // string, uint64, vector, etc.
    IsPrimaryKey bool  `json:"isPrimaryKey"`
    Description string `json:"description"`
}

type IndexSchema struct {
    FieldName  string                 `json:"fieldName"`
    IndexType  string                 `json:"indexType"`  // primaryKey, filter, vector
    MetricType string                 `json:"metricType"` // COSINE, L2, IP (for vector)
    Params     map[string]interface{} `json:"params"`
}

type CollectionDetail struct {
    Database    string        `json:"database"`
    Name        string        `json:"name"`
    ShardNum    int           `json:"shardNum"`
    ReplicaNum  int           `json:"replicaNum"`
    Description string        `json:"description"`
    Fields      []FieldSchema `json:"fields"`
    Indexes     []IndexSchema `json:"indexes"`
}
```

#### Document Query & Paging
```go
type QueryDocumentRequest struct {
    ConnectionID string `json:"connectionId"`
    Database     string `json:"database"`
    Collection   string `json:"collection"`
    Limit        int    `json:"limit"`        // default 20, max 100
    Offset       int    `json:"offset"`       // default 0
    Filter       string `json:"filter"`       // scalar filter expression, e.g., "id in ('doc-1')"
}

type QueryDocumentResponse struct {
    Total     int                      `json:"total"`
    Documents []map[string]interface{} `json:"documents"`
}
```

### 3.3 Backend Services (Exposed to Wails RPC)

1. `ConnectionService`:
   - `ListConnections() []ConnectionConfig`
   - `SaveConnection(config ConnectionConfig) error`
   - `DeleteConnection(id string) error`
   - `TestConnection(config ConnectionConfig) (bool, string, error)`

2. `VectorDBService`:
   - `ListDatabases(connId string) ([]string, error)`
   - `ListCollections(connId string, database string) ([]string, error)`
   - `DescribeCollection(connId string, database string, collection string) (*CollectionDetail, error)`
   - `QueryDocuments(req QueryDocumentRequest) (*QueryDocumentResponse, error)`

## 4. Frontend Design & Interaction

### 4.1 Layout Overview
- **Sidebar (Left, 280px)**:
  - Top: Connection selector dropdown with status badge (Connected/Disconnected/Error) + "New Connection" button.
  - Middle: Searchable TreeView of Databases and Collections. Clicking a collection loads its details and documents in the main view.
  - Bottom: Quick settings, refresh button.
- **Main Content Area (Right, flex-1)**:
  - Header: Breadcrumb (`Database / Collection`), basic metrics (shards, replicas, vector index metric).
  - Tabs:
    - **Tab 1: Data Explorer (Default)**
      - Filter bar: Input for filter expression, pagination controls (Previous, Next, Limit per page: 20/50/100).
      - Table: Dynamic columns based on returned document schema.
        - Primary key highlighted.
        - Vector fields truncated to `[0.123, -0.456, ... (N dims)]` with a quick "View" action to prevent UI lagging.
        - Row Action: "View JSON" opens full document in a side drawer or dialog.
    - **Tab 2: Collection Schema**
      - Displays field definitions, data types, primary key flags, and index definitions in a structured list.
- **Connection Modal**:
  - Modal form for creating/editing connection parameters with a "Test Connection" button.

### 4.2 Handling High-Dimensional Vectors
- Vectors can have 1536 or more floating-point numbers.
- Rendering raw arrays into standard table cells can freeze the browser DOM.
- The UI will format array/vector fields with compact summaries: `[float, float, ... (N dims)]`.
- Clicking "Expand/View" opens a formatted inspector with copy-to-clipboard functionality and dimension count.

## 5. Error Handling & Edge Cases
- **Network Timeout**: All HTTP requests are bounded by user-configured timeout (default 10s).
- **Authentication Failure**: Clear error display if API key or account is invalid.
- **Empty Collections**: Gracefully render empty state illustrations and clear message.
- **Malformed Filter**: Detailed error feedback when Tencent Cloud VectorDB returns syntax errors on filter expressions.

## 6. Verification & Quality Assurance
- **Go Unit Tests**:
  - Verify configuration storage serialization and deserialization.
  - Verify REST client URL, headers, and request formatting.
- **Build & Integration Test**:
  - Verify `wails3` build compiles without errors.
  - Verify Vue frontend builds via Vite.
  - Smoke test launching GUI application window.
