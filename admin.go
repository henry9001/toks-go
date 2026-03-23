package toks

import "context"

// CreateTenant creates a new tenant with an initial API key.
// Requires an admin key.
func (c *Client) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*CreateTenantResponse, error) {
	var resp CreateTenantResponse
	if err := c.do(ctx, "POST", "/api/tenants", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTenant retrieves a tenant by ID. Requires an admin key.
func (c *Client) GetTenant(ctx context.Context, tenantID string) (*CreateTenantResponse, error) {
	var resp CreateTenantResponse
	if err := c.do(ctx, "GET", "/api/tenants/"+tenantID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateAPIKey generates an additional API key for an existing tenant.
// Requires an admin key.
func (c *Client) CreateAPIKey(ctx context.Context, tenantID string, req *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	var resp CreateAPIKeyResponse
	if err := c.do(ctx, "POST", "/api/tenants/"+tenantID+"/keys", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RevokeAPIKey revokes an API key by its hash. Requires an admin key.
func (c *Client) RevokeAPIKey(ctx context.Context, tenantID, keyHash string) error {
	return c.do(ctx, "DELETE", "/api/tenants/"+tenantID+"/keys/"+keyHash, nil, nil)
}

// ListAPIKeys returns all API keys for a tenant. Requires an admin key.
func (c *Client) ListAPIKeys(ctx context.Context, tenantID string) (*APIKeyListResponse, error) {
	var resp APIKeyListResponse
	if err := c.do(ctx, "GET", "/api/tenants/"+tenantID+"/keys", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAPIKey returns details of a single API key. Requires an admin key.
func (c *Client) GetAPIKey(ctx context.Context, tenantID, keyHash string) (*APIKeyInfo, error) {
	var resp APIKeyInfo
	if err := c.do(ctx, "GET", "/api/tenants/"+tenantID+"/keys/"+keyHash, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUsage returns monthly usage stats for a tenant. Requires an admin key.
func (c *Client) GetUsage(ctx context.Context, tenantID string) (*UsageResponse, error) {
	var resp UsageResponse
	if err := c.do(ctx, "GET", "/api/tenants/"+tenantID+"/usage", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateRole changes the RBAC role for an API key. Requires admin-level access.
func (c *Client) UpdateRole(ctx context.Context, keyHash, role string) (*UpdateRoleResponse, error) {
	req := &UpdateRoleRequest{
		KeyHash: keyHash,
		Role:    role,
	}
	var resp UpdateRoleResponse
	if err := c.do(ctx, "PUT", "/admin/roles", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListRoles returns available roles and their permissions. Requires admin-level access.
func (c *Client) ListRoles(ctx context.Context) (*RoleListResponse, error) {
	var resp RoleListResponse
	if err := c.do(ctx, "GET", "/admin/roles", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
