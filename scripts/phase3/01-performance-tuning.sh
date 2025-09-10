#!/bin/bash

# =============================================================================
# LMSK2 Moodle Server - Performance Tuning Script
# =============================================================================
# Description: Comprehensive performance optimization for Moodle server
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
BACKUP_DIR="/backup/lmsk2/performance-tuning"

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
# Database Performance Optimization
# =============================================================================

optimize_database() {
    log "INFO" "Optimizing database performance..."
    
    # Backup current MySQL configuration
    backup_config "/etc/mysql/mariadb.conf.d/50-server.cnf"
    
    # Create MySQL performance optimization configuration
    cat > /etc/mysql/mariadb.conf.d/99-performance.cnf << 'EOF'
[mysqld]
# InnoDB optimizations
innodb_buffer_pool_size = 2G
innodb_log_file_size = 256M
innodb_log_buffer_size = 16M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT
innodb_file_per_table = 1
innodb_open_files = 400
innodb_io_capacity = 400
innodb_read_io_threads = 4
innodb_write_io_threads = 4

# Query cache optimizations
query_cache_type = 1
query_cache_size = 128M
query_cache_limit = 2M

# Connection optimizations
max_connections = 200
max_connect_errors = 10000
connect_timeout = 10
wait_timeout = 600
interactive_timeout = 600

# Temporary table optimizations
tmp_table_size = 128M
max_heap_table_size = 128M

# MyISAM optimizations
key_buffer_size = 32M
read_buffer_size = 2M
read_rnd_buffer_size = 8M
sort_buffer_size = 2M

# Slow query log
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2

# General optimizations
table_open_cache = 4000
thread_cache_size = 16
EOF

    # Restart MariaDB
    systemctl restart mariadb
    
    # Apply runtime optimizations
    mysql -u root -p"${DB_ROOT_PASSWORD:-}" << 'EOF'
-- Apply runtime optimizations
SET GLOBAL innodb_buffer_pool_size = 2147483648;
SET GLOBAL innodb_log_file_size = 268435456;
SET GLOBAL innodb_log_buffer_size = 16777216;
SET GLOBAL innodb_flush_log_at_trx_commit = 2;
SET GLOBAL innodb_flush_method = 'O_DIRECT';
SET GLOBAL innodb_file_per_table = 1;
SET GLOBAL innodb_open_files = 400;
SET GLOBAL innodb_io_capacity = 400;
SET GLOBAL innodb_read_io_threads = 4;
SET GLOBAL innodb_write_io_threads = 4;

SET GLOBAL query_cache_type = 1;
SET GLOBAL query_cache_size = 134217728;
SET GLOBAL query_cache_limit = 2097152;

SET GLOBAL max_connections = 200;
SET GLOBAL max_connect_errors = 10000;
SET GLOBAL connect_timeout = 10;
SET GLOBAL wait_timeout = 600;
SET GLOBAL interactive_timeout = 600;

SET GLOBAL tmp_table_size = 134217728;
SET GLOBAL max_heap_table_size = 134217728;

SET GLOBAL key_buffer_size = 33554432;
SET GLOBAL read_buffer_size = 2097152;
SET GLOBAL read_rnd_buffer_size = 8388608;
SET GLOBAL sort_buffer_size = 2097152;

-- Show current settings
SHOW VARIABLES LIKE 'innodb_buffer_pool_size';
SHOW VARIABLES LIKE 'query_cache_size';
SHOW VARIABLES LIKE 'max_connections';
EOF

    log "INFO" "Database performance optimization completed"
}

# =============================================================================
# PHP Performance Optimization
# =============================================================================

