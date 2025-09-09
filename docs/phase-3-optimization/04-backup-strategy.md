# 💾 Backup Strategy for Moodle 3.11 LTS

## 📋 Overview

Dokumen ini menjelaskan backup strategy untuk Moodle 3.11 LTS, termasuk database backup, file backup, dan disaster recovery.

## 🎯 Objectives

- [ ] Setup automated database backup
- [ ] Konfigurasi file backup system
- [ ] Setup backup verification
- [ ] Konfigurasi disaster recovery
- [ ] Setup backup monitoring

## 🔧 Step-by-Step Guide

### Step 1: Database Backup Setup

```bash
# Create database backup script
sudo nano /usr/local/bin/moodle-db-backup.sh
```

**Database Backup Script:**
```bash
#!/bin/bash

# Moodle Database Backup Script
BACKUP_DIR="/backup/moodle/database"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="moodle"
DB_USER="moodle"
DB_PASS="your_password"

# Create backup directory
mkdir -p $BACKUP_DIR

echo "Starting database backup at $(date)"

# Full database backup
mysqldump -u $DB_USER -p$DB_PASS \
    --single-transaction \
    --routines \
    --triggers \
    --events \
    --hex-blob \
    --opt \
    $DB_NAME | gzip > $BACKUP_DIR/moodle_full_$DATE.sql.gz

# Schema only backup
mysqldump -u $DB_USER -p$DB_PASS \
    --no-data \
    --routines \
    --triggers \
    --events \
    $DB_NAME | gzip > $BACKUP_DIR/moodle_schema_$DATE.sql.gz

# Cleanup old backups (keep 7 days)
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete

echo "Database backup completed at $(date)"
```

### Step 2: File Backup Setup

```bash
# Create file backup script
sudo nano /usr/local/bin/moodle-file-backup.sh
```

**File Backup Script:**
```bash
#!/bin/bash

# Moodle File Backup Script
BACKUP_DIR="/backup/moodle/files"
DATE=$(date +%Y%m%d_%H%M%S)
MOODLE_DIR="/var/www/moodle"

# Create backup directory
mkdir -p $BACKUP_DIR

echo "Starting file backup at $(date)"

# Backup Moodle files
tar -czf $BACKUP_DIR/moodle_files_$DATE.tar.gz -C /var/www moodle

# Backup Moodle data
tar -czf $BACKUP_DIR/moodle_data_$DATE.tar.gz -C /var/www moodle/moodledata

# Cleanup old backups (keep 7 days)
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

echo "File backup completed at $(date)"
```

### Step 3: Configuration Backup

```bash
# Create configuration backup script
sudo nano /usr/local/bin/moodle-config-backup.sh
```

**Configuration Backup Script:**
```bash
#!/bin/bash

# Moodle Configuration Backup Script
BACKUP_DIR="/backup/moodle/config"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

echo "Starting configuration backup at $(date)"

# Backup configuration files
tar -czf $BACKUP_DIR/moodle_config_$DATE.tar.gz \
    /etc/nginx/sites-available/moodle \
    /etc/php/8.1/fpm/php.ini \
    /etc/mysql/mariadb.conf.d/ \
    /var/www/moodle/config.php

# Cleanup old backups (keep 30 days)
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

echo "Configuration backup completed at $(date)"
```

### Step 4: Backup Verification

```bash
# Create backup verification script
sudo nano /usr/local/bin/backup-verification.sh
```

**Backup Verification Script:**
```bash
#!/bin/bash

# Backup Verification Script
BACKUP_DIR="/backup/moodle"
LOG_FILE="/var/log/backup-verification.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Backup verification..." >> $LOG_FILE

# Check database backup
DB_BACKUP=$(find $BACKUP_DIR/database -name "*.gz" -mtime -1 | wc -l)
if [ $DB_BACKUP -gt 0 ]; then
    echo "✓ Database backup found: $DB_BACKUP files" >> $LOG_FILE
else
    echo "✗ No recent database backup found" >> $LOG_FILE
fi

# Check file backup
FILE_BACKUP=$(find $BACKUP_DIR/files -name "*.tar.gz" -mtime -1 | wc -l)
if [ $FILE_BACKUP -gt 0 ]; then
    echo "✓ File backup found: $FILE_BACKUP files" >> $LOG_FILE
else
    echo "✗ No recent file backup found" >> $LOG_FILE
fi

# Check configuration backup
CONFIG_BACKUP=$(find $BACKUP_DIR/config -name "*.tar.gz" -mtime -1 | wc -l)
if [ $CONFIG_BACKUP -gt 0 ]; then
    echo "✓ Configuration backup found: $CONFIG_BACKUP files" >> $LOG_FILE
else
    echo "✗ No recent configuration backup found" >> $LOG_FILE
fi

echo "---" >> $LOG_FILE
```

### Step 5: Setup Cron Jobs

```bash
# Add backup cron jobs
sudo crontab -e
```

**Add to crontab:**
```
# Database backup daily at 2 AM
0 2 * * * /usr/local/bin/moodle-db-backup.sh

# File backup daily at 3 AM
0 3 * * * /usr/local/bin/moodle-file-backup.sh

# Configuration backup weekly on Sunday at 4 AM
0 4 * * 0 /usr/local/bin/moodle-config-backup.sh

# Backup verification daily at 5 AM
0 5 * * * /usr/local/bin/backup-verification.sh
```

### Step 6: Make Scripts Executable

```bash
# Make all backup scripts executable
sudo chmod +x /usr/local/bin/moodle-db-backup.sh
sudo chmod +x /usr/local/bin/moodle-file-backup.sh
sudo chmod +x /usr/local/bin/moodle-config-backup.sh
sudo chmod +x /usr/local/bin/backup-verification.sh
```

## ✅ Verification

### Backup Test

```bash
# Test backup scripts
sudo /usr/local/bin/moodle-db-backup.sh
sudo /usr/local/bin/moodle-file-backup.sh
sudo /usr/local/bin/moodle-config-backup.sh

# Check backup files
ls -la /backup/moodle/database/
ls -la /backup/moodle/files/
ls -la /backup/moodle/config/

# Test backup verification
sudo /usr/local/bin/backup-verification.sh
tail -f /var/log/backup-verification.log
```

### Expected Results

- ✅ Database backup working
- ✅ File backup working
- ✅ Configuration backup working
- ✅ Backup verification active
- ✅ Cron jobs configured

## 📝 Next Steps

Setelah backup strategy selesai, lanjutkan ke:
- [Phase 4: Production Deployment](../phase-4-deployment/README.md)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
