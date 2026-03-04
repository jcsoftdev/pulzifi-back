# Insight Module

AI-powered insight generation for page changes using LLM.

## Domain Entities

- `Insight` — AI-generated insight with title, content, metadata

## Use Cases

- `generate_insights` — generate insights for check changes (async, background)
- `list_insights` — list insights for a page or check

## HTTP Routes (`/insights/*`, tenant-aware)

- POST `/insights/generate` (returns 202 Accepted)
- GET `/insights`
- GET `/insights/{id}`
- GET `/insights/sse` — SSE stream for generation completion

## Domain Services

- `InsightGenerator` interface (OpenRouter implementation)

## Infrastructure

- OpenRouter AI client (configurable model)
- HTML text extraction from snapshots
- PostgreSQL: `insights` table (tenant-scoped)
- Pub/Sub Broker: SSE notifications when generation completes

## Notes

- Returns 202 immediately; generation runs in background
- Configured via `OPENROUTER_API_KEY` and `OPENROUTER_MODEL`
