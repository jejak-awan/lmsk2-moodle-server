#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: System Health Check Script
# =============================================================================
# Description: Comprehensive system health monitoring for LMSK2-Moodle-Server
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
SCRIPT_NAME="LMSK2 System Health Check"
SCRIPT_VERSION="1.0"
LOG_FILE="/var/log/lmsk2-monitoring/system-health-check.log"
CONFIG_DIR="/opt/lmsk2-moodle-server/scripts/config"
ALERT_EMAIL="admin@localhost"

# Health check thresholds
CPU_THRESHOLD=80
MEMORY_THRESHOLD=85
DISK_THRESHOLD=90
LOAD_THRESHOLD=2.0
SWAP_THRESHOLD=50

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
    esac
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# Send alert
send_alert() {
    local level="$1"
    local component="$2"
    local message="$3"
    
    local subject="LMSK2 Health Check Alert: $component - $level"
    local body="Component: $component
Level: $level
Message: $message
Timestamp: $(date)
Server: $(hostname)"

    echo "$body" | mail -s "$subject" "$ALERT_EMAIL"
    log "INFO" "Alert sent: $subject"
}

# =============================================================================
# Health Check Functions
# =============================================================================

# Check CPU usage
check_cpu() {
    log "INFO" "Checking CPU usage..."
    
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
    local cpu_int=${cpu_usage%.*}
    
    if [ "$cpu_int" -gt "$CPU_THRESHOLD" ]; then
        log "WARN" "High CPU usage: ${cpu_usage}% (threshold: ${CPU_THRESHOLD}%)"
        send_alert "WARNING" "CPU" "High CPU usage: ${cpu_usage}%"
        return 1
    else
        log "INFO" "CPU usage normal: ${cpu_usage}%"
        return 0
    fi
}

# Check memory usage
check_memory() {
    log "INFO" "Checking memory usage..."
    
    local memory_info=$(free | grep Mem)
    local total_memory=$(echo $memory_info | awk '{print $2}')
    local used_memory=$(echo $memory_info | awk '{print $3}')
    local memory_usage=$((used_memory * 100 / total_memory))
    
    if [ "$memory_usage" -gt "$MEMORY_THRESHOLD" ]; then
        log "WARN" "High memory usage: ${memory_usage}% (threshold: ${MEMORY_THRESHOLD}%)"
        send_alert "WARNING" "Memory" "High memory usage: ${memory_usage}%"
        return 1
    else
        log "INFO" "Memory usage normal: ${memory_usage}%"
        return 0
    fi
}

# Check swap usage
check_swap() {
    log "INFO" "Checking swap usage..."
    
    local swap_info=$(free | grep Swap)
    local total_swap=$(echo $swap_info | awk '{print $2}')
    local used_swap=$(echo $swap_info | awk '{print $3}')
    
    if [ "$total_swap" -gt 0 ]; then
        local swap_usage=$((used_swap * 100 / total_swap))
        
        if [ "$swap_usage" -gt "$SWAP_THRESHOLD" ]; then
            log "WARN" "High swap usage: ${swap_usage}% (threshold: ${SWAP_THRESHOLD}%)"
            send_alert "WARNING" "Swap" "High swap usage: ${swap_usage}%"
            return 1
        else
            log "INFO" "Swap usage normal: ${swap_usage}%"
        fi
    else
        log "INFO" "No swap configured"
    fi
    
    return 0
}

# Check disk usage
check_disk() {
    log "INFO" "Checking disk usage..."
    
    local disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
    
    if [ "$disk_usage" -gt "$DISK_THRESHOLD" ]; then
        log "WARN" "High disk usage: ${disk_usage}% (threshold: ${DISK_THRESHOLD}%)"
        send_alert "WARNING" "Disk" "High disk usage: ${disk_usage}%"
        return 1
    else
        log "INFO" "Disk usage normal: ${disk_usage}%"
        return 0
    fi
}

