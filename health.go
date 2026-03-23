package toks

import "context"

// Health returns the server's health status. No authentication required.
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	var resp HealthResponse
	if err := c.do(ctx, "GET", "/health", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
