#!/bin/bash

# =============================================================================
# System Verification Script
# =============================================================================
# Version: 1.0
# Author: jejakawan007
# Description: System verification untuk LMSK2-Moodle-Server
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Script information
SCRIPT_NAME="System Verification"
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

# Check service status
check_service() {
    local service=$1
    if systemctl is-active --quiet "$service"; then
        print_success "$service: Running"
        return 0
    else
        print_error "$service: Not running"
        return 1
    fi
}

# Check port status
check_port() {
    local port=$1
    local service=$2
    if netstat -tlnp 2>/dev/null | grep -q ":$port "; then
        print_success "Port $port ($service): Open"
        return 0
    else
        print_error "Port $port ($service): Closed"
        return 1
    fi
}

# =============================================================================
# Verification Functions
# =============================================================================

# System information
system_information() {
    print_section "System Information"
    
    print_info "Operating System:"
    if [[ -f /etc/os-release ]]; then
        source /etc/os-release
        echo "  OS: $PRETTY_NAME"
        echo "  Version: $VERSION_ID"
        echo "  Codename: $VERSION_CODENAME"
    else
        print_warning "Cannot determine OS information"
    fi
    
    print_info "Kernel Information:"
    echo "  Kernel: $(uname -r)"
    echo "  Architecture: $(uname -m)"
    echo "  Hostname: $(hostname)"
    
    print_info "System Uptime:"
    echo "  Uptime: $(uptime -p)"
    echo "  Load Average: $(uptime | awk -F'load average:' '{print $2}')"
    
    print_info "Date and Time:"
    echo "  Current Time: $(date)"
    echo "  Timezone: $(timedatectl show --property=Timezone --value 2>/dev/null || echo 'Unknown')"
    echo "  NTP Status: $(timedatectl show --property=NTPSynchronized --value 2>/dev/null || echo 'Unknown')"
}

# Hardware information
hardware_information() {
    print_section "Hardware Information"
    
    print_info "CPU Information:"
    echo "  CPU Model: $(lscpu | grep 'Model name' | cut -d: -f2 | xargs)"
    echo "  CPU Cores: $(nproc)"
    echo "  CPU Architecture: $(lscpu | grep 'Architecture' | cut -d: -f2 | xargs)"
    
    print_info "Memory Information:"
    free -h | while read line; do
        echo "  $line"
    done
    
    print_info "Disk Information:"
    df -h | while read line; do
        echo "  $line"
    done
    
    print_info "Swap Information:"
    if command -v swapon &> /dev/null; then
        swapon --show 2>/dev/null || echo "  No swap configured"
    else
        echo "  Swap information not available"
    fi
}

# Network configuration
network_configuration() {
    print_section "Network Configuration"
    
    print_info "Network Interfaces:"
    ip addr show | grep -E "(inet |UP)" | while read line; do
        echo "  $line"
    done
    
    print_info "Routing Table:"
    ip route show | while read line; do
        echo "  $line"
    done
    
    print_info "DNS Configuration:"
    if [[ -f /etc/resolv.conf ]]; then
        grep nameserver /etc/resolv.conf | while read line; do
            echo "  $line"
        done
    fi
    
    print_info "Network Connectivity:"
    if ping -c 1 8.8.8.8 >/dev/null 2>&1; then
        print_success "Internet connectivity: OK"
    else
        print_error "Internet connectivity: Failed"
    fi
    
    if nslookup google.com >/dev/null 2>&1; then
        print_success "DNS resolution: OK"
    else
        print_error "DNS resolution: Failed"
    fi
}

