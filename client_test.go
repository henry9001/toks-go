package toks

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestServer creates an httptest.Server that dispatches to handler and
// returns a Client pointed at it. The server is closed when the test ends.
func newTestServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New("test-api-key",
		WithBaseURL(srv.URL),
		WithAdminKey("test-admin-key"),
		WithUserAgent("toks-test/1.0"),
	)
}

// assertMethod fails the test if the request method does not match want.
func assertMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.Method != want {
		t.Fatalf("method = %s, want %s", r.Method, want)
	}
}

// writeJSON writes v as JSON with the given status code.
func writeJSON(t *testing.T, w http.ResponseWriter, status int, v interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Header injection
// ---------------------------------------------------------------------------

func TestHeaders(t *testing.T) {
	var gotHeaders http.Header
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		writeJSON(t, w, 200, map[string]string{"status": "ok"})
	})

	_, _ = c.Health(context.Background())

	if got := gotHeaders.Get("X-API-Key"); got != "test-api-key" {
		t.Errorf("X-API-Key = %q, want %q", got, "test-api-key")
	}
	if got := gotHeaders.Get("X-Admin-Key"); got != "test-admin-key" {
		t.Errorf("X-Admin-Key = %q, want %q", got, "test-admin-key")
	}
	if got := gotHeaders.Get("User-Agent"); got != "toks-test/1.0" {
		t.Errorf("User-Agent = %q, want %q", got, "toks-test/1.0")
	}
}

func TestContentTypeSetOnPOST(t *testing.T) {
	var gotCT string
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		writeJSON(t, w, 200, RunRegisterResponse{RunID: "r-1"})
	})

	_, _ = c.RegisterRun(context.Background(), &RunRegisterRequest{})
	if gotCT != "application/json" {
		t.Errorf("Content-Type = %q, want %q", gotCT, "application/json")
	}
}

func TestNoContentTypeOnGET(t *testing.T) {
	var gotCT string
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		writeJSON(t, w, 200, HealthResponse{Status: "ok"})
	})

	_, _ = c.Health(context.Background())
	if gotCT != "" {
		t.Errorf("Content-Type on GET = %q, want empty", gotCT)
	}
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

func TestHealth(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		writeJSON(t, w, 200, HealthResponse{
			Status: "ok", Uptime: "5m", Version: "1.0.0", Redis: true,
		})
	})

	resp, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
	if !resp.Redis {
		t.Error("Redis = false, want true")
	}
}

// ---------------------------------------------------------------------------
// Run
// ---------------------------------------------------------------------------

func TestRegisterRun(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/run/register" {
			t.Fatalf("path = %s, want /run/register", r.URL.Path)
		}

		var body RunRegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.InitiatedBy != "alice" {
			t.Errorf("InitiatedBy = %q, want %q", body.InitiatedBy, "alice")
		}

		writeJSON(t, w, 200, RunRegisterResponse{
			RunID:           "run-abc",
			InitiatedBy:     "alice",
			InitiatedByType: "user",
		})
	})

	resp, err := c.RegisterRun(context.Background(), &RunRegisterRequest{
		InitiatedBy:     "alice",
		InitiatedByType: "user",
	})
	if err != nil {
		t.Fatalf("RegisterRun: %v", err)
	}
	if resp.RunID != "run-abc" {
		t.Errorf("RunID = %q, want %q", resp.RunID, "run-abc")
	}
}

func TestGetRun(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		if r.URL.Path != "/run/run-abc" {
			t.Fatalf("path = %s, want /run/run-abc", r.URL.Path)
		}
		writeJSON(t, w, 200, Run{
			ID:    "run-abc",
			State: "active",
		})
	})

	resp, err := c.GetRun(context.Background(), "run-abc")
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if resp.State != "active" {
		t.Errorf("State = %q, want %q", resp.State, "active")
	}
}

