#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: Phase 3 - Backup Strategy Script
# =============================================================================
# Description: Comprehensive backup strategy for LMSK2-Moodle-Server
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
SCRIPT_NAME="LMSK2-Moodle-Server Backup Strategy"
SCRIPT_VERSION="1.0"
LOG_FILE="/var/log/lmsk2-backup-strategy.log"
CONFIG_DIR="/opt/lmsk2-moodle-server/scripts/config"
BACKUP_DIR="/opt/lmsk2-moodle-server/scripts/backup"

# Load configuration
if [ -f "$CONFIG_DIR/backup.conf" ]; then
    source "$CONFIG_DIR/backup.conf"
else
    echo -e "${YELLOW}Warning: Backup configuration file not found. Using defaults.${NC}"
fi

# Default configuration
BACKUP_ENABLE=${BACKUP_ENABLE:-"true"}
BACKUP_RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-"30"}
BACKUP_COMPRESSION=${BACKUP_COMPRESSION:-"gzip"}
BACKUP_ENCRYPTION=${BACKUP_ENCRYPTION:-"false"}
BACKUP_EMAIL=${BACKUP_EMAIL:-"admin@localhost"}
MOODLE_DATA_DIR=${MOODLE_DATA_DIR:-"/var/www/moodle"}
MOODLE_DB_NAME=${MOODLE_DB_NAME:-"moodle"}

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
        log "ERROR" "Backup strategy setup failed. Exit code: $exit_code"
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
# Backup Strategy Setup
# =============================================================================

# Create backup directory structure
create_backup_structure() {
    log "INFO" "Creating backup directory structure..."
    
    mkdir -p "$BACKUP_DIR"/{scripts,config,templates,logs}
    mkdir -p /var/backups/lmsk2/{database,files,config,logs}
    mkdir -p /var/backups/lmsk2/retention/{daily,weekly,monthly}
    
    # Set proper permissions
    chown -R www-data:www-data "$BACKUP_DIR"
    chmod -R 755 "$BACKUP_DIR"
    chmod -R 755 /var/backups/lmsk2
    
    log "INFO" "Backup directory structure created"
}

# Create backup configuration
create_backup_config() {
    log "INFO" "Creating backup configuration..."
    
    cat > "$CONFIG_DIR/backup.conf" << EOF
# LMSK2 Backup Configuration

# Backup settings
BACKUP_ENABLE=true
BACKUP_RETENTION_DAYS=30
BACKUP_COMPRESSION=gzip
BACKUP_ENCRYPTION=false
BACKUP_EMAIL=admin@localhost

# Database settings
MOODLE_DB_NAME=moodle
MOODLE_DB_USER=moodle
MOODLE_DB_PASSWORD=your_password_here
MOODLE_DB_HOST=localhost

# File paths
MOODLE_DATA_DIR=/var/www/moodle
MOODLE_CONFIG_DIR=/var/www/moodle/config.php
BACKUP_BASE_DIR=/var/backups/lmsk2

# Backup schedules
FULL_BACKUP_SCHEDULE="0 2 * * 0"  # Weekly on Sunday at 2 AM
INCREMENTAL_BACKUP_SCHEDULE="0 2 * * 1-6"  # Daily at 2 AM
CONFIG_BACKUP_SCHEDULE="0 3 * * *"  # Daily at 3 AM

# Retention policies
DAILY_RETENTION=7
WEEKLY_RETENTION=4
MONTHLY_RETENTION=12

# Notification settings
SEND_BACKUP_NOTIFICATIONS=true
BACKUP_SUCCESS_EMAIL=true
BACKUP_FAILURE_EMAIL=true

# Compression settings
COMPRESSION_LEVEL=6
COMPRESSION_EXTENSION=.gz

# Encryption settings (if enabled)
ENCRYPTION_KEY_FILE=/etc/lmsk2-backup/backup.key
ENCRYPTION_ALGORITHM=aes-256-cbc
EOF

    log "INFO" "Backup configuration created"
}

