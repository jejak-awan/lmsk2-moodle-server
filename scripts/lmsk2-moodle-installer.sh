#!/bin/bash

# =============================================================================
# LMSK2 Moodle Server Master Installer
# =============================================================================
# Version: 1.0
# Author: jejakawan007
# Description: Master installer script untuk LMSK2-Moodle-Server
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Script information
SCRIPT_NAME="LMSK2 Moodle Server Installer"
SCRIPT_VERSION="1.0"
SCRIPT_AUTHOR="jejakawan007"
SCRIPT_DATE="$(date '+%Y-%m-%d')"

# Default configuration
DEFAULT_VERSION="3.11-lts"
DEFAULT_DOMAIN="lms.yourdomain.com"
DEFAULT_EMAIL="admin@yourdomain.com"
DEFAULT_DB_PASSWORD=""
DEFAULT_ADMIN_PASSWORD=""

# Paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="/var/log/lmsk2"
CONFIG_DIR="${SCRIPT_DIR}/config"
BACKUP_DIR="/backup/lmsk2"

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

PHASE=""
VERSION=""
DOMAIN=""
EMAIL=""
DB_PASSWORD=""
ADMIN_PASSWORD=""
INTERACTIVE=false
DRY_RUN=false
DEBUG=false
VERBOSE=false
LOG_LEVEL="info"
SKIP_CONFIRMATION=false

# =============================================================================
# Utility Functions
# =============================================================================

# Print colored output
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Print header
print_header() {
    echo
    print_color $CYAN "=============================================================================="
    print_color $WHITE "  $SCRIPT_NAME v$SCRIPT_VERSION"
    print_color $CYAN "=============================================================================="
    echo
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

# Print debug message
print_debug() {
    if [[ "$DEBUG" == "true" ]]; then
        print_color $PURPLE "🐛 $1"
    fi
}

# Log message
log_message() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Create log directory if it doesn't exist
    mkdir -p "$LOG_DIR"
    
    # Write to log file
    echo "[$timestamp] [$level] $message" >> "$LOG_DIR/installer.log"
    
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
        "DEBUG")
            print_debug "$message"
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

# Check system requirements
check_requirements() {
    print_section "Checking System Requirements"
    
    # Check OS
    if [[ ! -f /etc/os-release ]]; then
        print_error "Tidak dapat mendeteksi sistem operasi"
        exit 1
    fi
    
    source /etc/os-release
    if [[ "$ID" != "ubuntu" ]]; then
        print_warning "Sistem operasi yang dideteksi: $ID $VERSION_ID"
        print_warning "Script ini dioptimalkan untuk Ubuntu. Lanjutkan? (y/N)"
        if [[ "$SKIP_CONFIRMATION" != "true" ]]; then
            read -r response
            if [[ ! "$response" =~ ^[Yy]$ ]]; then
                print_info "Installation dibatalkan"
                exit 0
            fi
        fi
    else
        print_success "Sistem operasi: $PRETTY_NAME"
    fi
    
    # Check Ubuntu version
    if [[ "$VERSION_ID" == "22.04" ]]; then
        print_success "Ubuntu version: $VERSION_ID (LTS) - Fully supported"
    elif [[ "$VERSION_ID" == "24.04" ]]; then
        print_success "Ubuntu version: $VERSION_ID (LTS) - Supported"
    else
        print_warning "Ubuntu version: $VERSION_ID - May not be fully supported"
    fi
    
    # Check memory
    local memory_gb=$(free -g | awk 'NR==2{print $2}')
    if [[ $memory_gb -lt 4 ]]; then
        print_warning "Memory: ${memory_gb}GB (Recommended: 4GB+)"
    else
        print_success "Memory: ${memory_gb}GB"
    fi
    
    # Check disk space
    local disk_gb=$(df -BG / | awk 'NR==2{print $4}' | sed 's/G//')
    if [[ $disk_gb -lt 20 ]]; then
        print_warning "Disk space: ${disk_gb}GB (Recommended: 20GB+)"
    else
        print_success "Disk space: ${disk_gb}GB"
    fi
    
    # Check internet connection
    if ping -c 1 8.8.8.8 >/dev/null 2>&1; then
        print_success "Internet connection: OK"
    else
        print_error "Internet connection: Failed"
        exit 1
    fi
    
    log_message "INFO" "System requirements check completed"
}

# Create necessary directories
create_directories() {
    print_section "Creating Directories"
    
    local dirs=("$LOG_DIR" "$CONFIG_DIR" "$BACKUP_DIR" "/var/www/moodle" "/var/www/moodle/moodledata")
    
    for dir in "${dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            mkdir -p "$dir"
            print_success "Created directory: $dir"
        else
            print_info "Directory exists: $dir"
        fi
    done
    
    log_message "INFO" "Directories created successfully"
}

