# ⚙️ Basic Configuration

## 📋 Overview

Dokumen ini menjelaskan konfigurasi dasar sistem yang diperlukan sebelum instalasi Moodle, termasuk optimasi kernel, konfigurasi cron jobs, dan setup monitoring dasar.

## 🎯 Objectives

- [ ] Optimasi kernel parameters untuk performa
- [ ] Konfigurasi cron jobs dan scheduled tasks
- [ ] Setup monitoring dan alerting dasar
- [ ] Konfigurasi backup otomatis
- [ ] Setup log management
- [ ] Final system verification

## 📋 Prerequisites

- Server Ubuntu 22.04 LTS sudah dikonfigurasi
- Nginx, PHP, MariaDB, Redis sudah terinstall
- Security hardening sudah selesai
- SSL certificate sudah terinstall

## 🔧 Step-by-Step Guide

### Step 1: Kernel Optimization

```bash
# Create kernel optimization configuration
sudo nano /etc/sysctl.d/99-moodle-optimization.conf
```

**Kernel Optimization Configuration:**
```
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
```

```bash
# Apply kernel parameters
sudo sysctl -p /etc/sysctl.d/99-moodle-optimization.conf

# Verify changes
sysctl net.core.rmem_max
sysctl vm.swappiness
```

### Step 2: System Limits Configuration

```bash
# Configure system limits
sudo nano /etc/security/limits.d/99-moodle.conf
```

**System Limits Configuration:**
```
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
```

### Step 3: Cron Jobs Configuration

```bash
# Create cron jobs for system maintenance
sudo crontab -e
```

**System Cron Jobs:**
```
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
```

### Step 4: System Health Monitoring

```bash
# Create system health check script
sudo nano /usr/local/bin/system-health-check.sh
```

**System Health Check Script:**
```bash
#!/bin/bash

# System health monitoring script
LOG_FILE="/var/log/system-health.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')
ALERT_EMAIL="admin@yourdomain.com"

echo "[$DATE] Starting system health check..." >> $LOG_FILE

# Check disk space
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    echo "[$DATE] ALERT: Disk usage critical: $DISK_USAGE%" >> $LOG_FILE
    # Send alert email (if mail is configured)
    echo "Disk usage is at $DISK_USAGE%" | mail -s "Disk Usage Alert" $ALERT_EMAIL
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
if ! mysql -u root -p'your_password' -e "SELECT 1;" > /dev/null 2>&1; then
    echo "[$DATE] ALERT: Database connection failed" >> $LOG_FILE
fi

# Check Redis connectivity
if ! redis-cli ping > /dev/null 2>&1; then
    echo "[$DATE] ALERT: Redis connection failed" >> $LOG_FILE
fi

echo "[$DATE] System health check completed." >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/system-health-check.sh

# Test script
sudo /usr/local/bin/system-health-check.sh
```

### Step 5: Backup Configuration

```bash
# Create backup script
sudo nano /usr/local/bin/moodle-backup.sh
```

**Moodle Backup Script:**
```bash
#!/bin/bash

# Moodle backup script
BACKUP_DIR="/backup/moodle"
DATE=$(date +%Y%m%d_%H%M%S)
MOODLE_DIR="/var/www/moodle"
MOODLE_DATA="/var/www/moodle/moodledata"
DB_NAME="moodle"
DB_USER="moodle"
DB_PASS="your_password"

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
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/moodle-backup.sh

# Create backup directory
sudo mkdir -p /backup/moodle
sudo chown root:root /backup/moodle
sudo chmod 700 /backup/moodle

# Add backup to crontab
sudo crontab -e
```

**Add backup cron job:**
```
# Daily backup at 2 AM
0 2 * * * /usr/local/bin/moodle-backup.sh
```

### Step 6: Log Management Setup

```bash
# Configure centralized logging
sudo nano /etc/rsyslog.d/50-moodle.conf
```

**Rsyslog Configuration:**
```
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
```

```bash
# Create log directories
sudo mkdir -p /var/log/moodle
sudo chown syslog:adm /var/log/moodle
sudo chmod 755 /var/log/moodle

# Restart rsyslog
sudo systemctl restart rsyslog
```

### Step 7: Performance Monitoring

```bash
# Install monitoring tools
sudo apt install -y htop iotop nethogs iftop

# Create performance monitoring script
sudo nano /usr/local/bin/performance-monitor.sh
```

**Performance Monitor Script:**
```bash
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
DB_CONNECTIONS=$(mysql -u root -p'your_password' -e "SHOW STATUS LIKE 'Threads_connected';" | awk 'NR==2 {print $2}')
echo "Database Connections: $DB_CONNECTIONS" >> $LOG_FILE

echo "---" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/performance-monitor.sh

# Add to crontab (every 5 minutes)
sudo crontab -e
```

**Add performance monitoring cron job:**
```
# Performance monitoring every 5 minutes
*/5 * * * * /usr/local/bin/performance-monitor.sh
```

### Step 8: Final System Verification

```bash
# Create system verification script
sudo nano /usr/local/bin/system-verification.sh
```

**System Verification Script:**
```bash
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
sudo ufw status | head -5
echo ""

# Check SSL certificate
echo "8. SSL Certificate:"
if [ -f "/etc/letsencrypt/live/lms.yourdomain.com/fullchain.pem" ]; then
    echo "   ✓ SSL certificate found"
    openssl x509 -in /etc/letsencrypt/live/lms.yourdomain.com/fullchain.pem -text -noout | grep -E "(Subject:|Not After:)"
else
    echo "   ✗ SSL certificate not found"
fi
echo ""

echo "=== Verification Complete ==="
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/system-verification.sh

# Run verification
sudo /usr/local/bin/system-verification.sh
```

## ✅ Verification

### Final System Check

```bash
# Run comprehensive system check
sudo /usr/local/bin/system-verification.sh

# Check all services
sudo systemctl status nginx php8.1-fpm mariadb redis-server fail2ban

# Check cron jobs
sudo crontab -l

# Check log files
sudo tail -f /var/log/system-health.log
sudo tail -f /var/log/performance.log

# Test backup script
sudo /usr/local/bin/moodle-backup.sh
```

### Expected Results

- ✅ All services running properly
- ✅ Kernel parameters optimized
- ✅ Cron jobs configured
- ✅ Monitoring scripts active
- ✅ Backup system ready
- ✅ Log management configured
- ✅ Performance monitoring active
- ✅ System verification passed

## 🚨 Troubleshooting

### Common Issues

**1. Kernel parameters not applied**
```bash
# Check current values
sysctl net.core.rmem_max
# Reload configuration
sudo sysctl -p /etc/sysctl.d/99-moodle-optimization.conf
```

**2. Cron jobs not running**
```bash
# Check cron service
sudo systemctl status cron
# Check cron logs
sudo tail -f /var/log/cron.log
```

**3. Backup script failing**
```bash
# Check backup directory permissions
ls -la /backup/moodle/
# Test database connection
mysql -u moodle -p -e "SELECT 1;"
```

## 📝 Next Steps

Setelah basic configuration selesai, lanjutkan ke:
- [Phase 2: Moodle Installation](../phase-2-moodle-installation/01-moodle-3.11-lts/01-requirements.md) - Install Moodle sesuai versi

## 📚 References

- [Ubuntu System Administration](https://ubuntu.com/server/docs)
- [Cron Jobs Guide](https://help.ubuntu.com/community/CronHowto)
- [System Monitoring](https://ubuntu.com/server/docs/monitoring)
- [Backup Strategies](https://ubuntu.com/server/docs/backups)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
