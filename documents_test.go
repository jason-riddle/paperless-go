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
			_, _ = w.Write([]byte("Internal Server Error"))
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
			_, _ = w.Write([]byte("Not Found"))
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

func TestClient_UpdateDocument(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tags := []int{1, 2}
		update := &DocumentUpdate{
			Tags: &tags,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/documents/1/" {
				t.Errorf("path = %v, want /api/documents/1/", r.URL.Path)
			}
			if r.Method != "PATCH" {
				t.Errorf("method = %v, want PATCH", r.Method)
			}

			// Verify body
			var decoded DocumentUpdate
			if err := json.NewDecoder(r.Body).Decode(&decoded); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if decoded.Tags == nil {
				t.Fatal("tags is nil")
			}
			if len(*decoded.Tags) != 2 || (*decoded.Tags)[0] != 1 || (*decoded.Tags)[1] != 2 {
				t.Errorf("tags = %v, want [1, 2]", *decoded.Tags)
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Document{
				ID:    1,
				Title: "Updated Document",
				Tags:  []int{1, 2},
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		doc, err := c.UpdateDocument(context.Background(), 1, update)
		if err != nil {
			t.Fatalf("UpdateDocument failed: %v", err)
		}
		if doc.ID != 1 {
			t.Errorf("ID = %d, want 1", doc.ID)
		}
		if len(doc.Tags) != 2 {
			t.Errorf("len(Tags) = %d, want 2", len(doc.Tags))
		}
	})
}

func TestClient_RenameDocument(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		newTitle := "New Document Title"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/documents/1/" {
				t.Errorf("path = %v, want /api/documents/1/", r.URL.Path)
			}
			if r.Method != "PATCH" {
				t.Errorf("method = %v, want PATCH", r.Method)
			}

			// Verify body contains title
			var decoded DocumentUpdate
			if err := json.NewDecoder(r.Body).Decode(&decoded); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if decoded.Title == nil {
				t.Fatal("title is nil, expected non-nil")
			}
			if *decoded.Title != newTitle {
				t.Errorf("title = %v, want %v", *decoded.Title, newTitle)
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Document{
				ID:    1,
				Title: newTitle,
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		doc, err := c.RenameDocument(context.Background(), 1, newTitle)
		if err != nil {
			t.Fatalf("RenameDocument failed: %v", err)
		}
		if doc.ID != 1 {
			t.Errorf("ID = %d, want 1", doc.ID)
		}
		if doc.Title != newTitle {
			t.Errorf("Title = %v, want %v", doc.Title, newTitle)
		}
	})

	t.Run("empty title error", func(t *testing.T) {
		c := NewClient("http://example.com", "test-token")
		_, err := c.RenameDocument(context.Background(), 1, "")
		if err == nil {
			t.Fatal("expected error for empty title, got nil")
		}
		expectedMsg := "RenameDocument: title cannot be empty"
		if err.Error() != expectedMsg {
			t.Errorf("error message = %v, want %v", err.Error(), expectedMsg)
		}
	})

	t.Run("invalid document ID error", func(t *testing.T) {
		c := NewClient("http://example.com", "test-token")
		_, err := c.RenameDocument(context.Background(), 0, "New Title")
		if err == nil {
			t.Fatal("expected error for invalid document ID, got nil")
		}
		expectedMsg := "RenameDocument: invalid document ID: 0"
		if err.Error() != expectedMsg {
			t.Errorf("error message = %v, want %v", err.Error(), expectedMsg)
		}
	})

	t.Run("negative document ID error", func(t *testing.T) {
		c := NewClient("http://example.com", "test-token")
		_, err := c.RenameDocument(context.Background(), -1, "New Title")
		if err == nil {
			t.Fatal("expected error for negative document ID, got nil")
		}
		expectedMsg := "RenameDocument: invalid document ID: -1"
		if err.Error() != expectedMsg {
			t.Errorf("error message = %v, want %v", err.Error(), expectedMsg)
		}
	})

	t.Run("not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not Found"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		_, err := c.RenameDocument(context.Background(), 999, "New Title")
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
		if apiErr.Op != "RenameDocument" {
			t.Errorf("op = %v, want RenameDocument", apiErr.Op)
		}
	})
}