func TestListRuns(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		if got := r.URL.Query().Get("status"); got != "active" {
			t.Errorf("status param = %q, want %q", got, "active")
		}
		writeJSON(t, w, 200, RunListResponse{
			Runs:  []Run{{ID: "r-1", State: "active"}},
			Count: 1,
		})
	})

	resp, err := c.ListRuns(context.Background(), &ListRunsOptions{Status: "active"})
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("Count = %d, want 1", resp.Count)
	}
}

func TestListRunsNilOpts(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/runs" {
			t.Fatalf("path = %s, want /runs", r.URL.Path)
		}
		if r.URL.RawQuery != "" {
			t.Errorf("query = %q, want empty", r.URL.RawQuery)
		}
		writeJSON(t, w, 200, RunListResponse{Count: 0})
	})

	_, err := c.ListRuns(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListRuns(nil): %v", err)
	}
}

func TestUpdateRunStatus(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "PUT")
		if r.URL.Path != "/run/run-abc/status" {
			t.Fatalf("path = %s, want /run/run-abc/status", r.URL.Path)
		}
		writeJSON(t, w, 200, RunStatusResponse{RunID: "run-abc", State: "completed"})
	})

	resp, err := c.UpdateRunStatus(context.Background(), "run-abc", &RunStatusRequest{
		State:  "completed",
		Reason: "done",
	})
	if err != nil {
		t.Fatalf("UpdateRunStatus: %v", err)
	}
	if resp.State != "completed" {
		t.Errorf("State = %q, want %q", resp.State, "completed")
	}
}

// ---------------------------------------------------------------------------
// Token
// ---------------------------------------------------------------------------

func TestIssueToken(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/token/issue" {
			t.Fatalf("path = %s, want /token/issue", r.URL.Path)
		}
		writeJSON(t, w, 200, TokenIssueResponse{
			Decision: PolicyDecision{Allowed: true, Reason: "match", TTLMinutes: 30},
			Token:    TokenResponse{Token: "tok-xyz", ExpiresAt: 1700000000, TTLMinutes: 30},
		})
	})

	resp, err := c.IssueToken(context.Background(), &PolicyRequest{
		AgentType:       "builder",
		TenantID:        "t-1",
		RequestedAction: "deploy",
		Resource:        "prod",
		RunID:           "run-1",
	})
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if !resp.Decision.Allowed {
		t.Error("Decision.Allowed = false, want true")
	}
	if resp.Token.Token != "tok-xyz" {
		t.Errorf("Token = %q, want %q", resp.Token.Token, "tok-xyz")
	}
}

func TestVerifyToken(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/token/verify" {
			t.Fatalf("path = %s, want /token/verify", r.URL.Path)
		}
		var body TokenVerifyRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Token != "tok-xyz" {
			t.Errorf("Token = %q, want %q", body.Token, "tok-xyz")
		}
		writeJSON(t, w, 200, TokenVerifyResponse{
			Valid:  true,
			Claims: map[string]interface{}{"agent_type": "builder"},
		})
	})

	resp, err := c.VerifyToken(context.Background(), "tok-xyz")
	if err != nil {
		t.Fatalf("VerifyToken: %v", err)
	}
	if !resp.Valid {
		t.Error("Valid = false, want true")
	}
}

func TestRevokeToken(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/token/revoke" {
			t.Fatalf("path = %s, want /token/revoke", r.URL.Path)
		}
		w.WriteHeader(200)
	})

	err := c.RevokeToken(context.Background(), "tok-xyz")
	if err != nil {
		t.Fatalf("RevokeToken: %v", err)
	}
}

