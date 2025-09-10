# 🔒 SSL Certificate Setup for Moodle

## 📋 Overview

Dokumen ini menjelaskan setup SSL certificate untuk Moodle production environment, termasuk Let's Encrypt, certificate management, dan security configuration.

## 🎯 Objectives

- [ ] Setup Let's Encrypt SSL certificate
- [ ] Konfigurasi certificate auto-renewal
- [ ] Setup security headers
- [ ] Konfigurasi HSTS dan security policies
- [ ] Verifikasi SSL configuration

## 🔧 Step-by-Step Guide

### Step 1: Install Certbot

```bash
# Update package list
sudo apt update

# Install Certbot and Nginx plugin
sudo apt install -y certbot python3-certbot-nginx

# Verify installation
certbot --version
```

### Step 2: Obtain SSL Certificate

```bash
# Obtain SSL certificate for your domain
sudo certbot --nginx -d lms.yourdomain.com

# Follow the prompts:
# 1. Enter email address for renewal notifications
# 2. Agree to terms of service
# 3. Choose whether to share email with EFF
# 4. Select redirect option (recommended: redirect HTTP to HTTPS)
```

### Step 3: Verify Certificate Installation

```bash
# Check certificate status
sudo certbot certificates

# Test certificate
openssl s_client -connect lms.yourdomain.com:443 -servername lms.yourdomain.com

# Check certificate details
echo | openssl s_client -servername lms.yourdomain.com -connect lms.yourdomain.com:443 2>/dev/null | openssl x509 -noout -dates
```

### Step 4: Configure Auto-Renewal

```bash
# Test auto-renewal
sudo certbot renew --dry-run

# Add renewal to crontab
sudo crontab -e
```

**Add to crontab:**
```
# SSL certificate renewal check twice daily
0 12 * * * /usr/bin/certbot renew --quiet
0 0 * * * /usr/bin/certbot renew --quiet
```

### Step 5: Configure Security Headers

```bash
# Update Nginx configuration
sudo nano /etc/nginx/sites-available/moodle
```

**Enhanced Nginx Configuration with Security Headers:**
```nginx
# Moodle Production Nginx Configuration with SSL
server {
    listen 80;
    server_name lms.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name lms.yourdomain.com;
    root /var/www/moodle;
    index index.php index.html index.htm;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/lms.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/lms.yourdomain.com/privkey.pem;
    ssl_trusted_certificate /etc/letsencrypt/live/lms.yourdomain.com/chain.pem;

    # SSL Security Settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    ssl_session_tickets off;
    ssl_stapling on;
    ssl_stapling_verify on;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'; frame-ancestors 'self';" always;
    add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;

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
        fastcgi_pass unix:/var/run/php/php8.1-fpm-moodle.sock;
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

### Step 6: Configure HSTS

```bash
# Create HSTS configuration
sudo nano /etc/nginx/conf.d/hsts.conf
```

**HSTS Configuration:**
```nginx
# HSTS Configuration
map $scheme $hsts_header {
    https "max-age=31536000; includeSubDomains; preload";
}

server {
    listen 443 ssl http2;
    server_name lms.yourdomain.com;
    
    # HSTS Header
    add_header Strict-Transport-Security $hsts_header always;
}
```

### Step 7: Configure Certificate Monitoring

```bash
# Create certificate monitoring script
sudo nano /usr/local/bin/ssl-monitor.sh
```

**SSL Monitoring Script:**
```bash
#!/bin/bash

# SSL Certificate Monitoring Script
LOG_FILE="/var/log/ssl-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')
DOMAIN="lms.yourdomain.com"

echo "[$DATE] SSL certificate monitoring..." >> $LOG_FILE

# Check certificate expiration
CERT_EXPIRY=$(echo | openssl s_client -servername $DOMAIN -connect $DOMAIN:443 2>/dev/null | openssl x509 -noout -dates | grep notAfter | cut -d= -f2)
CERT_EXPIRY_EPOCH=$(date -d "$CERT_EXPIRY" +%s)
CURRENT_EPOCH=$(date +%s)
DAYS_UNTIL_EXPIRY=$(( (CERT_EXPIRY_EPOCH - CURRENT_EPOCH) / 86400 ))

