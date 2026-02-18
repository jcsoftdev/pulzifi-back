package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultBaseURL = "https://openrouter.ai/api/v1"

// Message represents a single chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterClient is an HTTP client for the OpenRouter API (OpenAI-compatible).
type OpenRouterClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// NewOpenRouterClient creates a new OpenRouterClient with the given API key and model.
func NewOpenRouterClient(apiKey, model string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Complete sends a chat completion request and returns the assistant message content.
func (c *OpenRouterClient) Complete(ctx context.Context, messages []Message) (string, error) {
	payload := map[string]interface{}{
		"model":    c.model,
		"messages": messages,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("openrouter: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("openrouter: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openrouter: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return "", fmt.Errorf("openrouter: status %d: %v", resp.StatusCode, errBody)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("openrouter: decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openrouter: empty choices in response")
	}

	return result.Choices[0].Message.Content, nil
}
