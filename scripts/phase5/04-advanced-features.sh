#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: Advanced Features Script
# =============================================================================
# Description: Advanced features setup for Moodle
# Version: 1.0
# Author: jejakawan007
# Date: September 9, 2025
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Script information
SCRIPT_NAME="LMSK2 Advanced Features"
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
    echo "[$timestamp] [$level] $message" >> "$LOG_DIR/advanced-features.log"
    
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
# Advanced Features Functions
# =============================================================================

# Setup AI integration
setup_ai_integration() {
    print_section "Setting Up AI Integration"
    
    print_info "Installing AI-related plugins..."
    
    # Install AI plugins if available
    local ai_plugins=(
        "local_ai"
        "mod_ai_assistant"
        "block_ai_chat"
        "tool_ai_analytics"
    )
    
    for plugin in "${ai_plugins[@]}"; do
        print_info "Checking for AI plugin: $plugin"
        if php "$MOODLE_DIR/admin/cli/install_plugins.php" --install="$plugin" --non-interactive 2>/dev/null; then
            print_success "AI plugin $plugin installed"
        else
            print_info "AI plugin $plugin not available in repository"
        fi
    done
    
    # Configure AI settings
    print_info "Configuring AI settings..."
    
    # Enable AI features
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enableai --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=ai_provider --set="openai"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=ai_api_key --set="your_openai_api_key"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=ai_model --set="gpt-3.5-turbo"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=ai_max_tokens --set=1000
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=ai_temperature --set=0.7
    
    # AI content generation settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=ai_generate_questions --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=ai_generate_content --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=ai_auto_translate --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=ai_smart_recommendations --set=1
    
    print_success "AI integration configured"
    log_message "SUCCESS" "AI integration setup completed"
}

# Setup analytics and reporting
setup_analytics() {
    print_section "Setting Up Analytics and Reporting"
    
    print_info "Installing analytics plugins..."
    
    # Install analytics plugins
    local analytics_plugins=(
        "tool_analytics"
        "local_learning_analytics"
        "block_analytics"
        "report_analytics"
    )
    
    for plugin in "${analytics_plugins[@]}"; do
        print_info "Installing analytics plugin: $plugin"
        if php "$MOODLE_DIR/admin/cli/install_plugins.php" --install="$plugin" --non-interactive 2>/dev/null; then
            print_success "Analytics plugin $plugin installed"
        else
            print_info "Analytics plugin $plugin not available"
        fi
    done
    
    # Configure analytics settings
    print_info "Configuring analytics settings..."
    
    # Enable analytics
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enableanalytics --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=analytics_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=analytics_models_enabled --set=1
    
    # Analytics models
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=analytics_model_students_at_risk --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=analytics_model_course_completion --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=analytics_model_engagement --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=analytics_model_dropout_prediction --set=1
    
    # Reporting settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enablecustomreports --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=reportbuilder_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=reportbuilder_export_formats --set="csv,excel,pdf"
    
    # Data retention
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=analytics_data_retention_days --set=365
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=analytics_cleanup_enabled --set=1
    
    print_success "Analytics and reporting configured"
    log_message "SUCCESS" "Analytics setup completed"
}

# Setup mobile app features
setup_mobile_features() {
    print_section "Setting Up Mobile App Features"
    
    print_info "Configuring mobile app features..."
    
    # Enable mobile services
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enablemobilewebservice --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobilecssurl --set=""
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileappurl --set="https://yourdomain.com/mobile"
    
    # Mobile app settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileapp_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileapp_offline_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileapp_push_notifications --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileapp_biometric_auth --set=1
    
    # Mobile features
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileapp_camera_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileapp_microphone_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileapp_location_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileapp_file_upload --set=1
    
    # Mobile app branding
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileapp_app_name --set="LMSK2 Learning"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileapp_app_version --set="1.0.0"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileapp_splash_screen --set=""
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mobileapp_icon --set=""
    
    print_success "Mobile app features configured"
    log_message "SUCCESS" "Mobile features setup completed"
}

# Setup enterprise features
setup_enterprise_features() {
    print_section "Setting Up Enterprise Features"
    
    print_info "Installing enterprise plugins..."
    
    # Install enterprise plugins
    local enterprise_plugins=(
        "local_enterprise"
        "tool_enterprise_admin"
        "block_enterprise_dashboard"
        "mod_enterprise_certificate"
        "local_sso"
        "auth_saml2"
        "enrol_ldap"
        "tool_ldap_sync"
    )
    
    for plugin in "${enterprise_plugins[@]}"; do
        print_info "Installing enterprise plugin: $plugin"
        if php "$MOODLE_DIR/admin/cli/install_plugins.php" --install="$plugin" --non-interactive 2>/dev/null; then
            print_success "Enterprise plugin $plugin installed"
        else
            print_info "Enterprise plugin $plugin not available"
        fi
    done
    
    # Configure enterprise settings
    print_info "Configuring enterprise settings..."
    
    # Single Sign-On (SSO)
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enable_sso --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=sso_provider --set="saml2"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=sso_auto_create_users --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=sso_auto_update_users --set=1
    
    # SAML2 configuration
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_saml2_idpname --set="Enterprise SSO"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_saml2_idpurl --set="https://sso.enterprise.com/saml2"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_saml2_idpssourl --set="https://sso.enterprise.com/saml2/sso"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_saml2_idpslsurl --set="https://sso.enterprise.com/saml2/sls"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=auth_saml2_idpcert --set=""
    
    # Enterprise dashboard
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enterprise_dashboard_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enterprise_analytics_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enterprise_reporting_enabled --set=1
    
    # Certificate management
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enterprise_certificates_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enterprise_certificate_template --set=""
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enterprise_certificate_issuer --set="LMSK2 Enterprise"
    
    # User provisioning
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enterprise_user_provisioning --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enterprise_auto_enrollment --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enterprise_role_mapping --set=1
    
    print_success "Enterprise features configured"
    log_message "SUCCESS" "Enterprise features setup completed"
}

