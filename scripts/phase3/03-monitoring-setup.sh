#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: Phase 3 - Monitoring Setup Script
# =============================================================================
# Description: Comprehensive monitoring system setup for LMSK2-Moodle-Server
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
SCRIPT_NAME="LMSK2-Moodle-Server Monitoring Setup"
SCRIPT_VERSION="1.0"
LOG_FILE="/var/log/lmsk2-monitoring-setup.log"
CONFIG_DIR="/opt/lmsk2-moodle-server/scripts/config"
MONITORING_DIR="/opt/lmsk2-moodle-server/scripts/monitoring"

# Load configuration
if [ -f "$CONFIG_DIR/monitoring.conf" ]; then
    source "$CONFIG_DIR/monitoring.conf"
else
    echo -e "${YELLOW}Warning: Monitoring configuration file not found. Using defaults.${NC}"
fi

# Default configuration
MONITORING_ENABLE=${MONITORING_ENABLE:-"true"}
ALERT_EMAIL=${ALERT_EMAIL:-"admin@localhost"}
LOG_LEVEL=${LOG_LEVEL:-"info"}
MONITORING_INTERVAL=${MONITORING_INTERVAL:-"60"}
RETENTION_DAYS=${RETENTION_DAYS:-"30"}

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
        log "ERROR" "Monitoring setup failed. Exit code: $exit_code"
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

# Check system requirements
check_requirements() {
    log "INFO" "Checking system requirements..."
    
    # Check if required packages are installed
    local required_packages=("curl" "wget" "mailutils" "cron" "systemd")
    
    for package in "${required_packages[@]}"; do
        if ! command -v "$package" &> /dev/null; then
            log "WARN" "Package $package not found. Installing..."
            apt-get update -qq
            apt-get install -y "$package" || handle_error $? "Failed to install $package"
        fi
    done
    
    log "INFO" "System requirements check completed"
}

# =============================================================================
# Monitoring System Setup
# =============================================================================

# Create monitoring directory structure
create_monitoring_structure() {
    log "INFO" "Creating monitoring directory structure..."
    
    mkdir -p "$MONITORING_DIR"/{scripts,logs,config,templates}
    mkdir -p /var/log/lmsk2-monitoring
    mkdir -p /etc/lmsk2-monitoring
    
    # Set proper permissions
    chown -R www-data:www-data "$MONITORING_DIR"
    chmod -R 755 "$MONITORING_DIR"
    chmod -R 755 /var/log/lmsk2-monitoring
    chmod -R 755 /etc/lmsk2-monitoring
    
    log "INFO" "Monitoring directory structure created"
}

# Install monitoring tools
install_monitoring_tools() {
    log "INFO" "Installing monitoring tools..."
    
    # Update package list
    apt-get update -qq
    
    # Install essential monitoring tools
    local monitoring_packages=(
        "htop"
        "iotop"
        "nethogs"
        "iftop"
        "sysstat"
        "dstat"
        "netstat-nat"
        "lsof"
        "strace"
        "tcpdump"
        "wireshark-common"
        "tshark"
    )
    
    for package in "${monitoring_packages[@]}"; do
        log "INFO" "Installing $package..."
        apt-get install -y "$package" || log "WARN" "Failed to install $package"
    done
    
    # Install additional monitoring tools
    log "INFO" "Installing additional monitoring tools..."
    
    # Install Prometheus Node Exporter (optional)
    if [ "$INSTALL_PROMETHEUS" = "true" ]; then
        install_prometheus_node_exporter
    fi
    
    # Install Grafana (optional)
    if [ "$INSTALL_GRAFANA" = "true" ]; then
        install_grafana
    fi
    
    log "INFO" "Monitoring tools installation completed"
}

