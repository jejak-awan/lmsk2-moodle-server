# 🗄️ Database Setup for Moodle 3.11 LTS

## 📋 Overview

Dokumen ini menjelaskan setup database untuk Moodle 3.11 LTS, termasuk instalasi, konfigurasi, dan optimasi database server yang kompatibel.

## 🎯 Objectives

- [ ] Install dan konfigurasi MariaDB/MySQL
- [ ] Buat database dan user untuk Moodle
- [ ] Konfigurasi database untuk performa optimal
- [ ] Setup backup dan recovery
- [ ] Verifikasi koneksi database

## 📋 Prerequisites

- Server Ubuntu 22.04 LTS sudah dikonfigurasi
- User `moodle` sudah dibuat
- Firewall sudah dikonfigurasi
- Requirements check sudah passed

## 🔧 Step-by-Step Guide

### Step 1: Install MariaDB

```bash
# Update package list
sudo apt update

# Install MariaDB server and client
sudo apt install -y mariadb-server mariadb-client

# Start and enable MariaDB
sudo systemctl start mariadb
sudo systemctl enable mariadb

# Check MariaDB status
sudo systemctl status mariadb
```

### Step 2: Secure MariaDB Installation

```bash
# Run MariaDB secure installation
sudo mysql_secure_installation
```

**MariaDB Secure Installation:**
```
Enter current password for root (enter for none): [Press Enter]
Set root password? [Y/n]: Y
New password: [Enter strong password]
Re-enter new password: [Confirm password]
Remove anonymous users? [Y/n]: Y
Disallow root login remotely? [Y/n]: Y
Remove test database and access to it? [Y/n]: Y
Reload privilege tables now? [Y/n]: Y
```

### Step 3: Create Database and User

```bash
# Login to MariaDB as root
sudo mysql -u root -p

# Create database and user for Moodle
CREATE DATABASE moodle CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'moodle'@'localhost' IDENTIFIED BY 'strong_password_here';
GRANT ALL PRIVILEGES ON moodle.* TO 'moodle'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

**Database Configuration:**
```sql
-- Verify database creation
SHOW DATABASES;

-- Verify user creation
SELECT user, host FROM mysql.user WHERE user = 'moodle';

-- Test connection with moodle user
-- (Run this from command line)
mysql -u moodle -p moodle
```

### Step 4: MariaDB Configuration for Moodle 3.11 LTS

```bash
# Create MariaDB configuration for Moodle
sudo nano /etc/mysql/mariadb.conf.d/50-moodle.cnf
```

**MariaDB Configuration:**
```ini
[mysqld]
# Basic settings
bind-address = 127.0.0.1
port = 3306
socket = /var/run/mysqld/mysqld.sock

# Character set and collation
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci
init-connect = 'SET NAMES utf8mb4'

# InnoDB settings for Moodle 3.11 LTS
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
innodb_log_buffer_size = 16M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT
innodb_file_per_table = 1
innodb_open_files = 400
innodb_io_capacity = 400
innodb_read_io_threads = 4
innodb_write_io_threads = 4

# Query cache (for MySQL 5.7 compatibility)
query_cache_type = 1
query_cache_size = 64M
query_cache_limit = 2M

# Connection settings
max_connections = 200
max_connect_errors = 10000
connect_timeout = 10
wait_timeout = 600
interactive_timeout = 600

# Temporary tables
tmp_table_size = 64M
max_heap_table_size = 64M

# MyISAM settings
key_buffer_size = 32M
read_buffer_size = 2M
read_rnd_buffer_size = 8M
sort_buffer_size = 2M

# Logging
log_error = /var/log/mysql/error.log
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2
log_queries_not_using_indexes = 1

# Binary logging
log_bin = /var/log/mysql/mysql-bin.log
expire_logs_days = 7
max_binlog_size = 100M

# Security
local-infile = 0
symbolic-links = 0
skip-networking = 0
```

### Step 5: Restart and Verify MariaDB

```bash
# Restart MariaDB to apply configuration
sudo systemctl restart mariadb

# Check MariaDB status
sudo systemctl status mariadb

