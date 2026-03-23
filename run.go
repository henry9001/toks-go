package toks

import (
	"context"
	"fmt"
	"net/url"
)

// ListRunsOptions configures filtering for ListRuns.
type ListRunsOptions struct {
	Status    string // Filter by run state (e.g. "active").
	AgentType string // Filter by agent type.
	Project   string // Filter by project name.
	Since     string // Start time in RFC 3339 format.
	Until     string // End time in RFC 3339 format.
}

// RegisterRun registers a new run and returns its ID.
func (c *Client) RegisterRun(ctx context.Context, req *RunRegisterRequest) (*RunRegisterResponse, error) {
	var resp RunRegisterResponse
	if err := c.do(ctx, "POST", "/run/register", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRun retrieves a run by ID.
func (c *Client) GetRun(ctx context.Context, runID string) (*Run, error) {
	var resp Run
	if err := c.do(ctx, "GET", "/run/"+runID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListRuns lists runs, optionally filtered by the provided options.
func (c *Client) ListRuns(ctx context.Context, opts *ListRunsOptions) (*RunListResponse, error) {
	path := "/runs"
	if opts != nil {
		q := url.Values{}
		if opts.Status != "" {
			q.Set("status", opts.Status)
		}
		if opts.AgentType != "" {
			q.Set("agent_type", opts.AgentType)
		}
		if opts.Project != "" {
			q.Set("project", opts.Project)
		}
		if opts.Since != "" {
			q.Set("since", opts.Since)
		}
		if opts.Until != "" {
			q.Set("until", opts.Until)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}
	var resp RunListResponse
	if err := c.do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateRunStatus updates the state of a run.
func (c *Client) UpdateRunStatus(ctx context.Context, runID string, req *RunStatusRequest) (*RunStatusResponse, error) {
	var resp RunStatusResponse
	if err := c.do(ctx, "PUT", fmt.Sprintf("/run/%s/status", runID), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
