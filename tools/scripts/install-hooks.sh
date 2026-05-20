#!/bin/bash
# One-time per clone: point git at the tracked .githooks directory.
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_DIR"

git config core.hooksPath .githooks
chmod +x .githooks/pre-commit .githooks/pre-push

echo "git hooks installed: core.hooksPath -> .githooks"
