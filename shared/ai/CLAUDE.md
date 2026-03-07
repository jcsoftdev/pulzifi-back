# AI Package (`shared/ai/`)

OpenRouter LLM client for text completions and multimodal (vision) analysis.

## Files

- `openrouter.go` — HTTP client for the OpenRouter API (OpenAI-compatible)

## Exported API

### Types
- `Message` — Single chat message (`Role`, `Content` strings)
- `ContentBlock` — Multimodal content block (`Type`: "text" or "image_url", `Text`, `ImageURL`)
- `ImageURL` — Image URL holder (supports base64 data URIs)
- `MultimodalMessage` — Chat message with `[]ContentBlock` content
- `OpenRouterClient` — HTTP client with API key, model, and 120s timeout

### Functions
- `NewOpenRouterClient(apiKey, model) *OpenRouterClient` — Creates client with default base URL (`https://openrouter.ai/api/v1`)
- `NewOpenRouterClientWithModel(apiKey, model) *OpenRouterClient` — Alias, useful for creating separate vision client

### Methods (`*OpenRouterClient`)
- `Complete(ctx, []Message) (string, error)` — Text-only chat completion
- `CompleteMultimodal(ctx, []MultimodalMessage) (string, error)` — Multimodal completion (text + images)

## Usage

Used by:
- `modules/insight/infrastructure/ai/` — Text insight generation and vision-based screenshot analysis

## Dependencies

Standard library only (`bytes`, `context`, `encoding/json`, `fmt`, `net/http`, `time`).
