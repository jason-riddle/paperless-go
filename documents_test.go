package paperless

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_ListDocuments(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/documents/" {
				t.Errorf("path = %v, want /api/documents/", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(DocumentList{
				Count: 1,
				Results: []Document{
					{
						ID:    1,
						Title: "Test Document",
					},
				},
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		docs, err := c.ListDocuments(context.Background(), nil)
		if err != nil {
			t.Fatalf("ListDocuments failed: %v", err)
		}
		if docs.Count != 1 {
			t.Errorf("count = %d, want 1", docs.Count)
		}
		if len(docs.Results) != 1 {
			t.Errorf("len(results) = %d, want 1", len(docs.Results))
		}
		if docs.Results[0].ID != 1 {
			t.Errorf("document ID = %d, want 1", docs.Results[0].ID)
		}
	})

	t.Run("with options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			if query.Get("page") != "2" {
				t.Errorf("page = %v, want 2", query.Get("page"))
			}
			if query.Get("page_size") != "10" {
				t.Errorf("page_size = %v, want 10", query.Get("page_size"))
			}
			if query.Get("query") != "test" {
				t.Errorf("query = %v, want test", query.Get("query"))
			}
			if query.Get("ordering") != "-created" {
				t.Errorf("ordering = %v, want -created", query.Get("ordering"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(DocumentList{Count: 0, Results: []Document{}})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		opts := &ListOptions{
			Page:     2,
			PageSize: 10,
			Query:    "test",
			Ordering: "-created",
		}
		_, err := c.ListDocuments(context.Background(), opts)
		if err != nil {
			t.Fatalf("ListDocuments failed: %v", err)
		}
	})

	t.Run("title only search", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			if query.Get("title__icontains") != "invoice" {
				t.Errorf("title__icontains = %v, want invoice", query.Get("title__icontains"))
			}
			if query.Get("query") != "" {
				t.Errorf("query should be empty when title__icontains is set, got %v", query.Get("query"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(DocumentList{Count: 0, Results: []Document{}})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		opts := &ListOptions{
			Query:     "invoice",
			TitleOnly: true,
		}
		if _, err := c.ListDocuments(context.Background(), opts); err != nil {
			t.Fatalf("ListDocuments with title-only failed: %v", err)
		}
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		_, err := c.ListDocuments(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		apiErr, ok := err.(*Error)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}
		if apiErr.Op != "ListDocuments" {
			t.Errorf("op = %v, want ListDocuments", apiErr.Op)
		}
	})
}

func TestClient_GetDocument(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expectedDoc := Document{
			ID:               1,
			Title:            "Test Document",
			Content:          "This is test content",
			Created:          Date(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			Modified:         Date(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
			Added:            Date(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)),
			OriginalFileName: "test.pdf",
			Tags:             []int{1, 2, 3},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/documents/1/" {
				t.Errorf("path = %v, want /api/documents/1/", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(expectedDoc)
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		doc, err := c.GetDocument(context.Background(), 1)
		if err != nil {
			t.Fatalf("GetDocument failed: %v", err)
		}
		if doc.ID != expectedDoc.ID {
			t.Errorf("ID = %d, want %d", doc.ID, expectedDoc.ID)
		}
		if doc.Title != expectedDoc.Title {
			t.Errorf("Title = %v, want %v", doc.Title, expectedDoc.Title)
		}
		if len(doc.Tags) != len(expectedDoc.Tags) {
			t.Errorf("len(Tags) = %d, want %d", len(doc.Tags), len(expectedDoc.Tags))
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		_, err := c.GetDocument(context.Background(), 999)
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
		if apiErr.Op != "GetDocument" {
			t.Errorf("op = %v, want GetDocument", apiErr.Op)
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
		_, err := c.GetDocument(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
