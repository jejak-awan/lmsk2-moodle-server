# 📦 Software Installation

## 📋 Overview

Dokumen ini menjelaskan instalasi software yang diperlukan untuk menjalankan Moodle LMS, termasuk web server, database, PHP, dan tools pendukung lainnya.

## 🎯 Objectives

- [ ] Install dan konfigurasi Nginx web server
- [ ] Install PHP dengan extensions yang diperlukan
- [ ] Install dan konfigurasi MariaDB database
- [ ] Install Redis untuk caching
- [ ] Install tools pendukung (Composer, Node.js)

## 📋 Prerequisites

- Server Ubuntu 22.04 LTS sudah dikonfigurasi
- User `moodle` sudah dibuat
- Firewall sudah dikonfigurasi
- Koneksi internet stabil

## 🔧 Step-by-Step Guide

### Step 1: Install Nginx Web Server

```bash
# Update package list
sudo apt update

# Install Nginx
sudo apt install -y nginx

# Start and enable Nginx
sudo systemctl start nginx
sudo systemctl enable nginx

# Check Nginx status
sudo systemctl status nginx

# Test Nginx
curl -I http://localhost
```

### Step 2: Install PHP 8.1

```bash
# Add PHP repository
sudo apt install -y software-properties-common
sudo add-apt-repository ppa:ondrej/php -y
sudo apt update

# Install PHP 8.1 and extensions
sudo apt install -y php8.1-fpm php8.1-cli php8.1-common \
    php8.1-mysql php8.1-zip php8.1-gd php8.1-mbstring \
    php8.1-curl php8.1-xml php8.1-bcmath php8.1-intl \
    php8.1-soap php8.1-ldap php8.1-imagick php8.1-redis \
    php8.1-openssl php8.1-json php8.1-dom php8.1-fileinfo \
    php8.1-iconv php8.1-simplexml php8.1-tokenizer \
    php8.1-xmlreader php8.1-xmlwriter php8.1-exif \
    php8.1-ftp php8.1-gettext

# Start and enable PHP-FPM
sudo systemctl start php8.1-fpm
sudo systemctl enable php8.1-fpm

# Check PHP version
php -v
```

### Step 3: Configure PHP

```bash
# Edit PHP configuration
sudo nano /etc/php/8.1/fpm/php.ini
```

**PHP Configuration (php.ini):**
```ini
# Memory and execution time
memory_limit = 512M
max_execution_time = 300
max_input_time = 300

# File uploads
upload_max_filesize = 200M
post_max_size = 200M
max_file_uploads = 20

# Session configuration
session.gc_maxlifetime = 1440
session.save_handler = redis
session.save_path = "tcp://127.0.0.1:6379"

# OPcache configuration
opcache.enable = 1
opcache.memory_consumption = 256
opcache.interned_strings_buffer = 8
opcache.max_accelerated_files = 10000
opcache.revalidate_freq = 2
opcache.fast_shutdown = 1

# Error reporting (disable in production)
display_errors = Off
log_errors = On
error_log = /var/log/php_errors.log
```

```bash
# Restart PHP-FPM
sudo systemctl restart php8.1-fpm
```

### Step 4: Install MariaDB

```bash
# Install MariaDB
sudo apt install -y mariadb-server mariadb-client

# Start and enable MariaDB
sudo systemctl start mariadb
sudo systemctl enable mariadb

# Secure MariaDB installation
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

### Step 5: Configure MariaDB

```bash
# Login to MariaDB
sudo mysql -u root -p

# Create database and user for Moodle
CREATE DATABASE moodle CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'moodle'@'localhost' IDENTIFIED BY 'strong_password_here';
GRANT ALL PRIVILEGES ON moodle.* TO 'moodle'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

**MariaDB Configuration (/etc/mysql/mariadb.conf.d/50-server.cnf):**
```ini
[mysqld]
# Basic settings
bind-address = 127.0.0.1
port = 3306
socket = /var/run/mysqld/mysqld.sock

# Character set
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci

# InnoDB settings
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT

# Query cache
query_cache_type = 1
query_cache_size = 64M
query_cache_limit = 2M

# Connection settings
max_connections = 200
max_connect_errors = 10000
connect_timeout = 10
wait_timeout = 600

# Temporary tables
tmp_table_size = 64M
max_heap_table_size = 64M

# Logging
log_error = /var/log/mysql/error.log
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2
```

