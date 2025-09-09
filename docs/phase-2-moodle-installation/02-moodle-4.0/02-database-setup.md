# 🗄️ Database Setup for Moodle 4.0

## 📋 Overview

Dokumen ini menjelaskan setup database untuk Moodle 4.0, termasuk instalasi MariaDB/MySQL, konfigurasi database, dan optimasi performa.

## 🎯 Objectives

- [ ] Install dan konfigurasi MariaDB/MySQL
- [ ] Buat database dan user untuk Moodle 4.0
- [ ] Konfigurasi database untuk performa optimal
- [ ] Setup backup dan recovery
- [ ] Verifikasi koneksi database

## 🔧 Step-by-Step Guide

### Step 1: Install MariaDB/MySQL

```bash
# Update package list
sudo apt update

# Install MariaDB server
sudo apt install -y mariadb-server mariadb-client

# Start and enable MariaDB
sudo systemctl start mariadb
sudo systemctl enable mariadb

# Check status
sudo systemctl status mariadb
```

### Step 2: Secure MariaDB Installation

```bash
# Run security script
sudo mysql_secure_installation
```

**Security Configuration:**
```
Enter current password for root (enter for none): [Press Enter]
Set root password? [Y/n] Y
New password: [Enter strong password]
Re-enter new password: [Confirm password]
Remove anonymous users? [Y/n] Y
Disallow root login remotely? [Y/n] Y
Remove test database and access to it? [Y/n] Y
Reload privilege tables now? [Y/n] Y
```

### Step 3: Create Database and User

```bash
# Login to MariaDB
sudo mysql -u root -p

# Create database
CREATE DATABASE moodle CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# Create user
CREATE USER 'moodle'@'localhost' IDENTIFIED BY 'strong_password_here';

# Grant privileges
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, INDEX, ALTER ON moodle.* TO 'moodle'@'localhost';

# Flush privileges
FLUSH PRIVILEGES;

# Exit MariaDB
EXIT;
```

### Step 4: Configure MariaDB for Moodle 4.0

```bash
# Create MariaDB configuration for Moodle
sudo nano /etc/mysql/mariadb.conf.d/99-moodle.cnf
```

**MariaDB Configuration:**
```ini
[mysqld]
# Basic settings
default-storage-engine = InnoDB
innodb_file_per_table = 1
innodb_file_format = Barracuda

# Character set
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci
init-connect = 'SET NAMES utf8mb4'

# InnoDB settings
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
innodb_log_buffer_size = 16M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT

# Connection settings
max_connections = 200
max_connect_errors = 10000
connect_timeout = 10
wait_timeout = 600
interactive_timeout = 600

# Query cache
query_cache_type = 1
query_cache_size = 128M
query_cache_limit = 2M

# Temporary tables
tmp_table_size = 128M
max_heap_table_size = 128M

# MyISAM settings
key_buffer_size = 32M
read_buffer_size = 2M
read_rnd_buffer_size = 8M
sort_buffer_size = 2M

# Logging
log-error = /var/log/mysql/error.log
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2

# Security
local-infile = 0
```

### Step 5: Restart MariaDB

```bash
# Restart MariaDB
sudo systemctl restart mariadb

# Check status
sudo systemctl status mariadb
```

### Step 6: Test Database Connection

```bash
# Test connection
mysql -u moodle -p -e "SHOW DATABASES;"

# Test specific database
mysql -u moodle -p moodle -e "SELECT 1;"
```

### Step 7: Setup Database Backup

```bash
# Create backup script
sudo nano /usr/local/bin/moodle-db-backup.sh
```

**Database Backup Script:**
```bash
#!/bin/bash

# Moodle 4.0 Database Backup Script
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
    $DB_NAME | gzip > $BACKUP_DIR/moodle_4.0_$DATE.sql.gz

# Cleanup old backups (keep 7 days)
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete

echo "Database backup completed at $(date)"
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/moodle-db-backup.sh

# Test backup
sudo /usr/local/bin/moodle-db-backup.sh
```

### Step 8: Setup Automated Backup

```bash
# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# Database backup daily at 2 AM
0 2 * * * /usr/local/bin/moodle-db-backup.sh
```

## ✅ Verification

### Database Connection Test

```bash
# Test database connection
mysql -u moodle -p moodle -e "SELECT VERSION();"

# Test database performance
mysql -u moodle -p moodle -e "SHOW STATUS LIKE 'Questions';"

# Check database size
mysql -u moodle -p moodle -e "SELECT table_schema AS 'Database', ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS 'Size (MB)' FROM information_schema.tables WHERE table_schema = 'moodle' GROUP BY table_schema;"
```

### Expected Results

- ✅ MariaDB running and accessible
- ✅ Database 'moodle' created
- ✅ User 'moodle' created with proper privileges
- ✅ Database configuration optimized
- ✅ Backup system working
- ✅ Connection test successful

## 🚨 Troubleshooting

### Common Issues

**1. Connection refused**
```bash
# Check MariaDB status
sudo systemctl status mariadb

# Check MariaDB logs
sudo tail -f /var/log/mysql/error.log
```

**2. Access denied**
```bash
# Reset root password
sudo mysql -u root
ALTER USER 'root'@'localhost' IDENTIFIED BY 'new_password';
FLUSH PRIVILEGES;
```

**3. Database not found**
```bash
# Recreate database
mysql -u root -p
CREATE DATABASE moodle CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 📝 Next Steps

Setelah database setup selesai, lanjutkan ke:
- [03-web-server-config.md](03-web-server-config.md) - Setup web server

## 📚 References

- [MariaDB Documentation](https://mariadb.org/documentation/)
- [MySQL Documentation](https://dev.mysql.com/doc/)
- [Moodle Database Requirements](https://docs.moodle.org/400/en/Database)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
