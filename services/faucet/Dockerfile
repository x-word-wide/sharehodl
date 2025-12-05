FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git build-base

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy faucet source
COPY services/faucet/ ./services/faucet/
COPY x/ ./x/

# Build faucet binary
WORKDIR /app/services/faucet
RUN CGO_ENABLED=0 GOOS=linux go build -o faucet .

# Final stage
FROM alpine:latest

# Install required packages
RUN apk add --no-cache ca-certificates curl

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/services/faucet/faucet /usr/local/bin/faucet

# Create data directory
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Default command
CMD ["faucet"]