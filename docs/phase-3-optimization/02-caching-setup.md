# 🚀 Caching Setup for Moodle 3.11 LTS

## 📋 Overview

Dokumen ini menjelaskan setup caching system untuk Moodle 3.11 LTS, termasuk Redis, OPcache, dan Nginx caching untuk performa optimal.

## 🎯 Objectives

- [ ] Setup Redis untuk session dan data caching
- [ ] Konfigurasi OPcache untuk PHP bytecode caching
- [ ] Setup Nginx FastCGI caching
- [ ] Konfigurasi Moodle caching
- [ ] Monitoring cache performance

## 🔧 Step-by-Step Guide

### Step 1: Redis Caching Configuration

```bash
# Configure Redis for Moodle caching
sudo nano /etc/redis/redis.conf
```

**Redis Configuration:**
```
# Redis Caching Configuration
bind 127.0.0.1
port 6379
timeout 0
tcp-keepalive 300

# Memory management
maxmemory 512mb
maxmemory-policy allkeys-lru
maxmemory-samples 5

# Persistence
save 900 1
save 300 10
save 60 10000

# Performance
tcp-backlog 511
databases 16
```

```bash
# Restart Redis
sudo systemctl restart redis-server

# Test Redis
redis-cli ping
```

### Step 2: OPcache Configuration

```bash
# Configure OPcache
sudo nano /etc/php/8.1/fpm/conf.d/99-opcache.ini
```

**OPcache Configuration:**
```ini
; OPcache Configuration for Moodle
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

```bash
# Restart PHP-FPM
sudo systemctl restart php8.1-fpm
```

### Step 3: Nginx FastCGI Caching

```bash
# Configure Nginx caching
sudo nano /etc/nginx/conf.d/caching.conf
```

**Nginx Caching Configuration:**
```nginx
# FastCGI cache configuration
fastcgi_cache_path /var/cache/nginx/fastcgi levels=1:2 keys_zone=moodle:100m inactive=60m;
fastcgi_cache_key "$scheme$request_method$host$request_uri";

# Cache settings
fastcgi_cache_use_stale error timeout invalid_header http_500;
fastcgi_ignore_headers Cache-Control Expires Set-Cookie;
```

```bash
# Update Moodle site configuration
sudo nano /etc/nginx/sites-available/moodle
```

**Add to PHP location block:**
```nginx
location ~ \.php$ {
    # ... existing configuration ...
    
    # FastCGI cache
    fastcgi_cache moodle;
    fastcgi_cache_valid 200 60m;
    fastcgi_cache_valid 404 1m;
    fastcgi_cache_bypass $skip_cache;
    fastcgi_no_cache $skip_cache;
}

# Skip cache for admin and login
set $skip_cache 0;
if ($request_uri ~* "/admin/") {
    set $skip_cache 1;
}
if ($request_uri ~* "/login/") {
    set $skip_cache 1;
}
if ($request_method = POST) {
    set $skip_cache 1;
}
```

### Step 4: Moodle Caching Configuration

```bash
# Configure Moodle caching
cd /var/www/moodle
sudo -u www-data php admin/cli/cfg.php --name=cachejs --set=1
sudo -u www-data php admin/cli/cfg.php --name=cachetemplates --set=1
sudo -u www-data php admin/cli/cfg.php --name=session_handler_class --set="\core\session\redis"
sudo -u www-data php admin/cli/cfg.php --name=session_redis_host --set="127.0.0.1"
sudo -u www-data php admin/cli/cfg.php --name=session_redis_port --set=6379
```

### Step 5: Cache Monitoring

```bash
# Create cache monitoring script
sudo nano /usr/local/bin/cache-monitor.sh
```

**Cache Monitoring Script:**
```bash
#!/bin/bash

# Cache Monitoring Script
LOG_FILE="/var/log/cache-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Cache monitoring..." >> $LOG_FILE

# Redis cache info
REDIS_INFO=$(redis-cli info memory | grep used_memory_human)
echo "Redis Memory: $REDIS_INFO" >> $LOG_FILE

# OPcache info
OPCACHE_INFO=$(php -r "echo 'OPcache Hits: ' . opcache_get_status()['opcache_statistics']['hits'] . PHP_EOL;")
echo "$OPCACHE_INFO" >> $LOG_FILE

# Nginx cache info
NGINX_CACHE=$(du -sh /var/cache/nginx/fastcgi 2>/dev/null || echo "Cache directory not found")
echo "Nginx Cache Size: $NGINX_CACHE" >> $LOG_FILE

echo "---" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/cache-monitor.sh

# Add to crontab
sudo crontab -e
# Add: */5 * * * * /usr/local/bin/cache-monitor.sh
```

## ✅ Verification

### Cache Performance Test

```bash
# Test Redis
redis-cli ping
redis-cli info memory

# Test OPcache
php -r "var_dump(opcache_get_status());"

# Test Nginx cache
curl -I https://lms.yourdomain.com/
ls -la /var/cache/nginx/fastcgi/
```

### Expected Results

- ✅ Redis responding and caching data
- ✅ OPcache enabled and working
- ✅ Nginx FastCGI cache active
- ✅ Moodle caching configured
- ✅ Cache monitoring active

## 📝 Next Steps

Setelah caching setup selesai, lanjutkan ke:
- [03-monitoring-setup.md](03-monitoring-setup.md) - Setup monitoring system

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
