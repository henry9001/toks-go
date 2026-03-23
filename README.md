# toks-go

Go SDK for [TOKS](https://github.com/henry9001/agent-saas-toks) — the JIT privilege service for AI agents.

Zero external dependencies. Stdlib only.

## Install

```bash
go get github.com/henry9001/toks-go@latest
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	toks "github.com/henry9001/toks-go"
)

func main() {
	client := toks.New("your-api-key")

	ctx := context.Background()

	// Register a run
	run, err := client.RegisterRun(ctx, &toks.RunRegisterRequest{
		InitiatedBy:     "jane@example.com",
		InitiatedByType: "user",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Issue a scoped token
	resp, err := client.IssueToken(ctx, &toks.PolicyRequest{
		AgentType:       "ops-agent",
		TenantID:        "t1",
		RequestedAction: "read",
		Resource:        "ops:status",
		RunID:           run.RunID,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Token:", resp.Token.Token)

	// Verify
	v, err := client.VerifyToken(ctx, resp.Token.Token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Valid:", v.Valid)
}
```

## API Coverage

All 24+ TOKS endpoints are covered:

| Category | Methods |
|----------|---------|
| **Run** | `RegisterRun`, `GetRun`, `ListRuns` |
| **Token** | `IssueToken`, `VerifyToken`, `RevokeToken`, `DelegateToken`, `TokenChain` |
| **Policy** | `CheckPolicy`, `ListPolicyVersions`, `CreatePolicyVersion`, `ActivatePolicyVersion` |
| **Heartbeat** | `SendHeartbeat`, `ListAgents` |
| **Containment** | `Kill`, `Release`, `ContainmentStatus` |
| **Audit** | `ListAuditEvents`, `QueryAttribution` |
| **Admin** | `CreateTenant`, `GetTenant`, `CreateAPIKey`, `RevokeAPIKey`, `GetUsage`, `UpdateRole` |
| **Health** | `Health` |

## Configuration

```go
// Custom base URL
client := toks.New("api-key", toks.WithBaseURL("http://localhost:8080"))

// Custom HTTP client
client := toks.New("api-key", toks.WithHTTPClient(&http.Client{
	Timeout: 10 * time.Second,
}))

// Admin operations (kill switch, tenant management)
client := toks.New("api-key", toks.WithAdminKey("admin-key"))
```

## Error Handling

```go
resp, err := client.IssueToken(ctx, req)
if err != nil {
	var apiErr *toks.APIError
	if errors.As(err, &apiErr) {
		if apiErr.IsUnauthorized() { /* bad API key */ }
		if apiErr.IsForbidden()    { /* insufficient role */ }
		if apiErr.IsNotFound()     { /* resource missing */ }
		if apiErr.IsKilled()       { /* kill switch active */ }
	}
}
```

## Requirements

- Go 1.22+
- No external dependencies

## License

Proprietary.
