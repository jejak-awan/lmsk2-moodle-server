#!/bin/bash

# =============================================================================
# LMSK2 Moodle Server - Caching Setup Script
# =============================================================================
# Description: Setup comprehensive caching system for Moodle
# Author: jejakawan007
# Version: 1.0
# Date: September 2025
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Load configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/config/installer.conf"

if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
else
    echo "❌ Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Script configuration
SCRIPT_NAME="$(basename "$0")"
LOG_FILE="/var/log/lmsk2/${SCRIPT_NAME%.*}.log"
BACKUP_DIR="/backup/lmsk2/caching-setup"
MOODLE_DIR="/var/www/moodle"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# =============================================================================
# Logging Functions
# =============================================================================

log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case "$level" in
        "INFO")  echo -e "${GREEN}[INFO]${NC} $message" ;;
        "WARN")  echo -e "${YELLOW}[WARN]${NC} $message" ;;
        "ERROR") echo -e "${RED}[ERROR]${NC} $message" ;;
        "DEBUG") echo -e "${BLUE}[DEBUG]${NC} $message" ;;
    esac
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# =============================================================================
# Utility Functions
# =============================================================================

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log "ERROR" "This script must be run as root"
        exit 1
    fi
}

create_directories() {
    log "INFO" "Creating necessary directories..."
    
    mkdir -p "$BACKUP_DIR"
    mkdir -p "/var/cache/nginx/fastcgi"
    mkdir -p "/usr/local/bin"
    
    log "INFO" "Directories created successfully"
}

backup_config() {
    local config_file="$1"
    local backup_file="$BACKUP_DIR/$(basename "$config_file").backup.$(date +%Y%m%d_%H%M%S)"
    
    if [[ -f "$config_file" ]]; then
        cp "$config_file" "$backup_file"
        log "INFO" "Backed up $config_file to $backup_file"
    fi
}

# =============================================================================
# Redis Caching Configuration
# =============================================================================

setup_redis_caching() {
    log "INFO" "Setting up Redis caching..."
    
    # Backup current Redis configuration
    backup_config "/etc/redis/redis.conf"
    
    # Create Redis caching configuration
    cat > /etc/redis/redis.conf << 'EOF'
# Redis Caching Configuration for Moodle

# Network
bind 127.0.0.1
port 6379
timeout 0
tcp-keepalive 300

# Memory management
maxmemory 512mb
maxmemory-policy allkeys-lru
maxmemory-samples 5

# Persistence
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir /var/lib/redis

# Logging
loglevel notice
logfile /var/log/redis/redis-server.log
syslog-enabled no

# Performance
tcp-backlog 511
databases 16
always-show-logo yes

# Session storage
maxmemory-samples 5
EOF

    # Restart Redis
    systemctl restart redis-server
    
    # Test Redis connection
    if redis-cli ping >/dev/null 2>&1; then
        log "INFO" "Redis caching setup completed successfully"
    else
        log "ERROR" "Redis connection test failed"
        return 1
    fi
}

# =============================================================================
# OPcache Configuration
# =============================================================================

setup_opcache() {
    log "INFO" "Setting up OPcache..."
    
    # Backup current PHP configuration
    backup_config "/etc/php/8.1/fpm/php.ini"
    
    # Create OPcache configuration
    cat > /etc/php/8.1/fpm/conf.d/99-opcache.ini << 'EOF'
; OPcache Configuration for Moodle

; Enable OPcache
opcache.enable = 1
opcache.enable_cli = 1

; Memory settings
opcache.memory_consumption = 256
opcache.interned_strings_buffer = 8
opcache.max_accelerated_files = 10000

; Performance settings
opcache.revalidate_freq = 2
opcache.fast_shutdown = 1
opcache.validate_timestamps = 0
opcache.save_comments = 1
opcache.enable_file_override = 1

; Optimization level
opcache.optimization_level = 0x7FFFBFFF

; Blacklist
opcache.blacklist_filename = /etc/php/8.1/fpm/opcache-blacklist.txt

; Error handling
opcache.log_verbosity_level = 2
opcache.error_log = /var/log/php_opcache.log
EOF

    # Create OPcache blacklist
    cat > /etc/php/8.1/fpm/opcache-blacklist.txt << 'EOF'
/var/www/moodle/config.php
/var/www/moodle/version.php
/var/www/moodle/lib/setup.php
/var/www/moodle/admin/cli/
/var/www/moodle/install/
EOF

    # Restart PHP-FPM
    systemctl restart php8.1-fpm
    
    log "INFO" "OPcache setup completed successfully"
}