```bash
# Restart MariaDB
sudo systemctl restart mariadb
```

### Step 6: Install Redis

```bash
# Install Redis
sudo apt install -y redis-server

# Configure Redis
sudo nano /etc/redis/redis.conf
```

**Redis Configuration:**
```
# Network
bind 127.0.0.1
port 6379
timeout 0

# Memory management
maxmemory 256mb
maxmemory-policy allkeys-lru

# Persistence
save 900 1
save 300 10
save 60 10000

# Logging
loglevel notice
logfile /var/log/redis/redis-server.log
```

```bash
# Start and enable Redis
sudo systemctl start redis-server
sudo systemctl enable redis-server

# Test Redis
redis-cli ping
```

### Step 7: Install Additional Tools

```bash
# Install Composer
curl -sS https://getcomposer.org/installer | php
sudo mv composer.phar /usr/local/bin/composer
sudo chmod +x /usr/local/bin/composer

# Install Node.js (for Moodle development)
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt install -y nodejs

# Install additional tools
sudo apt install -y imagemagick ghostscript unoconv libreoffice
sudo apt install -y cron rsync tar gzip zip unzip
sudo apt install -y build-essential make cmake
```

### Step 8: Configure Nginx

```bash
# Create Nginx configuration for Moodle
sudo nano /etc/nginx/sites-available/moodle
```

**Nginx Configuration:**
```nginx
server {
    listen 80;
    server_name lms-server.local;
    root /var/www/moodle;
    index index.php index.html;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied expired no-cache no-store private must-revalidate auth;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss;

    # Main location
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    # PHP processing
    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php8.1-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }

    # Deny access to sensitive files
    location ~ /\. {
        deny all;
    }

    location ~ /(config|cache|local|moodledata|backup|temp|lang|pix|theme|userpix|upgrade|admin|lib|install|test|vendor)/ {
        deny all;
    }

    # Static files caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/moodle /etc/nginx/sites-enabled/
sudo rm /etc/nginx/sites-enabled/default

# Test Nginx configuration
sudo nginx -t

# Restart Nginx
sudo systemctl restart nginx
```

## ✅ Verification

### Service Status Check

```bash
# Check all services
sudo systemctl status nginx
sudo systemctl status php8.1-fpm
sudo systemctl status mariadb
sudo systemctl status redis-server

# Check PHP version and extensions
php -v
php -m | grep -E "(mysql|gd|curl|xml|mbstring|zip|intl|soap|ldap|imagick|redis)"

# Check MariaDB
sudo mysql -u root -p -e "SHOW DATABASES;"

# Check Redis
redis-cli ping

# Check Composer
composer --version

# Check Node.js
node --version
npm --version
```

### Expected Results

- ✅ Nginx running on port 80
- ✅ PHP 8.1-FPM running
- ✅ MariaDB running with moodle database
- ✅ Redis running on port 6379
- ✅ All PHP extensions installed
- ✅ Composer and Node.js installed

## 🚨 Troubleshooting

### Common Issues

**1. PHP-FPM not starting**
```bash
# Check PHP-FPM logs
sudo journalctl -u php8.1-fpm
sudo tail -f /var/log/php8.1-fpm.log
```

**2. MariaDB connection refused**
```bash
# Check MariaDB status
sudo systemctl status mariadb
sudo tail -f /var/log/mysql/error.log
```

**3. Nginx configuration error**
```bash
# Test configuration
sudo nginx -t
# Check syntax in configuration files
```

**4. Redis not responding**
```bash
# Check Redis logs
sudo tail -f /var/log/redis/redis-server.log
# Restart Redis
sudo systemctl restart redis-server
```

## 📝 Next Steps

Setelah software installation selesai, lanjutkan ke:
- [03-security-hardening.md](03-security-hardening.md) - Konfigurasi keamanan server

## 📚 References

- [Nginx Documentation](https://nginx.org/en/docs/)
- [PHP-FPM Configuration](https://www.php.net/manual/en/install.fpm.configuration.php)
- [MariaDB Documentation](https://mariadb.org/documentation/)
- [Redis Configuration](https://redis.io/topics/config)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
