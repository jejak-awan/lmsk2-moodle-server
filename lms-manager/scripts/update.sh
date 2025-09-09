#!/bin/bash

# LMS Manager Update Script
# Author: jejakawan007
# Company: K2NET
# Website: https://k2net.id

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/opt/lms-manager"
SERVICE_NAME="lms-manager"
SERVICE_USER="lms-manager"
SERVICE_GROUP="lms-manager"
BACKUP_DIR="/opt/lms-manager-backups"
UPDATE_URL="https://github.com/jejakawan007/lmsk2-moodle-server/releases/latest"

# Functions
print_header() {
    echo -e "${BLUE}"
    echo "=========================================="
    echo "    LMS Manager - Update Script"
    echo "=========================================="
    echo -e "${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root"
        exit 1
    fi
}

check_installation() {
    if [[ ! -d "$INSTALL_DIR" ]]; then
        print_error "LMS Manager is not installed"
        exit 1
    fi

    if [[ ! -f "$INSTALL_DIR/lms-manager" ]]; then
        print_error "LMS Manager binary not found"
        exit 1
    fi

    print_success "LMS Manager installation found"
}

get_current_version() {
    if [[ -f "$INSTALL_DIR/lms-manager" ]]; then
        CURRENT_VERSION=$($INSTALL_DIR/lms-manager --version 2>/dev/null || echo "unknown")
    else
        CURRENT_VERSION="unknown"
    fi
    print_info "Current version: $CURRENT_VERSION"
}

