# ⚡ Performance Tuning for Moodle 3.11 LTS

## 📋 Overview

Dokumen ini menjelaskan optimasi performa untuk Moodle 3.11 LTS, termasuk tuning database, PHP, web server, dan sistem operasi untuk mencapai performa maksimal.

## 🎯 Objectives

- [ ] Optimasi database performance
- [ ] Tuning PHP dan OPcache
- [ ] Optimasi web server (Nginx)
- [ ] Konfigurasi sistem operasi
- [ ] Setup monitoring performa
- [ ] Benchmark dan testing

## 📋 Prerequisites

- Moodle 3.11 LTS sudah terinstall
- Database sudah dikonfigurasi
- Web server sudah dikonfigurasi
- Monitoring tools sudah tersedia

## 🔧 Step-by-Step Guide

### Step 1: Database Performance Optimization

```bash
# Create database optimization script
sudo nano /usr/local/bin/mysql-performance-optimization.sh
```

**Database Performance Optimization Script:**
```bash
#!/bin/bash

# MySQL Performance Optimization for Moodle 3.11 LTS
echo "=== MySQL Performance Optimization ==="

# Login to MySQL
mysql -u root -p << EOF

-- Optimize InnoDB settings
SET GLOBAL innodb_buffer_pool_size = 2147483648; -- 2GB
SET GLOBAL innodb_log_file_size = 268435456; -- 256MB
SET GLOBAL innodb_log_buffer_size = 16777216; -- 16MB
SET GLOBAL innodb_flush_log_at_trx_commit = 2;
SET GLOBAL innodb_flush_method = O_DIRECT;
SET GLOBAL innodb_file_per_table = 1;
SET GLOBAL innodb_open_files = 400;
SET GLOBAL innodb_io_capacity = 400;
SET GLOBAL innodb_read_io_threads = 4;
SET GLOBAL innodb_write_io_threads = 4;

-- Optimize query cache
SET GLOBAL query_cache_type = 1;
SET GLOBAL query_cache_size = 134217728; -- 128MB
SET GLOBAL query_cache_limit = 2097152; -- 2MB

-- Optimize connection settings
SET GLOBAL max_connections = 200;
SET GLOBAL max_connect_errors = 10000;
SET GLOBAL connect_timeout = 10;
SET GLOBAL wait_timeout = 600;
SET GLOBAL interactive_timeout = 600;

-- Optimize temporary tables
SET GLOBAL tmp_table_size = 134217728; -- 128MB
SET GLOBAL max_heap_table_size = 134217728; -- 128MB

-- Optimize MyISAM settings
SET GLOBAL key_buffer_size = 33554432; -- 32MB
SET GLOBAL read_buffer_size = 2097152; -- 2MB
SET GLOBAL read_rnd_buffer_size = 8388608; -- 8MB
SET GLOBAL sort_buffer_size = 2097152; -- 2MB

-- Show current settings
SHOW VARIABLES LIKE 'innodb_buffer_pool_size';
SHOW VARIABLES LIKE 'query_cache_size';
SHOW VARIABLES LIKE 'max_connections';

EOF

echo "Database optimization completed"
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/mysql-performance-optimization.sh

# Run optimization
sudo /usr/local/bin/mysql-performance-optimization.sh
```

### Step 2: PHP Performance Optimization

```bash
# Create PHP optimization configuration
sudo nano /etc/php/8.1/fpm/conf.d/99-performance.ini
```

**PHP Performance Configuration:**
```ini
; PHP Performance Optimization for Moodle 3.11 LTS

; Memory and execution time
memory_limit = 512M
max_execution_time = 300
max_input_time = 300
max_input_vars = 3000

; File uploads
upload_max_filesize = 200M
post_max_size = 200M
max_file_uploads = 20

; OPcache optimization
opcache.enable = 1
opcache.memory_consumption = 256
opcache.interned_strings_buffer = 8
opcache.max_accelerated_files = 10000
opcache.revalidate_freq = 2
opcache.fast_shutdown = 1
opcache.validate_timestamps = 0
opcache.save_comments = 1
opcache.enable_file_override = 1
opcache.optimization_level = 0x7FFFBFFF
opcache.blacklist_filename = /etc/php/8.1/fpm/opcache-blacklist.txt

; Session optimization
session.gc_maxlifetime = 1440
session.gc_probability = 1
session.gc_divisor = 1000
session.save_handler = redis
session.save_path = "tcp://127.0.0.1:6379"

; Realpath cache
realpath_cache_size = 4096K
realpath_cache_ttl = 600

; Error handling
display_errors = Off
display_startup_errors = Off
log_errors = On
error_log = /var/log/php_errors.log
error_reporting = E_ALL & ~E_DEPRECATED & ~E_STRICT

; Performance settings
allow_url_fopen = Off
allow_url_include = Off
expose_php = Off
```

