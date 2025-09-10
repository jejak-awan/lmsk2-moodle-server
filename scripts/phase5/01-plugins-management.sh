#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: Plugins Management Script
# =============================================================================
# Description: Advanced plugins management for Moodle
# Version: 1.0
# Author: jejakawan007
# Date: September 9, 2025
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Script information
SCRIPT_NAME="LMSK2 Plugins Management"
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

# Paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="/var/log/lmsk2"
CONFIG_DIR="${SCRIPT_DIR}/../config"
MOODLE_DIR="/var/www/moodle"
MOODLE_DATA_DIR="/var/www/moodle/moodledata"

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
    mkdir -p "$LOG_DIR"
    
    # Write to log file
    echo "[$timestamp] [$level] $message" >> "$LOG_DIR/plugins-management.log"
    
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
# Plugin Management Functions
# =============================================================================

# Install essential plugins
install_essential_plugins() {
    print_section "Installing Essential Plugins"
    
    local essential_plugins=(
        "mod_bigbluebuttonbn"
        "mod_hvp"
        "mod_quiz"
        "mod_assign"
        "mod_forum"
        "mod_lesson"
        "mod_choice"
        "mod_feedback"
        "block_myoverview"
        "block_timeline"
        "block_recentlyaccesseditems"
        "block_starredcourses"
        "theme_boost"
        "auth_ldap"
        "enrol_ldap"
        "tool_moodlenet"
        "tool_usertours"
        "tool_mobile"
        "tool_analytics"
        "tool_dataprivacy"
        "tool_policy"
        "tool_recyclebin"
        "tool_task"
        "tool_uploaduser"
        "tool_xmldb"
    )
    
    for plugin in "${essential_plugins[@]}"; do
        print_info "Installing plugin: $plugin"
        
        # Check if plugin is already installed
        if [[ -d "$MOODLE_DIR/mod/$plugin" ]] || [[ -d "$MOODLE_DIR/blocks/$plugin" ]] || [[ -d "$MOODLE_DIR/auth/$plugin" ]]; then
            print_success "Plugin $plugin already installed"
            continue
        fi
        
        # Install plugin using Moodle CLI
        if php "$MOODLE_DIR/admin/cli/install_plugins.php" --install="$plugin" --non-interactive; then
            print_success "Plugin $plugin installed successfully"
            log_message "SUCCESS" "Plugin $plugin installed"
        else
            print_warning "Failed to install plugin $plugin"
            log_message "WARNING" "Failed to install plugin $plugin"
        fi
    done
    
    log_message "SUCCESS" "Essential plugins installation completed"
}

# Install third-party plugins
install_third_party_plugins() {
    print_section "Installing Third-Party Plugins"
    
    # Create plugins directory
    mkdir -p "$MOODLE_DIR/local/plugins"
    
    # Download and install popular third-party plugins
    local third_party_plugins=(
        "https://github.com/moodlehq/moodle-local_mobile/releases/latest/download/local_mobile.zip"
        "https://github.com/moodlehq/moodle-local_boost/releases/latest/download/local_boost.zip"
        "https://github.com/moodlehq/moodle-local_announcements/releases/latest/download/local_announcements.zip"
    )
    
    for plugin_url in "${third_party_plugins[@]}"; do
        local plugin_name=$(basename "$plugin_url" .zip)
        print_info "Installing third-party plugin: $plugin_name"
        
        # Download plugin
        if wget -q "$plugin_url" -O "/tmp/$plugin_name.zip"; then
            # Extract plugin
            if unzip -q "/tmp/$plugin_name.zip" -d "/tmp/"; then
                # Move to appropriate directory
                if [[ -d "/tmp/$plugin_name" ]]; then
                    cp -r "/tmp/$plugin_name" "$MOODLE_DIR/local/"
                    chown -R www-data:www-data "$MOODLE_DIR/local/$plugin_name"
                    chmod -R 755 "$MOODLE_DIR/local/$plugin_name"
                    print_success "Third-party plugin $plugin_name installed"
                    log_message "SUCCESS" "Third-party plugin $plugin_name installed"
                fi
            fi
            
            # Cleanup
            rm -f "/tmp/$plugin_name.zip"
            rm -rf "/tmp/$plugin_name"
        else
            print_warning "Failed to download plugin: $plugin_name"
            log_message "WARNING" "Failed to download plugin: $plugin_name"
        fi
    done
    
    log_message "SUCCESS" "Third-party plugins installation completed"
}

# Configure plugin settings
configure_plugin_settings() {
    print_section "Configuring Plugin Settings"
    
    # Configure BigBlueButton
    print_info "Configuring BigBlueButton plugin..."
    if [[ -d "$MOODLE_DIR/mod/bigbluebuttonbn" ]]; then
        # Set BigBlueButton server URL
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=bigbluebuttonbn_server_url --set="https://bbb.example.com/bigbluebutton/"
        
        # Set shared secret
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=bigbluebuttonbn_shared_secret --set="your_shared_secret_here"
        
        print_success "BigBlueButton configured"
    fi
    
    # Configure H5P
    print_info "Configuring H5P plugin..."
    if [[ -d "$MOODLE_DIR/mod/hvp" ]]; then
        # Enable H5P content types
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=hvp_enable_lrs_content_types --set=1
        
        print_success "H5P configured"
    fi
    
    # Configure Mobile app
    print_info "Configuring Mobile app..."
    if [[ -d "$MOODLE_DIR/admin/tool/mobile" ]]; then
        # Enable mobile services
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=enablemobilewebservice --set=1
        
        print_success "Mobile app configured"
    fi
    
    log_message "SUCCESS" "Plugin settings configuration completed"
}

