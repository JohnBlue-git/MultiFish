# MultiFish Dockerfile
# Multi-stage build for optimized production image

# Build stage
ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS builder

# Set working directory
WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty)" \
    -o multifish

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 multifish && \
    adduser -D -u 1000 -G multifish multifish

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/multifish .

# Copy example configuration (optional)
COPY --chown=multifish:multifish config.example.yaml ./config.yaml

# Create necessary directories
RUN mkdir -p /app/logs && \
    chown -R multifish:multifish /app

# Switch to non-root user
USER multifish

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/MultiFish/v1 || exit 1

# Run the application
# Can be overridden with docker-compose command or docker run
ENTRYPOINT ["./multifish"]
CMD []
