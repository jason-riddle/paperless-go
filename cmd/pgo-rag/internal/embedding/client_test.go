package embedding

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestNewClient(t *testing.T) {
	var client = NewClient("http://localhost:9999", "test-key", "test-model")

	if client == nil {
		t.Fatal("Client is nil")
	}

	if client.apiKey != "test-key" {
		t.Errorf("Expected apiKey 'test-key', got '%s'", client.apiKey)
	}

	if client.model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", client.model)
	}

	if client.baseURL != "http://localhost:9999" {
		t.Errorf("Expected baseURL 'http://localhost:9999', got '%s'", client.baseURL)
	}

	if client.client == nil {
		t.Error("HTTP client is nil")
	}
}

func TestGenerateEmbeddingSuccess(t *testing.T) {
	var server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/embeddings" {
			t.Errorf("Expected path /embeddings, got %s", r.URL.Path)
		}

		var authHeader = r.Header.Get("Authorization")
		if authHeader != "Bearer test-key" {
			t.Errorf("Expected Authorization header 'Bearer test-key', got '%s'", authHeader)
		}

		var contentType = r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		var response = EmbeddingResponse{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{
					Embedding: []float32{0.1, 0.2, 0.3},
					Index:     0,
				},
			},
			Model: "test-model",
			Usage: struct {
				PromptTokens int `json:"prompt_tokens"`
				TotalTokens  int `json:"total_tokens"`
			}{
				PromptTokens: 5,
				TotalTokens:  5,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	var client = &Client{
		apiKey:  "test-key",
		model:   "test-model",
		baseURL: server.URL,
		client:  &http.Client{},
	}

	var embedding, err = client.GenerateEmbedding("test text")
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}

	if len(embedding) != 3 {
		t.Errorf("Expected 3 dimensions, got %d", len(embedding))
	}

	if embedding[0] != 0.1 {
		t.Errorf("Expected first value 0.1, got %f", embedding[0])
	}
}

func TestGenerateEmbeddingRetriesWithBody(t *testing.T) {
	var requests int32
	var server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		if len(body) == 0 {
			t.Fatal("Expected non-empty request body")
		}

		var req EmbeddingRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to decode request JSON: %v", err)
		}
		if req.Model == "" || req.Input == "" {
			t.Fatalf("Expected model and input in request, got model=%q input=%q", req.Model, req.Input)
		}

		if atomic.LoadInt32(&requests) == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			var errResp = ErrorResponse{
				Error: struct {
					Message string `json:"message"`
					Type    string `json:"type"`
					Code    string `json:"code"`
				}{
					Message: "temporary error",
					Type:    "server_error",
					Code:    "temporary",
				},
			}
			json.NewEncoder(w).Encode(errResp)
			return
		}

		var response = EmbeddingResponse{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{
					Embedding: []float32{0.1, 0.2, 0.3},
					Index:     0,
				},
			},
			Model: "test-model",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	var client = &Client{
		apiKey:  "test-key",
		model:   "test-model",
		baseURL: server.URL,
		client:  &http.Client{},
	}

	var embedding, err = client.GenerateEmbedding("test text")
	if err != nil {
		t.Fatalf("Failed to generate embedding after retry: %v", err)
	}
	if len(embedding) != 3 {
		t.Fatalf("Expected 3 dimensions, got %d", len(embedding))
	}
	if atomic.LoadInt32(&requests) != 2 {
		t.Fatalf("Expected 2 requests, got %d", requests)
	}
}

func TestGenerateEmbeddingAPIError(t *testing.T) {
	var server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		var errResp = ErrorResponse{
			Error: struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			}{
				Message: "Invalid API key",
				Type:    "invalid_request_error",
				Code:    "invalid_api_key",
			},
		}
		json.NewEncoder(w).Encode(errResp)
	}))
	defer server.Close()

	var client = &Client{
		apiKey:  "invalid-key",
		model:   "test-model",
		baseURL: server.URL,
		client:  &http.Client{},
	}

	var _, err = client.GenerateEmbedding("test text")
	if err == nil {
		t.Error("Expected error for invalid API key, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("Expected API error message to include server message, got: %v", err)
	}
}

func TestGenerateEmbeddingEmptyResponse(t *testing.T) {
	var server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var response = EmbeddingResponse{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
				Index     int       `json:"index"`
			}{},
			Model: "test-model",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	var client = &Client{
		apiKey:  "test-key",
		model:   "test-model",
		baseURL: server.URL,
		client:  &http.Client{},
	}

	var _, err = client.GenerateEmbedding("test text")
	if err == nil {
		t.Error("Expected error for empty response data, got nil")
	}
}

func TestGenerateEmbeddingMissingConfig(t *testing.T) {
	client := &Client{
		apiKey:  "",
		model:   "model",
		baseURL: "http://localhost",
		client:  &http.Client{},
	}

	if _, err := client.GenerateEmbedding("test"); err == nil {
		t.Fatalf("expected error for missing api key")
	}

	client.apiKey = "key"
	client.baseURL = ""
	if _, err := client.GenerateEmbedding("test"); err == nil {
		t.Fatalf("expected error for missing base URL")
	}

	client.baseURL = "http://localhost"
	client.model = ""
	if _, err := client.GenerateEmbedding("test"); err == nil {
		t.Fatalf("expected error for missing model")
	}
}

func TestGenerateEmbeddingInvalidJSON(t *testing.T) {
	var server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	var client = &Client{
		apiKey:  "test-key",
		model:   "test-model",
		baseURL: server.URL,
		client:  &http.Client{},
	}

	var _, err = client.GenerateEmbedding("test text")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
