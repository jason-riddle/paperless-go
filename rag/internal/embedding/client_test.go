package embedding

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	var client = NewClient("test-key", "test-model")

	if client == nil {
		t.Fatal("Client is nil")
	}

	if client.apiKey != "test-key" {
		t.Errorf("Expected apiKey 'test-key', got '%s'", client.apiKey)
	}

	if client.model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", client.model)
	}

	if client.baseURL != "https://openrouter.ai/api/v1" {
		t.Errorf("Expected baseURL 'https://openrouter.ai/api/v1', got '%s'", client.baseURL)
	}

	if client.client == nil {
		t.Error("HTTP client is nil")
	}
}

func TestNewClientWithBaseURL(t *testing.T) {
	var client = NewClientWithBaseURL("test-key", "test-model", "http://localhost:1234")

	if client == nil {
		t.Fatal("Client is nil")
	}

	if client.apiKey != "test-key" {
		t.Errorf("Expected apiKey 'test-key', got '%s'", client.apiKey)
	}

	if client.model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", client.model)
	}

	if client.baseURL != "http://localhost:1234" {
		t.Errorf("Expected baseURL 'http://localhost:1234', got '%s'", client.baseURL)
	}
}

func TestNewOllamaClient(t *testing.T) {
	var client = NewOllamaClient("http://localhost:11434/v1", "nomic-embed-text")

	if client == nil {
		t.Fatal("Client is nil")
	}

	if client.apiKey != "" {
		t.Errorf("Expected empty apiKey for Ollama, got '%s'", client.apiKey)
	}

	if client.model != "nomic-embed-text" {
		t.Errorf("Expected model 'nomic-embed-text', got '%s'", client.model)
	}

	if client.baseURL != "http://localhost:11434/v1" {
		t.Errorf("Expected baseURL 'http://localhost:11434/v1', got '%s'", client.baseURL)
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

func TestGenerateEmbeddingOllamaNoAuth(t *testing.T) {
	var server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var authHeader = r.Header.Get("Authorization")
		if authHeader != "" {
			t.Errorf("Expected no Authorization header for Ollama, got '%s'", authHeader)
		}

		var response = EmbeddingResponse{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
				Index     int       `json:"index"`
			}{
				{
					Embedding: []float32{0.5, 0.6},
					Index:     0,
				},
			},
			Model: "nomic-embed-text",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	var client = NewOllamaClient(server.URL, "nomic-embed-text")
	var embedding, err = client.GenerateEmbedding("test text")
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}

	if len(embedding) != 2 {
		t.Errorf("Expected 2 dimensions, got %d", len(embedding))
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