# =============================================================================
# Nginx FastCGI Caching
# =============================================================================

setup_nginx_caching() {
    log "INFO" "Setting up Nginx FastCGI caching..."
    
    # Backup current Nginx configuration
    backup_config "/etc/nginx/nginx.conf"
    backup_config "/etc/nginx/sites-available/moodle"
    
    # Create Nginx caching configuration
    cat > /etc/nginx/conf.d/caching.conf << 'EOF'
# Nginx FastCGI Caching Configuration for Moodle

# FastCGI cache configuration
fastcgi_cache_path /var/cache/nginx/fastcgi levels=1:2 keys_zone=moodle:100m inactive=60m max_size=1g;
fastcgi_cache_key "$scheme$request_method$host$request_uri";
fastcgi_cache_lock on;
fastcgi_cache_lock_timeout 5s;

# Cache settings
fastcgi_cache_use_stale error timeout invalid_header http_500;
fastcgi_ignore_headers Cache-Control Expires Set-Cookie;

# Cache bypass conditions
map $request_method $purge_method {
    PURGE   1;
    default 0;
}

# Cache status header
add_header X-Cache-Status $upstream_cache_status;
EOF

    # Set proper ownership for cache directory
    chown -R www-data:www-data /var/cache/nginx
    
    # Update Moodle site configuration with caching
    cat > /etc/nginx/sites-available/moodle << EOF
# HTTP to HTTPS redirect
server {
    listen 80;
    server_name ${MOODLE_DOMAIN:-lms.example.com};
    return 301 https://\$server_name\$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name ${MOODLE_DOMAIN:-lms.example.com};
    root /var/www/moodle;
    index index.php index.html;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/${MOODLE_DOMAIN:-lms.example.com}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${MOODLE_DOMAIN:-lms.example.com}/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Rate limiting
    limit_req_zone \$binary_remote_addr zone=login:10m rate=5r/m;
    limit_req_zone \$binary_remote_addr zone=api:10m rate=10r/s;

    # Skip cache for admin and login
    set \$skip_cache 0;
    if (\$request_uri ~* "/admin/") {
        set \$skip_cache 1;
    }
    if (\$request_uri ~* "/login/") {
        set \$skip_cache 1;
    }
    if (\$request_uri ~* "/user/") {
        set \$skip_cache 1;
    }
    if (\$request_uri ~* "/course/edit/") {
        set \$skip_cache 1;
    }
    if (\$request_method = POST) {
        set \$skip_cache 1;
    }
    if (\$http_cookie ~* "MoodleSession") {
        set \$skip_cache 1;
    }

    # Main location
    location / {
        try_files \$uri \$uri/ /index.php?\$query_string;
    }

    # Login rate limiting
    location ~ ^/login/ {
        limit_req zone=login burst=3 nodelay;
        try_files \$uri \$uri/ /index.php?\$query_string;
    }

    # API rate limiting
    location ~ ^/webservice/ {
        limit_req zone=api burst=20 nodelay;
        try_files \$uri \$uri/ /index.php?\$query_string;
    }

    # PHP processing with FastCGI cache
    location ~ \.php\$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php8.1-fpm.sock;
        fastcgi_param SCRIPT_FILENAME \$document_root\$fastcgi_script_name;
        include fastcgi_params;
        
        # FastCGI cache
        fastcgi_cache moodle;
        fastcgi_cache_valid 200 60m;
        fastcgi_cache_valid 404 1m;
        fastcgi_cache_bypass \$skip_cache;
        fastcgi_no_cache \$skip_cache;
        
        # FastCGI settings
        fastcgi_connect_timeout 60s;
        fastcgi_send_timeout 60s;
        fastcgi_read_timeout 60s;
        fastcgi_buffer_size 128k;
        fastcgi_buffers 4 256k;
        fastcgi_busy_buffers_size 256k;
        fastcgi_temp_file_write_size 256k;
        
        # Security headers for PHP
        fastcgi_param HTTPS on;
        fastcgi_param HTTP_SCHEME https;
    }

    # Cache purge endpoint
    location ~ /purge(/.*) {
        fastcgi_cache_purge moodle "\$scheme\$request_method\$host\$1";
    }

    # Deny access to sensitive files
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }

    location ~ /(config|cache|local|moodledata|backup|temp|lang|pix|theme|userpix|upgrade|admin|lib|install|test|vendor)/ {
        deny all;
        access_log off;
        log_not_found off;
    }

    # Block common attack patterns
    location ~* /(wp-admin|wp-login|xmlrpc|admin|administrator) {
        deny all;
        access_log off;
        log_not_found off;
    }

    # Static files caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)\$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        add_header Vary "Accept-Encoding";
        
        # Enable gzip for static files
        gzip_static on;
    }
}
EOF

    # Test Nginx configuration
    nginx -t
    
    # Reload Nginx
    systemctl reload nginx
    
    log "INFO" "Nginx FastCGI caching setup completed successfully"
}

