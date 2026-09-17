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

	// Mock /database/list (supports GET with string array as per Tencent VectorDB official response)
	mux.HandleFunc("/database/list", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer account=root&api_key=test-key" {
			http.Error(w, `{"code":401,"msg":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		// Real Tencent VectorDB response format with string array
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"msg": "operation success",
			"databases": ["db_test", "db_prod"],
			"affectedCount": 0
		}`))
	})

	// Mock /collection/list
	mux.HandleFunc("/collection/list", func(w http.ResponseWriter, r *http.Request) {
		var req ListCollectionsRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Database != "db_test" {
			http.Error(w, `{"code":404,"msg":"db not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"msg": "operation success",
			"collections": [
				{
					"database": "db_test",
					"collection": "coll_articles",
					"shardNum": 1,
					"replicaNum": 0,
					"indexes": [
						{"fieldName": "id", "fieldType": "string", "indexType": "primaryKey"},
						{"fieldName": "vector", "fieldType": "vector", "indexType": "HNSW", "metricType": "COSINE"}
					]
				}
			]
		}`))
	})

	// Mock /collection/describe
	mux.HandleFunc("/collection/describe", func(w http.ResponseWriter, r *http.Request) {
		var req DescribeCollectionRequest
		json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"msg": "operation success",
			"collection": {
				"database": "db_test",
				"collection": "coll_articles",
				"shardNum": 1,
				"replicaNum": 0,
				"indexes": [
					{"fieldName": "id", "fieldType": "string", "indexType": "primaryKey"},
					{"fieldName": "vector", "fieldType": "vector", "indexType": "HNSW", "metricType": "COSINE"}
				]
			}
		}`))
	})

	// Mock /collection/drop
	mux.HandleFunc("/collection/drop", func(w http.ResponseWriter, r *http.Request) {
		var req DropCollectionRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Database == "" || req.Collection == "" {
			http.Error(w, `{"code":400,"msg":"database and collection required"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"msg": "operation success",
			"affectedCount": 1
		}`))
	})

	// Mock /document/update
	mux.HandleFunc("/document/update", func(w http.ResponseWriter, r *http.Request) {
		var req UpdateDocumentRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Database == "" || req.Collection == "" {
			http.Error(w, `{"code":400,"msg":"database and collection required"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"msg": "operation success",
			"affectedCount": 1
		}`))
	})

	// Mock /document/query
	mux.HandleFunc("/document/query", func(w http.ResponseWriter, r *http.Request) {
		var req QueryDocumentRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Query == nil || !req.Query.RetrieveVector {
			http.Error(w, `{"code":400,"msg":"retrieveVector must be true"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"code": 0,
			"msg": "operation success",
			"count": 1,
			"documents": [
				{"id": "doc-1", "vector": [0.1, 0.2]}
			]
		}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cli := NewClient(server.URL, "root", "test-key", 5*time.Second)

	// Test ListDatabases
	dbs, err := cli.ListDatabases(context.Background())
	if err != nil {
		t.Fatalf("ListDatabases failed: %v", err)
	}
	if len(dbs) != 2 || dbs[0] != "db_test" || dbs[1] != "db_prod" {
		t.Fatalf("unexpected databases: %+v", dbs)
	}

	// Test ListCollections
	colls, err := cli.ListCollections(context.Background(), "db_test", "base")
	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}
	if len(colls) != 1 || colls[0] != "coll_articles" {
		t.Fatalf("unexpected collections: %+v", colls)
	}

	// Test DescribeCollection (with auto-derived fields from indexes)
	meta, err := cli.DescribeCollection(context.Background(), "db_test", "coll_articles", "base")
	if err != nil {
		t.Fatalf("DescribeCollection failed: %v", err)
	}
	if meta.Collection != "coll_articles" || len(meta.Fields) != 2 {
		t.Fatalf("unexpected meta: %+v", meta)
	}

	// Test QueryDocuments
	qResp, err := cli.QueryDocuments(context.Background(), "db_test", "coll_articles", "base", 10, 0, "")
	if err != nil {
		t.Fatalf("QueryDocuments failed: %v", err)
	}
	if qResp.Count != 1 || len(qResp.Documents) != 1 {
		t.Fatalf("unexpected query resp: %+v", qResp)
	}

	// Test DropCollection
	if err := cli.DropCollection(context.Background(), "db_test", "coll_articles", "base"); err != nil {
		t.Fatalf("DropCollection failed: %v", err)
	}

	// Test UpdateDocument (Base)
	updateErr := cli.UpdateDocument(context.Background(), "db_test", "coll_articles", "base", UpdateDocumentQuery{
		DocumentIds: []string{"doc-1"},
	}, map[string]interface{}{"title": "updated title"})
	if updateErr != nil {
		t.Fatalf("UpdateDocument failed: %v", updateErr)
	}

	// Test Ping
	ok, msg, err := cli.Ping(context.Background())
	if err != nil || !ok {
		t.Fatalf("Ping failed: ok=%v, msg=%s, err=%v", ok, msg, err)
	}
}

