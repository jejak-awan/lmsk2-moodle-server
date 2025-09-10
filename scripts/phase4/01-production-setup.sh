#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: Phase 4 - Production Setup Script
# =============================================================================
# Description: Production environment setup for LMSK2-Moodle-Server
# Version: 1.0
# Author: jejakawan007
# Date: September 9, 2025
# =============================================================================

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Script configuration
SCRIPT_NAME="LMSK2-Moodle-Server Production Setup"
SCRIPT_VERSION="1.0"
LOG_FILE="/var/log/lmsk2-production-setup.log"
CONFIG_DIR="/opt/lmsk2-moodle-server/scripts/config"
PRODUCTION_DIR="/opt/lmsk2-moodle-server/scripts/phase4"

# Load configuration
if [ -f "$CONFIG_DIR/production.conf" ]; then
    source "$CONFIG_DIR/production.conf"
else
    echo -e "${YELLOW}Warning: Production configuration file not found. Using defaults.${NC}"
fi

# Default configuration
PRODUCTION_MODE=${PRODUCTION_MODE:-"true"}
SECURITY_LEVEL=${SECURITY_LEVEL:-"high"}
PERFORMANCE_MODE=${PERFORMANCE_MODE:-"high"}
MONITORING_ENABLE=${MONITORING_ENABLE:-"true"}
BACKUP_ENABLE=${BACKUP_ENABLE:-"true"}
SSL_ENABLE=${SSL_ENABLE:-"true"}
FIREWALL_ENABLE=${FIREWALL_ENABLE:-"true"}

# =============================================================================
# Utility Functions
# =============================================================================

# Logging function
log() {
    local level=$1
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case $level in
        "INFO")
            echo -e "${GREEN}[INFO]${NC} $message" | tee -a "$LOG_FILE"
            ;;
        "WARN")
            echo -e "${YELLOW}[WARN]${NC} $message" | tee -a "$LOG_FILE"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message" | tee -a "$LOG_FILE"
            ;;
        "DEBUG")
            if [ "$LOG_LEVEL" = "debug" ]; then
                echo -e "${BLUE}[DEBUG]${NC} $message" | tee -a "$LOG_FILE"
            fi
            ;;
    esac
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# Error handling
handle_error() {
    local exit_code=$1
    local error_message=$2
    
    if [ $exit_code -ne 0 ]; then
        log "ERROR" "$error_message"
        log "ERROR" "Production setup failed. Exit code: $exit_code"
        exit $exit_code
    fi
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log "ERROR" "This script must be run as root"
        exit 1
    fi
}

# =============================================================================
# Production Environment Setup
# =============================================================================

# Create production configuration
create_production_config() {
    log "INFO" "Creating production configuration..."
    
    cat > "$CONFIG_DIR/production.conf" << EOF
# LMSK2 Production Configuration

# Production settings
PRODUCTION_MODE=true
SECURITY_LEVEL=high
PERFORMANCE_MODE=high
MONITORING_ENABLE=true
BACKUP_ENABLE=true
SSL_ENABLE=true
FIREWALL_ENABLE=true

# Domain settings
PRODUCTION_DOMAIN=yourdomain.com
PRODUCTION_EMAIL=admin@yourdomain.com

# Security settings
SECURITY_HEADERS=true
HSTS_ENABLE=true
CSP_ENABLE=true
FAIL2BAN_ENABLE=true
INTRUSION_DETECTION=true

# Performance settings
OPCACHE_ENABLE=true
REDIS_CACHE=true
NGINX_CACHE=true
COMPRESSION_ENABLE=true
BROTLI_ENABLE=true

# Monitoring settings
LOG_LEVEL=info
ALERT_EMAIL=admin@yourdomain.com
MONITORING_INTERVAL=60
RETENTION_DAYS=30

# Backup settings
BACKUP_SCHEDULE="0 2 * * *"
BACKUP_RETENTION=30
BACKUP_ENCRYPTION=false

# SSL settings
SSL_PROVIDER=letsencrypt
SSL_RENEWAL=true
SSL_STRONG_CIPHERS=true

# Firewall settings
FIREWALL_PORTS="22,80,443"
FIREWALL_RATE_LIMIT=true
FIREWALL_GEO_BLOCKING=false

# System optimization
KERNEL_OPTIMIZATION=true
SYSTEM_LIMITS=true
SWAP_OPTIMIZATION=true
DISK_OPTIMIZATION=true
EOF

    log "INFO" "Production configuration created"
}

