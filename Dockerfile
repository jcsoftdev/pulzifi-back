# ============================================================
# Multi-stage Docker build for Pulzifi Backend
# Uses Go 1.25 and optimized for production deployment
# ============================================================

# ============================================================
# Stage 1: Build stage
# ============================================================
FROM golang:1.25-bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates build-essential librdkafka-dev

# Install air for live reloading
RUN go install github.com/air-verse/air@latest

# Set PATH to include go binaries
ENV PATH=$PATH:/root/go/bin
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Copy source code
COPY . .

# Download dependencies
RUN go mod download && go mod verify

# Build arguments for module selection
ARG MODULE_NAME
ENV MODULE_NAME=${MODULE_NAME}

# Build the specific module
# We don't build the binary in dev, air will do it

# ============================================================
# Stage 2: Runtime stage
# ============================================================
FROM golang:1.25-bookworm

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl librdkafka-dev && rm -rf /var/lib/apt/lists/*

# Install air for live reloading
RUN go install github.com/air-verse/air@latest

# Set PATH to include go binaries
ENV PATH=$PATH:/root/go/bin

# Create non-root user
RUN addgroup --system --gid 1001 appgroup && \
    adduser --system --uid 1001 --ingroup appgroup appuser

# Set working directory
WORKDIR /app

# Copy binary and air from builder stage
COPY --from=builder /go/bin/air /go/bin/air
COPY --from=builder /app/modules /app/modules
COPY --from=builder /app/go.mod /app/go.mod
COPY --from=builder /app/go.sum /app/
COPY --from=builder /app/shared /app/shared

# Copy any necessary config files
COPY --from=builder /app/.env.example .env.example

# Change ownership to root user (air needs to rebuild)
RUN chown -R root:root /app

# NOTE: Running as root for development with air hot-reload
# In production, use a separate slim image

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${HTTP_PORT:-8080}/health || exit 1

# Expose ports (will be overridden by docker-compose)
EXPOSE 8080 9000

# Run the application - air from the module directory if MODULE_NAME is set
CMD sh -c 'if [ -n "$MODULE_NAME" ]; then cd modules/$MODULE_NAME && air; else air; fi'