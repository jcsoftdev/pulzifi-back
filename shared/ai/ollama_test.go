package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOllamaClient_Disabled(t *testing.T) {
	c := NewOllamaClient("", "qwen3:4b", false)
	_, err := c.Complete(context.Background(), "prompt")
	if err == nil {
		t.Error("disabled client must return error, not call out")
	}
}

func TestOllamaClient_Complete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"response": `{"1":"Hero"}`, "done": true})
	}))
	defer srv.Close()

	c := NewOllamaClient(srv.URL, "qwen3:4b", true)
	c.timeout = 2 * time.Second
	out, err := c.Complete(context.Background(), "name the sections")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != `{"1":"Hero"}` {
		t.Errorf("response = %q, want {\"1\":\"Hero\"}", out)
	}
}