# =============================================================================
# Security Hardening
# =============================================================================

# Setup advanced security hardening
setup_advanced_security() {
    log "INFO" "Setting up advanced security hardening..."
    
    # Disable unnecessary services
    local services_to_disable=("bluetooth" "cups" "avahi-daemon" "cups-browsed")
    for service in "${services_to_disable[@]}"; do
        if systemctl is-enabled --quiet "$service" 2>/dev/null; then
            systemctl disable "$service" 2>/dev/null
            systemctl stop "$service" 2>/dev/null
            log "INFO" "Disabled unnecessary service: $service"
        fi
    done
    
    # Configure kernel security parameters
    cat > /etc/sysctl.d/99-lmsk2-security.conf << EOF
# LMSK2 Security Kernel Parameters

# Network security
net.ipv4.conf.all.send_redirects = 0
net.ipv4.conf.default.send_redirects = 0
net.ipv4.conf.all.accept_redirects = 0
net.ipv4.conf.default.accept_redirects = 0
net.ipv4.conf.all.accept_source_route = 0
net.ipv4.conf.default.accept_source_route = 0
net.ipv4.conf.all.log_martians = 1
net.ipv4.conf.default.log_martians = 1
net.ipv4.icmp_echo_ignore_broadcasts = 1
net.ipv4.icmp_ignore_bogus_error_responses = 1
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_rfc1337 = 1

# Memory protection
kernel.dmesg_restrict = 1
kernel.kptr_restrict = 2
kernel.yama.ptrace_scope = 1

# File system security
fs.protected_hardlinks = 1
fs.protected_symlinks = 1
fs.suid_dumpable = 0

# Process security
kernel.randomize_va_space = 2
EOF

    sysctl -p /etc/sysctl.d/99-lmsk2-security.conf
    
    # Configure system limits
    cat > /etc/security/limits.d/99-lmsk2.conf << EOF
# LMSK2 System Limits

# Core dumps
* soft core 0
* hard core 0

# File descriptors
* soft nofile 65536
* hard nofile 65536

# Processes
* soft nproc 32768
* hard nproc 32768

# Memory
* soft memlock unlimited
* hard memlock unlimited
EOF

    # Configure PAM security
    cat > /etc/pam.d/common-password << EOF
# LMSK2 PAM Password Configuration

password        [success=1 default=ignore]      pam_unix.so obscure sha512 minlen=12
password        requisite                       pam_deny.so
password        required                        pam_permit.so
password        optional        pam_gnome_keyring.so
EOF

    # Configure SSH security
    cat > /etc/ssh/sshd_config.d/99-lmsk2.conf << EOF
# LMSK2 SSH Security Configuration

# Authentication
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
AuthorizedKeysFile .ssh/authorized_keys

# Security
Protocol 2
MaxAuthTries 3
MaxSessions 2
ClientAliveInterval 300
ClientAliveCountMax 2
LoginGraceTime 60

# Logging
SyslogFacility AUTH
LogLevel INFO

# Network
AllowUsers www-data moodle
DenyUsers root
EOF

    systemctl reload sshd
    
    log "INFO" "Advanced security hardening completed"
}

