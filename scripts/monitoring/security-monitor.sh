#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: Security Monitor Script
# =============================================================================
# Description: Security monitoring and threat detection for LMSK2-Moodle-Server
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
SCRIPT_NAME="LMSK2 Security Monitor"
SCRIPT_VERSION="1.0"
LOG_FILE="/var/log/lmsk2-monitoring/security-monitor.log"
CONFIG_DIR="/opt/lmsk2-moodle-server/scripts/config"
ALERT_EMAIL="admin@localhost"

# Security thresholds
FAILED_LOGIN_THRESHOLD=10
SUSPICIOUS_ACTIVITY_THRESHOLD=5
FILE_INTEGRITY_THRESHOLD=1
PORT_SCAN_THRESHOLD=5

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
        "CRITICAL")
            echo -e "${RED}[CRITICAL]${NC} $message" | tee -a "$LOG_FILE"
            ;;
        "DEBUG")
            echo -e "${BLUE}[DEBUG]${NC} $message" | tee -a "$LOG_FILE"
            ;;
    esac
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# Send security alert
send_security_alert() {
    local level="$1"
    local component="$2"
    local message="$3"
    local details="$4"
    
    local subject="LMSK2 Security Alert: $component - $level"
    local body="Security Component: $component
Alert Level: $level
Message: $message
Details: $details
Timestamp: $(date)
Server: $(hostname)
IP Address: $(hostname -I | awk '{print $1}')"

    echo "$body" | mail -s "$subject" "$ALERT_EMAIL"
    log "INFO" "Security alert sent: $subject"
}

# =============================================================================
# Security Monitoring Functions
# =============================================================================

# Monitor failed login attempts
monitor_failed_logins() {
    log "INFO" "Monitoring failed login attempts..."
    
    local failed_logins=0
    local suspicious_ips=()
    
    # Check auth.log for failed password attempts
    if [ -f "/var/log/auth.log" ]; then
        local recent_failed=$(tail -n 1000 /var/log/auth.log | grep "Failed password" | grep "$(date '+%b %d')" | wc -l)
        failed_logins=$recent_failed
        
        # Get suspicious IPs
        suspicious_ips=($(tail -n 1000 /var/log/auth.log | grep "Failed password" | grep "$(date '+%b %d')" | awk '{print $11}' | sort | uniq -c | sort -nr | head -5 | awk '{print $2}'))
        
        log "INFO" "Failed login attempts today: $failed_logins"
        
        if [ "$failed_logins" -gt "$FAILED_LOGIN_THRESHOLD" ]; then
            log "WARN" "High number of failed login attempts: $failed_logins"
            send_security_alert "WARNING" "Authentication" "High number of failed login attempts: $failed_logins" "Suspicious IPs: ${suspicious_ips[*]}"
        fi
        
        # Check for brute force attempts
        for ip in "${suspicious_ips[@]}"; do
            if [ -n "$ip" ] && [ "$ip" != "for" ]; then
                local ip_attempts=$(tail -n 1000 /var/log/auth.log | grep "Failed password" | grep "$ip" | wc -l)
                if [ "$ip_attempts" -gt 5 ]; then
                    log "WARN" "Potential brute force attack from IP: $ip ($ip_attempts attempts)"
                    send_security_alert "CRITICAL" "Brute Force" "Potential brute force attack from IP: $ip" "Attempts: $ip_attempts"
                fi
            fi
        done
    else
        log "WARN" "Auth log file not found: /var/log/auth.log"
    fi
    
    # Check for SSH attacks
    if [ -f "/var/log/auth.log" ]; then
        local ssh_attacks=$(tail -n 1000 /var/log/auth.log | grep "Invalid user" | grep "$(date '+%b %d')" | wc -l)
        if [ "$ssh_attacks" -gt 5 ]; then
            log "WARN" "SSH attacks detected: $ssh_attacks"
            send_security_alert "WARNING" "SSH" "SSH attacks detected: $ssh_attacks" "Invalid user attempts"
        fi
    fi
    
    return $failed_logins
}

