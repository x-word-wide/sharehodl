#!/bin/bash

# ShareHODL SSL Domain Setup Script
# Sets up all subdomains with SSL certificates and auto-renewal

set -e

log_info() {
    echo -e "\033[0;34m[INFO]\033[0m $1"
}

log_success() {
    echo -e "\033[0;32m[SUCCESS]\033[0m $1"
}

log_error() {
    echo -e "\033[0;31m[ERROR]\033[0m $1"
}

# Domain configuration
MAIN_DOMAIN="sharehodl.com"
SUBDOMAINS=(
    "scan.sharehodl.com"      # Explorer
    "trade.sharehodl.com"     # Trading Platform  
    "gov.sharehodl.com"       # Governance Portal
    "business.sharehodl.com"  # Business Portal
    "wallet.sharehodl.com"    # Wallet Interface
    "api.sharehodl.com"       # API Endpoint
    "rpc.sharehodl.com"       # RPC Endpoint
)

log_info "Setting up SSL for ShareHODL domains..."
log_info "Main domain: $MAIN_DOMAIN"
log_info "Subdomains: ${SUBDOMAINS[*]}"

# Install Certbot if not already installed
if ! command -v certbot &> /dev/null; then
    log_info "Installing Certbot..."
    apt update
    apt install -y certbot python3-certbot-nginx
fi

# Stop nginx temporarily for certificate generation
log_info "Stopping Nginx for certificate generation..."
systemctl stop nginx

