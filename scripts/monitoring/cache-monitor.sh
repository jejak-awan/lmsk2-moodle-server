#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: Cache Monitor Script
# =============================================================================
# Description: Cache performance monitoring for LMSK2-Moodle-Server
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
SCRIPT_NAME="LMSK2 Cache Monitor"
SCRIPT_VERSION="1.0"
LOG_FILE="/var/log/lmsk2-monitoring/cache-monitor.log"
CONFIG_DIR="/opt/lmsk2-moodle-server/scripts/config"
ALERT_EMAIL="admin@localhost"

# Cache thresholds
REDIS_MEMORY_THRESHOLD=80
OPCACHE_MEMORY_THRESHOLD=80
NGINX_CACHE_THRESHOLD=80
CACHE_HIT_RATE_THRESHOLD=70

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

# Send cache alert
send_cache_alert() {
    local level="$1"
    local component="$2"
    local message="$3"
    local details="$4"
    
    local subject="LMSK2 Cache Alert: $component - $level"
    local body="Cache Component: $component
Alert Level: $level
Message: $message
Details: $details
Timestamp: $(date)
Server: $(hostname)"

    echo "$body" | mail -s "$subject" "$ALERT_EMAIL"
    log "INFO" "Cache alert sent: $subject"
}

# =============================================================================
# Cache Monitoring Functions
# =============================================================================

# Monitor Redis cache
monitor_redis_cache() {
    log "INFO" "Monitoring Redis cache..."
    
    local redis_issues=0
    
    if command -v redis-cli >/dev/null 2>&1; then
        # Get Redis information
        local redis_info=$(redis-cli info 2>/dev/null)
        
        if [ -n "$redis_info" ]; then
            # Memory usage
            local redis_used_memory=$(echo "$redis_info" | grep "used_memory_human:" | awk -F: '{print $2}' | tr -d '\r')
            local redis_max_memory=$(echo "$redis_info" | grep "maxmemory_human:" | awk -F: '{print $2}' | tr -d '\r')
            local redis_memory_usage=$(echo "$redis_info" | grep "used_memory_percentage:" | awk -F: '{print $2}' | tr -d '\r')
            
            log "INFO" "Redis Memory: Used: $redis_used_memory, Max: $redis_max_memory, Usage: ${redis_memory_usage}%"
            
            if [ -n "$redis_memory_usage" ] && [ "$redis_memory_usage" -gt "$REDIS_MEMORY_THRESHOLD" ]; then
                log "WARN" "High Redis memory usage: ${redis_memory_usage}%"
                redis_issues=$((redis_issues + 1))
            fi
            
            # Hit rate
            local redis_keyspace_hits=$(echo "$redis_info" | grep "keyspace_hits:" | awk -F: '{print $2}' | tr -d '\r')
            local redis_keyspace_misses=$(echo "$redis_info" | grep "keyspace_misses:" | awk -F: '{print $2}' | tr -d '\r')
            
            if [ -n "$redis_keyspace_hits" ] && [ -n "$redis_keyspace_misses" ]; then
                local total_requests=$((redis_keyspace_hits + redis_keyspace_misses))
                if [ "$total_requests" -gt 0 ]; then
                    local hit_rate=$((redis_keyspace_hits * 100 / total_requests))
                    log "INFO" "Redis Hit Rate: ${hit_rate}%"
                    
                    if [ "$hit_rate" -lt "$CACHE_HIT_RATE_THRESHOLD" ]; then
                        log "WARN" "Low Redis hit rate: ${hit_rate}%"
                        redis_issues=$((redis_issues + 1))
                    fi
                fi
            fi
            
            # Connected clients
            local redis_connected_clients=$(echo "$redis_info" | grep "connected_clients:" | awk -F: '{print $2}' | tr -d '\r')
            log "INFO" "Redis Connected Clients: $redis_connected_clients"
            
            # Operations per second
            local redis_ops_per_sec=$(echo "$redis_info" | grep "instantaneous_ops_per_sec:" | awk -F: '{print $2}' | tr -d '\r')
            log "INFO" "Redis Operations/sec: $redis_ops_per_sec"
            
            # Key count
            local redis_key_count=$(echo "$redis_info" | grep "db0:keys=" | awk -F: '{print $2}' | awk -F, '{print $1}' | tr -d '\r')
            if [ -n "$redis_key_count" ]; then
                log "INFO" "Redis Keys: $redis_key_count"
            fi
            
            # Expired keys
            local redis_expired_keys=$(echo "$redis_info" | grep "expired_keys:" | awk -F: '{print $2}' | tr -d '\r')
            log "INFO" "Redis Expired Keys: $redis_expired_keys"
            
            # Evicted keys
            local redis_evicted_keys=$(echo "$redis_info" | grep "evicted_keys:" | awk -F: '{print $2}' | tr -d '\r')
            if [ -n "$redis_evicted_keys" ] && [ "$redis_evicted_keys" -gt 0 ]; then
                log "WARN" "Redis keys evicted: $redis_evicted_keys"
                redis_issues=$((redis_issues + 1))
            fi
            
        else
            log "ERROR" "Failed to get Redis information"
            redis_issues=$((redis_issues + 1))
        fi
    else
        log "WARN" "Redis client not found, skipping Redis monitoring"
    fi
    
    if [ "$redis_issues" -gt 0 ]; then
        send_cache_alert "WARNING" "Redis" "Redis cache issues detected: $redis_issues" "Multiple Redis anomalies"
    fi
    
    return $redis_issues
}