# Install Prometheus Node Exporter
install_prometheus_node_exporter() {
    log "INFO" "Installing Prometheus Node Exporter..."
    
    # Download and install Node Exporter
    local node_exporter_version="1.6.1"
    local download_url="https://github.com/prometheus/node_exporter/releases/download/v${node_exporter_version}/node_exporter-${node_exporter_version}.linux-amd64.tar.gz"
    
    cd /tmp
    wget -q "$download_url" || handle_error $? "Failed to download Node Exporter"
    tar -xzf "node_exporter-${node_exporter_version}.linux-amd64.tar.gz"
    
    # Install binary
    cp "node_exporter-${node_exporter_version}.linux-amd64/node_exporter" /usr/local/bin/
    chmod +x /usr/local/bin/node_exporter
    
    # Create systemd service
    cat > /etc/systemd/system/node_exporter.service << EOF
[Unit]
Description=Node Exporter
After=network.target

[Service]
Type=simple
User=node_exporter
Group=node_exporter
ExecStart=/usr/local/bin/node_exporter
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    # Create user
    useradd --no-create-home --shell /bin/false node_exporter
    
    # Enable and start service
    systemctl daemon-reload
    systemctl enable node_exporter
    systemctl start node_exporter
    
    log "INFO" "Prometheus Node Exporter installed and started"
}

# Install Grafana
install_grafana() {
    log "INFO" "Installing Grafana..."
    
    # Add Grafana repository
    wget -q -O - https://packages.grafana.com/gpg.key | apt-key add -
    echo "deb https://packages.grafana.com/oss/deb stable main" > /etc/apt/sources.list.d/grafana.list
    
    # Update and install
    apt-get update -qq
    apt-get install -y grafana || handle_error $? "Failed to install Grafana"
    
    # Enable and start service
    systemctl enable grafana-server
    systemctl start grafana-server
    
    log "INFO" "Grafana installed and started on port 3000"
}

# =============================================================================
# System Monitoring Setup
# =============================================================================

# Setup system monitoring
setup_system_monitoring() {
    log "INFO" "Setting up system monitoring..."
    
    # Create system monitoring script
    cat > "$MONITORING_DIR/scripts/system-monitor.sh" << 'EOF'
#!/bin/bash

# System monitoring script
LOG_FILE="/var/log/lmsk2-monitoring/system-monitor.log"
ALERT_EMAIL="admin@localhost"

# Function to log with timestamp
log_monitor() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Check CPU usage
check_cpu() {
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
    local cpu_int=${cpu_usage%.*}
    
    if [ "$cpu_int" -gt 80 ]; then
        log_monitor "WARNING: High CPU usage: ${cpu_usage}%"
        echo "High CPU usage detected: ${cpu_usage}%" | mail -s "LMSK2 Alert: High CPU Usage" "$ALERT_EMAIL"
    fi
}

# Check memory usage
check_memory() {
    local memory_usage=$(free | grep Mem | awk '{printf("%.0f", $3/$2 * 100.0)}')
    
    if [ "$memory_usage" -gt 85 ]; then
        log_monitor "WARNING: High memory usage: ${memory_usage}%"
        echo "High memory usage detected: ${memory_usage}%" | mail -s "LMSK2 Alert: High Memory Usage" "$ALERT_EMAIL"
    fi
}

# Check disk usage
check_disk() {
    local disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
    
    if [ "$disk_usage" -gt 90 ]; then
        log_monitor "WARNING: High disk usage: ${disk_usage}%"
        echo "High disk usage detected: ${disk_usage}%" | mail -s "LMSK2 Alert: High Disk Usage" "$ALERT_EMAIL"
    fi
}

# Check load average
check_load() {
    local load_avg=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
    local cpu_cores=$(nproc)
    local load_threshold=$(echo "$cpu_cores * 2" | bc)
    
    if (( $(echo "$load_avg > $load_threshold" | bc -l) )); then
        log_monitor "WARNING: High load average: $load_avg (threshold: $load_threshold)"
        echo "High load average detected: $load_avg" | mail -s "LMSK2 Alert: High Load Average" "$ALERT_EMAIL"
    fi
}

# Main monitoring function
main() {
    log_monitor "Starting system monitoring check"
    check_cpu
    check_memory
    check_disk
    check_load
    log_monitor "System monitoring check completed"
}

main "$@"
EOF

    chmod +x "$MONITORING_DIR/scripts/system-monitor.sh"
    
    # Setup cron job for system monitoring
    (crontab -l 2>/dev/null; echo "*/5 * * * * $MONITORING_DIR/scripts/system-monitor.sh") | crontab -
    
    log "INFO" "System monitoring setup completed"
}

