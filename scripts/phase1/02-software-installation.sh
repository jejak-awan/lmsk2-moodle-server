#!/bin/bash

# =============================================================================
# Phase 1: Software Installation
# =============================================================================
# Version: 1.0
# Author: jejakawan007
# Description: Software installation untuk LMSK2-Moodle-Server
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Script information
SCRIPT_NAME="Phase 1: Software Installation"
SCRIPT_VERSION="1.0"
SCRIPT_AUTHOR="jejakawan007"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# =============================================================================
# Utility Functions
# =============================================================================

# Print colored output
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Print section header
print_section() {
    local section=$1
    echo
    print_color $YELLOW ">>> $section"
    print_color $YELLOW "=============================================================================="
}

# Print success message
print_success() {
    print_color $GREEN "✓ $1"
}

# Print error message
print_error() {
    print_color $RED "✗ $1"
}

# Print warning message
print_warning() {
    print_color $YELLOW "⚠ $1"
}

# Print info message
print_info() {
    print_color $BLUE "ℹ $1"
}

# Log message
log_message() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Create log directory if it doesn't exist
    mkdir -p "/var/log/lmsk2"
    
    # Write to log file
    echo "[$timestamp] [$level] $message" >> "/var/log/lmsk2/phase1.log"
    
    # Print to console based on log level
    case $level in
        "ERROR")
            print_error "$message"
            ;;
        "WARNING")
            print_warning "$message"
            ;;
        "INFO")
            print_info "$message"
            ;;
        "SUCCESS")
            print_success "$message"
            ;;
    esac
}

# =============================================================================
# Main Functions
# =============================================================================

# Install Nginx
install_nginx() {
    print_section "Installing Nginx"
    
    # Update package list
    print_info "Updating package list..."
    apt update
    
    # Install Nginx
    print_info "Installing Nginx..."
    apt install -y nginx
    
    # Start and enable Nginx
    print_info "Starting and enabling Nginx..."
    systemctl start nginx
    systemctl enable nginx
    
    # Check Nginx status
    print_info "Checking Nginx status..."
    systemctl status nginx --no-pager -l
    
    # Test Nginx
    print_info "Testing Nginx..."
    if curl -I http://localhost >/dev/null 2>&1; then
        print_success "Nginx is working correctly"
    else
        print_error "Nginx test failed"
        return 1
    fi
    
    log_message "SUCCESS" "Nginx installation completed"
}

# Install PHP
install_php() {
    print_section "Installing PHP"
    
    # Add PHP repository
    print_info "Adding PHP repository..."
    apt install -y software-properties-common
    add-apt-repository ppa:ondrej/php -y
    apt update
    
    # Install PHP 8.1 and extensions
    print_info "Installing PHP 8.1 and extensions..."
    apt install -y php8.1-fpm php8.1-cli php8.1-common \
        php8.1-mysql php8.1-zip php8.1-gd php8.1-mbstring \
        php8.1-curl php8.1-xml php8.1-bcmath php8.1-intl \
        php8.1-soap php8.1-ldap php8.1-imagick php8.1-redis \
        php8.1-openssl php8.1-json php8.1-dom php8.1-fileinfo \
        php8.1-iconv php8.1-simplexml php8.1-tokenizer \
        php8.1-xmlreader php8.1-xmlwriter php8.1-exif \
        php8.1-ftp php8.1-gettext
    
    # Start and enable PHP-FPM
    print_info "Starting and enabling PHP-FPM..."
    systemctl start php8.1-fpm
    systemctl enable php8.1-fpm
    
    # Check PHP version
    print_info "PHP version:"
    php -v
    
    # Check PHP extensions
    print_info "Checking PHP extensions..."
    php -m | grep -E "(mysql|gd|curl|xml|mbstring|zip|intl|soap|ldap|imagick|redis)" || true
    
    log_message "SUCCESS" "PHP installation completed"
}

