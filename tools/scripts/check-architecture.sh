#!/bin/bash
# Verify hexagonal architecture rules are not violated.
# Checks that domain layers do not import infrastructure packages
# and that modules do not import each other.

set -e

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
ERRORS=0

echo "Checking architecture rules..."

# Rule 1: Domain layer must not import infrastructure packages
echo "  [1/3] Domain layer imports..."
DOMAIN_VIOLATIONS=$(grep -rn '"github.com/jcsoftdev/pulzifi-back/modules/.*/infrastructure' "$PROJECT_DIR/modules/*/domain/" 2>/dev/null || true)
if [ -n "$DOMAIN_VIOLATIONS" ]; then
    echo "    FAIL: Domain importing infrastructure:"
    echo "$DOMAIN_VIOLATIONS"
    ERRORS=$((ERRORS + 1))
else
    echo "    OK"
fi

# Rule 2: No cross-module imports
echo "  [2/3] Cross-module imports..."
for module_dir in "$PROJECT_DIR"/modules/*/; do
    module_name=$(basename "$module_dir")
    CROSS_IMPORTS=$(grep -rn '"github.com/jcsoftdev/pulzifi-back/modules/' "$module_dir" 2>/dev/null | grep -v "modules/$module_name/" || true)
    if [ -n "$CROSS_IMPORTS" ]; then
        echo "    FAIL: $module_name imports other modules:"
        echo "$CROSS_IMPORTS"
        ERRORS=$((ERRORS + 1))
    fi
done
if [ $ERRORS -eq 0 ]; then
    echo "    OK"
fi

# Rule 3: shared/ has no business logic imports
echo "  [3/3] Shared package independence..."
SHARED_VIOLATIONS=$(grep -rn '"github.com/jcsoftdev/pulzifi-back/modules/' "$PROJECT_DIR/shared/" 2>/dev/null || true)
if [ -n "$SHARED_VIOLATIONS" ]; then
    echo "    FAIL: shared/ imports modules:"
    echo "$SHARED_VIOLATIONS"
    ERRORS=$((ERRORS + 1))
else
    echo "    OK"
fi

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAILED: $ERRORS architecture violation(s) found."
    exit 1
else
    echo ""
    echo "OK: All architecture rules pass."
fi
