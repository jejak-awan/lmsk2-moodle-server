#!/bin/bash

# =============================================================================
# LMSK2 Moodle Server - Basic Configuration Script
# =============================================================================
# Description: Basic system configuration for Moodle server
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
CONFIG_FILE="${SCRIPT_DIR}/config/phase1.conf"

if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
else
    echo "❌ Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Script configuration
SCRIPT_NAME="$(basename "$0")"
LOG_FILE="/var/log/lmsk2/${SCRIPT_NAME%.*}.log"
BACKUP_DIR="/backup/lmsk2/basic-config"

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

create_directories() {
    log "INFO" "Creating necessary directories..."
    
    mkdir -p "$BACKUP_DIR"
    mkdir -p "/var/log/moodle"
    mkdir -p "/backup/moodle"
    mkdir -p "/usr/local/bin"
    
    log "INFO" "Directories created successfully"
}

backup_config() {
    local config_file="$1"
    local backup_file="$BACKUP_DIR/$(basename "$config_file").backup.$(date +%Y%m%d_%H%M%S)"
    
    if [[ -f "$config_file" ]]; then
        cp "$config_file" "$backup_file"
        log "INFO" "Backed up $config_file to $backup_file"
    fi
}

# =============================================================================
# Configuration Functions
# =============================================================================

configure_kernel_optimization() {
    log "INFO" "Configuring kernel optimization..."
    
    # Backup existing sysctl configuration
    backup_config "/etc/sysctl.conf"
    
    # Create kernel optimization configuration
    cat > /etc/sysctl.d/99-moodle-optimization.conf << 'EOF'
# Network optimizations
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.core.rmem_default = 262144
net.core.wmem_default = 262144
net.ipv4.tcp_rmem = 4096 65536 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
net.ipv4.tcp_congestion_control = bbr
net.ipv4.tcp_slow_start_after_idle = 0
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.tcp_keepalive_intvl = 30
net.ipv4.tcp_keepalive_probes = 3

# File system optimizations
fs.file-max = 65536
fs.inotify.max_user_watches = 524288
fs.inotify.max_user_instances = 256
fs.inotify.max_queued_events = 32768

# Memory management
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
vm.vfs_cache_pressure = 50

# Process limits
kernel.pid_max = 4194304
kernel.threads-max = 2097152
EOF

    # Apply kernel parameters
    sysctl -p /etc/sysctl.d/99-moodle-optimization.conf
    
    # Verify changes
    log "INFO" "Verifying kernel parameters..."
    sysctl net.core.rmem_max
    sysctl vm.swappiness
    
    log "INFO" "Kernel optimization completed"
}

configure_system_limits() {
    log "INFO" "Configuring system limits..."
    
    # Create system limits configuration
    cat > /etc/security/limits.d/99-moodle.conf << 'EOF'
# Moodle user limits
moodle soft nofile 65536
moodle hard nofile 65536
moodle soft nproc 32768
moodle hard nproc 32768

# www-data user limits
www-data soft nofile 65536
www-data hard nofile 65536
www-data soft nproc 32768
www-data hard nproc 32768

# Root limits
root soft nofile 65536
root hard nofile 65536
root soft nproc 32768
root hard nproc 32768
EOF

    log "INFO" "System limits configuration completed"
}

configure_cron_jobs() {
    log "INFO" "Configuring cron jobs..."
    
    # Create system maintenance cron jobs
    (crontab -l 2>/dev/null; cat << 'EOF'
# System maintenance
0 2 * * * /usr/bin/apt update && /usr/bin/apt upgrade -y
0 3 * * 0 /usr/bin/apt autoremove -y && /usr/bin/apt autoclean

# Log rotation and cleanup
0 1 * * * /usr/sbin/logrotate /etc/logrotate.conf
0 4 * * * find /var/log -name "*.log" -mtime +30 -delete

# Security updates
0 5 * * * /usr/bin/unattended-upgrades

# System monitoring
*/15 * * * * /usr/local/bin/system-health-check.sh

# Backup verification
0 6 * * * /usr/local/bin/backup-verification.sh

# Daily backup at 2 AM
0 2 * * * /usr/local/bin/moodle-backup.sh

# Performance monitoring every 5 minutes
*/5 * * * * /usr/local/bin/performance-monitor.sh

# Security monitoring every hour
0 * * * * /usr/local/bin/security-monitor.sh
EOF
    ) | crontab -
    
    log "INFO" "Cron jobs configuration completed"
}

