# Insight Module

## Responsibility

AI-powered insight generation from detected page changes using OpenRouter LLM. Supports multiple insight types and real-time streaming of generation progress via SSE.

## Entities

- **Insight** — ID, PageID, CheckID, InsightType (seo/performance/content/accessibility), Title, Content, Metadata, CreatedAt

## Repository Interfaces

- `InsightRepository` — Create, ListByPageID, ListByCheckID, GetByID

## Domain Services

- `InsightGenerator` interface — implemented by OpenRouter adapter in `infrastructure/ai/`

## Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/insights/generate` | Generate insights (async, returns 202) |
| GET | `/insights` | List insights (filterable by page_id, check_id) |
| GET | `/insights/{id}` | Get insight |
| GET | `/insights/sse?check_id=...` | SSE stream for generation progress |

## Dependencies

- Monitoring module (provides check data)
- OpenRouter API (`shared/ai/`) for LLM inference
- HTML text extraction (`shared/html/`)
- InsightBroker (`shared/pubsub/`) for SSE streaming
- Email module (optional notification)

## Constraints

- Requires `OPENROUTER_API_KEY` to be set (disabled otherwise)
- Default model: `mistralai/mistral-7b-instruct:free`
- Insights generated asynchronously; client polls via SSE
- InsightBroker has 5-minute replay cache for late SSE subscribers
- SSE connection timeout: 120 seconds
