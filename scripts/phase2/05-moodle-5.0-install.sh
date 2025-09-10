#!/bin/bash

# =============================================================================
# Phase 2: Moodle 5.0 Installation
# =============================================================================
# Version: 1.0
# Author: jejakawan007
# Description: Moodle 5.0 installation untuk LMSK2-Moodle-Server
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Script information
SCRIPT_NAME="Phase 2: Moodle 5.0 Installation"
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

# Moodle 5.0 specific settings
MOODLE_VERSION="5.0"
MOODLE_DOWNLOAD_URL="https://download.moodle.org/releases/latest500/moodle-latest-500.tgz"
MOODLE_REQUIRED_PHP_VERSION="8.1"
MOODLE_REQUIRED_MYSQL_VERSION="8.0"

# Paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MOODLE_DIR="/var/www/moodle"
MOODLE_DATA_DIR="/var/www/moodle/moodledata"
BACKUP_DIR="/backup/moodle"
TEMP_DIR="/tmp/moodle-install"

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
    print_color $CYAN "=============================================================================="
    print_color $WHITE "  $section"
    print_color $CYAN "=============================================================================="
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
    echo "[$timestamp] [$level] $message" >> "/var/log/lmsk2/phase2-moodle-5.0.log"
    
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
# Pre-installation Checks
# =============================================================================

# Check system requirements
check_requirements() {
    print_section "Checking System Requirements"
    
    # Check PHP version
    local php_version=$(php -r "echo PHP_VERSION;" | cut -d. -f1,2)
    local required_php_version=$MOODLE_REQUIRED_PHP_VERSION
    
    print_info "Checking PHP version: $php_version (required: $required_php_version+)"
    if [[ $(echo "$php_version >= $required_php_version" | bc -l) -eq 1 ]]; then
        print_success "PHP version check passed"
    else
        print_error "PHP version $php_version is not supported. Required: $required_php_version+"
        exit 1
    fi
    
    # Check MySQL version
    local mysql_version=$(mysql --version | awk '{print $3}' | cut -d. -f1,2)
    local required_mysql_version=$MOODLE_REQUIRED_MYSQL_VERSION
    
    print_info "Checking MySQL version: $mysql_version (required: $required_mysql_version+)"
    if [[ $(echo "$mysql_version >= $required_mysql_version" | bc -l) -eq 1 ]]; then
        print_success "MySQL version check passed"
    else
        print_error "MySQL version $mysql_version is not supported. Required: $required_mysql_version+"
        exit 1
    fi
    
    # Check required PHP extensions
    local required_extensions=("mysqli" "curl" "xml" "zip" "gd" "intl" "mbstring" "soap" "ldap" "imagick" "redis" "opcache")
    print_info "Checking PHP extensions..."
    
    for extension in "${required_extensions[@]}"; do
        if php -m | grep -q "$extension"; then
            print_success "PHP extension $extension is installed"
        else
            print_error "PHP extension $extension is missing"
            exit 1
        fi
    done
    
    # Check disk space
    local available_space=$(df / | awk 'NR==2 {print $4}' | sed 's/G//')
    if [[ $available_space -lt 3 ]]; then
        print_error "Insufficient disk space. Available: ${available_space}GB, Required: 3GB+"
        exit 1
    else
        print_success "Disk space check passed: ${available_space}GB available"
    fi
    
    log_message "SUCCESS" "System requirements check completed"
}

# =============================================================================
# Moodle 5.0 Installation
# =============================================================================

# Download Moodle 5.0
download_moodle() {
    print_section "Downloading Moodle 5.0"
    
    # Create temporary directory
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    print_info "Downloading Moodle 5.0 from: $MOODLE_DOWNLOAD_URL"
    
    if wget -q --show-progress "$MOODLE_DOWNLOAD_URL" -O "moodle-5.0.tgz"; then
        print_success "Moodle 5.0 downloaded successfully"
    else
        print_error "Failed to download Moodle 5.0"
        exit 1
    fi
    
    # Verify download
    if [[ -f "moodle-5.0.tgz" ]]; then
        local file_size=$(du -h "moodle-5.0.tgz" | cut -f1)
        print_success "Download verification passed: $file_size"
    else
        print_error "Download verification failed"
        exit 1
    fi
    
    log_message "SUCCESS" "Moodle 5.0 download completed"
}