optimize_php() {
    log "INFO" "Optimizing PHP performance..."
    
    # Backup current PHP configuration
    backup_config "/etc/php/8.1/fpm/php.ini"
    
    # Create PHP performance configuration
    cat > /etc/php/8.1/fpm/conf.d/99-performance.ini << 'EOF'
; PHP Performance Optimization for Moodle

; Memory and execution time
memory_limit = 512M
max_execution_time = 300
max_input_time = 300
max_input_vars = 3000

; File uploads
upload_max_filesize = 200M
post_max_size = 200M
max_file_uploads = 20

; OPcache optimization
opcache.enable = 1
opcache.memory_consumption = 256
opcache.interned_strings_buffer = 8
opcache.max_accelerated_files = 10000
opcache.revalidate_freq = 2
opcache.fast_shutdown = 1
opcache.validate_timestamps = 0
opcache.save_comments = 1
opcache.enable_file_override = 1
opcache.optimization_level = 0x7FFFBFFF

; Session optimization
session.gc_maxlifetime = 1440
session.gc_probability = 1
session.gc_divisor = 1000
session.save_handler = redis
session.save_path = "tcp://127.0.0.1:6379"

; Realpath cache
realpath_cache_size = 4096K
realpath_cache_ttl = 600

; Error handling
display_errors = Off
display_startup_errors = Off
log_errors = On
error_log = /var/log/php_errors.log
error_reporting = E_ALL & ~E_DEPRECATED & ~E_STRICT

; Performance settings
allow_url_fopen = Off
allow_url_include = Off
expose_php = Off
EOF

    # Create OPcache blacklist
    cat > /etc/php/8.1/fpm/opcache-blacklist.txt << 'EOF'
/var/www/moodle/config.php
/var/www/moodle/version.php
/var/www/moodle/lib/setup.php
EOF

    # Restart PHP-FPM
    systemctl restart php8.1-fpm
    
    log "INFO" "PHP performance optimization completed"
}

# =============================================================================
# Nginx Performance Optimization
# =============================================================================

optimize_nginx() {
    log "INFO" "Optimizing Nginx performance..."
    
    # Backup current Nginx configuration
    backup_config "/etc/nginx/nginx.conf"
    backup_config "/etc/nginx/sites-available/moodle"
    
    # Create Nginx performance configuration
    cat > /etc/nginx/conf.d/performance.conf << 'EOF'
# Nginx Performance Optimization for Moodle

# Worker processes
worker_processes auto;
worker_cpu_affinity auto;

# Worker connections
events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

# HTTP optimization
http {
    # Basic settings
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    keepalive_requests 100;
    types_hash_max_size 2048;
    server_tokens off;

    # Buffer sizes
    client_body_buffer_size 128k;
    client_max_body_size 200m;
    client_header_buffer_size 1k;
    large_client_header_buffers 4 4k;
    output_buffers 1 32k;
    postpone_output 1460;

    # Timeouts
    client_body_timeout 12;
    client_header_timeout 12;
    send_timeout 10;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/x-javascript
        application/xml+rss
        application/javascript
        application/json
        image/svg+xml;

    # FastCGI optimization
    fastcgi_cache_path /var/cache/nginx/fastcgi levels=1:2 keys_zone=moodle:100m inactive=60m;
    fastcgi_cache_key "$scheme$request_method$host$request_uri";
    fastcgi_cache_use_stale error timeout invalid_header http_500;
    fastcgi_ignore_headers Cache-Control Expires Set-Cookie;

    # Open file cache
    open_file_cache max=1000 inactive=20s;
    open_file_cache_valid 30s;
    open_file_cache_min_uses 2;
    open_file_cache_errors on;
}
EOF

    # Set proper ownership for cache directory
    chown -R www-data:www-data /var/cache/nginx
    
    # Update Moodle site configuration with FastCGI cache
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
    if (\$request_method = POST) {
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
    }
}
EOF

    # Test Nginx configuration
    nginx -t
    
    # Reload Nginx
    systemctl reload nginx
    
    log "INFO" "Nginx performance optimization completed"
}

# =============================================================================
# Redis Performance Optimization
# =============================================================================

