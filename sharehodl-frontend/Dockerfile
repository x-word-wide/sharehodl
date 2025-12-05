# ShareHODL Frontend Dockerfile
# Multi-stage build for optimized production image

# Build stage
FROM node:20-alpine AS builder

# Install pnpm
RUN npm install -g pnpm

# Set working directory
WORKDIR /app

# Copy package files
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps ./apps
COPY packages ./packages

# Install dependencies
RUN pnpm install --frozen-lockfile

# Build all applications
RUN pnpm build

# Production stage
FROM node:20-alpine AS runner

# Install pnpm and curl
RUN npm install -g pnpm && apk add --no-cache curl

# Set working directory
WORKDIR /app

# Copy package files and source for runtime
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY --from=builder /app/apps ./apps
COPY --from=builder /app/packages ./packages

# Install ALL dependencies (dev mode needs dev dependencies)
RUN pnpm install --frozen-lockfile

# Create startup script
RUN cat > start-services.sh << 'EOF'
#!/bin/sh

echo "=========================================="
echo "Starting ShareHODL Services in Docker"
echo "=========================================="

# Function to start service and verify
start_service() {
    APP=$1
    PORT=$2
    echo "Starting $APP on port $PORT..."
    pnpm --filter $APP dev > /tmp/$APP.log 2>&1 &
    PID=$!
    echo $PID > /tmp/$APP.pid
    sleep 10
    if curl -f http://localhost:$PORT > /dev/null 2>&1; then
        echo "✅ $APP started successfully on port $PORT"
    else
        echo "❌ $APP failed to start on port $PORT"
    fi
}

# Start all services with verification
start_service "home" "3000"
start_service "governance" "3001"  
start_service "trading" "3002"
start_service "explorer" "3003"
start_service "wallet" "3004"
start_service "business" "3005"

echo ""
echo "All services started! Keeping container alive..."

# Keep container running
while true; do
    sleep 60
    # Check if all services are still running
    for app in home governance trading explorer wallet business; do
        if ! kill -0 $(cat /tmp/$app.pid) 2>/dev/null; then
            echo "Restarting $app..."
            case $app in
                home) start_service "home" "3000" ;;
                governance) start_service "governance" "3001" ;;
                trading) start_service "trading" "3002" ;;
                explorer) start_service "explorer" "3003" ;;
                wallet) start_service "wallet" "3004" ;;
                business) start_service "business" "3005" ;;
            esac
        fi
    done
done
EOF

RUN chmod +x start-services.sh

# Expose all ports
EXPOSE 3000 3001 3002 3003 3004 3005

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:3001 && curl -f http://localhost:3002 && curl -f http://localhost:3004

# Start all services
CMD ["./start-services.sh"]