# Monitor suspicious file activities
monitor_file_activities() {
    log "INFO" "Monitoring suspicious file activities..."
    
    local suspicious_activities=0
    
    # Check for unauthorized file modifications
    if [ -f "/var/log/auth.log" ]; then
        local file_modifications=$(tail -n 1000 /var/log/auth.log | grep "sudo" | grep "vi\|nano\|vim\|emacs" | wc -l)
        if [ "$file_modifications" -gt 10 ]; then
            log "WARN" "High number of file modifications: $file_modifications"
            suspicious_activities=$((suspicious_activities + 1))
        fi
    fi
    
    # Check for suspicious file permissions
    local suspicious_files=$(find /var/www -type f -perm -002 2>/dev/null | wc -l)
    if [ "$suspicious_files" -gt 0 ]; then
        log "WARN" "Files with world-writable permissions found: $suspicious_files"
        suspicious_activities=$((suspicious_activities + 1))
    fi
    
    # Check for suspicious file ownership
    local root_owned_files=$(find /var/www -type f -user root 2>/dev/null | wc -l)
    if [ "$root_owned_files" -gt 10 ]; then
        log "WARN" "High number of root-owned files in web directory: $root_owned_files"
        suspicious_activities=$((suspicious_activities + 1))
    fi
    
    # Check for recently modified critical files
    local critical_files=("/etc/passwd" "/etc/shadow" "/etc/sudoers" "/etc/ssh/sshd_config")
    for file in "${critical_files[@]}"; do
        if [ -f "$file" ]; then
            local file_age=$(($(date +%s) - $(stat -c %Y "$file")))
            if [ "$file_age" -lt 3600 ]; then  # Modified within last hour
                log "WARN" "Critical file recently modified: $file"
                suspicious_activities=$((suspicious_activities + 1))
            fi
        fi
    done
    
    if [ "$suspicious_activities" -gt "$SUSPICIOUS_ACTIVITY_THRESHOLD" ]; then
        send_security_alert "WARNING" "File System" "Suspicious file activities detected: $suspicious_activities" "Multiple file system anomalies"
    fi
    
    return $suspicious_activities
}

# Monitor network security
monitor_network_security() {
    log "INFO" "Monitoring network security..."
    
    local network_threats=0
    
    # Check for port scans
    if [ -f "/var/log/auth.log" ]; then
        local port_scans=$(tail -n 1000 /var/log/auth.log | grep "Connection refused" | wc -l)
        if [ "$port_scans" -gt "$PORT_SCAN_THRESHOLD" ]; then
            log "WARN" "Potential port scan detected: $port_scans connection refused"
            network_threats=$((network_threats + 1))
        fi
    fi
    
    # Check for suspicious network connections
    local suspicious_connections=$(ss -tuln | grep -E ":22|:23|:21|:25|:53|:80|:443|:993|:995" | wc -l)
    log "INFO" "Open network ports: $suspicious_connections"
    
    # Check for unusual network traffic
    local established_connections=$(ss -tuln | grep ESTAB | wc -l)
    if [ "$established_connections" -gt 1000 ]; then
        log "WARN" "High number of established connections: $established_connections"
        network_threats=$((network_threats + 1))
    fi
    
    # Check firewall status
    if command -v ufw >/dev/null 2>&1; then
        local firewall_status=$(ufw status | grep "Status:" | awk '{print $2}')
        if [ "$firewall_status" != "active" ]; then
            log "WARN" "Firewall is not active: $firewall_status"
            network_threats=$((network_threats + 1))
        fi
    fi
    
    # Check for listening services
    local listening_services=$(ss -tuln | grep LISTEN | wc -l)
    log "INFO" "Listening services: $listening_services"
    
    if [ "$network_threats" -gt 0 ]; then
        send_security_alert "WARNING" "Network" "Network security threats detected: $network_threats" "Multiple network anomalies"
    fi
    
    return $network_threats
}

# Monitor system integrity
monitor_system_integrity() {
    log "INFO" "Monitoring system integrity..."
    
    local integrity_issues=0
    
    # Check for unauthorized user accounts
    local suspicious_users=$(awk -F: '$3 >= 1000 && $3 < 65534 {print $1}' /etc/passwd | grep -v "www-data\|moodle\|backup" | wc -l)
    if [ "$suspicious_users" -gt 5 ]; then
        log "WARN" "High number of user accounts: $suspicious_users"
        integrity_issues=$((integrity_issues + 1))
    fi
    
    # Check for users with shell access
    local shell_users=$(awk -F: '$7 != "/bin/false" && $7 != "/usr/sbin/nologin" {print $1}' /etc/passwd | grep -v "root\|www-data" | wc -l)
    if [ "$shell_users" -gt 3 ]; then
        log "WARN" "Users with shell access: $shell_users"
        integrity_issues=$((integrity_issues + 1))
    fi
    
    # Check for sudo privileges
    local sudo_users=$(grep -v "^#" /etc/sudoers | grep -v "^$" | grep -v "root" | wc -l)
    if [ "$sudo_users" -gt 2 ]; then
        log "WARN" "Users with sudo privileges: $sudo_users"
        integrity_issues=$((integrity_issues + 1))
    fi
    
    # Check for unauthorized cron jobs
    local user_cron_jobs=$(find /var/spool/cron -type f 2>/dev/null | wc -l)
    if [ "$user_cron_jobs" -gt 3 ]; then
        log "WARN" "User cron jobs found: $user_cron_jobs"
        integrity_issues=$((integrity_issues + 1))
    fi
    
    # Check for unauthorized services
    local enabled_services=$(systemctl list-unit-files --state=enabled | grep -v "static\|disabled" | wc -l)
    log "INFO" "Enabled services: $enabled_services"
    
    if [ "$integrity_issues" -gt "$FILE_INTEGRITY_THRESHOLD" ]; then
        send_security_alert "WARNING" "System Integrity" "System integrity issues detected: $integrity_issues" "Multiple system anomalies"
    fi
    
    return $integrity_issues
}