# Setup intrusion detection
setup_intrusion_detection() {
    log "INFO" "Setting up intrusion detection..."
    
    # Install and configure AIDE (Advanced Intrusion Detection Environment)
    apt-get update -qq
    apt-get install -y aide aide-common || handle_error $? "Failed to install AIDE"
    
    # Initialize AIDE database
    aideinit --yes || handle_error $? "Failed to initialize AIDE database"
    
    # Move AIDE database to proper location
    mv /var/lib/aide/aide.db.new /var/lib/aide/aide.db
    
    # Create AIDE configuration
    cat > /etc/aide/aide.conf.d/99-lmsk2.conf << EOF
# LMSK2 AIDE Configuration

# Web directories
/var/www/moodle R
/var/www/html R

# Configuration files
/etc/nginx R
/etc/php R
/etc/mysql R
/etc/redis R

# System binaries
/bin R
/sbin R
/usr/bin R
/usr/sbin R

# System configuration
/etc/passwd R
/etc/shadow R
/etc/group R
/etc/sudoers R
EOF

    # Setup AIDE cron job
    (crontab -l 2>/dev/null; echo "0 3 * * * /usr/bin/aide --check | mail -s 'AIDE Report' admin@localhost") | crontab -
    
    # Install and configure rkhunter
    apt-get install -y rkhunter || handle_error $? "Failed to install rkhunter"
    
    # Configure rkhunter
    sed -i 's/UPDATE_MIRRORS=0/UPDATE_MIRRORS=1/' /etc/rkhunter.conf
    sed -i 's/MIRRORS_MODE=1/MIRRORS_MODE=0/' /etc/rkhunter.conf
    sed -i 's/WEB_CMD=""/WEB_CMD=""/' /etc/rkhunter.conf
    
    # Update rkhunter database
    rkhunter --update
    
    # Setup rkhunter cron job
    (crontab -l 2>/dev/null; echo "0 4 * * * /usr/bin/rkhunter --cronjob --update --quiet | mail -s 'RKHunter Report' admin@localhost") | crontab -
    
    log "INFO" "Intrusion detection setup completed"
}

# Setup advanced firewall
setup_advanced_firewall() {
    log "INFO" "Setting up advanced firewall..."
    
    # Install and configure UFW
    apt-get install -y ufw || handle_error $? "Failed to install UFW"
    
    # Reset UFW to defaults
    ufw --force reset
    
    # Set default policies
    ufw default deny incoming
    ufw default allow outgoing
    
    # Allow essential services
    ufw allow 22/tcp comment 'SSH'
    ufw allow 80/tcp comment 'HTTP'
    ufw allow 443/tcp comment 'HTTPS'
    
    # Allow monitoring ports (if needed)
    ufw allow 8080/tcp comment 'Monitoring Dashboard'
    
    # Enable rate limiting
    ufw limit 22/tcp comment 'SSH Rate Limit'
    
    # Enable UFW
    ufw --force enable
    
    # Install and configure fail2ban
    apt-get install -y fail2ban || handle_error $? "Failed to install fail2ban"
    
    # Create fail2ban configuration
    cat > /etc/fail2ban/jail.local << EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3
backend = systemd

[sshd]
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 3

[nginx-http-auth]
enabled = true
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 3

[nginx-limit-req]
enabled = true
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 3

[php-url-fopen]
enabled = true
port = http,https
logpath = /var/log/nginx/access.log
maxretry = 3
EOF

    systemctl enable fail2ban
    systemctl start fail2ban
    
    log "INFO" "Advanced firewall setup completed"
}

# =============================================================================
# Performance Optimization
# =============================================================================

# Setup advanced performance optimization
setup_advanced_performance() {
    log "INFO" "Setting up advanced performance optimization..."
    
    # Configure kernel performance parameters
    cat > /etc/sysctl.d/99-lmsk2-performance.conf << EOF
# LMSK2 Performance Kernel Parameters

# Network performance
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 65536 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
net.core.netdev_max_backlog = 5000
net.ipv4.tcp_congestion_control = bbr

# File system performance
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
vm.vfs_cache_pressure = 50

# Process performance
kernel.sched_rt_runtime_us = -1
kernel.sched_rt_period_us = 1000000
EOF

    sysctl -p /etc/sysctl.d/99-lmsk2-performance.conf
    
    # Configure systemd limits
    cat > /etc/systemd/system.conf.d/99-lmsk2.conf << EOF
[Manager]
DefaultLimitNOFILE=65536
DefaultLimitNPROC=32768
EOF

    systemctl daemon-reload
    
    # Optimize disk I/O
    echo noop > /sys/block/sda/queue/scheduler 2>/dev/null || true
    echo 1024 > /sys/block/sda/queue/nr_requests 2>/dev/null || true
    
    # Configure swap optimization
    if [ -f /swapfile ]; then
        swapon --show | grep -q swapfile || swapon /swapfile
        echo '/swapfile none swap sw 0 0' >> /etc/fstab
    fi
    
    log "INFO" "Advanced performance optimization completed"
}

