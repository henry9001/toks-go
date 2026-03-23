// Package toks provides a Go client for the TOKS JIT privilege service.
//
// TOKS manages just-in-time token issuance, policy evaluation, agent heartbeats,
// run lifecycle, containment kill switches, and audit events for multi-tenant
// agent infrastructure.
//
// Create a client with an API key:
//
//	client := toks.New("your-api-key")
//
// Or with options:
//
//	client := toks.New("your-api-key",
//	    toks.WithBaseURL("https://custom.endpoint.dev"),
//	    toks.WithUserAgent("my-agent/1.0"),
//	)
package toks
