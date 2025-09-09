# 🔒 Security Hardening

## 📋 Overview

Dokumen ini menjelaskan langkah-langkah untuk mengamankan server LMS Moodle dari berbagai ancaman keamanan, termasuk konfigurasi firewall, SSL/TLS, dan tools keamanan lainnya.

## 🎯 Objectives

- [ ] Konfigurasi firewall dan network security
- [ ] Setup SSL/TLS certificates
- [ ] Install dan konfigurasi Fail2ban
- [ ] Konfigurasi file permissions
- [ ] Setup log monitoring dan intrusion detection
- [ ] Implementasi security headers

## 📋 Prerequisites

- Server Ubuntu 22.04 LTS sudah dikonfigurasi
- Nginx, PHP, MariaDB, Redis sudah terinstall
- Firewall dasar sudah dikonfigurasi
- Domain name sudah tersedia (untuk SSL)

## 🔧 Step-by-Step Guide

### Step 1: Advanced Firewall Configuration

```bash
# Install additional security tools
sudo apt install -y fail2ban ufw iptables-persistent

# Configure UFW with more restrictive rules
sudo ufw --force reset
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow essential services
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Allow specific IP ranges (adjust as needed)
sudo ufw allow from 192.168.88.0/24 to any port 22
sudo ufw allow from 192.168.88.0/24 to any port 80
sudo ufw allow from 192.168.88.0/24 to any port 443

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status verbose
```

### Step 2: Install and Configure Fail2ban

```bash
# Create Fail2ban configuration
sudo nano /etc/fail2ban/jail.local
```

**Fail2ban Configuration:**
```ini
[DEFAULT]
# Ban hosts for 1 hour
bantime = 3600
# Override /etc/fail2ban/jail.d/00-firewalld.conf
banaction = ufw
# Number of failures before ban
maxretry = 3
# Time window for failures
findtime = 600
# Ignore local IPs
ignoreip = 127.0.0.1/8 ::1 192.168.88.0/24

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600

[nginx-http-auth]
enabled = true
filter = nginx-http-auth
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 3

[nginx-limit-req]
enabled = true
filter = nginx-limit-req
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 3

[php-url-fopen]
enabled = true
filter = php-url-fopen
port = http,https
logpath = /var/log/nginx/access.log
maxretry = 3
```

```bash
# Create custom filters
sudo nano /etc/fail2ban/filter.d/nginx-http-auth.conf
```

**Nginx HTTP Auth Filter:**
```ini
[Definition]
failregex = ^<HOST> -.*"(GET|POST).*HTTP.*" (401|403) .*$
ignoreregex =
```

```bash
# Create PHP URL fopen filter
sudo nano /etc/fail2ban/filter.d/php-url-fopen.conf
```

**PHP URL fopen Filter:**
```ini
[Definition]
failregex = ^<HOST> -.*"(GET|POST).*\.php.*" (200|404) .*$
ignoreregex =
```

```bash
# Start and enable Fail2ban
sudo systemctl start fail2ban
sudo systemctl enable fail2ban

# Check status
sudo fail2ban-client status
```

### Step 3: SSL/TLS Certificate Setup

```bash
# Install Certbot
sudo apt install -y certbot python3-certbot-nginx

# Generate SSL certificate (replace with your domain)
sudo certbot --nginx -d lms.yourdomain.com

# Test certificate renewal
sudo certbot renew --dry-run

# Setup automatic renewal
sudo crontab -e
```

**Add to crontab:**
```
0 12 * * * /usr/bin/certbot renew --quiet
```

### Step 4: Enhanced Nginx Security Configuration

```bash
# Update Nginx configuration with security headers
sudo nano /etc/nginx/sites-available/moodle
```

**Enhanced Nginx Configuration:**
```nginx
server {
    listen 80;
    server_name lms.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

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

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'self';" always;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=login:10m rate=5r/m;
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

    # Main location
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

    # PHP processing
    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php8.1-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
        
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
    }
}
```

### Step 5: File Permissions and Ownership

```bash
# Set proper ownership
sudo chown -R www-data:www-data /var/www/moodle
sudo chown -R moodle:www-data /var/www/moodle/moodledata

# Set proper permissions
sudo find /var/www/moodle -type d -exec chmod 755 {} \;
sudo find /var/www/moodle -type f -exec chmod 644 {} \;

# Make moodledata writable
sudo chmod -R 777 /var/www/moodle/moodledata

# Secure config.php (after Moodle installation)
sudo chmod 600 /var/www/moodle/config.php
sudo chown root:www-data /var/www/moodle/config.php
```

### Step 6: MariaDB Security Configuration

```bash
# Create MariaDB security configuration
sudo nano /etc/mysql/mariadb.conf.d/99-security.cnf
```

**MariaDB Security Configuration:**
```ini
[mysqld]
# Security settings
local-infile = 0
symbolic-links = 0
skip-networking = 0
bind-address = 127.0.0.1

# Logging
log_error = /var/log/mysql/error.log
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2

# Connection limits
max_connections = 100
max_connect_errors = 10
connect_timeout = 10
wait_timeout = 600
interactive_timeout = 600

# Disable dangerous functions
sql_mode = STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO
```