# Configure PHP
configure_php() {
    print_section "Configuring PHP"
    
    # Backup original php.ini
    print_info "Backing up original php.ini..."
    cp /etc/php/8.1/fpm/php.ini /etc/php/8.1/fpm/php.ini.backup.$(date +%Y%m%d_%H%M%S)
    
    # Configure PHP for Moodle
    print_info "Configuring PHP for Moodle..."
    
    # Create PHP configuration
    cat > /etc/php/8.1/fpm/conf.d/99-moodle.ini << EOF
; PHP Configuration for Moodle

; Memory and execution time
memory_limit = 512M
max_execution_time = 300
max_input_time = 300
max_input_vars = 3000

; File uploads
upload_max_filesize = 200M
post_max_size = 200M
max_file_uploads = 20

; Session configuration
session.gc_maxlifetime = 1440
session.save_handler = redis
session.save_path = "tcp://127.0.0.1:6379"

; OPcache configuration
opcache.enable = 1
opcache.memory_consumption = 256
opcache.interned_strings_buffer = 8
opcache.max_accelerated_files = 10000
opcache.revalidate_freq = 2
opcache.fast_shutdown = 1
opcache.validate_timestamps = 0

; Error reporting (disable in production)
display_errors = Off
log_errors = On
error_log = /var/log/php_errors.log

; Security settings
expose_php = Off
allow_url_fopen = Off
allow_url_include = Off
disable_functions = exec,passthru,shell_exec,system,proc_open,popen,curl_exec,curl_multi_exec,parse_ini_file,show_source

; File upload security
file_uploads = On

; Session security
session.cookie_httponly = 1
session.cookie_secure = 1
session.use_strict_mode = 1
session.cookie_samesite = "Strict"

; Realpath cache
realpath_cache_size = 4096K
realpath_cache_ttl = 600
EOF
    
    # Restart PHP-FPM
    print_info "Restarting PHP-FPM..."
    systemctl restart php8.1-fpm
    
    # Test PHP configuration
    print_info "Testing PHP configuration..."
    php -i | grep -E "(memory_limit|upload_max_filesize|opcache.enable)" || true
    
    log_message "SUCCESS" "PHP configuration completed"
}

# Install MariaDB
install_mariadb() {
    print_section "Installing MariaDB"
    
    # Install MariaDB
    print_info "Installing MariaDB..."
    apt install -y mariadb-server mariadb-client
    
    # Start and enable MariaDB
    print_info "Starting and enabling MariaDB..."
    systemctl start mariadb
    systemctl enable mariadb
    
    # Secure MariaDB installation
    print_info "Securing MariaDB installation..."
    
    # Generate random password if not provided
    local db_password="${LMSK2_DB_PASSWORD:-}"
    if [[ -z "$db_password" ]]; then
        db_password=$(openssl rand -base64 32)
        print_info "Generated database password: $db_password"
    fi
    
    # Create secure installation script
    cat > /tmp/mysql_secure.sql << EOF
-- Set root password
ALTER USER 'root'@'localhost' IDENTIFIED BY '$db_password';

-- Remove anonymous users
DELETE FROM mysql.user WHERE User='';

-- Remove test database
DROP DATABASE IF EXISTS test;
DELETE FROM mysql.db WHERE Db='test' OR Db='test\\_%';

-- Remove root remote access
DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');

-- Create moodle database and user
CREATE DATABASE IF NOT EXISTS moodle CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'moodle'@'localhost' IDENTIFIED BY '$db_password';
GRANT ALL PRIVILEGES ON moodle.* TO 'moodle'@'localhost';

-- Flush privileges
FLUSH PRIVILEGES;
EOF
    
    # Execute secure installation
    mysql -u root < /tmp/mysql_secure.sql
    rm /tmp/mysql_secure.sql
    
    # Store password for later use
    echo "LMSK2_DB_PASSWORD=$db_password" >> /etc/environment
    
    print_success "MariaDB secured and configured"
    
    # Test MariaDB connection
    print_info "Testing MariaDB connection..."
    if mysql -u moodle -p"$db_password" -e "SELECT 1;" >/dev/null 2>&1; then
        print_success "MariaDB connection test successful"
    else
        print_error "MariaDB connection test failed"
        return 1
    fi
    
    log_message "SUCCESS" "MariaDB installation completed"
}