func TestDelegateToken(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/token/delegate" {
			t.Fatalf("path = %s, want /token/delegate", r.URL.Path)
		}
		writeJSON(t, w, 200, DelegateResponse{
			Token:      "child-tok",
			ExpiresAt:  1700000000,
			TTLMinutes: 10,
			Depth:      1,
			ParentHash: "abc123",
		})
	})

	resp, err := c.DelegateToken(context.Background(), &DelegateRequest{
		ParentToken:     "parent-tok",
		ChildAgentType:  "scanner",
		ChildAction:     "read",
		ChildResource:   "logs",
		ChildTTLMinutes: 10,
	})
	if err != nil {
		t.Fatalf("DelegateToken: %v", err)
	}
	if resp.Depth != 1 {
		t.Errorf("Depth = %d, want 1", resp.Depth)
	}
}

func TestTokenChain(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		if got := r.URL.Query().Get("token_hash"); got != "abc123" {
			t.Errorf("token_hash = %q, want %q", got, "abc123")
		}
		writeJSON(t, w, 200, TokenChainResponse{
			TokenHash: "abc123",
			Depth:     2,
		})
	})

	resp, err := c.TokenChain(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("TokenChain: %v", err)
	}
	if resp.Depth != 2 {
		t.Errorf("Depth = %d, want 2", resp.Depth)
	}
}

// ---------------------------------------------------------------------------
// Policy
// ---------------------------------------------------------------------------

func TestCheckPolicy(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/policy/check" {
			t.Fatalf("path = %s, want /policy/check", r.URL.Path)
		}
		writeJSON(t, w, 200, PolicyDecision{
			Allowed:    true,
			Reason:     "default allow",
			Scope:      "tenant",
			TTLMinutes: 60,
		})
	})

	resp, err := c.CheckPolicy(context.Background(), &PolicyRequest{
		AgentType:       "builder",
		TenantID:        "t-1",
		RequestedAction: "deploy",
		Resource:        "staging",
	})
	if err != nil {
		t.Fatalf("CheckPolicy: %v", err)
	}
	if !resp.Allowed {
		t.Error("Allowed = false, want true")
	}
	if resp.TTLMinutes != 60 {
		t.Errorf("TTLMinutes = %d, want 60", resp.TTLMinutes)
	}
}

func TestListPolicyVersions(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		writeJSON(t, w, 200, PolicyVersionListResponse{
			Versions: []PolicyVersionResponse{
				{Version: 1, Active: true},
				{Version: 2, Active: false},
			},
		})
	})

	resp, err := c.ListPolicyVersions(context.Background())
	if err != nil {
		t.Fatalf("ListPolicyVersions: %v", err)
	}
	if len(resp.Versions) != 2 {
		t.Fatalf("len(Versions) = %d, want 2", len(resp.Versions))
	}
}

func TestCreatePolicyVersion(t *testing.T) {
	rules := json.RawMessage(`[{"action":"deploy","effect":"allow"}]`)
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/policy/rules" {
			t.Fatalf("path = %s, want /policy/rules", r.URL.Path)
		}
		writeJSON(t, w, 200, PolicyVersionResponse{Version: 3, Active: false})
	})

	resp, err := c.CreatePolicyVersion(context.Background(), rules)
	if err != nil {
		t.Fatalf("CreatePolicyVersion: %v", err)
	}
	if resp.Version != 3 {
		t.Errorf("Version = %d, want 3", resp.Version)
	}
}

func TestActivatePolicyVersion(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/policy/rules/3/activate" {
			t.Fatalf("path = %s, want /policy/rules/3/activate", r.URL.Path)
		}
		writeJSON(t, w, 200, PolicyVersionResponse{Version: 3, Active: true})
	})

	resp, err := c.ActivatePolicyVersion(context.Background(), 3)
	if err != nil {
		t.Fatalf("ActivatePolicyVersion: %v", err)
	}
	if !resp.Active {
		t.Error("Active = false, want true")
	}
}

// ---------------------------------------------------------------------------
// Heartbeat
// ---------------------------------------------------------------------------

