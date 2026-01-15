package paperless

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_ListTags(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/tags/" {
				t.Errorf("path = %v, want /api/tags/", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(TagList{
				Count: 2,
				Results: []Tag{
					{
						ID:            1,
						Name:          "Important",
						Slug:          "important",
						Color:         "#ff0000",
						DocumentCount: 5,
					},
					{
						ID:            2,
						Name:          "Work",
						Slug:          "work",
						Color:         "#00ff00",
						DocumentCount: 10,
					},
				},
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		tags, err := c.ListTags(context.Background(), nil)
		if err != nil {
			t.Fatalf("ListTags failed: %v", err)
		}
		if tags.Count != 2 {
			t.Errorf("count = %d, want 2", tags.Count)
		}
		if len(tags.Results) != 2 {
			t.Errorf("len(results) = %d, want 2", len(tags.Results))
		}
		if tags.Results[0].Name != "Important" {
			t.Errorf("tag name = %v, want Important", tags.Results[0].Name)
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
			_ = json.NewEncoder(w).Encode(TagList{Count: 0, Results: []Tag{}})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		opts := &ListOptions{
			Page:     1,
			Ordering: "name",
		}
		_, err := c.ListTags(context.Background(), opts)
		if err != nil {
			t.Fatalf("ListTags failed: %v", err)
		}
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("Unauthorized"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		_, err := c.ListTags(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		apiErr, ok := err.(*Error)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}
		if apiErr.Op != "ListTags" {
			t.Errorf("op = %v, want ListTags", apiErr.Op)
		}
		if apiErr.StatusCode != 401 {
			t.Errorf("status code = %d, want 401", apiErr.StatusCode)
		}
	})
}

func TestClient_GetTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expectedTag := Tag{
			ID:            1,
			Name:          "Important",
			Slug:          "important",
			Color:         "#ff0000",
			DocumentCount: 5,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/tags/1/" {
				t.Errorf("path = %v, want /api/tags/1/", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(expectedTag)
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		tag, err := c.GetTag(context.Background(), 1)
		if err != nil {
			t.Fatalf("GetTag failed: %v", err)
		}
		if tag.ID != expectedTag.ID {
			t.Errorf("ID = %d, want %d", tag.ID, expectedTag.ID)
		}
		if tag.Name != expectedTag.Name {
			t.Errorf("Name = %v, want %v", tag.Name, expectedTag.Name)
		}
		if tag.Slug != expectedTag.Slug {
			t.Errorf("Slug = %v, want %v", tag.Slug, expectedTag.Slug)
		}
		if tag.Color != expectedTag.Color {
			t.Errorf("Color = %v, want %v", tag.Color, expectedTag.Color)
		}
		if tag.DocumentCount != expectedTag.DocumentCount {
			t.Errorf("DocumentCount = %d, want %d", tag.DocumentCount, expectedTag.DocumentCount)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not Found"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		_, err := c.GetTag(context.Background(), 999)
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
		if apiErr.Op != "GetTag" {
			t.Errorf("op = %v, want GetTag", apiErr.Op)
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
		_, err := c.GetTag(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
