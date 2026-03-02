#!/bin/bash
# Scaffold a new domain module with the standard hexagonal structure.
# Usage: ./tools/scripts/new-module.sh <module-name>

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <module-name>"
    echo "Example: $0 billing"
    exit 1
fi

MODULE_NAME="$1"
PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
MODULE_DIR="$PROJECT_DIR/modules/$MODULE_NAME"

if [ -d "$MODULE_DIR" ]; then
    echo "Error: Module '$MODULE_NAME' already exists at $MODULE_DIR"
    exit 1
fi

echo "Creating module: $MODULE_NAME"

# Create directory structure
mkdir -p "$MODULE_DIR/domain/entities"
mkdir -p "$MODULE_DIR/domain/repositories"
mkdir -p "$MODULE_DIR/domain/services"
mkdir -p "$MODULE_DIR/domain/errors"
mkdir -p "$MODULE_DIR/application"
mkdir -p "$MODULE_DIR/infrastructure/http"
mkdir -p "$MODULE_DIR/infrastructure/persistence"

# Create CLAUDE.md
cat > "$MODULE_DIR/CLAUDE.md" << EOF
# ${MODULE_NAME^} Module

## Responsibility

TODO: Describe this module's responsibility.

## Entities

TODO: List entities.

## Repository Interfaces

TODO: List repository interfaces.

## Routes

| Method | Path | Description |
|--------|------|-------------|
| TODO | TODO | TODO |

## Dependencies

TODO: List dependencies.

## Constraints

- Tenant-scoped
EOF

echo "  Created: $MODULE_DIR/"
echo "  Created: domain/{entities,repositories,services,errors}/"
echo "  Created: application/"
echo "  Created: infrastructure/{http,persistence}/"
echo "  Created: CLAUDE.md"
echo ""
echo "Next steps:"
echo "  1. Define entities in domain/entities/"
echo "  2. Define repository interfaces in domain/repositories/"
echo "  3. Create use cases in application/<use_case>/"
echo "  4. Implement HTTP routes in infrastructure/http/module.go"
echo "  5. Implement PostgreSQL repos in infrastructure/persistence/"
echo "  6. Register the module in cmd/server/modules.go"
echo "  7. Update CLAUDE.md with actual details"