# Extract Moodle 5.0
extract_moodle() {
    print_section "Extracting Moodle 5.0"
    
    cd "$TEMP_DIR"
    
    print_info "Extracting Moodle 5.0..."
    
    if tar -xzf "moodle-5.0.tgz"; then
        print_success "Moodle 5.0 extracted successfully"
    else
        print_error "Failed to extract Moodle 5.0"
        exit 1
    fi
    
    # Check extracted files
    if [[ -d "moodle" ]]; then
        local extracted_files=$(find moodle -type f | wc -l)
        print_success "Extraction verification passed: $extracted_files files extracted"
    else
        print_error "Extraction verification failed"
        exit 1
    fi
    
    log_message "SUCCESS" "Moodle 5.0 extraction completed"
}

# Install Moodle 5.0
install_moodle() {
    print_section "Installing Moodle 5.0"
    
    # Backup existing installation if exists
    if [[ -d "$MOODLE_DIR" ]]; then
        print_info "Backing up existing Moodle installation..."
        local backup_name="moodle-backup-$(date +%Y%m%d_%H%M%S)"
        mv "$MOODLE_DIR" "$BACKUP_DIR/$backup_name"
        print_success "Existing installation backed up to: $BACKUP_DIR/$backup_name"
    fi
    
    # Create Moodle directory
    mkdir -p "$MOODLE_DIR"
    
    # Copy Moodle files
    print_info "Copying Moodle 5.0 files..."
    cp -r "$TEMP_DIR/moodle/"* "$MOODLE_DIR/"
    
    # Set proper permissions
    print_info "Setting file permissions..."
    chown -R www-data:www-data "$MOODLE_DIR"
    chmod -R 755 "$MOODLE_DIR"
    
    # Create moodledata directory
    mkdir -p "$MOODLE_DATA_DIR"
    chown -R www-data:www-data "$MOODLE_DATA_DIR"
    chmod -R 777 "$MOODLE_DATA_DIR"
    
    print_success "Moodle 5.0 installed successfully"
    log_message "SUCCESS" "Moodle 5.0 installation completed"
}

# =============================================================================
# Moodle 5.0 Configuration
# =============================================================================

