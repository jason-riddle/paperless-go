package paperless

import (
	"context"
	"fmt"
)

// ListTags retrieves all tags.
func (c *Client) ListTags(ctx context.Context, opts *ListOptions) (*TagList, error) {
	fullURL, err := c.buildURL(tagsAPIPath, opts)
	if err != nil {
		return nil, fmt.Errorf("build URL: %w", err)
	}

	var result TagList
	if err := c.doRequestWithURL(ctx, "GET", fullURL, nil, &result); err != nil {
		return nil, wrapError(err, "ListTags")
	}

	return &result, nil
}

// GetTag retrieves a single tag by ID.
func (c *Client) GetTag(ctx context.Context, id int) (*Tag, error) {
	path := fmt.Sprintf("/api/tags/%d/", id)

	var result Tag
	if err := c.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, wrapError(err, "GetTag")
	}

	return &result, nil
}

// CreateTag creates a new tag.
func (c *Client) CreateTag(ctx context.Context, tag *TagCreate) (*Tag, error) {
	var result Tag
	if err := c.doRequest(ctx, "POST", "/api/tags/", tag, &result); err != nil {
		return nil, wrapError(err, "CreateTag")
	}

	return &result, nil
}
