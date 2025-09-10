# 📊 Monitoring Setup for Moodle 3.11 LTS

## 📋 Overview

Dokumen ini menjelaskan setup monitoring system untuk Moodle 3.11 LTS, termasuk system monitoring, application monitoring, dan alerting.

## 🎯 Objectives

- [ ] Setup system monitoring (CPU, Memory, Disk)
- [ ] Konfigurasi application monitoring
- [ ] Setup log monitoring dan analysis
- [ ] Konfigurasi alerting system
- [ ] Setup performance dashboards

## 🔧 Step-by-Step Guide

### Step 1: System Monitoring Setup

```bash
# Install monitoring tools
sudo apt install -y htop iotop nethogs iftop nload

# Create system monitoring script
sudo nano /usr/local/bin/system-monitor.sh
```

**System Monitoring Script:**
```bash
#!/bin/bash

# System Monitoring Script
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

echo "---" >> $LOG_FILE
```

### Step 2: Application Monitoring

```bash
# Create application monitoring script
sudo nano /usr/local/bin/app-monitor.sh
```

**Application Monitoring Script:**
```bash
#!/bin/bash

# Application Monitoring Script
LOG_FILE="/var/log/app-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Application monitoring..." >> $LOG_FILE

# Check services
SERVICES=("nginx" "php8.1-fpm" "mariadb" "redis-server")
for service in "${SERVICES[@]}"; do
    if systemctl is-active --quiet $service; then
        echo "$service: Running" >> $LOG_FILE
    else
        echo "$service: Not running" >> $LOG_FILE
    fi
done

# Database connections
DB_CONN=$(mysql -u root -p'your_password' -e "SHOW STATUS LIKE 'Threads_connected';" | awk 'NR==2 {print $2}')
echo "Database Connections: $DB_CONN" >> $LOG_FILE

# PHP-FPM processes
PHP_PROC=$(ps aux | grep php-fpm | grep -v grep | wc -l)
echo "PHP-FPM Processes: $PHP_PROC" >> $LOG_FILE

# Nginx active connections
NGINX_CONN=$(netstat -an | grep :443 | grep ESTABLISHED | wc -l)
echo "Nginx Active Connections: $NGINX_CONN" >> $LOG_FILE

echo "---" >> $LOG_FILE
```

### Step 3: Log Monitoring

```bash
# Create log monitoring script
sudo nano /usr/local/bin/log-monitor.sh
```

**Log Monitoring Script:**
```bash
#!/bin/bash

# Log Monitoring Script
LOG_FILE="/var/log/log-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Log monitoring..." >> $LOG_FILE

# Check error logs
ERROR_COUNT=$(tail -100 /var/log/nginx/moodle-error.log | grep -c "$(date '+%Y/%m/%d')")
echo "Nginx Errors Today: $ERROR_COUNT" >> $LOG_FILE

PHP_ERRORS=$(tail -100 /var/log/php_errors.log | grep -c "$(date '+%Y-%m-%d')")
echo "PHP Errors Today: $PHP_ERRORS" >> $LOG_FILE

# Check access logs
ACCESS_COUNT=$(tail -100 /var/log/nginx/moodle-access.log | grep -c "$(date '+%Y/%m/%d')")
echo "Nginx Access Today: $ACCESS_COUNT" >> $LOG_FILE

echo "---" >> $LOG_FILE
```

### Step 4: Alerting System

```bash
# Create alerting script
sudo nano /usr/local/bin/alert-system.sh
```

**Alerting Script:**
```bash
#!/bin/bash

# Alerting System Script
ALERT_EMAIL="admin@yourdomain.com"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

# Check CPU usage
CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
if (( $(echo "$CPU_USAGE > 80" | bc -l) )); then
    echo "ALERT: High CPU usage: $CPU_USAGE%" | mail -s "CPU Alert" $ALERT_EMAIL
fi

# Check memory usage
MEMORY_USAGE=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
if [ $MEMORY_USAGE -gt 80 ]; then
    echo "ALERT: High memory usage: $MEMORY_USAGE%" | mail -s "Memory Alert" $ALERT_EMAIL
fi

# Check disk usage
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    echo "ALERT: High disk usage: $DISK_USAGE%" | mail -s "Disk Alert" $ALERT_EMAIL
fi

# Check services
SERVICES=("nginx" "php8.1-fpm" "mariadb" "redis-server")
for service in "${SERVICES[@]}"; do
    if ! systemctl is-active --quiet $service; then
        echo "ALERT: Service $service is not running" | mail -s "Service Alert" $ALERT_EMAIL
    fi
done
```

### Step 5: Setup Cron Jobs

```bash
# Add monitoring cron jobs
sudo crontab -e
```

**Add to crontab:**
```
# System monitoring every 5 minutes
*/5 * * * * /usr/local/bin/system-monitor.sh

# Application monitoring every 5 minutes
*/5 * * * * /usr/local/bin/app-monitor.sh

# Log monitoring every 10 minutes
*/10 * * * * /usr/local/bin/log-monitor.sh

# Alerting every 15 minutes
*/15 * * * * /usr/local/bin/alert-system.sh
```

### Step 6: Make Scripts Executable

```bash
# Make all monitoring scripts executable
sudo chmod +x /usr/local/bin/system-monitor.sh
sudo chmod +x /usr/local/bin/app-monitor.sh
sudo chmod +x /usr/local/bin/log-monitor.sh
sudo chmod +x /usr/local/bin/alert-system.sh
```

## ✅ Verification

### Monitoring Test

```bash
# Test monitoring scripts
sudo /usr/local/bin/system-monitor.sh
sudo /usr/local/bin/app-monitor.sh
sudo /usr/local/bin/log-monitor.sh

# Check log files
tail -f /var/log/system-monitor.log
tail -f /var/log/app-monitor.log
tail -f /var/log/log-monitor.log

# Check cron jobs
sudo crontab -l
```

### Expected Results

- ✅ System monitoring active
- ✅ Application monitoring active
- ✅ Log monitoring active
- ✅ Alerting system configured
- ✅ Cron jobs running

## 📝 Next Steps

Setelah monitoring setup selesai, lanjutkan ke:
- [04-backup-strategy.md](04-backup-strategy.md) - Setup backup strategy

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
