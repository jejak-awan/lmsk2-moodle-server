#!/bin/bash

# =============================================================================
# LMSK2 Moodle Server - Moodle 4.0 Installation Script
# =============================================================================
# Description: Install Moodle 4.0 with optimal configuration
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
BACKUP_DIR="/backup/lmsk2/moodle-4.0-install"
MOODLE_VERSION="4.0"
MOODLE_DOWNLOAD_URL="https://download.moodle.org/releases/latest/moodle-4.0-latest.tgz"
MOODLE_DIR="/var/www/moodle"
MOODLE_DATA_DIR="/var/www/moodledata"

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
    mkdir -p "/tmp/moodle-install"
    mkdir -p "$MOODLE_DIR"
    mkdir -p "$MOODLE_DATA_DIR"
    
    log "INFO" "Directories created successfully"
}

backup_existing() {
    if [[ -d "$MOODLE_DIR" && "$(ls -A "$MOODLE_DIR" 2>/dev/null)" ]]; then
        log "WARN" "Existing Moodle installation found, creating backup..."
        local backup_name="moodle-backup-$(date +%Y%m%d_%H%M%S)"
        cp -r "$MOODLE_DIR" "$BACKUP_DIR/$backup_name"
        log "INFO" "Backup created: $BACKUP_DIR/$backup_name"
    fi
}

# =============================================================================
# Installation Functions
# =============================================================================

download_moodle() {
    log "INFO" "Downloading Moodle $MOODLE_VERSION..."
    
    cd /tmp/moodle-install
    
    # Download Moodle
    if wget -O moodle-4.0-latest.tgz "$MOODLE_DOWNLOAD_URL"; then
        log "INFO" "Moodle downloaded successfully"
    else
        log "ERROR" "Failed to download Moodle"
        exit 1
    fi
    
    # Verify download
    if [[ -f "moodle-4.0-latest.tgz" ]]; then
        local file_size=$(stat -c%s "moodle-4.0-latest.tgz")
        log "INFO" "Downloaded file size: $file_size bytes"
    else
        log "ERROR" "Downloaded file not found"
        exit 1
    fi
    
    # Extract Moodle
    log "INFO" "Extracting Moodle..."
    if tar -xzf moodle-4.0-latest.tgz; then
        log "INFO" "Moodle extracted successfully"
    else
        log "ERROR" "Failed to extract Moodle"
        exit 1
    fi
    
    # Check extracted files
    if [[ -d "moodle" ]]; then
        log "INFO" "Moodle directory created successfully"
        ls -la moodle/ | head -10
    else
        log "ERROR" "Moodle directory not found after extraction"
        exit 1
    fi
}

