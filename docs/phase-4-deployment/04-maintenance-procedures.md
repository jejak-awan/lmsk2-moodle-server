# 🔧 Maintenance Procedures for Moodle

## 📋 Overview

Dokumen ini menjelaskan maintenance procedures untuk Moodle production environment, termasuk backup, updates, monitoring, dan disaster recovery.

## 🎯 Objectives

- [ ] Setup automated backup procedures
- [ ] Konfigurasi update procedures
- [ ] Setup monitoring dan alerting
- [ ] Konfigurasi disaster recovery
- [ ] Setup maintenance schedules

## 🔧 Step-by-Step Guide

### Step 1: Automated Backup Procedures

```bash
# Create comprehensive backup script
sudo nano /usr/local/bin/maintenance-backup.sh
```

**Maintenance Backup Script:**
```bash
#!/bin/bash

# Moodle Maintenance Backup Script
BACKUP_DIR="/backup/maintenance"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="moodle"
DB_USER="moodle"
DB_PASS="your_password"

# Create backup directory
mkdir -p $BACKUP_DIR/{database,files,config,logs,system}

echo "Starting maintenance backup at $(date)"

# Database backup
echo "Backing up database..."
mysqldump -u $DB_USER -p$DB_PASS \
    --single-transaction \
    --routines \
    --triggers \
    --events \
    --hex-blob \
    --opt \
    $DB_NAME | gzip > $BACKUP_DIR/database/moodle_maintenance_$DATE.sql.gz

# File backup
echo "Backing up Moodle files..."
tar -czf $BACKUP_DIR/files/moodle_files_$DATE.tar.gz -C /var/www moodle
tar -czf $BACKUP_DIR/files/moodle_data_$DATE.tar.gz -C /var/www moodledata

# Configuration backup
echo "Backing up system configuration..."
tar -czf $BACKUP_DIR/config/system_config_$DATE.tar.gz \
    /etc/nginx/sites-available/moodle \
    /etc/php/8.1/fpm/php.ini \
    /etc/mysql/mariadb.conf.d/ \
    /var/www/moodle/config.php \
    /etc/letsencrypt/live/lms.yourdomain.com/

# Log backup
echo "Backing up system logs..."
tar -czf $BACKUP_DIR/logs/system_logs_$DATE.tar.gz \
    /var/log/nginx/ \
    /var/log/php8.1-fpm/ \
    /var/log/mysql/ \
    /var/log/system-monitor.log \
    /var/log/ssl-monitor.log \
    /var/log/load-balancer-monitor.log

# System backup
echo "Backing up system files..."
tar -czf $BACKUP_DIR/system/system_files_$DATE.tar.gz \
    /etc/nginx/ \
    /etc/php/8.1/fpm/ \
    /etc/mysql/ \
    /etc/letsencrypt/ \
    /etc/ssl/

# Create backup manifest
echo "Creating backup manifest..."
cat > $BACKUP_DIR/backup_manifest_$DATE.txt << EOF
Backup Date: $(date)
Backup Type: Maintenance
Database: moodle_maintenance_$DATE.sql.gz
Files: moodle_files_$DATE.tar.gz
Data: moodle_data_$DATE.tar.gz
Config: system_config_$DATE.tar.gz
Logs: system_logs_$DATE.tar.gz
System: system_files_$DATE.tar.gz
EOF

# Cleanup old backups (keep 90 days)
find $BACKUP_DIR -name "*.gz" -mtime +90 -delete
find $BACKUP_DIR -name "*.txt" -mtime +90 -delete

echo "Maintenance backup completed at $(date)"
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/maintenance-backup.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# Maintenance backup weekly on Sunday at 3 AM
0 3 * * 0 /usr/local/bin/maintenance-backup.sh
```

### Step 2: Update Procedures

```bash
# Create update script
sudo nano /usr/local/bin/moodle-update.sh
```

**Moodle Update Script:**
```bash
#!/bin/bash

# Moodle Update Script
LOG_FILE="/var/log/moodle-update.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Starting Moodle update..." >> $LOG_FILE

# Create update backup
echo "Creating update backup..." >> $LOG_FILE
/usr/local/bin/maintenance-backup.sh

# Put site in maintenance mode
echo "Putting site in maintenance mode..." >> $LOG_FILE
cd /var/www/moodle
sudo -u www-data php admin/cli/maintenance.php --enable

# Update Moodle
echo "Updating Moodle..." >> $LOG_FILE
cd /var/www/moodle
sudo -u www-data php admin/cli/upgrade.php --non-interactive

# Update plugins
echo "Updating plugins..." >> $LOG_FILE
sudo -u www-data php admin/cli/upgrade.php --non-interactive --allow-unstable

# Clear caches
echo "Clearing caches..." >> $LOG_FILE
sudo -u www-data php admin/cli/purge_caches.php

# Disable maintenance mode
echo "Disabling maintenance mode..." >> $LOG_FILE
sudo -u www-data php admin/cli/maintenance.php --disable

# Restart services
echo "Restarting services..." >> $LOG_FILE
sudo systemctl restart php8.1-fpm
sudo systemctl restart nginx

echo "[$DATE] Moodle update completed" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/moodle-update.sh
```