setup_system_health_monitoring() {
    log "INFO" "Setting up system health monitoring..."
    
    # Create system health check script
    cat > /usr/local/bin/system-health-check.sh << 'EOF'
#!/bin/bash

# System health monitoring script
LOG_FILE="/var/log/system-health.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')
ALERT_EMAIL="admin@example.com"

echo "[$DATE] Starting system health check..." >> $LOG_FILE

# Check disk space
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    echo "[$DATE] ALERT: Disk usage critical: $DISK_USAGE%" >> $LOG_FILE
    # Send alert email (if mail is configured)
    echo "Disk usage is at $DISK_USAGE%" | mail -s "Disk Usage Alert" $ALERT_EMAIL 2>/dev/null || true
fi

# Check memory usage
MEMORY_USAGE=$(free | awk 'NR==2{printf "%.2f", $3*100/$2}')
if (( $(echo "$MEMORY_USAGE > 90" | bc -l) )); then
    echo "[$DATE] ALERT: Memory usage high: $MEMORY_USAGE%" >> $LOG_FILE
fi

# Check CPU load
CPU_LOAD=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
if (( $(echo "$CPU_LOAD > 5" | bc -l) )); then
    echo "[$DATE] ALERT: CPU load high: $CPU_LOAD" >> $LOG_FILE
fi

# Check service status
SERVICES=("nginx" "php8.1-fpm" "mariadb" "redis-server")
for service in "${SERVICES[@]}"; do
    if ! systemctl is-active --quiet $service; then
        echo "[$DATE] ALERT: Service $service is not running" >> $LOG_FILE
        systemctl restart $service
    fi
done

# Check database connectivity
if ! mysql -u root -p'${DB_ROOT_PASSWORD:-}' -e "SELECT 1;" > /dev/null 2>&1; then
    echo "[$DATE] ALERT: Database connection failed" >> $LOG_FILE
fi

# Check Redis connectivity
if ! redis-cli ping > /dev/null 2>&1; then
    echo "[$DATE] ALERT: Redis connection failed" >> $LOG_FILE
fi

echo "[$DATE] System health check completed." >> $LOG_FILE
EOF

    # Make script executable
    chmod +x /usr/local/bin/system-health-check.sh
    
    log "INFO" "System health monitoring setup completed"
}