install_moodle_files() {
    log "INFO" "Installing Moodle files..."
    
    # Stop web server temporarily
    log "INFO" "Stopping web server..."
    systemctl stop nginx
    
    # Move Moodle files to web directory
    log "INFO" "Moving Moodle files to web directory..."
    if [[ -d "/tmp/moodle-install/moodle" ]]; then
        cp -r /tmp/moodle-install/moodle/* "$MOODLE_DIR/"
        log "INFO" "Moodle files moved successfully"
    else
        log "ERROR" "Source Moodle directory not found"
        exit 1
    fi
    
    # Set proper ownership
    log "INFO" "Setting file ownership..."
    chown -R www-data:www-data "$MOODLE_DIR"
    
    # Set proper permissions
    log "INFO" "Setting file permissions..."
    chmod -R 755 "$MOODLE_DIR"
    
    # Set data directory permissions
    chmod -R 777 "$MOODLE_DATA_DIR"
    
    # Start web server
    log "INFO" "Starting web server..."
    systemctl start nginx
    
    log "INFO" "Moodle files installed successfully"
}

create_moodle_config() {
    log "INFO" "Creating Moodle configuration file..."
    
    # Create Moodle configuration file
    cat > "$MOODLE_DIR/config.php" << EOF
<?php
// Moodle 4.0 Configuration

unset(\$CFG);
global \$CFG;
\$CFG = new stdClass();

// Database configuration
\$CFG->dbtype    = '${DB_TYPE:-mariadb}';
\$CFG->dblibrary = 'native';
\$CFG->dbhost    = '${DB_HOST:-localhost}';
\$CFG->dbname    = '${DB_NAME:-moodle}';
\$CFG->dbuser    = '${DB_USER:-moodle}';
\$CFG->dbpass    = '${DB_PASSWORD:-}';
\$CFG->prefix    = 'mdl_';
\$CFG->dboptions = array(
    'dbpersist' => 0,
    'dbport' => ${DB_PORT:-3306},
    'dbsocket' => '',
    'dbcollation' => 'utf8mb4_unicode_ci',
);

// Directory configuration
\$CFG->wwwroot   = 'https://${MOODLE_DOMAIN:-lms.example.com}';
\$CFG->dataroot  = '$MOODLE_DATA_DIR';
\$CFG->admin     = 'admin';

// Security configuration
\$CFG->directorypermissions = 0777;
\$CFG->filepermissions = 0644;
\$CFG->dirpermissions = 0755;

// Performance configuration
\$CFG->preventexecpath = true;
\$CFG->pathtophp = '/usr/bin/php';
\$CFG->pathtodu = '/usr/bin/du';
\$CFG->aspellpath = '/usr/bin/aspell';
\$CFG->pathtodot = '/usr/bin/dot';

// Session configuration
\$CFG->session_handler_class = '\\core\\session\\redis';
\$CFG->session_redis_host = '127.0.0.1';
\$CFG->session_redis_port = 6379;
\$CFG->session_redis_database = 0;
\$CFG->session_redis_prefix = 'moodle_session_';

// Cache configuration
\$CFG->cachejs = true;
\$CFG->cachetemplates = true;

// Logging configuration
\$CFG->loglifetime = 0;
\$CFG->logguests = true;

// Email configuration
\$CFG->smtphosts = 'localhost';
\$CFG->smtpuser = '';
\$CFG->smtppass = '';
\$CFG->smtpsecure = '';
\$CFG->smtpauthtype = '';

// This is for PHP settings which can be used in case you cannot edit php.ini
\$CFG->phpunit_dataroot = '$MOODLE_DATA_DIR/phpunit';
\$CFG->phpunit_prefix = 'phpu_';

require_once(__DIR__ . '/lib/setup.php');
EOF

    # Set secure permissions for config.php
    chmod 644 "$MOODLE_DIR/config.php"
    chown www-data:www-data "$MOODLE_DIR/config.php"
    
    log "INFO" "Moodle configuration file created successfully"
}

run_moodle_installation() {
    log "INFO" "Running Moodle installation..."
    
    cd "$MOODLE_DIR"
    
    # Run installation via command line
    if sudo -u www-data php admin/cli/install.php \
        --lang="${LANGUAGE:-en}" \
        --wwwroot="https://${MOODLE_DOMAIN:-lms.example.com}" \
        --dataroot="$MOODLE_DATA_DIR" \
        --dbtype="${DB_TYPE:-mariadb}" \
        --dbhost="${DB_HOST:-localhost}" \
        --dbname="${DB_NAME:-moodle}" \
        --dbuser="${DB_USER:-moodle}" \
        --dbpass="${DB_PASSWORD:-}" \
        --dbport="${DB_PORT:-3306}" \
        --prefix="mdl_" \
        --fullname="${SITE_FULLNAME:-LMS K2NET}" \
        --shortname="${SITE_SHORTNAME:-LMSK2}" \
        --adminuser="${ADMIN_USERNAME:-admin}" \
        --adminpass="${ADMIN_PASSWORD:-}" \
        --adminemail="${ADMIN_EMAIL:-admin@example.com}" \
        --agree-license \
        --non-interactive; then
        log "INFO" "Moodle installation completed successfully"
    else
        log "ERROR" "Moodle installation failed"
        exit 1
    fi
}

configure_redis_session() {
    log "INFO" "Configuring Redis session..."
    
    # Install Redis if not already installed
    if ! command -v redis-server &> /dev/null; then
        log "INFO" "Installing Redis..."
        apt update
        apt install -y redis-server
    fi
    
    # Configure Redis
    cat > /etc/redis/redis.conf << 'EOF'
bind 127.0.0.1
port 6379
timeout 0
tcp-keepalive 300

maxmemory 512mb
maxmemory-policy allkeys-lru

save 900 1
save 300 10
save 60 10000
EOF

    # Start and enable Redis
    systemctl start redis-server
    systemctl enable redis-server
    
    log "INFO" "Redis session configuration completed"
}

configure_opcache() {
    log "INFO" "Configuring OPcache..."
    
    # Create OPcache configuration
    cat > /etc/php/8.1/fpm/conf.d/99-opcache.ini << 'EOF'
; OPcache Configuration for Moodle 4.0
opcache.enable = 1
opcache.memory_consumption = 256
opcache.interned_strings_buffer = 8
opcache.max_accelerated_files = 10000
opcache.revalidate_freq = 2
opcache.fast_shutdown = 1
opcache.validate_timestamps = 0
opcache.save_comments = 1
opcache.enable_file_override = 1
EOF

    # Restart PHP-FPM
    systemctl restart php8.1-fpm
    
    log "INFO" "OPcache configuration completed"
}

configure_moodle_cron() {
    log "INFO" "Configuring Moodle cron job..."
    
    # Add Moodle cron job
    (crontab -l 2>/dev/null; cat << EOF
# Moodle cron job (every 5 minutes)
*/5 * * * * /usr/bin/php $MOODLE_DIR/admin/cli/cron.php >/dev/null
EOF
    ) | crontab -
    
    log "INFO" "Moodle cron job configured successfully"
}

run_post_installation() {
    log "INFO" "Running post-installation configuration..."
    
    cd "$MOODLE_DIR"
    
    # Set proper permissions
    log "INFO" "Setting final file permissions..."
    chown -R www-data:www-data "$MOODLE_DIR"
    chmod -R 755 "$MOODLE_DIR"
    chmod -R 777 "$MOODLE_DATA_DIR"
    chmod 644 "$MOODLE_DIR/config.php"
    chown www-data:www-data "$MOODLE_DIR/config.php"
    
    # Run upgrade
    log "INFO" "Running Moodle upgrade..."
    sudo -u www-data php admin/cli/upgrade.php --non-interactive
    
    # Clear caches
    log "INFO" "Clearing Moodle caches..."
    sudo -u www-data php admin/cli/purge_caches.php
    
    log "INFO" "Post-installation configuration completed"
}

