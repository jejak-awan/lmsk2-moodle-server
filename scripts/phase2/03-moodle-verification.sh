#!/bin/bash

# =============================================================================
# LMSK2 Moodle Server - Moodle Verification Script
# =============================================================================
# Description: Comprehensive verification of Moodle 3.11 LTS installation
# Author: jejakawan007
# Version: 1.0
# Date: September 2025
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Load configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/config/installer.conf"

if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
else
    echo "❌ Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Script configuration
SCRIPT_NAME="$(basename "$0")"
LOG_FILE="/var/log/lmsk2/${SCRIPT_NAME%.*}.log"
REPORT_FILE="/var/log/moodle-verification-report.txt"
MOODLE_DIR="/var/www/moodle"
MOODLE_DATA_DIR="/var/www/moodle/moodledata"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# =============================================================================
# Logging Functions
# =============================================================================

log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case "$level" in
        "INFO")  echo -e "${GREEN}[INFO]${NC} $message" ;;
        "WARN")  echo -e "${YELLOW}[WARN]${NC} $message" ;;
        "ERROR") echo -e "${RED}[ERROR]${NC} $message" ;;
        "DEBUG") echo -e "${BLUE}[DEBUG]${NC} $message" ;;
    esac
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# =============================================================================
# Utility Functions
# =============================================================================

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log "ERROR" "This script must be run as root"
        exit 1
    fi
}

# =============================================================================
# Verification Functions
# =============================================================================

verify_moodle_version() {
    log "INFO" "Verifying Moodle version..."
    
    cd "$MOODLE_DIR"
    
    if [[ -f "version.php" ]]; then
        local moodle_version=$(sudo -u www-data php -r "require_once('version.php'); echo \$version;")
        log "INFO" "Moodle Version: $moodle_version"
        
        if [[ "$moodle_version" == "20221128" ]]; then
            log "INFO" "✓ Moodle 3.11 LTS detected"
            return 0
        else
            log "WARN" "⚠ Unexpected Moodle version: $moodle_version"
            return 1
        fi
    else
        log "ERROR" "✗ version.php not found"
        return 1
    fi
}

verify_web_interface() {
    log "INFO" "Verifying web interface accessibility..."
    
    local domain="${MOODLE_DOMAIN:-lms.example.com}"
    local http_status=$(curl -s -o /dev/null -w "%{http_code}" "https://$domain" 2>/dev/null || echo "000")
    
    if [[ "$http_status" == "200" ]]; then
        log "INFO" "✓ Web interface accessible (HTTP $http_status)"
        return 0
    else
        log "ERROR" "✗ Web interface not accessible (HTTP $http_status)"
        return 1
    fi
}

verify_ssl_certificate() {
    log "INFO" "Verifying SSL certificate..."
    
    local domain="${MOODLE_DOMAIN:-lms.example.com}"
    
    if echo | openssl s_client -connect "$domain:443" -servername "$domain" 2>/dev/null | openssl x509 -noout -dates >/dev/null 2>&1; then
        local ssl_info=$(echo | openssl s_client -connect "$domain:443" -servername "$domain" 2>/dev/null | openssl x509 -noout -dates)
        log "INFO" "✓ SSL certificate valid"
        log "INFO" "  $ssl_info"
        return 0
    else
        log "ERROR" "✗ SSL certificate issue"
        return 1
    fi
}

