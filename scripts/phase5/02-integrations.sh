#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: Integrations Script
# =============================================================================
# Description: Advanced integrations setup for Moodle
# Version: 1.0
# Author: jejakawan007
# Date: September 9, 2025
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Script information
SCRIPT_NAME="LMSK2 Integrations"
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
    echo "[$timestamp] [$level] $message" >> "$LOG_DIR/integrations.log"
    
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
# Integration Functions
# =============================================================================

# Setup LDAP integration
setup_ldap_integration() {
    print_section "Setting Up LDAP Integration"
    
    print_info "Installing LDAP authentication plugin..."
    
    # Install LDAP plugin if not already installed
    if [[ ! -d "$MOODLE_DIR/auth/ldap" ]]; then
        print_info "LDAP plugin not found, installing..."
        php "$MOODLE_DIR/admin/cli/install_plugins.php" --install="auth_ldap" --non-interactive
    fi
    
    # Configure LDAP settings
    print_info "Configuring LDAP settings..."
    
    # LDAP server configuration
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_host_url --set="ldap://ldap.example.com:389"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_version --set="3"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_ldap_encoding --set="utf-8"
    
    # LDAP bind settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_bind_dn --set="cn=admin,dc=example,dc=com"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_bind_pw --set="ldap_password"
    
    # User lookup settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_user_type --set="posixAccount"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_contexts --set="ou=users,dc=example,dc=com"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_search_sub --set="1"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_opt_deref --set="0"
    
    # User attribute mapping
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_user_attribute --set="uid"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_memberattribute --set="memberUid"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_memberattribute_isdn --set="0"
    
    # Field mapping
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_attrcreators --set=""
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_ldap_groupecreators --set=""
    
    # Enable LDAP authentication
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth --set="ldap,manual"
    
    print_success "LDAP integration configured"
    log_message "SUCCESS" "LDAP integration setup completed"
}

# Setup OAuth2 integration
setup_oauth2_integration() {
    print_section "Setting Up OAuth2 Integration"
    
    print_info "Installing OAuth2 authentication plugin..."
    
    # Install OAuth2 plugin
    if [[ ! -d "$MOODLE_DIR/auth/oauth2" ]]; then
        php "$MOODLE_DIR/admin/cli/install_plugins.php" --install="auth_oauth2" --non-interactive
    fi
    
    # Configure OAuth2 providers
    print_info "Configuring OAuth2 providers..."
    
    # Google OAuth2
    print_info "Setting up Google OAuth2..."
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_oauth2_google_clientid --set="your_google_client_id"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_oauth2_google_clientsecret --set="your_google_client_secret"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_oauth2_google_scope --set="openid profile email"
    
    # Microsoft OAuth2
    print_info "Setting up Microsoft OAuth2..."
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_oauth2_microsoft_clientid --set="your_microsoft_client_id"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_oauth2_microsoft_clientsecret --set="your_microsoft_client_secret"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_oauth2_microsoft_scope --set="openid profile email"
    
    # Facebook OAuth2
    print_info "Setting up Facebook OAuth2..."
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_oauth2_facebook_clientid --set="your_facebook_client_id"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_oauth2_facebook_clientsecret --set="your_facebook_client_secret"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_oauth2_facebook_scope --set="email"
    
    # Enable OAuth2 authentication
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth --set="oauth2,ldap,manual"
    
    print_success "OAuth2 integration configured"
    log_message "SUCCESS" "OAuth2 integration setup completed"
}

# Setup API integrations
setup_api_integrations() {
    print_section "Setting Up API Integrations"
    
    print_info "Enabling Moodle web services..."
    
    # Enable web services
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enablewebservices --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enablemobilewebservice --set=1
    
    # Configure REST protocol
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=webserviceprotocols --set="rest,soap"
    
    # Enable external services
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=allowexternaltools --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=allowembedding --set=1
    
    # Configure CORS
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=allowedcorsorigins --set="*"
    
    print_info "Creating API service user..."
    
    # Create service user for API access
    local service_user="moodle_api"
    local service_password=$(openssl rand -base64 32)
    
    # Create user if not exists
    if ! id "$service_user" &>/dev/null; then
        useradd -m -s /bin/bash "$service_user"
        echo "$service_user:$service_password" | chpasswd
        print_success "API service user created: $service_user"
    fi
    
    # Store credentials
    echo "API_USER=$service_user" > /etc/moodle-api-credentials
    echo "API_PASSWORD=$service_password" >> /etc/moodle-api-credentials
    chmod 600 /etc/moodle-api-credentials
    
    print_success "API integrations configured"
    log_message "SUCCESS" "API integrations setup completed"
}

# Setup third-party service integrations
setup_third_party_integrations() {
    print_section "Setting Up Third-Party Service Integrations"
    
    # BigBlueButton integration
    print_info "Configuring BigBlueButton integration..."
    if [[ -d "$MOODLE_DIR/mod/bigbluebuttonbn" ]]; then
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=bigbluebuttonbn_server_url --set="https://bbb.example.com/bigbluebutton/"
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=bigbluebuttonbn_shared_secret --set="your_bbb_secret"
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=bigbluebuttonbn_default_dpa_accepted --set=1
        print_success "BigBlueButton integration configured"
    fi
    
    # Zoom integration
    print_info "Configuring Zoom integration..."
    if [[ -d "$MOODLE_DIR/mod/zoom" ]]; then
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=zoom_apikey --set="your_zoom_api_key"
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=zoom_apisecret --set="your_zoom_api_secret"
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=zoom_webhook_secret --set="your_zoom_webhook_secret"
        print_success "Zoom integration configured"
    fi
    
    # Turnitin integration
    print_info "Configuring Turnitin integration..."
    if [[ -d "$MOODLE_DIR/mod/assign/feedback/editpdf" ]]; then
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=turnitin_enablepseudo --set=1
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=turnitin_enablegrademark --set=1
        print_success "Turnitin integration configured"
    fi
    
    # H5P integration
    print_info "Configuring H5P integration..."
    if [[ -d "$MOODLE_DIR/mod/hvp" ]]; then
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=hvp_enable_lrs_content_types --set=1
        php "$MOODLE_DIR/admin/cli/cfg.php" --name=hvp_hub_enabled --set=1
        print_success "H5P integration configured"
    fi
    
    log_message "SUCCESS" "Third-party integrations setup completed"
}

