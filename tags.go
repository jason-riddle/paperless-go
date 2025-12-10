package paperless

import (
	"context"
	"fmt"
)

// ListTags retrieves all tags.
func (c *Client) ListTags(ctx context.Context, opts *ListOptions) (*TagList, error) {
	fullURL, err := c.buildURL("/api/tags/", opts)
	if err != nil {
		return nil, fmt.Errorf("build URL: %w", err)
	}

	var result TagList
	if err := c.doRequestWithURL(ctx, "GET", fullURL, &result); err != nil {
		apiErr, ok := err.(*Error)
		if ok {
			apiErr.Op = "ListTags"
			return nil, apiErr
		}
		return nil, fmt.Errorf("ListTags: %w", err)
	}

	return &result, nil
}

// GetTag retrieves a single tag by ID.
func (c *Client) GetTag(ctx context.Context, id int) (*Tag, error) {
	path := fmt.Sprintf("/api/tags/%d/", id)

	var result Tag
	if err := c.doRequest(ctx, "GET", path, &result); err != nil {
		apiErr, ok := err.(*Error)
		if ok {
			apiErr.Op = "GetTag"
			return nil, apiErr
		}
		return nil, fmt.Errorf("GetTag: %w", err)
	}

	return &result, nil
}
