package paperless

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_ListDocumentTypes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/document_types/" {
				t.Errorf("path = %v, want /api/document_types/", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(DocumentTypeList{
				Count: 2,
				Results: []DocumentType{
					{
						ID:                1,
						Name:              "Bank Statement",
						Slug:              "bank-statement",
						Match:             "statement",
						MatchingAlgorithm: 1,
						IsInsensitive:     true,
						DocumentCount:     120,
					},
					{
						ID:                2,
						Name:              "Invoice",
						Slug:              "invoice",
						Match:             "invoice",
						MatchingAlgorithm: 1,
						IsInsensitive:     true,
						DocumentCount:     85,
					},
				},
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		docTypes, err := c.ListDocumentTypes(context.Background(), nil)
		if err != nil {
			t.Fatalf("ListDocumentTypes failed: %v", err)
		}
		if docTypes.Count != 2 {
			t.Errorf("count = %d, want 2", docTypes.Count)
		}
		if len(docTypes.Results) != 2 {
			t.Errorf("len(results) = %d, want 2", len(docTypes.Results))
		}
		if docTypes.Results[0].Name != "Bank Statement" {
			t.Errorf("document type name = %v, want Bank Statement", docTypes.Results[0].Name)
		}
	})

	t.Run("with options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			if query.Get("page") != "1" {
				t.Errorf("page = %v, want 1", query.Get("page"))
			}
			if query.Get("ordering") != "name" {
				t.Errorf("ordering = %v, want name", query.Get("ordering"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(DocumentTypeList{Count: 0, Results: []DocumentType{}})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		opts := &ListOptions{
			Page:     1,
			Ordering: "name",
		}
		_, err := c.ListDocumentTypes(context.Background(), opts)
		if err != nil {
			t.Fatalf("ListDocumentTypes failed: %v", err)
		}
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("Unauthorized"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		_, err := c.ListDocumentTypes(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		apiErr, ok := err.(*Error)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}
		if apiErr.Op != "ListDocumentTypes" {
			t.Errorf("op = %v, want ListDocumentTypes", apiErr.Op)
		}
		if apiErr.StatusCode != 401 {
			t.Errorf("status code = %d, want 401", apiErr.StatusCode)
		}
	})
}

func TestClient_GetDocumentType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expectedDocType := DocumentType{
			ID:                1,
			Name:              "Bank Statement",
			Slug:              "bank-statement",
			Match:             "statement",
			MatchingAlgorithm: 1,
			IsInsensitive:     true,
			DocumentCount:     120,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/document_types/1/" {
				t.Errorf("path = %v, want /api/document_types/1/", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(expectedDocType)
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		docType, err := c.GetDocumentType(context.Background(), 1)
		if err != nil {
			t.Fatalf("GetDocumentType failed: %v", err)
		}
		if docType.ID != expectedDocType.ID {
			t.Errorf("ID = %d, want %d", docType.ID, expectedDocType.ID)
		}
		if docType.Name != expectedDocType.Name {
			t.Errorf("Name = %v, want %v", docType.Name, expectedDocType.Name)
		}
		if docType.Slug != expectedDocType.Slug {
			t.Errorf("Slug = %v, want %v", docType.Slug, expectedDocType.Slug)
		}
		if docType.Match != expectedDocType.Match {
			t.Errorf("Match = %v, want %v", docType.Match, expectedDocType.Match)
		}
		if docType.MatchingAlgorithm != expectedDocType.MatchingAlgorithm {
			t.Errorf("MatchingAlgorithm = %d, want %d", docType.MatchingAlgorithm, expectedDocType.MatchingAlgorithm)
		}
		if docType.IsInsensitive != expectedDocType.IsInsensitive {
			t.Errorf("IsInsensitive = %v, want %v", docType.IsInsensitive, expectedDocType.IsInsensitive)
		}
		if docType.DocumentCount != expectedDocType.DocumentCount {
			t.Errorf("DocumentCount = %d, want %d", docType.DocumentCount, expectedDocType.DocumentCount)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not Found"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		_, err := c.GetDocumentType(context.Background(), 999)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected 404 error, got %v", err)
		}
		apiErr, ok := err.(*Error)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}
		if apiErr.Op != "GetDocumentType" {
			t.Errorf("op = %v, want GetDocumentType", apiErr.Op)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		_, err := c.GetDocumentType(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
