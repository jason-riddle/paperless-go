package paperless

import (
	"context"
	"fmt"
)

// ListDocuments retrieves documents with optional filtering.
func (c *Client) ListDocuments(ctx context.Context, opts *ListOptions) (*DocumentList, error) {
	fullURL, err := c.buildURL("/api/documents/", opts)
	if err != nil {
		return nil, fmt.Errorf("build URL: %w", err)
	}

	var result DocumentList
	if err := c.doRequestWithURL(ctx, "GET", fullURL, &result); err != nil {
		apiErr, ok := err.(*Error)
		if ok {
			apiErr.Op = "ListDocuments"
			return nil, apiErr
		}
		return nil, fmt.Errorf("ListDocuments: %w", err)
	}

	return &result, nil
}

// GetDocument retrieves a single document by ID.
func (c *Client) GetDocument(ctx context.Context, id int) (*Document, error) {
	path := fmt.Sprintf("/api/documents/%d/", id)

	var result Document
	if err := c.doRequest(ctx, "GET", path, &result); err != nil {
		apiErr, ok := err.(*Error)
		if ok {
			apiErr.Op = "GetDocument"
			return nil, apiErr
		}
		return nil, fmt.Errorf("GetDocument: %w", err)
	}

	return &result, nil
}
