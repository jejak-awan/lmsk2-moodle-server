# 🎓 Moodle 3.11 LTS Installation

## 📋 Overview

Dokumen ini menjelaskan instalasi Moodle 3.11 LTS (Long Term Support) yang merupakan versi stabil dan direkomendasikan untuk production environment.

## 🎯 Objectives

- [ ] Download dan extract Moodle 3.11 LTS
- [ ] Setup file permissions dan ownership
- [ ] Konfigurasi database connection
- [ ] Jalankan Moodle installation wizard
- [ ] Konfigurasi initial settings
- [ ] Verifikasi instalasi

## 📋 Prerequisites

- Server preparation sudah selesai
- Database sudah dikonfigurasi
- Web server sudah dikonfigurasi
- SSL certificate sudah terinstall
- Domain name sudah dikonfigurasi

## 🔧 Step-by-Step Guide

### Step 1: Download Moodle 3.11 LTS

```bash
# Create temporary directory
sudo mkdir -p /tmp/moodle-install
cd /tmp/moodle-install

# Download Moodle 3.11 LTS
sudo wget https://download.moodle.org/releases/latest311/moodle-latest-311.tgz

# Verify download
sudo ls -la moodle-latest-311.tgz

# Extract Moodle
sudo tar -xzf moodle-latest-311.tgz

# Check extracted files
sudo ls -la moodle/
```

### Step 2: Move Moodle to Web Directory

```bash
# Stop web server temporarily
sudo systemctl stop nginx

# Move Moodle files to web directory
sudo mv moodle/* /var/www/moodle/

# Set proper ownership
sudo chown -R www-data:www-data /var/www/moodle

# Set proper permissions
sudo find /var/www/moodle -type d -exec chmod 755 {} \;
sudo find /var/www/moodle -type f -exec chmod 644 {} \;

# Make moodledata writable
sudo chmod -R 777 /var/www/moodle/moodledata

# Start web server
sudo systemctl start nginx
```

### Step 3: Create Configuration File

```bash
# Create Moodle configuration file
sudo nano /var/www/moodle/config.php
```

**Moodle Configuration (config.php):**
```php
<?php  // Moodle configuration file

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

// Web server configuration
$CFG->wwwroot   = 'https://lms.yourdomain.com';
$CFG->dataroot  = '/var/www/moodle/moodledata';
$CFG->admin     = 'admin';

// Directory permissions
$CFG->directorypermissions = 0777;

// Security settings
$CFG->sslproxy = true;
$CFG->cookiehttponly = true;
$CFG->cookiesecure = true;

// Performance settings
$CFG->preventexecpath = true;
$CFG->pathtophp = '/usr/bin/php';
$CFG->pathtodu = '/usr/bin/du';
$CFG->aspellpath = '/usr/bin/aspell';
$CFG->pathtodot = '/usr/bin/dot';

// Caching configuration
$CFG->cachejs = true;
$CFG->cachetemplates = true;

// Logging configuration
$CFG->loglifetime = 30;
$CFG->logguests = true;

// Email configuration (configure later)
$CFG->smtphosts = 'localhost:25';
$CFG->smtpuser = '';
$CFG->smtppass = '';
$CFG->smtpsecure = '';
$CFG->smtpauthtype = '';

// File upload settings
$CFG->maxbytes = 200*1024*1024; // 200MB
$CFG->maxareabytes = 200*1024*1024; // 200MB

// Session configuration
$CFG->session_handler_class = '\core\session\redis';
$CFG->session_redis_host = '127.0.0.1';
$CFG->session_redis_port = 6379;
$CFG->session_redis_database = 0;
$CFG->session_redis_prefix = 'mdl_';

// Additional settings
$CFG->debug = false;
$CFG->debugdeveloper = false;
$CFG->perfdebug = false;
$CFG->debugpageinfo = false;
$CFG->allowthemechangeonurl = false;
$CFG->passwordpolicy = true;
$CFG->minpasswordlength = 8;
$CFG->minpassworddigits = 1;
$CFG->minpasswordlower = 1;
$CFG->minpasswordupper = 1;
$CFG->minpasswordnonalphanum = 1;

// Timezone
$CFG->timezone = 'Asia/Jakarta';

// Language
$CFG->lang = 'en';

// Site settings
$CFG->fullnamedisplay = 'firstname lastname';
$CFG->maxusersperpage = 50;
$CFG->maxcoursesperpage = 20;

// Maintenance mode
$CFG->maintenance_enabled = false;
$CFG->maintenance_message = 'This site is currently being upgraded and is not available. Please try again later.';

require_once(__DIR__ . '/lib/setup.php');
```