# Create database backup script
create_database_backup_script() {
    log "INFO" "Creating database backup script..."
    
    cat > "$BACKUP_DIR/scripts/database-backup.sh" << 'EOF'
#!/bin/bash

# Database backup script
CONFIG_FILE="/opt/lmsk2-moodle-server/scripts/config/backup.conf"
LOG_FILE="/var/log/lmsk2-backup/database-backup.log"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Function to log with timestamp
log_backup() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Create backup filename
create_backup_filename() {
    local backup_type="$1"
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    echo "${MOODLE_DB_NAME}_${backup_type}_${timestamp}.sql"
}

# Perform database backup
backup_database() {
    local backup_type="$1"
    local backup_file="$2"
    local backup_path="$BACKUP_BASE_DIR/database/$backup_file"
    
    log_backup "Starting $backup_type database backup..."
    
    # Create database dump
    mysqldump -h "$MOODLE_DB_HOST" \
              -u "$MOODLE_DB_USER" \
              -p"$MOODLE_DB_PASSWORD" \
              --single-transaction \
              --routines \
              --triggers \
              --events \
              --hex-blob \
              --opt \
              "$MOODLE_DB_NAME" > "$backup_path" 2>> "$LOG_FILE"
    
    if [ $? -eq 0 ]; then
        log_backup "Database backup completed successfully: $backup_path"
        
        # Compress backup if enabled
        if [ "$BACKUP_COMPRESSION" = "gzip" ]; then
            gzip "$backup_path"
            backup_path="${backup_path}.gz"
            log_backup "Backup compressed: $backup_path"
        fi
        
        # Encrypt backup if enabled
        if [ "$BACKUP_ENCRYPTION" = "true" ] && [ -f "$ENCRYPTION_KEY_FILE" ]; then
            openssl enc -"$ENCRYPTION_ALGORITHM" \
                        -in "$backup_path" \
                        -out "${backup_path}.enc" \
                        -pass file:"$ENCRYPTION_KEY_FILE"
            rm "$backup_path"
            backup_path="${backup_path}.enc"
            log_backup "Backup encrypted: $backup_path"
        fi
        
        # Set proper permissions
        chmod 600 "$backup_path"
        chown root:root "$backup_path"
        
        echo "$backup_path"
    else
        log_backup "ERROR: Database backup failed"
        return 1
    fi
}

# Verify backup integrity
verify_backup() {
    local backup_file="$1"
    
    log_backup "Verifying backup integrity: $backup_file"
    
    # Check if file exists and has content
    if [ ! -f "$backup_file" ] || [ ! -s "$backup_file" ]; then
        log_backup "ERROR: Backup file is missing or empty"
        return 1
    fi
    
    # Test backup restoration (dry run)
    if [[ "$backup_file" == *.gz ]]; then
        gunzip -t "$backup_file" 2>> "$LOG_FILE"
    elif [[ "$backup_file" == *.enc ]]; then
        # For encrypted files, just check if they exist and have content
        if [ -s "$backup_file" ]; then
            log_backup "Encrypted backup file verified"
        else
            log_backup "ERROR: Encrypted backup file is empty"
            return 1
        fi
    else
        # For plain SQL files, check for SQL syntax
        head -n 10 "$backup_file" | grep -q "CREATE\|INSERT\|UPDATE" 2>> "$LOG_FILE"
    fi
    
    if [ $? -eq 0 ]; then
        log_backup "Backup integrity verified successfully"
        return 0
    else
        log_backup "ERROR: Backup integrity verification failed"
        return 1
    fi
}

# Send backup notification
send_notification() {
    local backup_type="$1"
    local status="$2"
    local backup_file="$3"
    local message="$4"
    
    if [ "$SEND_BACKUP_NOTIFICATIONS" = "true" ]; then
        local subject="LMSK2 Backup $status: $backup_type"
        local body="Backup Type: $backup_type
Status: $status
File: $backup_file
Message: $message
Timestamp: $(date)"

        echo "$body" | mail -s "$subject" "$BACKUP_EMAIL"
        log_backup "Notification sent: $subject"
    fi
}

# Main backup function
main() {
    local backup_type="${1:-full}"
    local backup_file=$(create_backup_filename "$backup_type")
    
    log_backup "Starting database backup process"
    
    # Perform backup
    local backup_path=$(backup_database "$backup_type" "$backup_file")
    
    if [ $? -eq 0 ]; then
        # Verify backup
        if verify_backup "$backup_path"; then
            send_notification "$backup_type" "SUCCESS" "$backup_path" "Database backup completed successfully"
            log_backup "Database backup process completed successfully"
        else
            send_notification "$backup_type" "FAILED" "$backup_path" "Backup verification failed"
            log_backup "Database backup process failed: verification error"
            exit 1
        fi
    else
        send_notification "$backup_type" "FAILED" "$backup_file" "Database backup failed"
        log_backup "Database backup process failed"
        exit 1
    fi
}

# Run main function
main "$@"
EOF

    chmod +x "$BACKUP_DIR/scripts/database-backup.sh"
    log "INFO" "Database backup script created"
}

