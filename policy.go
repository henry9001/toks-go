package toks

import (
	"context"
	"encoding/json"
	"fmt"
)

// CheckPolicy evaluates a policy request and returns the server's decision.
// Requires any authenticated role.
func (c *Client) CheckPolicy(ctx context.Context, req *PolicyRequest) (*PolicyDecision, error) {
	var resp PolicyDecision
	if err := c.do(ctx, "POST", "/policy/check", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListPolicyVersions returns all policy rule versions for the caller's tenant.
// Requires operator-level access or above.
func (c *Client) ListPolicyVersions(ctx context.Context) (*PolicyVersionListResponse, error) {
	var resp PolicyVersionListResponse
	if err := c.do(ctx, "GET", "/policy/rules", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePolicyVersion uploads a new policy rule set. The new version is
// inactive until explicitly activated with ActivatePolicyVersion.
// Requires operator-level access or above.
func (c *Client) CreatePolicyVersion(ctx context.Context, rules json.RawMessage) (*PolicyVersionResponse, error) {
	req := &CreatePolicyVersionRequest{Rules: rules}
	var resp PolicyVersionResponse
	if err := c.do(ctx, "POST", "/policy/rules", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ActivatePolicyVersion makes the given version the active policy rule set.
// The version must already exist and belong to the caller's tenant.
// Requires operator-level access or above.
func (c *Client) ActivatePolicyVersion(ctx context.Context, version int) (*PolicyVersionResponse, error) {
	path := fmt.Sprintf("/policy/rules/%d/activate", version)
	var resp PolicyVersionResponse
	if err := c.do(ctx, "POST", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