```bash
# Create OPcache blacklist
sudo nano /etc/php/8.1/fpm/opcache-blacklist.txt
```

**OPcache Blacklist:**
```
/var/www/moodle/config.php
/var/www/moodle/version.php
/var/www/moodle/lib/setup.php
```

```bash
# Restart PHP-FPM
sudo systemctl restart php8.1-fpm
```

### Step 3: Nginx Performance Optimization

```bash
# Create Nginx performance configuration
sudo nano /etc/nginx/conf.d/performance.conf
```

**Nginx Performance Configuration:**
```nginx
# Nginx Performance Optimization for Moodle 3.11 LTS

# Worker processes
worker_processes auto;
worker_cpu_affinity auto;

# Worker connections
events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

# HTTP optimization
http {
    # Basic settings
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    keepalive_requests 100;
    types_hash_max_size 2048;
    server_tokens off;

    # Buffer sizes
    client_body_buffer_size 128k;
    client_max_body_size 200m;
    client_header_buffer_size 1k;
    large_client_header_buffers 4 4k;
    output_buffers 1 32k;
    postpone_output 1460;

    # Timeouts
    client_body_timeout 12;
    client_header_timeout 12;
    send_timeout 10;

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

    # FastCGI optimization
    fastcgi_cache_path /var/cache/nginx/fastcgi levels=1:2 keys_zone=moodle:100m inactive=60m;
    fastcgi_cache_key "$scheme$request_method$host$request_uri";
    fastcgi_cache_use_stale error timeout invalid_header http_500;
    fastcgi_ignore_headers Cache-Control Expires Set-Cookie;

    # Open file cache
    open_file_cache max=1000 inactive=20s;
    open_file_cache_valid 30s;
    open_file_cache_min_uses 2;
    open_file_cache_errors on;
}
```

```bash
# Create cache directory
sudo mkdir -p /var/cache/nginx/fastcgi
sudo chown -R www-data:www-data /var/cache/nginx

# Update Moodle site configuration
sudo nano /etc/nginx/sites-available/moodle
```

**Add to Moodle site configuration:**
```nginx
# FastCGI cache configuration
location ~ \.php$ {
    include snippets/fastcgi-php.conf;
    fastcgi_pass unix:/var/run/php/php8.1-fpm-moodle.sock;
    fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    include fastcgi_params;
    
    # FastCGI cache
    fastcgi_cache moodle;
    fastcgi_cache_valid 200 60m;
    fastcgi_cache_valid 404 1m;
    fastcgi_cache_bypass $skip_cache;
    fastcgi_no_cache $skip_cache;
    
    # FastCGI settings
    fastcgi_connect_timeout 60s;
    fastcgi_send_timeout 60s;
    fastcgi_read_timeout 60s;
    fastcgi_buffer_size 128k;
    fastcgi_buffers 4 256k;
    fastcgi_busy_buffers_size 256k;
    fastcgi_temp_file_write_size 256k;
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

```bash
# Test and reload Nginx
sudo nginx -t
sudo systemctl reload nginx
```

### Step 4: Redis Performance Optimization

```bash
# Create Redis performance configuration
sudo nano /etc/redis/redis.conf
```

**Redis Performance Configuration:**
```
# Redis Performance Optimization for Moodle 3.11 LTS

# Network
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
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir /var/lib/redis

# Logging
loglevel notice
logfile /var/log/redis/redis-server.log
syslog-enabled no

# Performance
tcp-backlog 511
databases 16
always-show-logo yes
```

```bash
# Restart Redis
sudo systemctl restart redis-server

# Test Redis performance
redis-cli --latency -h 127.0.0.1 -p 6379
```

### Step 5: System Performance Optimization

```bash
# Create system performance optimization script
sudo nano /usr/local/bin/system-performance-optimization.sh
```

**System Performance Optimization Script:**
```bash
#!/bin/bash

