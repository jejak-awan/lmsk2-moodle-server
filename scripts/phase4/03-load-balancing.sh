#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: Phase 4 - Load Balancing Script
# =============================================================================
# Description: Load balancing setup for LMSK2-Moodle-Server
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
SCRIPT_NAME="LMSK2-Moodle-Server Load Balancing"
SCRIPT_VERSION="1.0"
LOG_FILE="/var/log/lmsk2-load-balancing.log"
CONFIG_DIR="/opt/lmsk2-moodle-server/scripts/config"
LOAD_BALANCER_DIR="/opt/lmsk2-moodle-server/scripts/load-balancer"

# Load configuration
if [ -f "$CONFIG_DIR/load-balancer.conf" ]; then
    source "$CONFIG_DIR/load-balancer.conf"
else
    echo -e "${YELLOW}Warning: Load balancer configuration file not found. Using defaults.${NC}"
fi

# Default configuration
LOAD_BALANCER_ENABLE=${LOAD_BALANCER_ENABLE:-"true"}
BACKEND_SERVERS=${BACKEND_SERVERS:-"127.0.0.1:8080"}
BALANCING_METHOD=${BALANCING_METHOD:-"round_robin"}
HEALTH_CHECK_ENABLE=${HEALTH_CHECK_ENABLE:-"true"}
SESSION_PERSISTENCE=${SESSION_PERSISTENCE:-"true"}
SSL_TERMINATION=${SSL_TERMINATION:-"true"}

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
            if [ "$LOG_LEVEL" = "debug" ]; then
                echo -e "${BLUE}[DEBUG]${NC} $message" | tee -a "$LOG_FILE"
            fi
            ;;
    esac
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# Error handling
handle_error() {
    local exit_code=$1
    local error_message=$2
    
    if [ $exit_code -ne 0 ]; then
        log "ERROR" "$error_message"
        log "ERROR" "Load balancing setup failed. Exit code: $exit_code"
        exit $exit_code
    fi
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log "ERROR" "This script must be run as root"
        exit 1
    fi
}

# =============================================================================
# Load Balancer Configuration
# =============================================================================

# Create load balancer configuration
create_load_balancer_config() {
    log "INFO" "Creating load balancer configuration..."
    
    cat > "$CONFIG_DIR/load-balancer.conf" << EOF
# LMSK2 Load Balancer Configuration

# Load balancer settings
LOAD_BALANCER_ENABLE=true
BALANCING_METHOD=round_robin
HEALTH_CHECK_ENABLE=true
SESSION_PERSISTENCE=true
SSL_TERMINATION=true

# Backend servers
BACKEND_SERVERS="127.0.0.1:8080 127.0.0.1:8081 127.0.0.1:8082"
BACKEND_WEIGHTS="1 1 1"
BACKEND_MAX_FAILS=3
BACKEND_FAIL_TIMEOUT=30s

# Health check settings
HEALTH_CHECK_INTERVAL=10s
HEALTH_CHECK_TIMEOUT=5s
HEALTH_CHECK_PATH="/health"
HEALTH_CHECK_METHOD="GET"

# Session persistence
SESSION_COOKIE_NAME="LMSK2_SESSION"
SESSION_COOKIE_DOMAIN=".yourdomain.com"
SESSION_COOKIE_PATH="/"
SESSION_COOKIE_EXPIRES="1h"

# SSL settings
SSL_CERT_PATH="/etc/letsencrypt/live/yourdomain.com/fullchain.pem"
SSL_KEY_PATH="/etc/letsencrypt/live/yourdomain.com/privkey.pem"
SSL_PROTOCOLS="TLSv1.2 TLSv1.3"

# Load balancing algorithms
ROUND_ROBIN=true
LEAST_CONNECTIONS=true
IP_HASH=true
WEIGHTED_ROUND_ROBIN=true

# Monitoring settings
LOAD_BALANCER_MONITORING=true
STATS_ENABLE=true
STATS_PORT=8080
STATS_PATH="/nginx_status"

# Rate limiting
RATE_LIMIT_ENABLE=true
RATE_LIMIT_ZONE="api"
RATE_LIMIT_RATE="10r/s"
RATE_LIMIT_BURST=20

# Caching
CACHE_ENABLE=true
CACHE_ZONE="moodle_cache"
CACHE_SIZE="1g"
CACHE_INACTIVE="60m"
EOF

    log "INFO" "Load balancer configuration created"
}

