#!/bin/bash

# =============================================================================
# Backup and Restore Script
# =============================================================================
# Version: 1.0
# Author: jejakawan007
# Description: Backup and restore functionality for LMSK2-Moodle-Server
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Script information
SCRIPT_NAME="Backup and Restore"
SCRIPT_VERSION="1.0"
SCRIPT_AUTHOR="jejakawan007"

# Default configuration
BACKUP_DIR="/backup/moodle"
LOG_DIR="/var/log/lmsk2"
MOODLE_DIR="/var/www/moodle"
MOODLE_DATA_DIR="/var/www/moodle/moodledata"
DB_NAME="moodle"
DB_USER="moodle"
DB_HOST="localhost"
DB_PORT="3306"

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
# Global Variables
# =============================================================================

BACKUP_TYPE=""
RESTORE_TYPE=""
BACKUP_NAME=""
COMPRESS=false
ENCRYPT=false
VERBOSE=false
DRY_RUN=false

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
    mkdir -p "$LOG_DIR"
    
    # Write to log file
    echo "[$timestamp] [$level] $message" >> "$LOG_DIR/backup-restore.log"
    
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

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "Script ini harus dijalankan sebagai root atau dengan sudo"
        exit 1
    fi
}

# Create backup directory
create_backup_directory() {
    if [[ ! -d "$BACKUP_DIR" ]]; then
        mkdir -p "$BACKUP_DIR"
        chmod 700 "$BACKUP_DIR"
        print_success "Created backup directory: $BACKUP_DIR"
    fi
}

# Generate backup name
generate_backup_name() {
    local type=$1
    local timestamp=$(date +%Y%m%d_%H%M%S)
    echo "${type}_${timestamp}"
}

# =============================================================================
# Backup Functions
# =============================================================================

# Database backup
backup_database() {
    local backup_name=$1
    local backup_path="$BACKUP_DIR/$backup_name"
    
    print_section "Database Backup"
    
    # Get database password
    local db_password="${LMSK2_DB_PASSWORD:-}"
    if [[ -z "$db_password" ]]; then
        print_error "Database password not provided"
        return 1
    fi
    
    # Create backup directory
    mkdir -p "$backup_path"
    
    # Database backup
    print_info "Backing up database: $DB_NAME"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would backup database to $backup_path/database.sql"
        return 0
    fi
    
    # Create database backup
    mysqldump -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$db_password" \
        --single-transaction \
        --routines \
        --triggers \
        --events \
        --hex-blob \
        --opt \
        "$DB_NAME" > "$backup_path/database.sql"
    
    # Compress if requested
    if [[ "$COMPRESS" == "true" ]]; then
        print_info "Compressing database backup..."
        gzip "$backup_path/database.sql"
        print_success "Database backup compressed: $backup_path/database.sql.gz"
    else
        print_success "Database backup created: $backup_path/database.sql"
    fi
    
    log_message "SUCCESS" "Database backup completed: $backup_name"
}

# Files backup
backup_files() {
    local backup_name=$1
    local backup_path="$BACKUP_DIR/$backup_name"
    
    print_section "Files Backup"
    
    # Create backup directory
    mkdir -p "$backup_path"
    
    # Moodle files backup
    print_info "Backing up Moodle files: $MOODLE_DIR"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would backup files to $backup_path/moodle_files.tar.gz"
        return 0
    fi
    
    # Create files backup
    tar -czf "$backup_path/moodle_files.tar.gz" -C "$(dirname "$MOODLE_DIR")" "$(basename "$MOODLE_DIR")"
    
    print_success "Moodle files backup created: $backup_path/moodle_files.tar.gz"
    
    # Moodle data backup
    if [[ -d "$MOODLE_DATA_DIR" ]]; then
        print_info "Backing up Moodle data: $MOODLE_DATA_DIR"
        tar -czf "$backup_path/moodle_data.tar.gz" -C "$(dirname "$MOODLE_DATA_DIR")" "$(basename "$MOODLE_DATA_DIR")"
        print_success "Moodle data backup created: $backup_path/moodle_data.tar.gz"
    else
        print_warning "Moodle data directory not found: $MOODLE_DATA_DIR"
    fi
    
    log_message "SUCCESS" "Files backup completed: $backup_name"
}

