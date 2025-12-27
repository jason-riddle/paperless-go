package embedding

import (
	"testing"
)

func TestNewService(t *testing.T) {
	var client = NewClient("http://localhost:9999", "test-key", "test-model")
	var service = NewService(client)

	if service == nil {
		t.Fatal("Service is nil")
	}

	if service.client == nil {
		t.Error("Service client is nil")
	}
}

func TestServiceGenerateEmbeddingEmptyText(t *testing.T) {
	var client = NewClient("http://localhost:9999", "test-key", "test-model")
	var service = NewService(client)

	var _, err = service.GenerateEmbedding("")
	if err == nil {
		t.Error("Expected error for empty text, got nil")
	}

	var expectedMsg = "text cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestFormatDocumentText(t *testing.T) {
	var tests = []struct {
		name     string
		title    string
		tags     string
		expected string
	}{
		{
			name:     "title with tags",
			title:    "Financial Report",
			tags:     "finance, report",
			expected: "Financial Report. Tags: finance, report",
		},
		{
			name:     "title without tags",
			title:    "Simple Document",
			tags:     "",
			expected: "Simple Document",
		},
		{
			name:     "empty title with tags",
			title:    "",
			tags:     "tag1, tag2",
			expected: ". Tags: tag1, tag2",
		},
		{
			name:     "both empty",
			title:    "",
			tags:     "",
			expected: "",
		},
		{
			name:     "single tag",
			title:    "Invoice",
			tags:     "invoice",
			expected: "Invoice. Tags: invoice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result = FormatDocumentText(tt.title, tt.tags)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