### Step 3: Monitoring and Alerting

```bash
# Create monitoring script
sudo nano /usr/local/bin/maintenance-monitor.sh
```

**Maintenance Monitoring Script:**
```bash
#!/bin/bash

# Maintenance Monitoring Script
LOG_FILE="/var/log/maintenance-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')
ALERT_EMAIL="admin@yourdomain.com"

echo "[$DATE] Maintenance monitoring..." >> $LOG_FILE

# Check system resources
CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
MEMORY_USAGE=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')

echo "CPU Usage: $CPU_USAGE%" >> $LOG_FILE
echo "Memory Usage: $MEMORY_USAGE%" >> $LOG_FILE
echo "Disk Usage: $DISK_USAGE%" >> $LOG_FILE

# Check services
SERVICES=("nginx" "php8.1-fpm" "mariadb" "redis-server")
for service in "${SERVICES[@]}"; do
    if systemctl is-active --quiet $service; then
        echo "$service: Running" >> $LOG_FILE
    else
        echo "$service: Not running" >> $LOG_FILE
        echo "ALERT: Service $service is not running" | mail -s "Service Alert" $ALERT_EMAIL
    fi
done

# Check database
DB_CONNECTIONS=$(mysql -u moodle -p'your_password' -e "SHOW STATUS LIKE 'Threads_connected';" | awk 'NR==2 {print $2}')
echo "Database Connections: $DB_CONNECTIONS" >> $LOG_FILE

# Check SSL certificate
CERT_EXPIRY=$(echo | openssl s_client -servername lms.yourdomain.com -connect lms.yourdomain.com:443 2>/dev/null | openssl x509 -noout -dates | grep notAfter | cut -d= -f2)
CERT_EXPIRY_EPOCH=$(date -d "$CERT_EXPIRY" +%s)
CURRENT_EPOCH=$(date +%s)
DAYS_UNTIL_EXPIRY=$(( (CERT_EXPIRY_EPOCH - CURRENT_EPOCH) / 86400 ))

echo "SSL Certificate expires in: $DAYS_UNTIL_EXPIRY days" >> $LOG_FILE

# Alert if certificate expires in less than 30 days
if [ $DAYS_UNTIL_EXPIRY -lt 30 ]; then
    echo "ALERT: SSL Certificate expires in $DAYS_UNTIL_EXPIRY days" | mail -s "SSL Certificate Alert" $ALERT_EMAIL
fi

# Check backup status
LAST_BACKUP=$(find /backup/maintenance -name "*.gz" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2)
if [ -n "$LAST_BACKUP" ]; then
    BACKUP_AGE=$(( ($(date +%s) - $(stat -c %Y "$LAST_BACKUP")) / 86400 ))
    echo "Last backup: $BACKUP_AGE days ago" >> $LOG_FILE
    
    if [ $BACKUP_AGE -gt 7 ]; then
        echo "ALERT: Last backup was $BACKUP_AGE days ago" | mail -s "Backup Alert" $ALERT_EMAIL
    fi
fi

echo "---" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/maintenance-monitor.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# Maintenance monitoring every hour
0 * * * * /usr/local/bin/maintenance-monitor.sh
```

### Step 4: Disaster Recovery

```bash
# Create disaster recovery script
sudo nano /usr/local/bin/disaster-recovery.sh
```

**Disaster Recovery Script:**
```bash
#!/bin/bash

# Disaster Recovery Script
BACKUP_DIR="/backup/maintenance"
RESTORE_DIR="/tmp/disaster-recovery"
DATE=$(date +%Y%m%d_%H%M%S)

echo "Starting disaster recovery at $(date)"

# Create restore directory
mkdir -p $RESTORE_DIR

# Find latest backup
LATEST_BACKUP=$(find $BACKUP_DIR -name "backup_manifest_*.txt" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2)

if [ -z "$LATEST_BACKUP" ]; then
    echo "No backup found for disaster recovery"
    exit 1
fi

echo "Using backup: $LATEST_BACKUP"

# Extract backup manifest
BACKUP_DATE=$(basename $LATEST_BACKUP .txt | sed 's/backup_manifest_//')
echo "Backup date: $BACKUP_DATE"

# Stop services
echo "Stopping services..."
sudo systemctl stop nginx
sudo systemctl stop php8.1-fpm
sudo systemctl stop mariadb

# Restore database
echo "Restoring database..."
sudo systemctl start mariadb
mysql -u root -p -e "DROP DATABASE IF EXISTS moodle;"
mysql -u root -p -e "CREATE DATABASE moodle CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
gunzip -c $BACKUP_DIR/database/moodle_maintenance_$BACKUP_DATE.sql.gz | mysql -u root -p moodle

# Restore files
echo "Restoring files..."
sudo rm -rf /var/www/moodle
sudo rm -rf /var/www/moodledata
sudo tar -xzf $BACKUP_DIR/files/moodle_files_$BACKUP_DATE.tar.gz -C /var/www/
sudo tar -xzf $BACKUP_DIR/files/moodle_data_$BACKUP_DATE.tar.gz -C /var/www/

# Restore configuration
echo "Restoring configuration..."
sudo tar -xzf $BACKUP_DIR/config/system_config_$BACKUP_DATE.tar.gz -C /

# Set permissions
echo "Setting permissions..."
sudo chown -R www-data:www-data /var/www/moodle
sudo chown -R www-data:www-data /var/www/moodledata
sudo chmod -R 755 /var/www/moodle
sudo chmod -R 777 /var/www/moodledata

# Start services
echo "Starting services..."
sudo systemctl start php8.1-fpm
sudo systemctl start nginx

# Clear caches
echo "Clearing caches..."
cd /var/www/moodle
sudo -u www-data php admin/cli/purge_caches.php

echo "Disaster recovery completed at $(date)"
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/disaster-recovery.sh
```