# Setup advanced caching
setup_advanced_caching() {
    log "INFO" "Setting up advanced caching..."
    
    # Configure OPcache for production
    cat > /etc/php/8.1/mods-available/opcache-production.ini << EOF
; LMSK2 OPcache Production Configuration

[opcache]
opcache.enable=1
opcache.enable_cli=1
opcache.memory_consumption=256
opcache.interned_strings_buffer=16
opcache.max_accelerated_files=20000
opcache.revalidate_freq=0
opcache.validate_timestamps=0
opcache.save_comments=0
opcache.fast_shutdown=1
opcache.enable_file_override=1
opcache.optimization_level=0x7FFFBFFF
opcache.max_wasted_percentage=10
opcache.use_cwd=1
opcache.validate_permission=1
opcache.validate_root=1
opcache.file_update_protection=2
opcache.revalidate_path=0
opcache.save_comments=1
opcache.load_comments=1
opcache.dups_fix=0
opcache.blacklist_filename=/etc/php/8.1/mods-available/opcache-blacklist.txt
EOF

    # Create OPcache blacklist
    cat > /etc/php/8.1/mods-available/opcache-blacklist.txt << EOF
# LMSK2 OPcache Blacklist

# Development files
*.dev.php
*.test.php
*.debug.php

# Cache files
*/cache/*
*/temp/*
*/tmp/*

# Log files
*.log
EOF

    # Enable OPcache production configuration
    phpenmod opcache-production
    
    # Configure Redis for production
    cat > /etc/redis/redis-production.conf << EOF
# LMSK2 Redis Production Configuration

# Memory management
maxmemory 512mb
maxmemory-policy allkeys-lru

# Persistence
save 900 1
save 300 10
save 60 10000

# Performance
tcp-keepalive 60
timeout 300
tcp-backlog 511

# Security
requirepass your_redis_password_here
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command KEYS ""
rename-command CONFIG ""

# Logging
loglevel notice
logfile /var/log/redis/redis-server.log

# Network
bind 127.0.0.1
port 6379
EOF

    # Configure Nginx caching
    cat > /etc/nginx/conf.d/lmsk2-cache.conf << EOF
# LMSK2 Nginx Cache Configuration

# Proxy cache
proxy_cache_path /var/cache/nginx/proxy levels=1:2 keys_zone=proxy_cache:10m max_size=1g inactive=60m use_temp_path=off;

# FastCGI cache
fastcgi_cache_path /var/cache/nginx/fastcgi levels=1:2 keys_zone=fastcgi_cache:10m max_size=1g inactive=60m use_temp_path=off;

# Cache zones
proxy_cache_key \$scheme\$request_method\$host\$request_uri;
fastcgi_cache_key \$scheme\$request_method\$host\$request_uri;

# Cache bypass
proxy_cache_bypass \$http_pragma \$http_authorization;
fastcgi_cache_bypass \$http_pragma \$http_authorization;

# Cache methods
proxy_cache_methods GET HEAD;
fastcgi_cache_methods GET HEAD;
EOF

    # Create cache directories
    mkdir -p /var/cache/nginx/{proxy,fastcgi}
    chown -R www-data:www-data /var/cache/nginx
    chmod -R 755 /var/cache/nginx
    
    log "INFO" "Advanced caching setup completed"
}

# =============================================================================
# Monitoring and Alerting
# =============================================================================

# Setup production monitoring
setup_production_monitoring() {
    log "INFO" "Setting up production monitoring..."
    
    # Create production monitoring configuration
    cat > "$CONFIG_DIR/monitoring-production.conf" << EOF
# LMSK2 Production Monitoring Configuration

# Monitoring settings
MONITORING_ENABLE=true
ALERT_EMAIL=admin@yourdomain.com
LOG_LEVEL=info
MONITORING_INTERVAL=30

# Alert thresholds
CPU_THRESHOLD=70
MEMORY_THRESHOLD=80
DISK_THRESHOLD=85
LOAD_THRESHOLD=2.0
RESPONSE_TIME_THRESHOLD=1000

# Service monitoring
SERVICE_MONITORING=true
DATABASE_MONITORING=true
CACHE_MONITORING=true
WEB_MONITORING=true

# Log monitoring
LOG_MONITORING=true
ERROR_MONITORING=true
SECURITY_MONITORING=true

# Performance monitoring
PERFORMANCE_MONITORING=true
BENCHMARK_MONITORING=true
EOF

    # Setup enhanced cron jobs for monitoring
    cat > /etc/cron.d/lmsk2-production-monitoring << EOF
# LMSK2 Production Monitoring Cron Jobs

# System health check every 5 minutes
*/5 * * * * root /opt/lmsk2-moodle-server/scripts/monitoring/system-health-check.sh

# Performance monitoring every 10 minutes
*/10 * * * * root /opt/lmsk2-moodle-server/scripts/monitoring/performance-monitor.sh

# Security monitoring every 15 minutes
*/15 * * * * root /opt/lmsk2-moodle-server/scripts/monitoring/security-monitor.sh

# Cache monitoring every 20 minutes
*/20 * * * * root /opt/lmsk2-moodle-server/scripts/monitoring/cache-monitor.sh

# Log rotation check daily
0 1 * * * root /usr/sbin/logrotate /etc/logrotate.conf

# System cleanup weekly
0 2 * * 0 root /opt/lmsk2-moodle-server/scripts/utilities/system-cleanup.sh
EOF

    # Setup log aggregation
    cat > /etc/rsyslog.d/99-lmsk2.conf << EOF
# LMSK2 Log Aggregation

# Local logging
local0.*    /var/log/lmsk2/lmsk2.log
local1.*    /var/log/lmsk2/security.log
local2.*    /var/log/lmsk2/performance.log
local3.*    /var/log/lmsk2/application.log

# Remote logging (if configured)
# *.* @@remote-log-server:514
EOF

    # Create log directories
    mkdir -p /var/log/lmsk2
    chown -R syslog:adm /var/log/lmsk2
    chmod -R 755 /var/log/lmsk2
    
    systemctl restart rsyslog
    
    log "INFO" "Production monitoring setup completed"
}

