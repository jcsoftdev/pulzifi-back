#!/bin/bash
# Watch script that regenerates Swagger docs on Go file changes
# Requires: fswatch (brew install fswatch)

PROJECT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
GENERATE_SCRIPT="$PROJECT_DIR/scripts/generate-swagger.sh"

echo "👀 Watching for Go file changes..."
echo "📝 When files change, Swagger docs will be regenerated automatically"
echo "🛑 Press Ctrl+C to stop watching"

# Watch all .go files and regenerate docs on change
fswatch -r "$PROJECT_DIR" --include='\.go$' | while read file; do
    echo "📝 Detected change in: $file"
    bash "$GENERATE_SCRIPT"
    echo "⏳ Watching for more changes..."
done
