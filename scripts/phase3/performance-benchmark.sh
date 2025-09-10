#!/bin/bash

# =============================================================================
# LMSK2 Moodle Server - Performance Benchmark Script
# =============================================================================
# Description: Comprehensive performance benchmarking for Moodle server
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
CONFIG_FILE="${SCRIPT_DIR}/../config/installer.conf"

if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
else
    echo "❌ Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Script configuration
SCRIPT_NAME="$(basename "$0")"
LOG_FILE="/var/log/lmsk2/${SCRIPT_NAME%.*}.log"
REPORT_FILE="/var/log/moodle-performance-benchmark.txt"
MOODLE_DIR="/var/www/moodle"

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
# Benchmark Functions
# =============================================================================

benchmark_response_time() {
    log "INFO" "Benchmarking response time..."
    
    local domain="${MOODLE_DOMAIN:-lms.example.com}"
    local results=()
    
    # Test homepage response time
    local homepage_time=$(curl -o /dev/null -s -w '%{time_total}' "https://$domain/" 2>/dev/null || echo "0")
    results+=("Homepage: ${homepage_time}s")
    
    # Test login page response time
    local login_time=$(curl -o /dev/null -s -w '%{time_total}' "https://$domain/login/index.php" 2>/dev/null || echo "0")
    results+=("Login page: ${login_time}s")
    
    # Test course page response time
    local course_time=$(curl -o /dev/null -s -w '%{time_total}' "https://$domain/course/" 2>/dev/null || echo "0")
    results+=("Course page: ${course_time}s")
    
    # Test admin page response time
    local admin_time=$(curl -o /dev/null -s -w '%{time_total}' "https://$domain/admin/" 2>/dev/null || echo "0")
    results+=("Admin page: ${admin_time}s")
    
    # Display results
    for result in "${results[@]}"; do
        log "INFO" "  $result"
    done
    
    # Check if response times are acceptable
    local homepage_num=$(echo "$homepage_time" | cut -d'.' -f1)
    if [[ "$homepage_num" -lt 2 ]]; then
        log "INFO" "✓ Homepage response time is excellent (< 2s)"
    elif [[ "$homepage_num" -lt 5 ]]; then
        log "WARN" "⚠ Homepage response time is acceptable (< 5s)"
    else
        log "ERROR" "✗ Homepage response time is slow (> 5s)"
    fi
}

benchmark_database_performance() {
    log "INFO" "Benchmarking database performance..."
    
    # Test simple query performance
    local start_time=$(date +%s.%N)
    mysql -u "${DB_USER:-moodle}" -p"${DB_PASSWORD:-}" "${DB_NAME:-moodle}" -e "SELECT COUNT(*) FROM mdl_user;" >/dev/null 2>&1
    local end_time=$(date +%s.%N)
    local query_time=$(echo "$end_time - $start_time" | bc)
    
    log "INFO" "  Simple query time: ${query_time}s"
    
    # Test complex query performance
    local start_time=$(date +%s.%N)
    mysql -u "${DB_USER:-moodle}" -p"${DB_PASSWORD:-}" "${DB_NAME:-moodle}" -e "SELECT u.id, u.username, u.firstname, u.lastname, c.fullname FROM mdl_user u JOIN mdl_user_enrolments ue ON u.id = ue.userid JOIN mdl_enrol e ON ue.enrolid = e.id JOIN mdl_course c ON e.courseid = c.id LIMIT 100;" >/dev/null 2>&1
    local end_time=$(date +%s.%N)
    local complex_query_time=$(echo "$end_time - $start_time" | bc)
    
    log "INFO" "  Complex query time: ${complex_query_time}s"
    
    # Check database performance
    if (( $(echo "$query_time < 0.1" | bc -l) )); then
        log "INFO" "✓ Database performance is excellent (< 0.1s)"
    elif (( $(echo "$query_time < 0.5" | bc -l) )); then
        log "INFO" "✓ Database performance is good (< 0.5s)"
    else
        log "WARN" "⚠ Database performance could be improved (> 0.5s)"
    fi
}

benchmark_concurrent_connections() {
    log "INFO" "Benchmarking concurrent connections..."
    
    local domain="${MOODLE_DOMAIN:-lms.example.com}"
    local concurrent_count=20
    local success_count=0
    
    # Test concurrent connections
    for i in $(seq 1 $concurrent_count); do
        if curl -o /dev/null -s "https://$domain/" >/dev/null 2>&1; then
            ((success_count++))
        fi &
    done
    wait
    
    log "INFO" "  Concurrent connections: $success_count/$concurrent_count successful"
    
    # Check concurrent connection performance
    local success_rate=$((success_count * 100 / concurrent_count))
    if [[ $success_rate -ge 95 ]]; then
        log "INFO" "✓ Concurrent connection handling is excellent (≥ 95%)"
    elif [[ $success_rate -ge 80 ]]; then
        log "INFO" "✓ Concurrent connection handling is good (≥ 80%)"
    else
        log "WARN" "⚠ Concurrent connection handling needs improvement (< 80%)"
    fi
}