# Load configuration
load_config() {
    print_section "Loading Configuration"
    
    # Load from config file if exists
    if [[ -f "$CONFIG_DIR/installer.conf" ]]; then
        source "$CONFIG_DIR/installer.conf"
        print_success "Configuration loaded from: $CONFIG_DIR/installer.conf"
    else
        print_info "No configuration file found, using defaults"
    fi
    
    # Set defaults if not provided
    VERSION=${VERSION:-$DEFAULT_VERSION}
    DOMAIN=${DOMAIN:-$DEFAULT_DOMAIN}
    EMAIL=${EMAIL:-$DEFAULT_EMAIL}
    
    log_message "INFO" "Configuration loaded: version=$VERSION, domain=$DOMAIN, email=$EMAIL"
}

# Validate configuration
validate_config() {
    print_section "Validating Configuration"
    
    # Validate version
    case $VERSION in
        "3.11-lts"|"4.0"|"4.1"|"5.0")
            print_success "Moodle version: $VERSION"
            ;;
        *)
            print_error "Invalid Moodle version: $VERSION"
            print_info "Supported versions: 3.11-lts, 4.0, 4.1, 5.0"
            exit 1
            ;;
    esac
    
    # Validate domain
    if [[ ! "$DOMAIN" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
        print_error "Invalid domain format: $DOMAIN"
        exit 1
    else
        print_success "Domain: $DOMAIN"
    fi
    
    # Validate email
    if [[ ! "$EMAIL" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
        print_error "Invalid email format: $EMAIL"
        exit 1
    else
        print_success "Email: $EMAIL"
    fi
    
    # Check if passwords are provided
    if [[ -z "$DB_PASSWORD" ]]; then
        if [[ "$INTERACTIVE" == "true" ]]; then
            print_info "Enter database password:"
            read -s DB_PASSWORD
            echo
        else
            print_error "Database password is required"
            exit 1
        fi
    fi
    
    if [[ -z "$ADMIN_PASSWORD" ]]; then
        if [[ "$INTERACTIVE" == "true" ]]; then
            print_info "Enter admin password:"
            read -s ADMIN_PASSWORD
            echo
        else
            print_error "Admin password is required"
            exit 1
        fi
    fi
    
    log_message "INFO" "Configuration validation completed"
}

# Show configuration summary
show_config_summary() {
    print_section "Configuration Summary"
    
    echo "Phase: $PHASE"
    echo "Moodle Version: $VERSION"
    echo "Domain: $DOMAIN"
    echo "Email: $EMAIL"
    echo "Interactive Mode: $INTERACTIVE"
    echo "Dry Run: $DRY_RUN"
    echo "Debug: $DEBUG"
    echo "Verbose: $VERBOSE"
    echo "Log Level: $LOG_LEVEL"
    echo
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_warning "DRY RUN MODE - No changes will be made"
    fi
    
    if [[ "$SKIP_CONFIRMATION" != "true" ]]; then
        print_info "Continue with installation? (y/N)"
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            print_info "Installation cancelled"
            exit 0
        fi
    fi
}

# Execute phase script
execute_phase() {
    local phase_num=$1
    local phase_name=$2
    local script_name=$3
    
    print_section "Executing Phase $phase_num: $phase_name"
    
    local script_path="$SCRIPT_DIR/$script_name"
    
    if [[ ! -f "$script_path" ]]; then
        print_error "Script not found: $script_path"
        return 1
    fi
    
    if [[ ! -x "$script_path" ]]; then
        print_info "Making script executable: $script_path"
        chmod +x "$script_path"
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "DRY RUN: Would execute $script_path"
        return 0
    fi
    
    # Set environment variables for the script
    export LMSK2_VERSION="$VERSION"
    export LMSK2_DOMAIN="$DOMAIN"
    export LMSK2_EMAIL="$EMAIL"
    export LMSK2_DB_PASSWORD="$DB_PASSWORD"
    export LMSK2_ADMIN_PASSWORD="$ADMIN_PASSWORD"
    export LMSK2_DEBUG="$DEBUG"
    export LMSK2_VERBOSE="$VERBOSE"
    export LMSK2_LOG_LEVEL="$LOG_LEVEL"
    
    # Execute the script
    print_info "Executing: $script_path"
    if "$script_path"; then
        print_success "Phase $phase_num completed successfully"
        log_message "SUCCESS" "Phase $phase_num ($phase_name) completed"
        return 0
    else
        print_error "Phase $phase_num failed"
        log_message "ERROR" "Phase $phase_num ($phase_name) failed"
        return 1
    fi
}

# Execute all phases
execute_all_phases() {
    print_section "Executing All Phases"
    
    local phases=(
        "1:Server Preparation:01-server-preparation.sh"
        "2:Software Installation:02-software-installation.sh"
        "3:Security Hardening:03-security-hardening.sh"
        "4:Basic Configuration:04-basic-configuration.sh"
        "5:Moodle Installation:01-moodle-${VERSION}-install.sh"
        "6:Performance Tuning:01-performance-tuning.sh"
        "7:Caching Setup:02-caching-setup.sh"
        "8:Monitoring Setup:03-monitoring-setup.sh"
        "9:Backup Strategy:04-backup-strategy.sh"
    )
    
    for phase_info in "${phases[@]}"; do
        IFS=':' read -r phase_num phase_name script_name <<< "$phase_info"
        
        if execute_phase "$phase_num" "$phase_name" "$script_name"; then
            continue
        else
            print_error "Installation failed at phase $phase_num"
            return 1
        fi
    done
    
    print_success "All phases completed successfully"
    return 0
}

# Execute specific phase
execute_specific_phase() {
    case $PHASE in
        "1")
            execute_phase "1" "Server Preparation" "01-server-preparation.sh"
            ;;
        "2")
            execute_phase "2" "Software Installation" "02-software-installation.sh"
            ;;
        "3")
            execute_phase "3" "Security Hardening" "03-security-hardening.sh"
            ;;
        "4")
            execute_phase "4" "Basic Configuration" "04-basic-configuration.sh"
            ;;
        "5")
            execute_phase "5" "Moodle Installation" "01-moodle-${VERSION}-install.sh"
            ;;
        *)
            print_error "Invalid phase: $PHASE"
            print_info "Valid phases: 1, 2, 3, 4, 5, all"
            exit 1
            ;;
    esac
}

