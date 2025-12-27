package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is an HTTP client for an OpenAI-compatible embeddings API.
type Client struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewClient creates a new embeddings client with the provided base URL.
func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

// GenerateEmbedding generates an embedding vector for the given text
func (c *Client) GenerateEmbedding(text string) ([]float32, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return nil, fmt.Errorf("api key is required")
	}
	if strings.TrimSpace(c.baseURL) == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if strings.TrimSpace(c.model) == "" {
		return nil, fmt.Errorf("model is required")
	}

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

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
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
