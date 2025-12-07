#!/bin/bash
# Script to generate Swagger documentation for the monolith
# Used by Makefile targets to ensure docs are always up to date

set -e

# Get project root (should be called from root via make)
PROJECT_DIR="${PROJECT_DIR:-.}"
SWAG_CMD="$(go env GOPATH)/bin/swag"
DOCS_OUTPUT_DIR="$PROJECT_DIR/docs"

# Ensure swag is installed
if [ ! -f "$SWAG_CMD" ]; then
    echo "📦 Installing swag CLI..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Generate swagger docs from project root
echo "🔄 Generating Swagger documentation..."
cd "$PROJECT_DIR"
"$SWAG_CMD" init -g ./cmd/server/main.go --output "$DOCS_OUTPUT_DIR" --quiet

# Check if docs were generated
if [ -f "$DOCS_OUTPUT_DIR/docs.go" ]; then
    echo "✅ Swagger docs generated successfully at $DOCS_OUTPUT_DIR"
    echo "📄 Swagger UI available at: http://localhost:8080/swagger/index.html"
else
    echo "⚠️  Warning: Swagger docs may not have been generated properly"
    exit 1
fi
