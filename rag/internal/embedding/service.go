package embedding

import (
	"fmt"
	"log/slog"

	"github.com/jason-riddle/paperless-go/rag/internal/metrics"
)

// Service provides embedding generation with additional logic
type Service struct {
	client *Client
}

// NewService creates a new embedding service
func NewService(client *Client) *Service {
	return &Service{
		client: client,
	}
}

// GenerateEmbedding generates an embedding for the given text
func (s *Service) GenerateEmbedding(text string) ([]float32, error) {
	if text == "" {
		metrics.EmbeddingsFailedTotal.Add(1)
		return nil, fmt.Errorf("text cannot be empty")
	}

	slog.Debug("Generating embedding", "text_length", len(text))

	vector, err := s.client.GenerateEmbedding(text)
	if err != nil {
		metrics.EmbeddingsFailedTotal.Add(1)
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	metrics.EmbeddingsGeneratedTotal.Add(1)
	slog.Debug("Generated embedding", "dimensions", len(vector))
	return vector, nil
}

// FormatDocumentText formats a document's title and tags for embedding
func FormatDocumentText(title string, tags string) string {
	if tags == "" {
		return title
	}
	return fmt.Sprintf("%s. Tags: %s", title, tags)
}
