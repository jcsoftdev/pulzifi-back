# Infra Module

Infrastructure services — headless browser page extraction.

## Implementation

- Bun + TypeScript service using Hono framework
- Runs Playwright with rebrowser-playwright (stealth mode)
- Exposes HTTP API for page screenshot and HTML extraction

## Components

- `index.ts` — Hono server with Chromium automation
- Stealth scripts to hide automation signals
- DOM mutation detection for page render completion

## Notes

- Separate TypeScript service (not Go)
- Containerized with Dockerfile
- Uses rebrowser-playwright for anti-detection