# Monitor web application security
monitor_web_security() {
    log "INFO" "Monitoring web application security..."
    
    local web_threats=0
    
    # Check for PHP errors that might indicate attacks
    if [ -f "/var/log/php8.1-fpm.log" ]; then
        local php_errors=$(tail -n 1000 /var/log/php8.1-fpm.log | grep -i "error\|warning" | wc -l)
        if [ "$php_errors" -gt 50 ]; then
            log "WARN" "High number of PHP errors: $php_errors"
            web_threats=$((web_threats + 1))
        fi
    fi
    
    # Check for suspicious web requests
    if [ -f "/var/log/nginx/access.log" ]; then
        local suspicious_requests=$(tail -n 1000 /var/log/nginx/access.log | grep -E "(\.\./|\.\.\\|%2e%2e|%2f|%5c)" | wc -l)
        if [ "$suspicious_requests" -gt 10 ]; then
            log "WARN" "Suspicious web requests detected: $suspicious_requests"
            web_threats=$((web_threats + 1))
        fi
        
        # Check for SQL injection attempts
        local sql_injection_attempts=$(tail -n 1000 /var/log/nginx/access.log | grep -i "union\|select\|insert\|delete\|drop\|script" | wc -l)
        if [ "$sql_injection_attempts" -gt 5 ]; then
            log "WARN" "Potential SQL injection attempts: $sql_injection_attempts"
            web_threats=$((web_threats + 1))
        fi
        
        # Check for XSS attempts
        local xss_attempts=$(tail -n 1000 /var/log/nginx/access.log | grep -i "script\|javascript\|onload\|onerror" | wc -l)
        if [ "$xss_attempts" -gt 5 ]; then
            log "WARN" "Potential XSS attempts: $xss_attempts"
            web_threats=$((web_threats + 1))
        fi
    fi
    
    # Check for file upload attempts
    if [ -f "/var/log/nginx/access.log" ]; then
        local upload_attempts=$(tail -n 1000 /var/log/nginx/access.log | grep -i "upload\|file" | wc -l)
        if [ "$upload_attempts" -gt 20 ]; then
            log "WARN" "High number of file upload attempts: $upload_attempts"
            web_threats=$((web_threats + 1))
        fi
    fi
    
    if [ "$web_threats" -gt 0 ]; then
        send_security_alert "WARNING" "Web Application" "Web security threats detected: $web_threats" "Multiple web application anomalies"
    fi
    
    return $web_threats
}

# Monitor database security
monitor_database_security() {
    log "INFO" "Monitoring database security..."
    
    local db_threats=0
    
    if command -v mysql >/dev/null 2>&1; then
        # Check for failed database connections
        local failed_db_connections=$(mysql -e "SHOW GLOBAL STATUS LIKE 'Connection_errors_max_connections';" 2>/dev/null | awk '{print $2}')
        if [ -n "$failed_db_connections" ] && [ "$failed_db_connections" -gt 10 ]; then
            log "WARN" "High number of failed database connections: $failed_db_connections"
            db_threats=$((db_threats + 1))
        fi
        
        # Check for slow queries
        local slow_queries=$(mysql -e "SHOW GLOBAL STATUS LIKE 'Slow_queries';" 2>/dev/null | awk '{print $2}')
        if [ -n "$slow_queries" ] && [ "$slow_queries" -gt 100 ]; then
            log "WARN" "High number of slow queries: $slow_queries"
            db_threats=$((db_threats + 1))
        fi
        
        # Check for database users
        local db_users=$(mysql -e "SELECT COUNT(*) FROM mysql.user;" 2>/dev/null | tail -1)
        if [ -n "$db_users" ] && [ "$db_users" -gt 10 ]; then
            log "WARN" "High number of database users: $db_users"
            db_threats=$((db_threats + 1))
        fi
    else
        log "WARN" "MySQL client not found, skipping database security monitoring"
    fi
    
    if [ "$db_threats" -gt 0 ]; then
        send_security_alert "WARNING" "Database" "Database security threats detected: $db_threats" "Multiple database anomalies"
    fi
    
    return $db_threats
}

