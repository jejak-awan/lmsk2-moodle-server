# 🌐 Web Server Configuration for Moodle 3.11 LTS

## 📋 Overview

Dokumen ini menjelaskan konfigurasi web server (Nginx) untuk Moodle 3.11 LTS, termasuk virtual host, PHP-FPM integration, SSL configuration, dan optimasi performa.

## 🎯 Objectives

- [ ] Konfigurasi Nginx virtual host untuk Moodle
- [ ] Setup PHP-FPM integration
- [ ] Konfigurasi SSL/TLS certificates
- [ ] Optimasi performa dan caching
- [ ] Setup security headers
- [ ] Konfigurasi log management

## 📋 Prerequisites

- Nginx sudah terinstall
- PHP 8.1-FPM sudah terinstall
- MariaDB sudah dikonfigurasi
- SSL certificate sudah tersedia
- Domain name sudah dikonfigurasi

## 🔧 Step-by-Step Guide

### Step 1: Create Moodle Directory Structure

```bash
# Create Moodle directory structure
sudo mkdir -p /var/www/moodle
sudo mkdir -p /var/www/moodle/moodledata

# Set proper ownership
sudo chown -R www-data:www-data /var/www/moodle
sudo chown -R moodle:www-data /var/www/moodle/moodledata

# Set proper permissions
sudo chmod -R 755 /var/www/moodle
sudo chmod -R 777 /var/www/moodle/moodledata
```

### Step 2: Nginx Configuration for Moodle 3.11 LTS

```bash
# Create Nginx configuration for Moodle
sudo nano /etc/nginx/sites-available/moodle
```

**Nginx Configuration:**
```nginx
# Rate limiting zones
limit_req_zone $binary_remote_addr zone=login:10m rate=5r/m;
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=general:10m rate=30r/s;

# Upstream for PHP-FPM
upstream php-fpm {
    server unix:/var/run/php/php8.1-fpm.sock;
}

# HTTP to HTTPS redirect
server {
    listen 80;
    server_name lms.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

# Main HTTPS server block
server {
    listen 443 ssl http2;
    server_name lms.yourdomain.com;
    root /var/www/moodle;
    index index.php index.html;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/lms.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/lms.yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    ssl_stapling on;
    ssl_stapling_verify on;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'self';" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied expired no-cache no-store private must-revalidate auth;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/x-javascript
        application/xml+rss
        application/javascript
        application/json
        image/svg+xml;

    # Client settings
    client_max_body_size 200M;
    client_body_timeout 60s;
    client_header_timeout 60s;

    # Main location block
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    # Login rate limiting
    location ~ ^/login/ {
        limit_req zone=login burst=3 nodelay;
        try_files $uri $uri/ /index.php?$query_string;
    }

    # API rate limiting
    location ~ ^/webservice/ {
        limit_req zone=api burst=20 nodelay;
        try_files $uri $uri/ /index.php?$query_string;
    }

    # General rate limiting
    location ~ ^/ {
        limit_req zone=general burst=50 nodelay;
        try_files $uri $uri/ /index.php?$query_string;
    }

    # PHP processing
    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass php-fpm;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
        
        # FastCGI settings
        fastcgi_connect_timeout 60s;
        fastcgi_send_timeout 60s;
        fastcgi_read_timeout 60s;
        fastcgi_buffer_size 128k;
        fastcgi_buffers 4 256k;
        fastcgi_busy_buffers_size 256k;
        fastcgi_temp_file_write_size 256k;
        
        # Security headers for PHP
        fastcgi_param HTTPS on;
        fastcgi_param HTTP_SCHEME https;
    }

    # Deny access to sensitive files
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }

    location ~ /(config|cache|local|moodledata|backup|temp|lang|pix|theme|userpix|upgrade|admin|lib|install|test|vendor)/ {
        deny all;
        access_log off;
        log_not_found off;
    }

    # Block common attack patterns
    location ~* /(wp-admin|wp-login|xmlrpc|admin|administrator) {
        deny all;
        access_log off;
        log_not_found off;
    }

    # Static files caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        add_header Vary "Accept-Encoding";
        access_log off;
    }

    # Moodle specific files
    location ~* \.(pdf|doc|docx|xls|xlsx|ppt|pptx|zip|rar|7z)$ {
        expires 30d;
        add_header Cache-Control "public";
        access_log off;
    }

    # Logging
    access_log /var/log/nginx/moodle-access.log;
    error_log /var/log/nginx/moodle-error.log;
}
```