# Setup email integrations
setup_email_integrations() {
    print_section "Setting Up Email Integrations"
    
    print_info "Configuring SMTP settings..."
    
    # SMTP configuration
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=smtphosts --set="smtp.gmail.com:587"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=smtpuser --set="your_email@gmail.com"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=smtppass --set="your_app_password"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=smtpsecure --set="tls"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=smtpauthtype --set="LOGIN"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=smtpmaxbulk --set="250"
    
    # Email settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=noreplyaddress --set="noreply@yourdomain.com"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=emailonlyfromnoreplyaddress --set=1
    
    print_info "Configuring email templates..."
    
    # Enable HTML emails
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=allowusermailcharset --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=emailcharset --set="utf-8"
    
    print_success "Email integrations configured"
    log_message "SUCCESS" "Email integrations setup completed"
}

# Setup calendar integrations
setup_calendar_integrations() {
    print_section "Setting Up Calendar Integrations"
    
    print_info "Configuring calendar settings..."
    
    # Enable calendar
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=calendar_site_timeformat --set="12"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=calendar_startwday --set="1"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=calendar_weekend --set="0,6"
    
    # Google Calendar integration
    print_info "Setting up Google Calendar integration..."
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=googlecalendar_clientid --set="your_google_client_id"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=googlecalendar_clientsecret --set="your_google_client_secret"
    
    # Outlook Calendar integration
    print_info "Setting up Outlook Calendar integration..."
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=outlook_calendar_clientid --set="your_outlook_client_id"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=outlook_calendar_clientsecret --set="your_outlook_client_secret"
    
    print_success "Calendar integrations configured"
    log_message "SUCCESS" "Calendar integrations setup completed"
}

# Test integrations
test_integrations() {
    print_section "Testing Integrations"
    
    print_info "Testing LDAP connection..."
    if php "$MOODLE_DIR/admin/cli/test_ldap.php"; then
        print_success "LDAP connection test passed"
    else
        print_warning "LDAP connection test failed"
    fi
    
    print_info "Testing OAuth2 providers..."
    # Test OAuth2 configuration
    local oauth2_providers=("google" "microsoft" "facebook")
    for provider in "${oauth2_providers[@]}"; do
        if php "$MOODLE_DIR/admin/cli/cfg.php" --name="auth_oauth2_${provider}_clientid" --get | grep -q "your_.*_client_id"; then
            print_warning "OAuth2 $provider not configured (using placeholder)"
        else
            print_success "OAuth2 $provider configured"
        fi
    done
    
    print_info "Testing web services..."
    if curl -s -o /dev/null -w "%{http_code}" "http://localhost/moodle/webservice/rest/server.php" | grep -q "200"; then
        print_success "Web services are accessible"
    else
        print_warning "Web services test failed"
    fi
    
    print_info "Testing email configuration..."
    if php "$MOODLE_DIR/admin/cli/cfg.php" --name=smtphosts --get | grep -q "smtp"; then
        print_success "SMTP configuration found"
    else
        print_warning "SMTP not configured"
    fi
    
    log_message "SUCCESS" "Integration tests completed"
}

# =============================================================================
# Main Function
# =============================================================================

main() {
    print_color $CYAN "=============================================================================="
    print_color $WHITE "  $SCRIPT_NAME v$SCRIPT_VERSION"
    print_color $CYAN "=============================================================================="
    
    log_message "INFO" "Starting integrations setup"
    
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
            --ldap)
                action="ldap"
                shift
                ;;
            --oauth2)
                action="oauth2"
                shift
                ;;
            --api)
                action="api"
                shift
                ;;
            --third-party)
                action="third-party"
                shift
                ;;
            --email)
                action="email"
                shift
                ;;
            --calendar)
                action="calendar"
                shift
                ;;
            --test)
                action="test"
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --ldap         Setup LDAP integration"
                echo "  --oauth2       Setup OAuth2 integration"
                echo "  --api          Setup API integrations"
                echo "  --third-party  Setup third-party integrations"
                echo "  --email        Setup email integrations"
                echo "  --calendar     Setup calendar integrations"
                echo "  --test         Test all integrations"
                echo "  --help         Show this help"
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
        "ldap")
            setup_ldap_integration
            ;;
        "oauth2")
            setup_oauth2_integration
            ;;
        "api")
            setup_api_integrations
            ;;
        "third-party")
            setup_third_party_integrations
            ;;
        "email")
            setup_email_integrations
            ;;
        "calendar")
            setup_calendar_integrations
            ;;
        "test")
            test_integrations
            ;;
        "all")
            setup_ldap_integration
            setup_oauth2_integration
            setup_api_integrations
            setup_third_party_integrations
            setup_email_integrations
            setup_calendar_integrations
            test_integrations
            ;;
    esac
    
    print_section "Integrations Setup Complete"
    print_success "Integrations setup completed successfully!"
    print_info "Log file: $LOG_DIR/integrations.log"
    
    log_message "SUCCESS" "Integrations setup completed successfully"
}

# Trap errors
trap 'print_error "Script failed at line $LINENO"; log_message "ERROR" "Script failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