# Check load average
check_load() {
    log "INFO" "Checking load average..."
    
    local load_avg=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
    local cpu_cores=$(nproc)
    local load_threshold=$(echo "$cpu_cores * $LOAD_THRESHOLD" | bc)
    
    if (( $(echo "$load_avg > $load_threshold" | bc -l) )); then
        log "WARN" "High load average: $load_avg (threshold: $load_threshold)"
        send_alert "WARNING" "Load" "High load average: $load_avg"
        return 1
    else
        log "INFO" "Load average normal: $load_avg"
        return 0
    fi
}

# Check network connectivity
check_network() {
    log "INFO" "Checking network connectivity..."
    
    local network_issues=0
    
    # Check if network interfaces are up
    local interfaces=$(ip link show | grep -E "^[0-9]+:" | grep -v lo | awk -F: '{print $2}' | tr -d ' ')
    
    for interface in $interfaces; do
        if ! ip link show "$interface" | grep -q "state UP"; then
            log "WARN" "Network interface $interface is down"
            network_issues=$((network_issues + 1))
        fi
    done
    
    # Check DNS resolution
    if ! nslookup google.com >/dev/null 2>&1; then
        log "WARN" "DNS resolution failed"
        network_issues=$((network_issues + 1))
    fi
    
    # Check internet connectivity
    if ! ping -c 1 8.8.8.8 >/dev/null 2>&1; then
        log "WARN" "Internet connectivity failed"
        network_issues=$((network_issues + 1))
    fi
    
    if [ "$network_issues" -gt 0 ]; then
        send_alert "WARNING" "Network" "Network issues detected: $network_issues problems"
        return 1
    else
        log "INFO" "Network connectivity normal"
        return 0
    fi
}