# =============================================================================
# Application Monitoring Setup
# =============================================================================

# Setup application monitoring
setup_application_monitoring() {
    log "INFO" "Setting up application monitoring..."
    
    # Create application monitoring script
    cat > "$MONITORING_DIR/scripts/application-monitor.sh" << 'EOF'
#!/bin/bash

# Application monitoring script
LOG_FILE="/var/log/lmsk2-monitoring/application-monitor.log"
ALERT_EMAIL="admin@localhost"

# Function to log with timestamp
log_monitor() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Check Nginx status
check_nginx() {
    if ! systemctl is-active --quiet nginx; then
        log_monitor "ERROR: Nginx is not running"
        echo "Nginx service is down" | mail -s "LMSK2 Alert: Nginx Down" "$ALERT_EMAIL"
        systemctl start nginx
    else
        log_monitor "INFO: Nginx is running"
    fi
}

# Check PHP-FPM status
check_php_fpm() {
    if ! systemctl is-active --quiet php8.1-fpm; then
        log_monitor "ERROR: PHP-FPM is not running"
        echo "PHP-FPM service is down" | mail -s "LMSK2 Alert: PHP-FPM Down" "$ALERT_EMAIL"
        systemctl start php8.1-fpm
    else
        log_monitor "INFO: PHP-FPM is running"
    fi
}

# Check MariaDB status
check_mariadb() {
    if ! systemctl is-active --quiet mariadb; then
        log_monitor "ERROR: MariaDB is not running"
        echo "MariaDB service is down" | mail -s "LMSK2 Alert: MariaDB Down" "$ALERT_EMAIL"
        systemctl start mariadb
    else
        log_monitor "INFO: MariaDB is running"
    fi
}

# Check Redis status
check_redis() {
    if ! systemctl is-active --quiet redis-server; then
        log_monitor "ERROR: Redis is not running"
        echo "Redis service is down" | mail -s "LMSK2 Alert: Redis Down" "$ALERT_EMAIL"
        systemctl start redis-server
    else
        log_monitor "INFO: Redis is running"
    fi
}

# Check Moodle accessibility
check_moodle() {
    local moodle_url="http://localhost"
    local response_code=$(curl -s -o /dev/null -w "%{http_code}" "$moodle_url")
    
    if [ "$response_code" != "200" ]; then
        log_monitor "ERROR: Moodle is not accessible (HTTP $response_code)"
        echo "Moodle is not accessible (HTTP $response_code)" | mail -s "LMSK2 Alert: Moodle Down" "$ALERT_EMAIL"
    else
        log_monitor "INFO: Moodle is accessible"
    fi
}

# Main monitoring function
main() {
    log_monitor "Starting application monitoring check"
    check_nginx
    check_php_fpm
    check_mariadb
    check_redis
    check_moodle
    log_monitor "Application monitoring check completed"
}

main "$@"
EOF

    chmod +x "$MONITORING_DIR/scripts/application-monitor.sh"
    
    # Setup cron job for application monitoring
    (crontab -l 2>/dev/null; echo "*/2 * * * * $MONITORING_DIR/scripts/application-monitor.sh") | crontab -
    
    log "INFO" "Application monitoring setup completed"
}

# =============================================================================
# Log Monitoring Setup
# =============================================================================