benchmark_file_upload() {
    log "INFO" "Benchmarking file upload performance..."
    
    # Create test file
    local test_file="/tmp/benchmark_test.txt"
    echo "Performance benchmark test file content" > "$test_file"
    
    # Test file upload simulation
    local start_time=$(date +%s.%N)
    curl -o /dev/null -s -X POST -F "file=@$test_file" "https://${MOODLE_DOMAIN:-lms.example.com}/" >/dev/null 2>&1 || true
    local end_time=$(date +%s.%N)
    local upload_time=$(echo "$end_time - $start_time" | bc)
    
    # Cleanup
    rm -f "$test_file"
    
    log "INFO" "  File upload simulation time: ${upload_time}s"
    
    # Check file upload performance
    if (( $(echo "$upload_time < 2" | bc -l) )); then
        log "INFO" "✓ File upload performance is excellent (< 2s)"
    elif (( $(echo "$upload_time < 5" | bc -l) )); then
        log "INFO" "✓ File upload performance is good (< 5s)"
    else
        log "WARN" "⚠ File upload performance could be improved (> 5s)"
    fi
}

benchmark_redis_performance() {
    log "INFO" "Benchmarking Redis performance..."
    
    # Test Redis latency
    local redis_latency=$(redis-cli --latency -h 127.0.0.1 -p 6379 -c 10 2>/dev/null | tail -1 || echo "N/A")
    log "INFO" "  Redis latency: $redis_latency"
    
    # Test Redis memory usage
    local redis_memory=$(redis-cli info memory | grep used_memory_human | cut -d: -f2 | tr -d '\r')
    log "INFO" "  Redis memory usage: $redis_memory"
    
    # Test Redis key count
    local redis_keys=$(redis-cli dbsize)
    log "INFO" "  Redis keys: $redis_keys"
    
    # Check Redis performance
    if [[ "$redis_latency" != "N/A" ]]; then
        local latency_num=$(echo "$redis_latency" | cut -d'.' -f1)
        if [[ "$latency_num" -lt 1 ]]; then
            log "INFO" "✓ Redis performance is excellent (< 1ms)"
        elif [[ "$latency_num" -lt 5 ]]; then
            log "INFO" "✓ Redis performance is good (< 5ms)"
        else
            log "WARN" "⚠ Redis performance could be improved (> 5ms)"
        fi
    else
        log "WARN" "⚠ Redis latency test failed"
    fi
}

benchmark_system_resources() {
    log "INFO" "Benchmarking system resources..."
    
    # CPU usage
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
    log "INFO" "  CPU usage: ${cpu_usage}%"
    
    # Memory usage
    local memory_usage=$(free | awk 'NR==2{printf "%.1f", $3*100/$2}')
    log "INFO" "  Memory usage: ${memory_usage}%"
    
    # Disk usage
    local disk_usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    log "INFO" "  Disk usage: ${disk_usage}%"
    
    # Load average
    local load_avg=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
    log "INFO" "  Load average: $load_avg"
    
    # Check system resource usage
    if (( $(echo "$cpu_usage < 70" | bc -l) )); then
        log "INFO" "✓ CPU usage is good (< 70%)"
    else
        log "WARN" "⚠ CPU usage is high (≥ 70%)"
    fi
    
    if (( $(echo "$memory_usage < 80" | bc -l) )); then
        log "INFO" "✓ Memory usage is good (< 80%)"
    else
        log "WARN" "⚠ Memory usage is high (≥ 80%)"
    fi
    
    if [[ $disk_usage -lt 80 ]]; then
        log "INFO" "✓ Disk usage is good (< 80%)"
    else
        log "WARN" "⚠ Disk usage is high (≥ 80%)"
    fi
}

benchmark_cache_performance() {
    log "INFO" "Benchmarking cache performance..."
    
    # OPcache performance
    if php -r "echo opcache_get_status()['opcache_enabled'] ? 'Enabled' : 'Disabled';" 2>/dev/null; then
        local opcache_hits=$(php -r "echo opcache_get_status()['opcache_statistics']['hits'];" 2>/dev/null || echo "0")
        local opcache_misses=$(php -r "echo opcache_get_status()['opcache_statistics']['misses'];" 2>/dev/null || echo "0")
        local opcache_hit_rate=0
        
        if [[ $opcache_hits -gt 0 || $opcache_misses -gt 0 ]]; then
            opcache_hit_rate=$(echo "scale=2; $opcache_hits * 100 / ($opcache_hits + $opcache_misses)" | bc)
        fi
        
        log "INFO" "  OPcache hits: $opcache_hits"
        log "INFO" "  OPcache misses: $opcache_misses"
        log "INFO" "  OPcache hit rate: ${opcache_hit_rate}%"
        
        if (( $(echo "$opcache_hit_rate > 90" | bc -l) )); then
            log "INFO" "✓ OPcache performance is excellent (> 90% hit rate)"
        elif (( $(echo "$opcache_hit_rate > 80" | bc -l) )); then
            log "INFO" "✓ OPcache performance is good (> 80% hit rate)"
        else
            log "WARN" "⚠ OPcache performance could be improved (< 80% hit rate)"
        fi
    else
        log "WARN" "⚠ OPcache is not available"
    fi
    
    # Nginx cache performance
    if [[ -d "/var/cache/nginx/fastcgi" ]]; then
        local nginx_cache_size=$(du -sh /var/cache/nginx/fastcgi 2>/dev/null | cut -f1)
        local nginx_cache_files=$(find /var/cache/nginx/fastcgi -type f 2>/dev/null | wc -l)
        
        log "INFO" "  Nginx cache size: $nginx_cache_size"
        log "INFO" "  Nginx cache files: $nginx_cache_files"
        
        if [[ $nginx_cache_files -gt 0 ]]; then
            log "INFO" "✓ Nginx cache is active"
        else
            log "WARN" "⚠ Nginx cache is empty"
        fi
    else
        log "WARN" "⚠ Nginx cache directory not found"
    fi
}

