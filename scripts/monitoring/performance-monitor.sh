#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: Performance Monitor Script
# =============================================================================
# Description: Real-time performance monitoring for LMSK2-Moodle-Server
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
SCRIPT_NAME="LMSK2 Performance Monitor"
SCRIPT_VERSION="1.0"
LOG_FILE="/var/log/lmsk2-monitoring/performance-monitor.log"
CONFIG_DIR="/opt/lmsk2-moodle-server/scripts/config"
ALERT_EMAIL="admin@localhost"

# Performance thresholds
CPU_THRESHOLD=80
MEMORY_THRESHOLD=85
DISK_IO_THRESHOLD=80
NETWORK_THRESHOLD=80
RESPONSE_TIME_THRESHOLD=2000  # milliseconds

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
        "DEBUG")
            echo -e "${BLUE}[DEBUG]${NC} $message" | tee -a "$LOG_FILE"
            ;;
    esac
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# Send alert
send_alert() {
    local level="$1"
    local component="$2"
    local message="$3"
    
    local subject="LMSK2 Performance Alert: $component - $level"
    local body="Component: $component
Level: $level
Message: $message
Timestamp: $(date)
Server: $(hostname)"

    echo "$body" | mail -s "$subject" "$ALERT_EMAIL"
    log "INFO" "Performance alert sent: $subject"
}

# =============================================================================
# Performance Monitoring Functions
# =============================================================================

# Monitor CPU performance
monitor_cpu() {
    log "INFO" "Monitoring CPU performance..."
    
    # Get CPU usage from /proc/stat
    local cpu_info=$(cat /proc/stat | grep "cpu " | awk '{print $2, $3, $4, $5, $6, $7, $8}')
    local user=$(echo $cpu_info | awk '{print $1}')
    local nice=$(echo $cpu_info | awk '{print $2}')
    local system=$(echo $cpu_info | awk '{print $3}')
    local idle=$(echo $cpu_info | awk '{print $4}')
    local iowait=$(echo $cpu_info | awk '{print $5}')
    local irq=$(echo $cpu_info | awk '{print $6}')
    local softirq=$(echo $cpu_info | awk '{print $7}')
    
    local total=$((user + nice + system + idle + iowait + irq + softirq))
    local used=$((user + nice + system + irq + softirq))
    local cpu_usage=$((used * 100 / total))
    
    # Get load average
    local load_avg=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
    local cpu_cores=$(nproc)
    
    # Get CPU frequency
    local cpu_freq=$(cat /proc/cpuinfo | grep "cpu MHz" | head -1 | awk '{print $4}')
    
    log "INFO" "CPU Usage: ${cpu_usage}%, Load Average: $load_avg, Cores: $cpu_cores, Frequency: ${cpu_freq}MHz"
    
    if [ "$cpu_usage" -gt "$CPU_THRESHOLD" ]; then
        log "WARN" "High CPU usage detected: ${cpu_usage}%"
        send_alert "WARNING" "CPU" "High CPU usage: ${cpu_usage}%"
    fi
    
    # Return CPU usage for summary
    echo "$cpu_usage"
}

# Monitor memory performance
monitor_memory() {
    log "INFO" "Monitoring memory performance..."
    
    # Get memory information
    local memory_info=$(free -m)
    local total_mem=$(echo "$memory_info" | grep "Mem:" | awk '{print $2}')
    local used_mem=$(echo "$memory_info" | grep "Mem:" | awk '{print $3}')
    local free_mem=$(echo "$memory_info" | grep "Mem:" | awk '{print $4}')
    local available_mem=$(echo "$memory_info" | grep "Mem:" | awk '{print $7}')
    local cached_mem=$(echo "$memory_info" | grep "Mem:" | awk '{print $6}')
    local buffers_mem=$(echo "$memory_info" | grep "Mem:" | awk '{print $5}')
    
    local memory_usage=$((used_mem * 100 / total_mem))
    local memory_available_percent=$((available_mem * 100 / total_mem))
    
    # Get swap information
    local swap_info=$(echo "$memory_info" | grep "Swap:")
    local total_swap=$(echo "$swap_info" | awk '{print $2}')
    local used_swap=$(echo "$swap_info" | awk '{print $3}')
    
    log "INFO" "Memory Usage: ${memory_usage}% (${used_mem}MB/${total_mem}MB), Available: ${memory_available_percent}% (${available_mem}MB)"
    log "INFO" "Cached: ${cached_mem}MB, Buffers: ${buffers_mem}MB"
    
    if [ "$total_swap" -gt 0 ]; then
        local swap_usage=$((used_swap * 100 / total_swap))
        log "INFO" "Swap Usage: ${swap_usage}% (${used_swap}MB/${total_swap}MB)"
    fi
    
    if [ "$memory_usage" -gt "$MEMORY_THRESHOLD" ]; then
        log "WARN" "High memory usage detected: ${memory_usage}%"
        send_alert "WARNING" "Memory" "High memory usage: ${memory_usage}%"
    fi
    
    # Return memory usage for summary
    echo "$memory_usage"
}

