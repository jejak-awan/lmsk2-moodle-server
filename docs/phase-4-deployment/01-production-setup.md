# 🚀 Production Setup for Moodle

## 📋 Overview

Dokumen ini menjelaskan setup production environment untuk Moodle, termasuk konfigurasi server, security hardening, dan optimasi untuk production use.

## 🎯 Objectives

- [ ] Setup production server environment
- [ ] Konfigurasi security hardening
- [ ] Setup monitoring dan logging
- [ ] Konfigurasi backup dan disaster recovery
- [ ] Optimasi performa untuk production

## 🔧 Step-by-Step Guide

### Step 1: Production Server Preparation

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install essential packages
sudo apt install -y curl wget git htop iotop nethogs iftop nload

# Configure timezone
sudo timedatectl set-timezone Asia/Jakarta

# Configure hostname
sudo hostnamectl set-hostname lms-prod.yourdomain.com
```

### Step 2: Security Hardening

```bash
# Create security hardening script
sudo nano /usr/local/bin/security-hardening.sh
```

**Security Hardening Script:**
```bash
#!/bin/bash

# Production Security Hardening Script
echo "=== Production Security Hardening ==="

# Disable root login
sudo sed -i 's/#PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
sudo sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config

# Configure SSH
sudo sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config

# Restart SSH
sudo systemctl restart sshd

# Configure firewall
sudo ufw --force enable
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 'Nginx Full'
sudo ufw allow 443/tcp
sudo ufw allow 80/tcp

# Install fail2ban
sudo apt install -y fail2ban

# Configure fail2ban
sudo nano /etc/fail2ban/jail.local
```

**Fail2ban Configuration:**
```ini
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 3

[nginx-http-auth]
enabled = true
filter = nginx-http-auth
port = http,https
logpath = /var/log/nginx/error.log

[nginx-limit-req]
enabled = true
filter = nginx-limit-req
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 10
```

```bash
# Start fail2ban
sudo systemctl start fail2ban
sudo systemctl enable fail2ban

# Install ClamAV
sudo apt install -y clamav clamav-daemon

# Update ClamAV
sudo freshclam

# Configure ClamAV
sudo systemctl start clamav-daemon
sudo systemctl enable clamav-daemon
```

### Step 3: System Monitoring Setup

```bash
# Install monitoring tools
sudo apt install -y htop iotop nethogs iftop nload

# Create system monitoring script
sudo nano /usr/local/bin/system-monitor.sh
```

**System Monitoring Script:**
```bash
#!/bin/bash

# Production System Monitoring Script
LOG_FILE="/var/log/system-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] System monitoring..." >> $LOG_FILE

# CPU usage
CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
echo "CPU Usage: $CPU_USAGE%" >> $LOG_FILE

# Memory usage
MEMORY_USAGE=$(free | awk 'NR==2{printf "%.2f", $3*100/$2}')
echo "Memory Usage: $MEMORY_USAGE%" >> $LOG_FILE

# Disk usage
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
echo "Disk Usage: $DISK_USAGE%" >> $LOG_FILE

# Load average
LOAD_AVG=$(uptime | awk -F'load average:' '{print $2}')
echo "Load Average: $LOAD_AVG" >> $LOG_FILE

# Network connections
NET_CONN=$(netstat -an | grep ESTABLISHED | wc -l)
echo "Network Connections: $NET_CONN" >> $LOG_FILE

# Check services
SERVICES=("nginx" "php8.1-fpm" "mariadb" "redis-server")
for service in "${SERVICES[@]}"; do
    if systemctl is-active --quiet $service; then
        echo "$service: Running" >> $LOG_FILE
    else
        echo "$service: Not running" >> $LOG_FILE
    fi
done

echo "---" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/system-monitor.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# System monitoring every 5 minutes
*/5 * * * * /usr/local/bin/system-monitor.sh
```

### Step 4: Log Management

```bash
# Configure logrotate
sudo nano /etc/logrotate.d/moodle
```

**Logrotate Configuration:**
```
/var/log/nginx/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 www-data adm
    postrotate
        if [ -f /var/run/nginx.pid ]; then
            kill -USR1 $(cat /var/run/nginx.pid)
        fi
    endscript
}

/var/log/php8.1-fpm*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 644 www-data adm
    postrotate
        /bin/kill -SIGUSR1 $(cat /var/run/php/php8.1-fpm.pid 2>/dev/null) 2>/dev/null || true
    endscript
}