# Create Moodle configuration
create_moodle_config() {
    print_section "Creating Moodle 5.0 Configuration"
    
    # Get configuration values
    local db_host="${LMSK2_DB_HOST:-localhost}"
    local db_name="${LMSK2_DB_NAME:-moodle}"
    local db_user="${LMSK2_DB_USER:-moodle}"
    local db_password="${LMSK2_DB_PASSWORD:-}"
    local db_prefix="${LMSK2_DB_PREFIX:-mdl_}"
    local wwwroot="${LMSK2_DOMAIN:-http://localhost/moodle}"
    local dataroot="$MOODLE_DATA_DIR"
    local admin_user="${LMSK2_ADMIN_USERNAME:-admin}"
    local admin_password="${LMSK2_ADMIN_PASSWORD:-}"
    local admin_email="${LMSK2_ADMIN_EMAIL:-admin@yourdomain.com}"
    local site_name="${LMSK2_SITE_NAME:-LMSK2 Learning Platform}"
    local site_shortname="${LMSK2_SITE_SHORTNAME:-LMSK2}"
    
    # Create config.php
    print_info "Creating config.php..."
    
    cat > "$MOODLE_DIR/config.php" << EOF
<?php  // Moodle configuration file

unset(\$CFG);
global \$CFG;
\$CFG = new stdClass();

// Database settings
\$CFG->dbtype    = 'mariadb';
\$CFG->dblibrary = 'native';
\$CFG->dbhost    = '$db_host';
\$CFG->dbname    = '$db_name';
\$CFG->dbuser    = '$db_user';
\$CFG->dbpass    = '$db_password';
\$CFG->prefix    = '$db_prefix';
\$CFG->dboptions = array(
    'dbpersist' => 0,
    'dbport' => 3306,
    'dbsocket' => '',
    'dbcollation' => 'utf8mb4_unicode_ci',
);

// Site settings
\$CFG->wwwroot   = '$wwwroot';
\$CFG->dataroot  = '$dataroot';
\$CFG->admin     = 'admin';

// Directory permissions
\$CFG->directorypermissions = 0777;

// Performance settings
\$CFG->preventexecpath = true;
\$CFG->pathtophp = '/usr/bin/php';
\$CFG->pathtodu = '/usr/bin/du';
\$CFG->aspellpath = '/usr/bin/aspell';
\$CFG->pathtodot = '/usr/bin/dot';

// Security settings
\$CFG->slasharguments = true;
\$CFG->enablewebservices = true;
\$CFG->enablemobilewebservice = true;
\$CFG->maintenance_enabled = false;
\$CFG->maintenance_message = 'LMSK2 is currently under maintenance. Please try again later.';

// Cache settings
\$CFG->cachejs = true;
\$CFG->cachecss = true;
\$CFG->cachetemplates = true;

// Session settings
\$CFG->session_handler_class = '\core\session\redis';
\$CFG->session_redis_host = '127.0.0.1';
\$CFG->session_redis_port = 6379;
\$CFG->session_redis_database = 0;
\$CFG->session_redis_prefix = 'moodle_session_';

// File settings
\$CFG->maxbytes = 268435456;  // 256MB
\$CFG->maxareabytes = 1073741824;  // 1GB

// Logging settings
\$CFG->loglifetime = 30;
\$CFG->logguests = true;

// Theme settings
\$CFG->theme = 'boost';

// Language settings
\$CFG->lang = 'en';
\$CFG->country = 'ID';
\$CFG->timezone = 'Asia/Jakarta';

// Additional settings for Moodle 5.0
\$CFG->enablecompletion = true;
\$CFG->enablebadges = true;
\$CFG->enableportfolios = true;
\$CFG->enableblogs = true;
\$CFG->enablemessaging = true;
\$CFG->enableanalytics = true;
\$CFG->enableh5p = true;
\$CFG->enablecompetencies = true;

// AI and Machine Learning settings
\$CFG->enableai = true;
\$CFG->ai_provider = 'openai';
\$CFG->ai_api_key = '';
\$CFG->ai_model = 'gpt-3.5-turbo';

// Advanced features
\$CFG->enablemobileapp = true;
\$CFG->enablewebservices = true;
\$CFG->enableapi = true;

// End of configuration
require_once(__DIR__ . '/lib/setup.php');
EOF

    # Set proper permissions for config.php
    chown www-data:www-data "$MOODLE_DIR/config.php"
    chmod 600 "$MOODLE_DIR/config.php"
    
    print_success "Moodle 5.0 configuration created"
    log_message "SUCCESS" "Moodle 5.0 configuration completed"
}

# =============================================================================
# Database Setup
# =============================================================================

# Setup database
setup_database() {
    print_section "Setting Up Database for Moodle 5.0"
    
    local db_host="${LMSK2_DB_HOST:-localhost}"
    local db_name="${LMSK2_DB_NAME:-moodle}"
    local db_user="${LMSK2_DB_USER:-moodle}"
    local db_password="${LMSK2_DB_PASSWORD:-}"
    local db_prefix="${LMSK2_DB_PREFIX:-mdl_}"
    
    # Create database if not exists
    print_info "Creating database: $db_name"
    mysql -u root -p"${LMSK2_DB_PASSWORD:-}" -e "CREATE DATABASE IF NOT EXISTS $db_name CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
    
    # Create user if not exists
    print_info "Creating database user: $db_user"
    mysql -u root -p"${LMSK2_DB_PASSWORD:-}" -e "CREATE USER IF NOT EXISTS '$db_user'@'$db_host' IDENTIFIED BY '$db_password';"
    mysql -u root -p"${LMSK2_DB_PASSWORD:-}" -e "GRANT ALL PRIVILEGES ON $db_name.* TO '$db_user'@'$db_host';"
    mysql -u root -p"${LMSK2_DB_PASSWORD:-}" -e "FLUSH PRIVILEGES;"
    
    print_success "Database setup completed"
    log_message "SUCCESS" "Database setup completed"
}