# =============================================================================
# Moodle Caching Configuration
# =============================================================================

setup_moodle_caching() {
    log "INFO" "Setting up Moodle caching configuration..."
    
    cd "$MOODLE_DIR"
    
    # Configure Moodle caching settings
    sudo -u www-data php admin/cli/cfg.php --name=cachejs --set=1
    sudo -u www-data php admin/cli/cfg.php --name=cachetemplates --set=1
    sudo -u www-data php admin/cli/cfg.php --name=cachecss --set=1
    
    # Configure session handler
    sudo -u www-data php admin/cli/cfg.php --name=session_handler_class --set="\\core\\session\\redis"
    sudo -u www-data php admin/cli/cfg.php --name=session_redis_host --set="127.0.0.1"
    sudo -u www-data php admin/cli/cfg.php --name=session_redis_port --set=6379
    sudo -u www-data php admin/cli/cfg.php --name=session_redis_database --set=0
    sudo -u www-data php admin/cli/cfg.php --name=session_redis_prefix --set="moodle_session_"
    
    # Configure application caching
    sudo -u www-data php admin/cli/cfg.php --name=application_cache --set=1
    sudo -u www-data php admin/cli/cfg.php --name=localcachedir --set="/var/www/moodledata/cache"
    
    # Configure string caching
    sudo -u www-data php admin/cli/cfg.php --name=stringcaches --set=1
    
    # Configure language caching
    sudo -u www-data php admin/cli/cfg.php --name=langstringcache --set=1
    
    # Clear all caches
    sudo -u www-data php admin/cli/purge_caches.php
    
    log "INFO" "Moodle caching configuration completed successfully"
}

# =============================================================================
# Cache Monitoring Setup
# =============================================================================