```bash
# Restart MariaDB
sudo systemctl restart mariadb

# Remove test database and users
sudo mysql -u root -p
```

**MariaDB Security Commands:**
```sql
-- Remove test database
DROP DATABASE IF EXISTS test;
DELETE FROM mysql.db WHERE Db='test' OR Db='test\\_%';

-- Remove anonymous users
DELETE FROM mysql.user WHERE User='';
DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');

-- Create dedicated user for backups
CREATE USER 'backup'@'localhost' IDENTIFIED BY 'strong_backup_password';
GRANT SELECT, LOCK TABLES, SHOW VIEW, EVENT, TRIGGER ON *.* TO 'backup'@'localhost';

-- Flush privileges
FLUSH PRIVILEGES;
EXIT;
```

### Step 7: PHP Security Configuration

```bash
# Create PHP security configuration
sudo nano /etc/php/8.1/fpm/conf.d/99-security.ini
```

**PHP Security Configuration:**
```ini
; Security settings
expose_php = Off
allow_url_fopen = Off
allow_url_include = Off
disable_functions = exec,passthru,shell_exec,system,proc_open,popen,curl_exec,curl_multi_exec,parse_ini_file,show_source
disable_classes = 

; File upload security
file_uploads = On
upload_max_filesize = 200M
post_max_size = 200M
max_file_uploads = 20

; Session security
session.cookie_httponly = 1
session.cookie_secure = 1
session.use_strict_mode = 1
session.cookie_samesite = "Strict"

; Error handling
display_errors = Off
display_startup_errors = Off
log_errors = On
error_log = /var/log/php_errors.log
```

```bash
# Restart PHP-FPM
sudo systemctl restart php8.1-fpm
```

### Step 8: Log Monitoring Setup

```bash
# Install log monitoring tools
sudo apt install -y logwatch rsyslog

# Configure log rotation
sudo nano /etc/logrotate.d/moodle
```

**Log Rotation Configuration:**
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

/var/log/php_errors.log {
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

```bash
# Setup log monitoring script
sudo nano /usr/local/bin/security-monitor.sh
```

**Security Monitor Script:**
```bash
#!/bin/bash

# Security monitoring script
LOG_FILE="/var/log/security-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Starting security check..." >> $LOG_FILE

# Check for failed login attempts
FAILED_LOGINS=$(grep "Failed password" /var/log/auth.log | wc -l)
if [ $FAILED_LOGINS -gt 10 ]; then
    echo "[$DATE] WARNING: High number of failed login attempts: $FAILED_LOGINS" >> $LOG_FILE
fi

# Check for suspicious PHP errors
PHP_ERRORS=$(grep "PHP" /var/log/nginx/error.log | wc -l)
if [ $PHP_ERRORS -gt 50 ]; then
    echo "[$DATE] WARNING: High number of PHP errors: $PHP_ERRORS" >> $LOG_FILE
fi

# Check disk space
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    echo "[$DATE] WARNING: Disk usage high: $DISK_USAGE%" >> $LOG_FILE
fi

echo "[$DATE] Security check completed." >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/security-monitor.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# Security monitoring every hour
0 * * * * /usr/local/bin/security-monitor.sh
```

## ✅ Verification

### Security Check

```bash
# Check firewall status
sudo ufw status verbose

# Check Fail2ban status
sudo fail2ban-client status

# Check SSL certificate
sudo certbot certificates

# Check file permissions
ls -la /var/www/moodle/
ls -la /var/www/moodle/config.php

# Check MariaDB security
sudo mysql -u root -p -e "SELECT user, host FROM mysql.user;"

# Check PHP security
php -i | grep -E "(expose_php|allow_url_fopen|disable_functions)"

# Check log files
sudo tail -f /var/log/fail2ban.log
sudo tail -f /var/log/security-monitor.log
```

### Expected Results

- ✅ Firewall configured with restrictive rules
- ✅ Fail2ban protecting against brute force attacks
- ✅ SSL certificate installed and working
- ✅ Security headers implemented
- ✅ File permissions properly set
- ✅ MariaDB secured with strong passwords
- ✅ PHP security settings applied
- ✅ Log monitoring active

## 🚨 Troubleshooting

### Common Issues

**1. SSL certificate not working**
```bash
# Check certificate status
sudo certbot certificates
# Renew certificate
sudo certbot renew --force-renewal
```

**2. Fail2ban not banning IPs**
```bash
# Check Fail2ban logs
sudo tail -f /var/log/fail2ban.log
# Test jail status
sudo fail2ban-client status sshd
```

**3. File permission errors**
```bash
# Fix ownership
sudo chown -R www-data:www-data /var/www/moodle
# Fix permissions
sudo find /var/www/moodle -type d -exec chmod 755 {} \;
```

## 📝 Next Steps

Setelah security hardening selesai, lanjutkan ke:
- [04-basic-configuration.md](04-basic-configuration.md) - Konfigurasi dasar sistem

## 📚 References

- [Ubuntu Security Guide](https://ubuntu.com/security)
- [Fail2ban Documentation](https://www.fail2ban.org/wiki/index.php/Main_Page)
- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
- [OWASP Security Headers](https://owasp.org/www-project-secure-headers/)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
