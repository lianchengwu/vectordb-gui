package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestDeleteDocument_Base verifies deletion of a Base collection document
// sends POST /document/delete with documentIds and propagates success.
func TestDeleteDocument_Base(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/document/delete", func(w http.ResponseWriter, r *http.Request) {
		var req DeleteDocumentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"code":400,"msg":"bad json"}`, http.StatusBadRequest)
			return
		}
		if req.Database != "db_test" || req.Collection != "coll_articles" {
			http.Error(w, `{"code":404,"msg":"not found"}`, http.StatusNotFound)
			return
		}
		if len(req.Query.DocumentIds) != 1 || req.Query.DocumentIds[0] != "doc-1" {
			t.Errorf("unexpected query documentIds: %+v", req.Query.DocumentIds)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code": 0, "msg": "operation success", "affectedCount": 1}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cli := NewClient(server.URL, "root", "test-key", 5*time.Second)
	if err := cli.DeleteDocument(context.Background(), "db_test", "coll_articles", "base", UpdateDocumentQuery{
		DocumentIds: []string{"doc-1"},
	}); err != nil {
		t.Fatalf("DeleteDocument failed: %v", err)
	}
}

// TestDeleteDocument_AI verifies deletion of an AI knowledge-base DocumentSet
// sends POST /ai/documentSet/delete with documentSetId(s).
func TestDeleteDocument_AI(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ai/documentSet/delete", func(w http.ResponseWriter, r *http.Request) {
		var req AIDocumentSetDeleteReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"code":400,"msg":"bad json"}`, http.StatusBadRequest)
			return
		}
		if req.Database != "db_test" || req.CollectionView != "coll_kb" {
			t.Errorf("unexpected db/collectionView: %s/%s", req.Database, req.CollectionView)
		}
		if len(req.Query.DocumentSetId) != 1 || req.Query.DocumentSetId[0] != "ds-1" {
			t.Errorf("unexpected query documentSetId: %+v", req.Query.DocumentSetId)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code": 0, "msg": "operation success", "affectedCount": 1}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cli := NewClient(server.URL, "root", "test-key", 5*time.Second)
	if err := cli.DeleteDocument(context.Background(), "db_test", "coll_kb", "ai", UpdateDocumentQuery{
		DocumentIds: []string{"ds-1"},
	}); err != nil {
		t.Fatalf("DeleteDocument failed: %v", err)
	}
}

// TestDeleteDocument_BaseFallsBackToAI verifies that when Base delete
// returns 11100 (AI-only data), the client retries via the AI DocumentSet API.
func TestDeleteDocument_BaseFallsBackToAI(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/document/delete", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code": 11100, "msg": "user can only handle base data"}`))
	})
	aiCalled := false
	mux.HandleFunc("/ai/documentSet/delete", func(w http.ResponseWriter, r *http.Request) {
		aiCalled = true
		var req AIDocumentSetDeleteReq
		json.NewDecoder(r.Body).Decode(&req)
		if len(req.Query.DocumentSetId) != 1 || req.Query.DocumentSetId[0] != "ds-1" {
			t.Errorf("unexpected query documentSetId: %+v", req.Query.DocumentSetId)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code": 0, "msg": "operation success", "affectedCount": 1}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cli := NewClient(server.URL, "root", "test-key", 5*time.Second)
	if err := cli.DeleteDocument(context.Background(), "db_test", "coll_kb", "base", UpdateDocumentQuery{
		DocumentIds: []string{"ds-1"},
	}); err != nil {
		t.Fatalf("DeleteDocument failed: %v", err)
	}
	if !aiCalled {
		t.Fatal("expected fallback to /ai/documentSet/delete")
	}
}

// TestDeleteDocument_EmptyQuery verifies deletion without any selector is rejected.
func TestDeleteDocument_EmptyQuery(t *testing.T) {
	server := httptest.NewServer(http.NewServeMux())
	defer server.Close()

	cli := NewClient(server.URL, "root", "test-key", 5*time.Second)
	err := cli.DeleteDocument(context.Background(), "db_test", "coll_articles", "base", UpdateDocumentQuery{})
	if err == nil {
		t.Fatal("expected error for empty query, got nil")
	}
}

// TestDeleteDocument_ApiError verifies non-zero API codes surface as errors.
func TestDeleteDocument_ApiError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/document/delete", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code": 1200, "msg": "document not exist"}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	cli := NewClient(server.URL, "root", "test-key", 5*time.Second)
	err := cli.DeleteDocument(context.Background(), "db_test", "coll_articles", "base", UpdateDocumentQuery{
		DocumentIds: []string{"missing"},
	})
	if err == nil {
		t.Fatal("expected error for api code 1200, got nil")
	}
}