setup_cache_monitoring() {
    log "INFO" "Setting up cache monitoring..."
    
    # Create cache monitoring script
    cat > /usr/local/bin/cache-monitor.sh << 'EOF'
#!/bin/bash

# Cache Monitoring Script for Moodle
LOG_FILE="/var/log/cache-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Cache monitoring..." >> $LOG_FILE

# Redis cache info
REDIS_INFO=$(redis-cli info memory | grep used_memory_human)
echo "Redis Memory: $REDIS_INFO" >> $LOG_FILE

REDIS_KEYS=$(redis-cli dbsize)
echo "Redis Keys: $REDIS_KEYS" >> $LOG_FILE

# OPcache info
OPCACHE_STATUS=$(php -r "echo json_encode(opcache_get_status());" 2>/dev/null || echo "OPcache not available")
if [[ "$OPCACHE_STATUS" != "OPcache not available" ]]; then
    OPCACHE_HITS=$(php -r "echo opcache_get_status()['opcache_statistics']['hits'];" 2>/dev/null || echo "N/A")
    OPCACHE_MISSES=$(php -r "echo opcache_get_status()['opcache_statistics']['misses'];" 2>/dev/null || echo "N/A")
    echo "OPcache Hits: $OPCACHE_HITS" >> $LOG_FILE
    echo "OPcache Misses: $OPCACHE_MISSES" >> $LOG_FILE
fi

# Nginx cache info
NGINX_CACHE=$(du -sh /var/cache/nginx/fastcgi 2>/dev/null || echo "Cache directory not found")
echo "Nginx Cache Size: $NGINX_CACHE" >> $LOG_FILE

NGINX_CACHE_FILES=$(find /var/cache/nginx/fastcgi -type f 2>/dev/null | wc -l)
echo "Nginx Cache Files: $NGINX_CACHE_FILES" >> $LOG_FILE

# PHP-FPM cache info
PHP_FPM_PROCESSES=$(ps aux | grep php-fpm | grep -v grep | wc -l)
echo "PHP-FPM Processes: $PHP_FPM_PROCESSES" >> $LOG_FILE

echo "---" >> $LOG_FILE
EOF

    # Make script executable
    chmod +x /usr/local/bin/cache-monitor.sh
    
    # Add to crontab
    (crontab -l 2>/dev/null; echo "*/5 * * * * /usr/local/bin/cache-monitor.sh") | crontab -
    
    log "INFO" "Cache monitoring setup completed successfully"
}

# =============================================================================
# Cache Management Functions
# =============================================================================

