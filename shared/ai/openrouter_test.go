package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenRouterClient_WithBaseURL_Default(t *testing.T) {
	c := NewOpenRouterClient("key", "model")
	if c.baseURL != defaultBaseURL {
		t.Fatalf("baseURL = %q, want default %q", c.baseURL, defaultBaseURL)
	}
	// Empty override keeps the default so existing OpenRouter deployments are untouched.
	if got := c.WithBaseURL("").baseURL; got != defaultBaseURL {
		t.Errorf("empty override changed baseURL to %q, want %q", got, defaultBaseURL)
	}
}

func TestOpenRouterClient_WithBaseURL_Override(t *testing.T) {
	c := NewOpenRouterClient("key", "model").WithBaseURL("http://ollama:11434/v1")
	if c.baseURL != "http://ollama:11434/v1" {
		t.Errorf("baseURL = %q, want override", c.baseURL)
	}
}

// Complete must hit the overridden endpoint, proving any OpenAI-compatible
// backend (Ollama /v1, DeepInfra, Together) works without code changes.
func TestOpenRouterClient_Complete_HonorsBaseURL(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "ok"}},
			},
		})
	}))
	defer srv.Close()

	c := NewOpenRouterClient("key", "model").WithBaseURL(srv.URL)
	out, err := c.Complete(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/chat/completions" {
		t.Errorf("path = %q, want /chat/completions", gotPath)
	}
	if out != "ok" {
		t.Errorf("content = %q, want ok", out)
	}
}