func TestSendHeartbeat(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/heartbeat" {
			t.Fatalf("path = %s, want /heartbeat", r.URL.Path)
		}
		writeJSON(t, w, 200, Agent{
			AgentID:  "a-1",
			TenantID: "t-1",
			Role:     "builder",
			State:    "healthy",
			LastSeen: time.Now(),
		})
	})

	resp, err := c.SendHeartbeat(context.Background(), &HeartbeatRequest{
		AgentID:  "a-1",
		TenantID: "t-1",
		Role:     "builder",
		Session:  "s-1",
	})
	if err != nil {
		t.Fatalf("SendHeartbeat: %v", err)
	}
	if resp.State != "healthy" {
		t.Errorf("State = %q, want %q", resp.State, "healthy")
	}
}

func TestListAgents(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		if got := r.URL.Query().Get("tenant_id"); got != "t-1" {
			t.Errorf("tenant_id = %q, want %q", got, "t-1")
		}
		writeJSON(t, w, 200, AgentListResponse{
			Agents: []Agent{{AgentID: "a-1", State: "healthy"}},
		})
	})

	resp, err := c.ListAgents(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("ListAgents: %v", err)
	}
	if len(resp.Agents) != 1 {
		t.Fatalf("len(Agents) = %d, want 1", len(resp.Agents))
	}
}

func TestListAgentsNoTenant(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("query = %q, want empty", r.URL.RawQuery)
		}
		writeJSON(t, w, 200, AgentListResponse{})
	})

	_, err := c.ListAgents(context.Background(), "")
	if err != nil {
		t.Fatalf("ListAgents: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Containment
// ---------------------------------------------------------------------------

func TestKill(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/containment/kill" {
			t.Fatalf("path = %s, want /containment/kill", r.URL.Path)
		}
		writeJSON(t, w, 200, KillResponse{
			Status: "activated",
			KillSwitch: KillSwitch{
				Scope:      "agent",
				ScopeValue: "a-bad",
				Reason:     "rogue",
			},
		})
	})

	resp, err := c.Kill(context.Background(), &KillRequest{
		Scope:      "agent",
		ScopeValue: "a-bad",
		TenantID:   "t-1",
		Reason:     "rogue",
	})
	if err != nil {
		t.Fatalf("Kill: %v", err)
	}
	if resp.Status != "activated" {
		t.Errorf("Status = %q, want %q", resp.Status, "activated")
	}
}

func TestRelease(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/containment/release" {
			t.Fatalf("path = %s, want /containment/release", r.URL.Path)
		}
		w.WriteHeader(200)
	})

	err := c.Release(context.Background(), &ReleaseRequest{
		Scope:      "agent",
		ScopeValue: "a-bad",
		TenantID:   "t-1",
	})
	if err != nil {
		t.Fatalf("Release: %v", err)
	}
}

func TestContainmentStatus(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		if got := r.URL.Query().Get("tenant_id"); got != "t-1" {
			t.Errorf("tenant_id = %q, want %q", got, "t-1")
		}
		writeJSON(t, w, 200, ContainmentStatusResponse{ActiveCount: 1})
	})

	resp, err := c.ContainmentStatus(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("ContainmentStatus: %v", err)
	}
	if resp.ActiveCount != 1 {
		t.Errorf("ActiveCount = %d, want 1", resp.ActiveCount)
	}
}