check_latest_version() {
    print_info "Checking for latest version..."

    # Get latest release info
    LATEST_VERSION=$(curl -s https://api.github.com/repos/jejakawan007/lmsk2-moodle-server/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [[ -z "$LATEST_VERSION" ]]; then
        print_error "Failed to get latest version information"
        exit 1
    fi

    print_info "Latest version: $LATEST_VERSION"

    if [[ "$CURRENT_VERSION" == "$LATEST_VERSION" ]]; then
        print_success "You are already running the latest version"
        exit 0
    fi
}

create_backup() {
    print_info "Creating backup..."

    # Create backup directory
    mkdir -p "$BACKUP_DIR"
    
    # Create timestamped backup
    BACKUP_NAME="lms-manager-backup-$(date +%Y%m%d-%H%M%S)"
    BACKUP_PATH="$BACKUP_DIR/$BACKUP_NAME"
    
    mkdir -p "$BACKUP_PATH"
    
    # Backup application files
    cp -r "$INSTALL_DIR"/* "$BACKUP_PATH/"
    
    # Backup configuration
    if [[ -f "$INSTALL_DIR/config/config.json" ]]; then
        cp "$INSTALL_DIR/config/config.json" "$BACKUP_PATH/config.json"
    fi
    
    # Backup database
    if [[ -f "$INSTALL_DIR/lms-manager.db" ]]; then
        cp "$INSTALL_DIR/lms-manager.db" "$BACKUP_PATH/lms-manager.db"
    fi
    
    print_success "Backup created: $BACKUP_PATH"
}

stop_service() {
    print_info "Stopping LMS Manager service..."

    if systemctl is-active --quiet "$SERVICE_NAME"; then
        systemctl stop "$SERVICE_NAME"
        print_success "Service stopped"
    else
        print_info "Service is not running"
    fi
}

download_update() {
    print_info "Downloading update..."

    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"

    # Download latest release
    wget -q "https://github.com/jejakawan007/lmsk2-moodle-server/archive/refs/tags/$LATEST_VERSION.tar.gz"
    
    # Extract
    tar -xzf "$LATEST_VERSION.tar.gz"
    
    # Find the extracted directory
    EXTRACTED_DIR=$(find . -name "lmsk2-moodle-server-*" -type d | head -1)
    
    if [[ -z "$EXTRACTED_DIR" ]]; then
        print_error "Failed to extract update files"
        exit 1
    fi
    
    # Copy to installation directory
    cp -r "$EXTRACTED_DIR/lms-manager"/* "$INSTALL_DIR/"
    
    # Cleanup
    cd /
    rm -rf "$TEMP_DIR"
    
    print_success "Update downloaded and extracted"
}

build_application() {
    print_info "Building application..."

    cd "$INSTALL_DIR"
    
    # Update dependencies
    go mod tidy
    
    # Build application
    go build -o lms-manager main.go
    
    # Set permissions
    chown "$SERVICE_USER:$SERVICE_GROUP" lms-manager
    chmod +x lms-manager
    
    print_success "Application built"
}

restore_config() {
    print_info "Restoring configuration..."

    # Restore configuration from backup
    if [[ -f "$BACKUP_PATH/config.json" ]]; then
        cp "$BACKUP_PATH/config.json" "$INSTALL_DIR/config/config.json"
        chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/config/config.json"
        chmod 600 "$INSTALL_DIR/config/config.json"
        print_success "Configuration restored"
    else
        print_warning "No configuration backup found"
    fi
}

restore_database() {
    print_info "Restoring database..."

    # Restore database from backup
    if [[ -f "$BACKUP_PATH/lms-manager.db" ]]; then
        cp "$BACKUP_PATH/lms-manager.db" "$INSTALL_DIR/lms-manager.db"
        chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/lms-manager.db"
        chmod 600 "$INSTALL_DIR/lms-manager.db"
        print_success "Database restored"
    else
        print_warning "No database backup found"
    fi
}

start_service() {
    print_info "Starting LMS Manager service..."

    systemctl start "$SERVICE_NAME"
    sleep 3

    if systemctl is-active --quiet "$SERVICE_NAME"; then
        print_success "Service started successfully"
    else
        print_error "Failed to start service"
        systemctl status "$SERVICE_NAME"
        exit 1
    fi
}

verify_update() {
    print_info "Verifying update..."

    # Check service status
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        print_success "Service is running"
    else
        print_error "Service is not running"
        exit 1
    fi

    # Check version
    NEW_VERSION=$($INSTALL_DIR/lms-manager --version 2>/dev/null || echo "unknown")
    print_info "New version: $NEW_VERSION"

    # Test health endpoint
    sleep 5
    if curl -f -s http://localhost:8080/health > /dev/null; then
        print_success "Health check passed"
    else
        print_warning "Health check failed, but service is running"
    fi
}

cleanup_old_backups() {
    print_info "Cleaning up old backups..."

    # Keep only last 5 backups
    cd "$BACKUP_DIR"
    ls -t | tail -n +6 | xargs -r rm -rf
    
    print_success "Old backups cleaned"
}

show_completion() {
    echo -e "${GREEN}"
    echo "=========================================="
    echo "    Update Completed Successfully!"
    echo "=========================================="
    echo -e "${NC}"
    echo ""
    echo "LMS Manager has been updated to version $LATEST_VERSION"
    echo ""
    echo "Update Summary:"
    echo "  ✓ Service stopped"
    echo "  ✓ Backup created"
    echo "  ✓ Update downloaded"
    echo "  ✓ Application built"
    echo "  ✓ Configuration restored"
    echo "  ✓ Database restored"
    echo "  ✓ Service started"
    echo "  ✓ Update verified"
    echo ""
    echo "Access Information:"
    echo "  URL: http://localhost:8080"
    echo "  Service Status: systemctl status $SERVICE_NAME"
    echo "  Service Logs: journalctl -u $SERVICE_NAME -f"
    echo ""
    echo "Backup Location: $BACKUP_PATH"
    echo ""
    echo "Support:"
    echo "  Website: https://k2net.id"
    echo "  Email: support@k2net.id"
    echo ""
}

# Main update process
main() {
    print_header

    check_root
    check_installation
    get_current_version
    check_latest_version
    create_backup
    stop_service
    download_update
    build_application
    restore_config
    restore_database
    start_service
    verify_update
    cleanup_old_backups
    show_completion

    print_success "Update completed successfully!"
}

# Run main function
main "$@"