# Setup log monitoring
setup_log_monitoring() {
    log "INFO" "Setting up log monitoring..."
    
    # Create log monitoring script
    cat > "$MONITORING_DIR/scripts/log-monitor.sh" << 'EOF'
#!/bin/bash

# Log monitoring script
LOG_FILE="/var/log/lmsk2-monitoring/log-monitor.log"
ALERT_EMAIL="admin@localhost"

# Function to log with timestamp
log_monitor() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Monitor Nginx error logs
monitor_nginx_errors() {
    local nginx_error_log="/var/log/nginx/error.log"
    local error_count=$(tail -n 100 "$nginx_error_log" | grep -c "error\|crit\|alert\|emerg" 2>/dev/null || echo "0")
    
    if [ "$error_count" -gt 10 ]; then
        log_monitor "WARNING: High number of Nginx errors: $error_count"
        echo "High number of Nginx errors detected: $error_count" | mail -s "LMSK2 Alert: Nginx Errors" "$ALERT_EMAIL"
    fi
}

# Monitor PHP error logs
monitor_php_errors() {
    local php_error_log="/var/log/php8.1-fpm.log"
    local error_count=$(tail -n 100 "$php_error_log" | grep -c "ERROR\|CRITICAL\|ALERT\|EMERGENCY" 2>/dev/null || echo "0")
    
    if [ "$error_count" -gt 5 ]; then
        log_monitor "WARNING: High number of PHP errors: $error_count"
        echo "High number of PHP errors detected: $error_count" | mail -s "LMSK2 Alert: PHP Errors" "$ALERT_EMAIL"
    fi
}

# Monitor MariaDB error logs
monitor_mariadb_errors() {
    local mariadb_error_log="/var/log/mysql/error.log"
    local error_count=$(tail -n 100 "$mariadb_error_log" | grep -c "ERROR\|CRITICAL\|ALERT\|EMERGENCY" 2>/dev/null || echo "0")
    
    if [ "$error_count" -gt 3 ]; then
        log_monitor "WARNING: High number of MariaDB errors: $error_count"
        echo "High number of MariaDB errors detected: $error_count" | mail -s "LMSK2 Alert: MariaDB Errors" "$ALERT_EMAIL"
    fi
}

# Monitor system logs for security events
monitor_security_logs() {
    local auth_log="/var/log/auth.log"
    local failed_logins=$(tail -n 100 "$auth_log" | grep -c "Failed password" 2>/dev/null || echo "0")
    
    if [ "$failed_logins" -gt 20 ]; then
        log_monitor "WARNING: High number of failed login attempts: $failed_logins"
        echo "High number of failed login attempts detected: $failed_logins" | mail -s "LMSK2 Alert: Security Threat" "$ALERT_EMAIL"
    fi
}

# Main monitoring function
main() {
    log_monitor "Starting log monitoring check"
    monitor_nginx_errors
    monitor_php_errors
    monitor_mariadb_errors
    monitor_security_logs
    log_monitor "Log monitoring check completed"
}

main "$@"
EOF

    chmod +x "$MONITORING_DIR/scripts/log-monitor.sh"
    
    # Setup cron job for log monitoring
    (crontab -l 2>/dev/null; echo "*/10 * * * * $MONITORING_DIR/scripts/log-monitor.sh") | crontab -
    
    log "INFO" "Log monitoring setup completed"
}

# =============================================================================
# Alerting System Setup
# =============================================================================