# =============================================================================
# Backup and Recovery
# =============================================================================

# Setup production backup
setup_production_backup() {
    log "INFO" "Setting up production backup..."
    
    # Create production backup configuration
    cat > "$CONFIG_DIR/backup-production.conf" << EOF
# LMSK2 Production Backup Configuration

# Backup settings
BACKUP_ENABLE=true
BACKUP_ENCRYPTION=true
BACKUP_COMPRESSION=gzip
BACKUP_EMAIL=admin@yourdomain.com

# Backup schedules
FULL_BACKUP_SCHEDULE="0 1 * * 0"  # Weekly on Sunday at 1 AM
INCREMENTAL_BACKUP_SCHEDULE="0 1 * * 1-6"  # Daily at 1 AM
CONFIG_BACKUP_SCHEDULE="0 2 * * *"  # Daily at 2 AM

# Retention policies
DAILY_RETENTION=7
WEEKLY_RETENTION=4
MONTHLY_RETENTION=12
YEARLY_RETENTION=2

# Backup locations
BACKUP_BASE_DIR=/var/backups/lmsk2
BACKUP_REMOTE_DIR=/mnt/backup/lmsk2

# Database settings
MOODLE_DB_NAME=moodle
MOODLE_DB_USER=moodle
MOODLE_DB_PASSWORD=your_password_here
MOODLE_DB_HOST=localhost

# File paths
MOODLE_DATA_DIR=/var/www/moodle
MOODLE_CONFIG_DIR=/var/www/moodle/config.php

# Notification settings
SEND_BACKUP_NOTIFICATIONS=true
BACKUP_SUCCESS_EMAIL=true
BACKUP_FAILURE_EMAIL=true
EOF

    # Setup backup encryption
    if [ ! -f /etc/lmsk2-backup/backup.key ]; then
        mkdir -p /etc/lmsk2-backup
        openssl rand -base64 32 > /etc/lmsk2-backup/backup.key
        chmod 600 /etc/lmsk2-backup/backup.key
        chown root:root /etc/lmsk2-backup/backup.key
    fi
    
    # Setup backup verification
    cat > /opt/lmsk2-moodle-server/scripts/utilities/backup-verification.sh << 'EOF'
#!/bin/bash

# Backup verification script
CONFIG_FILE="/opt/lmsk2-moodle-server/scripts/config/backup-production.conf"
LOG_FILE="/var/log/lmsk2-backup/backup-verification.log"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Function to log with timestamp
log_verification() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Verify database backup
verify_database_backup() {
    local backup_file="$1"
    
    log_verification "Verifying database backup: $backup_file"
    
    if [ ! -f "$backup_file" ]; then
        log_verification "ERROR: Database backup file not found: $backup_file"
        return 1
    fi
    
    # Test backup restoration (dry run)
    if [[ "$backup_file" == *.gz ]]; then
        gunzip -t "$backup_file" 2>> "$LOG_FILE"
    elif [[ "$backup_file" == *.enc ]]; then
        # For encrypted files, just check if they exist and have content
        if [ -s "$backup_file" ]; then
            log_verification "Encrypted database backup verified"
        else
            log_verification "ERROR: Encrypted database backup is empty"
            return 1
        fi
    else
        # For plain SQL files, check for SQL syntax
        head -n 10 "$backup_file" | grep -q "CREATE\|INSERT\|UPDATE" 2>> "$LOG_FILE"
    fi
    
    if [ $? -eq 0 ]; then
        log_verification "Database backup verification successful"
        return 0
    else
        log_verification "ERROR: Database backup verification failed"
        return 1
    fi
}

# Verify file backup
verify_file_backup() {
    local backup_file="$1"
    
    log_verification "Verifying file backup: $backup_file"
    
    if [ ! -f "$backup_file" ]; then
        log_verification "ERROR: File backup file not found: $backup_file"
        return 1
    fi
    
    # Test backup extraction (dry run)
    if [[ "$backup_file" == *.gz ]]; then
        gunzip -t "$backup_file" 2>> "$LOG_FILE"
    elif [[ "$backup_file" == *.enc ]]; then
        # For encrypted files, just check if they exist and have content
        if [ -s "$backup_file" ]; then
            log_verification "Encrypted file backup verified"
        else
            log_verification "ERROR: Encrypted file backup is empty"
            return 1
        fi
    else
        # For plain tar files, test extraction
        tar -tf "$backup_file" > /dev/null 2>> "$LOG_FILE"
    fi
    
    if [ $? -eq 0 ]; then
        log_verification "File backup verification successful"
        return 0
    else
        log_verification "ERROR: File backup verification failed"
        return 1
    fi
}

# Main verification function
main() {
    log_verification "Starting backup verification process"
    
    # Find latest backups
    local latest_db_backup=$(find "$BACKUP_BASE_DIR/database" -name "*.sql*" -type f -printf '%T@ %p\n' 2>/dev/null | sort -n | tail -1 | cut -d' ' -f2-)
    local latest_file_backup=$(find "$BACKUP_BASE_DIR/files" -name "*.tar*" -type f -printf '%T@ %p\n' 2>/dev/null | sort -n | tail -1 | cut -d' ' -f2-)
    
    local verification_failed=0
    
    # Verify database backup
    if [ -n "$latest_db_backup" ]; then
        if ! verify_database_backup "$latest_db_backup"; then
            verification_failed=1
        fi
    else
        log_verification "WARNING: No database backup found for verification"
    fi
    
    # Verify file backup
    if [ -n "$latest_file_backup" ]; then
        if ! verify_file_backup "$latest_file_backup"; then
            verification_failed=1
        fi
    else
        log_verification "WARNING: No file backup found for verification"
    fi
    
    if [ "$verification_failed" -eq 0 ]; then
        log_verification "Backup verification process completed successfully"
        exit 0
    else
        log_verification "Backup verification process failed"
        exit 1
    fi
}

# Run main function
main "$@"
EOF

    chmod +x /opt/lmsk2-moodle-server/scripts/utilities/backup-verification.sh
    
    # Setup backup verification cron job
    (crontab -l 2>/dev/null; echo "0 3 * * * /opt/lmsk2-moodle-server/scripts/utilities/backup-verification.sh") | crontab -
    
    log "INFO" "Production backup setup completed"
}