# Update plugins
update_plugins() {
    print_section "Updating Plugins"
    
    print_info "Checking for plugin updates..."
    
    # Update all plugins
    if php "$MOODLE_DIR/admin/cli/upgrade.php" --non-interactive; then
        print_success "All plugins updated successfully"
        log_message "SUCCESS" "Plugins updated successfully"
    else
        print_warning "Some plugins failed to update"
        log_message "WARNING" "Some plugins failed to update"
    fi
    
    # Clear caches
    print_info "Clearing caches..."
    php "$MOODLE_DIR/admin/cli/purge_caches.php"
    
    log_message "SUCCESS" "Plugin update completed"
}

# Backup plugins
backup_plugins() {
    print_section "Backing Up Plugins"
    
    local backup_dir="/backup/lmsk2/plugins/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    print_info "Creating plugin backup in: $backup_dir"
    
    # Backup custom plugins
    if [[ -d "$MOODLE_DIR/local" ]]; then
        cp -r "$MOODLE_DIR/local" "$backup_dir/"
        print_success "Custom plugins backed up"
    fi
    
    # Backup plugin configurations
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=* > "$backup_dir/plugin_configs.txt"
    print_success "Plugin configurations backed up"
    
    # Create backup archive
    cd "$backup_dir/.."
    tar -czf "$(basename "$backup_dir").tar.gz" "$(basename "$backup_dir")"
    rm -rf "$backup_dir"
    
    print_success "Plugin backup created: $(basename "$backup_dir").tar.gz"
    log_message "SUCCESS" "Plugin backup completed"
}

# Monitor plugin performance
monitor_plugin_performance() {
    print_section "Monitoring Plugin Performance"
    
    print_info "Checking plugin performance..."
    
    # Check for slow plugins
    local slow_plugins=$(php "$MOODLE_DIR/admin/cli/task.php" --list | grep -i "slow\|timeout" || true)
    if [[ -n "$slow_plugins" ]]; then
        print_warning "Slow plugins detected:"
        echo "$slow_plugins"
        log_message "WARNING" "Slow plugins detected"
    else
        print_success "No slow plugins detected"
    fi
    
    # Check plugin errors
    local plugin_errors=$(tail -n 100 "$MOODLE_DATA_DIR/error.log" | grep -i "plugin\|error" | wc -l)
    if [[ "$plugin_errors" -gt 0 ]]; then
        print_warning "Plugin errors detected: $plugin_errors"
        log_message "WARNING" "Plugin errors detected: $plugin_errors"
    else
        print_success "No plugin errors detected"
    fi
    
    log_message "SUCCESS" "Plugin performance monitoring completed"
}

# =============================================================================
# Main Function
# =============================================================================

main() {
    print_color $CYAN "=============================================================================="
    print_color $WHITE "  $SCRIPT_NAME v$SCRIPT_VERSION"
    print_color $CYAN "=============================================================================="
    
    log_message "INFO" "Starting plugins management"
    
    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root"
        exit 1
    fi
    
    # Check if Moodle is installed
    if [[ ! -d "$MOODLE_DIR" ]]; then
        print_error "Moodle directory not found: $MOODLE_DIR"
        exit 1
    fi
    
    # Parse command line arguments
    local action="all"
    while [[ $# -gt 0 ]]; do
        case $1 in
            --install-essential)
                action="essential"
                shift
                ;;
            --install-third-party)
                action="third-party"
                shift
                ;;
            --configure)
                action="configure"
                shift
                ;;
            --update)
                action="update"
                shift
                ;;
            --backup)
                action="backup"
                shift
                ;;
            --monitor)
                action="monitor"
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --install-essential    Install essential plugins"
                echo "  --install-third-party  Install third-party plugins"
                echo "  --configure           Configure plugin settings"
                echo "  --update              Update all plugins"
                echo "  --backup              Backup plugins"
                echo "  --monitor             Monitor plugin performance"
                echo "  --help                Show this help"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Execute based on action
    case $action in
        "essential")
            install_essential_plugins
            ;;
        "third-party")
            install_third_party_plugins
            ;;
        "configure")
            configure_plugin_settings
            ;;
        "update")
            update_plugins
            ;;
        "backup")
            backup_plugins
            ;;
        "monitor")
            monitor_plugin_performance
            ;;
        "all")
            install_essential_plugins
            install_third_party_plugins
            configure_plugin_settings
            update_plugins
            backup_plugins
            monitor_plugin_performance
            ;;
    esac
    
    print_section "Plugins Management Complete"
    print_success "Plugins management completed successfully!"
    print_info "Log file: $LOG_DIR/plugins-management.log"
    
    log_message "SUCCESS" "Plugins management completed successfully"
}

# Trap errors
trap 'print_error "Script failed at line $LINENO"; log_message "ERROR" "Script failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