# Create comprehensive Nginx configuration
log_info "Creating Nginx configuration for all domains..."
cat > /etc/nginx/sites-available/sharehodl-ssl << EOF
# ShareHODL Main Domain
server {
    listen 80;
    server_name sharehodl.com www.sharehodl.com;
    return 301 https://sharehodl.com\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name sharehodl.com www.sharehodl.com;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/sharehodl.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/sharehodl.com/privkey.pem;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Main frontend (governance by default)
    location / {
        proxy_pass http://localhost:3001;
        include /etc/nginx/proxy_params;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Health check for ALB
    location /health {
        access_log off;
        return 200 "healthy\\n";
        add_header Content-Type text/plain;
    }
}

# Explorer Subdomain
server {
    listen 80;
    server_name scan.sharehodl.com;
    return 301 https://scan.sharehodl.com\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name scan.sharehodl.com;

    ssl_certificate /etc/letsencrypt/live/scan.sharehodl.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/scan.sharehodl.com/privkey.pem;
    
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    location / {
        proxy_pass http://localhost:3003;
        include /etc/nginx/proxy_params;
    }
}

# Trading Subdomain
server {
    listen 80;
    server_name trade.sharehodl.com;
    return 301 https://trade.sharehodl.com\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name trade.sharehodl.com;

    ssl_certificate /etc/letsencrypt/live/trade.sharehodl.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/trade.sharehodl.com/privkey.pem;
    
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    location / {
        proxy_pass http://localhost:3002;
        include /etc/nginx/proxy_params;
    }
}

# Governance Subdomain
server {
    listen 80;
    server_name gov.sharehodl.com;
    return 301 https://gov.sharehodl.com\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name gov.sharehodl.com;

    ssl_certificate /etc/letsencrypt/live/gov.sharehodl.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/gov.sharehodl.com/privkey.pem;
    
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    location / {
        proxy_pass http://localhost:3001;
        include /etc/nginx/proxy_params;
    }
}

# Business Subdomain
server {
    listen 80;
    server_name business.sharehodl.com;
    return 301 https://business.sharehodl.com\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name business.sharehodl.com;

    ssl_certificate /etc/letsencrypt/live/business.sharehodl.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/business.sharehodl.com/privkey.pem;
    
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    location / {
        proxy_pass http://localhost:3005;
        include /etc/nginx/proxy_params;
    }
}

# Wallet Subdomain
server {
    listen 80;
    server_name wallet.sharehodl.com;
    return 301 https://wallet.sharehodl.com\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name wallet.sharehodl.com;

    ssl_certificate /etc/letsencrypt/live/wallet.sharehodl.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/wallet.sharehodl.com/privkey.pem;
    
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    location / {
        proxy_pass http://localhost:3004;
        include /etc/nginx/proxy_params;
    }
}

# API Subdomain
server {
    listen 80;
    server_name api.sharehodl.com;
    return 301 https://api.sharehodl.com\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.sharehodl.com;

    ssl_certificate /etc/letsencrypt/live/api.sharehodl.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.sharehodl.com/privkey.pem;
    
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    location / {
        proxy_pass http://localhost:1317;
        include /etc/nginx/proxy_params;
        
        # CORS headers for API
        add_header Access-Control-Allow-Origin "*" always;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
        add_header Access-Control-Allow-Headers "DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization" always;
        
        if (\$request_method = 'OPTIONS') {
            add_header Access-Control-Allow-Origin "*";
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
            add_header Access-Control-Allow-Headers "DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization";
            add_header Access-Control-Max-Age 1728000;
            add_header Content-Type 'text/plain; charset=utf-8';
            add_header Content-Length 0;
            return 204;
        }
    }
}

# RPC Subdomain  
server {
    listen 80;
    server_name rpc.sharehodl.com;
    return 301 https://rpc.sharehodl.com\$request_uri;
}

server {
    listen 443 ssl http2;
    server_name rpc.sharehodl.com;

    ssl_certificate /etc/letsencrypt/live/rpc.sharehodl.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/rpc.sharehodl.com/privkey.pem;
    
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    location / {
        proxy_pass http://localhost:26657;
        include /etc/nginx/proxy_params;
        
        # WebSocket support for RPC
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # CORS headers for RPC
        add_header Access-Control-Allow-Origin "*" always;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
        add_header Access-Control-Allow-Headers "DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization" always;
    }
}
EOF

# Generate SSL certificates for all domains
log_info "Generating SSL certificates..."

# Main domain
log_info "Getting certificate for main domain: $MAIN_DOMAIN"
certbot certonly --standalone -d sharehodl.com -d www.sharehodl.com \
    --email admin@sharehodl.com \
    --agree-tos \
    --non-interactive

# Subdomains
for subdomain in "${SUBDOMAINS[@]}"; do
    log_info "Getting certificate for: $subdomain"
    certbot certonly --standalone -d $subdomain \
        --email admin@sharehodl.com \
        --agree-tos \
        --non-interactive
done

# Enable the new site configuration
log_info "Enabling SSL configuration..."
ln -sf /etc/nginx/sites-available/sharehodl-ssl /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default
rm -f /etc/nginx/sites-enabled/sharehodl

# Test nginx configuration
log_info "Testing Nginx configuration..."
nginx -t

# Start nginx
log_info "Starting Nginx..."
systemctl start nginx
systemctl enable nginx

# Create SSL renewal script
log_info "Setting up SSL auto-renewal..."
cat > /etc/cron.d/sharehodl-ssl-renewal << EOF
# ShareHODL SSL Certificate Auto-Renewal
# Runs twice daily to check for certificate renewal
SHELL=/bin/sh
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

# Check for renewal twice daily
0 2,14 * * * root /usr/bin/certbot renew --quiet --nginx --post-hook "systemctl reload nginx"
EOF

# Create renewal test script
cat > /usr/local/bin/test-ssl-renewal.sh << 'EOF'
#!/bin/bash
# Test SSL renewal process
echo "Testing SSL certificate renewal..."
certbot renew --dry-run
if [ $? -eq 0 ]; then
    echo "SSL renewal test PASSED ✓"
else
    echo "SSL renewal test FAILED ✗"
    exit 1
fi
EOF

chmod +x /usr/local/bin/test-ssl-renewal.sh

# Test the renewal process
log_info "Testing SSL renewal process..."
/usr/local/bin/test-ssl-renewal.sh

log_success "SSL setup completed successfully!"
log_info ""
log_info "ShareHODL is now accessible at:"
log_info "  Main Site: https://sharehodl.com"
log_info "  Explorer: https://scan.sharehodl.com"
log_info "  Trading: https://trade.sharehodl.com"
log_info "  Governance: https://gov.sharehodl.com"
log_info "  Business: https://business.sharehodl.com"
log_info "  Wallet: https://wallet.sharehodl.com"
log_info "  API: https://api.sharehodl.com"
log_info "  RPC: https://rpc.sharehodl.com"
log_info ""
log_info "SSL Features:"
log_info "  ✓ HTTPS enforced for all domains"
log_info "  ✓ HTTP automatically redirects to HTTPS"
log_info "  ✓ HSTS security headers enabled"
log_info "  ✓ Auto-renewal configured (runs twice daily)"
log_info "  ✓ Renewal test passed"
log_info ""
log_info "Cron job installed: /etc/cron.d/sharehodl-ssl-renewal"
log_info "Test script: /usr/local/bin/test-ssl-renewal.sh"