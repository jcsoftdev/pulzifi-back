#!/usr/bin/env bash
# Enforce minimum Go coverage floor.
# Usage: COVERAGE_FLOOR=15 ./tools/scripts/coverage-gate.sh coverage.out
set -euo pipefail

PROFILE="${1:-coverage.out}"
FLOOR="${COVERAGE_FLOOR:-15}"

if [ ! -f "$PROFILE" ]; then
  echo "coverage-gate: profile not found: $PROFILE" >&2
  exit 2
fi

ACTUAL=$(go tool cover -func="$PROFILE" | awk '/^total:/{gsub("%","",$3); print $3}')

if [ -z "$ACTUAL" ]; then
  echo "coverage-gate: failed to parse total coverage from $PROFILE" >&2
  exit 2
fi

awk -v a="$ACTUAL" -v f="$FLOOR" 'BEGIN { exit !(a+0 >= f+0) }'
RC=$?

if [ "$RC" -eq 0 ]; then
  printf "coverage-gate: OK — %s%% >= %s%% floor\n" "$ACTUAL" "$FLOOR"
  exit 0
fi

printf "coverage-gate: FAIL — %s%% < %s%% floor\n" "$ACTUAL" "$FLOOR" >&2
exit 1