# Create file backup script
create_file_backup_script() {
    log "INFO" "Creating file backup script..."
    
    cat > "$BACKUP_DIR/scripts/file-backup.sh" << 'EOF'
#!/bin/bash

# File backup script
CONFIG_FILE="/opt/lmsk2-moodle-server/scripts/config/backup.conf"
LOG_FILE="/var/log/lmsk2-backup/file-backup.log"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Function to log with timestamp
log_backup() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Create backup filename
create_backup_filename() {
    local backup_type="$1"
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    echo "moodle_files_${backup_type}_${timestamp}.tar"
}

# Perform file backup
backup_files() {
    local backup_type="$1"
    local backup_file="$2"
    local backup_path="$BACKUP_BASE_DIR/files/$backup_file"
    
    log_backup "Starting $backup_type file backup..."
    
    # Create tar archive
    tar -cf "$backup_path" \
        -C "$(dirname "$MOODLE_DATA_DIR")" \
        "$(basename "$MOODLE_DATA_DIR")" \
        --exclude="*/cache/*" \
        --exclude="*/temp/*" \
        --exclude="*/sessions/*" \
        --exclude="*/trashdir/*" \
        2>> "$LOG_FILE"
    
    if [ $? -eq 0 ]; then
        log_backup "File backup completed successfully: $backup_path"
        
        # Compress backup if enabled
        if [ "$BACKUP_COMPRESSION" = "gzip" ]; then
            gzip "$backup_path"
            backup_path="${backup_path}.gz"
            log_backup "Backup compressed: $backup_path"
        fi
        
        # Encrypt backup if enabled
        if [ "$BACKUP_ENCRYPTION" = "true" ] && [ -f "$ENCRYPTION_KEY_FILE" ]; then
            openssl enc -"$ENCRYPTION_ALGORITHM" \
                        -in "$backup_path" \
                        -out "${backup_path}.enc" \
                        -pass file:"$ENCRYPTION_KEY_FILE"
            rm "$backup_path"
            backup_path="${backup_path}.enc"
            log_backup "Backup encrypted: $backup_path"
        fi
        
        # Set proper permissions
        chmod 600 "$backup_path"
        chown root:root "$backup_path"
        
        echo "$backup_path"
    else
        log_backup "ERROR: File backup failed"
        return 1
    fi
}

# Verify backup integrity
verify_backup() {
    local backup_file="$1"
    
    log_backup "Verifying backup integrity: $backup_file"
    
    # Check if file exists and has content
    if [ ! -f "$backup_file" ] || [ ! -s "$backup_file" ]; then
        log_backup "ERROR: Backup file is missing or empty"
        return 1
    fi
    
    # Test backup extraction (dry run)
    if [[ "$backup_file" == *.gz ]]; then
        gunzip -t "$backup_file" 2>> "$LOG_FILE"
    elif [[ "$backup_file" == *.enc ]]; then
        # For encrypted files, just check if they exist and have content
        if [ -s "$backup_file" ]; then
            log_backup "Encrypted backup file verified"
        else
            log_backup "ERROR: Encrypted backup file is empty"
            return 1
        fi
    else
        # For plain tar files, test extraction
        tar -tf "$backup_file" > /dev/null 2>> "$LOG_FILE"
    fi
    
    if [ $? -eq 0 ]; then
        log_backup "Backup integrity verified successfully"
        return 0
    else
        log_backup "ERROR: Backup integrity verification failed"
        return 1
    fi
}

# Send backup notification
send_notification() {
    local backup_type="$1"
    local status="$2"
    local backup_file="$3"
    local message="$4"
    
    if [ "$SEND_BACKUP_NOTIFICATIONS" = "true" ]; then
        local subject="LMSK2 Backup $status: $backup_type"
        local body="Backup Type: $backup_type
Status: $status
File: $backup_file
Message: $message
Timestamp: $(date)"

        echo "$body" | mail -s "$subject" "$BACKUP_EMAIL"
        log_backup "Notification sent: $subject"
    fi
}

# Main backup function
main() {
    local backup_type="${1:-full}"
    local backup_file=$(create_backup_filename "$backup_type")
    
    log_backup "Starting file backup process"
    
    # Perform backup
    local backup_path=$(backup_files "$backup_type" "$backup_file")
    
    if [ $? -eq 0 ]; then
        # Verify backup
        if verify_backup "$backup_path"; then
            send_notification "$backup_type" "SUCCESS" "$backup_path" "File backup completed successfully"
            log_backup "File backup process completed successfully"
        else
            send_notification "$backup_type" "FAILED" "$backup_path" "Backup verification failed"
            log_backup "File backup process failed: verification error"
            exit 1
        fi
    else
        send_notification "$backup_type" "FAILED" "$backup_file" "File backup failed"
        log_backup "File backup process failed"
        exit 1
    fi
}

# Run main function
main "$@"
EOF

    chmod +x "$BACKUP_DIR/scripts/file-backup.sh"
    log "INFO" "File backup script created"
}

