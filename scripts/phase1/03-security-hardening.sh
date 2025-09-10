#!/bin/bash

# =============================================================================
# LMSK2 Moodle Server - Security Hardening Script
# =============================================================================
# Description: Advanced security configuration for Moodle server
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
CONFIG_FILE="${SCRIPT_DIR}/config/phase1.conf"

if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
else
    echo "❌ Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Script configuration
SCRIPT_NAME="$(basename "$0")"
LOG_FILE="/var/log/lmsk2/${SCRIPT_NAME%.*}.log"
BACKUP_DIR="/backup/lmsk2/security"

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
    mkdir -p "/etc/fail2ban/jail.d"
    mkdir -p "/etc/fail2ban/filter.d"
    mkdir -p "/etc/nginx/snippets"
    mkdir -p "/etc/mysql/mariadb.conf.d"
    mkdir -p "/etc/php/8.1/fpm/conf.d"
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
# Security Functions
# =============================================================================

advanced_firewall_configuration() {
    log "INFO" "Configuring advanced firewall settings..."
    
    # Install additional security tools
    apt update
    apt install -y fail2ban ufw iptables-persistent
    
    # Backup current UFW configuration
    backup_config "/etc/ufw/before.rules"
    backup_config "/etc/ufw/after.rules"
    
    # Reset and configure UFW
    ufw --force reset
    ufw default deny incoming
    ufw default allow outgoing
    
    # Allow essential services
    ufw allow ssh
    ufw allow 80/tcp
    ufw allow 443/tcp
    
    # Allow specific IP ranges if configured
    if [[ -n "${ALLOWED_IP_RANGE:-}" ]]; then
        ufw allow from "$ALLOWED_IP_RANGE" to any port 22
        ufw allow from "$ALLOWED_IP_RANGE" to any port 80
        ufw allow from "$ALLOWED_IP_RANGE" to any port 443
        log "INFO" "Allowed IP range: $ALLOWED_IP_RANGE"
    fi
    
    # Enable firewall
    ufw --force enable
    
    # Check status
    ufw status verbose
    
    log "INFO" "Advanced firewall configuration completed"
}

configure_fail2ban() {
    log "INFO" "Configuring Fail2ban..."
    
    # Create Fail2ban jail configuration
    cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
# Ban hosts for 1 hour
bantime = 3600
# Override /etc/fail2ban/jail.d/00-firewalld.conf
banaction = ufw
# Number of failures before ban
maxretry = 3
# Time window for failures
findtime = 600
# Ignore local IPs
ignoreip = 127.0.0.1/8 ::1 192.168.88.0/24

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600

[nginx-http-auth]
enabled = true
filter = nginx-http-auth
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 3

[nginx-limit-req]
enabled = true
filter = nginx-limit-req
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 3

[php-url-fopen]
enabled = true
filter = php-url-fopen
port = http,https
logpath = /var/log/nginx/access.log
maxretry = 3
EOF

    # Create custom filters
    cat > /etc/fail2ban/filter.d/nginx-http-auth.conf << 'EOF'
[Definition]
failregex = ^<HOST> -.*"(GET|POST).*HTTP.*" (401|403) .*$
ignoreregex =
EOF

    cat > /etc/fail2ban/filter.d/nginx-limit-req.conf << 'EOF'
[Definition]
failregex = limiting requests, excess: .* by zone .*, client: <HOST>
ignoreregex =
EOF

    cat > /etc/fail2ban/filter.d/php-url-fopen.conf << 'EOF'
[Definition]
failregex = ^<HOST> -.*"(GET|POST).*\.php.*" (200|404) .*$
ignoreregex =
EOF

    # Start and enable Fail2ban
    systemctl start fail2ban
    systemctl enable fail2ban
    
    # Check status
    fail2ban-client status
    
    log "INFO" "Fail2ban configuration completed"
}

setup_ssl_certificate() {
    log "INFO" "Setting up SSL certificate..."
    
    # Install Certbot
    apt install -y certbot python3-certbot-nginx
    
    # Check if domain is configured
    if [[ -z "${MOODLE_DOMAIN:-}" ]]; then
        log "WARN" "MOODLE_DOMAIN not configured, skipping SSL setup"
        return 0
    fi
    
    # Generate SSL certificate
    if certbot --nginx -d "$MOODLE_DOMAIN" --non-interactive --agree-tos --email "${ADMIN_EMAIL:-admin@example.com}"; then
        log "INFO" "SSL certificate generated successfully for $MOODLE_DOMAIN"
        
        # Test certificate renewal
        certbot renew --dry-run
        
        # Setup automatic renewal
        (crontab -l 2>/dev/null; echo "0 12 * * * /usr/bin/certbot renew --quiet") | crontab -
        
        log "INFO" "SSL certificate setup completed"
    else
        log "ERROR" "Failed to generate SSL certificate"
        return 1
    fi
}