optimize_redis() {
    log "INFO" "Optimizing Redis performance..."
    
    # Backup current Redis configuration
    backup_config "/etc/redis/redis.conf"
    
    # Create Redis performance configuration
    cat > /etc/redis/redis.conf << 'EOF'
# Redis Performance Optimization for Moodle

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
EOF

    # Restart Redis
    systemctl restart redis-server
    
    log "INFO" "Redis performance optimization completed"
}

# =============================================================================
# System Performance Optimization
# =============================================================================

optimize_system() {
    log "INFO" "Optimizing system performance..."
    
    # Create system performance optimization configuration
    cat > /etc/sysctl.d/99-moodle-performance.conf << 'EOF'
# Network optimizations
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.core.rmem_default = 262144
net.core.wmem_default = 262144
net.ipv4.tcp_rmem = 4096 65536 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
net.ipv4.tcp_congestion_control = bbr
net.ipv4.tcp_slow_start_after_idle = 0
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.tcp_keepalive_intvl = 30
net.ipv4.tcp_keepalive_probes = 3

# File system optimizations
fs.file-max = 65536
fs.inotify.max_user_watches = 524288
fs.inotify.max_user_instances = 256
fs.inotify.max_queued_events = 32768

# Memory management
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
vm.vfs_cache_pressure = 50

# Process limits
kernel.pid_max = 4194304
kernel.threads-max = 2097152
EOF

    # Apply kernel parameters
    sysctl -p /etc/sysctl.d/99-moodle-performance.conf
    
    # Optimize file limits
    cat > /etc/security/limits.d/99-moodle-performance.conf << 'EOF'
# Moodle performance limits
moodle soft nofile 65536
moodle hard nofile 65536
moodle soft nproc 32768
moodle hard nproc 32768

www-data soft nofile 65536
www-data hard nofile 65536
www-data soft nproc 32768
www-data hard nproc 32768

root soft nofile 65536
root hard nofile 65536
root soft nproc 32768
root hard nproc 32768
EOF

    # Optimize I/O scheduler (if possible)
    if [[ -f "/sys/block/sda/queue/scheduler" ]]; then
        echo mq-deadline > /sys/block/sda/queue/scheduler
        echo 1024 > /sys/block/sda/queue/nr_requests
        log "INFO" "I/O scheduler optimized"
    fi
    
    # Optimize CPU governor (if possible)
    if [[ -d "/sys/devices/system/cpu/cpu0/cpufreq" ]]; then
        echo performance > /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor 2>/dev/null || true
        log "INFO" "CPU governor optimized"
    fi
    
    log "INFO" "System performance optimization completed"
}

# =============================================================================
# Performance Monitoring Setup
# =============================================================================

setup_performance_monitoring() {
    log "INFO" "Setting up performance monitoring..."
    
    # Create performance monitoring script
    cat > /usr/local/bin/performance-monitor.sh << 'EOF'
#!/bin/bash

# Performance Monitoring for Moodle
LOG_FILE="/var/log/performance-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Performance monitoring..." >> $LOG_FILE

# CPU usage
CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
echo "CPU Usage: $CPU_USAGE%" >> $LOG_FILE

# Memory usage
MEMORY_USAGE=$(free | awk 'NR==2{printf "%.2f", $3*100/$2}')
echo "Memory Usage: $MEMORY_USAGE%" >> $LOG_FILE

# Disk I/O
DISK_IO=$(iostat -x 1 1 | grep -E "(Device|sda)" | tail -1)
echo "Disk I/O: $DISK_IO" >> $LOG_FILE

# Network usage
NETWORK_USAGE=$(iftop -t -s 1 -L 1 | grep -E "(TX|RX)" | tail -2)
echo "Network Usage: $NETWORK_USAGE" >> $LOG_FILE

# Database connections
DB_CONNECTIONS=$(mysql -u root -p'${DB_ROOT_PASSWORD:-}' -e "SHOW STATUS LIKE 'Threads_connected';" 2>/dev/null | awk 'NR==2 {print $2}' || echo "N/A")
echo "Database Connections: $DB_CONNECTIONS" >> $LOG_FILE

# Redis memory usage
REDIS_MEMORY=$(redis-cli info memory | grep used_memory_human | cut -d: -f2)
echo "Redis Memory: $REDIS_MEMORY" >> $LOG_FILE

# Nginx active connections
NGINX_CONNECTIONS=$(netstat -an | grep :443 | grep ESTABLISHED | wc -l)
echo "Nginx Active Connections: $NGINX_CONNECTIONS" >> $LOG_FILE

# PHP-FPM processes
PHP_FPM_PROCESSES=$(ps aux | grep php-fpm | grep -v grep | wc -l)
echo "PHP-FPM Processes: $PHP_FPM_PROCESSES" >> $LOG_FILE

echo "---" >> $LOG_FILE
EOF

    # Make script executable
    chmod +x /usr/local/bin/performance-monitor.sh
    
    # Add to crontab
    (crontab -l 2>/dev/null; echo "*/5 * * * * /usr/local/bin/performance-monitor.sh") | crontab -
    
    log "INFO" "Performance monitoring setup completed"
}

