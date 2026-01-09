package paperless

import (
	"context"
	"fmt"
)

// ListDocuments retrieves documents with optional filtering.
func (c *Client) ListDocuments(ctx context.Context, opts *ListOptions) (*DocumentList, error) {
	fullURL, err := c.buildURL(documentsAPIPath, opts)
	if err != nil {
		return nil, fmt.Errorf("build URL: %w", err)
	}

	var result DocumentList
	if err := c.doRequestWithURL(ctx, "GET", fullURL, nil, &result); err != nil {
		return nil, wrapError(err, "ListDocuments")
	}

	return &result, nil
}

// GetDocument retrieves a single document by ID.
func (c *Client) GetDocument(ctx context.Context, id int) (*Document, error) {
	path := fmt.Sprintf("/api/documents/%d/", id)

	var result Document
	if err := c.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, wrapError(err, "GetDocument")
	}

	return &result, nil
}

// UpdateDocument updates a document.
func (c *Client) UpdateDocument(ctx context.Context, id int, update *DocumentUpdate) (*Document, error) {
	path := fmt.Sprintf("/api/documents/%d/", id)

	var result Document
	if err := c.doRequest(ctx, "PATCH", path, update, &result); err != nil {
		return nil, wrapError(err, "UpdateDocument")
	}

	return &result, nil
}

// RenameDocument renames a document by updating its title.
// This is a convenience wrapper around UpdateDocument that only updates the title field.
// Returns an error if the new title is empty or if the document ID is invalid.
func (c *Client) RenameDocument(ctx context.Context, id int, newTitle string) (*Document, error) {
	if id <= 0 {
		return nil, fmt.Errorf("RenameDocument: invalid document ID: %d", id)
	}
	if newTitle == "" {
		return nil, fmt.Errorf("RenameDocument: title cannot be empty")
	}

	update := &DocumentUpdate{
		Title: &newTitle,
	}

	doc, err := c.UpdateDocument(ctx, id, update)
	if err != nil {
		return nil, wrapError(err, "RenameDocument")
	}

	return doc, nil
}

// UpdateDocumentTags updates the tags for a document.
// This is a convenience wrapper around UpdateDocument that only updates the tags field.
// Pass an empty slice to remove all tags from the document.
// Returns an error if the document ID is invalid or if any tag IDs are invalid.
func (c *Client) UpdateDocumentTags(ctx context.Context, id int, tagIDs []int) (*Document, error) {
	if id <= 0 {
		return nil, fmt.Errorf("UpdateDocumentTags: invalid document ID: %d", id)
	}

	// Validate tag IDs
	for i, tagID := range tagIDs {
		if tagID <= 0 {
			return nil, fmt.Errorf("UpdateDocumentTags: invalid tag ID at index %d: %d", i, tagID)
		}
	}

	// Ensure we have a non-nil slice (empty slice is valid - it removes all tags)
	if tagIDs == nil {
		tagIDs = []int{}
	}

	update := &DocumentUpdate{
		Tags: &tagIDs,
	}

	doc, err := c.UpdateDocument(ctx, id, update)
	if err != nil {
		return nil, wrapError(err, "UpdateDocumentTags")
	}

	return doc, nil
}
