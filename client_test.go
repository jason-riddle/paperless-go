package paperless

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	baseURL := "http://localhost:8000"
	token := "test-token"

	t.Run("default client", func(t *testing.T) {
		c := NewClient(baseURL, token)
		if c.baseURL != baseURL {
			t.Errorf("baseURL = %v, want %v", c.baseURL, baseURL)
		}
		if c.token != token {
			t.Errorf("token = %v, want %v", c.token, token)
		}
		if c.httpClient == nil {
			t.Error("httpClient is nil")
		}
		if c.httpClient.Timeout != 30*time.Second {
			t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, 30*time.Second)
		}
	})

	t.Run("with custom HTTP client", func(t *testing.T) {
		customClient := &http.Client{Timeout: 10 * time.Second}
		c := NewClient(baseURL, token, WithHTTPClient(customClient))
		if c.httpClient != customClient {
			t.Error("custom HTTP client not set")
		}
	})

	t.Run("with custom timeout", func(t *testing.T) {
		timeout := 5 * time.Second
		c := NewClient(baseURL, token, WithTimeout(timeout))
		if c.httpClient.Timeout != timeout {
			t.Errorf("timeout = %v, want %v", c.httpClient.Timeout, timeout)
		}
	})
}

func TestClient_doRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Token test-token" {
				t.Error("authorization header not set correctly")
			}
			if r.Header.Get("Accept") != "application/json" {
				t.Error("accept header not set correctly")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		var result map[string]string
		err := c.doRequest(context.Background(), "GET", "/api/test/", &result)
		if err != nil {
			t.Fatalf("doRequest failed: %v", err)
		}
		if result["status"] != "ok" {
			t.Errorf("status = %v, want ok", result["status"])
		}
	})

	t.Run("404 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		err := c.doRequest(context.Background(), "GET", "/api/test/", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected 404 error, got %v", err)
		}
	})

	t.Run("500 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		err := c.doRequest(context.Background(), "GET", "/api/test/", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		apiErr, ok := err.(*Error)
		if !ok {
			t.Fatalf("expected *Error, got %T", err)
		}
		if apiErr.StatusCode != 500 {
			t.Errorf("status code = %d, want 500", apiErr.StatusCode)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		c := NewClient(server.URL, "test-token")
		var result map[string]string
		err := c.doRequest(context.Background(), "GET", "/api/test/", &result)
		if err == nil {
			t.Fatal("expected error, got nil")
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
		err := c.doRequest(ctx, "GET", "/api/test/", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestClient_buildURL(t *testing.T) {
	c := NewClient("http://localhost:8000", "test-token")

	tests := []struct {
		name    string
		path    string
		opts    *ListOptions
		want    string
		wantErr bool
	}{
		{
			name: "no options",
			path: "/api/documents/",
			opts: nil,
			want: "http://localhost:8000/api/documents/",
		},
		{
			name: "with page",
			path: "/api/documents/",
			opts: &ListOptions{Page: 2},
			want: "http://localhost:8000/api/documents/?page=2",
		},
		{
			name: "with page size",
			path: "/api/documents/",
			opts: &ListOptions{PageSize: 50},
			want: "http://localhost:8000/api/documents/?page_size=50",
		},
		{
			name: "with query",
			path: "/api/documents/",
			opts: &ListOptions{Query: "test search"},
			want: "http://localhost:8000/api/documents/?query=test+search",
		},
		{
			name: "with ordering",
			path: "/api/documents/",
			opts: &ListOptions{Ordering: "-created"},
			want: "http://localhost:8000/api/documents/?ordering=-created",
		},
		{
			name: "with all options",
			path: "/api/documents/",
			opts: &ListOptions{
				Page:     2,
				PageSize: 50,
				Query:    "test",
				Ordering: "-created",
			},
			want: "http://localhost:8000/api/documents/?ordering=-created&page=2&page_size=50&query=test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.buildURL(tt.path, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
