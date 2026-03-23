package toks

import (
	"encoding/json"
	"time"
)

// ---------------------------------------------------------------------------
// Policy
// ---------------------------------------------------------------------------

// PolicyRequest is the body for POST /policy/check and POST /token/issue.
type PolicyRequest struct {
	AgentType       string `json:"agent_type"`
	TenantID        string `json:"tenant_id"`
	RequestedAction string `json:"requested_action"`
	Resource        string `json:"resource"`
	RunID           string `json:"run_id"`
}

// PolicyDecision is the server's evaluation of a policy request.
type PolicyDecision struct {
	Allowed    bool   `json:"allowed"`
	Reason     string `json:"reason"`
	Scope      string `json:"scope"`
	TTLMinutes int    `json:"ttl_minutes"`
}

// ---------------------------------------------------------------------------
// Token
// ---------------------------------------------------------------------------

// TokenResponse is returned by POST /token/issue.
type TokenResponse struct {
	Token      string `json:"token"`
	ExpiresAt  int64  `json:"expires_at"`
	TTLMinutes int    `json:"ttl_minutes"`
}

// TokenIssueResponse wraps both the decision and the issued token.
type TokenIssueResponse struct {
	Decision PolicyDecision `json:"decision"`
	Token    TokenResponse  `json:"token"`
}

// TokenVerifyRequest is the body for POST /token/verify.
type TokenVerifyRequest struct {
	Token string `json:"token"`
}

// TokenVerifyResponse is returned by POST /token/verify.
type TokenVerifyResponse struct {
	Valid  bool                   `json:"valid"`
	Reason string                 `json:"reason,omitempty"`
	Claims map[string]interface{} `json:"claims,omitempty"`
}

// TokenRevokeRequest is the body for POST /token/revoke.
type TokenRevokeRequest struct {
	Token string `json:"token"`
}

// ---------------------------------------------------------------------------
// Delegation
// ---------------------------------------------------------------------------

// DelegateRequest is the body for POST /token/delegate.
type DelegateRequest struct {
	ParentToken     string `json:"parent_token"`
	ChildAgentType  string `json:"child_agent_type"`
	ChildAction     string `json:"child_action"`
	ChildResource   string `json:"child_resource"`
	ChildTTLMinutes int    `json:"child_ttl_minutes"`
}

// DelegateResponse is returned by POST /token/delegate.
type DelegateResponse struct {
	Token      string `json:"token"`
	ExpiresAt  int64  `json:"expires_at"`
	TTLMinutes int    `json:"ttl_minutes"`
	Depth      int    `json:"depth"`
	ParentHash string `json:"parent_hash"`
}

// TokenChainResponse is returned by GET /token/chain.
type TokenChainResponse struct {
	TokenHash string                   `json:"token_hash"`
	Chain     []map[string]interface{} `json:"chain"`
	Depth     int                      `json:"depth"`
}

// ---------------------------------------------------------------------------
// Heartbeat
// ---------------------------------------------------------------------------

// HeartbeatRequest is the body for POST /heartbeat.
type HeartbeatRequest struct {
	AgentID  string `json:"agent_id"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	Session  string `json:"session"`
}

// Agent represents an agent's heartbeat state.
type Agent struct {
	AgentID          string    `json:"agent_id"`
	TenantID         string    `json:"tenant_id"`
	Role             string    `json:"role"`
	Session          string    `json:"session"`
	State            string    `json:"state"` // "healthy", "stale", "stuck"
	LastSeen         time.Time `json:"last_seen"`
	ActiveTokenCount int       `json:"active_token_count"`
}

// AgentListResponse is returned by GET /api/agents.
type AgentListResponse struct {
	Agents []Agent `json:"agents"`
}

// ---------------------------------------------------------------------------
// Run
// ---------------------------------------------------------------------------

// RunRegisterRequest is the body for POST /run/register.
type RunRegisterRequest struct {
	InitiatedBy     string `json:"initiated_by,omitempty"`
	InitiatedByType string `json:"initiated_by_type,omitempty"` // "user", "agent", "system"
}

// RunRegisterResponse is returned by POST /run/register.
type RunRegisterResponse struct {
	RunID           string `json:"run_id"`
	InitiatedBy     string `json:"initiated_by"`
	InitiatedByType string `json:"initiated_by_type"`
}

// RunStatusRequest is the body for PUT /run/{id}/status.
type RunStatusRequest struct {
	State  string `json:"state"` // "completed", "failed", "expired"
	Reason string `json:"reason,omitempty"`
}

// RunStatusResponse is returned by PUT /run/{id}/status.
type RunStatusResponse struct {
	RunID string `json:"run_id"`
	State string `json:"state"`
}

// Run represents a run record.
type Run struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	AgentID         string     `json:"agent_id"`
	RegisteredAt    time.Time  `json:"registered_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	State           string     `json:"state"`
	InitiatedBy     string     `json:"initiated_by"`
	InitiatedByType string     `json:"initiated_by_type"`
}