# Configure MariaDB
configure_mariadb() {
    print_section "Configuring MariaDB"
    
    # Backup original configuration
    print_info "Backing up original MariaDB configuration..."
    cp /etc/mysql/mariadb.conf.d/50-server.cnf /etc/mysql/mariadb.conf.d/50-server.cnf.backup.$(date +%Y%m%d_%H%M%S)
    
    # Create MariaDB optimization configuration
    cat > /etc/mysql/mariadb.conf.d/99-moodle-optimization.cnf << EOF
[mysqld]
# Basic settings
bind-address = 127.0.0.1
port = 3306
socket = /var/run/mysqld/mysqld.sock

# Character set
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci

# InnoDB settings
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT
innodb_file_per_table = 1
innodb_open_files = 400
innodb_io_capacity = 400
innodb_read_io_threads = 4
innodb_write_io_threads = 4

# Query cache
query_cache_type = 1
query_cache_size = 64M
query_cache_limit = 2M

# Connection settings
max_connections = 200
max_connect_errors = 10000
connect_timeout = 10
wait_timeout = 600
interactive_timeout = 600

# Temporary tables
tmp_table_size = 64M
max_heap_table_size = 64M

# MyISAM settings
key_buffer_size = 32M
read_buffer_size = 2M
read_rnd_buffer_size = 8M
sort_buffer_size = 2M

# Logging
log_error = /var/log/mysql/error.log
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2

# Security settings
local-infile = 0
symbolic-links = 0
skip-networking = 0

# SQL mode
sql_mode = STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO
EOF
    
    # Restart MariaDB
    print_info "Restarting MariaDB..."
    systemctl restart mariadb
    
    # Test MariaDB configuration
    print_info "Testing MariaDB configuration..."
    mysql -u root -p"${LMSK2_DB_PASSWORD:-}" -e "SHOW VARIABLES LIKE 'innodb_buffer_pool_size';" || true
    
    log_message "SUCCESS" "MariaDB configuration completed"
}

# Install Redis
install_redis() {
    print_section "Installing Redis"
    
    # Install Redis
    print_info "Installing Redis..."
    apt install -y redis-server
    
    # Configure Redis
    print_info "Configuring Redis..."
    
    # Backup original configuration
    cp /etc/redis/redis.conf /etc/redis/redis.conf.backup.$(date +%Y%m%d_%H%M%S)
    
    # Create Redis configuration
    cat > /etc/redis/redis.conf << EOF
# Redis Configuration for Moodle

# Network
bind 127.0.0.1
port 6379
timeout 0
tcp-keepalive 300

# Memory management
maxmemory 256mb
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
    
    # Start and enable Redis
    print_info "Starting and enabling Redis..."
    systemctl start redis-server
    systemctl enable redis-server
    
    # Test Redis
    print_info "Testing Redis..."
    if redis-cli ping >/dev/null 2>&1; then
        print_success "Redis is working correctly"
    else
        print_error "Redis test failed"
        return 1
    fi
    
    log_message "SUCCESS" "Redis installation completed"
}

# Install additional tools
install_additional_tools() {
    print_section "Installing Additional Tools"
    
    # Install Composer
    print_info "Installing Composer..."
    curl -sS https://getcomposer.org/installer | php
    mv composer.phar /usr/local/bin/composer
    chmod +x /usr/local/bin/composer
    
    # Install Node.js
    print_info "Installing Node.js..."
    curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
    apt install -y nodejs
    
    # Install additional tools
    print_info "Installing additional tools..."
    apt install -y imagemagick ghostscript unoconv libreoffice
    apt install -y cron rsync tar gzip zip unzip
    apt install -y build-essential make cmake
    apt install -y htop iotop nethogs iftop nload
    
    # Check installations
    print_info "Checking installations..."
    composer --version
    node --version
    npm --version
    
    log_message "SUCCESS" "Additional tools installation completed"
}