# Check system services
check_services() {
    log "INFO" "Checking system services..."
    
    local services=("nginx" "php8.1-fpm" "mariadb" "redis-server")
    local failed_services=()
    
    for service in "${services[@]}"; do
        if ! systemctl is-active --quiet "$service"; then
            log "ERROR" "Service $service is not running"
            failed_services+=("$service")
        else
            log "INFO" "Service $service is running"
        fi
    done
    
    if [ ${#failed_services[@]} -gt 0 ]; then
        local failed_list=$(IFS=', '; echo "${failed_services[*]}")
        send_alert "CRITICAL" "Services" "Failed services: $failed_list"
        return 1
    else
        log "INFO" "All system services are running"
        return 0
    fi
}

# Check database connectivity
check_database() {
    log "INFO" "Checking database connectivity..."
    
    if command -v mysql >/dev/null 2>&1; then
        if mysql -e "SELECT 1;" >/dev/null 2>&1; then
            log "INFO" "Database connectivity normal"
            return 0
        else
            log "ERROR" "Database connectivity failed"
            send_alert "CRITICAL" "Database" "Database connectivity failed"
            return 1
        fi
    else
        log "WARN" "MySQL client not found, skipping database check"
        return 0
    fi
}

# Check Redis connectivity
check_redis() {
    log "INFO" "Checking Redis connectivity..."
    
    if command -v redis-cli >/dev/null 2>&1; then
        if redis-cli ping >/dev/null 2>&1; then
            log "INFO" "Redis connectivity normal"
            return 0
        else
            log "ERROR" "Redis connectivity failed"
            send_alert "CRITICAL" "Redis" "Redis connectivity failed"
            return 1
        fi
    else
        log "WARN" "Redis client not found, skipping Redis check"
        return 0
    fi
}

# Check file system health
check_filesystem() {
    log "INFO" "Checking file system health..."
    
    local filesystem_issues=0
    
    # Check for read-only file systems
    if mount | grep -q "ro,"; then
        log "WARN" "Read-only file system detected"
        filesystem_issues=$((filesystem_issues + 1))
    fi
    
    # Check for disk errors
    if dmesg | grep -i "I/O error\|disk error" | tail -5 | grep -q .; then
        log "WARN" "Disk I/O errors detected in system logs"
        filesystem_issues=$((filesystem_issues + 1))
    fi
    
    # Check inode usage
    local inode_usage=$(df -i / | tail -1 | awk '{print $5}' | sed 's/%//')
    if [ "$inode_usage" -gt 90 ]; then
        log "WARN" "High inode usage: ${inode_usage}%"
        filesystem_issues=$((filesystem_issues + 1))
    fi
    
    if [ "$filesystem_issues" -gt 0 ]; then
        send_alert "WARNING" "Filesystem" "File system issues detected: $filesystem_issues problems"
        return 1
    else
        log "INFO" "File system health normal"
        return 0
    fi
}

# Check system uptime
check_uptime() {
    log "INFO" "Checking system uptime..."
    
    local uptime_seconds=$(cat /proc/uptime | awk '{print $1}')
    local uptime_days=$(echo "$uptime_seconds / 86400" | bc)
    
    log "INFO" "System uptime: ${uptime_days} days"
    
    # Alert if system has been up for more than 365 days (potential for memory leaks)
    if [ "$uptime_days" -gt 365 ]; then
        log "WARN" "System uptime is very high: ${uptime_days} days"
        send_alert "INFO" "Uptime" "System uptime is very high: ${uptime_days} days"
    fi
    
    return 0
}

# =============================================================================
# Main Health Check Function
# =============================================================================

# Main health check function
main() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  $SCRIPT_NAME v$SCRIPT_VERSION${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo
    
    log "INFO" "Starting system health check..."
    
    local health_score=0
    local total_checks=10
    
    # Run all health checks
    check_cpu && health_score=$((health_score + 1))
    check_memory && health_score=$((health_score + 1))
    check_swap && health_score=$((health_score + 1))
    check_disk && health_score=$((health_score + 1))
    check_load && health_score=$((health_score + 1))
    check_network && health_score=$((health_score + 1))
    check_services && health_score=$((health_score + 1))
    check_database && health_score=$((health_score + 1))
    check_redis && health_score=$((health_score + 1))
    check_filesystem && health_score=$((health_score + 1))
    check_uptime && health_score=$((health_score + 1))
    
    # Calculate health percentage
    local health_percentage=$((health_score * 100 / total_checks))
    
    echo
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Health Check Summary${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "Health Score: ${health_score}/${total_checks} (${health_percentage}%)"
    echo -e "Timestamp: $(date)"
    echo -e "Server: $(hostname)"
    echo
    
    # Determine overall health status
    if [ "$health_percentage" -ge 90 ]; then
        log "INFO" "System health: EXCELLENT (${health_percentage}%)"
        echo -e "${GREEN}Overall Status: EXCELLENT${NC}"
    elif [ "$health_percentage" -ge 75 ]; then
        log "INFO" "System health: GOOD (${health_percentage}%)"
        echo -e "${GREEN}Overall Status: GOOD${NC}"
    elif [ "$health_percentage" -ge 50 ]; then
        log "WARN" "System health: FAIR (${health_percentage}%)"
        echo -e "${YELLOW}Overall Status: FAIR${NC}"
    else
        log "ERROR" "System health: POOR (${health_percentage}%)"
        echo -e "${RED}Overall Status: POOR${NC}"
        send_alert "CRITICAL" "System" "System health is poor: ${health_percentage}%"
    fi
    
    echo
    log "INFO" "System health check completed"
    
    # Return exit code based on health
    if [ "$health_percentage" -lt 75 ]; then
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
            echo "  --cpu-threshold     Set CPU usage threshold (default: 80)"
            echo "  --memory-threshold  Set memory usage threshold (default: 85)"
            echo "  --disk-threshold    Set disk usage threshold (default: 90)"
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
        --cpu-threshold)
            CPU_THRESHOLD="$2"
            shift 2
            ;;
        --memory-threshold)
            MEMORY_THRESHOLD="$2"
            shift 2
            ;;
        --disk-threshold)
            DISK_THRESHOLD="$2"
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