configure_nginx_security() {
    log "INFO" "Configuring Nginx security settings..."
    
    # Backup current Nginx configuration
    backup_config "/etc/nginx/sites-available/moodle"
    
    # Create enhanced Nginx configuration
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
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'self';" always;

    # Rate limiting
    limit_req_zone \$binary_remote_addr zone=login:10m rate=5r/m;
    limit_req_zone \$binary_remote_addr zone=api:10m rate=10r/s;

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

    # PHP processing
    location ~ \.php\$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php8.1-fpm.sock;
        fastcgi_param SCRIPT_FILENAME \$document_root\$fastcgi_script_name;
        include fastcgi_params;
        
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
    
    log "INFO" "Nginx security configuration completed"
}

configure_file_permissions() {
    log "INFO" "Configuring file permissions and ownership..."
    
    # Set proper ownership
    chown -R www-data:www-data /var/www/moodle
    
    # Set proper permissions
    find /var/www/moodle -type d -exec chmod 755 {} \;
    find /var/www/moodle -type f -exec chmod 644 {} \;
    
    # Make moodledata writable
    chmod -R 777 /var/www/moodle/moodledata
    
    # Secure config.php if it exists
    if [[ -f "/var/www/moodle/config.php" ]]; then
        chmod 600 /var/www/moodle/config.php
        chown root:www-data /var/www/moodle/config.php
        log "INFO" "Secured config.php permissions"
    fi
    
    log "INFO" "File permissions configuration completed"
}

configure_mariadb_security() {
    log "INFO" "Configuring MariaDB security..."
    
    # Create MariaDB security configuration
    cat > /etc/mysql/mariadb.conf.d/99-security.cnf << 'EOF'
[mysqld]
# Security settings
local-infile = 0
symbolic-links = 0
skip-networking = 0
bind-address = 127.0.0.1

# Logging
log_error = /var/log/mysql/error.log
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2

# Connection limits
max_connections = 100
max_connect_errors = 10
connect_timeout = 10
wait_timeout = 600
interactive_timeout = 600

# Disable dangerous functions
sql_mode = STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO
EOF

    # Restart MariaDB
    systemctl restart mariadb
    
    # Secure MariaDB installation
    mysql -u root -p"${DB_ROOT_PASSWORD:-}" << 'EOF'
-- Remove test database
DROP DATABASE IF EXISTS test;
DELETE FROM mysql.db WHERE Db='test' OR Db='test\_%';

-- Remove anonymous users
DELETE FROM mysql.user WHERE User='';
DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');

-- Create dedicated user for backups
CREATE USER IF NOT EXISTS 'backup'@'localhost' IDENTIFIED BY 'strong_backup_password_2025';
GRANT SELECT, LOCK TABLES, SHOW VIEW, EVENT, TRIGGER ON *.* TO 'backup'@'localhost';

-- Flush privileges
FLUSH PRIVILEGES;
EOF

    log "INFO" "MariaDB security configuration completed"
}

configure_php_security() {
    log "INFO" "Configuring PHP security settings..."
    
    # Create PHP security configuration
    cat > /etc/php/8.1/fpm/conf.d/99-security.ini << 'EOF'
; Security settings
expose_php = Off
allow_url_fopen = Off
allow_url_include = Off
disable_functions = exec,passthru,shell_exec,system,proc_open,popen,curl_exec,curl_multi_exec,parse_ini_file,show_source
disable_classes = 

; File upload security
file_uploads = On
upload_max_filesize = 200M
post_max_size = 200M
max_file_uploads = 20

; Session security
session.cookie_httponly = 1
session.cookie_secure = 1
session.use_strict_mode = 1
session.cookie_samesite = "Strict"

; Error handling
display_errors = Off
display_startup_errors = Off
log_errors = On
error_log = /var/log/php_errors.log
EOF

    # Restart PHP-FPM
    systemctl restart php8.1-fpm
    
    log "INFO" "PHP security configuration completed"
}

