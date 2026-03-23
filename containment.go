package toks

import "context"

// Kill activates a kill switch. Requires an admin key.
func (c *Client) Kill(ctx context.Context, req *KillRequest) (*KillResponse, error) {
	var resp KillResponse
	if err := c.do(ctx, "POST", "/containment/kill", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Release deactivates a kill switch. Requires an admin key.
func (c *Client) Release(ctx context.Context, req *ReleaseRequest) error {
	return c.do(ctx, "POST", "/containment/release", req, nil)
}

// ContainmentStatus lists active kill switches for a tenant.
// If tenantID is empty, all active switches are returned.
// Requires an admin key.
func (c *Client) ContainmentStatus(ctx context.Context, tenantID string) (*ContainmentStatusResponse, error) {
	path := "/containment/status"
	if tenantID != "" {
		path += "?tenant_id=" + tenantID
	}
	var resp ContainmentStatusResponse
	if err := c.do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
