#!/usr/bin/env bash
# Run the full E2E pipeline: bring up the stack, run the API-driven Playwright test,
# tear down. The test uses Playwright's APIRequestContext (no browser), so there's no
# web service and no Chromium install — it drives the real backend stack via HTTP.
set -euo pipefail

# Absolute compose path so the EXIT trap still works after we `cd frontend`.
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
COMPOSE="docker compose -f $ROOT/docker-compose.e2e.yml"
cleanup() { $COMPOSE down -v --remove-orphans || true; }
trap cleanup EXIT

echo "==> Cleaning any prior stack (fresh volumes + fixture at v1)"
$COMPOSE down -v --remove-orphans 2>/dev/null || true

echo "==> Building & starting E2E stack"
$COMPOSE up -d --build

echo "==> Waiting for monolith health (max ~3m)"
for i in $(seq 1 90); do
  if curl -sf http://jcsoftdev-inc.lvh.me:3000/health >/dev/null; then break; fi
  sleep 2
  if [ "$i" -eq 90 ]; then echo "monolith never became healthy" >&2; $COMPOSE logs monolith; exit 1; fi
done

echo "==> Running Playwright (API-driven, no browser)"
cd frontend
bun install --frozen-lockfile
bun run e2e