### Step 5: Maintenance Schedules

```bash
# Create maintenance schedule
sudo nano /usr/local/bin/maintenance-schedule.sh
```

**Maintenance Schedule Script:**
```bash
#!/bin/bash

# Maintenance Schedule Script
LOG_FILE="/var/log/maintenance-schedule.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Maintenance schedule check..." >> $LOG_FILE

# Get current day and time
CURRENT_DAY=$(date +%u)  # 1=Monday, 7=Sunday
CURRENT_HOUR=$(date +%H)

# Daily maintenance (every day at 2 AM)
if [ $CURRENT_HOUR -eq 2 ]; then
    echo "Running daily maintenance..." >> $LOG_FILE
    
    # Clear temporary files
    sudo find /tmp -type f -mtime +7 -delete
    sudo find /var/tmp -type f -mtime +7 -delete
    
    # Clear log files
    sudo find /var/log -name "*.log" -mtime +30 -delete
    
    # Update package list
    sudo apt update
    
    echo "Daily maintenance completed" >> $LOG_FILE
fi

# Weekly maintenance (Sunday at 3 AM)
if [ $CURRENT_DAY -eq 7 ] && [ $CURRENT_HOUR -eq 3 ]; then
    echo "Running weekly maintenance..." >> $LOG_FILE
    
    # Run maintenance backup
    /usr/local/bin/maintenance-backup.sh
    
    # Update system packages
    sudo apt upgrade -y
    
    # Restart services
    sudo systemctl restart nginx
    sudo systemctl restart php8.1-fpm
    sudo systemctl restart mariadb
    sudo systemctl restart redis-server
    
    echo "Weekly maintenance completed" >> $LOG_FILE
fi

# Monthly maintenance (1st of month at 4 AM)
if [ $(date +%d) -eq 1 ] && [ $CURRENT_HOUR -eq 4 ]; then
    echo "Running monthly maintenance..." >> $LOG_FILE
    
    # Update Moodle
    /usr/local/bin/moodle-update.sh
    
    # Clean up old backups
    find /backup/maintenance -name "*.gz" -mtime +90 -delete
    
    # Update SSL certificate
    sudo certbot renew
    
    echo "Monthly maintenance completed" >> $LOG_FILE
fi

echo "---" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/maintenance-schedule.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# Maintenance schedule check every hour
0 * * * * /usr/local/bin/maintenance-schedule.sh
```

## ✅ Verification

### Maintenance Test

```bash
# Test backup script
sudo /usr/local/bin/maintenance-backup.sh

# Test monitoring script
sudo /usr/local/bin/maintenance-monitor.sh

# Test update script (in test environment)
sudo /usr/local/bin/moodle-update.sh

# Check maintenance logs
tail -f /var/log/maintenance-monitor.log
tail -f /var/log/maintenance-schedule.log
```

### Expected Results

- ✅ Backup procedures working
- ✅ Update procedures configured
- ✅ Monitoring and alerting active
- ✅ Disaster recovery ready
- ✅ Maintenance schedules configured

## 🚨 Troubleshooting

### Common Issues

**1. Backup failures**
```bash
# Check backup logs
tail -f /var/log/maintenance-backup.log

# Check disk space
df -h

# Check backup directory permissions
ls -la /backup/maintenance/
```

**2. Update failures**
```bash
# Check update logs
tail -f /var/log/moodle-update.log

# Check Moodle status
cd /var/www/moodle
sudo -u www-data php admin/cli/status.php
```

**3. Monitoring issues**
```bash
# Check monitoring logs
tail -f /var/log/maintenance-monitor.log

# Check cron jobs
sudo crontab -l
```

## 📝 Next Steps

Setelah maintenance procedures selesai, lanjutkan ke:
- [Phase 5: Advanced Features](../phase-5-advanced/README.md)

## 📚 References

- [Moodle Maintenance](https://docs.moodle.org/400/en/Maintenance_mode)
- [Moodle Updates](https://docs.moodle.org/400/en/Upgrading)
- [Disaster Recovery](https://docs.moodle.org/400/en/Restore)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
