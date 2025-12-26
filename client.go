package paperless

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client is a Paperless-ngx API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(client *Client) {
		client.httpClient = c
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(client *Client) {
		client.httpClient.Timeout = d
	}
}

// NewClient creates a new Paperless-ngx API client.
// baseURL is the Paperless instance URL (e.g., "http://localhost:8000").
// token is the API authentication token.
func NewClient(baseURL, token string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// doRequest performs an HTTP request and decodes the JSON response.
func (c *Client) doRequest(ctx context.Context, method, path string, result interface{}) error {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path

	return c.doRequestWithURL(ctx, method, u.String(), result)
}

// wrapError wraps an error with an operation name if it's an API error.
func wrapError(err error, op string) error {
	if err == nil {
		return nil
	}
	apiErr, ok := err.(*Error)
	if ok {
		apiErr.Op = op
		return apiErr
	}
	return fmt.Errorf("%s: %w", op, err)
}

// buildURL constructs a URL with query parameters from ListOptions.
func (c *Client) buildURL(path string, opts *ListOptions) (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path

	if opts != nil {
		q := u.Query()
		if opts.Page > 0 {
			q.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			q.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.Query != "" {
			if opts.TitleOnly {
				q.Set("title__icontains", opts.Query)
			} else {
				q.Set("query", opts.Query)
			}
		}
		if opts.Ordering != "" {
			q.Set("ordering", opts.Ordering)
		}
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

// doRequestWithURL performs an HTTP request using a full URL and decodes the JSON response.
// This is the common helper function used by both doRequest and direct calls.
func (c *Client) doRequestWithURL(ctx context.Context, method, fullURL string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &Error{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
