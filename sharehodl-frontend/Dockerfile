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

# Install pnpm
RUN npm install -g pnpm

# Set working directory
WORKDIR /app

# Copy package files and source for runtime
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY --from=builder /app/apps ./apps
COPY --from=builder /app/packages ./packages

# Install only production dependencies
RUN pnpm install --frozen-lockfile --prod

# Create startup script
RUN cat > start-services.sh << 'EOF'
#!/bin/sh

echo "Starting all ShareHODL frontend services in development mode..."

# Start all Next.js applications in development mode (works better than production)
echo "Starting Home on port 3000..."
pnpm --filter home dev > /tmp/home.log 2>&1 &

echo "Starting Governance on port 3001..."
pnpm --filter governance dev > /tmp/governance.log 2>&1 &

echo "Starting Trading on port 3002..."
pnpm --filter trading dev > /tmp/trading.log 2>&1 &

echo "Starting Explorer on port 3003..."
pnpm --filter explorer dev > /tmp/explorer.log 2>&1 &

echo "Starting Wallet on port 3004..."
pnpm --filter wallet dev > /tmp/wallet.log 2>&1 &

echo "Starting Business on port 3005..."
pnpm --filter business dev > /tmp/business.log 2>&1 &

echo "All services started in development mode! Logs available in /tmp/"

# Wait for all background processes
wait
EOF

RUN chmod +x start-services.sh

# Expose all ports
EXPOSE 3000 3001 3002 3003 3004 3005

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:3001 && curl -f http://localhost:3002 && curl -f http://localhost:3004

# Start all services
CMD ["./start-services.sh"]