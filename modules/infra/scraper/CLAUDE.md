# Infra Scraper Module

TypeScript/Bun Playwright-based web scraper and page extractor service.

## Technology Stack

- **Runtime:** Bun
- **Framework:** Hono (lightweight HTTP)
- **Browser:** Patchright (Playwright fork with stealth/anti-detection)
- **Image Processing:** Sharp
- **Anti-Bot:** fingerprint-generator, fingerprint-injector, Ghostery adblocker
- **AI:** Anthropic SDK (element analysis)

## Architecture

Follows hexagonal architecture in TypeScript (kebab-case directories):
```
src/
├── domain/
│   ├── entities/        # ExtractionResult, PreviewResult
│   ├── errors/          # ScraperError with structured error codes
│   ├── services/        # BrowserService, ImageProcessor interfaces
│   └── value-objects/   # Viewport, ProxyConfig, SelectorConfig
├── application/
│   ├── extract-page/    # Full page extraction (screenshot + HTML + text)
│   ├── preview-page/    # Page preview with element mapping
│   └── health-check/    # Browser health status
└── infrastructure/
    ├── http/            # Hono routes (app.ts)
    ├── browser/         # Patchright browser implementation
    │   ├── blocking/    # Ad blocker, Cloudflare handler, cookie blocker
    │   └── stealth/     # Fingerprint randomization
    ├── image/           # Sharp image processor
    └── logger.ts        # Structured logging
```

## HTTP Routes (default port 3000)

- `GET /health` — Browser health check (200 ok / 503 unhealthy)
- `POST /extract` — Extract page content (screenshot, HTML, text, optional sections)
- `POST /preview` — Preview page elements (SSE streaming or JSON based on Accept header)

## Domain Entities

- `ExtractionResult` — title, html, text, screenshot_base64, optional section results
- `SectionResult` — per-section screenshot, html, text, selector match status
- `PreviewResult` — screenshot, viewport, page height, element list with selectors/rects
- `PreviewElement` — selector, xpath, tag, bounding rect, text preview, semantic role

## Commands

```bash
bun run index.ts           # Start server
bun run --watch index.ts   # Start with hot reload (dev)
```

## Notes

- Not part of the Go monolith — runs as a standalone service
- Called by the snapshot module via HTTP (`snapshot/infrastructure/extractor/client.go`)
- Called by the page module for live preview (`POST /preview` with SSE streaming)
- Has its own Dockerfile for deployment
- The composition root in `index.ts` wires all dependencies and handles graceful shutdown
- Preview endpoint supports dual modes: SSE streaming (default) or JSON (via `Accept: application/json`)

## Architecture Improvements

### Security

#### SSRF Protection
The scraper accepts arbitrary URLs with no validation beyond checking presence. This is an SSRF risk — an attacker could use the scraper to access internal services. Add:
1. URL protocol validation (only allow `http://` and `https://`)
2. Block private/internal IP ranges (10.x, 172.16-31.x, 192.168.x, 127.x, 169.254.x, ::1, fc00::/7)
3. DNS resolution check before navigation to prevent DNS rebinding attacks

#### Service-to-Service Authentication
No API key or auth mechanism protects the scraper endpoints. Any network-adjacent service (or attacker with network access) can submit extraction requests. Add:
- API key header validation (e.g., `X-API-Key` checked against `SCRAPER_API_KEY` env var)
- Or mutual TLS for service-to-service auth

#### Request Body Size Limits
No body size limits on `/extract` and `/preview` endpoints. Add Hono middleware to reject oversized payloads (e.g., 1MB limit).

### Code Quality

#### Dead Dependency
`@anthropic-ai/sdk` is declared in `package.json` but never imported in any source file. Remove to reduce image size and dependency surface.

#### No Tests
The scraper has zero test files. Priority areas for testing:
- URL validation (when SSRF protection is added)
- Element selector matching logic
- Screenshot cropping/resizing
- SSE streaming output format

#### Concurrency
`acquireSlot()` in the browser service has no timeout — if all slots are occupied, requests wait indefinitely. Add a configurable timeout (e.g., 30s) that returns 503 Service Unavailable when the scraper is at capacity.

### Scaling
- Run multiple scraper instances behind a load balancer for higher throughput
- The Go backend can round-robin `EXTRACTOR_URL` across multiple instances
- Consider adding a request queue (Redis or SQS) between the Go backend and scraper instances for better backpressure handling