# Configuration backup
backup_configuration() {
    local backup_name=$1
    local backup_path="$BACKUP_DIR/$backup_name"
    
    print_section "Configuration Backup"
    
    # Create backup directory
    mkdir -p "$backup_path"
    
    # Configuration files to backup
    local config_files=(
        "/etc/nginx/sites-available/moodle"
        "/etc/php/8.1/fpm/php.ini"
        "/etc/php/8.1/fpm/conf.d/99-moodle.ini"
        "/etc/mysql/mariadb.conf.d/99-moodle-optimization.cnf"
        "/etc/redis/redis.conf"
        "/etc/sysctl.d/99-moodle-optimization.conf"
        "/etc/security/limits.d/99-moodle.conf"
        "/var/www/moodle/config.php"
    )
    
    print_info "Backing up configuration files..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would backup configuration files to $backup_path/config.tar.gz"
        return 0
    fi
    
    # Create temporary directory for config files
    local temp_dir="/tmp/config_backup_$$"
    mkdir -p "$temp_dir"
    
    # Copy configuration files
    for config_file in "${config_files[@]}"; do
        if [[ -f "$config_file" ]]; then
            local dir_name=$(dirname "$config_file")
            local file_name=$(basename "$config_file")
            mkdir -p "$temp_dir$dir_name"
            cp "$config_file" "$temp_dir$dir_name/$file_name"
            print_info "Backed up: $config_file"
        else
            print_warning "Configuration file not found: $config_file"
        fi
    done
    
    # Create configuration backup archive
    tar -czf "$backup_path/config.tar.gz" -C "$temp_dir" .
    rm -rf "$temp_dir"
    
    print_success "Configuration backup created: $backup_path/config.tar.gz"
    
    log_message "SUCCESS" "Configuration backup completed: $backup_name"
}

# Full backup
backup_full() {
    local backup_name=$(generate_backup_name "full")
    
    print_section "Full Backup: $backup_name"
    
    print_info "Starting full backup..."
    
    # Create backup directory
    mkdir -p "$BACKUP_DIR/$backup_name"
    
    # Backup database
    backup_database "$backup_name"
    
    # Backup files
    backup_files "$backup_name"
    
    # Backup configuration
    backup_configuration "$backup_name"
    
    # Create backup manifest
    create_backup_manifest "$backup_name"
    
    print_success "Full backup completed: $backup_name"
    log_message "SUCCESS" "Full backup completed: $backup_name"
}

# Incremental backup
backup_incremental() {
    local backup_name=$(generate_backup_name "incremental")
    
    print_section "Incremental Backup: $backup_name"
    
    print_info "Starting incremental backup..."
    
    # Create backup directory
    mkdir -p "$BACKUP_DIR/$backup_name"
    
    # Find last full backup
    local last_full_backup=$(find "$BACKUP_DIR" -name "full_*" -type d | sort | tail -1)
    
    if [[ -z "$last_full_backup" ]]; then
        print_warning "No full backup found, creating full backup instead"
        backup_full
        return 0
    fi
    
    print_info "Last full backup: $(basename "$last_full_backup")"
    
    # Backup only changed files since last full backup
    backup_files_incremental "$backup_name" "$last_full_backup"
    
    # Create backup manifest
    create_backup_manifest "$backup_name"
    
    print_success "Incremental backup completed: $backup_name"
    log_message "SUCCESS" "Incremental backup completed: $backup_name"
}

# Incremental files backup
backup_files_incremental() {
    local backup_name=$1
    local last_backup=$2
    local backup_path="$BACKUP_DIR/$backup_name"
    
    print_info "Creating incremental files backup..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would create incremental backup"
        return 0
    fi
    
    # Create incremental backup using rsync
    rsync -av --link-dest="$last_backup" "$MOODLE_DIR/" "$backup_path/moodle_files/"
    
    print_success "Incremental files backup created"
}

# Create backup manifest
create_backup_manifest() {
    local backup_name=$1
    local backup_path="$BACKUP_DIR/$backup_name"
    local manifest_file="$backup_path/manifest.txt"
    
    print_info "Creating backup manifest..."
    
    cat > "$manifest_file" << EOF
LMSK2 Moodle Server Backup Manifest
====================================
Backup Name: $backup_name
Backup Type: $BACKUP_TYPE
Created: $(date)
Created By: $(whoami)
Hostname: $(hostname)
OS: $(lsb_release -d | cut -f2)
Kernel: $(uname -r)

Backup Contents:
EOF
    
    # List backup contents
    if [[ -d "$backup_path" ]]; then
        find "$backup_path" -type f -exec ls -la {} \; >> "$manifest_file"
    fi
    
    print_success "Backup manifest created: $manifest_file"
}

# =============================================================================
# Restore Functions
# =============================================================================