# Create configuration backup script
create_config_backup_script() {
    log "INFO" "Creating configuration backup script..."
    
    cat > "$BACKUP_DIR/scripts/config-backup.sh" << 'EOF'
#!/bin/bash

# Configuration backup script
CONFIG_FILE="/opt/lmsk2-moodle-server/scripts/config/backup.conf"
LOG_FILE="/var/log/lmsk2-backup/config-backup.log"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Function to log with timestamp
log_backup() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Create backup filename
create_backup_filename() {
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    echo "moodle_config_${timestamp}.tar"
}

# Perform configuration backup
backup_config() {
    local backup_file="$1"
    local backup_path="$BACKUP_BASE_DIR/config/$backup_file"
    
    log_backup "Starting configuration backup..."
    
    # Create tar archive with configuration files
    tar -cf "$backup_path" \
        -C /etc nginx \
        -C /etc php \
        -C /etc mysql \
        -C /etc redis \
        -C /etc cron.d . \
        -C /opt/lmsk2-moodle-server/scripts config \
        "$MOODLE_CONFIG_DIR" \
        2>> "$LOG_FILE"
    
    if [ $? -eq 0 ]; then
        log_backup "Configuration backup completed successfully: $backup_path"
        
        # Compress backup if enabled
        if [ "$BACKUP_COMPRESSION" = "gzip" ]; then
            gzip "$backup_path"
            backup_path="${backup_path}.gz"
            log_backup "Backup compressed: $backup_path"
        fi
        
        # Encrypt backup if enabled
        if [ "$BACKUP_ENCRYPTION" = "true" ] && [ -f "$ENCRYPTION_KEY_FILE" ]; then
            openssl enc -"$ENCRYPTION_ALGORITHM" \
                        -in "$backup_path" \
                        -out "${backup_path}.enc" \
                        -pass file:"$ENCRYPTION_KEY_FILE"
            rm "$backup_path"
            backup_path="${backup_path}.enc"
            log_backup "Backup encrypted: $backup_path"
        fi
        
        # Set proper permissions
        chmod 600 "$backup_path"
        chown root:root "$backup_path"
        
        echo "$backup_path"
    else
        log_backup "ERROR: Configuration backup failed"
        return 1
    fi
}

# Send backup notification
send_notification() {
    local status="$1"
    local backup_file="$2"
    local message="$3"
    
    if [ "$SEND_BACKUP_NOTIFICATIONS" = "true" ]; then
        local subject="LMSK2 Config Backup $status"
        local body="Backup Type: Configuration
Status: $status
File: $backup_file
Message: $message
Timestamp: $(date)"

        echo "$body" | mail -s "$subject" "$BACKUP_EMAIL"
        log_backup "Notification sent: $subject"
    fi
}

# Main backup function
main() {
    local backup_file=$(create_backup_filename)
    
    log_backup "Starting configuration backup process"
    
    # Perform backup
    local backup_path=$(backup_config "$backup_file")
    
    if [ $? -eq 0 ]; then
        send_notification "SUCCESS" "$backup_path" "Configuration backup completed successfully"
        log_backup "Configuration backup process completed successfully"
    else
        send_notification "FAILED" "$backup_file" "Configuration backup failed"
        log_backup "Configuration backup process failed"
        exit 1
    fi
}

# Run main function
main "$@"
EOF

    chmod +x "$BACKUP_DIR/scripts/config-backup.sh"
    log "INFO" "Configuration backup script created"
}

