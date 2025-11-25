# =============================================================================
# GoConnect Server - Production Dockerfile
# Multi-stage build for minimal image size
# =============================================================================

# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY server/go.mod server/go.sum ./
RUN go mod download

# Copy source code
COPY server/ ./

# Build arguments for version info
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}" \
    -o /goconnect-server \
    ./cmd/server/main.go

# Stage 2: Runtime
FROM alpine:3.19

# Labels
LABEL org.opencontainers.image.title="GoConnect Server"
LABEL org.opencontainers.image.description="GoConnect VPN Management Server"
LABEL org.opencontainers.image.vendor="orhaniscoding"
LABEL org.opencontainers.image.source="https://github.com/orhaniscoding/goconnect"
LABEL org.opencontainers.image.licenses="MIT"

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wireguard-tools \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 goconnect && \
    adduser -D -u 1000 -G goconnect goconnect

# Create necessary directories
RUN mkdir -p /var/log/goconnect /etc/goconnect /data && \
    chown -R goconnect:goconnect /var/log/goconnect /etc/goconnect /data

# Copy binary from builder
COPY --from=builder /goconnect-server /usr/local/bin/goconnect-server

# Copy migrations if they exist
COPY --from=builder /build/migrations /migrations

# Set working directory
WORKDIR /app

# Environment variables with secure defaults
ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080 \
    ENVIRONMENT=production \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_NAME=goconnect \
    DB_USER=goconnect \
    DB_SSLMODE=require \
    REDIS_HOST=redis \
    REDIS_PORT=6379

# Expose ports
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run as non-root user
USER goconnect

# Entry point
ENTRYPOINT ["/usr/local/bin/goconnect-server"]