# Setup alerting system
setup_alerting_system() {
    log "INFO" "Setting up alerting system..."
    
    # Create alerting configuration
    cat > "$MONITORING_DIR/config/alerting.conf" << EOF
# LMSK2 Monitoring Alerting Configuration

# Email settings
ALERT_EMAIL="$ALERT_EMAIL"
SMTP_SERVER="localhost"
SMTP_PORT="587"
SMTP_USER=""
SMTP_PASS=""

# Alert thresholds
CPU_THRESHOLD=80
MEMORY_THRESHOLD=85
DISK_THRESHOLD=90
LOAD_THRESHOLD=2.0

# Service check intervals (in minutes)
SYSTEM_CHECK_INTERVAL=5
APPLICATION_CHECK_INTERVAL=2
LOG_CHECK_INTERVAL=10

# Log retention (in days)
LOG_RETENTION_DAYS=$RETENTION_DAYS

# Notification settings
ENABLE_EMAIL_ALERTS=true
ENABLE_SYSLOG_ALERTS=true
ENABLE_WEBHOOK_ALERTS=false
WEBHOOK_URL=""
EOF

    # Create alerting script
    cat > "$MONITORING_DIR/scripts/alerting.sh" << 'EOF'
#!/bin/bash

# Alerting system script
CONFIG_FILE="/opt/lmsk2-moodle-server/scripts/monitoring/config/alerting.conf"
LOG_FILE="/var/log/lmsk2-monitoring/alerting.log"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Function to log with timestamp
log_alert() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Send email alert
send_email_alert() {
    local subject="$1"
    local message="$2"
    
    if [ "$ENABLE_EMAIL_ALERTS" = "true" ]; then
        echo "$message" | mail -s "$subject" "$ALERT_EMAIL"
        log_alert "Email alert sent: $subject"
    fi
}

# Send syslog alert
send_syslog_alert() {
    local message="$1"
    
    if [ "$ENABLE_SYSLOG_ALERTS" = "true" ]; then
        logger -t "lmsk2-monitoring" "$message"
        log_alert "Syslog alert sent: $message"
    fi
}

# Send webhook alert
send_webhook_alert() {
    local message="$1"
    
    if [ "$ENABLE_WEBHOOK_ALERTS" = "true" ] && [ -n "$WEBHOOK_URL" ]; then
        curl -X POST -H "Content-Type: application/json" \
             -d "{\"text\":\"$message\"}" \
             "$WEBHOOK_URL" >/dev/null 2>&1
        log_alert "Webhook alert sent: $message"
    fi
}

# Main alerting function
send_alert() {
    local level="$1"
    local component="$2"
    local message="$3"
    
    local full_message="[$level] $component: $message"
    local subject="LMSK2 Alert: $component - $level"
    
    send_email_alert "$subject" "$full_message"
    send_syslog_alert "$full_message"
    send_webhook_alert "$full_message"
}

# Test alerting system
test_alerting() {
    log_alert "Testing alerting system"
    send_alert "INFO" "TEST" "This is a test alert from LMSK2 monitoring system"
    log_alert "Alerting system test completed"
}

# Main function
main() {
    case "$1" in
        "test")
            test_alerting
            ;;
        *)
            echo "Usage: $0 {test}"
            exit 1
            ;;
    esac
}

main "$@"
EOF

    chmod +x "$MONITORING_DIR/scripts/alerting.sh"
    
    # Test alerting system
    "$MONITORING_DIR/scripts/alerting.sh" test
    
    log "INFO" "Alerting system setup completed"
}

# =============================================================================
# Log Rotation Setup
# =============================================================================