# Service status
service_status() {
    print_section "Service Status"
    
    local services=("nginx" "php8.1-fpm" "mariadb" "redis-server" "fail2ban")
    local failed_services=()
    
    for service in "${services[@]}"; do
        if ! check_service "$service"; then
            failed_services+=("$service")
        fi
    done
    
    if [[ ${#failed_services[@]} -gt 0 ]]; then
        print_warning "Failed services: ${failed_services[*]}"
        return 1
    else
        print_success "All services are running"
        return 0
    fi
}

# Port status
port_status() {
    print_section "Port Status"
    
    local ports=("80:HTTP" "443:HTTPS" "3306:MySQL" "6379:Redis" "22:SSH")
    local failed_ports=()
    
    for port_info in "${ports[@]}"; do
        IFS=':' read -r port service <<< "$port_info"
        if ! check_port "$port" "$service"; then
            failed_ports+=("$port")
        fi
    done
    
    if [[ ${#failed_ports[@]} -gt 0 ]]; then
        print_warning "Failed ports: ${failed_ports[*]}"
        return 1
    else
        print_success "All required ports are open"
        return 0
    fi
}

# Firewall status
firewall_status() {
    print_section "Firewall Status"
    
    if command -v ufw &> /dev/null; then
        print_info "UFW Firewall Status:"
        ufw status | while read line; do
            echo "  $line"
        done
    else
        print_warning "UFW firewall not installed"
    fi
    
    if command -v iptables &> /dev/null; then
        print_info "IPTables Rules:"
        iptables -L -n | head -10 | while read line; do
            echo "  $line"
        done
    fi
}

# SSL certificate status
ssl_certificate_status() {
    print_section "SSL Certificate Status"
    
    local domain="${LMSK2_DOMAIN:-lms-server.local}"
    
    if [[ -d "/etc/letsencrypt/live/$domain" ]]; then
        print_success "SSL certificate found for $domain"
        
        local cert_file="/etc/letsencrypt/live/$domain/fullchain.pem"
        if [[ -f "$cert_file" ]]; then
            print_info "Certificate Information:"
            openssl x509 -in "$cert_file" -text -noout | grep -E "(Subject:|Not After:)" | while read line; do
                echo "  $line"
            done
        fi
    else
        print_warning "SSL certificate not found for $domain"
    fi
}

# Database connectivity
database_connectivity() {
    print_section "Database Connectivity"
    
    local db_password="${LMSK2_DB_PASSWORD:-}"
    
    if [[ -n "$db_password" ]]; then
        print_info "Testing MariaDB connection..."
        if mysql -u root -p"$db_password" -e "SELECT 1;" >/dev/null 2>&1; then
            print_success "MariaDB root connection: OK"
        else
            print_error "MariaDB root connection: Failed"
        fi
        
        if mysql -u moodle -p"$db_password" -e "SELECT 1;" >/dev/null 2>&1; then
            print_success "MariaDB moodle user connection: OK"
        else
            print_error "MariaDB moodle user connection: Failed"
        fi
        
        print_info "Database Information:"
        mysql -u root -p"$db_password" -e "SHOW DATABASES;" 2>/dev/null | while read line; do
            echo "  $line"
        done
    else
        print_warning "Database password not provided, skipping connectivity test"
    fi
}

# Redis connectivity
redis_connectivity() {
    print_section "Redis Connectivity"
    
    print_info "Testing Redis connection..."
    if redis-cli ping >/dev/null 2>&1; then
        print_success "Redis connection: OK"
        
        print_info "Redis Information:"
        redis-cli info server | grep -E "(redis_version|uptime_in_seconds|connected_clients)" | while read line; do
            echo "  $line"
        done
    else
        print_error "Redis connection: Failed"
    fi
}

# PHP configuration
php_configuration() {
    print_section "PHP Configuration"
    
    print_info "PHP Version:"
    php -v | head -1
    
    print_info "PHP Extensions:"
    local required_extensions=("mysql" "gd" "curl" "xml" "mbstring" "zip" "intl" "soap" "ldap" "imagick" "redis")
    local missing_extensions=()
    
    for ext in "${required_extensions[@]}"; do
        if php -m | grep -q "^$ext$"; then
            print_success "Extension $ext: Installed"
        else
            print_error "Extension $ext: Missing"
            missing_extensions+=("$ext")
        fi
    done
    
    if [[ ${#missing_extensions[@]} -gt 0 ]]; then
        print_warning "Missing PHP extensions: ${missing_extensions[*]}"
    fi
    
    print_info "PHP Configuration:"
    php -i | grep -E "(memory_limit|upload_max_filesize|post_max_size|opcache.enable)" | while read line; do
        echo "  $line"
    done
}

# File permissions
file_permissions() {
    print_section "File Permissions"
    
    local directories=("/var/www/moodle" "/var/www/moodle/moodledata" "/var/log/lmsk2" "/backup/moodle")
    
    for dir in "${directories[@]}"; do
        if [[ -d "$dir" ]]; then
            print_info "Directory: $dir"
            ls -la "$dir" | head -5 | while read line; do
                echo "  $line"
            done
        else
            print_warning "Directory not found: $dir"
        fi
    done
}

# Log files
log_files() {
    print_section "Log Files"
    
    local log_files=(
        "/var/log/lmsk2/installer.log"
        "/var/log/lmsk2/phase1.log"
        "/var/log/nginx/error.log"
        "/var/log/php8.1-fpm.log"
        "/var/log/mysql/error.log"
        "/var/log/redis/redis-server.log"
    )
    
    for log_file in "${log_files[@]}"; do
        if [[ -f "$log_file" ]]; then
            print_info "Log file: $log_file"
            echo "  Size: $(du -h "$log_file" | cut -f1)"
            echo "  Last modified: $(stat -c %y "$log_file")"
        else
            print_warning "Log file not found: $log_file"
        fi
    done
}

# Performance metrics
performance_metrics() {
    print_section "Performance Metrics"
    
    print_info "CPU Usage:"
    top -bn1 | grep "Cpu(s)" | awk '{print "  CPU Usage: " $2}'
    
    print_info "Memory Usage:"
    free | awk 'NR==2{printf "  Memory Usage: %.2f%%\n", $3*100/$2}'
    
    print_info "Disk Usage:"
    df -h / | awk 'NR==2{print "  Root partition: " $5 " used"}'
    
    print_info "Load Average:"
    uptime | awk -F'load average:' '{print "  Load Average:" $2}'
    
    print_info "Active Connections:"
    netstat -an | grep ESTABLISHED | wc -l | awk '{print "  Active connections: " $1}'
}

# Security check
security_check() {
    print_section "Security Check"
    
    print_info "SSH Configuration:"
    if [[ -f /etc/ssh/sshd_config ]]; then
        grep -E "(PermitRootLogin|PasswordAuthentication|Port)" /etc/ssh/sshd_config | while read line; do
            echo "  $line"
        done
    fi
    
    print_info "Fail2ban Status:"
    if command -v fail2ban-client &> /dev/null; then
        fail2ban-client status 2>/dev/null | while read line; do
            echo "  $line"
        done
    else
        print_warning "Fail2ban not installed"
    fi
    
    print_info "Recent Failed Login Attempts:"
    if [[ -f /var/log/auth.log ]]; then
        grep "Failed password" /var/log/auth.log | tail -5 | while read line; do
            echo "  $line"
        done
    fi
}

# Generate report
generate_report() {
    print_section "Verification Report"
    
    local report_file="/var/log/lmsk2/verification-report-$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "LMSK2 Moodle Server Verification Report"
        echo "Generated: $(date)"
        echo "Hostname: $(hostname)"
        echo "OS: $(lsb_release -d | cut -f2)"
        echo "Kernel: $(uname -r)"
        echo
        echo "=============================================================================="
        echo
        
        # System information
        echo "SYSTEM INFORMATION:"
        systemctl status nginx --no-pager -l | head -5
        systemctl status php8.1-fpm --no-pager -l | head -5
        systemctl status mariadb --no-pager -l | head -5
        systemctl status redis-server --no-pager -l | head -5
        echo
        
        # Performance metrics
        echo "PERFORMANCE METRICS:"
        free -h
        df -h /
        uptime
        echo
        
        # Network status
        echo "NETWORK STATUS:"
        ip addr show | grep -E "(inet |UP)"
        echo
        
    } > "$report_file"
    
    print_success "Verification report generated: $report_file"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    print_color $CYAN "=============================================================================="
    print_color $WHITE "  $SCRIPT_NAME v$SCRIPT_VERSION"
    print_color $CYAN "=============================================================================="
    
    # Run all verification functions
    system_information
    hardware_information
    network_configuration
    service_status
    port_status
    firewall_status
    ssl_certificate_status
    database_connectivity
    redis_connectivity
    php_configuration
    file_permissions
    log_files
    performance_metrics
    security_check
    generate_report
    
    print_section "Verification Complete"
    print_success "System verification completed successfully!"
    print_info "Check the verification report for detailed information"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --full-check)
            # Run all checks (default behavior)
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --full-check    Run all verification checks (default)"
            echo "  --help          Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main function
main "$@"