# Configure Nginx
configure_nginx() {
    print_section "Configuring Nginx"
    
    # Backup original configuration
    print_info "Backing up original Nginx configuration..."
    cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup.$(date +%Y%m%d_%H%M%S)
    
    # Create Nginx configuration for Moodle
    print_info "Creating Nginx configuration for Moodle..."
    
    cat > /etc/nginx/sites-available/moodle << EOF
server {
    listen 80;
    server_name ${LMSK2_DOMAIN:-lms-server.local};
    root /var/www/moodle;
    index index.php index.html;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied expired no-cache no-store private must-revalidate auth;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss;

    # Main location
    location / {
        try_files \$uri \$uri/ /index.php?\$query_string;
    }

    # PHP processing
    location ~ \.php\$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php8.1-fpm.sock;
        fastcgi_param SCRIPT_FILENAME \$document_root\$fastcgi_script_name;
        include fastcgi_params;
    }

    # Deny access to sensitive files
    location ~ /\. {
        deny all;
    }

    location ~ /(config|cache|local|moodledata|backup|temp|lang|pix|theme|userpix|upgrade|admin|lib|install|test|vendor)/ {
        deny all;
    }

    # Static files caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)\$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
EOF
    
    # Enable site
    print_info "Enabling Moodle site..."
    ln -sf /etc/nginx/sites-available/moodle /etc/nginx/sites-enabled/
    rm -f /etc/nginx/sites-enabled/default
    
    # Test Nginx configuration
    print_info "Testing Nginx configuration..."
    if nginx -t; then
        print_success "Nginx configuration is valid"
    else
        print_error "Nginx configuration test failed"
        return 1
    fi
    
    # Restart Nginx
    print_info "Restarting Nginx..."
    systemctl restart nginx
    
    log_message "SUCCESS" "Nginx configuration completed"
}

# Verification
verification() {
    print_section "Verification"
    
    print_info "Service status:"
    systemctl status nginx --no-pager -l | head -10
    echo
    systemctl status php8.1-fpm --no-pager -l | head -10
    echo
    systemctl status mariadb --no-pager -l | head -10
    echo
    systemctl status redis-server --no-pager -l | head -10
    echo
    
    print_info "PHP version and extensions:"
    php -v
    echo
    php -m | grep -E "(mysql|gd|curl|xml|mbstring|zip|intl|soap|ldap|imagick|redis)" || true
    echo
    
    print_info "MariaDB test:"
    mysql -u root -p"${LMSK2_DB_PASSWORD:-}" -e "SHOW DATABASES;" || true
    echo
    
    print_info "Redis test:"
    redis-cli ping || true
    echo
    
    print_info "Composer version:"
    composer --version || true
    echo
    
    print_info "Node.js version:"
    node --version || true
    npm --version || true
    echo
    
    print_info "Nginx test:"
    curl -I http://localhost || true
    echo
    
    print_success "Verification completed"
    log_message "SUCCESS" "Verification completed"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    print_color $CYAN "=============================================================================="
    print_color $WHITE "  $SCRIPT_NAME v$SCRIPT_VERSION"
    print_color $CYAN "=============================================================================="
    
    log_message "INFO" "Starting Phase 1: Software Installation"
    
    # Execute all functions
    install_nginx
    install_php
    configure_php
    install_mariadb
    configure_mariadb
    install_redis
    install_additional_tools
    configure_nginx
    verification
    
    print_section "Phase 1 Complete"
    print_success "Software installation completed successfully!"
    print_info "Log file: /var/log/lmsk2/phase1.log"
    
    log_message "SUCCESS" "Phase 1: Software Installation completed successfully"
}

# Trap errors
trap 'print_error "Script failed at line $LINENO"; log_message "ERROR" "Script failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