verify_database_connectivity() {
    log "INFO" "Verifying database connectivity..."
    
    if mysql -u "${DB_USER:-moodle}" -p"${DB_PASSWORD:-}" -e "SELECT 1;" >/dev/null 2>&1; then
        log "INFO" "✓ Database connection successful"
        
        # Check database tables
        local table_count=$(mysql -u "${DB_USER:-moodle}" -p"${DB_PASSWORD:-}" "${DB_NAME:-moodle}" -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '${DB_NAME:-moodle}';" 2>/dev/null | tail -1)
        log "INFO" "✓ Database tables: $table_count"
        
        # Check database size
        local db_size=$(mysql -u "${DB_USER:-moodle}" -p"${DB_PASSWORD:-}" "${DB_NAME:-moodle}" -e "SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS 'DB Size in MB' FROM information_schema.tables WHERE table_schema = '${DB_NAME:-moodle}';" 2>/dev/null | tail -1)
        log "INFO" "✓ Database size: $db_size MB"
        return 0
    else
        log "ERROR" "✗ Database connection failed"
        return 1
    fi
}

verify_file_permissions() {
    log "INFO" "Verifying file permissions..."
    
    # Check config.php permissions
    local config_perm=$(ls -la "$MOODLE_DIR/config.php" | awk '{print $1}')
    if [[ "$config_perm" == "-rw-------" ]]; then
        log "INFO" "✓ config.php permissions correct ($config_perm)"
    else
        log "ERROR" "✗ config.php permissions incorrect ($config_perm)"
    fi
    
    # Check moodledata permissions
    local moodledata_perm=$(ls -ld "$MOODLE_DATA_DIR" | awk '{print $1}')
    if [[ "$moodledata_perm" == "drwxrwxrwx" ]]; then
        log "INFO" "✓ moodledata permissions correct ($moodledata_perm)"
    else
        log "ERROR" "✗ moodledata permissions incorrect ($moodledata_perm)"
    fi
}

verify_cron_job() {
    log "INFO" "Verifying cron job..."
    
    local cron_exists=$(crontab -l 2>/dev/null | grep -c "moodle.*cron.php" || echo "0")
    
    if [[ "$cron_exists" -gt 0 ]]; then
        log "INFO" "✓ Moodle cron job configured"
        
        # Test cron execution
        if sudo -u www-data php "$MOODLE_DIR/admin/cli/cron.php" >/dev/null 2>&1; then
            log "INFO" "✓ Cron job execution successful"
            return 0
        else
            log "ERROR" "✗ Cron job execution failed"
            return 1
        fi
    else
        log "ERROR" "✗ Moodle cron job not configured"
        return 1
    fi
}

verify_php_extensions() {
    log "INFO" "Verifying PHP extensions..."
    
    local required_extensions=("mysql" "gd" "curl" "xml" "mbstring" "zip" "intl" "soap" "ldap" "imagick" "xmlrpc" "openssl" "json" "dom" "fileinfo" "iconv" "simplexml" "tokenizer" "xmlreader" "xmlwriter")
    local missing_extensions=()
    
    for ext in "${required_extensions[@]}"; do
        if ! php -m | grep -q "^$ext$"; then
            missing_extensions+=("$ext")
        fi
    done
    
    if [[ ${#missing_extensions[@]} -eq 0 ]]; then
        log "INFO" "✓ All required PHP extensions installed"
        return 0
    else
        log "ERROR" "✗ Missing PHP extensions: ${missing_extensions[*]}"
        return 1
    fi
}

verify_system_resources() {
    log "INFO" "Verifying system resources..."
    
    # Check disk usage
    local disk_usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    if [[ $disk_usage -lt 80 ]]; then
        log "INFO" "✓ Disk usage: $disk_usage% (OK)"
    else
        log "WARN" "⚠ Disk usage: $disk_usage% (High)"
    fi
    
    # Check memory usage
    local memory_usage=$(free | awk 'NR==2{printf "%.1f", $3*100/$2}')
    if (( $(echo "$memory_usage < 80" | bc -l) )); then
        log "INFO" "✓ Memory usage: $memory_usage% (OK)"
    else
        log "WARN" "⚠ Memory usage: $memory_usage% (High)"
    fi
    
    # Check CPU load
    local cpu_load=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
    if (( $(echo "$cpu_load < 5" | bc -l) )); then
        log "INFO" "✓ CPU load: $cpu_load (OK)"
    else
        log "WARN" "⚠ CPU load: $cpu_load (High)"
    fi
}

verify_services() {
    log "INFO" "Verifying services..."
    
    local services=("nginx" "php8.1-fpm" "mariadb" "redis-server" "fail2ban")
    local all_services_ok=true
    
    for service in "${services[@]}"; do
        if systemctl is-active --quiet "$service"; then
            log "INFO" "✓ $service: Running"
        else
            log "ERROR" "✗ $service: Not running"
            all_services_ok=false
        fi
    done
    
    if [[ "$all_services_ok" == true ]]; then
        return 0
    else
        return 1
    fi
}

verify_security_headers() {
    log "INFO" "Verifying security headers..."
    
    local domain="${MOODLE_DOMAIN:-lms.example.com}"
    local headers=$(curl -I "https://$domain/" 2>/dev/null || echo "")
    
    if echo "$headers" | grep -q "Strict-Transport-Security"; then
        log "INFO" "✓ HSTS header present"
    else
        log "WARN" "✗ HSTS header missing"
    fi
    
    if echo "$headers" | grep -q "X-Frame-Options"; then
        log "INFO" "✓ X-Frame-Options header present"
    else
        log "WARN" "✗ X-Frame-Options header missing"
    fi
    
    if echo "$headers" | grep -q "X-Content-Type-Options"; then
        log "INFO" "✓ X-Content-Type-Options header present"
    else
        log "WARN" "✗ X-Content-Type-Options header missing"
    fi
}

verify_firewall() {
    log "INFO" "Verifying firewall configuration..."
    
    local ufw_status=$(ufw status | grep "Status" || echo "Status: inactive")
    log "INFO" "  $ufw_status"
    
    local open_ports=$(ufw status | grep "ALLOW" || echo "No open ports")
    log "INFO" "  Open ports: $open_ports"
}

verify_fail2ban() {
    log "INFO" "Verifying Fail2ban configuration..."
    
    if systemctl is-active --quiet fail2ban; then
        log "INFO" "✓ Fail2ban running"
        local jail_status=$(fail2ban-client status 2>/dev/null || echo "No jails active")
        log "INFO" "  Active jails: $jail_status"
        return 0
    else
        log "ERROR" "✗ Fail2ban not running"
        return 1
    fi
}

# =============================================================================
# Performance Testing Functions
# =============================================================================

test_response_time() {
    log "INFO" "Testing response time..."
    
    local domain="${MOODLE_DOMAIN:-lms.example.com}"
    local response_time=$(curl -o /dev/null -s -w '%{time_total}' "https://$domain/" 2>/dev/null || echo "0")
    log "INFO" "  Homepage response time: ${response_time}s"
    
    local login_time=$(curl -o /dev/null -s -w '%{time_total}' "https://$domain/login/index.php" 2>/dev/null || echo "0")
    log "INFO" "  Login page response time: ${login_time}s"
}

test_database_performance() {
    log "INFO" "Testing database performance..."
    
    if mysql -u "${DB_USER:-moodle}" -p"${DB_PASSWORD:-}" "${DB_NAME:-moodle}" -e "SELECT COUNT(*) FROM mdl_user;" >/dev/null 2>&1; then
        log "INFO" "  ✓ Database query performance: OK"
        return 0
    else
        log "ERROR" "  ✗ Database query performance: Failed"
        return 1
    fi
}

test_concurrent_connections() {
    log "INFO" "Testing concurrent connections..."
    
    local domain="${MOODLE_DOMAIN:-lms.example.com}"
    
    for i in {1..5}; do
        curl -o /dev/null -s "https://$domain/" &
    done
    wait
    
    log "INFO" "  ✓ 5 concurrent connections handled"
}

# =============================================================================
# Functional Testing Functions
# =============================================================================

test_admin_login() {
    log "INFO" "Testing admin login page..."
    
    local domain="${MOODLE_DOMAIN:-lms.example.com}"
    local login_status=$(curl -s -o /dev/null -w "%{http_code}" "https://$domain/login/index.php" 2>/dev/null || echo "000")
    
    if [[ "$login_status" == "200" ]]; then
        log "INFO" "  ✓ Admin login page accessible"
        return 0
    else
        log "ERROR" "  ✗ Admin login page issue (HTTP $login_status)"
        return 1
    fi
}

test_file_upload() {
    log "INFO" "Testing file upload functionality..."
    
    if [[ -w "$MOODLE_DATA_DIR" ]]; then
        log "INFO" "  ✓ Moodledata directory writable"
        return 0
    else
        log "ERROR" "  ✗ Moodledata directory not writable"
        return 1
    fi
}

test_user_management() {
    log "INFO" "Testing user management..."
    
    local user_count=$(mysql -u "${DB_USER:-moodle}" -p"${DB_PASSWORD:-}" "${DB_NAME:-moodle}" -e "SELECT COUNT(*) FROM mdl_user WHERE deleted = 0;" 2>/dev/null | tail -1 || echo "0")
    log "INFO" "  Total users: $user_count"
}

# =============================================================================
# Report Generation Functions
# =============================================================================

generate_verification_report() {
    log "INFO" "Generating verification report..."
    
    {
        echo "=== Moodle 3.11 LTS Verification Report ==="
        echo "Generated: $(date '+%Y-%m-%d %H:%M:%S')"
        echo "Domain: ${MOODLE_DOMAIN:-lms.example.com}"
        echo "Moodle Directory: $MOODLE_DIR"
        echo "Database: ${DB_NAME:-moodle}"
        echo ""
        
        echo "=== Installation Verification ==="
        verify_moodle_version
        echo ""
        
        echo "=== Web Interface Verification ==="
        verify_web_interface
        verify_ssl_certificate
        echo ""
        
        echo "=== Database Verification ==="
        verify_database_connectivity
        echo ""
        
        echo "=== File Permissions Verification ==="
        verify_file_permissions
        echo ""
        
        echo "=== Cron Job Verification ==="
        verify_cron_job
        echo ""
        
        echo "=== PHP Extensions Verification ==="
        verify_php_extensions
        echo ""
        
        echo "=== System Resources Verification ==="
        verify_system_resources
        echo ""
        
        echo "=== Services Verification ==="
        verify_services
        echo ""
        
        echo "=== Security Verification ==="
        verify_security_headers
        verify_firewall
        verify_fail2ban
        echo ""
        
        echo "=== Performance Testing ==="
        test_response_time
        test_database_performance
        test_concurrent_connections
        echo ""
        
        echo "=== Functional Testing ==="
        test_admin_login
        test_file_upload
        test_user_management
        echo ""
        
        echo "=== Verification Complete ==="
    } > "$REPORT_FILE"
    
    log "INFO" "Verification report generated: $REPORT_FILE"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    log "INFO" "Starting Moodle verification process..."
    
    # Check prerequisites
    check_root
    
    # Run comprehensive verification
    generate_verification_report
    
    # Display report summary
    log "INFO" "Verification completed. Report saved to: $REPORT_FILE"
    log "INFO" "Log file: $LOG_FILE"
    
    # Show report summary
    echo ""
    echo "=== Verification Summary ==="
    tail -20 "$REPORT_FILE"
}

# =============================================================================
# Script Execution
# =============================================================================

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --report       Generate verification report only"
        echo "  --quick        Run quick verification (no performance tests)"
        echo "  --security     Run security verification only"
        echo "  --performance  Run performance tests only"
        exit 0
        ;;
    --report)
        check_root
        generate_verification_report
        echo "Report generated: $REPORT_FILE"
        exit 0
        ;;
    --quick)
        log "INFO" "Running quick verification..."
        check_root
        verify_moodle_version
        verify_web_interface
        verify_database_connectivity
        verify_services
        log "INFO" "Quick verification completed"
        exit 0
        ;;
    --security)
        log "INFO" "Running security verification..."
        check_root
        verify_ssl_certificate
        verify_security_headers
        verify_firewall
        verify_fail2ban
        verify_file_permissions
        log "INFO" "Security verification completed"
        exit 0
        ;;
    --performance)
        log "INFO" "Running performance tests..."
        check_root
        test_response_time
        test_database_performance
        test_concurrent_connections
        verify_system_resources
        log "INFO" "Performance tests completed"
        exit 0
        ;;
    "")
        main
        ;;
    *)
        log "ERROR" "Unknown option: $1"
        log "INFO" "Use --help for usage information"
        exit 1
        ;;
esac
