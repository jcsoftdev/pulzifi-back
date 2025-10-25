# ============================================================
# Multi-stage Docker build for Pulzifi Backend
# Uses Go 1.25 and optimized for production deployment
# ============================================================

# ============================================================
# Stage 1: Build stage
# ============================================================
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments for module selection
ARG MODULE_NAME
ENV MODULE_NAME=${MODULE_NAME}

# Build the specific module
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o app ./modules/${MODULE_NAME}/main.go

# ============================================================
# Stage 2: Runtime stage
# ============================================================
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/app .

# Copy any necessary config files
COPY --from=builder /app/.env.example .env.example

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${HTTP_PORT:-8080}/health || exit 1

# Expose ports (will be overridden by docker-compose)
EXPOSE 8080 9000

# Run the application
CMD ["./app"]