create_cache_management_scripts() {
    log "INFO" "Creating cache management scripts..."
    
    # Create cache clear script
    cat > /usr/local/bin/clear-cache.sh << 'EOF'
#!/bin/bash

# Cache Clear Script for Moodle
echo "=== Clearing All Caches ==="

# Clear Moodle caches
echo "Clearing Moodle caches..."
cd /var/www/moodle
sudo -u www-data php admin/cli/purge_caches.php

# Clear OPcache
echo "Clearing OPcache..."
php -r "if (function_exists('opcache_reset')) { opcache_reset(); echo 'OPcache cleared'; } else { echo 'OPcache not available'; }"

# Clear Redis cache
echo "Clearing Redis cache..."
redis-cli flushall

# Clear Nginx cache
echo "Clearing Nginx cache..."
rm -rf /var/cache/nginx/fastcgi/*

# Restart services
echo "Restarting services..."
systemctl restart php8.1-fpm
systemctl reload nginx

echo "All caches cleared successfully"
EOF

    # Create cache status script
    cat > /usr/local/bin/cache-status.sh << 'EOF'
#!/bin/bash

# Cache Status Script for Moodle
echo "=== Cache Status ==="

# Redis status
echo "Redis Status:"
redis-cli ping
redis-cli info memory | grep used_memory_human
redis-cli dbsize

echo ""

# OPcache status
echo "OPcache Status:"
php -r "if (function_exists('opcache_get_status')) { \$status = opcache_get_status(); echo 'Enabled: ' . (\$status['opcache_enabled'] ? 'Yes' : 'No') . PHP_EOL; echo 'Hits: ' . \$status['opcache_statistics']['hits'] . PHP_EOL; echo 'Misses: ' . \$status['opcache_statistics']['misses'] . PHP_EOL; } else { echo 'OPcache not available'; }"

echo ""

# Nginx cache status
echo "Nginx Cache Status:"
du -sh /var/cache/nginx/fastcgi 2>/dev/null || echo "Cache directory not found"
find /var/cache/nginx/fastcgi -type f 2>/dev/null | wc -l

echo ""

# Moodle cache status
echo "Moodle Cache Status:"
cd /var/www/moodle
sudo -u www-data php admin/cli/cfg.php --name=cachejs --get
sudo -u www-data php admin/cli/cfg.php --name=cachetemplates --get
sudo -u www-data php admin/cli/cfg.php --name=session_handler_class --get
EOF

    # Make scripts executable
    chmod +x /usr/local/bin/clear-cache.sh
    chmod +x /usr/local/bin/cache-status.sh
    
    log "INFO" "Cache management scripts created successfully"
}

# =============================================================================
# Verification Functions
# =============================================================================

verify_caching_setup() {
    log "INFO" "Verifying caching setup..."
    
    # Test Redis
    log "INFO" "Testing Redis..."
    if redis-cli ping >/dev/null 2>&1; then
        log "INFO" "✓ Redis is responding"
        redis-cli info memory | grep used_memory_human
    else
        log "ERROR" "✗ Redis is not responding"
    fi
    
    # Test OPcache
    log "INFO" "Testing OPcache..."
    if php -r "echo opcache_get_status()['opcache_enabled'] ? 'Enabled' : 'Disabled';" 2>/dev/null; then
        log "INFO" "✓ OPcache is working"
    else
        log "ERROR" "✗ OPcache is not working"
    fi
    
    # Test Nginx cache
    log "INFO" "Testing Nginx cache..."
    if [[ -d "/var/cache/nginx/fastcgi" ]]; then
        log "INFO" "✓ Nginx cache directory exists"
        ls -la /var/cache/nginx/fastcgi/ | head -5
    else
        log "ERROR" "✗ Nginx cache directory not found"
    fi
    
    # Test Moodle caching
    log "INFO" "Testing Moodle caching..."
    cd "$MOODLE_DIR"
    local cachejs=$(sudo -u www-data php admin/cli/cfg.php --name=cachejs --get 2>/dev/null || echo "0")
    local cachetemplates=$(sudo -u www-data php admin/cli/cfg.php --name=cachetemplates --get 2>/dev/null || echo "0")
    
    if [[ "$cachejs" == "1" && "$cachetemplates" == "1" ]]; then
        log "INFO" "✓ Moodle caching is configured"
    else
        log "WARN" "⚠ Moodle caching may not be fully configured"
    fi
    
    log "INFO" "Caching setup verification completed"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    log "INFO" "Starting caching setup process..."
    
    # Check prerequisites
    check_root
    create_directories
    
    # Execute caching setup steps
    setup_redis_caching
    setup_opcache
    setup_nginx_caching
    setup_moodle_caching
    setup_cache_monitoring
    create_cache_management_scripts
    
    # Verify setup
    verify_caching_setup
    
    log "INFO" "Caching setup process completed successfully!"
    log "INFO" "Log file: $LOG_FILE"
    log "INFO" "Backup directory: $BACKUP_DIR"
    log "INFO" "Cache management scripts: /usr/local/bin/clear-cache.sh, /usr/local/bin/cache-status.sh"
}

# =============================================================================
# Script Execution
# =============================================================================

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --verify       Only verify current caching setup"
        echo "  --dry-run      Show what would be done without executing"
        echo "  --redis        Setup Redis caching only"
        echo "  --opcache      Setup OPcache only"
        echo "  --nginx        Setup Nginx caching only"
        echo "  --moodle       Setup Moodle caching only"
        echo "  --monitoring   Setup cache monitoring only"
        exit 0
        ;;
    --verify)
        check_root
        verify_caching_setup
        exit 0
        ;;
    --dry-run)
        log "INFO" "DRY RUN MODE - No changes will be made"
        log "INFO" "Would execute: setup_redis_caching, setup_opcache, setup_nginx_caching, setup_moodle_caching, setup_cache_monitoring, create_cache_management_scripts"
        exit 0
        ;;
    --redis)
        check_root
        create_directories
        setup_redis_caching
        log "INFO" "Redis caching setup completed"
        ;;
    --opcache)
        check_root
        create_directories
        setup_opcache
        log "INFO" "OPcache setup completed"
        ;;
    --nginx)
        check_root
        create_directories
        setup_nginx_caching
        log "INFO" "Nginx caching setup completed"
        ;;
    --moodle)
        check_root
        create_directories
        setup_moodle_caching
        log "INFO" "Moodle caching setup completed"
        ;;
    --monitoring)
        check_root
        create_directories
        setup_cache_monitoring
        create_cache_management_scripts
        log "INFO" "Cache monitoring setup completed"
        ;;
    "")
        main
        ;;
    *)
        log "ERROR" "Unknown option: $1"
        log "INFO" "Use --help for usage information"
        exit 1
        ;;
esac