setup_log_monitoring() {
    log "INFO" "Setting up log monitoring..."
    
    # Install log monitoring tools
    apt install -y logwatch rsyslog
    
    # Configure log rotation
    cat > /etc/logrotate.d/moodle << 'EOF'
/var/log/nginx/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 www-data adm
    postrotate
        if [ -f /var/run/nginx.pid ]; then
            kill -USR1 $(cat /var/run/nginx.pid)
        fi
    endscript
}

/var/log/php_errors.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 www-data adm
    postrotate
        systemctl reload php8.1-fpm
    endscript
}
EOF

    # Create security monitoring script
    cat > /usr/local/bin/security-monitor.sh << 'EOF'
#!/bin/bash

# Security monitoring script
LOG_FILE="/var/log/security-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Starting security check..." >> $LOG_FILE

# Check for failed login attempts
FAILED_LOGINS=$(grep "Failed password" /var/log/auth.log 2>/dev/null | wc -l)
if [ $FAILED_LOGINS -gt 10 ]; then
    echo "[$DATE] WARNING: High number of failed login attempts: $FAILED_LOGINS" >> $LOG_FILE
fi

# Check for suspicious PHP errors
PHP_ERRORS=$(grep "PHP" /var/log/nginx/error.log 2>/dev/null | wc -l)
if [ $PHP_ERRORS -gt 50 ]; then
    echo "[$DATE] WARNING: High number of PHP errors: $PHP_ERRORS" >> $LOG_FILE
fi

# Check disk space
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    echo "[$DATE] WARNING: Disk usage high: $DISK_USAGE%" >> $LOG_FILE
fi

echo "[$DATE] Security check completed." >> $LOG_FILE
EOF

    # Make script executable
    chmod +x /usr/local/bin/security-monitor.sh
    
    # Add to crontab
    (crontab -l 2>/dev/null; echo "0 * * * * /usr/local/bin/security-monitor.sh") | crontab -
    
    log "INFO" "Log monitoring setup completed"
}

# =============================================================================
# Verification Functions
# =============================================================================

verify_security_configuration() {
    log "INFO" "Verifying security configuration..."
    
    # Check firewall status
    log "INFO" "Checking firewall status..."
    ufw status verbose
    
    # Check Fail2ban status
    log "INFO" "Checking Fail2ban status..."
    fail2ban-client status
    
    # Check SSL certificate if domain is configured
    if [[ -n "${MOODLE_DOMAIN:-}" ]]; then
        log "INFO" "Checking SSL certificate..."
        certbot certificates
    fi
    
    # Check file permissions
    log "INFO" "Checking file permissions..."
    ls -la /var/www/moodle/ | head -10
    
    # Check MariaDB security
    log "INFO" "Checking MariaDB users..."
    mysql -u root -p"${DB_ROOT_PASSWORD:-}" -e "SELECT user, host FROM mysql.user;" 2>/dev/null || log "WARN" "Could not check MariaDB users"
    
    # Check PHP security
    log "INFO" "Checking PHP security settings..."
    php -i | grep -E "(expose_php|allow_url_fopen|disable_functions)" || log "WARN" "Could not check PHP settings"
    
    # Check log files
    log "INFO" "Checking log files..."
    ls -la /var/log/ | grep -E "(fail2ban|security-monitor)" || log "WARN" "Security log files not found"
    
    log "INFO" "Security configuration verification completed"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    log "INFO" "Starting security hardening process..."
    
    # Check prerequisites
    check_root
    create_directories
    
    # Execute security configurations
    advanced_firewall_configuration
    configure_fail2ban
    setup_ssl_certificate
    configure_nginx_security
    configure_file_permissions
    configure_mariadb_security
    configure_php_security
    setup_log_monitoring
    
    # Verify configuration
    verify_security_configuration
    
    log "INFO" "Security hardening process completed successfully!"
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
        echo "  --verify       Only verify current configuration"
        echo "  --dry-run      Show what would be done without executing"
        exit 0
        ;;
    --verify)
        check_root
        verify_security_configuration
        exit 0
        ;;
    --dry-run)
        log "INFO" "DRY RUN MODE - No changes will be made"
        log "INFO" "Would execute: advanced_firewall_configuration, configure_fail2ban, setup_ssl_certificate, configure_nginx_security, configure_file_permissions, configure_mariadb_security, configure_php_security, setup_log_monitoring"
        exit 0
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
