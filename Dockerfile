# ==========================================
# Dockerfile for Farmers Module
# Usage: docker build -t farmers-module .
# Context: Root of farmers-module folder
# ==========================================

# ---------------------------------
# Stage 1: Build
# ---------------------------------
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependencies first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# Main entrypoint is in cmd/farmers-service/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server cmd/farmers-service/main.go

# ---------------------------------
# Stage 2: Runtime
# ---------------------------------
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Copy binary from builder
COPY --from=builder /bin/server .

# Copy migration files if needed
# COPY migrations ./migrations

# Expose port
EXPOSE 8000

# Healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8000/health || exit 1

# Run the server
CMD ["./server"]