# =============================================================================
# Verification Functions
# =============================================================================

verify_performance_optimization() {
    log "INFO" "Verifying performance optimization..."
    
    # Check database settings
    log "INFO" "Checking database settings..."
    mysql -u root -p"${DB_ROOT_PASSWORD:-}" -e "SHOW VARIABLES LIKE 'innodb_buffer_pool_size';" 2>/dev/null || log "WARN" "Could not check database settings"
    
    # Check PHP settings
    log "INFO" "Checking PHP settings..."
    php -i | grep -E "(memory_limit|opcache)" | head -5 || log "WARN" "Could not check PHP settings"
    
    # Check Nginx settings
    log "INFO" "Checking Nginx settings..."
    nginx -T | grep -E "(worker_processes|worker_connections)" || log "WARN" "Could not check Nginx settings"
    
    # Check Redis settings
    log "INFO" "Checking Redis settings..."
    redis-cli info memory | head -5 || log "WARN" "Could not check Redis settings"
    
    # Check system settings
    log "INFO" "Checking system settings..."
    sysctl net.core.rmem_max
    sysctl vm.swappiness
    
    log "INFO" "Performance optimization verification completed"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    log "INFO" "Starting performance tuning process..."
    
    # Check prerequisites
    check_root
    create_directories
    
    # Execute optimization steps
    optimize_database
    optimize_php
    optimize_nginx
    optimize_redis
    optimize_system
    setup_performance_monitoring
    
    # Verify optimization
    verify_performance_optimization
    
    log "INFO" "Performance tuning process completed successfully!"
    log "INFO" "Log file: $LOG_FILE"
    log "INFO" "Backup directory: $BACKUP_DIR"
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
        echo "  --verify       Only verify current optimization"
        echo "  --dry-run      Show what would be done without executing"
        echo "  --database     Optimize database only"
        echo "  --php          Optimize PHP only"
        echo "  --nginx        Optimize Nginx only"
        echo "  --redis        Optimize Redis only"
        echo "  --system       Optimize system only"
        exit 0
        ;;
    --verify)
        check_root
        verify_performance_optimization
        exit 0
        ;;
    --dry-run)
        log "INFO" "DRY RUN MODE - No changes will be made"
        log "INFO" "Would execute: optimize_database, optimize_php, optimize_nginx, optimize_redis, optimize_system, setup_performance_monitoring"
        exit 0
        ;;
    --database)
        check_root
        create_directories
        optimize_database
        log "INFO" "Database optimization completed"
        ;;
    --php)
        check_root
        create_directories
        optimize_php
        log "INFO" "PHP optimization completed"
        ;;
    --nginx)
        check_root
        create_directories
        optimize_nginx
        log "INFO" "Nginx optimization completed"
        ;;
    --redis)
        check_root
        create_directories
        optimize_redis
        log "INFO" "Redis optimization completed"
        ;;
    --system)
        check_root
        create_directories
        optimize_system
        log "INFO" "System optimization completed"
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
