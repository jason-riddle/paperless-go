package paperless

import (
	"fmt"
	"time"
)

// Date represents a date/timestamp from the Paperless API
type Date time.Time

// UnmarshalJSON implements json.Unmarshaler for both date-only and RFC3339 formats
func (d *Date) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	// Remove quotes
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	// Try RFC3339 format first (full timestamp)
	if parsed, err := time.Parse(time.RFC3339, str); err == nil {
		*d = Date(parsed)
		return nil
	}

	// Try date-only format
	if parsed, err := time.Parse("2006-01-02", str); err == nil {
		*d = Date(parsed)
		return nil
	}

	// Try RFC3339Nano format
	if parsed, err := time.Parse(time.RFC3339Nano, str); err == nil {
		*d = Date(parsed)
		return nil
	}

	return fmt.Errorf("unable to parse date: %s", str)
}

// MarshalJSON implements json.Marshaler
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(d).Format("2006-01-02") + `"`), nil
}

// Time returns the underlying time.Time
func (d Date) Time() time.Time {
	return time.Time(d)
}

// String returns the date as a string
func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}

// Document represents a Paperless-ngx document.
type Document struct {
	ID                  int    `json:"id"`
	Title               string `json:"title"`
	Content             string `json:"content"`
	Created             Date   `json:"created"`
	Modified            Date   `json:"modified"`
	Added               Date   `json:"added"`
	ArchiveSerialNumber *int   `json:"archive_serial_number"`
	OriginalFileName    string `json:"original_file_name"`
	Tags                []int  `json:"tags"`
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