### Step 3: Enable Nginx Site

```bash
# Enable Moodle site
sudo ln -s /etc/nginx/sites-available/moodle /etc/nginx/sites-enabled/

# Remove default site
sudo rm /etc/nginx/sites-enabled/default

# Test Nginx configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

### Step 4: PHP-FPM Configuration for Moodle 3.11 LTS

```bash
# Create PHP-FPM pool configuration for Moodle
sudo nano /etc/php/8.1/fpm/pool.d/moodle.conf
```

**PHP-FPM Pool Configuration:**
```ini
[moodle]
user = www-data
group = www-data
listen = /var/run/php/php8.1-fpm-moodle.sock
listen.owner = www-data
listen.group = www-data
listen.mode = 0660

pm = dynamic
pm.max_children = 50
pm.start_servers = 5
pm.min_spare_servers = 5
pm.max_spare_servers = 35
pm.max_requests = 1000

pm.status_path = /fpm-status
ping.path = /fpm-ping

; Environment variables
env[HOSTNAME] = $HOSTNAME
env[PATH] = /usr/local/bin:/usr/bin:/bin
env[TMP] = /tmp
env[TMPDIR] = /tmp
env[TEMP] = /tmp

; PHP settings
php_admin_value[error_log] = /var/log/php8.1-fpm-moodle.log
php_admin_flag[log_errors] = on
php_admin_value[memory_limit] = 512M
php_admin_value[max_execution_time] = 300
php_admin_value[max_input_time] = 300
php_admin_value[upload_max_filesize] = 200M
php_admin_value[post_max_size] = 200M
php_admin_value[max_file_uploads] = 20

; Security settings
php_admin_value[disable_functions] = exec,passthru,shell_exec,system,proc_open,popen
php_admin_value[open_basedir] = /var/www/moodle:/tmp:/var/tmp
```

### Step 5: Update Nginx Configuration for Custom PHP-FPM Pool

```bash
# Update Nginx configuration to use custom PHP-FPM pool
sudo nano /etc/nginx/sites-available/moodle
```

**Update the upstream section:**
```nginx
# Upstream for PHP-FPM
upstream php-fpm {
    server unix:/var/run/php/php8.1-fpm-moodle.sock;
}
```

```bash
# Restart PHP-FPM
sudo systemctl restart php8.1-fpm

# Test Nginx configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

### Step 6: SSL Certificate Configuration

```bash
# Install Certbot for Nginx
sudo apt install -y certbot python3-certbot-nginx

# Generate SSL certificate
sudo certbot --nginx -d lms.yourdomain.com

# Test certificate renewal
sudo certbot renew --dry-run

# Setup automatic renewal
sudo crontab -e
```

**Add to crontab:**
```
# SSL certificate renewal
0 12 * * * /usr/bin/certbot renew --quiet
```

### Step 7: Log Management Configuration

```bash
# Create log rotation for Moodle
sudo nano /etc/logrotate.d/moodle
```

**Log Rotation Configuration:**
```
/var/log/nginx/moodle-*.log {
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

/var/log/php8.1-fpm-moodle.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 www-data adm
    postrotate
        systemctl reload php8.1-fpm
    endscript
}
```

### Step 8: Performance Monitoring

```bash
# Create Nginx monitoring script
sudo nano /usr/local/bin/nginx-monitor.sh
```

