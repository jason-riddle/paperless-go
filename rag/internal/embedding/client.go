package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for OpenAI-compatible embeddings API (OpenRouter or Ollama)
type Client struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewClient creates a new embeddings client for OpenRouter
func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://openrouter.ai/api/v1",
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

// NewOllamaClient creates a new embeddings client for Ollama
func NewOllamaClient(baseURL, model string) *Client {
	return &Client{
		apiKey:  "", // Ollama doesn't require API key
		model:   model,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

// GenerateEmbedding generates an embedding vector for the given text
func (c *Client) GenerateEmbedding(text string) ([]float32, error) {
	// Prepare request body
	reqBody := EmbeddingRequest{
		Model: c.model,
		Input: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	embeddingsURL := c.baseURL + "/embeddings"
	req, err := http.NewRequest("POST", embeddingsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header only if API key is provided (not needed for Ollama)
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request with retry logic
	var resp *http.Response
	var lastErr error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		resp, lastErr = c.client.Do(req)

		// Success case
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			break
		}

		// Cleanup response body if we're retrying
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		// Don't sleep after last attempt
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to execute request after %d retries: %w", maxRetries, lastErr)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var embeddingResp EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return embeddingResp.Data[0].Embedding, nil
}