# Monitor OPcache
monitor_opcache() {
    log "INFO" "Monitoring OPcache..."
    
    local opcache_issues=0
    
    # Check if OPcache is enabled
    if php -m | grep -q "Zend OPcache"; then
        # Get OPcache statistics
        local opcache_stats=$(php -r "if (function_exists('opcache_get_status')) { print_r(opcache_get_status()); } else { echo 'OPcache not available'; }" 2>/dev/null)
        
        if [ -n "$opcache_stats" ] && [ "$opcache_stats" != "OPcache not available" ]; then
            # Parse OPcache statistics
            local opcache_enabled=$(echo "$opcache_stats" | grep "opcache_enabled" | awk '{print $3}')
            local opcache_memory_usage=$(echo "$opcache_stats" | grep "used_memory" | awk '{print $3}')
            local opcache_memory_free=$(echo "$opcache_stats" | grep "free_memory" | awk '{print $3}')
            local opcache_hit_rate=$(echo "$opcache_stats" | grep "opcache_hit_rate" | awk '{print $3}')
            local opcache_cached_scripts=$(echo "$opcache_stats" | grep "num_cached_scripts" | awk '{print $3}')
            
            if [ "$opcache_enabled" = "1" ]; then
                log "INFO" "OPcache: Enabled"
                log "INFO" "OPcache Memory: Used: ${opcache_memory_usage} bytes, Free: ${opcache_memory_free} bytes"
                log "INFO" "OPcache Hit Rate: ${opcache_hit_rate}%"
                log "INFO" "OPcache Cached Scripts: $opcache_cached_scripts"
                
                # Check hit rate
                if [ -n "$opcache_hit_rate" ] && [ "$opcache_hit_rate" -lt "$CACHE_HIT_RATE_THRESHOLD" ]; then
                    log "WARN" "Low OPcache hit rate: ${opcache_hit_rate}%"
                    opcache_issues=$((opcache_issues + 1))
                fi
                
                # Check memory usage
                if [ -n "$opcache_memory_usage" ] && [ -n "$opcache_memory_free" ]; then
                    local total_memory=$((opcache_memory_usage + opcache_memory_free))
                    if [ "$total_memory" -gt 0 ]; then
                        local memory_usage_percent=$((opcache_memory_usage * 100 / total_memory))
                        if [ "$memory_usage_percent" -gt "$OPCACHE_MEMORY_THRESHOLD" ]; then
                            log "WARN" "High OPcache memory usage: ${memory_usage_percent}%"
                            opcache_issues=$((opcache_issues + 1))
                        fi
                    fi
                fi
            else
                log "WARN" "OPcache is disabled"
                opcache_issues=$((opcache_issues + 1))
            fi
        else
            log "WARN" "OPcache statistics not available"
            opcache_issues=$((opcache_issues + 1))
        fi
    else
        log "WARN" "OPcache module not loaded"
        opcache_issues=$((opcache_issues + 1))
    fi
    
    if [ "$opcache_issues" -gt 0 ]; then
        send_cache_alert "WARNING" "OPcache" "OPcache issues detected: $opcache_issues" "Multiple OPcache anomalies"
    fi
    
    return $opcache_issues
}

