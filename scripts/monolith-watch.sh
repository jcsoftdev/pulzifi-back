#!/bin/bash
# Hot reload script for monolith using fswatch
# Watches Go files and restarts the server on changes

set -e

PROJECT_DIR="${PROJECT_DIR:-.}"
BIN_PATH="$PROJECT_DIR/bin/pulzifi-monolith"
PID=""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Function to kill the server
kill_server() {
    if [ ! -z "$PID" ] && kill -0 $PID 2>/dev/null; then
        echo -e "${YELLOW}Stopping server (PID: $PID)...${NC}"
        kill $PID 2>/dev/null || true
        sleep 1
    fi
}

# Function to build and start server
rebuild_and_start() {
    echo -e "${YELLOW}Building monolith...${NC}"
    if go build -o "$BIN_PATH" ./cmd/server 2>&1; then
        echo -e "${GREEN}✓ Build successful${NC}"
        kill_server
        
        echo -e "${GREEN}Starting server...${NC}"
        "$BIN_PATH" &
        PID=$!
        echo -e "${GREEN}✓ Server started (PID: $PID)${NC}"
    else
        echo -e "${RED}✗ Build failed${NC}"
    fi
}

# Trap signals
trap 'kill_server; exit 0' SIGINT SIGTERM

echo -e "${GREEN}Starting hot reload watcher...${NC}"
echo -e "${YELLOW}Watching Go files for changes...${NC}"

# Initial build
rebuild_and_start

# Watch for changes
fswatch -o --exclude ".*" --include "\\.go$" ./cmd ./modules ./shared 2>/dev/null | while read; do
    clear
    rebuild_and_start
done
