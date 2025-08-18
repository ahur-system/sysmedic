# SysMedic Dockerfile
# Multi-stage build for optimized production image

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s -extldflags '-static'" \
    -a -installsuffix cgo \
    -o sysmedic ./cmd/sysmedic

# Production stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && update-ca-certificates

# Create non-root user (though SysMedic needs root for system monitoring)
RUN addgroup -g 1001 sysmedic && \
    adduser -D -s /bin/sh -u 1001 -G sysmedic sysmedic

# Create directories
RUN mkdir -p /etc/sysmedic /var/lib/sysmedic && \
    chown -R sysmedic:sysmedic /etc/sysmedic /var/lib/sysmedic

# Copy binary from builder
COPY --from=builder /app/sysmedic /usr/local/bin/sysmedic
RUN chmod +x /usr/local/bin/sysmedic

# Copy configuration
COPY scripts/config.example.yaml /etc/sysmedic/config.yaml

# Set environment variables
ENV SYSMEDIC_CONFIG=/etc/sysmedic/config.yaml
ENV SYSMEDIC_DATA=/var/lib/sysmedic

# Expose any ports if needed (none for current implementation)
# EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD /usr/local/bin/sysmedic daemon status || exit 1

# Volume for persistent data
VOLUME ["/var/lib/sysmedic", "/etc/sysmedic"]

# Note: Container needs privileged access to monitor host system
# Run with: docker run --privileged --pid=host --net=host -v /proc:/host/proc:ro -v /sys:/host/sys:ro

# Switch to root for system monitoring (required for /proc access)
USER root

# Default command
CMD ["/usr/local/bin/sysmedic", "daemon", "start"]