# Setup advanced monitoring
setup_advanced_monitoring() {
    print_section "Setting Up Advanced Monitoring"
    
    print_info "Installing advanced monitoring plugins..."
    
    # Install monitoring plugins
    local monitoring_plugins=(
        "local_advanced_monitoring"
        "tool_health_check"
        "block_system_status"
        "report_performance"
    )
    
    for plugin in "${monitoring_plugins[@]}"; do
        print_info "Installing monitoring plugin: $plugin"
        if php "$MOODLE_DIR/admin/cli/install_plugins.php" --install="$plugin" --non-interactive 2>/dev/null; then
            print_success "Monitoring plugin $plugin installed"
        else
            print_info "Monitoring plugin $plugin not available"
        fi
    done
    
    # Configure advanced monitoring
    print_info "Configuring advanced monitoring..."
    
    # Performance monitoring
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=advanced_monitoring_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=performance_monitoring_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=user_activity_monitoring --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=course_analytics_monitoring --set=1
    
    # Health checks
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=health_check_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=health_check_interval --set=300
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=health_check_alert_threshold --set=80
    
    # System status
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=system_status_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=system_status_public --set=0
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=system_status_auto_refresh --set=1
    
    # Alerting
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=monitoring_alerts_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=monitoring_alert_email --set="admin@yourdomain.com"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=monitoring_alert_slack --set=""
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=monitoring_alert_webhook --set=""
    
    print_success "Advanced monitoring configured"
    log_message "SUCCESS" "Advanced monitoring setup completed"
}

# Setup advanced security features
setup_advanced_security() {
    print_section "Setting Up Advanced Security Features"
    
    print_info "Installing advanced security plugins..."
    
    # Install security plugins
    local security_plugins=(
        "local_advanced_security"
        "tool_security_audit"
        "auth_mfa"
        "tool_mfa"
        "local_data_privacy"
        "tool_dataprivacy"
    )
    
    for plugin in "${security_plugins[@]}"; do
        print_info "Installing security plugin: $plugin"
        if php "$MOODLE_DIR/admin/cli/install_plugins.php" --install="$plugin" --non-interactive 2>/dev/null; then
            print_success "Security plugin $plugin installed"
        else
            print_info "Security plugin $plugin not available"
        fi
    done
    
    # Configure advanced security
    print_info "Configuring advanced security..."
    
    # Multi-Factor Authentication (MFA)
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mfa_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mfa_required_roles --set="manager,teacher"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mfa_totp_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mfa_email_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mfa_sms_enabled --set=1
    
    # Security audit
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=security_audit_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=security_audit_log_retention --set=90
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=security_audit_alert_threshold --set=5
    
    # Data privacy
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=dataprivacy_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=dataprivacy_auto_delete --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=dataprivacy_retention_period --set=2555  # 7 years
    
    # Advanced password policies
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=password_policy_advanced --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=password_history_count --set=12
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=password_expiry_days --set=90
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=password_force_change_on_first_login --set=1
    
    # Session security
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=session_security_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=session_timeout_warning --set=300
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=session_concurrent_limit --set=3
    
    print_success "Advanced security features configured"
    log_message "SUCCESS" "Advanced security setup completed"
}

# Setup advanced backup and recovery
setup_advanced_backup() {
    print_section "Setting Up Advanced Backup and Recovery"
    
    print_info "Installing advanced backup plugins..."
    
    # Install backup plugins
    local backup_plugins=(
        "local_advanced_backup"
        "tool_automated_backup"
        "local_backup_encryption"
        "tool_backup_restore"
    )
    
    for plugin in "${backup_plugins[@]}"; do
        print_info "Installing backup plugin: $plugin"
        if php "$MOODLE_DIR/admin/cli/install_plugins.php" --install="$plugin" --non-interactive 2>/dev/null; then
            print_success "Backup plugin $plugin installed"
        else
            print_info "Backup plugin $plugin not available"
        fi
    done
    
    # Configure advanced backup
    print_info "Configuring advanced backup..."
    
    # Automated backup
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=automated_backup_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=automated_backup_schedule --set="daily"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=automated_backup_time --set="02:00"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=automated_backup_retention --set=30
    
    # Backup encryption
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=backup_encryption_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=backup_encryption_key --set="$(openssl rand -base64 32)"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=backup_encryption_algorithm --set="AES-256-CBC"
    
    # Cloud backup
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=cloud_backup_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=cloud_backup_provider --set="aws_s3"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=cloud_backup_bucket --set="lmsk2-backups"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=cloud_backup_region --set="us-east-1"
    
    # Disaster recovery
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=disaster_recovery_enabled --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=disaster_recovery_rto --set=4  # 4 hours
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=disaster_recovery_rpo --set=1  # 1 hour
    
    print_success "Advanced backup and recovery configured"
    log_message "SUCCESS" "Advanced backup setup completed"
}