# Verify configuration
sudo mysql -u root -p -e "SHOW VARIABLES LIKE 'character_set%';"
sudo mysql -u root -p -e "SHOW VARIABLES LIKE 'innodb_buffer_pool_size';"
```

### Step 6: Database Performance Tuning

```bash
# Create performance tuning script
sudo nano /usr/local/bin/mysql-performance-tuning.sh
```

**Performance Tuning Script:**
```bash
#!/bin/bash

# MySQL Performance Tuning for Moodle 3.11 LTS
echo "=== MySQL Performance Tuning ==="

# Login to MySQL
mysql -u root -p << EOF

-- Optimize tables
USE moodle;
OPTIMIZE TABLE mdl_user;
OPTIMIZE TABLE mdl_course;
OPTIMIZE TABLE mdl_log;

-- Analyze tables
ANALYZE TABLE mdl_user;
ANALYZE TABLE mdl_course;
ANALYZE TABLE mdl_log;

-- Check table status
SHOW TABLE STATUS;

-- Check process list
SHOW PROCESSLIST;

-- Check variables
SHOW VARIABLES LIKE 'innodb_buffer_pool_size';
SHOW VARIABLES LIKE 'query_cache_size';
SHOW VARIABLES LIKE 'max_connections';

EOF

echo "Performance tuning completed"
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/mysql-performance-tuning.sh
```

### Step 7: Database Backup Configuration

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
echo "Creating full database backup..."
mysqldump -u $DB_USER -p$DB_PASS \
    --single-transaction \
    --routines \
    --triggers \
    --events \
    --hex-blob \
    --opt \
    $DB_NAME | gzip > $BACKUP_DIR/moodle_full_$DATE.sql.gz

# Schema only backup
echo "Creating schema backup..."
mysqldump -u $DB_USER -p$DB_PASS \
    --no-data \
    --routines \
    --triggers \
    --events \
    $DB_NAME | gzip > $BACKUP_DIR/moodle_schema_$DATE.sql.gz

# Data only backup
echo "Creating data backup..."
mysqldump -u $DB_USER -p$DB_PASS \
    --no-create-info \
    --single-transaction \
    --hex-blob \
    $DB_NAME | gzip > $BACKUP_DIR/moodle_data_$DATE.sql.gz

# Cleanup old backups (keep 7 days)
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete

echo "Database backup completed at $(date)"
echo "Backup files:"
ls -lh $BACKUP_DIR/*$DATE*
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/moodle-db-backup.sh

# Test backup
sudo /usr/local/bin/moodle-db-backup.sh
```

### Step 8: Database Monitoring

```bash
# Create database monitoring script
sudo nano /usr/local/bin/mysql-monitor.sh
```

**Database Monitoring Script:**
```bash
#!/bin/bash

# MySQL Monitoring Script for Moodle
LOG_FILE="/var/log/mysql-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] MySQL monitoring..." >> $LOG_FILE

# Check MySQL status
if systemctl is-active --quiet mariadb; then
    echo "[$DATE] ✓ MariaDB is running" >> $LOG_FILE
else
    echo "[$DATE] ✗ MariaDB is not running" >> $LOG_FILE
    systemctl restart mariadb
fi

# Check connections
CONNECTIONS=$(mysql -u root -p'your_password' -e "SHOW STATUS LIKE 'Threads_connected';" | awk 'NR==2 {print $2}')
echo "[$DATE] Active connections: $CONNECTIONS" >> $LOG_FILE

# Check slow queries
SLOW_QUERIES=$(mysql -u root -p'your_password' -e "SHOW STATUS LIKE 'Slow_queries';" | awk 'NR==2 {print $2}')
echo "[$DATE] Slow queries: $SLOW_QUERIES" >> $LOG_FILE

# Check table locks
TABLE_LOCKS=$(mysql -u root -p'your_password' -e "SHOW STATUS LIKE 'Table_locks_waited';" | awk 'NR==2 {print $2}')
echo "[$DATE] Table locks waited: $TABLE_LOCKS" >> $LOG_FILE

# Check InnoDB status
INNODB_BUFFER_POOL_HIT=$(mysql -u root -p'your_password' -e "SHOW STATUS LIKE 'Innodb_buffer_pool_read_requests';" | awk 'NR==2 {print $2}')
INNODB_BUFFER_POOL_MISS=$(mysql -u root -p'your_password' -e "SHOW STATUS LIKE 'Innodb_buffer_pool_reads';" | awk 'NR==2 {print $2}')
if [ $INNODB_BUFFER_POOL_HIT -gt 0 ]; then
    HIT_RATIO=$((INNODB_BUFFER_POOL_HIT * 100 / (INNODB_BUFFER_POOL_HIT + INNODB_BUFFER_POOL_MISS)))
    echo "[$DATE] InnoDB buffer pool hit ratio: $HIT_RATIO%" >> $LOG_FILE
fi

echo "[$DATE] MySQL monitoring completed" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/mysql-monitor.sh

# Add to crontab (every 5 minutes)
sudo crontab -e
```