# Restore database
restore_database() {
    local backup_path=$1
    
    print_section "Database Restore"
    
    # Get database password
    local db_password="${LMSK2_DB_PASSWORD:-}"
    if [[ -z "$db_password" ]]; then
        print_error "Database password not provided"
        return 1
    fi
    
    # Find database backup file
    local db_file=""
    if [[ -f "$backup_path/database.sql.gz" ]]; then
        db_file="$backup_path/database.sql.gz"
        print_info "Found compressed database backup: $db_file"
    elif [[ -f "$backup_path/database.sql" ]]; then
        db_file="$backup_path/database.sql"
        print_info "Found database backup: $db_file"
    else
        print_error "Database backup file not found in $backup_path"
        return 1
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would restore database from $db_file"
        return 0
    fi
    
    # Drop and recreate database
    print_info "Dropping and recreating database: $DB_NAME"
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$db_password" -e "DROP DATABASE IF EXISTS $DB_NAME;"
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$db_password" -e "CREATE DATABASE $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
    
    # Restore database
    print_info "Restoring database from: $db_file"
    if [[ "$db_file" == *.gz ]]; then
        gunzip -c "$db_file" | mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$db_password" "$DB_NAME"
    else
        mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$db_password" "$DB_NAME" < "$db_file"
    fi
    
    print_success "Database restored successfully"
    log_message "SUCCESS" "Database restore completed from $backup_path"
}

# Restore files
restore_files() {
    local backup_path=$1
    
    print_section "Files Restore"
    
    # Find files backup
    local files_backup=""
    if [[ -f "$backup_path/moodle_files.tar.gz" ]]; then
        files_backup="$backup_path/moodle_files.tar.gz"
        print_info "Found Moodle files backup: $files_backup"
    else
        print_error "Moodle files backup not found in $backup_path"
        return 1
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would restore files from $files_backup"
        return 0
    fi
    
    # Backup current files
    if [[ -d "$MOODLE_DIR" ]]; then
        print_info "Backing up current Moodle files..."
        local current_backup="/tmp/moodle_current_backup_$(date +%Y%m%d_%H%M%S).tar.gz"
        tar -czf "$current_backup" -C "$(dirname "$MOODLE_DIR")" "$(basename "$MOODLE_DIR")"
        print_success "Current files backed up to: $current_backup"
    fi
    
    # Restore files
    print_info "Restoring Moodle files from: $files_backup"
    tar -xzf "$files_backup" -C "$(dirname "$MOODLE_DIR")"
    
    # Set proper permissions
    chown -R www-data:www-data "$MOODLE_DIR"
    chmod -R 755 "$MOODLE_DIR"
    
    print_success "Moodle files restored successfully"
    
    # Restore Moodle data if available
    if [[ -f "$backup_path/moodle_data.tar.gz" ]]; then
        print_info "Restoring Moodle data..."
        tar -xzf "$backup_path/moodle_data.tar.gz" -C "$(dirname "$MOODLE_DATA_DIR")"
        chown -R www-data:www-data "$MOODLE_DATA_DIR"
        chmod -R 777 "$MOODLE_DATA_DIR"
        print_success "Moodle data restored successfully"
    fi
    
    log_message "SUCCESS" "Files restore completed from $backup_path"
}

# Restore configuration
restore_configuration() {
    local backup_path=$1
    
    print_section "Configuration Restore"
    
    # Find configuration backup
    local config_backup=""
    if [[ -f "$backup_path/config.tar.gz" ]]; then
        config_backup="$backup_path/config.tar.gz"
        print_info "Found configuration backup: $config_backup"
    else
        print_error "Configuration backup not found in $backup_path"
        return 1
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would restore configuration from $config_backup"
        return 0
    fi
    
    # Create temporary directory for config files
    local temp_dir="/tmp/config_restore_$$"
    mkdir -p "$temp_dir"
    
    # Extract configuration files
    print_info "Extracting configuration files..."
    tar -xzf "$config_backup" -C "$temp_dir"
    
    # Restore configuration files
    print_info "Restoring configuration files..."
    find "$temp_dir" -type f | while read -r config_file; do
        local relative_path="${config_file#$temp_dir}"
        local target_file="$relative_path"
        
        if [[ -f "$target_file" ]]; then
            # Backup existing file
            cp "$target_file" "$target_file.backup.$(date +%Y%m%d_%H%M%S)"
        fi
        
        # Create directory if it doesn't exist
        mkdir -p "$(dirname "$target_file")"
        
        # Copy configuration file
        cp "$config_file" "$target_file"
        print_info "Restored: $target_file"
    done
    
    # Clean up temporary directory
    rm -rf "$temp_dir"
    
    print_success "Configuration restored successfully"
    log_message "SUCCESS" "Configuration restore completed from $backup_path"
}

# Full restore
restore_full() {
    local backup_name=$1
    local backup_path="$BACKUP_DIR/$backup_name"
    
    print_section "Full Restore: $backup_name"
    
    if [[ ! -d "$backup_path" ]]; then
        print_error "Backup not found: $backup_path"
        return 1
    fi
    
    print_info "Starting full restore from: $backup_path"
    
    # Restore database
    restore_database "$backup_path"
    
    # Restore files
    restore_files "$backup_path"
    
    # Restore configuration
    restore_configuration "$backup_path"
    
    print_success "Full restore completed: $backup_name"
    log_message "SUCCESS" "Full restore completed: $backup_name"
}

