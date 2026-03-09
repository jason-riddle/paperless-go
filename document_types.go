package paperless

import (
	"context"
	"fmt"
)

// ListDocumentTypes retrieves all document types with optional filtering.
// Document types represent classifications like Invoice, Statement, Receipt, Contract, etc.
func (c *Client) ListDocumentTypes(ctx context.Context, opts *ListOptions) (*DocumentTypeList, error) {
	fullURL, err := c.buildURL(documentTypesAPIPath, opts)
	if err != nil {
		return nil, fmt.Errorf("build URL: %w", err)
	}

	var result DocumentTypeList
	if err := c.doRequestWithURL(ctx, "GET", fullURL, nil, &result); err != nil {
		return nil, wrapError(err, "ListDocumentTypes")
	}

	return &result, nil
}

// GetDocumentType retrieves a single document type by ID.
func (c *Client) GetDocumentType(ctx context.Context, id int) (*DocumentType, error) {
	path := fmt.Sprintf("/api/document_types/%d/", id)

	var result DocumentType
	if err := c.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, wrapError(err, "GetDocumentType")
	}

	return &result, nil
}