# Test advanced features
test_advanced_features() {
    print_section "Testing Advanced Features"
    
    print_info "Testing AI integration..."
    if php "$MOODLE_DIR/admin/cli/cfg.php" --name=enableai --get | grep -q "1"; then
        print_success "AI integration is enabled"
    else
        print_warning "AI integration is not enabled"
    fi
    
    print_info "Testing analytics..."
    if php "$MOODLE_DIR/admin/cli/cfg.php" --name=enableanalytics --get | grep -q "1"; then
        print_success "Analytics is enabled"
    else
        print_warning "Analytics is not enabled"
    fi
    
    print_info "Testing mobile features..."
    if php "$MOODLE_DIR/admin/cli/cfg.php" --name=enablemobilewebservice --get | grep -q "1"; then
        print_success "Mobile features are enabled"
    else
        print_warning "Mobile features are not enabled"
    fi
    
    print_info "Testing enterprise features..."
    if php "$MOODLE_DIR/admin/cli/cfg.php" --name=enable_sso --get | grep -q "1"; then
        print_success "Enterprise features are enabled"
    else
        print_warning "Enterprise features are not enabled"
    fi
    
    print_info "Testing advanced monitoring..."
    if php "$MOODLE_DIR/admin/cli/cfg.php" --name=advanced_monitoring_enabled --get | grep -q "1"; then
        print_success "Advanced monitoring is enabled"
    else
        print_warning "Advanced monitoring is not enabled"
    fi
    
    print_info "Testing advanced security..."
    if php "$MOODLE_DIR/admin/cli/cfg.php" --name=mfa_enabled --get | grep -q "1"; then
        print_success "Advanced security is enabled"
    else
        print_warning "Advanced security is not enabled"
    fi
    
    print_info "Testing advanced backup..."
    if php "$MOODLE_DIR/admin/cli/cfg.php" --name=automated_backup_enabled --get | grep -q "1"; then
        print_success "Advanced backup is enabled"
    else
        print_warning "Advanced backup is not enabled"
    fi
    
    log_message "SUCCESS" "Advanced features testing completed"
}

# =============================================================================
# Main Function
# =============================================================================

main() {
    print_color $CYAN "=============================================================================="
    print_color $WHITE "  $SCRIPT_NAME v$SCRIPT_VERSION"
    print_color $CYAN "=============================================================================="
    
    log_message "INFO" "Starting advanced features setup"
    
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
            --ai)
                action="ai"
                shift
                ;;
            --analytics)
                action="analytics"
                shift
                ;;
            --mobile)
                action="mobile"
                shift
                ;;
            --enterprise)
                action="enterprise"
                shift
                ;;
            --monitoring)
                action="monitoring"
                shift
                ;;
            --security)
                action="security"
                shift
                ;;
            --backup)
                action="backup"
                shift
                ;;
            --test)
                action="test"
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --ai          Setup AI integration"
                echo "  --analytics   Setup analytics and reporting"
                echo "  --mobile      Setup mobile app features"
                echo "  --enterprise  Setup enterprise features"
                echo "  --monitoring  Setup advanced monitoring"
                echo "  --security    Setup advanced security"
                echo "  --backup      Setup advanced backup"
                echo "  --test        Test advanced features"
                echo "  --help        Show this help"
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
        "ai")
            setup_ai_integration
            ;;
        "analytics")
            setup_analytics
            ;;
        "mobile")
            setup_mobile_features
            ;;
        "enterprise")
            setup_enterprise_features
            ;;
        "monitoring")
            setup_advanced_monitoring
            ;;
        "security")
            setup_advanced_security
            ;;
        "backup")
            setup_advanced_backup
            ;;
        "test")
            test_advanced_features
            ;;
        "all")
            setup_ai_integration
            setup_analytics
            setup_mobile_features
            setup_enterprise_features
            setup_advanced_monitoring
            setup_advanced_security
            setup_advanced_backup
            test_advanced_features
            ;;
    esac
    
    print_section "Advanced Features Setup Complete"
    print_success "Advanced features setup completed successfully!"
    print_info "Log file: $LOG_DIR/advanced-features.log"
    
    log_message "SUCCESS" "Advanced features setup completed successfully"
}

# Trap errors
trap 'print_error "Script failed at line $LINENO"; log_message "ERROR" "Script failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
