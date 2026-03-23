package toks_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	toks "github.com/henry9001/toks-go"
)

// fakeServer returns an httptest.Server that responds with canned JSON for
// the endpoints used in the examples below.
func fakeServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(toks.HealthResponse{
			Status: "ok", Uptime: "10m", Version: "1.0.0", Redis: true,
		})
	})

	mux.HandleFunc("/run/register", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(toks.RunRegisterResponse{
			RunID: "run-abc123", InitiatedBy: "alice", InitiatedByType: "user",
		})
	})

	mux.HandleFunc("/token/issue", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(toks.TokenIssueResponse{
			Decision: toks.PolicyDecision{Allowed: true, Reason: "default allow", TTLMinutes: 30},
			Token:    toks.TokenResponse{Token: "tok-secret", ExpiresAt: 1700000000, TTLMinutes: 30},
		})
	})

	mux.HandleFunc("/token/verify", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(toks.TokenVerifyResponse{
			Valid:  true,
			Claims: map[string]interface{}{"agent_type": "builder"},
		})
	})

	mux.HandleFunc("/policy/check", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(toks.PolicyDecision{
			Allowed: true, Reason: "default allow", Scope: "tenant", TTLMinutes: 60,
		})
	})

	mux.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(toks.Agent{
			AgentID: "agent-1", TenantID: "t-1", Role: "builder", State: "healthy",
		})
	})

	return httptest.NewServer(mux)
}

func ExampleNew() {
	// Create a client with default settings.
	client := toks.New("your-api-key")

	// The client is ready to use. Point it at your TOKS server:
	//   client := toks.New("your-api-key", toks.WithBaseURL("https://toks.example.com"))
	_ = client
	fmt.Println("client created")
	// Output: client created
}

func ExampleNew_withOptions() {
	// Create a client with custom base URL and user agent.
	_ = toks.New("your-api-key",
		toks.WithBaseURL("https://toks.example.com"),
		toks.WithUserAgent("my-agent/1.0"),
	)
	fmt.Println("configured")
	// Output: configured
}

func ExampleClient_RegisterRun() {
	srv := fakeServer()
	defer srv.Close()
	client := toks.New("my-key", toks.WithBaseURL(srv.URL))

	run, err := client.RegisterRun(context.Background(), &toks.RunRegisterRequest{
		InitiatedBy:     "alice",
		InitiatedByType: "user",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("run:", run.RunID)
	// Output: run: run-abc123
}

func ExampleClient_IssueToken() {
	srv := fakeServer()
	defer srv.Close()
	client := toks.New("my-key", toks.WithBaseURL(srv.URL))

	resp, err := client.IssueToken(context.Background(), &toks.PolicyRequest{
		AgentType:       "builder",
		TenantID:        "t-1",
		RequestedAction: "deploy",
		Resource:        "staging",
		RunID:           "run-abc123",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("allowed:", resp.Decision.Allowed)
	fmt.Println("ttl:", resp.Token.TTLMinutes)
	// Output:
	// allowed: true
	// ttl: 30
}

func ExampleClient_VerifyToken() {
	srv := fakeServer()
	defer srv.Close()
	client := toks.New("my-key", toks.WithBaseURL(srv.URL))

	resp, err := client.VerifyToken(context.Background(), "tok-secret")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("valid:", resp.Valid)
	// Output: valid: true
}

func ExampleClient_CheckPolicy() {
	srv := fakeServer()
	defer srv.Close()
	client := toks.New("my-key", toks.WithBaseURL(srv.URL))

	decision, err := client.CheckPolicy(context.Background(), &toks.PolicyRequest{
		AgentType:       "builder",
		TenantID:        "t-1",
		RequestedAction: "deploy",
		Resource:        "prod",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("allowed:", decision.Allowed)
	fmt.Println("ttl:", decision.TTLMinutes, "minutes")
	// Output:
	// allowed: true
	// ttl: 60 minutes
}

func ExampleClient_SendHeartbeat() {
	srv := fakeServer()
	defer srv.Close()
	client := toks.New("my-key", toks.WithBaseURL(srv.URL))

	agent, err := client.SendHeartbeat(context.Background(), &toks.HeartbeatRequest{
		AgentID:  "agent-1",
		TenantID: "t-1",
		Role:     "builder",
		Session:  "sess-1",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("state:", agent.State)
	// Output: state: healthy
}

func ExampleIsNotFound() {
	err := &toks.APIError{StatusCode: 404, Code: "not_found"}
	fmt.Println(toks.IsNotFound(err))
	// Output: true
}

func ExampleIsKilled() {
	err := &toks.APIError{StatusCode: 403, Code: "killed", Reason: "containment active"}
	fmt.Println(toks.IsKilled(err))
	// Output: true
}
