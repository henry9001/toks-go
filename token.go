package toks

import (
	"context"
	"net/url"
)

// IssueToken requests a JIT privilege token based on a policy evaluation.
func (c *Client) IssueToken(ctx context.Context, req *PolicyRequest) (*TokenIssueResponse, error) {
	var resp TokenIssueResponse
	if err := c.do(ctx, "POST", "/token/issue", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// VerifyToken checks whether a token is valid and returns its claims.
func (c *Client) VerifyToken(ctx context.Context, token string) (*TokenVerifyResponse, error) {
	var resp TokenVerifyResponse
	if err := c.do(ctx, "POST", "/token/verify", &TokenVerifyRequest{Token: token}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RevokeToken revokes a token so it can no longer be used.
func (c *Client) RevokeToken(ctx context.Context, token string) error {
	return c.do(ctx, "POST", "/token/revoke", &TokenRevokeRequest{Token: token}, nil)
}

// DelegateToken creates a child token scoped under an existing parent token.
func (c *Client) DelegateToken(ctx context.Context, req *DelegateRequest) (*DelegateResponse, error) {
	var resp DelegateResponse
	if err := c.do(ctx, "POST", "/token/delegate", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// TokenChain retrieves the delegation chain for a token identified by its hash.
func (c *Client) TokenChain(ctx context.Context, tokenHash string) (*TokenChainResponse, error) {
	q := url.Values{}
	q.Set("token_hash", tokenHash)
	var resp TokenChainResponse
	if err := c.do(ctx, "GET", "/token/chain?"+q.Encode(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
