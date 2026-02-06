#!/bin/bash

# ShareHODL Server Hardening Script
# For Ubuntu 24.04 - Run as root on fresh server
# Usage: bash 01-server-hardening.sh <new_username> <ssh_port>

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Configuration
NEW_USER="${1:-sharehodl}"
SSH_PORT="${2:-2222}"
SERVER_IP="178.63.13.190"

echo "=============================================="
echo "  ShareHODL Server Hardening Script"
echo "  Ubuntu 24.04 - Security First Deployment"
echo "=============================================="
echo ""
log_info "New sudo user: $NEW_USER"
log_info "SSH Port: $SSH_PORT"
log_info "Server IP: $SERVER_IP"
echo ""

# Verify running as root
if [[ $EUID -ne 0 ]]; then
   log_error "This script must be run as root"
fi

# Step 1: Update System
log_info "Step 1: Updating system packages..."
apt update && apt upgrade -y
apt dist-upgrade -y
apt autoremove -y
log_success "System updated"

# Step 2: Create new sudo user
log_info "Step 2: Creating sudo user '$NEW_USER'..."
if id "$NEW_USER" &>/dev/null; then
    log_warning "User $NEW_USER already exists"
else
    # Generate a random password
    USER_PASSWORD=$(openssl rand -base64 24)

    # Create user with home directory
    useradd -m -s /bin/bash -G sudo "$NEW_USER"
    echo "$NEW_USER:$USER_PASSWORD" | chpasswd

    # Save password securely (will be deleted after noting)
    echo "$USER_PASSWORD" > /root/.sharehodl_user_password
    chmod 600 /root/.sharehodl_user_password

    log_success "User '$NEW_USER' created"
    log_warning "IMPORTANT: Password saved to /root/.sharehodl_user_password"
    log_warning "Note this password, then delete the file!"
    echo ""
    echo "==========================================="
    echo "  NEW USER CREDENTIALS"
    echo "  Username: $NEW_USER"
    echo "  Password: $USER_PASSWORD"
    echo "==========================================="
    echo ""
fi

# Step 3: Setup SSH Key Authentication
log_info "Step 3: Setting up SSH key authentication..."
mkdir -p /home/$NEW_USER/.ssh
chmod 700 /home/$NEW_USER/.ssh

# Create authorized_keys file
touch /home/$NEW_USER/.ssh/authorized_keys
chmod 600 /home/$NEW_USER/.ssh/authorized_keys
chown -R $NEW_USER:$NEW_USER /home/$NEW_USER/.ssh

log_warning "ADD YOUR SSH PUBLIC KEY to /home/$NEW_USER/.ssh/authorized_keys"
log_warning "Run: ssh-copy-id -p $SSH_PORT $NEW_USER@$SERVER_IP"
echo ""

# Step 4: Configure SSH Daemon
log_info "Step 4: Hardening SSH configuration..."
cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup

cat > /etc/ssh/sshd_config <<EOF
# ShareHODL Hardened SSH Configuration
# Generated: $(date)

# Basic Settings
Port $SSH_PORT
Protocol 2
AddressFamily inet

# Authentication
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
AuthorizedKeysFile .ssh/authorized_keys
PermitEmptyPasswords no
ChallengeResponseAuthentication no
UsePAM yes

# Security Settings
X11Forwarding no
PrintMotd no
PrintLastLog yes
TCPKeepAlive yes
UseDNS no
MaxAuthTries 3
MaxSessions 3
LoginGraceTime 30
ClientAliveInterval 300
ClientAliveCountMax 2
MaxStartups 3:50:10

# Allowed Users (add more as needed)
AllowUsers $NEW_USER

# Logging
SyslogFacility AUTH
LogLevel VERBOSE

# Subsystems
Subsystem sftp /usr/lib/openssh/sftp-server -f AUTHPRIV -l INFO
EOF

log_success "SSH hardened (Port: $SSH_PORT, Root disabled, Key-only auth)"

# Step 5: Configure Firewall (UFW)
log_info "Step 5: Configuring UFW firewall..."
apt install -y ufw

# Reset UFW
ufw --force reset

# Default policies
ufw default deny incoming
ufw default allow outgoing

# Allow SSH on new port
ufw allow $SSH_PORT/tcp comment 'SSH'

# ShareHODL Blockchain ports
ufw allow 26656/tcp comment 'Tendermint P2P'
ufw allow 26657/tcp comment 'Tendermint RPC'
ufw allow 1317/tcp comment 'REST API'
ufw allow 9090/tcp comment 'gRPC'

# Web ports
ufw allow 80/tcp comment 'HTTP'
ufw allow 443/tcp comment 'HTTPS'

# Enable UFW
ufw --force enable

log_success "Firewall configured"
ufw status verbose

# Step 6: Install and Configure Fail2ban
log_info "Step 6: Installing and configuring Fail2ban..."
apt install -y fail2ban

cat > /etc/fail2ban/jail.local <<EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3
ignoreip = 127.0.0.1/8 ::1

[sshd]
enabled = true
port = $SSH_PORT
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 86400
findtime = 600

[sshd-ddos]
enabled = true
port = $SSH_PORT
filter = sshd-ddos
logpath = /var/log/auth.log
maxretry = 6
bantime = 172800
findtime = 600

[nginx-http-auth]
enabled = true
filter = nginx-http-auth
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 3
bantime = 3600

[nginx-limit-req]
enabled = true
filter = nginx-limit-req
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 10
bantime = 7200
EOF

systemctl enable fail2ban
systemctl restart fail2ban
log_success "Fail2ban installed and configured"