# =============================================================================
# Moodle 5.0 Installation via CLI
# =============================================================================

# Install Moodle via CLI
install_moodle_cli() {
    print_section "Installing Moodle 5.0 via CLI"
    
    local admin_user="${LMSK2_ADMIN_USERNAME:-admin}"
    local admin_password="${LMSK2_ADMIN_PASSWORD:-}"
    local admin_email="${LMSK2_ADMIN_EMAIL:-admin@yourdomain.com}"
    local site_name="${LMSK2_SITE_NAME:-LMSK2 Learning Platform}"
    local site_shortname="${LMSK2_SITE_SHORTNAME:-LMSK2}"
    
    cd "$MOODLE_DIR"
    
    print_info "Installing Moodle 5.0 via CLI..."
    
    # Run Moodle installation
    if php admin/cli/install.php \
        --lang=en \
        --wwwroot="${LMSK2_DOMAIN:-http://localhost/moodle}" \
        --dataroot="$MOODLE_DATA_DIR" \
        --dbtype=mariadb \
        --dbhost="${LMSK2_DB_HOST:-localhost}" \
        --dbname="${LMSK2_DB_NAME:-moodle}" \
        --dbuser="${LMSK2_DB_USER:-moodle}" \
        --dbpass="${LMSK2_DB_PASSWORD:-}" \
        --dbprefix="${LMSK2_DB_PREFIX:-mdl_}" \
        --fullname="$site_name" \
        --shortname="$site_shortname" \
        --adminuser="$admin_user" \
        --adminpass="$admin_password" \
        --adminemail="$admin_email" \
        --agree-license \
        --non-interactive; then
        print_success "Moodle 5.0 CLI installation completed"
    else
        print_error "Moodle 5.0 CLI installation failed"
        exit 1
    fi
    
    log_message "SUCCESS" "Moodle 5.0 CLI installation completed"
}

# =============================================================================
# Post-installation Configuration
# =============================================================================

# Configure Moodle 5.0 settings
configure_moodle_settings() {
    print_section "Configuring Moodle 5.0 Settings"
    
    cd "$MOODLE_DIR"
    
    # Configure Moodle settings
    print_info "Configuring Moodle 5.0 settings..."
    
    # Enable completion tracking
    php admin/cli/cfg.php --name=enablecompletion --set=1
    
    # Enable badges
    php admin/cli/cfg.php --name=enablebadges --set=1
    
    # Enable portfolios
    php admin/cli/cfg.php --name=enableportfolios --set=1
    
    # Enable blogs
    php admin/cli/cfg.php --name=enableblogs --set=1
    
    # Enable messaging
    php admin/cli/cfg.php --name=enablemessaging --set=1
    
    # Enable analytics
    php admin/cli/cfg.php --name=enableanalytics --set=1
    
    # Enable H5P
    php admin/cli/cfg.php --name=enableh5p --set=1
    
    # Enable competencies
    php admin/cli/cfg.php --name=enablecompetencies --set=1
    
    # Enable AI features
    php admin/cli/cfg.php --name=enableai --set=1
    
    # Configure theme
    php admin/cli/cfg.php --name=theme --set=boost
    
    # Configure language
    php admin/cli/cfg.php --name=lang --set=en
    
    # Configure timezone
    php admin/cli/cfg.php --name=timezone --set=Asia/Jakarta
    
    # Configure country
    php admin/cli/cfg.php --name=country --set=ID
    
    # Configure file upload limits
    php admin/cli/cfg.php --name=maxbytes --set=268435456  # 256MB
    php admin/cli/cfg.php --name=maxareabytes --set=1073741824  # 1GB
    
    # Configure session settings
    php admin/cli/cfg.php --name=session_handler_class --set="\core\session\redis"
    php admin/cli/cfg.php --name=session_redis_host --set=127.0.0.1
    php admin/cli/cfg.php --name=session_redis_port --set=6379
    php admin/cli/cfg.php --name=session_redis_database --set=0
    php admin/cli/cfg.php --name=session_redis_prefix --set="moodle_session_"
    
    # Configure cache settings
    php admin/cli/cfg.php --name=cachejs --set=1
    php admin/cli/cfg.php --name=cachecss --set=1
    php admin/cli/cfg.php --name=cachetemplates --set=1
    
    # Configure mobile app settings
    php admin/cli/cfg.php --name=enablemobileapp --set=1
    
    # Configure API settings
    php admin/cli/cfg.php --name=enableapi --set=1
    
    print_success "Moodle 5.0 settings configured"
    log_message "SUCCESS" "Moodle 5.0 settings configuration completed"
}