# =============================================================================
# Main Installation Function
# =============================================================================

# Main installation function
main() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  $SCRIPT_NAME v$SCRIPT_VERSION${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo
    
    log "INFO" "Starting production setup..."
    
    # Check prerequisites
    check_root
    
    # Setup production environment
    create_production_config
    setup_advanced_security
    setup_intrusion_detection
    setup_advanced_firewall
    setup_advanced_performance
    setup_advanced_caching
    setup_production_monitoring
    setup_production_backup
    
    # Final verification
    log "INFO" "Verifying production setup..."
    
    # Check security configurations
    if [ -f "/etc/sysctl.d/99-lmsk2-security.conf" ]; then
        log "INFO" "✓ Security kernel parameters configured"
    else
        log "ERROR" "✗ Security kernel parameters not configured"
    fi
    
    # Check firewall status
    if ufw status | grep -q "Status: active"; then
        log "INFO" "✓ Firewall is active"
    else
        log "WARN" "✗ Firewall is not active"
    fi
    
    # Check fail2ban status
    if systemctl is-active --quiet fail2ban; then
        log "INFO" "✓ Fail2ban is running"
    else
        log "WARN" "✗ Fail2ban is not running"
    fi
    
    # Check monitoring cron jobs
    if crontab -l | grep -q "lmsk2-production-monitoring"; then
        log "INFO" "✓ Production monitoring cron jobs configured"
    else
        log "WARN" "✗ Production monitoring cron jobs not found"
    fi
    
    # Display summary
    echo
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Production Setup Completed!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo
    echo -e "${WHITE}Production Components Installed:${NC}"
    echo -e "  • Advanced security hardening"
    echo -e "  • Intrusion detection (AIDE, rkhunter)"
    echo -e "  • Advanced firewall (UFW, fail2ban)"
    echo -e "  • Performance optimization"
    echo -e "  • Advanced caching (OPcache, Redis, Nginx)"
    echo -e "  • Production monitoring"
    echo -e "  • Enhanced backup system"
    echo -e "  • Log aggregation"
    echo
    echo -e "${WHITE}Configuration Files:${NC}"
    echo -e "  • $CONFIG_DIR/production.conf"
    echo -e "  • $CONFIG_DIR/monitoring-production.conf"
    echo -e "  • $CONFIG_DIR/backup-production.conf"
    echo -e "  • /etc/sysctl.d/99-lmsk2-security.conf"
    echo -e "  • /etc/sysctl.d/99-lmsk2-performance.conf"
    echo
    echo -e "${WHITE}Security Features:${NC}"
    echo -e "  • Kernel security parameters"
    echo -e "  • System limits and PAM security"
    echo -e "  • SSH security hardening"
    echo -e "  • Intrusion detection"
    echo -e "  • Advanced firewall rules"
    echo
    echo -e "${WHITE}Performance Features:${NC}"
    echo -e "  • Kernel performance parameters"
    echo -e "  • Systemd limits"
    echo -e "  • Disk I/O optimization"
    echo -e "  • Advanced caching"
    echo -e "  • Swap optimization"
    echo
    echo -e "${WHITE}Next Steps:${NC}"
    echo -e "  1. Configure domain and email in production.conf"
    echo -e "  2. Set up SSL certificates"
    echo -e "  3. Configure load balancing (if needed)"
    echo -e "  4. Test all monitoring systems"
    echo -e "  5. Verify backup and recovery procedures"
    echo -e "  6. Run security audit"
    echo -e "  7. Performance testing"
    echo
    
    log "INFO" "Production setup completed successfully"
}

# =============================================================================
# Script Execution
# =============================================================================

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo
            echo "Options:"
            echo "  --help, -h          Show this help message"
            echo "  --version, -v       Show version information"
            echo "  --config FILE       Use custom configuration file"
            echo "  --security-level    Set security level (low|medium|high)"
            echo "  --performance-mode  Set performance mode (low|medium|high)"
            echo "  --monitoring        Enable/disable monitoring"
            echo "  --backup            Enable/disable backup"
            echo
            exit 0
            ;;
        --version|-v)
            echo "$SCRIPT_NAME v$SCRIPT_VERSION"
            exit 0
            ;;
        --config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        --security-level)
            SECURITY_LEVEL="$2"
            shift 2
            ;;
        --performance-mode)
            PERFORMANCE_MODE="$2"
            shift 2
            ;;
        --monitoring)
            MONITORING_ENABLE="$2"
            shift 2
            ;;
        --backup)
            BACKUP_ENABLE="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run main function
main "$@"

