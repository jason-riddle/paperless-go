package paperless

import "time"

// Document represents a Paperless-ngx document.
type Document struct {
	ID                  int       `json:"id"`
	Title               string    `json:"title"`
	Content             string    `json:"content"`
	Created             time.Time `json:"created"`
	Modified            time.Time `json:"modified"`
	Added               time.Time `json:"added"`
	ArchiveSerialNumber *int      `json:"archive_serial_number"`
	OriginalFileName    string    `json:"original_file_name"`
	Tags                []int     `json:"tags"`
}

// Tag represents a Paperless-ngx tag.
type Tag struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Color         string `json:"color"`
	DocumentCount int    `json:"document_count"`
}

// List is a paginated response.
type List[T any] struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []T     `json:"results"`
}

// DocumentList is a paginated list of documents.
type DocumentList = List[Document]

// TagList is a paginated list of tags.
type TagList = List[Tag]

// ListOptions configures list operations.
type ListOptions struct {
	Page     int    // Page number (1-indexed), 0 means default
	PageSize int    // Results per page, 0 means default
	Query    string // Full-text search query
	Ordering string // Sort field (prefix with - for descending)
}
