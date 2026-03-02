#!/bin/bash
# Create a new database migration file pair (up + down).
# Usage: ./tools/scripts/new-migration.sh <scope> <description>
#   scope: "public" or "tenant"
#   description: snake_case description (e.g., add_billing_table)

set -e

if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Usage: $0 <scope> <description>"
    echo "  scope: public | tenant"
    echo "  description: snake_case migration name"
    echo ""
    echo "Example: $0 tenant add_billing_table"
    exit 1
fi

SCOPE="$1"
DESCRIPTION="$2"
PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
MIGRATIONS_DIR="$PROJECT_DIR/shared/database/migrations/$SCOPE"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "Error: Invalid scope '$SCOPE'. Use 'public' or 'tenant'."
    exit 1
fi

# Find the next sequence number
LAST_NUM=$(ls "$MIGRATIONS_DIR"/*.up.sql 2>/dev/null | sort -V | tail -1 | grep -oP '\d{6}' | head -1)
if [ -z "$LAST_NUM" ]; then
    NEXT_NUM="000001"
else
    NEXT_NUM=$(printf "%06d" $((10#$LAST_NUM + 1)))
fi

UP_FILE="$MIGRATIONS_DIR/${NEXT_NUM}_${DESCRIPTION}.up.sql"
DOWN_FILE="$MIGRATIONS_DIR/${NEXT_NUM}_${DESCRIPTION}.down.sql"

cat > "$UP_FILE" << EOF
-- Migration: $DESCRIPTION
-- Scope: $SCOPE
-- Created: $(date -u +"%Y-%m-%dT%H:%M:%SZ")

-- TODO: Write your migration SQL here
EOF

cat > "$DOWN_FILE" << EOF
-- Rollback: $DESCRIPTION
-- Scope: $SCOPE

-- TODO: Write your rollback SQL here
EOF

echo "Created migration files:"
echo "  UP:   $UP_FILE"
echo "  DOWN: $DOWN_FILE"