# Create backup cleanup script
create_backup_cleanup_script() {
    log "INFO" "Creating backup cleanup script..."
    
    cat > "$BACKUP_DIR/scripts/backup-cleanup.sh" << 'EOF'
#!/bin/bash

# Backup cleanup script
CONFIG_FILE="/opt/lmsk2-moodle-server/scripts/config/backup.conf"
LOG_FILE="/var/log/lmsk2-backup/backup-cleanup.log"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Function to log with timestamp
log_cleanup() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Cleanup old backups
cleanup_old_backups() {
    local backup_type="$1"
    local retention_days="$2"
    local backup_dir="$BACKUP_BASE_DIR/$backup_type"
    
    log_cleanup "Cleaning up $backup_type backups older than $retention_days days..."
    
    if [ -d "$backup_dir" ]; then
        # Find and remove old backup files
        find "$backup_dir" -name "*.sql*" -o -name "*.tar*" | \
        while read -r backup_file; do
            if [ -f "$backup_file" ]; then
                local file_age=$(($(date +%s) - $(stat -c %Y "$backup_file")))
                local age_days=$((file_age / 86400))
                
                if [ "$age_days" -gt "$retention_days" ]; then
                    log_cleanup "Removing old backup: $backup_file (age: $age_days days)"
                    rm -f "$backup_file"
                fi
            fi
        done
        
        log_cleanup "Cleanup completed for $backup_type backups"
    else
        log_cleanup "Backup directory not found: $backup_dir"
    fi
}

# Main cleanup function
main() {
    log_cleanup "Starting backup cleanup process"
    
    # Cleanup different backup types
    cleanup_old_backups "database" "$DAILY_RETENTION"
    cleanup_old_backups "files" "$DAILY_RETENTION"
    cleanup_old_backups "config" "$DAILY_RETENTION"
    
    log_cleanup "Backup cleanup process completed"
}

# Run main function
main "$@"
EOF

    chmod +x "$BACKUP_DIR/scripts/backup-cleanup.sh"
    log "INFO" "Backup cleanup script created"
}

