# Farmers Module - Multi-stage Docker Build
# Produces an optimized, secure container with minimal footprint

# ============================================================================
# Stage 1: Builder - Compile the Go application
# ============================================================================
ARG GO_VERSION=1.24.4
FROM golang:${GO_VERSION}-alpine AS builder

# Build arguments for metadata
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

# Install build dependencies
# - git: Required for go modules that reference git repositories
# - ca-certificates: SSL certificates for HTTPS connections
# - make: For running Makefile targets if needed
RUN apk add --no-cache git ca-certificates make tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first for better layer caching
# This allows Docker to cache the dependency download step
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Generate Swagger documentation
RUN if [ -f /usr/local/go/bin/swag ] || command -v swag > /dev/null 2>&1; then \
        swag init -g cmd/farmers-service/main.go || echo "Swagger generation skipped"; \
    else \
        echo "Swagger not installed, skipping documentation generation"; \
    fi

# Build the application
# - CGO_ENABLED=0: Build static binary without C dependencies
# - GOOS=linux: Target Linux operating system
# - -ldflags: Inject build metadata and strip debug symbols
#   - -s: Omit symbol table and debug information
#   - -w: Omit DWARF symbol table
#   - -X: Set variable values at link time
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w \
    -X 'main.Version=${VERSION}' \
    -X 'main.GitCommit=${GIT_COMMIT}' \
    -X 'main.BuildDate=${BUILD_DATE}'" \
    -o /build/farmers-server \
    cmd/farmers-service/main.go

# Verify the binary is executable and get its size
RUN ls -lh /build/farmers-server && \
    chmod +x /build/farmers-server

# ============================================================================
# Stage 2: Runtime - Minimal Alpine image with only the binary
# ============================================================================
FROM alpine:3.20 AS runtime

# Install runtime dependencies
# - ca-certificates: For HTTPS connections (database, external APIs)
# - tzdata: Timezone data for accurate timestamps
# - curl: For health checks
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user for security
# - nobody: Use the standard nobody user (UID 65534)
# - nogroup: Use the standard nogroup group (GID 65534)
RUN addgroup -g 65534 -S appgroup && \
    adduser -u 65534 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/farmers-server /app/farmers-server

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Create directory for logs (if needed)
RUN mkdir -p /app/logs && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
# 8000: HTTP REST API
# 8081: gRPC gateway (if enabled)
EXPOSE 8000 8081

# Health check configuration
# - interval: Check every 30 seconds
# - timeout: Wait 3 seconds for response
# - start-period: Wait 40 seconds before starting checks (for initialization)
# - retries: Mark unhealthy after 3 consecutive failures
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Set metadata labels
LABEL maintainer="Kisanlink Team" \
      version="${VERSION}" \
      git.commit="${GIT_COMMIT}" \
      build.date="${BUILD_DATE}" \
      description="Farmers Module - Farm Management Service"

# Entry point
ENTRYPOINT ["/app/farmers-server"]