# Monitor disk I/O performance
monitor_disk_io() {
    log "INFO" "Monitoring disk I/O performance..."
    
    # Get disk I/O statistics
    local disk_stats=$(iostat -x 1 1 | tail -n +4)
    local total_reads=0
    local total_writes=0
    local total_util=0
    local disk_count=0
    
    while IFS= read -r line; do
        if [ -n "$line" ]; then
            local device=$(echo "$line" | awk '{print $1}')
            local reads=$(echo "$line" | awk '{print $4}')
            local writes=$(echo "$line" | awk '{print $5}')
            local util=$(echo "$line" | awk '{print $10}')
            
            if [ "$device" != "Device" ] && [ "$device" != "avg-cpu:" ]; then
                total_reads=$(echo "$total_reads + $reads" | bc)
                total_writes=$(echo "$total_writes + $writes" | bc)
                total_util=$(echo "$total_util + $util" | bc)
                disk_count=$((disk_count + 1))
                
                log "DEBUG" "Disk $device: Reads: ${reads}MB/s, Writes: ${writes}MB/s, Utilization: ${util}%"
            fi
        fi
    done <<< "$disk_stats"
    
    if [ "$disk_count" -gt 0 ]; then
        local avg_util=$(echo "scale=2; $total_util / $disk_count" | bc)
        log "INFO" "Average Disk I/O Utilization: ${avg_util}%"
        
        if (( $(echo "$avg_util > $DISK_IO_THRESHOLD" | bc -l) )); then
            log "WARN" "High disk I/O utilization detected: ${avg_util}%"
            send_alert "WARNING" "Disk I/O" "High disk I/O utilization: ${avg_util}%"
        fi
        
        # Return average utilization for summary
        echo "${avg_util%.*}"
    else
        echo "0"
    fi
}

# Monitor network performance
monitor_network() {
    log "INFO" "Monitoring network performance..."
    
    # Get network statistics
    local network_stats=$(cat /proc/net/dev | grep -E "(eth|ens|enp)" | head -5)
    local total_rx=0
    local total_tx=0
    local interface_count=0
    
    while IFS= read -r line; do
        if [ -n "$line" ]; then
            local interface=$(echo "$line" | awk -F: '{print $1}' | tr -d ' ')
            local rx_bytes=$(echo "$line" | awk '{print $2}')
            local tx_bytes=$(echo "$line" | awk '{print $10}')
            
            total_rx=$((total_rx + rx_bytes))
            total_tx=$((total_tx + tx_bytes))
            interface_count=$((interface_count + 1))
            
            log "DEBUG" "Interface $interface: RX: ${rx_bytes} bytes, TX: ${tx_bytes} bytes"
        fi
    done <<< "$network_stats"
    
    # Convert to MB
    local total_rx_mb=$((total_rx / 1024 / 1024))
    local total_tx_mb=$((total_tx / 1024 / 1024))
    
    log "INFO" "Total Network: RX: ${total_rx_mb}MB, TX: ${total_tx_mb}MB"
    
    # Get network connections
    local connections=$(ss -tuln | wc -l)
    local established_connections=$(ss -tuln | grep ESTAB | wc -l)
    
    log "INFO" "Network Connections: Total: $connections, Established: $established_connections"
    
    # Return network utilization (simplified)
    echo "0"
}

