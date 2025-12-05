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
COPY apps/*/package.json ./apps/*/
COPY packages/*/package.json ./packages/*/

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy source code
COPY . .

# Build all applications
RUN pnpm build

# Production stage
FROM node:20-alpine AS runner

# Install pnpm
RUN npm install -g pnpm

# Set working directory
WORKDIR /app

# Copy package files for runtime dependencies
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/*/package.json ./apps/*/
COPY packages/*/package.json ./packages/*/

# Install only production dependencies
RUN pnpm install --frozen-lockfile --prod

# Copy built applications
COPY --from=builder /app/apps ./apps
COPY --from=builder /app/packages ./packages

# Create startup script
RUN cat > start-services.sh << 'EOF'
#!/bin/sh

# Start all Next.js applications in parallel
pnpm --filter home start --port 3000 &
pnpm --filter governance start --port 3001 &
pnpm --filter trading start --port 3002 &
pnpm --filter explorer start --port 3003 &
pnpm --filter wallet start --port 3004 &
pnpm --filter business start --port 3005 &

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