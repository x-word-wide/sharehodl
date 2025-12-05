# ShareHODL Domain & SSL Setup Guide

## 🌐 Domain Configuration

### DNS Records to Add

Point these domains to your EC2 instance: **54.198.239.167**

```dns
# A Records
sharehodl.com                A    54.198.239.167
www.sharehodl.com           A    54.198.239.167
scan.sharehodl.com          A    54.198.239.167
trade.sharehodl.com         A    54.198.239.167
gov.sharehodl.com           A    54.198.239.167
business.sharehodl.com      A    54.198.239.167
wallet.sharehodl.com        A    54.198.239.167
api.sharehodl.com           A    54.198.239.167
rpc.sharehodl.com           A    54.198.239.167
```

### Subdomain Purpose

| Domain | Purpose | Frontend Port |
|--------|---------|---------------|
| `sharehodl.com` | Main portal (Governance) | 3001 |
| `scan.sharehodl.com` | Blockchain Explorer | 3003 |
| `trade.sharehodl.com` | Trading Platform | 3002 |
| `gov.sharehodl.com` | Governance Portal | 3001 |
| `business.sharehodl.com` | Business Portal | 3005 |
| `wallet.sharehodl.com` | Wallet Interface | 3004 |
| `api.sharehodl.com` | REST API | 1317 |
| `rpc.sharehodl.com` | RPC Endpoint | 26657 |

## 🔒 SSL Setup

### Automatic SSL Setup

Once your DNS records are propagated (5-30 minutes), run:

```bash
# SSH into your EC2 instance
ssh -i ~/.ssh/sharehodl-key-1764931859.pem ubuntu@54.198.239.167

# Download and run SSL setup
wget https://raw.githubusercontent.com/sharehodl/sharehodl-blockchain/main/scripts/setup-ssl-domains.sh
chmod +x setup-ssl-domains.sh
sudo ./setup-ssl-domains.sh
```

### What the Script Does

✅ **Domain Configuration**
- Configures Nginx for all subdomains
- Sets up proper proxy routing
- Enables HTTP/2 for better performance

✅ **SSL Certificates**  
- Generates Let's Encrypt certificates for all domains
- Forces HTTPS redirects from HTTP
- Adds security headers (HSTS, XSS protection)

✅ **Auto-Renewal**
- Sets up cron job for automatic renewal
- Runs twice daily (2 AM & 2 PM UTC)
- Tests renewal process to ensure it works

✅ **Security Features**
- HSTS (HTTP Strict Transport Security)
- X-Frame-Options: DENY
- X-Content-Type-Options: nosniff
- X-XSS-Protection: enabled
- Referrer-Policy: strict-origin-when-cross-origin

## 🔧 Manual SSL Management

### Check Certificate Status
```bash
sudo certbot certificates
```

### Force Renewal
```bash
sudo certbot renew --force-renewal
```

### Test Renewal Process
```bash
sudo /usr/local/bin/test-ssl-renewal.sh
```

### View Cron Job
```bash
sudo cat /etc/cron.d/sharehodl-ssl-renewal
```

## 🌐 After SSL Setup

### Access Your Secure ShareHODL Platform

- **Main Portal**: https://sharehodl.com
- **Explorer**: https://scan.sharehodl.com  
- **Trading**: https://trade.sharehodl.com
- **Governance**: https://gov.sharehodl.com
- **Business**: https://business.sharehodl.com
- **Wallet**: https://wallet.sharehodl.com
- **API**: https://api.sharehodl.com/node_info
- **RPC**: https://rpc.sharehodl.com/status

### SSL Health Checks

All domains automatically redirect HTTP → HTTPS:
```bash
curl -I http://sharehodl.com
# Returns: 301 Moved Permanently
# Location: https://sharehodl.com

curl -I https://sharehodl.com  
# Returns: 200 OK with security headers
```

## 🔐 Certificate Details

### Certificate Validity
- **Issuer**: Let's Encrypt Authority X3
- **Validity**: 90 days (auto-renewed every 60 days)
- **Encryption**: RSA 2048-bit / ECDSA P-256
- **Protocol**: TLS 1.2, TLS 1.3

### Auto-Renewal Schedule
```cron
# Runs at 2 AM and 2 PM daily
0 2,14 * * * root /usr/bin/certbot renew --quiet --nginx --post-hook "systemctl reload nginx"
```

### Certificate Paths
```
/etc/letsencrypt/live/sharehodl.com/fullchain.pem
/etc/letsencrypt/live/sharehodl.com/privkey.pem
/etc/letsencrypt/live/scan.sharehodl.com/fullchain.pem
[... for each subdomain]
```

## 🚨 Troubleshooting

### DNS Not Propagated
```bash
# Check DNS propagation
nslookup sharehodl.com
dig +short sharehodl.com
```

### SSL Certificate Issues
```bash
# Check certificate status
sudo certbot certificates

# View Nginx error logs
sudo tail -f /var/log/nginx/error.log

# Test SSL configuration
sudo nginx -t
```

### Auto-Renewal Not Working
```bash
# Test renewal manually
sudo certbot renew --dry-run

# Check cron logs
sudo journalctl -u cron.service -f
```

## 🌟 Professional SSL Features

✅ **A+ SSL Rating** (SSLLabs test)
✅ **HTTP/2 Support** for faster loading  
✅ **Perfect Forward Secrecy**
✅ **OCSP Stapling** for faster handshakes
✅ **Security Headers** for XSS/Clickjacking protection
✅ **HSTS Preload Ready** for maximum security
✅ **Auto-Renewal** with zero downtime
✅ **Monitoring** via cron job logs

Your ShareHODL platform will have enterprise-grade SSL security! 🔐