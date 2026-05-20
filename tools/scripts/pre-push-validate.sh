#!/bin/bash
# Global pre-push gate. Each check runs to completion; failures collected.
# Output: only failed sections, prefixed with [name], capped at 30 lines.

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
ERRORS=""

run() {
    local name="$1"
    shift
    local output
    output=$("$@" 2>&1) || ERRORS="${ERRORS}[${name}] ${output}\n"
}

run "go-test"      go test ./...
run "arch"         bash "$PROJECT_DIR/tools/scripts/check-architecture.sh"
run "fe-test"      bash -c "cd '$PROJECT_DIR/frontend/apps/web' && bun test features/cms"
run "fe-types"     bash -c "cd '$PROJECT_DIR/frontend' && bun run type-check"
run "fe-lint"      bash -c "cd '$PROJECT_DIR/frontend' && bun run lint"

if [ -n "$ERRORS" ]; then
    printf "%b" "$ERRORS" | head -30
    exit 1
fi