# Monitor SSL/TLS security
monitor_ssl_security() {
    log "INFO" "Monitoring SSL/TLS security..."
    
    local ssl_issues=0
    
    # Check SSL certificate expiration
    if command -v openssl >/dev/null 2>&1; then
        local ssl_cert=$(echo | openssl s_client -servername localhost -connect localhost:443 2>/dev/null | openssl x509 -noout -dates 2>/dev/null)
        if [ -n "$ssl_cert" ]; then
            local cert_expiry=$(echo "$ssl_cert" | grep "notAfter" | cut -d= -f2)
            local cert_expiry_epoch=$(date -d "$cert_expiry" +%s)
            local current_epoch=$(date +%s)
            local days_until_expiry=$(( (cert_expiry_epoch - current_epoch) / 86400 ))
            
            if [ "$days_until_expiry" -lt 30 ]; then
                log "WARN" "SSL certificate expires in $days_until_expiry days"
                ssl_issues=$((ssl_issues + 1))
            fi
            
            log "INFO" "SSL certificate expires in $days_until_expiry days"
        else
            log "WARN" "SSL certificate not found or not accessible"
            ssl_issues=$((ssl_issues + 1))
        fi
    fi
    
    # Check for weak SSL ciphers
    if command -v nmap >/dev/null 2>&1; then
        local weak_ciphers=$(nmap --script ssl-enum-ciphers -p 443 localhost 2>/dev/null | grep -i "weak\|deprecated" | wc -l)
        if [ "$weak_ciphers" -gt 0 ]; then
            log "WARN" "Weak SSL ciphers detected: $weak_ciphers"
            ssl_issues=$((ssl_issues + 1))
        fi
    fi
    
    if [ "$ssl_issues" -gt 0 ]; then
        send_security_alert "WARNING" "SSL/TLS" "SSL security issues detected: $ssl_issues" "Multiple SSL/TLS anomalies"
    fi
    
    return $ssl_issues
}

# =============================================================================
# Main Security Monitoring Function
# =============================================================================

# Main security monitoring function
main() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  $SCRIPT_NAME v$SCRIPT_VERSION${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo
    
    log "INFO" "Starting security monitoring..."
    
    # Run all security monitors
    local failed_logins=$(monitor_failed_logins)
    local suspicious_activities=$(monitor_file_activities)
    local network_threats=$(monitor_network_security)
    local integrity_issues=$(monitor_system_integrity)
    local web_threats=$(monitor_web_security)
    local db_threats=$(monitor_database_security)
    local ssl_issues=$(monitor_ssl_security)
    
    # Calculate total security score
    local total_threats=$((failed_logins + suspicious_activities + network_threats + integrity_issues + web_threats + db_threats + ssl_issues))
    
    # Display security summary
    echo
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Security Summary${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "Failed Logins: $failed_logins"
    echo -e "Suspicious Activities: $suspicious_activities"
    echo -e "Network Threats: $network_threats"
    echo -e "Integrity Issues: $integrity_issues"
    echo -e "Web Threats: $web_threats"
    echo -e "Database Threats: $db_threats"
    echo -e "SSL Issues: $ssl_issues"
    echo -e "Total Threats: $total_threats"
    echo -e "Timestamp: $(date)"
    echo -e "Server: $(hostname)"
    echo
    
    # Determine overall security status
    if [ "$total_threats" -eq 0 ]; then
        log "INFO" "Security monitoring: All systems secure"
        echo -e "${GREEN}Overall Security: EXCELLENT${NC}"
    elif [ "$total_threats" -le 2 ]; then
        log "WARN" "Security monitoring: Minor issues detected"
        echo -e "${YELLOW}Overall Security: GOOD${NC}"
    elif [ "$total_threats" -le 5 ]; then
        log "WARN" "Security monitoring: Some issues detected"
        echo -e "${YELLOW}Overall Security: FAIR${NC}"
    else
        log "ERROR" "Security monitoring: Multiple threats detected"
        echo -e "${RED}Overall Security: POOR${NC}"
        send_security_alert "CRITICAL" "Security" "Multiple security threats detected: $total_threats" "Immediate attention required"
    fi
    
    echo
    log "INFO" "Security monitoring completed"
    
    # Return exit code based on security status
    if [ "$total_threats" -gt 5 ]; then
        exit 1
    else
        exit 0
    fi
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
            echo "  --email EMAIL       Set alert email address"
            echo "  --login-threshold   Set failed login threshold (default: 10)"
            echo "  --activity-threshold Set suspicious activity threshold (default: 5)"
            echo
            exit 0
            ;;
        --version|-v)
            echo "$SCRIPT_NAME v$SCRIPT_VERSION"
            exit 0
            ;;
        --email)
            ALERT_EMAIL="$2"
            shift 2
            ;;
        --login-threshold)
            FAILED_LOGIN_THRESHOLD="$2"
            shift 2
            ;;
        --activity-threshold)
            SUSPICIOUS_ACTIVITY_THRESHOLD="$2"
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