# Setup cron jobs for automated backups
setup_backup_cron_jobs() {
    log "INFO" "Setting up backup cron jobs..."
    
    # Create cron job entries
    local cron_entries=(
        "# LMSK2 Backup Jobs"
        "$FULL_BACKUP_SCHEDULE $BACKUP_DIR/scripts/database-backup.sh full"
        "$INCREMENTAL_BACKUP_SCHEDULE $BACKUP_DIR/scripts/database-backup.sh incremental"
        "$CONFIG_BACKUP_SCHEDULE $BACKUP_DIR/scripts/config-backup.sh"
        "0 4 * * * $BACKUP_DIR/scripts/file-backup.sh full"
        "0 5 * * * $BACKUP_DIR/scripts/backup-cleanup.sh"
    )
    
    # Add cron jobs
    for entry in "${cron_entries[@]}"; do
        (crontab -l 2>/dev/null; echo "$entry") | crontab -
    done
    
    log "INFO" "Backup cron jobs configured"
}

# Create disaster recovery script
create_disaster_recovery_script() {
    log "INFO" "Creating disaster recovery script..."
    
    cat > "$BACKUP_DIR/scripts/disaster-recovery.sh" << 'EOF'
#!/bin/bash

# Disaster recovery script
CONFIG_FILE="/opt/lmsk2-moodle-server/scripts/config/backup.conf"
LOG_FILE="/var/log/lmsk2-backup/disaster-recovery.log"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Function to log with timestamp
log_recovery() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Restore database
restore_database() {
    local backup_file="$1"
    
    log_recovery "Starting database restoration from: $backup_file"
    
    # Stop services
    systemctl stop nginx php8.1-fpm
    
    # Decrypt backup if needed
    local temp_file="$backup_file"
    if [[ "$backup_file" == *.enc ]]; then
        temp_file="${backup_file%.enc}"
        openssl enc -"$ENCRYPTION_ALGORITHM" \
                    -d \
                    -in "$backup_file" \
                    -out "$temp_file" \
                    -pass file:"$ENCRYPTION_KEY_FILE"
    fi
    
    # Decompress backup if needed
    if [[ "$temp_file" == *.gz ]]; then
        gunzip "$temp_file"
        temp_file="${temp_file%.gz}"
    fi
    
    # Restore database
    mysql -h "$MOODLE_DB_HOST" \
          -u "$MOODLE_DB_USER" \
          -p"$MOODLE_DB_PASSWORD" \
          "$MOODLE_DB_NAME" < "$temp_file"
    
    if [ $? -eq 0 ]; then
        log_recovery "Database restoration completed successfully"
        
        # Cleanup temporary files
        if [ "$temp_file" != "$backup_file" ]; then
            rm -f "$temp_file"
        fi
        
        # Start services
        systemctl start php8.1-fpm nginx
        
        return 0
    else
        log_recovery "ERROR: Database restoration failed"
        return 1
    fi
}

# Restore files
restore_files() {
    local backup_file="$1"
    
    log_recovery "Starting file restoration from: $backup_file"
    
    # Stop services
    systemctl stop nginx php8.1-fpm
    
    # Decrypt backup if needed
    local temp_file="$backup_file"
    if [[ "$backup_file" == *.enc ]]; then
        temp_file="${backup_file%.enc}"
        openssl enc -"$ENCRYPTION_ALGORITHM" \
                    -d \
                    -in "$backup_file" \
                    -out "$temp_file" \
                    -pass file:"$ENCRYPTION_KEY_FILE"
    fi
    
    # Decompress backup if needed
    if [[ "$temp_file" == *.gz ]]; then
        gunzip "$temp_file"
        temp_file="${temp_file%.gz}"
    fi
    
    # Restore files
    tar -xf "$temp_file" -C "$(dirname "$MOODLE_DATA_DIR")"
    
    if [ $? -eq 0 ]; then
        log_recovery "File restoration completed successfully"
        
        # Set proper permissions
        chown -R www-data:www-data "$MOODLE_DATA_DIR"
        chmod -R 755 "$MOODLE_DATA_DIR"
        
        # Cleanup temporary files
        if [ "$temp_file" != "$backup_file" ]; then
            rm -f "$temp_file"
        fi
        
        # Start services
        systemctl start php8.1-fpm nginx
        
        return 0
    else
        log_recovery "ERROR: File restoration failed"
        return 1
    fi
}

# Main recovery function
main() {
    local recovery_type="$1"
    local backup_file="$2"
    
    if [ -z "$recovery_type" ] || [ -z "$backup_file" ]; then
        echo "Usage: $0 {database|files} <backup_file>"
        exit 1
    fi
    
    log_recovery "Starting disaster recovery process"
    log_recovery "Recovery type: $recovery_type"
    log_recovery "Backup file: $backup_file"
    
    case "$recovery_type" in
        "database")
            restore_database "$backup_file"
            ;;
        "files")
            restore_files "$backup_file"
            ;;
        *)
            echo "Invalid recovery type: $recovery_type"
            echo "Valid types: database, files"
            exit 1
            ;;
    esac
    
    if [ $? -eq 0 ]; then
        log_recovery "Disaster recovery process completed successfully"
    else
        log_recovery "Disaster recovery process failed"
        exit 1
    fi
}