func TestClient_UpdateDocumentTags(t *testing.T) {
	t.Run("success with multiple tags", func(t *testing.T) {
		tagIDs := []int{1, 2, 3}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/documents/1/" {
				t.Errorf("path = %v, want /api/documents/1/", r.URL.Path)
			}
			if r.Method != "PATCH" {
				t.Errorf("method = %v, want PATCH", r.Method)
			}

			// Verify body contains tags
			var decoded DocumentUpdate
			if err := json.NewDecoder(r.Body).Decode(&decoded); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if decoded.Tags == nil {
				t.Fatal("tags is nil")
			}
			if len(*decoded.Tags) != 3 {
				t.Errorf("len(tags) = %d, want 3", len(*decoded.Tags))
			}
			for i, tag := range *decoded.Tags {
				if tag != tagIDs[i] {
					t.Errorf("tags[%d] = %d, want %d", i, tag, tagIDs[i])
				}
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Document{
				ID:   1,
				Tags: tagIDs,
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		doc, err := c.UpdateDocumentTags(context.Background(), 1, tagIDs)
		if err != nil {
			t.Fatalf("UpdateDocumentTags failed: %v", err)
		}
		if doc.ID != 1 {
			t.Errorf("ID = %d, want 1", doc.ID)
		}
		if len(doc.Tags) != 3 {
			t.Errorf("len(Tags) = %d, want 3", len(doc.Tags))
		}
	})

	t.Run("success with empty tags (remove all)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify body contains empty tags array
			var decoded DocumentUpdate
			if err := json.NewDecoder(r.Body).Decode(&decoded); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if decoded.Tags == nil {
				t.Fatal("tags is nil, expected non-nil pointer to empty slice")
			}
			if len(*decoded.Tags) != 0 {
				t.Errorf("len(tags) = %d, want 0", len(*decoded.Tags))
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Document{
				ID:   1,
				Tags: []int{},
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		doc, err := c.UpdateDocumentTags(context.Background(), 1, []int{})
		if err != nil {
			t.Fatalf("UpdateDocumentTags failed: %v", err)
		}
		if doc.ID != 1 {
			t.Errorf("ID = %d, want 1", doc.ID)
		}
		if len(doc.Tags) != 0 {
			t.Errorf("len(Tags) = %d, want 0", len(doc.Tags))
		}
	})

	t.Run("nil tags converted to empty slice", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify body contains empty tags array
			var decoded DocumentUpdate
			if err := json.NewDecoder(r.Body).Decode(&decoded); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}
			if decoded.Tags == nil {
				t.Fatal("tags is nil, expected non-nil pointer to empty slice")
			}
			if len(*decoded.Tags) != 0 {
				t.Errorf("len(tags) = %d, want 0", len(*decoded.Tags))
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(Document{
				ID:   1,
				Tags: []int{},
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		doc, err := c.UpdateDocumentTags(context.Background(), 1, nil)
		if err != nil {
			t.Fatalf("UpdateDocumentTags failed: %v", err)
		}
		if doc.ID != 1 {
			t.Errorf("ID = %d, want 1", doc.ID)
		}
	})

	t.Run("invalid document ID error", func(t *testing.T) {
		c := NewClient("http://example.com", "test-token")
		_, err := c.UpdateDocumentTags(context.Background(), 0, []int{1, 2})
		if err == nil {
			t.Fatal("expected error for invalid document ID, got nil")
		}
		expectedMsg := "UpdateDocumentTags: invalid document ID: 0"
		if err.Error() != expectedMsg {
			t.Errorf("error message = %v, want %v", err.Error(), expectedMsg)
		}
	})

	t.Run("negative document ID error", func(t *testing.T) {
		c := NewClient("http://example.com", "test-token")
		_, err := c.UpdateDocumentTags(context.Background(), -1, []int{1, 2})
		if err == nil {
			t.Fatal("expected error for negative document ID, got nil")
		}
		expectedMsg := "UpdateDocumentTags: invalid document ID: -1"
		if err.Error() != expectedMsg {
			t.Errorf("error message = %v, want %v", err.Error(), expectedMsg)
		}
	})

	t.Run("invalid tag ID error", func(t *testing.T) {
		c := NewClient("http://example.com", "test-token")
		_, err := c.UpdateDocumentTags(context.Background(), 1, []int{1, 0, 3})
		if err == nil {
			t.Fatal("expected error for invalid tag ID, got nil")
		}
		expectedMsg := "UpdateDocumentTags: invalid tag ID at index 1: 0"
		if err.Error() != expectedMsg {
			t.Errorf("error message = %v, want %v", err.Error(), expectedMsg)
		}
	})

	t.Run("negative tag ID error", func(t *testing.T) {
		c := NewClient("http://example.com", "test-token")
		_, err := c.UpdateDocumentTags(context.Background(), 1, []int{1, -5, 3})
		if err == nil {
			t.Fatal("expected error for negative tag ID, got nil")
		}
		expectedMsg := "UpdateDocumentTags: invalid tag ID at index 1: -5"
		if err.Error() != expectedMsg {
			t.Errorf("error message = %v, want %v", err.Error(), expectedMsg)
		}
	})

	t.Run("not found error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not Found"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		_, err := c.UpdateDocumentTags(context.Background(), 999, []int{1, 2})
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
		if apiErr.Op != "UpdateDocumentTags" {
			t.Errorf("op = %v, want UpdateDocumentTags", apiErr.Op)
		}
	})
}
