# 🚀 Moodle 4.0 Installation

## 📋 Overview

Dokumen ini menjelaskan proses instalasi Moodle 4.0, termasuk download, extract, konfigurasi, dan setup initial settings.

## 🎯 Objectives

- [ ] Download dan extract Moodle 4.0
- [ ] Setup file permissions dan ownership
- [ ] Konfigurasi database connection
- [ ] Jalankan Moodle installation wizard
- [ ] Konfigurasi initial settings

## 🔧 Step-by-Step Guide

### Step 1: Download Moodle 4.0

```bash
# Create temporary directory
mkdir -p /tmp/moodle-install
cd /tmp/moodle-install

# Download Moodle 4.0
wget https://download.moodle.org/releases/latest/moodle-4.0-latest.tgz

# Verify download
ls -la moodle-4.0-latest.tgz
```

### Step 2: Extract Moodle

```bash
# Extract Moodle
tar -xzf moodle-4.0-latest.tgz

# Check extracted files
ls -la moodle/

# Move to web directory
sudo mv moodle/* /var/www/moodle/

# Set ownership
sudo chown -R www-data:www-data /var/www/moodle

# Set permissions
sudo chmod -R 755 /var/www/moodle
```

### Step 3: Create Moodle Data Directory

```bash
# Create data directory
sudo mkdir -p /var/www/moodledata

# Set ownership
sudo chown -R www-data:www-data /var/www/moodledata

# Set permissions
sudo chmod -R 777 /var/www/moodledata
```

### Step 4: Configure Database Connection

```bash
# Create config.php
sudo nano /var/www/moodle/config.php
```

**Moodle Configuration:**
```php
<?php
// Moodle 4.0 Configuration

unset($CFG);
global $CFG;
$CFG = new stdClass();

// Database configuration
$CFG->dbtype    = 'mariadb';
$CFG->dblibrary = 'native';
$CFG->dbhost    = 'localhost';
$CFG->dbname    = 'moodle';
$CFG->dbuser    = 'moodle';
$CFG->dbpass    = 'your_password_here';
$CFG->prefix    = 'mdl_';
$CFG->dboptions = array(
    'dbpersist' => 0,
    'dbport' => 3306,
    'dbsocket' => '',
    'dbcollation' => 'utf8mb4_unicode_ci',
);

// Directory configuration
$CFG->wwwroot   = 'https://lms.yourdomain.com';
$CFG->dataroot  = '/var/www/moodledata';
$CFG->admin     = 'admin';

// Security configuration
$CFG->directorypermissions = 0777;
$CFG->filepermissions = 0644;
$CFG->dirpermissions = 0755;

// Performance configuration
$CFG->preventexecpath = true;
$CFG->pathtophp = '/usr/bin/php';
$CFG->pathtodu = '/usr/bin/du';
$CFG->aspellpath = '/usr/bin/aspell';
$CFG->pathtodot = '/usr/bin/dot';

// Session configuration
$CFG->session_handler_class = '\core\session\redis';
$CFG->session_redis_host = '127.0.0.1';
$CFG->session_redis_port = 6379;
$CFG->session_redis_database = 0;
$CFG->session_redis_prefix = 'moodle_session_';

// Cache configuration
$CFG->cachejs = true;
$CFG->cachetemplates = true;

// Logging configuration
$CFG->loglifetime = 0;
$CFG->logguests = true;

// Email configuration
$CFG->smtphosts = 'localhost';
$CFG->smtpuser = '';
$CFG->smtppass = '';
$CFG->smtpsecure = '';
$CFG->smtpauthtype = '';

// This is for PHP settings which can be used in case you cannot edit php.ini
$CFG->phpunit_dataroot = '/var/www/moodledata/phpunit';
$CFG->phpunit_prefix = 'phpu_';

require_once(__DIR__ . '/lib/setup.php');
```

### Step 5: Set File Permissions

