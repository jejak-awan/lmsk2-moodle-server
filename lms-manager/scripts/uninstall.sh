#!/bin/bash

# LMS Manager Uninstallation Script
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
SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"

# Functions
print_header() {
    echo -e "${BLUE}"
    echo "=========================================="
    echo "    LMS Manager - Uninstallation Script"
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

confirm_uninstall() {
    echo -e "${YELLOW}"
    echo "This will completely remove LMS Manager from your system."
    echo "This action cannot be undone!"
    echo ""
    echo "The following will be removed:"
    echo "  - LMS Manager application files"
    echo "  - Systemd service"
    echo "  - Configuration files"
    echo "  - Database files"
    echo "  - Log files"
    echo "  - Service user account"
    echo ""
    echo -e "${NC}"
    
    read -p "Are you sure you want to continue? (yes/no): " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        print_info "Uninstallation cancelled"
        exit 0
    fi
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

disable_service() {
    print_info "Disabling LMS Manager service..."

    if systemctl is-enabled --quiet "$SERVICE_NAME"; then
        systemctl disable "$SERVICE_NAME"
        print_success "Service disabled"
    else
        print_info "Service is not enabled"
    fi
}

remove_systemd_service() {
    print_info "Removing systemd service file..."

    if [[ -f "$SERVICE_FILE" ]]; then
        rm -f "$SERVICE_FILE"
        systemctl daemon-reload
        print_success "Systemd service file removed"
    else
        print_info "Systemd service file not found"
    fi
}

remove_application_files() {
    print_info "Removing application files..."

    if [[ -d "$INSTALL_DIR" ]]; then
        rm -rf "$INSTALL_DIR"
        print_success "Application files removed"
    else
        print_info "Application directory not found"
    fi
}

remove_user() {
    print_info "Removing service user..."

    if id "$SERVICE_USER" &>/dev/null; then
        userdel "$SERVICE_USER"
        print_success "Service user removed"
    else
        print_info "Service user not found"
    fi
}

remove_firewall_rules() {
    if command -v ufw &> /dev/null; then
        print_info "Removing firewall rules..."
        ufw delete allow 8080/tcp 2>/dev/null || true
        print_success "Firewall rules removed"
    elif command -v firewall-cmd &> /dev/null; then
        print_info "Removing firewall rules..."
        firewall-cmd --permanent --remove-port=8080/tcp 2>/dev/null || true
        firewall-cmd --reload 2>/dev/null || true
        print_success "Firewall rules removed"
    else
        print_info "No firewall detected"
    fi
}

cleanup_logs() {
    print_info "Cleaning up system logs..."

    # Remove journal logs
    journalctl --vacuum-time=1s --quiet 2>/dev/null || true

    print_success "System logs cleaned"
}

show_completion() {
    echo -e "${GREEN}"
    echo "=========================================="
    echo "    Uninstallation Completed Successfully!"
    echo "=========================================="
    echo -e "${NC}"
    echo ""
    echo "LMS Manager has been completely removed from your system."
    echo ""
    echo "The following have been removed:"
    echo "  ✓ Application files"
    echo "  ✓ Systemd service"
    echo "  ✓ Configuration files"
    echo "  ✓ Database files"
    echo "  ✓ Log files"
    echo "  ✓ Service user account"
    echo "  ✓ Firewall rules"
    echo ""
    echo "Thank you for using LMS Manager!"
    echo ""
    echo "Support:"
    echo "  Website: https://k2net.id"
    echo "  Email: support@k2net.id"
    echo ""
}

# Main uninstallation process
main() {
    print_header

    check_root
    confirm_uninstall
    stop_service
    disable_service
    remove_systemd_service
    remove_application_files
    remove_user
    remove_firewall_rules
    cleanup_logs
    show_completion

    print_success "Uninstallation completed successfully!"
}

# Run main function
main "$@"