restart_services() {
    log "INFO" "Restarting services..."
    
    # Restart PHP-FPM
    systemctl restart php8.1-fpm
    
    # Restart Nginx
    systemctl restart nginx
    
    # Restart Redis
    systemctl restart redis-server
    
    log "INFO" "Services restarted successfully"
}

# =============================================================================
# Verification Functions
# =============================================================================

verify_installation() {
    log "INFO" "Verifying Moodle installation..."
    
    # Check Moodle installation
    log "INFO" "Checking Moodle web interface..."
    if curl -I "https://${MOODLE_DOMAIN:-lms.example.com}" >/dev/null 2>&1; then
        log "INFO" "Moodle web interface is accessible"
    else
        log "WARN" "Moodle web interface may not be accessible"
    fi
    
    # Check database connection
    log "INFO" "Checking database connection..."
    if mysql -u "${DB_USER:-moodle}" -p"${DB_PASSWORD:-}" "${DB_NAME:-moodle}" -e "SELECT COUNT(*) FROM mdl_user;" >/dev/null 2>&1; then
        log "INFO" "Database connection is working"
    else
        log "WARN" "Database connection may have issues"
    fi
    
    # Check Redis connection
    log "INFO" "Checking Redis connection..."
    if redis-cli ping >/dev/null 2>&1; then
        log "INFO" "Redis connection is working"
    else
        log "WARN" "Redis connection may have issues"
    fi
    
    # Check file permissions
    log "INFO" "Checking file permissions..."
    ls -la "$MOODLE_DIR/config.php"
    ls -la "$MOODLE_DATA_DIR/" | head -5
    
    # Check cron job
    log "INFO" "Checking cron job..."
    if crontab -l | grep -q "moodle"; then
        log "INFO" "Moodle cron job is configured"
    else
        log "WARN" "Moodle cron job not found"
    fi
    
    # Check Moodle status
    log "INFO" "Checking Moodle status..."
    cd "$MOODLE_DIR"
    sudo -u www-data php admin/cli/upgrade.php --non-interactive >/dev/null 2>&1 || log "WARN" "Moodle upgrade check failed"
    
    log "INFO" "Moodle installation verification completed"
}

# =============================================================================
# Cleanup Functions
# =============================================================================

cleanup_installation() {
    log "INFO" "Cleaning up installation files..."
    
    # Remove temporary files
    rm -rf /tmp/moodle-install
    
    log "INFO" "Installation cleanup completed"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    log "INFO" "Starting Moodle $MOODLE_VERSION installation process..."
    
    # Check prerequisites
    check_root
    create_directories
    backup_existing
    
    # Execute installation steps
    download_moodle
    install_moodle_files
    create_moodle_config
    run_moodle_installation
    configure_redis_session
    configure_opcache
    configure_moodle_cron
    run_post_installation
    restart_services
    
    # Verify installation
    verify_installation
    
    # Cleanup
    cleanup_installation
    
    log "INFO" "Moodle $MOODLE_VERSION installation process completed successfully!"
    log "INFO" "Log file: $LOG_FILE"
    log "INFO" "Backup directory: $BACKUP_DIR"
    log "INFO" "Moodle URL: https://${MOODLE_DOMAIN:-lms.example.com}"
    log "INFO" "Admin username: ${ADMIN_USERNAME:-admin}"
    log "INFO" "Admin email: ${ADMIN_EMAIL:-admin@example.com}"
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
        echo "  --verify       Only verify current installation"
        echo "  --dry-run      Show what would be done without executing"
        echo "  --config       Show current configuration"
        exit 0
        ;;
    --verify)
        check_root
        verify_installation
        exit 0
        ;;
    --dry-run)
        log "INFO" "DRY RUN MODE - No changes will be made"
        log "INFO" "Would execute: download_moodle, install_moodle_files, create_moodle_config, run_moodle_installation, configure_redis_session, configure_opcache, configure_moodle_cron, run_post_installation, restart_services"
        exit 0
        ;;
    --config)
        echo "Current Configuration:"
        echo "  Moodle Version: $MOODLE_VERSION"
        echo "  Moodle Domain: ${MOODLE_DOMAIN:-lms.example.com}"
        echo "  Database Type: ${DB_TYPE:-mariadb}"
        echo "  Database Host: ${DB_HOST:-localhost}"
        echo "  Database Name: ${DB_NAME:-moodle}"
        echo "  Database User: ${DB_USER:-moodle}"
        echo "  Admin Username: ${ADMIN_USERNAME:-admin}"
        echo "  Admin Email: ${ADMIN_EMAIL:-admin@example.com}"
        echo "  Timezone: ${TIMEZONE:-Asia/Jakarta}"
        echo "  Language: ${LANGUAGE:-en}"
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