# Setup log rotation
setup_log_rotation() {
    log "INFO" "Setting up log rotation..."
    
    # Create logrotate configuration
    cat > /etc/logrotate.d/lmsk2-monitoring << EOF
/var/log/lmsk2-monitoring/*.log {
    daily
    missingok
    rotate $RETENTION_DAYS
    compress
    delaycompress
    notifempty
    create 644 www-data www-data
    postrotate
        # Reload monitoring scripts if needed
        systemctl reload lmsk2-monitoring 2>/dev/null || true
    endscript
}
EOF

    # Test logrotate configuration
    logrotate -d /etc/logrotate.d/lmsk2-monitoring
    
    log "INFO" "Log rotation setup completed"
}

# =============================================================================
# Monitoring Dashboard Setup
# =============================================================================

# Setup monitoring dashboard
setup_monitoring_dashboard() {
    log "INFO" "Setting up monitoring dashboard..."
    
    # Create simple HTML dashboard
    cat > "$MONITORING_DIR/templates/dashboard.html" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LMSK2 Monitoring Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background-color: #2c3e50; color: white; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .card { background-color: white; padding: 20px; margin-bottom: 20px; border-radius: 5px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
        .status-ok { color: #27ae60; }
        .status-warning { color: #f39c12; }
        .status-error { color: #e74c3c; }
        .metric { display: inline-block; margin: 10px; padding: 10px; background-color: #ecf0f1; border-radius: 3px; }
        .refresh-btn { background-color: #3498db; color: white; padding: 10px 20px; border: none; border-radius: 3px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>LMSK2 Monitoring Dashboard</h1>
            <p>Real-time system and application monitoring</p>
        </div>
        
        <div class="card">
            <h2>System Status</h2>
            <div class="metric">
                <strong>CPU Usage:</strong> <span id="cpu-usage">Loading...</span>
            </div>
            <div class="metric">
                <strong>Memory Usage:</strong> <span id="memory-usage">Loading...</span>
            </div>
            <div class="metric">
                <strong>Disk Usage:</strong> <span id="disk-usage">Loading...</span>
            </div>
            <div class="metric">
                <strong>Load Average:</strong> <span id="load-average">Loading...</span>
            </div>
        </div>
        
        <div class="card">
            <h2>Service Status</h2>
            <div class="metric">
                <strong>Nginx:</strong> <span id="nginx-status">Loading...</span>
            </div>
            <div class="metric">
                <strong>PHP-FPM:</strong> <span id="php-fpm-status">Loading...</span>
            </div>
            <div class="metric">
                <strong>MariaDB:</strong> <span id="mariadb-status">Loading...</span>
            </div>
            <div class="metric">
                <strong>Redis:</strong> <span id="redis-status">Loading...</span>
            </div>
        </div>
        
        <div class="card">
            <h2>Actions</h2>
            <button class="refresh-btn" onclick="refreshData()">Refresh Data</button>
            <button class="refresh-btn" onclick="testAlerting()">Test Alerting</button>
        </div>
    </div>
    
    <script>
        function refreshData() {
            // This would typically fetch data from monitoring scripts
            document.getElementById('cpu-usage').textContent = '75%';
            document.getElementById('memory-usage').textContent = '60%';
            document.getElementById('disk-usage').textContent = '45%';
            document.getElementById('load-average').textContent = '1.2';
            document.getElementById('nginx-status').textContent = 'Running';
            document.getElementById('php-fpm-status').textContent = 'Running';
            document.getElementById('mariadb-status').textContent = 'Running';
            document.getElementById('redis-status').textContent = 'Running';
        }
        
        function testAlerting() {
            alert('Alerting system test initiated. Check your email and logs.');
        }
        
        // Auto-refresh every 30 seconds
        setInterval(refreshData, 30000);
        
        // Initial load
        refreshData();
    </script>
</body>
</html>
EOF

    # Create dashboard script
    cat > "$MONITORING_DIR/scripts/dashboard.sh" << 'EOF'
#!/bin/bash

# Monitoring dashboard script
DASHBOARD_DIR="/opt/lmsk2-moodle-server/scripts/monitoring/templates"
NGINX_SITES_DIR="/etc/nginx/sites-available"

# Create Nginx configuration for dashboard
cat > "$NGINX_SITES_DIR/lmsk2-monitoring" << 'EOF'
server {
    listen 8080;
    server_name localhost;
    
    root /opt/lmsk2-moodle-server/scripts/monitoring/templates;
    index dashboard.html;
    
    location / {
        try_files $uri $uri/ =404;
    }
    
    location /api/ {
        # API endpoints for monitoring data
        return 200 '{"status":"ok","message":"Monitoring API"}';
        add_header Content-Type application/json;
    }
}
EOF

# Enable the site
ln -sf "$NGINX_SITES_DIR/lmsk2-monitoring" /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx

echo "Monitoring dashboard available at: http://localhost:8080"
EOF

    chmod +x "$MONITORING_DIR/scripts/dashboard.sh"
    "$MONITORING_DIR/scripts/dashboard.sh"
    
    log "INFO" "Monitoring dashboard setup completed"
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
    
    log "INFO" "Starting monitoring setup..."
    
    # Check prerequisites
    check_root
    check_requirements
    
    # Setup monitoring system
    create_monitoring_structure
    install_monitoring_tools
    setup_system_monitoring
    setup_application_monitoring
    setup_log_monitoring
    setup_alerting_system
    setup_log_rotation
    setup_monitoring_dashboard
    
    # Final verification
    log "INFO" "Verifying monitoring setup..."
    
    # Check if monitoring scripts are executable
    local monitoring_scripts=(
        "$MONITORING_DIR/scripts/system-monitor.sh"
        "$MONITORING_DIR/scripts/application-monitor.sh"
        "$MONITORING_DIR/scripts/log-monitor.sh"
        "$MONITORING_DIR/scripts/alerting.sh"
    )
    
    for script in "${monitoring_scripts[@]}"; do
        if [ -x "$script" ]; then
            log "INFO" "✓ $script is executable"
        else
            log "ERROR" "✗ $script is not executable"
        fi
    done
    
    # Check cron jobs
    if crontab -l | grep -q "lmsk2-monitoring"; then
        log "INFO" "✓ Monitoring cron jobs are configured"
    else
        log "WARN" "✗ Monitoring cron jobs not found"
    fi
    
    # Display summary
    echo
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Monitoring Setup Completed Successfully!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo
    echo -e "${WHITE}Monitoring Components Installed:${NC}"
    echo -e "  • System monitoring (CPU, Memory, Disk, Load)"
    echo -e "  • Application monitoring (Nginx, PHP-FPM, MariaDB, Redis)"
    echo -e "  • Log monitoring (Error logs, Security events)"
    echo -e "  • Alerting system (Email, Syslog, Webhook)"
    echo -e "  • Log rotation and retention"
    echo -e "  • Monitoring dashboard (http://localhost:8080)"
    echo
    echo -e "${WHITE}Configuration Files:${NC}"
    echo -e "  • $MONITORING_DIR/config/alerting.conf"
    echo -e "  • /etc/logrotate.d/lmsk2-monitoring"
    echo
    echo -e "${WHITE}Log Files:${NC}"
    echo -e "  • $LOG_FILE"
    echo -e "  • /var/log/lmsk2-monitoring/"
    echo
    echo -e "${WHITE}Next Steps:${NC}"
    echo -e "  1. Configure email settings in alerting.conf"
    echo -e "  2. Test monitoring system: $MONITORING_DIR/scripts/alerting.sh test"
    echo -e "  3. Access dashboard: http://localhost:8080"
    echo -e "  4. Monitor logs: tail -f /var/log/lmsk2-monitoring/*.log"
    echo
    
    log "INFO" "Monitoring setup completed successfully"
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
            echo "  --config FILE       Use custom configuration file"
            echo "  --email EMAIL       Set alert email address"
            echo "  --interval MINUTES  Set monitoring interval"
            echo "  --retention DAYS    Set log retention days"
            echo
            exit 0
            ;;
        --version|-v)
            echo "$SCRIPT_NAME v$SCRIPT_VERSION"
            exit 0
            ;;
        --config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        --email)
            ALERT_EMAIL="$2"
            shift 2
            ;;
        --interval)
            MONITORING_INTERVAL="$2"
            shift 2
            ;;
        --retention)
            RETENTION_DAYS="$2"
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

