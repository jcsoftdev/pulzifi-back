// Package ai holds technical clients for AI inference backends.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ErrOllamaDisabled is returned when the client is configured as disabled.
var ErrOllamaDisabled = errors.New("ollama: disabled")

// OllamaClient is a minimal client for the Ollama /api/generate endpoint.
type OllamaClient struct {
	baseURL string
	model   string
	enabled bool
	timeout time.Duration
	http    *http.Client
}

// NewOllamaClient builds a client. When enabled is false, Complete always returns
// ErrOllamaDisabled so callers degrade gracefully without a network call.
func NewOllamaClient(baseURL, model string, enabled bool) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		enabled: enabled,
		timeout: 8 * time.Second,
		http:    &http.Client{},
	}
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format,omitempty"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Complete sends a single non-streaming generation request and returns the raw
// model response text. Format is forced to "json" to bias structured output.
func (c *OllamaClient) Complete(ctx context.Context, prompt string) (string, error) {
	if !c.enabled || c.baseURL == "" {
		return "", ErrOllamaDisabled
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	body, err := json.Marshal(ollamaRequest{Model: c.model, Prompt: prompt, Stream: false, Format: "json"})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("ollama: status %d: %s", resp.StatusCode, string(b))
	}

	var decoded ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}
	return decoded.Response, nil
}