### Step 4: Set Configuration File Permissions

```bash
# Set secure permissions for config.php
sudo chmod 600 /var/www/moodle/config.php
sudo chown root:www-data /var/www/moodle/config.php
```

### Step 5: Run Moodle Installation

```bash
# Access Moodle installation via web browser
# Go to: https://lms.yourdomain.com

# Or run installation via command line
cd /var/www/moodle
sudo -u www-data php admin/cli/install.php \
    --lang=en \
    --wwwroot=https://lms.yourdomain.com \
    --dataroot=/var/www/moodle/moodledata \
    --dbtype=mariadb \
    --dbhost=localhost \
    --dbname=moodle \
    --dbuser=moodle \
    --dbpass=your_password_here \
    --fullname="LMS K2 Moodle Server" \
    --shortname="LMS" \
    --adminuser=admin \
    --adminpass=admin_password_here \
    --adminemail=admin@yourdomain.com \
    --agree-license
```

### Step 6: Post-Installation Configuration

```bash
# Create post-installation script
sudo nano /usr/local/bin/moodle-post-install.sh
```

**Post-Installation Script:**
```bash
#!/bin/bash

# Moodle Post-Installation Configuration
echo "=== Moodle Post-Installation Configuration ==="

# Set proper permissions
echo "Setting file permissions..."
sudo chown -R www-data:www-data /var/www/moodle
sudo chmod -R 755 /var/www/moodle
sudo chmod -R 777 /var/www/moodle/moodledata
sudo chmod 600 /var/www/moodle/config.php
sudo chown root:www-data /var/www/moodle/config.php

# Create cron job for Moodle
echo "Setting up Moodle cron job..."
sudo crontab -e
# Add: */5 * * * * /usr/bin/php /var/www/moodle/admin/cli/cron.php >/dev/null

# Install additional plugins (optional)
echo "Installing additional plugins..."
cd /var/www/moodle
sudo -u www-data php admin/cli/upgrade.php --non-interactive

# Clear caches
echo "Clearing caches..."
sudo -u www-data php admin/cli/purge_caches.php

# Update database
echo "Updating database..."
sudo -u www-data php admin/cli/upgrade.php --non-interactive

echo "Post-installation configuration completed"
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/moodle-post-install.sh

# Run post-installation
sudo /usr/local/bin/moodle-post-install.sh
```

### Step 7: Configure Moodle Cron Job

```bash
# Add Moodle cron job
sudo crontab -e
```

**Add to crontab:**
```
# Moodle cron job (every 5 minutes)
*/5 * * * * /usr/bin/php /var/www/moodle/admin/cli/cron.php >/dev/null

# Moodle maintenance (daily at 2 AM)
0 2 * * * /usr/bin/php /var/www/moodle/admin/cli/maintenance.php --enable
0 3 * * * /usr/bin/php /var/www/moodle/admin/cli/upgrade.php --non-interactive
0 4 * * * /usr/bin/php /var/www/moodle/admin/cli/maintenance.php --disable
```

### Step 8: Initial Moodle Configuration

```bash
# Create initial configuration script
sudo nano /usr/local/bin/moodle-initial-config.sh
```