```bash
# Set proper permissions
sudo chown -R www-data:www-data /var/www/moodle
sudo chmod -R 755 /var/www/moodle

# Set config.php permissions
sudo chmod 644 /var/www/moodle/config.php

# Set data directory permissions
sudo chmod -R 777 /var/www/moodledata
```

### Step 6: Run Moodle Installation

```bash
# Access Moodle in browser
# Go to: https://lms.yourdomain.com

# Or run CLI installation
cd /var/www/moodle
sudo -u www-data php admin/cli/install.php
```

**CLI Installation Parameters:**
```
--lang=en
--wwwroot=https://lms.yourdomain.com
--dataroot=/var/www/moodledata
--dbtype=mariadb
--dbhost=localhost
--dbname=moodle
--dbuser=moodle
--dbpass=your_password
--dbport=3306
--prefix=mdl_
--fullname="LMS K2NET"
--shortname="LMSK2"
--adminuser=admin
--adminpass=admin_password
--adminemail=admin@yourdomain.com
--agree-license
--non-interactive
```

### Step 7: Configure Initial Settings

```bash
# Set up cron job
sudo crontab -e
```

**Add to crontab:**
```
# Moodle cron job
*/5 * * * * /usr/bin/php /var/www/moodle/admin/cli/cron.php >/dev/null
```

### Step 8: Configure Redis Session

```bash
# Install Redis
sudo apt install -y redis-server

# Configure Redis
sudo nano /etc/redis/redis.conf
```

**Redis Configuration:**
```
bind 127.0.0.1
port 6379
timeout 0
tcp-keepalive 300

maxmemory 512mb
maxmemory-policy allkeys-lru

save 900 1
save 300 10
save 60 10000
```

```bash
# Start Redis
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

### Step 9: Configure OPcache

```bash
# Configure OPcache
sudo nano /etc/php/8.1/fpm/conf.d/99-opcache.ini
```

**OPcache Configuration:**
```ini
; OPcache Configuration for Moodle 4.0
opcache.enable = 1
opcache.memory_consumption = 256
opcache.interned_strings_buffer = 8
opcache.max_accelerated_files = 10000
opcache.revalidate_freq = 2
opcache.fast_shutdown = 1
opcache.validate_timestamps = 0
opcache.save_comments = 1
opcache.enable_file_override = 1
```

### Step 10: Restart Services

```bash
# Restart PHP-FPM
sudo systemctl restart php8.1-fpm

# Restart Nginx
sudo systemctl restart nginx

# Restart Redis
sudo systemctl restart redis-server
```

## ✅ Verification

### Installation Test

```bash
# Test Moodle installation
curl -I https://lms.yourdomain.com

# Test database connection
mysql -u moodle -p moodle -e "SELECT COUNT(*) FROM mdl_user;"

# Test Redis connection
redis-cli ping

# Test cron job
sudo -u www-data php /var/www/moodle/admin/cli/cron.php
```

### Expected Results

- ✅ Moodle accessible via web browser
- ✅ Database connection working
- ✅ Redis session working
- ✅ Cron job running
- ✅ File permissions correct
- ✅ SSL certificate working

## 🚨 Troubleshooting

### Common Issues

**1. Installation wizard not accessible**
```bash
# Check file permissions
sudo chown -R www-data:www-data /var/www/moodle
sudo chmod -R 755 /var/www/moodle
```

**2. Database connection error**
```bash
# Check database credentials
mysql -u moodle -p moodle -e "SELECT 1;"

# Check config.php
sudo nano /var/www/moodle/config.php
```

**3. Redis connection error**
```bash
# Check Redis status
sudo systemctl status redis-server

# Test Redis connection
redis-cli ping
```

## 📝 Next Steps

Setelah Moodle 4.0 installation selesai, lanjutkan ke:
- [05-verification.md](05-verification.md) - Verifikasi instalasi

## 📚 References

- [Moodle 4.0 Installation](https://docs.moodle.org/400/en/Installation)
- [Moodle CLI Installation](https://docs.moodle.org/400/en/Installation_via_command_line)
- [Moodle Configuration](https://docs.moodle.org/400/en/Configuration_file)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