func TestContainmentStatusNoTenant(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("query = %q, want empty", r.URL.RawQuery)
		}
		writeJSON(t, w, 200, ContainmentStatusResponse{ActiveCount: 0})
	})

	_, err := c.ContainmentStatus(context.Background(), "")
	if err != nil {
		t.Fatalf("ContainmentStatus: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Audit
// ---------------------------------------------------------------------------

func TestListAuditEvents(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		if r.URL.Path != "/audit/events" {
			t.Fatalf("path = %s, want /audit/events", r.URL.Path)
		}
		writeJSON(t, w, 200, AuditEventsResponse{
			Events: []map[string]interface{}{{"event_type": "token_issued"}},
		})
	})

	resp, err := c.ListAuditEvents(context.Background())
	if err != nil {
		t.Fatalf("ListAuditEvents: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Fatalf("len(Events) = %d, want 1", len(resp.Events))
	}
}

func TestQueryAttribution(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		if got := r.URL.Query().Get("user"); got != "alice" {
			t.Errorf("user = %q, want %q", got, "alice")
		}
		if r.URL.Query().Get("since") == "" {
			t.Error("since param missing")
		}
		writeJSON(t, w, 200, AttributionResponse{
			User:  "alice",
			Count: 2,
		})
	})

	resp, err := c.QueryAttribution(context.Background(), &AttributionOptions{
		User:  "alice",
		Since: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("QueryAttribution: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("Count = %d, want 2", resp.Count)
	}
}

// ---------------------------------------------------------------------------
// Admin
// ---------------------------------------------------------------------------

func TestCreateTenant(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/api/tenants" {
			t.Fatalf("path = %s, want /api/tenants", r.URL.Path)
		}
		writeJSON(t, w, 200, CreateTenantResponse{
			TenantID: "t-new",
			Name:     "Acme",
			Plan:     "pro",
			APIKey:   "key-abc",
		})
	})

	resp, err := c.CreateTenant(context.Background(), &CreateTenantRequest{
		Name: "Acme",
		Plan: "pro",
	})
	if err != nil {
		t.Fatalf("CreateTenant: %v", err)
	}
	if resp.TenantID != "t-new" {
		t.Errorf("TenantID = %q, want %q", resp.TenantID, "t-new")
	}
}

func TestGetTenant(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		if r.URL.Path != "/api/tenants/t-1" {
			t.Fatalf("path = %s, want /api/tenants/t-1", r.URL.Path)
		}
		writeJSON(t, w, 200, CreateTenantResponse{TenantID: "t-1", Name: "Acme"})
	})

	resp, err := c.GetTenant(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("GetTenant: %v", err)
	}
	if resp.Name != "Acme" {
		t.Errorf("Name = %q, want %q", resp.Name, "Acme")
	}
}

func TestCreateAPIKey(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "POST")
		if r.URL.Path != "/api/tenants/t-1/keys" {
			t.Fatalf("path = %s, want /api/tenants/t-1/keys", r.URL.Path)
		}
		writeJSON(t, w, 200, CreateAPIKeyResponse{
			TenantID: "t-1",
			APIKey:   "key-new",
			KeyHash:  "hash-new",
			Label:    "ci",
		})
	})

	resp, err := c.CreateAPIKey(context.Background(), "t-1", &CreateAPIKeyRequest{Label: "ci"})
	if err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}
	if resp.Label != "ci" {
		t.Errorf("Label = %q, want %q", resp.Label, "ci")
	}
}

func TestRevokeAPIKey(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "DELETE")
		if r.URL.Path != "/api/tenants/t-1/keys/hash-old" {
			t.Fatalf("path = %s, want /api/tenants/t-1/keys/hash-old", r.URL.Path)
		}
		w.WriteHeader(200)
	})

	err := c.RevokeAPIKey(context.Background(), "t-1", "hash-old")
	if err != nil {
		t.Fatalf("RevokeAPIKey: %v", err)
	}
}

func TestListAPIKeys(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		writeJSON(t, w, 200, APIKeyListResponse{
			Keys:  []APIKeyInfo{{Label: "default", Role: "agent"}},
			Count: 1,
		})
	})

	resp, err := c.ListAPIKeys(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("ListAPIKeys: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("Count = %d, want 1", resp.Count)
	}
}

func TestGetAPIKey(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		if r.URL.Path != "/api/tenants/t-1/keys/hash-abc" {
			t.Fatalf("path = %s, want /api/tenants/t-1/keys/hash-abc", r.URL.Path)
		}
		writeJSON(t, w, 200, APIKeyInfo{Label: "ci", Role: "agent"})
	})

	resp, err := c.GetAPIKey(context.Background(), "t-1", "hash-abc")
	if err != nil {
		t.Fatalf("GetAPIKey: %v", err)
	}
	if resp.Role != "agent" {
		t.Errorf("Role = %q, want %q", resp.Role, "agent")
	}
}