# Monitor Nginx cache
monitor_nginx_cache() {
    log "INFO" "Monitoring Nginx cache..."
    
    local nginx_cache_issues=0
    
    # Check if Nginx is running
    if systemctl is-active --quiet nginx; then
        # Check Nginx cache directory
        local nginx_cache_dir="/var/cache/nginx"
        if [ -d "$nginx_cache_dir" ]; then
            local cache_size=$(du -sh "$nginx_cache_dir" 2>/dev/null | awk '{print $1}')
            local cache_files=$(find "$nginx_cache_dir" -type f 2>/dev/null | wc -l)
            
            log "INFO" "Nginx Cache: Size: $cache_size, Files: $cache_files"
            
            # Check cache directory permissions
            local cache_permissions=$(stat -c "%a" "$nginx_cache_dir" 2>/dev/null)
            if [ "$cache_permissions" != "755" ] && [ "$cache_permissions" != "700" ]; then
                log "WARN" "Nginx cache directory has unusual permissions: $cache_permissions"
                nginx_cache_issues=$((nginx_cache_issues + 1))
            fi
        else
            log "WARN" "Nginx cache directory not found: $nginx_cache_dir"
            nginx_cache_issues=$((nginx_cache_issues + 1))
        fi
        
        # Check Nginx cache configuration
        local nginx_config="/etc/nginx/nginx.conf"
        if [ -f "$nginx_config" ]; then
            local cache_config=$(grep -c "proxy_cache\|fastcgi_cache" "$nginx_config")
            if [ "$cache_config" -eq 0 ]; then
                log "WARN" "No cache configuration found in Nginx"
                nginx_cache_issues=$((nginx_cache_issues + 1))
            else
                log "INFO" "Nginx cache configuration found: $cache_config directives"
            fi
        fi
        
        # Check Nginx cache status (if available)
        local nginx_status=$(curl -s http://localhost/nginx_status 2>/dev/null)
        if [ -n "$nginx_status" ]; then
            log "INFO" "Nginx status available"
        else
            log "DEBUG" "Nginx status not available"
        fi
        
    else
        log "ERROR" "Nginx is not running"
        nginx_cache_issues=$((nginx_cache_issues + 1))
    fi
    
    if [ "$nginx_cache_issues" -gt 0 ]; then
        send_cache_alert "WARNING" "Nginx Cache" "Nginx cache issues detected: $nginx_cache_issues" "Multiple Nginx cache anomalies"
    fi
    
    return $nginx_cache_issues
}

# Monitor Moodle cache
monitor_moodle_cache() {
    log "INFO" "Monitoring Moodle cache..."
    
    local moodle_cache_issues=0
    
    # Check Moodle cache directory
    local moodle_cache_dir="/var/www/moodle/cache"
    if [ -d "$moodle_cache_dir" ]; then
        local cache_size=$(du -sh "$moodle_cache_dir" 2>/dev/null | awk '{print $1}')
        local cache_files=$(find "$moodle_cache_dir" -type f 2>/dev/null | wc -l)
        local cache_dirs=$(find "$moodle_cache_dir" -type d 2>/dev/null | wc -l)
        
        log "INFO" "Moodle Cache: Size: $cache_size, Files: $cache_files, Directories: $cache_dirs"
        
        # Check cache directory permissions
        local cache_permissions=$(stat -c "%a" "$moodle_cache_dir" 2>/dev/null)
        if [ "$cache_permissions" != "755" ] && [ "$cache_permissions" != "777" ]; then
            log "WARN" "Moodle cache directory has unusual permissions: $cache_permissions"
            moodle_cache_issues=$((moodle_cache_issues + 1))
        fi
        
        # Check for old cache files
        local old_cache_files=$(find "$moodle_cache_dir" -type f -mtime +7 2>/dev/null | wc -l)
        if [ "$old_cache_files" -gt 1000 ]; then
            log "WARN" "High number of old cache files: $old_cache_files"
            moodle_cache_issues=$((moodle_cache_issues + 1))
        fi
        
        # Check cache directory ownership
        local cache_owner=$(stat -c "%U:%G" "$moodle_cache_dir" 2>/dev/null)
        if [ "$cache_owner" != "www-data:www-data" ]; then
            log "WARN" "Moodle cache directory has unusual ownership: $cache_owner"
            moodle_cache_issues=$((moodle_cache_issues + 1))
        fi
        
    else
        log "WARN" "Moodle cache directory not found: $moodle_cache_dir"
        moodle_cache_issues=$((moodle_cache_issues + 1))
    fi
    
    # Check Moodle temp directory
    local moodle_temp_dir="/var/www/moodle/temp"
    if [ -d "$moodle_temp_dir" ]; then
        local temp_size=$(du -sh "$moodle_temp_dir" 2>/dev/null | awk '{print $1}')
        local temp_files=$(find "$moodle_temp_dir" -type f 2>/dev/null | wc -l)
        
        log "INFO" "Moodle Temp: Size: $temp_size, Files: $temp_files"
        
        # Check for old temp files
        local old_temp_files=$(find "$moodle_temp_dir" -type f -mtime +1 2>/dev/null | wc -l)
        if [ "$old_temp_files" -gt 100 ]; then
            log "WARN" "High number of old temp files: $old_temp_files"
            moodle_cache_issues=$((moodle_cache_issues + 1))
        fi
    fi
    
    if [ "$moodle_cache_issues" -gt 0 ]; then
        send_cache_alert "WARNING" "Moodle Cache" "Moodle cache issues detected: $moodle_cache_issues" "Multiple Moodle cache anomalies"
    fi
    
    return $moodle_cache_issues
}

# Monitor system cache
monitor_system_cache() {
    log "INFO" "Monitoring system cache..."
    
    local system_cache_issues=0
    
    # Check system cache usage
    local system_cache=$(free | grep "buff/cache" | awk '{print $3}')
    local total_memory=$(free | grep "Mem:" | awk '{print $2}')
    local cache_percentage=$((system_cache * 100 / total_memory))
    
    log "INFO" "System Cache: ${system_cache}MB (${cache_percentage}% of total memory)"
    
    # Check page cache
    local page_cache=$(cat /proc/meminfo | grep "Cached:" | awk '{print $2}')
    local page_cache_mb=$((page_cache / 1024))
    log "INFO" "Page Cache: ${page_cache_mb}MB"
    
    # Check buffer cache
    local buffer_cache=$(cat /proc/meminfo | grep "Buffers:" | awk '{print $2}')
    local buffer_cache_mb=$((buffer_cache / 1024))
    log "INFO" "Buffer Cache: ${buffer_cache_mb}MB"
    
    # Check swap cache
    local swap_cache=$(cat /proc/meminfo | grep "SwapCached:" | awk '{print $2}')
    local swap_cache_mb=$((swap_cache / 1024))
    if [ "$swap_cache_mb" -gt 0 ]; then
        log "INFO" "Swap Cache: ${swap_cache_mb}MB"
    fi
    
    # Check for cache pressure
    local cache_pressure=$(cat /proc/sys/vm/vfs_cache_pressure 2>/dev/null)
    if [ -n "$cache_pressure" ]; then
        log "INFO" "Cache Pressure: $cache_pressure"
        if [ "$cache_pressure" -gt 100 ]; then
            log "WARN" "High cache pressure: $cache_pressure"
            system_cache_issues=$((system_cache_issues + 1))
        fi
    fi
    
    if [ "$system_cache_issues" -gt 0 ]; then
        send_cache_alert "WARNING" "System Cache" "System cache issues detected: $system_cache_issues" "Multiple system cache anomalies"
    fi
    
    return $system_cache_issues
}

# =============================================================================
# Main Cache Monitoring Function
# =============================================================================

# Main cache monitoring function
main() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  $SCRIPT_NAME v$SCRIPT_VERSION${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo
    
    log "INFO" "Starting cache monitoring..."
    
    # Run all cache monitors
    local redis_issues=$(monitor_redis_cache)
    local opcache_issues=$(monitor_opcache)
    local nginx_cache_issues=$(monitor_nginx_cache)
    local moodle_cache_issues=$(monitor_moodle_cache)
    local system_cache_issues=$(monitor_system_cache)
    
    # Calculate total cache issues
    local total_issues=$((redis_issues + opcache_issues + nginx_cache_issues + moodle_cache_issues + system_cache_issues))
    
    # Display cache summary
    echo
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  Cache Summary${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo -e "Redis Issues: $redis_issues"
    echo -e "OPcache Issues: $opcache_issues"
    echo -e "Nginx Cache Issues: $nginx_cache_issues"
    echo -e "Moodle Cache Issues: $moodle_cache_issues"
    echo -e "System Cache Issues: $system_cache_issues"
    echo -e "Total Issues: $total_issues"
    echo -e "Timestamp: $(date)"
    echo -e "Server: $(hostname)"
    echo
    
    # Determine overall cache status
    if [ "$total_issues" -eq 0 ]; then
        log "INFO" "Cache monitoring: All caches healthy"
        echo -e "${GREEN}Overall Cache Status: EXCELLENT${NC}"
    elif [ "$total_issues" -le 2 ]; then
        log "WARN" "Cache monitoring: Minor issues detected"
        echo -e "${YELLOW}Overall Cache Status: GOOD${NC}"
    elif [ "$total_issues" -le 5 ]; then
        log "WARN" "Cache monitoring: Some issues detected"
        echo -e "${YELLOW}Overall Cache Status: FAIR${NC}"
    else
        log "ERROR" "Cache monitoring: Multiple issues detected"
        echo -e "${RED}Overall Cache Status: POOR${NC}"
        send_cache_alert "CRITICAL" "Cache System" "Multiple cache issues detected: $total_issues" "Immediate attention required"
    fi
    
    echo
    log "INFO" "Cache monitoring completed"
    
    # Return exit code based on cache status
    if [ "$total_issues" -gt 5 ]; then
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
            echo "  --redis-threshold   Set Redis memory threshold (default: 80)"
            echo "  --opcache-threshold Set OPcache memory threshold (default: 80)"
            echo "  --hit-rate-threshold Set cache hit rate threshold (default: 70)"
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
        --redis-threshold)
            REDIS_MEMORY_THRESHOLD="$2"
            shift 2
            ;;
        --opcache-threshold)
            OPCACHE_MEMORY_THRESHOLD="$2"
            shift 2
            ;;
        --hit-rate-threshold)
            CACHE_HIT_RATE_THRESHOLD="$2"
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

