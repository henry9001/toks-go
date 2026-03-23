package toks

import "net/http"

const defaultBaseURL = "https://api.toks.dev"

// Option configures a Client.
type Option func(*Client)

// WithBaseURL sets the API base URL. Default: https://api.toks.dev
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.http = hc }
}

// WithAdminKey sets the X-Admin-Key header for admin endpoints.
func WithAdminKey(key string) Option {
	return func(c *Client) { c.adminKey = key }
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}