# Step 7: Kernel Security Parameters
log_info "Step 7: Configuring kernel security parameters..."
cat > /etc/sysctl.d/99-sharehodl-security.conf <<EOF
# ShareHODL Security and Performance Parameters

# Network Security
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1
net.ipv4.tcp_syncookies = 1
net.ipv4.conf.all.accept_redirects = 0
net.ipv4.conf.default.accept_redirects = 0
net.ipv4.conf.all.send_redirects = 0
net.ipv4.conf.default.send_redirects = 0
net.ipv4.conf.all.accept_source_route = 0
net.ipv4.conf.default.accept_source_route = 0
net.ipv4.conf.all.log_martians = 1
net.ipv4.conf.default.log_martians = 1
net.ipv4.icmp_echo_ignore_broadcasts = 1
net.ipv4.icmp_ignore_bogus_error_responses = 1

# IPv6 (disable if not needed)
net.ipv6.conf.all.disable_ipv6 = 1
net.ipv6.conf.default.disable_ipv6 = 1

# Performance Tuning for Blockchain
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_fin_timeout = 15
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_keepalive_time = 300
net.ipv4.tcp_keepalive_probes = 5
net.ipv4.tcp_keepalive_intvl = 15
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216

# File Handles
fs.file-max = 2097152
fs.nr_open = 2097152

# Memory
vm.swappiness = 10
vm.dirty_ratio = 60
vm.dirty_background_ratio = 2
EOF

sysctl -p /etc/sysctl.d/99-sharehodl-security.conf
log_success "Kernel parameters configured"

# Step 8: Set up file limits
log_info "Step 8: Configuring file limits..."
cat >> /etc/security/limits.conf <<EOF

# ShareHODL Limits
* soft nofile 65535
* hard nofile 65535
* soft nproc 65535
* hard nproc 65535
$NEW_USER soft nofile 65535
$NEW_USER hard nofile 65535
EOF

log_success "File limits configured"

# Step 9: Automatic Security Updates
log_info "Step 9: Configuring automatic security updates..."
apt install -y unattended-upgrades apt-listchanges

cat > /etc/apt/apt.conf.d/50unattended-upgrades <<EOF
Unattended-Upgrade::Allowed-Origins {
    "\${distro_id}:\${distro_codename}";
    "\${distro_id}:\${distro_codename}-security";
    "\${distro_id}ESMApps:\${distro_codename}-apps-security";
    "\${distro_id}ESM:\${distro_codename}-infra-security";
};
Unattended-Upgrade::AutoFixInterruptedDpkg "true";
Unattended-Upgrade::MinimalSteps "true";
Unattended-Upgrade::Remove-Unused-Kernel-Packages "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
Unattended-Upgrade::Automatic-Reboot "false";
EOF

cat > /etc/apt/apt.conf.d/20auto-upgrades <<EOF
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
EOF

systemctl enable unattended-upgrades
log_success "Automatic security updates enabled"

# Step 10: Install essential security tools
log_info "Step 10: Installing security tools..."
apt install -y \
    htop \
    iotop \
    net-tools \
    curl \
    wget \
    git \
    vim \
    tmux \
    rsync \
    logwatch \
    rkhunter \
    lynis

log_success "Security tools installed"

# Step 11: Disable unnecessary services
log_info "Step 11: Disabling unnecessary services..."
systemctl disable --now cups 2>/dev/null || true
systemctl disable --now avahi-daemon 2>/dev/null || true
systemctl disable --now bluetooth 2>/dev/null || true
log_success "Unnecessary services disabled"

# Step 12: Set proper permissions on critical files
log_info "Step 12: Setting secure file permissions..."
chmod 600 /etc/ssh/sshd_config
chmod 700 /root
chmod 644 /etc/passwd
chmod 644 /etc/group
chmod 600 /etc/shadow
chmod 600 /etc/gshadow
log_success "File permissions secured"

# Step 13: Create ShareHODL deployment directory
log_info "Step 13: Creating deployment directories..."
mkdir -p /opt/sharehodl
mkdir -p /var/log/sharehodl
chown -R $NEW_USER:$NEW_USER /opt/sharehodl
chown -R $NEW_USER:$NEW_USER /var/log/sharehodl
log_success "Deployment directories created"

# Summary
echo ""
echo "=============================================="
echo "  SERVER HARDENING COMPLETE"
echo "=============================================="
echo ""
log_success "Server hardening completed successfully!"
echo ""
echo "IMPORTANT NEXT STEPS:"
echo "1. Note down the user password from above"
echo "2. Delete /root/.sharehodl_user_password after noting"
echo "3. Add your SSH public key:"
echo "   ssh-copy-id -p $SSH_PORT $NEW_USER@$SERVER_IP"
echo ""
echo "4. Test SSH connection in a NEW terminal:"
echo "   ssh -p $SSH_PORT $NEW_USER@$SERVER_IP"
echo ""
echo "5. Once SSH key works, restart SSH:"
echo "   sudo systemctl restart sshd"
echo ""
echo "6. Run the deployment script:"
echo "   bash 02-deploy-sharehodl.sh"
echo ""
log_warning "DO NOT close this terminal until you verify SSH key access!"
echo ""

# Create a revert script in case of issues
cat > /root/revert-ssh.sh <<EOF
#!/bin/bash
# Emergency SSH revert script
cp /etc/ssh/sshd_config.backup /etc/ssh/sshd_config
systemctl restart sshd
echo "SSH config reverted to backup"
EOF
chmod +x /root/revert-ssh.sh
log_info "Emergency revert script created at /root/revert-ssh.sh"
