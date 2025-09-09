# 🌐 Web Server Configuration for Moodle 5.0

## 📋 Overview

Dokumen ini menjelaskan konfigurasi web server untuk Moodle 5.0, termasuk setup Nginx, PHP-FPM, SSL/TLS, dan optimasi performa.

## 🎯 Objectives

- [ ] Install dan konfigurasi Nginx
- [ ] Setup PHP-FPM untuk Moodle 5.0
- [ ] Konfigurasi virtual host
- [ ] Setup SSL/TLS certificates
- [ ] Optimasi performa dan security

## 🔧 Step-by-Step Guide

### Step 1: Install Nginx

```bash
# Update package list
sudo apt update

# Install Nginx
sudo apt install -y nginx

# Start and enable Nginx
sudo systemctl start nginx
sudo systemctl enable nginx

# Check status
sudo systemctl status nginx
```

### Step 2: Install PHP-FPM

```bash
# Install PHP 8.2 and extensions
sudo apt install -y php8.2-fpm php8.2-cli php8.2-common \
    php8.2-mysql php8.2-zip php8.2-gd php8.2-mbstring \
    php8.2-curl php8.2-xml php8.2-bcmath php8.2-intl \
    php8.2-soap php8.2-ldap php8.2-imagick php8.2-redis \
    php8.2-openssl php8.2-json php8.2-dom php8.2-fileinfo \
    php8.2-iconv php8.2-simplexml php8.2-tokenizer \
    php8.2-xmlreader php8.2-xmlwriter php8.2-exif \
    php8.2-ftp php8.2-gettext php8.2-sodium php8.2-hash \
    php8.2-filter

# Start and enable PHP-FPM
sudo systemctl start php8.2-fpm
sudo systemctl enable php8.2-fpm

# Check status
sudo systemctl status php8.2-fpm
```

### Step 3: Configure PHP-FPM

```bash
# Configure PHP-FPM for Moodle 5.0
sudo nano /etc/php/8.2/fpm/php.ini
```

**PHP Configuration:**
```ini
; PHP Configuration for Moodle 5.0

; Memory and execution time
memory_limit = 1G
max_execution_time = 900
max_input_time = 900
max_input_vars = 5000

; File uploads
upload_max_filesize = 500M
post_max_size = 500M
max_file_uploads = 20

; Session configuration
session.gc_maxlifetime = 1440
session.cookie_httponly = 1
session.use_strict_mode = 1

; OPcache configuration
opcache.enable = 1
opcache.memory_consumption = 512
opcache.interned_strings_buffer = 16
opcache.max_accelerated_files = 20000
opcache.revalidate_freq = 2
opcache.fast_shutdown = 1
opcache.validate_timestamps = 0

; Error reporting
display_errors = Off
log_errors = On
error_log = /var/log/php_errors.log
error_reporting = E_ALL & ~E_DEPRECATED & ~E_STRICT

; Security
allow_url_fopen = Off
allow_url_include = Off
expose_php = Off
```

### Step 4: Configure PHP-FPM Pool

```bash
# Configure PHP-FPM pool for Moodle
sudo nano /etc/php/8.2/fpm/pool.d/moodle.conf
```

**PHP-FPM Pool Configuration:**
```ini
[moodle]
user = www-data
group = www-data
listen = /var/run/php/php8.2-fpm-moodle.sock
listen.owner = www-data
listen.group = www-data
listen.mode = 0660

pm = dynamic
pm.max_children = 100
pm.start_servers = 10
pm.min_spare_servers = 10
pm.max_spare_servers = 70
pm.max_requests = 1000

; Logging
access.log = /var/log/php8.2-fpm-moodle-access.log
slowlog = /var/log/php8.2-fpm-moodle-slow.log
request_slowlog_timeout = 10s

; Environment variables
env[HOSTNAME] = $HOSTNAME
env[PATH] = /usr/local/bin:/usr/bin:/bin
env[TMP] = /tmp
env[TMPDIR] = /tmp
env[TEMP] = /tmp
```

### Step 5: Create Nginx Virtual Host

```bash
# Create Nginx virtual host for Moodle
sudo nano /etc/nginx/sites-available/moodle
```

**Nginx Virtual Host Configuration:**
```nginx
# Moodle 5.0 Nginx Configuration
server {
    listen 80;
    server_name lms.yourdomain.com;
    root /var/www/moodle;
    index index.php index.html index.htm;

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
    gzip_proxied any;
    gzip_comp_level 6;
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

    # Main location block
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    # PHP processing
    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php8.2-fpm-moodle.sock;
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
    }

    # Deny access to sensitive files
    location ~ /\. {
        deny all;
    }

    location ~ /(config|install|lib|lang|pix|theme|vendor)/.*\.php$ {
        deny all;
    }

    # Static files caching
    location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # Moodle specific locations
    location /dataroot/ {
        deny all;
    }

    location /admin/ {
        allow 192.168.1.0/24;
        deny all;
    }
}
```

### Step 6: Enable Site

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

### Step 7: Setup SSL/TLS Certificate

```bash
# Install Certbot
sudo apt install -y certbot python3-certbot-nginx

# Obtain SSL certificate
sudo certbot --nginx -d lms.yourdomain.com

# Test certificate renewal
sudo certbot renew --dry-run
```

### Step 8: Configure Firewall

```bash
# Configure UFW firewall
sudo ufw allow 'Nginx Full'
sudo ufw allow ssh
sudo ufw enable

# Check status
sudo ufw status
```

### Step 9: Create Moodle Directory

```bash
# Create Moodle directory
sudo mkdir -p /var/www/moodle

# Set permissions
sudo chown -R www-data:www-data /var/www/moodle
sudo chmod -R 755 /var/www/moodle
```

## ✅ Verification

### Web Server Test

```bash
# Test Nginx configuration
sudo nginx -t

# Test PHP-FPM
sudo systemctl status php8.2-fpm

# Test Nginx
sudo systemctl status nginx

# Test SSL certificate
openssl s_client -connect lms.yourdomain.com:443 -servername lms.yourdomain.com
```

### Expected Results

- ✅ Nginx running and accessible
- ✅ PHP-FPM running and accessible
- ✅ Virtual host configured
- ✅ SSL certificate installed
- ✅ Firewall configured
- ✅ Directory permissions set

## 🚨 Troubleshooting

### Common Issues

**1. Nginx configuration error**
```bash
# Check configuration
sudo nginx -t

# Check error logs
sudo tail -f /var/log/nginx/error.log
```

**2. PHP-FPM not working**
```bash
# Check PHP-FPM status
sudo systemctl status php8.2-fpm

# Check PHP-FPM logs
sudo tail -f /var/log/php8.2-fpm.log
```

**3. SSL certificate issues**
```bash
# Check certificate
sudo certbot certificates

# Renew certificate
sudo certbot renew
```

## 📝 Next Steps

Setelah web server configuration selesai, lanjutkan ke:
- [04-moodle-installation.md](04-moodle-installation.md) - Install Moodle 5.0

## 📚 References

- [Nginx Documentation](https://nginx.org/en/docs/)
- [PHP-FPM Documentation](https://www.php.net/manual/en/install.fpm.php)
- [Certbot Documentation](https://certbot.eff.org/docs/)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
