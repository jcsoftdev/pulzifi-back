#!/bin/bash
# Pre-commit validation. Only outputs errors to keep Claude context minimal.

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
ERRORS=""

output=$(go build ./... 2>&1) || ERRORS="${ERRORS}[build] ${output}\n"
output=$(go vet ./... 2>&1) || ERRORS="${ERRORS}[vet] ${output}\n"
# Mirror CI's "Frontend lint + types + tests" job (.github/workflows/ci.yml) so a
# frontend break is caught at commit time, not after push. lint is what CI runs
# (biome lint, not format) — the gap that let a noDangerouslySetInnerHtml error ship.
output=$(cd "$PROJECT_DIR/frontend" && bun run lint 2>&1) || ERRORS="${ERRORS}[fe-lint] ${output}\n"
output=$(cd "$PROJECT_DIR/frontend" && bun run type-check 2>&1) || ERRORS="${ERRORS}[fe-types] ${output}\n"
# `.test.` filters to unit tests only — Playwright specs are `.spec.ts` and must
# not be collected by bun test (they crash on test() setup). bun's bunfig
# pathIgnorePatterns is unreliable, so the CLI filter is the authoritative guard.
output=$(cd "$PROJECT_DIR/frontend" && bun test .test. 2>&1) || ERRORS="${ERRORS}[fe-test] ${output}\n"
output=$(cd "$PROJECT_DIR" && make check-arch 2>&1) || ERRORS="${ERRORS}[arch] ${output}\n"

if [ -n "$ERRORS" ]; then
    printf "%b" "$ERRORS" | head -20
    exit 1
fi