# Run system verification
run_verification() {
    print_section "System Verification"
    
    local verification_script="$SCRIPT_DIR/system-verification.sh"
    
    if [[ -f "$verification_script" && -x "$verification_script" ]]; then
        print_info "Running system verification..."
        if "$verification_script"; then
            print_success "System verification passed"
            return 0
        else
            print_error "System verification failed"
            return 1
        fi
    else
        print_warning "System verification script not found or not executable"
        return 0
    fi
}

# Show usage information
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

LMSK2 Moodle Server Master Installer v$SCRIPT_VERSION

OPTIONS:
    --phase=PHASE           Phase to execute (1, 2, 3, 4, 5, all)
    --version=VERSION       Moodle version (3.11-lts, 4.0, 4.1, 5.0)
    --domain=DOMAIN         Domain name for Moodle
    --email=EMAIL           Admin email address
    --db-password=PASSWORD  Database password
    --admin-password=PASS   Admin password
    --interactive           Interactive mode
    --dry-run              Test mode (no changes)
    --debug                Debug mode
    --verbose              Verbose output
    --log-level=LEVEL      Log level (debug, info, warning, error)
    --skip-confirmation    Skip confirmation prompts
    --help                 Show this help message

EXAMPLES:
    $0 --phase=all --version=3.11-lts --domain=lms.example.com --interactive
    $0 --phase=1 --dry-run
    $0 --phase=2 --version=4.0 --domain=lms.example.com --email=admin@example.com

PHASES:
    1 - Server Preparation
    2 - Software Installation  
    3 - Security Hardening
    4 - Basic Configuration
    5 - Moodle Installation
    all - Execute all phases

EOF
}

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --phase=*)
                PHASE="${1#*=}"
                shift
                ;;
            --version=*)
                VERSION="${1#*=}"
                shift
                ;;
            --domain=*)
                DOMAIN="${1#*=}"
                shift
                ;;
            --email=*)
                EMAIL="${1#*=}"
                shift
                ;;
            --db-password=*)
                DB_PASSWORD="${1#*=}"
                shift
                ;;
            --admin-password=*)
                ADMIN_PASSWORD="${1#*=}"
                shift
                ;;
            --interactive)
                INTERACTIVE=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --debug)
                DEBUG=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --log-level=*)
                LOG_LEVEL="${1#*=}"
                shift
                ;;
            --skip-confirmation)
                SKIP_CONFIRMATION=true
                shift
                ;;
            --help)
                show_usage
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

# Main function
main() {
    # Parse arguments
    parse_arguments "$@"
    
    # Show header
    print_header
    
    # Check if running as root
    check_root
    
    # Check system requirements
    check_requirements
    
    # Create necessary directories
    create_directories
    
    # Load configuration
    load_config
    
    # Validate configuration
    validate_config
    
    # Show configuration summary
    show_config_summary
    
    # Execute phases
    if [[ "$PHASE" == "all" ]]; then
        execute_all_phases
    elif [[ -n "$PHASE" ]]; then
        execute_specific_phase
    else
        print_error "No phase specified"
        show_usage
        exit 1
    fi
    
    # Run verification
    run_verification
    
    # Final message
    print_section "Installation Complete"
    print_success "LMSK2 Moodle Server installation completed successfully!"
    print_info "Log files: $LOG_DIR/installer.log"
    print_info "Configuration: $CONFIG_DIR/"
    print_info "Backup directory: $BACKUP_DIR/"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_warning "This was a dry run - no changes were made"
    fi
    
    log_message "SUCCESS" "Installation completed successfully"
}

# Trap errors
trap 'print_error "Script failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