func TestVectorDBClient_ObjectDatabases(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/database/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Test object format
		w.Write([]byte(`{
			"code": 0,
			"msg": "success",
			"databases": [{"database": "db1"}, {"name": "db2"}]
		}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	cli := NewClient(server.URL, "root", "test-key", 5*time.Second)
	dbs, err := cli.ListDatabases(context.Background())
	if err != nil {
		t.Fatalf("ListDatabases failed with object array: %v", err)
	}
	if len(dbs) != 2 || dbs[0] != "db1" || dbs[1] != "db2" {
		t.Fatalf("unexpected dbs: %+v", dbs)
	}
}

func TestSanitizeUpdate(t *testing.T) {
	// Nested metaData should be unpacked to top-level scalars
	input := map[string]interface{}{
		"title": "Document Title",
		"metaData": map[string]interface{}{
			"is_enabled": true,
			"author":     "Alice",
		},
	}
	res := sanitizeUpdate(input)
	if res["title"] != "Document Title" {
		t.Errorf("expected title to be 'Document Title', got %v", res["title"])
	}
	if res["is_enabled"] != true {
		t.Errorf("expected is_enabled to be true, got %v", res["is_enabled"])
	}
	if res["author"] != "Alice" {
		t.Errorf("expected author to be 'Alice', got %v", res["author"])
	}
	if _, exists := res["metaData"]; exists {
		t.Errorf("expected metaData to be flattened, but it still exists")
	}

	// Deeply nested non-meta object
	input2 := map[string]interface{}{
		"extra": map[string]interface{}{
			"count": 42,
		},
	}
	res2 := sanitizeUpdate(input2)
	if res2["extra_count"] != 42 {
		t.Errorf("expected extra_count to be 42, got %v", res2["extra_count"])
	}
}

func TestFilterAIRuntimeFields(t *testing.T) {
	input := map[string]interface{}{
		"documentSetName":    "my_doc.pdf",
		"documentsetid":      "12345",
		"status":             "ready",
		"byteSize":           1024,
		"appendTitleToChunk": false,
		"author":             "tencent",
		"is_enabled":         true,
		"is_open":            false,
		"weight":             float64(42),
	}
	filtered := filterAIRuntimeFields(input)
	if len(filtered) != 4 {
		t.Errorf("expected 4 custom fields, got %d: %+v", len(filtered), filtered)
	}
	if filtered["author"] != "tencent" {
		t.Errorf("expected author to be 'tencent', got %v", filtered["author"])
	}
	if filtered["is_enabled"] != uint64(1) {
		t.Errorf("expected is_enabled to be uint64(1), got %v (type %T)", filtered["is_enabled"], filtered["is_enabled"])
	}
	if filtered["is_open"] != uint64(0) {
		t.Errorf("expected is_open to be uint64(0), got %v (type %T)", filtered["is_open"], filtered["is_open"])
	}
	if filtered["weight"] != uint64(42) {
		t.Errorf("expected weight to be uint64(42), got %v (type %T)", filtered["weight"], filtered["weight"])
	}
	if _, ok := filtered["documentSetName"]; ok {
		t.Errorf("expected documentSetName to be filtered out")
	}
	if _, ok := filtered["appendTitleToChunk"]; ok {
		t.Errorf("expected appendTitleToChunk to be filtered out")
	}
	if _, ok := filtered["status"]; ok {
		t.Errorf("expected status to be filtered out")
	}
}


