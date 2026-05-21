#!/bin/bash
# Pre-commit validation. Only outputs errors to keep Claude context minimal.

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
ERRORS=""

output=$(go build ./... 2>&1) || ERRORS="${ERRORS}[build] ${output}\n"
output=$(go vet ./... 2>&1) || ERRORS="${ERRORS}[vet] ${output}\n"
output=$(cd "$PROJECT_DIR/frontend" && bun run type-check 2>&1) || ERRORS="${ERRORS}[types] ${output}\n"
output=$(cd "$PROJECT_DIR" && make check-arch 2>&1) || ERRORS="${ERRORS}[arch] ${output}\n"

if [ -n "$ERRORS" ]; then
    printf "%b" "$ERRORS" | head -20
    exit 1
fi