# System Performance Optimization for Moodle 3.11 LTS
echo "=== System Performance Optimization ==="

# Kernel parameters optimization
echo "Optimizing kernel parameters..."
cat >> /etc/sysctl.d/99-moodle-performance.conf << EOF
# Network optimizations
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.core.rmem_default = 262144
net.core.wmem_default = 262144
net.ipv4.tcp_rmem = 4096 65536 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
net.ipv4.tcp_congestion_control = bbr
net.ipv4.tcp_slow_start_after_idle = 0
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.tcp_keepalive_intvl = 30
net.ipv4.tcp_keepalive_probes = 3

# File system optimizations
fs.file-max = 65536
fs.inotify.max_user_watches = 524288
fs.inotify.max_user_instances = 256
fs.inotify.max_queued_events = 32768

# Memory management
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
vm.vfs_cache_pressure = 50

# Process limits
kernel.pid_max = 4194304
kernel.threads-max = 2097152
EOF

# Apply kernel parameters
sysctl -p /etc/sysctl.d/99-moodle-performance.conf

# Optimize file limits
echo "Optimizing file limits..."
cat >> /etc/security/limits.d/99-moodle-performance.conf << EOF
# Moodle performance limits
moodle soft nofile 65536
moodle hard nofile 65536
moodle soft nproc 32768
moodle hard nproc 32768

www-data soft nofile 65536
www-data hard nofile 65536
www-data soft nproc 32768
www-data hard nproc 32768

root soft nofile 65536
root hard nofile 65536
root soft nproc 32768
root hard nproc 32768
EOF

# Optimize I/O scheduler
echo "Optimizing I/O scheduler..."
echo mq-deadline > /sys/block/sda/queue/scheduler
echo 1024 > /sys/block/sda/queue/nr_requests

# Optimize CPU governor
echo "Optimizing CPU governor..."
echo performance > /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

echo "System performance optimization completed"
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/system-performance-optimization.sh

# Run optimization
sudo /usr/local/bin/system-performance-optimization.sh
```

### Step 6: Performance Monitoring Setup

```bash
# Create performance monitoring script
sudo nano /usr/local/bin/performance-monitor.sh
```

**Performance Monitoring Script:**
```bash
#!/bin/bash

# Performance Monitoring for Moodle 3.11 LTS
LOG_FILE="/var/log/performance-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Performance monitoring..." >> $LOG_FILE

# CPU usage
CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')
echo "CPU Usage: $CPU_USAGE%" >> $LOG_FILE

# Memory usage
MEMORY_USAGE=$(free | awk 'NR==2{printf "%.2f", $3*100/$2}')
echo "Memory Usage: $MEMORY_USAGE%" >> $LOG_FILE

# Disk I/O
DISK_IO=$(iostat -x 1 1 | grep -E "(Device|sda)" | tail -1)
echo "Disk I/O: $DISK_IO" >> $LOG_FILE

# Network usage
NETWORK_USAGE=$(iftop -t -s 1 -L 1 | grep -E "(TX|RX)" | tail -2)
echo "Network Usage: $NETWORK_USAGE" >> $LOG_FILE

# Database connections
DB_CONNECTIONS=$(mysql -u root -p'your_password' -e "SHOW STATUS LIKE 'Threads_connected';" | awk 'NR==2 {print $2}')
echo "Database Connections: $DB_CONNECTIONS" >> $LOG_FILE

# Redis memory usage
REDIS_MEMORY=$(redis-cli info memory | grep used_memory_human | cut -d: -f2)
echo "Redis Memory: $REDIS_MEMORY" >> $LOG_FILE

# Nginx active connections
NGINX_CONNECTIONS=$(netstat -an | grep :443 | grep ESTABLISHED | wc -l)
echo "Nginx Active Connections: $NGINX_CONNECTIONS" >> $LOG_FILE

# PHP-FPM processes
PHP_FPM_PROCESSES=$(ps aux | grep php-fpm | grep -v grep | wc -l)
echo "PHP-FPM Processes: $PHP_FPM_PROCESSES" >> $LOG_FILE

echo "---" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/performance-monitor.sh

# Add to crontab (every 5 minutes)
sudo crontab -e
```

**Add to crontab:**
```
# Performance monitoring every 5 minutes
*/5 * * * * /usr/local/bin/performance-monitor.sh
```

### Step 7: Performance Benchmarking

```bash
# Create performance benchmark script
sudo nano /usr/local/bin/performance-benchmark.sh
```

**Performance Benchmark Script:**
```bash
#!/bin/bash

