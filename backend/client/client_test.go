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

	// Mock /collection/describe
	mux.HandleFunc("/collection/describe", func(w http.ResponseWriter, r *http.Request) {
		var req DescribeCollectionRequest
		json.NewDecoder(r.Body).Decode(&req)
		resp := DescribeCollectionResponse{
			ResponseHeader: ResponseHeader{Code: 0, Message: "Success"},
		}
		resp.Collection = CollectionMeta{
			Database:   req.Database,
			Collection: req.Collection,
			ShardNum:   1,
			ReplicaNum: 1,
			Fields: []FieldMeta{
				{FieldName: "id", FieldType: "string", PrimaryKey: true},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Mock /document/query
	mux.HandleFunc("/document/query", func(w http.ResponseWriter, r *http.Request) {
		resp := QueryDocumentResponse{
			ResponseHeader: ResponseHeader{Code: 0, Message: "Success"},
			Count:          1,
			Documents: []map[string]interface{}{
				{"id": "doc-1", "vector": []interface{}{0.1, 0.2}},
			},
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

	// Test DescribeCollection
	meta, err := cli.DescribeCollection(context.Background(), "db_test", "coll_articles")
	if err != nil {
		t.Fatalf("DescribeCollection failed: %v", err)
	}
	if meta.Collection != "coll_articles" || len(meta.Fields) != 1 {
		t.Fatalf("unexpected meta: %+v", meta)
	}

	// Test QueryDocuments
	qResp, err := cli.QueryDocuments(context.Background(), "db_test", "coll_articles", 10, 0, "")
	if err != nil {
		t.Fatalf("QueryDocuments failed: %v", err)
	}
	if qResp.Count != 1 || len(qResp.Documents) != 1 {
		t.Fatalf("unexpected query resp: %+v", qResp)
	}

	// Test Ping
	ok, msg, err := cli.Ping(context.Background())
	if err != nil || !ok {
		t.Fatalf("Ping failed: ok=%v, msg=%s, err=%v", ok, msg, err)
	}
}

func TestVectorDBClient_Errors(t *testing.T) {
	mux := http.NewServeMux()

	// Unauthorized database list
	mux.HandleFunc("/database/list", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code":401,"msg":"invalid api key"}`))
	})

	// API error with code != 0 but HTTP 200
	mux.HandleFunc("/collection/list", func(w http.ResponseWriter, r *http.Request) {
		resp := ListCollectionsResponse{
			ResponseHeader: ResponseHeader{Code: 1002, Message: "database not found"},
		}
		json.NewEncoder(w).Encode(resp)
	})
	// API error on describe
	mux.HandleFunc("/collection/describe", func(w http.ResponseWriter, r *http.Request) {
		resp := DescribeCollectionResponse{
			ResponseHeader: ResponseHeader{Code: 1003, Message: "collection not found"},
		}
		json.NewEncoder(w).Encode(resp)
	})

	// API error on query
	mux.HandleFunc("/document/query", func(w http.ResponseWriter, r *http.Request) {
		resp := QueryDocumentResponse{
			ResponseHeader: ResponseHeader{Code: 1004, Message: "filter syntax error"},
		}
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cli := NewClient(server.URL, "root", "wrong-key", 2*time.Second)

	// Test unauthorized Ping
	ok, msg, err := cli.Ping(context.Background())
	if ok || err == nil {
		t.Fatalf("expected ping to fail, got ok=%v, msg=%s, err=%v", ok, msg, err)
	}

	// Test ListDatabases returning HTTP error
	_, err = cli.ListDatabases(context.Background())
	if err == nil {
		t.Fatalf("expected ListDatabases error, got nil")
	}

	// Test ListCollections returning API error
	_, err = cli.ListCollections(context.Background(), "db_test")
	if err == nil {
		t.Fatalf("expected ListCollections error, got nil")
	}

	// Test DescribeCollection returning API error
	_, err = cli.DescribeCollection(context.Background(), "db_test", "missing_coll")
	if err == nil {
		t.Fatalf("expected DescribeCollection error, got nil")
	}

	// Test QueryDocuments returning API error
	_, err = cli.QueryDocuments(context.Background(), "db_test", "coll", 0, 0, "invalid filter")
	if err == nil {
		t.Fatalf("expected QueryDocuments error, got nil")
	}
}

func TestVectorDBClient_URLFormattingAndDefaults(t *testing.T) {
	// Test without scheme and with trailing slash
	cli := NewClient("127.0.0.1:8080/", "root", "key", 0)
	if cli.baseURL != "http://127.0.0.1:8080" {
		t.Fatalf("unexpected baseURL: %s", cli.baseURL)
	}
	if cli.httpClient.Timeout != 10*time.Second {
		t.Fatalf("unexpected default timeout: %v", cli.httpClient.Timeout)
	}

	// Test with https
	cliHTTPS := NewClient("https://vdb.tencentcloudapi.com/", "root", "key", 5*time.Second)
	if cliHTTPS.baseURL != "https://vdb.tencentcloudapi.com" {
		t.Fatalf("unexpected baseURL: %s", cliHTTPS.baseURL)
	}
}