func TestGetUsage(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		if r.URL.Path != "/api/tenants/t-1/usage" {
			t.Fatalf("path = %s, want /api/tenants/t-1/usage", r.URL.Path)
		}
		writeJSON(t, w, 200, UsageResponse{
			TenantID:       "t-1",
			Month:          "2026-03",
			EventsIngested: 1500,
			TokensIssued:   42,
		})
	})

	resp, err := c.GetUsage(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("GetUsage: %v", err)
	}
	if resp.TokensIssued != 42 {
		t.Errorf("TokensIssued = %d, want 42", resp.TokensIssued)
	}
}

func TestUpdateRole(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "PUT")
		if r.URL.Path != "/admin/roles" {
			t.Fatalf("path = %s, want /admin/roles", r.URL.Path)
		}
		writeJSON(t, w, 200, UpdateRoleResponse{
			KeyHash: "hash-abc",
			Role:    "operator",
			Status:  "updated",
		})
	})

	resp, err := c.UpdateRole(context.Background(), "hash-abc", "operator")
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if resp.Role != "operator" {
		t.Errorf("Role = %q, want %q", resp.Role, "operator")
	}
}

func TestListRoles(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		writeJSON(t, w, 200, RoleListResponse{
			Roles: []RoleInfo{
				{Role: "agent", Description: "Agent role"},
				{Role: "operator", Description: "Operator role"},
			},
		})
	})

	resp, err := c.ListRoles(context.Background())
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if len(resp.Roles) != 2 {
		t.Fatalf("len(Roles) = %d, want 2", len(resp.Roles))
	}
}

// ---------------------------------------------------------------------------
// Error paths
// ---------------------------------------------------------------------------

func TestError4xx(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 401, map[string]string{
			"error":  "unauthorized",
			"reason": "invalid api key",
		})
	})

	_, err := c.Health(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
	if apiErr.Code != "unauthorized" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "unauthorized")
	}
	if !IsUnauthorized(err) {
		t.Error("IsUnauthorized = false, want true")
	}
	if IsNotFound(err) {
		t.Error("IsNotFound = true, want false")
	}
}

func TestError404(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 404, map[string]string{"error": "not_found"})
	})

	_, err := c.GetRun(context.Background(), "no-such-run")
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true (err = %v)", err)
	}
}

func TestError403(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 403, map[string]string{"error": "forbidden"})
	})

	_, err := c.ListRoles(context.Background())
	if !IsForbidden(err) {
		t.Errorf("IsForbidden = false, want true (err = %v)", err)
	}
}

func TestError5xx(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	})

	_, err := c.Health(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
	// When body isn't valid JSON, Code falls back to status text.
	if apiErr.Code != "Internal Server Error" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "Internal Server Error")
	}
}

func TestIsKilled(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 403, map[string]string{
			"error":  "killed",
			"reason": "containment active",
		})
	})

	_, err := c.IssueToken(context.Background(), &PolicyRequest{
		AgentType: "builder", TenantID: "t-1",
		RequestedAction: "deploy", Resource: "prod",
	})
	if !IsKilled(err) {
		t.Errorf("IsKilled = false, want true (err = %v)", err)
	}
}