# Performance Benchmark for Moodle 3.11 LTS
echo "=== Moodle Performance Benchmark ==="
echo "Date: $(date)"
echo ""

# Test response time
echo "1. Response Time Test:"
RESPONSE_TIME=$(curl -o /dev/null -s -w '%{time_total}' https://lms.yourdomain.com/)
echo "   Homepage response time: ${RESPONSE_TIME}s"

LOGIN_TIME=$(curl -o /dev/null -s -w '%{time_total}' https://lms.yourdomain.com/login/index.php)
echo "   Login page response time: ${LOGIN_TIME}s"

COURSE_TIME=$(curl -o /dev/null -s -w '%{time_total}' https://lms.yourdomain.com/course/)
echo "   Course page response time: ${COURSE_TIME}s"
echo ""

# Test database performance
echo "2. Database Performance Test:"
DB_QUERY_TIME=$(time mysql -u moodle -p'your_password' moodle -e "SELECT COUNT(*) FROM mdl_user;" 2>&1 | grep real)
echo "   Database query performance: $DB_QUERY_TIME"
echo ""

# Test concurrent connections
echo "3. Concurrent Connection Test:"
for i in {1..20}; do
    curl -o /dev/null -s https://lms.yourdomain.com/ &
done
wait
echo "   ✓ 20 concurrent connections handled"
echo ""

# Test file upload performance
echo "4. File Upload Test:"
echo "Test file content" > /tmp/test_upload.txt
UPLOAD_TEST=$(curl -o /dev/null -s -w '%{time_total}' -X POST -F "file=@/tmp/test_upload.txt" https://lms.yourdomain.com/)
echo "   File upload simulation: ${UPLOAD_TEST}s"
rm /tmp/test_upload.txt
echo ""

# Test Redis performance
echo "5. Redis Performance Test:"
REDIS_TEST=$(redis-cli --latency -h 127.0.0.1 -p 6379 -c 10 | tail -1)
echo "   Redis latency: $REDIS_TEST"
echo ""

echo "=== Benchmark Complete ==="
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/performance-benchmark.sh

# Run benchmark
sudo /usr/local/bin/performance-benchmark.sh
```

## ✅ Verification

### Performance Check

```bash
# Run comprehensive performance check
sudo /usr/local/bin/performance-benchmark.sh

# Check system resources
htop
iotop
nethogs

# Check database performance
mysql -u root -p -e "SHOW STATUS LIKE 'Questions';"
mysql -u root -p -e "SHOW STATUS LIKE 'Slow_queries';"

# Check Redis performance
redis-cli info stats

# Check Nginx performance
sudo nginx -T | grep -E "(worker_processes|worker_connections)"
```

### Expected Results

- ✅ Response time < 2 seconds
- ✅ Database queries < 100ms
- ✅ Memory usage < 80%
- ✅ CPU usage < 70%
- ✅ Redis latency < 1ms
- ✅ Concurrent connections handled
- ✅ File uploads working

## 🚨 Troubleshooting

### Common Issues

**1. High memory usage**
```bash
# Check memory usage
free -h
ps aux --sort=-%mem | head -10

# Optimize PHP memory
sudo nano /etc/php/8.1/fpm/conf.d/99-performance.ini
# Adjust memory_limit
```

**2. Slow database queries**
```bash
# Check slow query log
sudo tail -f /var/log/mysql/slow.log

# Optimize database
sudo /usr/local/bin/mysql-performance-optimization.sh
```

**3. High CPU usage**
```bash
# Check CPU usage
top
htop

# Check PHP-FPM processes
ps aux | grep php-fpm
```

## 📝 Next Steps

Setelah performance tuning selesai, lanjutkan ke:
- [02-caching-setup.md](02-caching-setup.md) - Setup caching system

## 📚 References

- [Moodle Performance](https://docs.moodle.org/311/en/Performance)
- [MySQL Performance Tuning](https://dev.mysql.com/doc/refman/8.0/en/optimization.html)
- [Nginx Performance Tuning](https://nginx.org/en/docs/http/ngx_http_core_module.html)
- [PHP OPcache](https://www.php.net/manual/en/book.opcache.php)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