**Add to crontab:**
```
# MySQL monitoring every 5 minutes
*/5 * * * * /usr/local/bin/mysql-monitor.sh
```

## ✅ Verification

### Database Connection Test

```bash
# Test database connection
mysql -u moodle -p moodle -e "SELECT 1 as test;"

# Test database creation
mysql -u moodle -p moodle -e "CREATE TABLE test_table (id INT PRIMARY KEY, name VARCHAR(50));"
mysql -u moodle -p moodle -e "INSERT INTO test_table VALUES (1, 'test');"
mysql -u moodle -p moodle -e "SELECT * FROM test_table;"
mysql -u moodle -p moodle -e "DROP TABLE test_table;"
```

### Performance Check

```bash
# Check MariaDB performance
sudo mysql -u root -p -e "SHOW STATUS LIKE 'Uptime';"
sudo mysql -u root -p -e "SHOW STATUS LIKE 'Questions';"
sudo mysql -u root -p -e "SHOW STATUS LIKE 'Slow_queries';"
sudo mysql -u root -p -e "SHOW STATUS LIKE 'Connections';"
```

### Expected Results

- ✅ MariaDB running and accessible
- ✅ Database `moodle` created with utf8mb4 charset
- ✅ User `moodle` created with proper permissions
- ✅ Configuration optimized for Moodle 3.11 LTS
- ✅ Backup system configured
- ✅ Monitoring system active

## 🚨 Troubleshooting

### Common Issues

**1. MariaDB won't start**
```bash
# Check error logs
sudo tail -f /var/log/mysql/error.log

# Check configuration
sudo mysql --help --verbose | head -20

# Reset root password if needed
sudo systemctl stop mariadb
sudo mysqld_safe --skip-grant-tables &
mysql -u root
```

**2. Connection refused**
```bash
# Check if MariaDB is listening
sudo netstat -tlnp | grep 3306

# Check bind-address in configuration
sudo grep bind-address /etc/mysql/mariadb.conf.d/50-moodle.cnf
```

**3. Character set issues**
```bash
# Check current character set
mysql -u root -p -e "SHOW VARIABLES LIKE 'character_set%';"

# Fix character set
mysql -u root -p -e "ALTER DATABASE moodle CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

**4. Performance issues**
```bash
# Check slow query log
sudo tail -f /var/log/mysql/slow.log

# Check process list
mysql -u root -p -e "SHOW PROCESSLIST;"

# Check table status
mysql -u root -p moodle -e "SHOW TABLE STATUS;"
```

## 📝 Next Steps

Setelah database setup selesai, lanjutkan ke:
- [03-web-server-config.md](03-web-server-config.md) - Konfigurasi web server untuk Moodle 3.11 LTS

## 📚 References

- [MariaDB Documentation](https://mariadb.org/documentation/)
- [MySQL Performance Tuning](https://dev.mysql.com/doc/refman/8.0/en/optimization.html)
- [Moodle Database Requirements](https://docs.moodle.org/311/en/Database_requirements)
- [UTF8MB4 Support](https://mariadb.org/documentation/mariadb/character-sets-collations/)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