# =============================================================================
# Cleanup
# =============================================================================

# Cleanup installation files
cleanup() {
    print_section "Cleaning Up Installation Files"
    
    # Remove temporary files
    if [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
        print_success "Temporary files cleaned up"
    fi
    
    # Clear caches
    cd "$MOODLE_DIR"
    php admin/cli/purge_caches.php
    
    print_success "Cleanup completed"
    log_message "SUCCESS" "Cleanup completed"
}

# =============================================================================
# Verification
# =============================================================================

# Verify installation
verify_installation() {
    print_section "Verifying Moodle 5.0 Installation"
    
    # Check if Moodle is accessible
    print_info "Checking Moodle accessibility..."
    if curl -s -o /dev/null -w "%{http_code}" "${LMSK2_DOMAIN:-http://localhost/moodle}" | grep -q "200"; then
        print_success "Moodle 5.0 is accessible"
    else
        print_warning "Moodle 5.0 accessibility check failed"
    fi
    
    # Check database connection
    print_info "Checking database connection..."
    cd "$MOODLE_DIR"
    if php admin/cli/cfg.php --name=dbname --get | grep -q "moodle"; then
        print_success "Database connection verified"
    else
        print_warning "Database connection check failed"
    fi
    
    # Check file permissions
    print_info "Checking file permissions..."
    if [[ -r "$MOODLE_DIR/config.php" ]] && [[ -w "$MOODLE_DATA_DIR" ]]; then
        print_success "File permissions verified"
    else
        print_warning "File permissions check failed"
    fi
    
    # Display installation summary
    print_info "Installation Summary:"
    echo "  Moodle Version: $MOODLE_VERSION"
    echo "  Installation Directory: $MOODLE_DIR"
    echo "  Data Directory: $MOODLE_DATA_DIR"
    echo "  Database: ${LMSK2_DB_NAME:-moodle}"
    echo "  Admin User: ${LMSK2_ADMIN_USERNAME:-admin}"
    echo "  Site URL: ${LMSK2_DOMAIN:-http://localhost/moodle}"
    
    print_success "Moodle 5.0 installation verification completed"
    log_message "SUCCESS" "Moodle 5.0 installation verification completed"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    print_color $CYAN "=============================================================================="
    print_color $WHITE "  $SCRIPT_NAME v$SCRIPT_VERSION"
    print_color $CYAN "=============================================================================="
    
    log_message "INFO" "Starting Moodle 5.0 installation"
    
    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root"
        exit 1
    fi
    
    # Execute installation steps
    check_requirements
    download_moodle
    extract_moodle
    install_moodle
    create_moodle_config
    setup_database
    install_moodle_cli
    configure_moodle_settings
    cleanup
    verify_installation
    
    print_section "Moodle 5.0 Installation Complete"
    print_success "Moodle 5.0 installation completed successfully!"
    print_info "Log file: /var/log/lmsk2/phase2-moodle-5.0.log"
    print_info "Access your Moodle at: ${LMSK2_DOMAIN:-http://localhost/moodle}"
    print_info "Admin credentials: ${LMSK2_ADMIN_USERNAME:-admin} / ${LMSK2_ADMIN_PASSWORD:-[your password]}"
    
    log_message "SUCCESS" "Moodle 5.0 installation completed successfully"
}

# Trap errors
trap 'print_error "Script failed at line $LINENO"; log_message "ERROR" "Script failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
