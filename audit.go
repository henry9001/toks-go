package toks

import (
	"context"
	"net/url"
	"time"
)

// ListAuditEvents returns recent audit log entries.
// Requires viewer-level access or above.
func (c *Client) ListAuditEvents(ctx context.Context) (*AuditEventsResponse, error) {
	var resp AuditEventsResponse
	if err := c.do(ctx, "GET", "/audit/events", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AttributionOptions configures the attribution query.
type AttributionOptions struct {
	User  string    // Required: the user to query attribution for.
	Since time.Time // Optional: only return records after this time.
}

// QueryAttribution returns a cross-entity audit trail for a user.
// Requires viewer-level access or above.
func (c *Client) QueryAttribution(ctx context.Context, opts *AttributionOptions) (*AttributionResponse, error) {
	v := url.Values{}
	v.Set("user", opts.User)
	if !opts.Since.IsZero() {
		v.Set("since", opts.Since.Format(time.RFC3339))
	}
	var resp AttributionResponse
	if err := c.do(ctx, "GET", "/audit/attribution?"+v.Encode(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