func TestAPIErrorString(t *testing.T) {
	e := &APIError{StatusCode: 400, Code: "bad_request", Reason: "missing field", Detail: "tenant_id required"}
	got := e.Error()
	want := "toks: 400 bad_request: missing field (tenant_id required)"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAPIErrorStringMinimal(t *testing.T) {
	e := &APIError{StatusCode: 500, Code: "Internal Server Error"}
	got := e.Error()
	want := "toks: 500 Internal Server Error"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Context cancellation
// ---------------------------------------------------------------------------

func TestContextCancellation(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow server — the context should cancel before we respond.
		<-r.Context().Done()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := c.Health(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("error = %q, want context canceled", err)
	}
}

// ---------------------------------------------------------------------------
// Client construction
// ---------------------------------------------------------------------------

func TestNewDefaults(t *testing.T) {
	c := New("my-key")
	if c.apiKey != "my-key" {
		t.Errorf("apiKey = %q, want %q", c.apiKey, "my-key")
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
	if c.http != http.DefaultClient {
		t.Error("http client is not http.DefaultClient")
	}
}

func TestWithOptions(t *testing.T) {
	hc := &http.Client{Timeout: 5 * time.Second}
	c := New("k",
		WithBaseURL("https://custom.dev/"),
		WithHTTPClient(hc),
		WithAdminKey("admin"),
		WithUserAgent("ua/1"),
	)
	if c.baseURL != "https://custom.dev" { // trailing slash trimmed
		t.Errorf("baseURL = %q", c.baseURL)
	}
	if c.adminKey != "admin" {
		t.Errorf("adminKey = %q", c.adminKey)
	}
	if c.userAgent != "ua/1" {
		t.Errorf("userAgent = %q", c.userAgent)
	}
	if c.http != hc {
		t.Error("http client not set")
	}
}

// ---------------------------------------------------------------------------
// Edge cases in do()
// ---------------------------------------------------------------------------

func TestDoInvalidURL(t *testing.T) {
	c := New("k", WithBaseURL("://bad"))
	_, err := c.Health(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestDoReadBodyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set Content-Length to lie about size, then close.
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
		// The http client may or may not flag this; depends on implementation.
		// We mainly test that the SDK doesn't panic.
	}))
	defer srv.Close()
	c := New("k", WithBaseURL(srv.URL))
	// This should succeed or return a non-nil error — no panics.
	_, _ = c.Health(context.Background())
}

func TestDoMalformedJSON(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{not json}`))
	})

	_, err := c.Health(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("error = %q, want 'decode response'", err)
	}
}

func TestDoEmptyBody(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		// No body at all.
	})

	// RevokeToken expects no response body — should succeed.
	err := c.RevokeToken(context.Background(), "tok-x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNoAPIKeyHeader(t *testing.T) {
	var gotHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		writeJSON(t, w, 200, HealthResponse{Status: "ok"})
	}))
	defer srv.Close()

	c := New("", WithBaseURL(srv.URL))
	c.Health(context.Background())
	if got := gotHeaders.Get("X-API-Key"); got != "" {
		t.Errorf("X-API-Key = %q, want empty for blank key", got)
	}
}

// ---------------------------------------------------------------------------
// Error paths for all methods (cover the err != nil return branches)
// ---------------------------------------------------------------------------

func newErrorServer(t *testing.T) *Client {
	t.Helper()
	return newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 500, map[string]string{"error": "boom"})
	})
}

func TestRegisterRunError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.RegisterRun(context.Background(), &RunRegisterRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateRunStatusError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.UpdateRunStatus(context.Background(), "r-1", &RunStatusRequest{State: "failed"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIssueTokenError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.IssueToken(context.Background(), &PolicyRequest{AgentType: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifyTokenError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.VerifyToken(context.Background(), "tok-x")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDelegateTokenError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.DelegateToken(context.Background(), &DelegateRequest{ParentToken: "p"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTokenChainError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.TokenChain(context.Background(), "hash")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckPolicyError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.CheckPolicy(context.Background(), &PolicyRequest{AgentType: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListPolicyVersionsError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.ListPolicyVersions(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreatePolicyVersionError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.CreatePolicyVersion(context.Background(), json.RawMessage(`[]`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestActivatePolicyVersionError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.ActivatePolicyVersion(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSendHeartbeatError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.SendHeartbeat(context.Background(), &HeartbeatRequest{AgentID: "a"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListAgentsError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.ListAgents(context.Background(), "t-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestKillError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.Kill(context.Background(), &KillRequest{Scope: "agent"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReleaseError(t *testing.T) {
	c := newErrorServer(t)
	err := c.Release(context.Background(), &ReleaseRequest{Scope: "agent"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestContainmentStatusError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.ContainmentStatus(context.Background(), "t-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListAuditEventsError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.ListAuditEvents(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestQueryAttributionError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.QueryAttribution(context.Background(), &AttributionOptions{User: "alice"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateTenantError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.CreateTenant(context.Background(), &CreateTenantRequest{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetTenantError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.GetTenant(context.Background(), "t-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateAPIKeyError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.CreateAPIKey(context.Background(), "t-1", &CreateAPIKeyRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListAPIKeysError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.ListAPIKeys(context.Background(), "t-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetAPIKeyError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.GetAPIKey(context.Background(), "t-1", "hash")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetUsageError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.GetUsage(context.Background(), "t-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateRoleError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.UpdateRole(context.Background(), "hash", "admin")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListRolesError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.ListRoles(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListRunsError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.ListRuns(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHealthError(t *testing.T) {
	c := newErrorServer(t)
	_, err := c.Health(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRevokeTokenError(t *testing.T) {
	c := newErrorServer(t)
	err := c.RevokeToken(context.Background(), "tok-x")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRevokeAPIKeyError(t *testing.T) {
	c := newErrorServer(t)
	err := c.RevokeAPIKey(context.Background(), "t-1", "hash")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---------------------------------------------------------------------------
// ListRuns with all query params
// ---------------------------------------------------------------------------

func TestListRunsAllOptions(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, r, "GET")
		q := r.URL.Query()
		if got := q.Get("status"); got != "active" {
			t.Errorf("status = %q, want %q", got, "active")
		}
		if got := q.Get("agent_type"); got != "builder" {
			t.Errorf("agent_type = %q, want %q", got, "builder")
		}
		if got := q.Get("project"); got != "myproj" {
			t.Errorf("project = %q, want %q", got, "myproj")
		}
		if got := q.Get("since"); got != "2026-01-01T00:00:00Z" {
			t.Errorf("since = %q", got)
		}
		if got := q.Get("until"); got != "2026-03-01T00:00:00Z" {
			t.Errorf("until = %q", got)
		}
		writeJSON(t, w, 200, RunListResponse{Count: 5})
	})

	resp, err := c.ListRuns(context.Background(), &ListRunsOptions{
		Status:    "active",
		AgentType: "builder",
		Project:   "myproj",
		Since:     "2026-01-01T00:00:00Z",
		Until:     "2026-03-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if resp.Count != 5 {
		t.Errorf("Count = %d, want 5", resp.Count)
	}
}

// ---------------------------------------------------------------------------
// QueryAttribution without Since
// ---------------------------------------------------------------------------

func TestQueryAttributionNoSince(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("since") != "" {
			t.Error("since param should be empty when not set")
		}
		writeJSON(t, w, 200, AttributionResponse{User: "bob", Count: 0})
	})

	_, err := c.QueryAttribution(context.Background(), &AttributionOptions{User: "bob"})
	if err != nil {
		t.Fatalf("QueryAttribution: %v", err)
	}
}

func TestRequestBodySentCorrectly(t *testing.T) {
	c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req HeartbeatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if req.AgentID != "a-1" || req.TenantID != "t-1" || req.Role != "worker" || req.Session != "s-42" {
			t.Errorf("body mismatch: %s", body)
		}
		writeJSON(t, w, 200, Agent{AgentID: "a-1", State: "healthy"})
	})

	c.SendHeartbeat(context.Background(), &HeartbeatRequest{
		AgentID:  "a-1",
		TenantID: "t-1",
		Role:     "worker",
		Session:  "s-42",
	})
}
