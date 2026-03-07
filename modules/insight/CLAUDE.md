# Insight Module

AI-powered insight generation for page changes using LLM.

## Domain Entities

- `Insight` — AI-generated insight with title, content, metadata

## Use Cases (application/ directories)

- `generate_insights` — generate insights for check changes (async, background)
- `list_insights` — list insights for a page or check

## HTTP Routes (`/insights/*`, tenant-aware)

- POST `/insights/generate` (returns 202 Accepted)
- GET `/insights` — list insights
- GET `/insights/{id}` — get insight details
- GET `/insights/sse` — SSE stream for generation completion

## Domain Services

- `InsightGenerator` interface (OpenRouter implementation)
- `VisionAnalyzer` interface (OpenRouter multimodal vision implementation)

## Infrastructure

- OpenRouter AI client: `infrastructure/ai/openrouter_generator.go` (text completions)
- Vision analyzer: `infrastructure/ai/vision_analyzer.go` (multimodal image analysis)
- HTML text extraction from snapshots
- PostgreSQL: `insights` table (tenant-scoped)
- Pub/Sub Broker (InsightBroker): SSE notifications when generation completes

## Notes

- Returns 202 immediately; generation runs in background
- Configured via `OPENROUTER_API_KEY`, `OPENROUTER_MODEL`, `OPENROUTER_VISION_MODEL`

## Cross-Module Dependencies (violations)

- Imports `modules/monitoring/infrastructure/persistence` (CheckPostgresRepository, MonitoringConfigPostgresRepository)

**Recommended:** Define check/config repository interfaces in this module's domain layer. Inject the monitoring module's implementations from `cmd/server/modules.go`.

## Architecture Improvements

- **LLM calls have no retry logic.** OpenRouter API can be flaky — add exponential backoff retries with circuit breaker.
- **No response caching.** Identical check pairs will re-generate insights. Consider caching by content hash pair.
- **Vision analysis is synchronous within the background goroutine.** For large images, this blocks the goroutine. Consider a dedicated worker pool for AI calls.