# Monitor web server performance
monitor_web_server() {
    log "INFO" "Monitoring web server performance..."
    
    # Check Nginx status
    if systemctl is-active --quiet nginx; then
        # Get Nginx statistics
        local nginx_connections=$(curl -s http://localhost/nginx_status 2>/dev/null | grep "Active connections" | awk '{print $3}')
        local nginx_requests=$(curl -s http://localhost/nginx_status 2>/dev/null | grep "server accepts handled requests" | awk '{print $4}')
        
        if [ -n "$nginx_connections" ]; then
            log "INFO" "Nginx Active Connections: $nginx_connections"
        fi
        
        if [ -n "$nginx_requests" ]; then
            log "INFO" "Nginx Total Requests: $nginx_requests"
        fi
        
        # Test response time
        local response_time=$(curl -o /dev/null -s -w "%{time_total}" http://localhost/ 2>/dev/null)
        if [ -n "$response_time" ]; then
            local response_ms=$(echo "$response_time * 1000" | bc)
            log "INFO" "Web Response Time: ${response_ms}ms"
            
            if (( $(echo "$response_ms > $RESPONSE_TIME_THRESHOLD" | bc -l) )); then
                log "WARN" "Slow web response time detected: ${response_ms}ms"
                send_alert "WARNING" "Web Server" "Slow response time: ${response_ms}ms"
            fi
        fi
    else
        log "ERROR" "Nginx is not running"
        send_alert "CRITICAL" "Web Server" "Nginx is not running"
    fi
    
    # Return response time for summary
    if [ -n "$response_time" ]; then
        echo "${response_ms%.*}"
    else
        echo "0"
    fi
}

# Monitor database performance
monitor_database() {
    log "INFO" "Monitoring database performance..."
    
    if command -v mysql >/dev/null 2>&1; then
        # Get database statistics
        local db_stats=$(mysql -e "SHOW GLOBAL STATUS LIKE 'Connections'; SHOW GLOBAL STATUS LIKE 'Threads_connected'; SHOW GLOBAL STATUS LIKE 'Questions'; SHOW GLOBAL STATUS LIKE 'Uptime';" 2>/dev/null)
        
        local connections=$(echo "$db_stats" | grep "Connections" | awk '{print $2}')
        local threads_connected=$(echo "$db_stats" | grep "Threads_connected" | awk '{print $2}')
        local questions=$(echo "$db_stats" | grep "Questions" | awk '{print $2}')
        local uptime=$(echo "$db_stats" | grep "Uptime" | awk '{print $2}')
        
        if [ -n "$connections" ]; then
            log "INFO" "Database Connections: $connections, Active Threads: $threads_connected"
        fi
        
        if [ -n "$questions" ]; then
            log "INFO" "Database Questions: $questions"
        fi
        
        if [ -n "$uptime" ]; then
            local uptime_hours=$((uptime / 3600))
            log "INFO" "Database Uptime: ${uptime_hours} hours"
        fi
        
        # Get database size
        local db_size=$(mysql -e "SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS 'DB Size in MB' FROM information_schema.tables WHERE table_schema='moodle';" 2>/dev/null | tail -1)
        if [ -n "$db_size" ] && [ "$db_size" != "NULL" ]; then
            log "INFO" "Database Size: ${db_size}MB"
        fi
    else
        log "WARN" "MySQL client not found, skipping database monitoring"
    fi
    
    # Return database performance metric (simplified)
    echo "0"
}

# Monitor Redis performance
monitor_redis() {
    log "INFO" "Monitoring Redis performance..."
    
    if command -v redis-cli >/dev/null 2>&1; then
        # Get Redis statistics
        local redis_info=$(redis-cli info 2>/dev/null)
        
        if [ -n "$redis_info" ]; then
            local redis_connected_clients=$(echo "$redis_info" | grep "connected_clients:" | awk -F: '{print $2}')
            local redis_used_memory=$(echo "$redis_info" | grep "used_memory_human:" | awk -F: '{print $2}')
            local redis_ops_per_sec=$(echo "$redis_info" | grep "instantaneous_ops_per_sec:" | awk -F: '{print $2}')
            local redis_keyspace_hits=$(echo "$redis_info" | grep "keyspace_hits:" | awk -F: '{print $2}')
            local redis_keyspace_misses=$(echo "$redis_info" | grep "keyspace_misses:" | awk -F: '{print $2}')
            
            if [ -n "$redis_connected_clients" ]; then
                log "INFO" "Redis Connected Clients: $redis_connected_clients"
            fi
            
            if [ -n "$redis_used_memory" ]; then
                log "INFO" "Redis Used Memory: $redis_used_memory"
            fi
            
            if [ -n "$redis_ops_per_sec" ]; then
                log "INFO" "Redis Operations/sec: $redis_ops_per_sec"
            fi
            
            if [ -n "$redis_keyspace_hits" ] && [ -n "$redis_keyspace_misses" ]; then
                local total_requests=$((redis_keyspace_hits + redis_keyspace_misses))
                if [ "$total_requests" -gt 0 ]; then
                    local hit_rate=$((redis_keyspace_hits * 100 / total_requests))
                    log "INFO" "Redis Hit Rate: ${hit_rate}%"
                fi
            fi
        fi
    else
        log "WARN" "Redis client not found, skipping Redis monitoring"
    fi
    
    # Return Redis performance metric (simplified)
    echo "0"
}

# =============================================================================
# Main Performance Monitoring Function
# =============================================================================

# Main performance monitoring function
main() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  $SCRIPT_NAME v$SCRIPT_VERSION${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo
    
    log "INFO" "Starting performance monitoring..."
    
    # Run all performance monitors
    local cpu_usage=$(monitor_cpu)
    local memory_usage=$(monitor_memory)
    local disk_io_usage=$(monitor_disk_io)
    local network_usage=$(monitor_network)
    local web_response_time=$(monitor_web_server)
    local db_performance=$(monitor_database)
    local redis_performance=$(monitor_redis)
    
    # Display performance summary
    echo
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Performance Summary${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "CPU Usage: ${cpu_usage}%"
    echo -e "Memory Usage: ${memory_usage}%"
    echo -e "Disk I/O Usage: ${disk_io_usage}%"
    echo -e "Network Usage: ${network_usage}%"
    echo -e "Web Response Time: ${web_response_time}ms"
    echo -e "Database Performance: ${db_performance}"
    echo -e "Redis Performance: ${redis_performance}"
    echo -e "Timestamp: $(date)"
    echo -e "Server: $(hostname)"
    echo
    
    # Determine overall performance status
    local performance_issues=0
    
    if [ "$cpu_usage" -gt "$CPU_THRESHOLD" ]; then
        performance_issues=$((performance_issues + 1))
    fi
    
    if [ "$memory_usage" -gt "$MEMORY_THRESHOLD" ]; then
        performance_issues=$((performance_issues + 1))
    fi
    
    if [ "$disk_io_usage" -gt "$DISK_IO_THRESHOLD" ]; then
        performance_issues=$((performance_issues + 1))
    fi
    
    if [ "$web_response_time" -gt "$RESPONSE_TIME_THRESHOLD" ]; then
        performance_issues=$((performance_issues + 1))
    fi
    
    if [ "$performance_issues" -eq 0 ]; then
        log "INFO" "Performance monitoring: All systems normal"
        echo -e "${GREEN}Overall Performance: EXCELLENT${NC}"
    elif [ "$performance_issues" -le 2 ]; then
        log "WARN" "Performance monitoring: Some issues detected"
        echo -e "${YELLOW}Overall Performance: GOOD${NC}"
    else
        log "ERROR" "Performance monitoring: Multiple issues detected"
        echo -e "${RED}Overall Performance: POOR${NC}"
        send_alert "CRITICAL" "Performance" "Multiple performance issues detected"
    fi
    
    echo
    log "INFO" "Performance monitoring completed"
    
    # Return exit code based on performance
    if [ "$performance_issues" -gt 2 ]; then
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
            echo "  --response-threshold Set response time threshold in ms (default: 2000)"
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
        --response-threshold)
            RESPONSE_TIME_THRESHOLD="$2"
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