**Initial Configuration Script:**
```bash
#!/bin/bash

# Moodle Initial Configuration
echo "=== Moodle Initial Configuration ==="

# Set site settings
cd /var/www/moodle
sudo -u www-data php admin/cli/cfg.php --name=sitename --set="LMS K2 Moodle Server"
sudo -u www-data php admin/cli/cfg.php --name=sitesummary --set="Professional Learning Management System"

# Configure email settings
sudo -u www-data php admin/cli/cfg.php --name=smtphosts --set="localhost:25"
sudo -u www-data php admin/cli/cfg.php --name=noreplyaddress --set="noreply@yourdomain.com"

# Configure file upload settings
sudo -u www-data php admin/cli/cfg.php --name=maxbytes --set=209715200
sudo -u www-data php admin/cli/cfg.php --name=maxareabytes --set=209715200

# Configure session settings
sudo -u www-data php admin/cli/cfg.php --name=session_handler_class --set="\core\session\redis"
sudo -u www-data php admin/cli/cfg.php --name=session_redis_host --set="127.0.0.1"
sudo -u www-data php admin/cli/cfg.php --name=session_redis_port --set=6379

# Configure caching
sudo -u www-data php admin/cli/cfg.php --name=cachejs --set=1
sudo -u www-data php admin/cli/cfg.php --name=cachetemplates --set=1

# Configure security settings
sudo -u www-data php admin/cli/cfg.php --name=passwordpolicy --set=1
sudo -u www-data php admin/cli/cfg.php --name=minpasswordlength --set=8
sudo -u www-data php admin/cli/cfg.php --name=minpassworddigits --set=1
sudo -u www-data php admin/cli/cfg.php --name=minpasswordlower --set=1
sudo -u www-data php admin/cli/cfg.php --name=minpasswordupper --set=1
sudo -u www-data php admin/cli/cfg.php --name=minpasswordnonalphanum --set=1

# Configure timezone
sudo -u www-data php admin/cli/cfg.php --name=timezone --set="Asia/Jakarta"

# Configure language
sudo -u www-data php admin/cli/cfg.php --name=lang --set="en"

# Clear caches
sudo -u www-data php admin/cli/purge_caches.php

echo "Initial configuration completed"
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/moodle-initial-config.sh

# Run initial configuration
sudo /usr/local/bin/moodle-initial-config.sh
```

## ✅ Verification

### Installation Verification

```bash
# Check Moodle installation
curl -I https://lms.yourdomain.com

# Check database connection
mysql -u moodle -p moodle -e "SELECT COUNT(*) as tables FROM information_schema.tables WHERE table_schema = 'moodle';"

# Check file permissions
ls -la /var/www/moodle/config.php
ls -la /var/www/moodle/moodledata/

# Check cron job
sudo crontab -l | grep moodle

# Check Moodle status
cd /var/www/moodle
sudo -u www-data php admin/cli/upgrade.php --non-interactive
```

### Web Interface Verification

1. **Access Moodle**: https://lms.yourdomain.com
2. **Login as admin**: Use admin credentials
3. **Check site administration**: Go to Site administration
4. **Verify settings**: Check all configuration settings
5. **Test functionality**: Create a test course

### Expected Results

- ✅ Moodle accessible via web browser
- ✅ Admin login working
- ✅ Database connection established
- ✅ File permissions correct
- ✅ Cron job running
- ✅ All configurations applied
- ✅ Site administration accessible

## 🚨 Troubleshooting

### Common Issues

**1. Installation wizard not accessible**
```bash
# Check file permissions
sudo chown -R www-data:www-data /var/www/moodle
sudo chmod -R 755 /var/www/moodle

# Check Nginx configuration
sudo nginx -t
sudo systemctl status nginx
```

**2. Database connection failed**
```bash
# Check database credentials in config.php
sudo cat /var/www/moodle/config.php | grep db

# Test database connection
mysql -u moodle -p moodle -e "SELECT 1;"
```

**3. File permission errors**
```bash
# Fix permissions
sudo chown -R www-data:www-data /var/www/moodle
sudo chmod -R 755 /var/www/moodle
sudo chmod -R 777 /var/www/moodle/moodledata
sudo chmod 600 /var/www/moodle/config.php
```

**4. Cron job not working**
```bash
# Check cron service
sudo systemctl status cron

# Test cron manually
sudo -u www-data php /var/www/moodle/admin/cli/cron.php

# Check cron logs
sudo tail -f /var/log/cron.log
```

## 📝 Next Steps

Setelah Moodle installation selesai, lanjutkan ke:
- [05-verification.md](05-verification.md) - Verifikasi instalasi Moodle 3.11 LTS

## 📚 References

- [Moodle 3.11 LTS Installation Guide](https://docs.moodle.org/311/en/Installation)
- [Moodle Command Line Interface](https://docs.moodle.org/311/en/Command_line_interface)
- [Moodle Configuration](https://docs.moodle.org/311/en/Configuration)
- [Moodle Cron](https://docs.moodle.org/311/en/Cron)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