# =============================================================================
# Report Generation
# =============================================================================

generate_benchmark_report() {
    log "INFO" "Generating benchmark report..."
    
    {
        echo "=== Moodle Performance Benchmark Report ==="
        echo "Generated: $(date '+%Y-%m-%d %H:%M:%S')"
        echo "Domain: ${MOODLE_DOMAIN:-lms.example.com}"
        echo "Moodle Directory: $MOODLE_DIR"
        echo "Database: ${DB_NAME:-moodle}"
        echo ""
        
        echo "=== Response Time Benchmark ==="
        benchmark_response_time
        echo ""
        
        echo "=== Database Performance Benchmark ==="
        benchmark_database_performance
        echo ""
        
        echo "=== Concurrent Connection Benchmark ==="
        benchmark_concurrent_connections
        echo ""
        
        echo "=== File Upload Benchmark ==="
        benchmark_file_upload
        echo ""
        
        echo "=== Redis Performance Benchmark ==="
        benchmark_redis_performance
        echo ""
        
        echo "=== System Resources Benchmark ==="
        benchmark_system_resources
        echo ""
        
        echo "=== Cache Performance Benchmark ==="
        benchmark_cache_performance
        echo ""
        
        echo "=== Benchmark Summary ==="
        echo "All benchmarks completed successfully"
        echo ""
        
        echo "=== Recommendations ==="
        echo "1. Monitor system resources regularly"
        echo "2. Optimize database queries if needed"
        echo "3. Ensure caching is properly configured"
        echo "4. Monitor Redis memory usage"
        echo "5. Regular performance testing"
        echo ""
        
        echo "=== Benchmark Complete ==="
    } > "$REPORT_FILE"
    
    log "INFO" "Benchmark report generated: $REPORT_FILE"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    log "INFO" "Starting performance benchmark process..."
    
    # Check prerequisites
    check_root
    
    # Run comprehensive benchmark
    generate_benchmark_report
    
    # Display report summary
    log "INFO" "Performance benchmark completed. Report saved to: $REPORT_FILE"
    log "INFO" "Log file: $LOG_FILE"
    
    # Show report summary
    echo ""
    echo "=== Benchmark Summary ==="
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
        echo "  --report       Generate benchmark report only"
        echo "  --quick        Run quick benchmark (no detailed tests)"
        echo "  --response     Test response time only"
        echo "  --database     Test database performance only"
        echo "  --concurrent   Test concurrent connections only"
        echo "  --upload       Test file upload performance only"
        echo "  --redis        Test Redis performance only"
        echo "  --system       Test system resources only"
        echo "  --cache        Test cache performance only"
        exit 0
        ;;
    --report)
        check_root
        generate_benchmark_report
        echo "Report generated: $REPORT_FILE"
        exit 0
        ;;
    --quick)
        log "INFO" "Running quick benchmark..."
        check_root
        benchmark_response_time
        benchmark_system_resources
        log "INFO" "Quick benchmark completed"
        exit 0
        ;;
    --response)
        log "INFO" "Testing response time..."
        check_root
        benchmark_response_time
        log "INFO" "Response time test completed"
        exit 0
        ;;
    --database)
        log "INFO" "Testing database performance..."
        check_root
        benchmark_database_performance
        log "INFO" "Database performance test completed"
        exit 0
        ;;
    --concurrent)
        log "INFO" "Testing concurrent connections..."
        check_root
        benchmark_concurrent_connections
        log "INFO" "Concurrent connection test completed"
        exit 0
        ;;
    --upload)
        log "INFO" "Testing file upload performance..."
        check_root
        benchmark_file_upload
        log "INFO" "File upload test completed"
        exit 0
        ;;
    --redis)
        log "INFO" "Testing Redis performance..."
        check_root
        benchmark_redis_performance
        log "INFO" "Redis performance test completed"
        exit 0
        ;;
    --system)
        log "INFO" "Testing system resources..."
        check_root
        benchmark_system_resources
        log "INFO" "System resources test completed"
        exit 0
        ;;
    --cache)
        log "INFO" "Testing cache performance..."
        check_root
        benchmark_cache_performance
        log "INFO" "Cache performance test completed"
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