// RunListResponse is returned by GET /runs.
type RunListResponse struct {
	Runs  []Run `json:"runs"`
	Count int   `json:"count"`
}

// ---------------------------------------------------------------------------
// Events
// ---------------------------------------------------------------------------

// IngestEventRequest is the body for POST /events/ingest.
type IngestEventRequest struct {
	AgentID   string          `json:"agent_id"`
	EventType string          `json:"event_type"`
	RunID     string          `json:"run_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

// IngestEventResponse is returned by POST /events/ingest.
type IngestEventResponse struct {
	EventID int64 `json:"event_id"`
}

// AuditEventsResponse is returned by GET /audit/events.
type AuditEventsResponse struct {
	Events []map[string]interface{} `json:"events"`
}

// ---------------------------------------------------------------------------
// Attribution
// ---------------------------------------------------------------------------

// AttributionRecord represents a single entry in the attribution trail.
type AttributionRecord struct {
	EntityType      string    `json:"entity_type"` // "run", "event", "token"
	EntityID        string    `json:"entity_id"`
	InitiatedBy     string    `json:"initiated_by"`
	InitiatedByType string    `json:"initiated_by_type"`
	Timestamp       time.Time `json:"timestamp"`
	Detail          string    `json:"detail"`
}

// AttributionResponse is returned by GET /audit/attribution.
type AttributionResponse struct {
	User    string              `json:"user"`
	Records []AttributionRecord `json:"records"`
	Count   int                 `json:"count"`
}

// ---------------------------------------------------------------------------
// Containment
// ---------------------------------------------------------------------------

// KillRequest is the body for POST /containment/kill.
type KillRequest struct {
	Scope      string `json:"scope"` // "agent", "run", "tenant"
	ScopeValue string `json:"scope_value"`
	TenantID   string `json:"tenant_id"`
	Reason     string `json:"reason"`
}

// KillSwitch represents an active containment action.
type KillSwitch struct {
	Scope       string    `json:"scope"`
	ScopeValue  string    `json:"scope_value"`
	TenantID    string    `json:"tenant_id"`
	Reason      string    `json:"reason"`
	ActivatedAt time.Time `json:"activated_at"`
	ActivatedBy string    `json:"activated_by"`
}

// KillResponse is returned by POST /containment/kill.
type KillResponse struct {
	Status     string     `json:"status"`
	KillSwitch KillSwitch `json:"kill_switch"`
}

// ReleaseRequest is the body for POST /containment/release.
type ReleaseRequest struct {
	Scope      string `json:"scope"`
	ScopeValue string `json:"scope_value"`
	TenantID   string `json:"tenant_id"`
}

// ContainmentStatusResponse is returned by GET /containment/status.
type ContainmentStatusResponse struct {
	ActiveCount  int          `json:"active_count"`
	KillSwitches []KillSwitch `json:"kill_switches"`
}

// ---------------------------------------------------------------------------
// Tenants & API Keys (Admin)
// ---------------------------------------------------------------------------

// CreateTenantRequest is the body for POST /api/tenants.
type CreateTenantRequest struct {
	Name  string `json:"name"`
	Plan  string `json:"plan,omitempty"`
	Label string `json:"label,omitempty"`
	Role  string `json:"role,omitempty"`
}

// CreateTenantResponse is returned by POST /api/tenants.
type CreateTenantResponse struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
	Plan     string `json:"plan"`
	APIKey   string `json:"api_key"`
	KeyHash  string `json:"key_hash"`
	Role     string `json:"role"`
}

// CreateAPIKeyRequest is the body for POST /api/tenants/{id}/keys.
type CreateAPIKeyRequest struct {
	Label string `json:"label,omitempty"`
	Role  string `json:"role,omitempty"`
}

// CreateAPIKeyResponse is returned by POST /api/tenants/{id}/keys.
type CreateAPIKeyResponse struct {
	TenantID string `json:"tenant_id"`
	APIKey   string `json:"api_key"`
	KeyHash  string `json:"key_hash"`
	Label    string `json:"label"`
	Role     string `json:"role"`
}

// APIKeyInfo represents an API key in list/get responses.
type APIKeyInfo struct {
	KeyHashPrefix string  `json:"key_hash_prefix"`
	TenantID      string  `json:"tenant_id,omitempty"`
	Label         string  `json:"label"`
	CreatedAt     string  `json:"created_at"`
	LastUsedAt    *string `json:"last_used_at"`
	Revoked       bool    `json:"revoked"`
	Role          string  `json:"role"`
}

// APIKeyListResponse is returned by GET /api/tenants/{id}/keys.
type APIKeyListResponse struct {
	Keys  []APIKeyInfo `json:"keys"`
	Count int          `json:"count"`
}

// ---------------------------------------------------------------------------
// Usage
// ---------------------------------------------------------------------------

// UsageResponse is returned by GET /api/tenants/{id}/usage.
type UsageResponse struct {
	TenantID        string `json:"tenant_id"`
	Month           string `json:"month"`
	EventsIngested  int64  `json:"events_ingested"`
	TokensIssued    int64  `json:"tokens_issued"`
	AgentsMonitored int    `json:"agents_monitored"`
}

// ---------------------------------------------------------------------------
// Policy Management
// ---------------------------------------------------------------------------

// CreatePolicyVersionRequest is the body for POST /policy/rules.
type CreatePolicyVersionRequest struct {
	Rules json.RawMessage `json:"rules"`
}

// PolicyVersionResponse is returned by POST /policy/rules and POST /policy/rules/{v}/activate.
type PolicyVersionResponse struct {
	Version     int             `json:"version"`
	Rules       json.RawMessage `json:"rules,omitempty"`
	Active      bool            `json:"active"`
	CreatedAt   *time.Time      `json:"created_at,omitempty"`
	ActivatedAt *time.Time      `json:"activated_at,omitempty"`
}

// PolicyVersionListResponse is returned by GET /policy/rules.
type PolicyVersionListResponse struct {
	Versions []PolicyVersionResponse `json:"versions"`
}

// ---------------------------------------------------------------------------
// Roles
// ---------------------------------------------------------------------------

// RoleInfo describes a role and its permissions.
type RoleInfo struct {
	Role        string   `json:"role"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// RoleListResponse is returned by GET /admin/roles.
type RoleListResponse struct {
	Roles []RoleInfo `json:"roles"`
}

// UpdateRoleRequest is the body for PUT /admin/roles.
type UpdateRoleRequest struct {
	KeyHash string `json:"key_hash"`
	Role    string `json:"role"`
}

// UpdateRoleResponse is returned by PUT /admin/roles.
type UpdateRoleResponse struct {
	KeyHash string `json:"key_hash"`
	Role    string `json:"role"`
	Status  string `json:"status"`
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

// HealthResponse is returned by GET /health.
type HealthResponse struct {
	Status  string `json:"status"`
	Uptime  string `json:"uptime"`
	Version string `json:"version"`
	Redis   bool   `json:"redis"`
}
