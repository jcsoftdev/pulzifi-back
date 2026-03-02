#!/bin/bash
# Run all validation checks before committing or deploying.
# Exits with non-zero if any check fails.

set -e

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
ERRORS=0

echo "=== Build Validation ==="
echo ""

# Go backend checks
echo "[1/5] Go build..."
if go build ./... 2>&1; then
    echo "  OK"
else
    echo "  FAIL"
    ERRORS=$((ERRORS + 1))
fi

echo "[2/5] Go vet..."
if go vet ./... 2>&1; then
    echo "  OK"
else
    echo "  FAIL"
    ERRORS=$((ERRORS + 1))
fi

echo "[3/5] Go tests..."
if go test ./... 2>&1; then
    echo "  OK"
else
    echo "  FAIL"
    ERRORS=$((ERRORS + 1))
fi

# Frontend checks
echo "[4/5] Frontend type-check..."
if cd "$PROJECT_DIR/frontend" && bun run type-check 2>&1; then
    echo "  OK"
else
    echo "  FAIL"
    ERRORS=$((ERRORS + 1))
fi
cd "$PROJECT_DIR"

echo "[5/5] Architecture rules..."
if bash "$PROJECT_DIR/tools/scripts/check-architecture.sh" 2>&1; then
    echo "  OK"
else
    echo "  FAIL"
    ERRORS=$((ERRORS + 1))
fi

echo ""
if [ $ERRORS -gt 0 ]; then
    echo "FAILED: $ERRORS check(s) failed."
    exit 1
else
    echo "ALL CHECKS PASSED."
fi