# =============================================================================
# Nginx Load Balancer Setup
# =============================================================================

# Setup Nginx load balancer
setup_nginx_load_balancer() {
    log "INFO" "Setting up Nginx load balancer..."
    
    # Create upstream configuration
    cat > /etc/nginx/conf.d/lmsk2-upstream.conf << EOF
# LMSK2 Upstream Configuration

# Define upstream servers
upstream lmsk2_backend {
    # Load balancing method
    $BALANCING_METHOD;
    
    # Backend servers
    server 127.0.0.1:8080 weight=1 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8081 weight=1 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8082 weight=1 max_fails=3 fail_timeout=30s;
    
    # Health check
    keepalive 32;
    keepalive_requests 100;
    keepalive_timeout 60s;
}

# Health check upstream
upstream lmsk2_health {
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
}
EOF

    # Create load balancer configuration
    cat > /etc/nginx/sites-available/lmsk2-load-balancer << EOF
# LMSK2 Load Balancer Configuration

# Rate limiting
limit_req_zone \$binary_remote_addr zone=api:10m rate=10r/s;
limit_req_zone \$binary_remote_addr zone=login:10m rate=5r/s;

# Cache configuration
proxy_cache_path /var/cache/nginx/lmsk2 levels=1:2 keys_zone=moodle_cache:10m max_size=1g inactive=60m use_temp_path=off;

# Main load balancer server
server {
    listen 80;
    server_name yourdomain.com;
    
    # Redirect to HTTPS
    return 301 https://\$server_name\$request_uri;
}

# HTTPS load balancer server
server {
    listen 443 ssl http2;
    server_name yourdomain.com;
    
    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    include /etc/nginx/snippets/ssl-params.conf;
    
    # Security headers
    include /etc/nginx/snippets/security-headers.conf;
    
    # Rate limiting
    limit_req zone=api burst=20 nodelay;
    
    # Health check endpoint
    location /health {
        access_log off;
        proxy_pass http://lmsk2_health;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    # Nginx status
    location /nginx_status {
        stub_status on;
        access_log off;
        allow 127.0.0.1;
        allow 10.0.0.0/8;
        allow 172.16.0.0/12;
        allow 192.168.0.0/16;
        deny all;
    }
    
    # Main application
    location / {
        # Session persistence
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header X-Forwarded-Host \$host;
        proxy_set_header X-Forwarded-Port \$server_port;
        
        # Load balancing
        proxy_pass http://lmsk2_backend;
        
        # Caching
        proxy_cache moodle_cache;
        proxy_cache_valid 200 302 10m;
        proxy_cache_valid 404 1m;
        proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504;
        proxy_cache_lock on;
        proxy_cache_lock_timeout 5s;
        
        # Timeouts
        proxy_connect_timeout 5s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Buffering
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        proxy_busy_buffers_size 8k;
        
        # Headers
        proxy_hide_header X-Powered-By;
        proxy_hide_header Server;
    }
    
    # Login rate limiting
    location /login {
        limit_req zone=login burst=10 nodelay;
        
        proxy_pass http://lmsk2_backend;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    # Static files caching
    location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://lmsk2_backend;
        proxy_cache moodle_cache;
        proxy_cache_valid 200 1y;
        expires 1y;
        add_header Cache-Control "public, immutable";
        add_header Vary Accept-Encoding;
    }
    
    # API endpoints
    location /api/ {
        limit_req zone=api burst=50 nodelay;
        
        proxy_pass http://lmsk2_backend;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # No caching for API
        proxy_cache off;
    }
}
EOF

    # Enable the load balancer site
    ln -sf /etc/nginx/sites-available/lmsk2-load-balancer /etc/nginx/sites-enabled/
    
    # Test Nginx configuration
    nginx -t || handle_error $? "Nginx load balancer configuration test failed"
    
    # Reload Nginx
    systemctl reload nginx
    
    log "INFO" "Nginx load balancer setup completed"
}

# =============================================================================
# Health Check System
# =============================================================================

# Setup health check system
setup_health_check_system() {
    log "INFO" "Setting up health check system..."
    
    # Create health check script
    cat > "$LOAD_BALANCER_DIR/scripts/health-check.sh" << 'EOF'
#!/bin/bash

# Load Balancer Health Check Script
CONFIG_FILE="/opt/lmsk2-moodle-server/scripts/config/load-balancer.conf"
LOG_FILE="/var/log/lmsk2-load-balancer/health-check.log"
ALERT_EMAIL="admin@localhost"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Function to log with timestamp
log_health() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Check backend server health
check_backend_health() {
    local server="$1"
    local health_url="http://$server$HEALTH_CHECK_PATH"
    
    log_health "Checking health of server: $server"
    
    # Check if server is responding
    local response_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 --max-time 10 "$health_url" 2>/dev/null)
    
    if [ "$response_code" = "200" ]; then
        log_health "Server $server is healthy (HTTP $response_code)"
        return 0
    else
        log_health "Server $server is unhealthy (HTTP $response_code)"
        return 1
    fi
}

# Check all backend servers
check_all_backends() {
    local unhealthy_servers=()
    local total_servers=0
    local healthy_servers=0
    
    # Parse backend servers
    IFS=' ' read -ra SERVERS <<< "$BACKEND_SERVERS"
    
    for server in "${SERVERS[@]}"; do
        total_servers=$((total_servers + 1))
        
        if check_backend_health "$server"; then
            healthy_servers=$((healthy_servers + 1))
        else
            unhealthy_servers+=("$server")
        fi
    done
    
    log_health "Health check summary: $healthy_servers/$total_servers servers healthy"
    
    # Alert if too many servers are unhealthy
    local unhealthy_count=${#unhealthy_servers[@]}
    if [ "$unhealthy_count" -gt 0 ]; then
        log_health "WARNING: $unhealthy_count servers are unhealthy: ${unhealthy_servers[*]}"
        
        # Send alert if more than half are unhealthy
        if [ "$unhealthy_count" -gt $((total_servers / 2)) ]; then
            if [ -n "$ALERT_EMAIL" ]; then
                echo "Load balancer health check failed: $unhealthy_count/$total_servers servers unhealthy" | mail -s "Load Balancer Health Alert" "$ALERT_EMAIL"
            fi
        fi
    fi
    
    return $unhealthy_count
}

# Update Nginx upstream configuration
update_upstream_config() {
    local unhealthy_servers=("$@")
    
    if [ ${#unhealthy_servers[@]} -gt 0 ]; then
        log_health "Updating upstream configuration to remove unhealthy servers"
        
        # Create temporary upstream configuration
        cat > /tmp/lmsk2-upstream-temp.conf << EOF
# LMSK2 Upstream Configuration (Updated)

upstream lmsk2_backend {
    $BALANCING_METHOD;
EOF
        
        # Add only healthy servers
        IFS=' ' read -ra SERVERS <<< "$BACKEND_SERVERS"
        for server in "${SERVERS[@]}"; do
            local is_unhealthy=false
            for unhealthy in "${unhealthy_servers[@]}"; do
                if [ "$server" = "$unhealthy" ]; then
                    is_unhealthy=true
                    break
                fi
            done
            
            if [ "$is_unhealthy" = false ]; then
                echo "    server $server weight=1 max_fails=3 fail_timeout=30s;" >> /tmp/lmsk2-upstream-temp.conf
            fi
        done
        
        echo "    keepalive 32;" >> /tmp/lmsk2-upstream-temp.conf
        echo "    keepalive_requests 100;" >> /tmp/lmsk2-upstream-temp.conf
        echo "    keepalive_timeout 60s;" >> /tmp/lmsk2-upstream-temp.conf
        echo "}" >> /tmp/lmsk2-upstream-temp.conf
        
        # Replace upstream configuration
        cp /tmp/lmsk2-upstream-temp.conf /etc/nginx/conf.d/lmsk2-upstream.conf
        rm /tmp/lmsk2-upstream-temp.conf
        
        # Test and reload Nginx
        if nginx -t; then
            systemctl reload nginx
            log_health "Nginx configuration updated and reloaded"
        else
            log_health "ERROR: Nginx configuration test failed"
        fi
    fi
}

# Main health check function
main() {
    log_health "Starting load balancer health check"
    
    # Check all backend servers
    if check_all_backends; then
        log_health "All backend servers are healthy"
        exit 0
    else
        # Get unhealthy servers
        local unhealthy_servers=()
        IFS=' ' read -ra SERVERS <<< "$BACKEND_SERVERS"
        
        for server in "${SERVERS[@]}"; do
            if ! check_backend_health "$server"; then
                unhealthy_servers+=("$server")
            fi
        done
        
        # Update upstream configuration
        update_upstream_config "${unhealthy_servers[@]}"
        
        log_health "Health check completed with issues"
        exit 1
    fi
}

# Run main function
main "$@"
EOF

    chmod +x "$LOAD_BALANCER_DIR/scripts/health-check.sh"
    
    # Setup health check cron job
    (crontab -l 2>/dev/null; echo "*/30 * * * * $LOAD_BALANCER_DIR/scripts/health-check.sh") | crontab -
    
    log "INFO" "Health check system setup completed"
}

# =============================================================================
# Session Persistence
# =============================================================================

# Setup session persistence
setup_session_persistence() {
    log "INFO" "Setting up session persistence..."
    
    # Create session persistence configuration
    cat > /etc/nginx/conf.d/lmsk2-session.conf << EOF
# LMSK2 Session Persistence Configuration

# Session cookie configuration
map \$cookie_LMSK2_SESSION \$backend_pool {
    default lmsk2_backend;
    ~^(.+)$ lmsk2_backend;
}

# Session persistence upstream
upstream lmsk2_backend {
    $BALANCING_METHOD;
    
    # Backend servers with session affinity
    server 127.0.0.1:8080 weight=1 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8081 weight=1 max_fails=3 fail_timeout=30s;
    server 127.0.0.1:8082 weight=1 max_fails=3 fail_timeout=30s;
    
    # Session persistence settings
    ip_hash;
    keepalive 32;
    keepalive_requests 100;
    keepalive_timeout 60s;
}
EOF

    # Create session management script
    cat > "$LOAD_BALANCER_DIR/scripts/session-manager.sh" << 'EOF'
#!/bin/bash

# Session Management Script
CONFIG_FILE="/opt/lmsk2-moodle-server/scripts/config/load-balancer.conf"
LOG_FILE="/var/log/lmsk2-load-balancer/session-manager.log"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Function to log with timestamp
log_session() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Clean expired sessions
clean_expired_sessions() {
    log_session "Cleaning expired sessions"
    
    # Clean Redis sessions
    if command -v redis-cli >/dev/null 2>&1; then
        local expired_sessions=$(redis-cli --scan --pattern "session:*" | wc -l)
        redis-cli --scan --pattern "session:*" | xargs -r redis-cli del
        log_session "Cleaned $expired_sessions expired Redis sessions"
    fi
    
    # Clean file-based sessions
    if [ -d "/var/lib/php/sessions" ]; then
        local expired_files=$(find /var/lib/php/sessions -type f -mmin +60 | wc -l)
        find /var/lib/php/sessions -type f -mmin +60 -delete
        log_session "Cleaned $expired_files expired file sessions"
    fi
}

# Monitor session usage
monitor_session_usage() {
    log_session "Monitoring session usage"
    
    # Check Redis session count
    if command -v redis-cli >/dev/null 2>&1; then
        local redis_sessions=$(redis-cli dbsize)
        log_session "Redis sessions: $redis_sessions"
    fi
    
    # Check file session count
    if [ -d "/var/lib/php/sessions" ]; then
        local file_sessions=$(find /var/lib/php/sessions -type f | wc -l)
        log_session "File sessions: $file_sessions"
    fi
}

# Main session management function
main() {
    log_session "Starting session management"
    
    clean_expired_sessions
    monitor_session_usage
    
    log_session "Session management completed"
}

# Run main function
main "$@"
EOF

    chmod +x "$LOAD_BALANCER_DIR/scripts/session-manager.sh"
    
    # Setup session management cron job
    (crontab -l 2>/dev/null; echo "0 */6 * * * $LOAD_BALANCER_DIR/scripts/session-manager.sh") | crontab -
    
    log "INFO" "Session persistence setup completed"
}

# =============================================================================
# Load Balancer Monitoring
# =============================================================================

# Setup load balancer monitoring
setup_load_balancer_monitoring() {
    log "INFO" "Setting up load balancer monitoring..."
    
    # Create load balancer monitoring script
    cat > "$LOAD_BALANCER_DIR/scripts/load-balancer-monitor.sh" << 'EOF'
#!/bin/bash

# Load Balancer Monitoring Script
CONFIG_FILE="/opt/lmsk2-moodle-server/scripts/config/load-balancer.conf"
LOG_FILE="/var/log/lmsk2-load-balancer/load-balancer-monitor.log"
ALERT_EMAIL="admin@localhost"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Function to log with timestamp
log_monitor() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Monitor load balancer statistics
monitor_load_balancer_stats() {
    log_monitor "Monitoring load balancer statistics"
    
    # Get Nginx status
    local nginx_status=$(curl -s http://localhost/nginx_status 2>/dev/null)
    
    if [ -n "$nginx_status" ]; then
        local active_connections=$(echo "$nginx_status" | grep "Active connections" | awk '{print $3}')
        local server_handled=$(echo "$nginx_status" | grep "server accepts handled requests" | awk '{print $4}')
        local server_requests=$(echo "$nginx_status" | grep "server accepts handled requests" | awk '{print $5}')
        
        log_monitor "Active connections: $active_connections"
        log_monitor "Server handled: $server_handled"
        log_monitor "Server requests: $server_requests"
        
        # Check for high connection count
        if [ "$active_connections" -gt 1000 ]; then
            log_monitor "WARNING: High number of active connections: $active_connections"
            if [ -n "$ALERT_EMAIL" ]; then
                echo "High number of active connections: $active_connections" | mail -s "Load Balancer High Connections" "$ALERT_EMAIL"
            fi
        fi
    else
        log_monitor "ERROR: Unable to get Nginx status"
    fi
}

# Monitor backend server performance
monitor_backend_performance() {
    log_monitor "Monitoring backend server performance"
    
    # Parse backend servers
    IFS=' ' read -ra SERVERS <<< "$BACKEND_SERVERS"
    
    for server in "${SERVERS[@]}"; do
        local response_time=$(curl -o /dev/null -s -w "%{time_total}" "http://$server$HEALTH_CHECK_PATH" 2>/dev/null)
        
        if [ -n "$response_time" ]; then
            log_monitor "Server $server response time: ${response_time}s"
            
            # Check for slow response
            if (( $(echo "$response_time > 2.0" | bc -l) )); then
                log_monitor "WARNING: Slow response from server $server: ${response_time}s"
            fi
        else
            log_monitor "ERROR: Unable to get response time from server $server"
        fi
    done
}

# Monitor cache performance
monitor_cache_performance() {
    log_monitor "Monitoring cache performance"
    
    # Check cache directory size
    if [ -d "/var/cache/nginx/lmsk2" ]; then
        local cache_size=$(du -sh /var/cache/nginx/lmsk2 | awk '{print $1}')
        local cache_files=$(find /var/cache/nginx/lmsk2 -type f | wc -l)
        
        log_monitor "Cache size: $cache_size"
        log_monitor "Cache files: $cache_files"
        
        # Check for large cache
        local cache_size_mb=$(du -sm /var/cache/nginx/lmsk2 | awk '{print $1}')
        if [ "$cache_size_mb" -gt 1000 ]; then
            log_monitor "WARNING: Large cache size: ${cache_size_mb}MB"
        fi
    fi
}

# Main monitoring function
main() {
    log_monitor "Starting load balancer monitoring"
    
    monitor_load_balancer_stats
    monitor_backend_performance
    monitor_cache_performance
    
    log_monitor "Load balancer monitoring completed"
}

# Run main function
main "$@"
EOF

    chmod +x "$LOAD_BALANCER_DIR/scripts/load-balancer-monitor.sh"
    
    # Setup monitoring cron job
    (crontab -l 2>/dev/null; echo "*/5 * * * * $LOAD_BALANCER_DIR/scripts/load-balancer-monitor.sh") | crontab -
    
    log "INFO" "Load balancer monitoring setup completed"
}

# =============================================================================
# Main Installation Function
# =============================================================================

# Main installation function
main() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  $SCRIPT_NAME v$SCRIPT_VERSION${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo
    
    log "INFO" "Starting load balancing setup..."
    
    # Check prerequisites
    check_root
    
    # Create load balancer directory
    mkdir -p "$LOAD_BALANCER_DIR"/{scripts,config,logs}
    mkdir -p /var/log/lmsk2-load-balancer
    
    # Setup load balancing
    create_load_balancer_config
    setup_nginx_load_balancer
    setup_health_check_system
    setup_session_persistence
    setup_load_balancer_monitoring
    
    # Final verification
    log "INFO" "Verifying load balancing setup..."
    
    # Check Nginx configuration
    if nginx -t >/dev/null 2>&1; then
        log "INFO" "✓ Nginx load balancer configuration is valid"
    else
        log "ERROR" "✗ Nginx load balancer configuration is invalid"
    fi
    
    # Check upstream configuration
    if [ -f "/etc/nginx/conf.d/lmsk2-upstream.conf" ]; then
        log "INFO" "✓ Upstream configuration created"
    else
        log "ERROR" "✗ Upstream configuration not found"
    fi
    
    # Check health check script
    if [ -f "$LOAD_BALANCER_DIR/scripts/health-check.sh" ]; then
        log "INFO" "✓ Health check script created"
    else
        log "ERROR" "✗ Health check script not found"
    fi
    
    # Display summary
    echo
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Load Balancing Setup Completed!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo
    echo -e "${WHITE}Load Balancer Components Installed:${NC}"
    echo -e "  • Nginx load balancer"
    echo -e "  • Health check system"
    echo -e "  • Session persistence"
    echo -e "  • Rate limiting"
    echo -e "  • Caching"
    echo -e "  • SSL termination"
    echo -e "  • Monitoring system"
    echo
    echo -e "${WHITE}Configuration Files:${NC}"
    echo -e "  • $CONFIG_DIR/load-balancer.conf"
    echo -e "  • /etc/nginx/conf.d/lmsk2-upstream.conf"
    echo -e "  • /etc/nginx/sites-available/lmsk2-load-balancer"
    echo -e "  • /etc/nginx/conf.d/lmsk2-session.conf"
    echo
    echo -e "${WHITE}Load Balancing Features:${NC}"
    echo -e "  • Round-robin balancing"
    echo -e "  • Health checks"
    echo -e "  • Session persistence"
    echo -e "  • Rate limiting"
    echo -e "  • SSL termination"
    echo -e "  • Caching"
    echo -e "  • Monitoring"
    echo
    echo -e "${WHITE}Backend Servers:${NC}"
    echo -e "  • 127.0.0.1:8080"
    echo -e "  • 127.0.0.1:8081"
    echo -e "  • 127.0.0.1:8082"
    echo
    echo -e "${WHITE}Next Steps:${NC}"
    echo -e "  1. Configure backend servers"
    echo -e "  2. Update domain in configuration"
    echo -e "  3. Test load balancing"
    echo -e "  4. Monitor performance"
    echo -e "  5. Adjust balancing algorithms"
    echo -e "  6. Configure SSL certificates"
    echo
    
    log "INFO" "Load balancing setup completed successfully"
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
            echo "  --backend SERVERS   Set backend servers"
            echo "  --method METHOD     Set balancing method"
            echo "  --health-check      Enable/disable health checks"
            echo "  --session-persistence Enable/disable session persistence"
            echo "  --ssl-termination   Enable/disable SSL termination"
            echo
            exit 0
            ;;
        --version|-v)
            echo "$SCRIPT_NAME v$SCRIPT_VERSION"
            exit 0
            ;;
        --backend)
            BACKEND_SERVERS="$2"
            shift 2
            ;;
        --method)
            BALANCING_METHOD="$2"
            shift 2
            ;;
        --health-check)
            HEALTH_CHECK_ENABLE="$2"
            shift 2
            ;;
        --session-persistence)
            SESSION_PERSISTENCE="$2"
            shift 2
            ;;
        --ssl-termination)
            SSL_TERMINATION="$2"
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