**Nginx Monitoring Script:**
```bash
#!/bin/bash

# Nginx Monitoring Script for Moodle
LOG_FILE="/var/log/nginx-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Nginx monitoring..." >> $LOG_FILE

# Check Nginx status
if systemctl is-active --quiet nginx; then
    echo "[$DATE] ✓ Nginx is running" >> $LOG_FILE
else
    echo "[$DATE] ✗ Nginx is not running" >> $LOG_FILE
    systemctl restart nginx
fi

# Check PHP-FPM status
if systemctl is-active --quiet php8.1-fpm; then
    echo "[$DATE] ✓ PHP-FPM is running" >> $LOG_FILE
else
    echo "[$DATE] ✗ PHP-FPM is not running" >> $LOG_FILE
    systemctl restart php8.1-fpm
fi

# Check active connections
ACTIVE_CONNECTIONS=$(netstat -an | grep :443 | grep ESTABLISHED | wc -l)
echo "[$DATE] Active HTTPS connections: $ACTIVE_CONNECTIONS" >> $LOG_FILE

# Check error rate
ERROR_COUNT=$(tail -100 /var/log/nginx/moodle-error.log | grep -c "$(date '+%Y/%m/%d')")
echo "[$DATE] Error count today: $ERROR_COUNT" >> $LOG_FILE

# Check response time
RESPONSE_TIME=$(curl -o /dev/null -s -w '%{time_total}' https://lms.yourdomain.com/)
echo "[$DATE] Response time: ${RESPONSE_TIME}s" >> $LOG_FILE

echo "[$DATE] Nginx monitoring completed" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/nginx-monitor.sh

# Add to crontab (every 5 minutes)
sudo crontab -e
```

**Add to crontab:**
```
# Nginx monitoring every 5 minutes
*/5 * * * * /usr/local/bin/nginx-monitor.sh
```

## ✅ Verification

### Web Server Test

```bash
# Test Nginx configuration
sudo nginx -t

# Check Nginx status
sudo systemctl status nginx

# Check PHP-FPM status
sudo systemctl status php8.1-fpm

# Test HTTP to HTTPS redirect
curl -I http://lms.yourdomain.com

# Test HTTPS connection
curl -I https://lms.yourdomain.com

# Test PHP processing
echo "<?php phpinfo(); ?>" | sudo tee /var/www/moodle/info.php
curl https://lms.yourdomain.com/info.php
sudo rm /var/www/moodle/info.php
```

### SSL Certificate Test

```bash
# Check SSL certificate
openssl s_client -connect lms.yourdomain.com:443 -servername lms.yourdomain.com

# Test SSL configuration
curl -I https://lms.yourdomain.com

# Check certificate expiration
sudo certbot certificates
```

### Expected Results

- ✅ Nginx running and accessible
- ✅ PHP-FPM processing PHP files
- ✅ SSL certificate installed and working
- ✅ HTTP to HTTPS redirect working
- ✅ Security headers implemented
- ✅ Rate limiting configured
- ✅ Log management active
- ✅ Performance monitoring running

## 🚨 Troubleshooting

### Common Issues

**1. Nginx configuration error**
```bash
# Test configuration
sudo nginx -t

# Check syntax
sudo nginx -T

# Check error logs
sudo tail -f /var/log/nginx/error.log
```

**2. PHP-FPM not processing**
```bash
# Check PHP-FPM status
sudo systemctl status php8.1-fpm

# Check PHP-FPM logs
sudo tail -f /var/log/php8.1-fpm.log

# Test PHP-FPM socket
sudo ls -la /var/run/php/php8.1-fpm-moodle.sock
```

**3. SSL certificate issues**
```bash
# Check certificate status
sudo certbot certificates

# Renew certificate
sudo certbot renew --force-renewal

# Check certificate files
sudo ls -la /etc/letsencrypt/live/lms.yourdomain.com/
```

**4. Permission issues**
```bash
# Fix ownership
sudo chown -R www-data:www-data /var/www/moodle

# Fix permissions
sudo chmod -R 755 /var/www/moodle
sudo chmod -R 777 /var/www/moodle/moodledata
```

## 📝 Next Steps

Setelah web server configuration selesai, lanjutkan ke:
- [04-moodle-installation.md](04-moodle-installation.md) - Install Moodle 3.11 LTS

## 📚 References

- [Nginx Documentation](https://nginx.org/en/docs/)
- [PHP-FPM Configuration](https://www.php.net/manual/en/install.fpm.configuration.php)
- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
- [Moodle Web Server Requirements](https://docs.moodle.org/311/en/Web_server_requirements)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
