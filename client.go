package toks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client is a TOKS API client.
type Client struct {
	apiKey    string
	adminKey  string
	baseURL   string
	userAgent string
	http      *http.Client
}

// New creates a Client authenticated with the given API key.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    http.DefaultClient,
	}
	for _, o := range opts {
		o(c)
	}
	c.baseURL = strings.TrimRight(c.baseURL, "/")
	return c
}

// do executes an HTTP request and decodes the JSON response into dst.
// If the server returns a non-2xx status, an *APIError is returned.
func (c *Client) do(ctx context.Context, method, path string, body, dst interface{}) error {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("toks: marshal request: %w", err)
		}
		reqBody = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("toks: create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	if c.adminKey != "" {
		req.Header.Set("X-Admin-Key", c.adminKey)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("toks: request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("toks: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		_ = json.Unmarshal(data, apiErr)
		if apiErr.Code == "" {
			apiErr.Code = http.StatusText(resp.StatusCode)
		}
		return apiErr
	}

	if dst != nil && len(data) > 0 {
		if err := json.Unmarshal(data, dst); err != nil {
			return fmt.Errorf("toks: decode response: %w", err)
		}
	}
	return nil
}
