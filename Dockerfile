# ShareHODL Blockchain Dockerfile
# Multi-stage build for optimized production image

# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    gcc \
    musl-dev \
    linux-headers

# Set working directory
WORKDIR /src

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN make build

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    jq \
    curl \
    bash

# Create sharehodl user
RUN adduser -D -s /bin/bash sharehodl

# Set working directory
WORKDIR /home/sharehodl

# Copy binary from builder stage
COPY --from=builder /src/build/sharehodld /usr/local/bin/sharehodld

# Copy scripts
COPY --from=builder /src/scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
COPY --from=builder /src/scripts/healthcheck.sh /usr/local/bin/healthcheck.sh

# Make scripts executable
RUN chmod +x /usr/local/bin/entrypoint.sh /usr/local/bin/healthcheck.sh

# Create data directory
RUN mkdir -p /home/sharehodl/.sharehodl && \
    chown -R sharehodl:sharehodl /home/sharehodl

# Switch to sharehodl user
USER sharehodl

# Expose ports
# P2P port
EXPOSE 26656
# RPC port
EXPOSE 26657
# API port
EXPOSE 1317
# gRPC port
EXPOSE 9090
# Prometheus metrics port
EXPOSE 26660

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD /usr/local/bin/healthcheck.sh

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Default command
CMD ["start"]