setup_backup_system() {
    log "INFO" "Setting up backup system..."
    
    # Create backup directory
    mkdir -p /backup/moodle
    chown root:root /backup/moodle
    chmod 700 /backup/moodle
    
    # Create Moodle backup script
    cat > /usr/local/bin/moodle-backup.sh << 'EOF'
#!/bin/bash

# Moodle backup script
BACKUP_DIR="/backup/moodle"
DATE=$(date +%Y%m%d_%H%M%S)
MOODLE_DIR="/var/www/moodle"
MOODLE_DATA="/var/www/moodle/moodledata"
DB_NAME="moodle"
DB_USER="moodle"
DB_PASS="${DB_PASSWORD:-}"

# Create backup directory
mkdir -p $BACKUP_DIR

echo "Starting Moodle backup at $(date)"

# Database backup
echo "Backing up database..."
mysqldump -u $DB_USER -p$DB_PASS $DB_NAME | gzip > $BACKUP_DIR/moodle_db_$DATE.sql.gz

# Files backup
echo "Backing up Moodle files..."
tar -czf $BACKUP_DIR/moodle_files_$DATE.tar.gz -C /var/www moodle

# Moodle data backup
echo "Backing up Moodle data..."
tar -czf $BACKUP_DIR/moodle_data_$DATE.tar.gz -C /var/www moodle/moodledata

# Configuration backup
echo "Backing up configuration files..."
tar -czf $BACKUP_DIR/moodle_config_$DATE.tar.gz /etc/nginx/sites-available/moodle /etc/php/8.1/fpm/php.ini /etc/mysql/mariadb.conf.d/

# Cleanup old backups (keep 7 days)
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete

echo "Backup completed at $(date)"
echo "Backup files:"
ls -lh $BACKUP_DIR/*$DATE*
EOF

    # Make script executable
    chmod +x /usr/local/bin/moodle-backup.sh
    
    # Create backup verification script
    cat > /usr/local/bin/backup-verification.sh << 'EOF'
#!/bin/bash

# Backup verification script
BACKUP_DIR="/backup/moodle"
LOG_FILE="/var/log/backup-verification.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Starting backup verification..." >> $LOG_FILE

# Check if backup directory exists
if [ ! -d "$BACKUP_DIR" ]; then
    echo "[$DATE] ERROR: Backup directory not found" >> $LOG_FILE
    exit 1
fi

# Check latest backup files
LATEST_DB=$(ls -t $BACKUP_DIR/moodle_db_*.sql.gz 2>/dev/null | head -1)
LATEST_FILES=$(ls -t $BACKUP_DIR/moodle_files_*.tar.gz 2>/dev/null | head -1)
LATEST_DATA=$(ls -t $BACKUP_DIR/moodle_data_*.tar.gz 2>/dev/null | head -1)

if [ -z "$LATEST_DB" ] || [ -z "$LATEST_FILES" ] || [ -z "$LATEST_DATA" ]; then
    echo "[$DATE] ERROR: Backup files not found" >> $LOG_FILE
    exit 1
fi

# Check file sizes
DB_SIZE=$(stat -c%s "$LATEST_DB")
FILES_SIZE=$(stat -c%s "$LATEST_FILES")
DATA_SIZE=$(stat -c%s "$LATEST_DATA")

echo "[$DATE] Backup verification completed:" >> $LOG_FILE
echo "[$DATE] Database backup: $LATEST_DB ($DB_SIZE bytes)" >> $LOG_FILE
echo "[$DATE] Files backup: $LATEST_FILES ($FILES_SIZE bytes)" >> $LOG_FILE
echo "[$DATE] Data backup: $LATEST_DATA ($DATA_SIZE bytes)" >> $LOG_FILE
EOF

    # Make script executable
    chmod +x /usr/local/bin/backup-verification.sh
    
    log "INFO" "Backup system setup completed"
}

configure_log_management() {
    log "INFO" "Configuring log management..."
    
    # Create log directories
    mkdir -p /var/log/moodle
    chown syslog:adm /var/log/moodle
    chmod 755 /var/log/moodle
    
    # Configure centralized logging
    cat > /etc/rsyslog.d/50-moodle.conf << 'EOF'
# Moodle application logs
local0.*    /var/log/moodle/app.log
local1.*    /var/log/moodle/error.log
local2.*    /var/log/moodle/access.log

# Security logs
auth.*      /var/log/moodle/security.log
authpriv.*  /var/log/moodle/security.log

# System logs
kern.*      /var/log/moodle/system.log
mail.*      /var/log/moodle/mail.log
EOF

    # Restart rsyslog
    systemctl restart rsyslog
    
    log "INFO" "Log management configuration completed"
}

setup_performance_monitoring() {
    log "INFO" "Setting up performance monitoring..."
    
    # Install monitoring tools
    apt update
    apt install -y htop iotop nethogs iftop bc
    
    # Create performance monitoring script
    cat > /usr/local/bin/performance-monitor.sh << 'EOF'
#!/bin/bash

# Performance monitoring script
LOG_FILE="/var/log/performance.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Performance monitoring..." >> $LOG_FILE

# CPU usage
CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
echo "CPU Usage: $CPU_USAGE%" >> $LOG_FILE

# Memory usage
MEMORY_USAGE=$(free | awk 'NR==2{printf "%.2f", $3*100/$2}')
echo "Memory Usage: $MEMORY_USAGE%" >> $LOG_FILE

# Disk I/O
DISK_IO=$(iostat -x 1 1 | grep -E "(Device|sda)" | tail -1)
echo "Disk I/O: $DISK_IO" >> $LOG_FILE

# Network usage
NETWORK_USAGE=$(iftop -t -s 1 -L 1 | grep -E "(TX|RX)" | tail -2)
echo "Network Usage: $NETWORK_USAGE" >> $LOG_FILE

# Database connections
DB_CONNECTIONS=$(mysql -u root -p'${DB_ROOT_PASSWORD:-}' -e "SHOW STATUS LIKE 'Threads_connected';" 2>/dev/null | awk 'NR==2 {print $2}' || echo "N/A")
echo "Database Connections: $DB_CONNECTIONS" >> $LOG_FILE

echo "---" >> $LOG_FILE
EOF

    # Make script executable
    chmod +x /usr/local/bin/performance-monitor.sh
    
    log "INFO" "Performance monitoring setup completed"
}

create_system_verification_script() {
    log "INFO" "Creating system verification script..."
    
    # Create system verification script
    cat > /usr/local/bin/system-verification.sh << 'EOF'
#!/bin/bash

# System verification script
echo "=== LMS Server System Verification ==="
echo "Date: $(date)"
echo ""

# Check system information
echo "1. System Information:"
echo "   OS: $(lsb_release -d | cut -f2)"
echo "   Kernel: $(uname -r)"
echo "   Uptime: $(uptime -p)"
echo ""

# Check services
echo "2. Service Status:"
SERVICES=("nginx" "php8.1-fpm" "mariadb" "redis-server" "fail2ban")
for service in "${SERVICES[@]}"; do
    if systemctl is-active --quiet $service; then
        echo "   ✓ $service: Running"
    else
        echo "   ✗ $service: Not running"
    fi
done
echo ""

# Check ports
echo "3. Port Status:"
PORTS=("80" "443" "3306" "6379")
for port in "${PORTS[@]}"; do
    if netstat -tlnp | grep -q ":$port "; then
        echo "   ✓ Port $port: Open"
    else
        echo "   ✗ Port $port: Closed"
    fi
done
echo ""

# Check disk space
echo "4. Disk Space:"
df -h | grep -E "(Filesystem|/dev/)"
echo ""

# Check memory
echo "5. Memory Usage:"
free -h
echo ""

# Check network
echo "6. Network Configuration:"
ip addr show | grep -E "(inet |UP)"
echo ""

# Check firewall
echo "7. Firewall Status:"
ufw status | head -5
echo ""

# Check SSL certificate
echo "8. SSL Certificate:"
if [ -f "/etc/letsencrypt/live/${MOODLE_DOMAIN:-lms.example.com}/fullchain.pem" ]; then
    echo "   ✓ SSL certificate found"
    openssl x509 -in /etc/letsencrypt/live/${MOODLE_DOMAIN:-lms.example.com}/fullchain.pem -text -noout | grep -E "(Subject:|Not After:)"
else
    echo "   ✗ SSL certificate not found"
fi
echo ""

echo "=== Verification Complete ==="
EOF

    # Make script executable
    chmod +x /usr/local/bin/system-verification.sh
    
    log "INFO" "System verification script created"
}

# =============================================================================
# Verification Functions
# =============================================================================

verify_basic_configuration() {
    log "INFO" "Verifying basic configuration..."
    
    # Check kernel parameters
    log "INFO" "Checking kernel parameters..."
    sysctl net.core.rmem_max
    sysctl vm.swappiness
    
    # Check system limits
    log "INFO" "Checking system limits..."
    cat /etc/security/limits.d/99-moodle.conf
    
    # Check cron jobs
    log "INFO" "Checking cron jobs..."
    crontab -l | head -10
    
    # Check monitoring scripts
    log "INFO" "Checking monitoring scripts..."
    ls -la /usr/local/bin/*.sh
    
    # Check log directories
    log "INFO" "Checking log directories..."
    ls -la /var/log/moodle/
    
    # Check backup directory
    log "INFO" "Checking backup directory..."
    ls -la /backup/moodle/
    
    # Run system verification
    log "INFO" "Running system verification..."
    /usr/local/bin/system-verification.sh
    
    log "INFO" "Basic configuration verification completed"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    log "INFO" "Starting basic configuration process..."
    
    # Check prerequisites
    check_root
    create_directories
    
    # Execute configuration steps
    configure_kernel_optimization
    configure_system_limits
    configure_cron_jobs
    setup_system_health_monitoring
    setup_backup_system
    configure_log_management
    setup_performance_monitoring
    create_system_verification_script
    
    # Verify configuration
    verify_basic_configuration
    
    log "INFO" "Basic configuration process completed successfully!"
    log "INFO" "Log file: $LOG_FILE"
    log "INFO" "Backup directory: $BACKUP_DIR"
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
        echo "  --verify       Only verify current configuration"
        echo "  --dry-run      Show what would be done without executing"
        exit 0
        ;;
    --verify)
        check_root
        verify_basic_configuration
        exit 0
        ;;
    --dry-run)
        log "INFO" "DRY RUN MODE - No changes will be made"
        log "INFO" "Would execute: configure_kernel_optimization, configure_system_limits, configure_cron_jobs, setup_system_health_monitoring, setup_backup_system, configure_log_management, setup_performance_monitoring, create_system_verification_script"
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
