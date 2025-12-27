package embedding

import (
	"encoding/json"
	"testing"
)

func TestEmbeddingRequestSerialization(t *testing.T) {
	var req = EmbeddingRequest{
		Model: "openai/text-embedding-3-small",
		Input: "test text",
	}

	var jsonData, err = json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var decoded EmbeddingRequest
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if decoded.Model != req.Model {
		t.Errorf("Model mismatch: expected %s, got %s", req.Model, decoded.Model)
	}
	if decoded.Input != req.Input {
		t.Errorf("Input mismatch: expected %s, got %s", req.Input, decoded.Input)
	}
}

func TestEmbeddingResponseDeserialization(t *testing.T) {
	var jsonResponse = `{
		"data": [
			{
				"embedding": [0.1, 0.2, 0.3],
				"index": 0
			}
		],
		"model": "openai/text-embedding-3-small",
		"usage": {
			"prompt_tokens": 5,
			"total_tokens": 5
		}
	}`

	var resp EmbeddingResponse
	var err = json.Unmarshal([]byte(jsonResponse), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("Expected 1 data item, got %d", len(resp.Data))
	}

	if len(resp.Data[0].Embedding) != 3 {
		t.Errorf("Expected 3 embedding values, got %d", len(resp.Data[0].Embedding))
	}

	if resp.Data[0].Embedding[0] != 0.1 {
		t.Errorf("Expected first embedding value 0.1, got %f", resp.Data[0].Embedding[0])
	}

	if resp.Model != "openai/text-embedding-3-small" {
		t.Errorf("Model mismatch: expected openai/text-embedding-3-small, got %s", resp.Model)
	}

	if resp.Usage.PromptTokens != 5 {
		t.Errorf("Expected 5 prompt tokens, got %d", resp.Usage.PromptTokens)
	}
}

func TestErrorResponseDeserialization(t *testing.T) {
	var jsonError = `{
		"error": {
			"message": "Invalid API key",
			"type": "invalid_request_error",
			"code": "invalid_api_key"
		}
	}`

	var errResp ErrorResponse
	var err = json.Unmarshal([]byte(jsonError), &errResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if errResp.Error.Message != "Invalid API key" {
		t.Errorf("Message mismatch: expected 'Invalid API key', got '%s'", errResp.Error.Message)
	}

	if errResp.Error.Type != "invalid_request_error" {
		t.Errorf("Type mismatch: expected 'invalid_request_error', got '%s'", errResp.Error.Type)
	}

	if errResp.Error.Code != "invalid_api_key" {
		t.Errorf("Code mismatch: expected 'invalid_api_key', got '%s'", errResp.Error.Code)
	}
}

func TestEmbeddingResponseEmptyData(t *testing.T) {
	var jsonResponse = `{
		"data": [],
		"model": "openai/text-embedding-3-small",
		"usage": {
			"prompt_tokens": 0,
			"total_tokens": 0
		}
	}`

	var resp EmbeddingResponse
	var err = json.Unmarshal([]byte(jsonResponse), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(resp.Data) != 0 {
		t.Errorf("Expected 0 data items, got %d", len(resp.Data))
	}
}

func TestEmbeddingResponseMultipleEmbeddings(t *testing.T) {
	var jsonResponse = `{
		"data": [
			{
				"embedding": [0.1, 0.2],
				"index": 0
			},
			{
				"embedding": [0.3, 0.4],
				"index": 1
			}
		],
		"model": "test-model",
		"usage": {
			"prompt_tokens": 10,
			"total_tokens": 10
		}
	}`

	var resp EmbeddingResponse
	var err = json.Unmarshal([]byte(jsonResponse), &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("Expected 2 data items, got %d", len(resp.Data))
	}

	if resp.Data[0].Index != 0 {
		t.Errorf("Expected index 0, got %d", resp.Data[0].Index)
	}

	if resp.Data[1].Index != 1 {
		t.Errorf("Expected index 1, got %d", resp.Data[1].Index)
	}

	if len(resp.Data[1].Embedding) != 2 {
		t.Errorf("Expected 2 embedding values, got %d", len(resp.Data[1].Embedding))
	}

	if resp.Data[1].Embedding[0] != 0.3 {
		t.Errorf("Expected embedding value 0.3, got %f", resp.Data[1].Embedding[0])
	}
}