# Run main function
main "$@"
EOF

    chmod +x "$BACKUP_DIR/scripts/disaster-recovery.sh"
    log "INFO" "Disaster recovery script created"
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
    
    log "INFO" "Starting backup strategy setup..."
    
    # Check prerequisites
    check_root
    
    # Setup backup strategy
    create_backup_structure
    create_backup_config
    create_database_backup_script
    create_file_backup_script
    create_config_backup_script
    create_backup_cleanup_script
    create_disaster_recovery_script
    setup_backup_cron_jobs
    
    # Final verification
    log "INFO" "Verifying backup strategy setup..."
    
    # Check if backup scripts are executable
    local backup_scripts=(
        "$BACKUP_DIR/scripts/database-backup.sh"
        "$BACKUP_DIR/scripts/file-backup.sh"
        "$BACKUP_DIR/scripts/config-backup.sh"
        "$BACKUP_DIR/scripts/backup-cleanup.sh"
        "$BACKUP_DIR/scripts/disaster-recovery.sh"
    )
    
    for script in "${backup_scripts[@]}"; do
        if [ -x "$script" ]; then
            log "INFO" "✓ $script is executable"
        else
            log "ERROR" "✗ $script is not executable"
        fi
    done
    
    # Check cron jobs
    if crontab -l | grep -q "lmsk2-backup"; then
        log "INFO" "✓ Backup cron jobs are configured"
    else
        log "WARN" "✗ Backup cron jobs not found"
    fi
    
    # Display summary
    echo
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Backup Strategy Setup Completed!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo
    echo -e "${WHITE}Backup Components Installed:${NC}"
    echo -e "  • Database backup (MySQL/MariaDB)"
    echo -e "  • File backup (Moodle data directory)"
    echo -e "  • Configuration backup (System configs)"
    echo -e "  • Automated cleanup (Retention management)"
    echo -e "  • Disaster recovery (Restore procedures)"
    echo -e "  • Automated scheduling (Cron jobs)"
    echo
    echo -e "${WHITE}Configuration Files:${NC}"
    echo -e "  • $CONFIG_DIR/backup.conf"
    echo -e "  • $BACKUP_DIR/scripts/"
    echo
    echo -e "${WHITE}Backup Locations:${NC}"
    echo -e "  • /var/backups/lmsk2/database/"
    echo -e "  • /var/backups/lmsk2/files/"
    echo -e "  • /var/backups/lmsk2/config/"
    echo
    echo -e "${WHITE}Next Steps:${NC}"
    echo -e "  1. Configure database credentials in backup.conf"
    echo -e "  2. Test backup scripts manually"
    echo -e "  3. Verify cron jobs are running"
    echo -e "  4. Test disaster recovery procedures"
    echo -e "  5. Monitor backup logs: /var/log/lmsk2-backup/"
    echo
    
    log "INFO" "Backup strategy setup completed successfully"
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
            echo "  --email EMAIL       Set backup notification email"
            echo "  --retention DAYS    Set backup retention days"
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
        --email)
            BACKUP_EMAIL="$2"
            shift 2
            ;;
        --retention)
            BACKUP_RETENTION_DAYS="$2"
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

