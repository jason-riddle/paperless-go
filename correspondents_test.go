package paperless

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_ListCorrespondents(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/correspondents/" {
				t.Errorf("path = %v, want /api/correspondents/", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(CorrespondentList{
				Count: 2,
				Results: []Correspondent{
					{
						ID:                1,
						Name:              "SheerValue Property Management",
						Slug:              "sheervalue-property-management",
						Match:             "SheerValue",
						MatchingAlgorithm: 1,
						IsInsensitive:     true,
						DocumentCount:     45,
					},
					{
						ID:                2,
						Name:              "Bank of America",
						Slug:              "bank-of-america",
						Match:             "Bank of America",
						MatchingAlgorithm: 1,
						IsInsensitive:     true,
						DocumentCount:     30,
					},
				},
			})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		correspondents, err := c.ListCorrespondents(context.Background(), nil)
		if err != nil {
			t.Fatalf("ListCorrespondents failed: %v", err)
		}
		if correspondents.Count != 2 {
			t.Errorf("count = %d, want 2", correspondents.Count)
		}
		if len(correspondents.Results) != 2 {
			t.Errorf("len(results) = %d, want 2", len(correspondents.Results))
		}
		if correspondents.Results[0].Name != "SheerValue Property Management" {
			t.Errorf("correspondent name = %v, want SheerValue Property Management", correspondents.Results[0].Name)
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
			_ = json.NewEncoder(w).Encode(CorrespondentList{Count: 0, Results: []Correspondent{}})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		opts := &ListOptions{
			Page:     1,
			Ordering: "name",
		}
		_, err := c.ListCorrespondents(context.Background(), opts)
		if err != nil {
			t.Fatalf("ListCorrespondents failed: %v", err)
		}
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("Unauthorized"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		_, err := c.ListCorrespondents(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		apiErr, ok := err.(*Error)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}
		if apiErr.Op != "ListCorrespondents" {
			t.Errorf("op = %v, want ListCorrespondents", apiErr.Op)
		}
		if apiErr.StatusCode != 401 {
			t.Errorf("status code = %d, want 401", apiErr.StatusCode)
		}
	})
}

func TestClient_GetCorrespondent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expectedCorrespondent := Correspondent{
			ID:                1,
			Name:              "SheerValue Property Management",
			Slug:              "sheervalue-property-management",
			Match:             "SheerValue",
			MatchingAlgorithm: 1,
			IsInsensitive:     true,
			DocumentCount:     45,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/correspondents/1/" {
				t.Errorf("path = %v, want /api/correspondents/1/", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(expectedCorrespondent)
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		correspondent, err := c.GetCorrespondent(context.Background(), 1)
		if err != nil {
			t.Fatalf("GetCorrespondent failed: %v", err)
		}
		if correspondent.ID != expectedCorrespondent.ID {
			t.Errorf("ID = %d, want %d", correspondent.ID, expectedCorrespondent.ID)
		}
		if correspondent.Name != expectedCorrespondent.Name {
			t.Errorf("Name = %v, want %v", correspondent.Name, expectedCorrespondent.Name)
		}
		if correspondent.Slug != expectedCorrespondent.Slug {
			t.Errorf("Slug = %v, want %v", correspondent.Slug, expectedCorrespondent.Slug)
		}
		if correspondent.Match != expectedCorrespondent.Match {
			t.Errorf("Match = %v, want %v", correspondent.Match, expectedCorrespondent.Match)
		}
		if correspondent.MatchingAlgorithm != expectedCorrespondent.MatchingAlgorithm {
			t.Errorf("MatchingAlgorithm = %d, want %d", correspondent.MatchingAlgorithm, expectedCorrespondent.MatchingAlgorithm)
		}
		if correspondent.IsInsensitive != expectedCorrespondent.IsInsensitive {
			t.Errorf("IsInsensitive = %v, want %v", correspondent.IsInsensitive, expectedCorrespondent.IsInsensitive)
		}
		if correspondent.DocumentCount != expectedCorrespondent.DocumentCount {
			t.Errorf("DocumentCount = %d, want %d", correspondent.DocumentCount, expectedCorrespondent.DocumentCount)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not Found"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		_, err := c.GetCorrespondent(context.Background(), 999)
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
		if apiErr.Op != "GetCorrespondent" {
			t.Errorf("op = %v, want GetCorrespondent", apiErr.Op)
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
		_, err := c.GetCorrespondent(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