/var/log/mysql/*.log {
    daily
    missingok
    rotate 52
    compress
    delaycompress
    notifempty
    create 640 mysql adm
    postrotate
        if [ -f /var/run/mysqld/mysqld.pid ]; then
            kill -USR1 $(cat /var/run/mysqld/mysqld.pid)
        fi
    endscript
}
```

### Step 5: Backup Strategy

```bash
# Create comprehensive backup script
sudo nano /usr/local/bin/production-backup.sh
```

**Production Backup Script:**
```bash
#!/bin/bash

# Production Backup Script
BACKUP_DIR="/backup/production"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="moodle"
DB_USER="moodle"
DB_PASS="your_password"

# Create backup directory
mkdir -p $BACKUP_DIR/{database,files,config,logs}

echo "Starting production backup at $(date)"

# Database backup
mysqldump -u $DB_USER -p$DB_PASS \
    --single-transaction \
    --routines \
    --triggers \
    --events \
    --hex-blob \
    --opt \
    $DB_NAME | gzip > $BACKUP_DIR/database/moodle_$DATE.sql.gz

# File backup
tar -czf $BACKUP_DIR/files/moodle_files_$DATE.tar.gz -C /var/www moodle
tar -czf $BACKUP_DIR/files/moodle_data_$DATE.tar.gz -C /var/www moodledata

# Configuration backup
tar -czf $BACKUP_DIR/config/system_config_$DATE.tar.gz \
    /etc/nginx/sites-available/moodle \
    /etc/php/8.1/fpm/php.ini \
    /etc/mysql/mariadb.conf.d/ \
    /var/www/moodle/config.php

# Log backup
tar -czf $BACKUP_DIR/logs/system_logs_$DATE.tar.gz \
    /var/log/nginx/ \
    /var/log/php8.1-fpm/ \
    /var/log/mysql/ \
    /var/log/system-monitor.log

# Cleanup old backups (keep 30 days)
find $BACKUP_DIR -name "*.gz" -mtime +30 -delete

echo "Production backup completed at $(date)"
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/production-backup.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# Production backup daily at 2 AM
0 2 * * * /usr/local/bin/production-backup.sh
```

### Step 6: Performance Optimization

```bash
# Create performance optimization script
sudo nano /usr/local/bin/performance-optimization.sh
```

**Performance Optimization Script:**
```bash
#!/bin/bash

# Production Performance Optimization Script
echo "=== Production Performance Optimization ==="

# Kernel parameters optimization
cat >> /etc/sysctl.d/99-production-performance.conf << EOF
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
sysctl -p /etc/sysctl.d/99-production-performance.conf

# Optimize file limits
cat >> /etc/security/limits.d/99-production-performance.conf << EOF
# Production performance limits
moodle soft nofile 65536
moodle hard nofile 65536
moodle soft nproc 32768
moodle hard nproc 32768

www-data soft nofile 65536
www-data hard nofile 65536
www-data soft nproc 32768
www-data hard nproc 32768

root soft nofile 65536
root hard nofile 65536
root soft nproc 32768
root hard nproc 32768
EOF

echo "Production performance optimization completed"
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/performance-optimization.sh

# Run optimization
sudo /usr/local/bin/performance-optimization.sh
```

## ✅ Verification

### Production Setup Test

```bash
# Test security hardening
sudo ufw status
sudo fail2ban-client status

# Test monitoring
sudo /usr/local/bin/system-monitor.sh
tail -f /var/log/system-monitor.log

# Test backup
sudo /usr/local/bin/production-backup.sh
ls -la /backup/production/

# Test performance
htop
iotop
nethogs
```

### Expected Results

- ✅ Security hardening applied
- ✅ Monitoring system active
- ✅ Backup system working
- ✅ Performance optimized
- ✅ Log management configured

## 🚨 Troubleshooting

### Common Issues

**1. Security issues**
```bash
# Check firewall status
sudo ufw status

# Check fail2ban status
sudo fail2ban-client status
```

**2. Performance issues**
```bash
# Check system resources
htop
df -h
free -h

# Check kernel parameters
sysctl -a | grep -E "(net\.|vm\.|fs\.)"
```

## 📝 Next Steps

Setelah production setup selesai, lanjutkan ke:
- [02-ssl-certificate.md](02-ssl-certificate.md) - Setup SSL certificate

## 📚 References

- [Ubuntu Security Hardening](https://ubuntu.com/security)
- [Fail2ban Documentation](https://www.fail2ban.org/wiki/index.php/Main_Page)
- [ClamAV Documentation](https://www.clamav.net/documents)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
