package toks

import "context"

// SendHeartbeat records a heartbeat for the given agent and returns its
// updated state. Requires agent-level access or above.
func (c *Client) SendHeartbeat(ctx context.Context, req *HeartbeatRequest) (*Agent, error) {
	var resp Agent
	if err := c.do(ctx, "POST", "/heartbeat", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListAgents returns all known agents. If tenantID is non-empty, results are
// filtered to that tenant. Requires viewer-level access or above.
func (c *Client) ListAgents(ctx context.Context, tenantID string) (*AgentListResponse, error) {
	path := "/api/agents"
	if tenantID != "" {
		path += "?tenant_id=" + tenantID
	}
	var resp AgentListResponse
	if err := c.do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
