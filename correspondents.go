package paperless

import (
	"context"
	"fmt"
)

// ListCorrespondents retrieves all correspondents with optional filtering.
// Correspondents represent entities like property managers, banks, insurance companies, etc.
func (c *Client) ListCorrespondents(ctx context.Context, opts *ListOptions) (*CorrespondentList, error) {
	fullURL, err := c.buildURL(correspondentsAPIPath, opts)
	if err != nil {
		return nil, fmt.Errorf("build URL: %w", err)
	}

	var result CorrespondentList
	if err := c.doRequestWithURL(ctx, "GET", fullURL, nil, &result); err != nil {
		return nil, wrapError(err, "ListCorrespondents")
	}

	return &result, nil
}

// GetCorrespondent retrieves a single correspondent by ID.
func (c *Client) GetCorrespondent(ctx context.Context, id int) (*Correspondent, error) {
	path := fmt.Sprintf("/api/correspondents/%d/", id)

	var result Correspondent
	if err := c.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, wrapError(err, "GetCorrespondent")
	}

	return &result, nil
}