echo "Certificate expires in: $DAYS_UNTIL_EXPIRY days" >> $LOG_FILE

# Check certificate validity
if openssl s_client -connect $DOMAIN:443 -servername $DOMAIN < /dev/null 2>/dev/null | grep -q "Verify return code: 0"; then
    echo "Certificate is valid" >> $LOG_FILE
else
    echo "Certificate validation failed" >> $LOG_FILE
fi

# Check SSL grade
SSL_GRADE=$(curl -s "https://api.ssllabs.com/api/v3/analyze?host=$DOMAIN" | grep -o '"grade":"[^"]*"' | cut -d'"' -f4)
echo "SSL Grade: $SSL_GRADE" >> $LOG_FILE

# Alert if certificate expires in less than 30 days
if [ $DAYS_UNTIL_EXPIRY -lt 30 ]; then
    echo "WARNING: Certificate expires in $DAYS_UNTIL_EXPIRY days" >> $LOG_FILE
    # Send email notification (configure as needed)
    # echo "SSL Certificate expires in $DAYS_UNTIL_EXPIRY days" | mail -s "SSL Certificate Expiry Warning" admin@yourdomain.com
fi

echo "---" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/ssl-monitor.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# SSL monitoring daily
0 9 * * * /usr/local/bin/ssl-monitor.sh
```

### Step 8: Configure SSL Security

```bash
# Create SSL security configuration
sudo nano /etc/nginx/conf.d/ssl-security.conf
```

**SSL Security Configuration:**
```nginx
# SSL Security Configuration
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-SHA384;
ssl_prefer_server_ciphers off;
ssl_session_cache shared:SSL:10m;
ssl_session_timeout 10m;
ssl_session_tickets off;
ssl_stapling on;
ssl_stapling_verify on;
ssl_stapling_file /etc/ssl/certs/stapling-ocsp.der;

# OCSP Stapling
ssl_trusted_certificate /etc/letsencrypt/live/lms.yourdomain.com/chain.pem;
resolver 8.8.8.8 8.8.4.4 valid=300s;
resolver_timeout 5s;
```

### Step 9: Test and Reload Configuration

```bash
# Test Nginx configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx

# Test SSL configuration
curl -I https://lms.yourdomain.com
```

## ✅ Verification

### SSL Configuration Test

```bash
# Test SSL certificate
openssl s_client -connect lms.yourdomain.com:443 -servername lms.yourdomain.com

# Test SSL grade
curl -s "https://api.ssllabs.com/api/v3/analyze?host=lms.yourdomain.com"

# Test security headers
curl -I https://lms.yourdomain.com | grep -E "(Strict-Transport-Security|X-Frame-Options|X-XSS-Protection|X-Content-Type-Options|Referrer-Policy|Content-Security-Policy)"

# Test certificate renewal
sudo certbot renew --dry-run
```

### Expected Results

- ✅ SSL certificate installed and valid
- ✅ Auto-renewal configured
- ✅ Security headers present
- ✅ HSTS configured
- ✅ SSL monitoring active

## 🚨 Troubleshooting

### Common Issues

**1. Certificate not working**
```bash
# Check certificate status
sudo certbot certificates

# Check Nginx configuration
sudo nginx -t

# Check certificate files
ls -la /etc/letsencrypt/live/lms.yourdomain.com/
```

**2. Auto-renewal failing**
```bash
# Test renewal manually
sudo certbot renew --dry-run

# Check renewal logs
sudo tail -f /var/log/letsencrypt/letsencrypt.log
```

**3. Security headers not working**
```bash
# Check Nginx configuration
sudo nginx -t

# Test headers
curl -I https://lms.yourdomain.com
```

## 📝 Next Steps

Setelah SSL certificate setup selesai, lanjutkan ke:
- [03-load-balancing.md](03-load-balancing.md) - Setup load balancing

## 📚 References

- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
- [Certbot Documentation](https://certbot.eff.org/docs/)
- [SSL Labs Test](https://www.ssllabs.com/ssltest/)
- [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