# =============================================================================
# Utility Functions
# =============================================================================

# List backups
list_backups() {
    print_section "Available Backups"
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        print_warning "Backup directory not found: $BACKUP_DIR"
        return 1
    fi
    
    print_info "Backups in $BACKUP_DIR:"
    echo
    
    # List all backup directories
    find "$BACKUP_DIR" -maxdepth 1 -type d -name "*_*" | sort -r | while read -r backup_dir; do
        local backup_name=$(basename "$backup_dir")
        local backup_date=$(stat -c %y "$backup_dir" | cut -d' ' -f1)
        local backup_time=$(stat -c %y "$backup_dir" | cut -d' ' -f2 | cut -d'.' -f1)
        local backup_size=$(du -sh "$backup_dir" | cut -f1)
        
        echo "  $backup_name"
        echo "    Date: $backup_date $backup_time"
        echo "    Size: $backup_size"
        echo
    done
}

# Cleanup old backups
cleanup_backups() {
    local retention_days=${1:-30}
    
    print_section "Backup Cleanup"
    
    print_info "Cleaning up backups older than $retention_days days..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would delete backups older than $retention_days days"
        find "$BACKUP_DIR" -maxdepth 1 -type d -name "*_*" -mtime +$retention_days -exec echo "Would delete: {}" \;
        return 0
    fi
    
    # Find and delete old backups
    local deleted_count=0
    find "$BACKUP_DIR" -maxdepth 1 -type d -name "*_*" -mtime +$retention_days | while read -r backup_dir; do
        print_info "Deleting old backup: $(basename "$backup_dir")"
        rm -rf "$backup_dir"
        ((deleted_count++))
    done
    
    print_success "Backup cleanup completed"
    log_message "SUCCESS" "Backup cleanup completed: $retention_days days retention"
}

# Show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] COMMAND

LMSK2 Moodle Server Backup and Restore Script v$SCRIPT_VERSION

COMMANDS:
    backup [TYPE]     Create backup (full, incremental)
    restore NAME      Restore from backup
    list              List available backups
    cleanup [DAYS]    Cleanup old backups (default: 30 days)

OPTIONS:
    --backup-dir=DIR  Backup directory (default: $BACKUP_DIR)
    --compress        Compress backup files
    --encrypt         Encrypt backup files
    --verbose         Verbose output
    --dry-run         Test mode (no changes)
    --help            Show this help message

EXAMPLES:
    $0 backup full
    $0 backup incremental --compress
    $0 restore full_20240101_120000
    $0 list
    $0 cleanup 7
    $0 backup full --dry-run

EOF
}

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --backup-dir=*)
                BACKUP_DIR="${1#*=}"
                shift
                ;;
            --compress)
                COMPRESS=true
                shift
                ;;
            --encrypt)
                ENCRYPT=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --help)
                show_usage
                exit 0
                ;;
            backup)
                BACKUP_TYPE="${2:-full}"
                shift 2
                ;;
            restore)
                RESTORE_TYPE="full"
                BACKUP_NAME="$2"
                shift 2
                ;;
            list)
                list_backups
                exit 0
                ;;
            cleanup)
                cleanup_backups "${2:-30}"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    # Check if running as root
    check_root
    
    # Create backup directory
    create_backup_directory
    
    # Parse arguments
    parse_arguments "$@"
    
    # Show header
    print_color $CYAN "=============================================================================="
    print_color $WHITE "  $SCRIPT_NAME v$SCRIPT_VERSION"
    print_color $CYAN "=============================================================================="
    
    # Execute based on command
    if [[ -n "$BACKUP_TYPE" ]]; then
        case $BACKUP_TYPE in
            "full")
                backup_full
                ;;
            "incremental")
                backup_incremental
                ;;
            *)
                print_error "Invalid backup type: $BACKUP_TYPE"
                show_usage
                exit 1
                ;;
        esac
    elif [[ -n "$RESTORE_TYPE" ]]; then
        if [[ -z "$BACKUP_NAME" ]]; then
            print_error "Backup name required for restore"
            show_usage
            exit 1
        fi
        restore_full "$BACKUP_NAME"
    else
        print_error "No command specified"
        show_usage
        exit 1
    fi
    
    print_section "Operation Complete"
    print_success "Backup/restore operation completed successfully!"
    print_info "Log file: $LOG_DIR/backup-restore.log"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_warning "This was a dry run - no changes were made"
    fi
}

# Trap errors
trap 'print_error "Script failed at line $LINENO"; log_message "ERROR" "Script failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
