package storage

import "time"

// Document represents a Paperless document in the database
type Document struct {
	ID           int       `json:"id"`
	PaperlessID  int       `json:"paperless_id"`
	PaperlessURL string    `json:"paperless_url"`
	Title        string    `json:"title"`
	Tags         string    `json:"tags"`
	EmbeddedAt   time.Time `json:"embedded_at"`
	LastModified time.Time `json:"last_modified"`
}

// Embedding represents a vector embedding for a document
type Embedding struct {
	ID         int       `json:"id"`
	DocumentID int       `json:"document_id"`
	Content    string    `json:"content"`
	Vector     []float32 `json:"vector"`
	CreatedAt  time.Time `json:"created_at"`
}

// SearchResult represents a search result with similarity score
type SearchResult struct {
	DocumentID      int       `json:"document_id"`
	PaperlessURL    string    `json:"paperless_url"`
	Title           string    `json:"title"`
	Tags            string    `json:"tags"`
	SimilarityScore float64   `json:"similarity_score"`
	LastModified    time.Time `json:"last_modified"`
}
