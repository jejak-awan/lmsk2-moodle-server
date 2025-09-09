# ⚖️ Load Balancing Setup for Moodle

## 📋 Overview

Dokumen ini menjelaskan setup load balancing untuk Moodle production environment, termasuk Nginx load balancer, health checks, dan failover configuration.

## 🎯 Objectives

- [ ] Setup Nginx load balancer
- [ ] Konfigurasi health checks
- [ ] Setup failover dan redundancy
- [ ] Konfigurasi session persistence
- [ ] Optimasi load balancing algorithms

## 🔧 Step-by-Step Guide

### Step 1: Install Nginx Load Balancer

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

### Step 2: Configure Load Balancer

```bash
# Create load balancer configuration
sudo nano /etc/nginx/sites-available/load-balancer
```

**Load Balancer Configuration:**
```nginx
# Moodle Load Balancer Configuration
upstream moodle_backend {
    # Load balancing method
    least_conn;
    
    # Backend servers
    server 192.168.1.10:80 weight=3 max_fails=3 fail_timeout=30s;
    server 192.168.1.11:80 weight=3 max_fails=3 fail_timeout=30s;
    server 192.168.1.12:80 weight=2 max_fails=3 fail_timeout=30s;
    
    # Health check
    keepalive 32;
}

# Health check endpoint
server {
    listen 80;
    server_name health.yourdomain.com;
    
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}

# Main load balancer
server {
    listen 80;
    server_name lms.yourdomain.com;
    
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name lms.yourdomain.com;
    
    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/lms.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/lms.yourdomain.com/privkey.pem;
    
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
    
    # Proxy settings
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $host;
    proxy_set_header X-Forwarded-Port $server_port;
    
    # Timeout settings
    proxy_connect_timeout 30s;
    proxy_send_timeout 30s;
    proxy_read_timeout 30s;
    
    # Buffer settings
    proxy_buffering on;
    proxy_buffer_size 4k;
    proxy_buffers 8 4k;
    proxy_busy_buffers_size 8k;
    
    # Main location block
    location / {
        proxy_pass http://moodle_backend;
        proxy_redirect off;
        
        # Health check
        proxy_next_upstream error timeout invalid_header http_500 http_502 http_503 http_504;
        proxy_next_upstream_tries 3;
        proxy_next_upstream_timeout 30s;
    }
    
    # Static files caching
    location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://moodle_backend;
        proxy_cache_valid 200 1y;
        proxy_cache_valid 404 1m;
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
    }
    
    # Admin access restriction
    location /admin/ {
        allow 192.168.1.0/24;
        deny all;
        proxy_pass http://moodle_backend;
    }
}
```

### Step 3: Enable Load Balancer

```bash
# Enable load balancer site
sudo ln -s /etc/nginx/sites-available/load-balancer /etc/nginx/sites-enabled/

# Remove default site
sudo rm /etc/nginx/sites-enabled/default

# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

### Step 4: Configure Health Checks

```bash
# Create health check script
sudo nano /usr/local/bin/health-check.sh
```

**Health Check Script:**
```bash
#!/bin/bash

# Load Balancer Health Check Script
LOG_FILE="/var/log/health-check.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

# Backend servers
SERVERS=("192.168.1.10" "192.168.1.11" "192.168.1.12")

echo "[$DATE] Health check started..." >> $LOG_FILE

for server in "${SERVERS[@]}"; do
    # Check HTTP response
    if curl -s --max-time 10 "http://$server/health" | grep -q "healthy"; then
        echo "Server $server: Healthy" >> $LOG_FILE
    else
        echo "Server $server: Unhealthy" >> $LOG_FILE
        
        # Remove from load balancer (if using dynamic configuration)
        # sed -i "/server $server:80/d" /etc/nginx/sites-available/load-balancer
        # nginx -s reload
    fi
done

echo "---" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/health-check.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# Health check every minute
* * * * * /usr/local/bin/health-check.sh
```

### Step 5: Configure Session Persistence

```bash
# Update load balancer configuration for session persistence
sudo nano /etc/nginx/sites-available/load-balancer
```

**Add to upstream block:**
```nginx
upstream moodle_backend {
    # Load balancing method
    ip_hash;  # Use IP hash for session persistence
    
    # Backend servers
    server 192.168.1.10:80 weight=3 max_fails=3 fail_timeout=30s;
    server 192.168.1.11:80 weight=3 max_fails=3 fail_timeout=30s;
    server 192.168.1.12:80 weight=2 max_fails=3 fail_timeout=30s;
    
    # Health check
    keepalive 32;
}
```

### Step 6: Configure Failover

```bash
# Create failover configuration
sudo nano /etc/nginx/conf.d/failover.conf
```

**Failover Configuration:**
```nginx
# Failover Configuration
upstream moodle_backend {
    # Primary servers
    server 192.168.1.10:80 weight=3 max_fails=3 fail_timeout=30s;
    server 192.168.1.11:80 weight=3 max_fails=3 fail_timeout=30s;
    
    # Backup servers
    server 192.168.1.12:80 weight=2 max_fails=3 fail_timeout=30s backup;
    server 192.168.1.13:80 weight=2 max_fails=3 fail_timeout=30s backup;
}

# Error page configuration
error_page 502 503 504 /50x.html;
location = /50x.html {
    root /usr/share/nginx/html;
}
```

### Step 7: Configure Monitoring

```bash
# Create load balancer monitoring script
sudo nano /usr/local/bin/load-balancer-monitor.sh
```

**Load Balancer Monitoring Script:**
```bash
#!/bin/bash

# Load Balancer Monitoring Script
LOG_FILE="/var/log/load-balancer-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Load balancer monitoring..." >> $LOG_FILE

# Check Nginx status
if systemctl is-active --quiet nginx; then
    echo "Nginx: Running" >> $LOG_FILE
else
    echo "Nginx: Not running" >> $LOG_FILE
fi

# Check backend servers
SERVERS=("192.168.1.10" "192.168.1.11" "192.168.1.12")
for server in "${SERVERS[@]}"; do
    RESPONSE_TIME=$(curl -o /dev/null -s -w '%{time_total}' "http://$server/health")
    if [ $? -eq 0 ]; then
        echo "Server $server: Response time ${RESPONSE_TIME}s" >> $LOG_FILE
    else
        echo "Server $server: Unreachable" >> $LOG_FILE
    fi
done

# Check load balancer response
LB_RESPONSE_TIME=$(curl -o /dev/null -s -w '%{time_total}' "https://lms.yourdomain.com/")
echo "Load balancer response time: ${LB_RESPONSE_TIME}s" >> $LOG_FILE

# Check active connections
ACTIVE_CONNECTIONS=$(netstat -an | grep :443 | grep ESTABLISHED | wc -l)
echo "Active connections: $ACTIVE_CONNECTIONS" >> $LOG_FILE

echo "---" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/load-balancer-monitor.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# Load balancer monitoring every 5 minutes
*/5 * * * * /usr/local/bin/load-balancer-monitor.sh
```

### Step 8: Configure Logging

```bash
# Configure Nginx logging
sudo nano /etc/nginx/nginx.conf
```

**Add to http block:**
```nginx
http {
    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for" '
                    'rt=$request_time uct="$upstream_connect_time" '
                    'uht="$upstream_header_time" urt="$upstream_response_time"';
    
    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log;
}
```

## ✅ Verification

### Load Balancer Test

```bash
# Test load balancer configuration
sudo nginx -t

# Test backend servers
curl -I http://192.168.1.10/health
curl -I http://192.168.1.11/health
curl -I http://192.168.1.12/health

# Test load balancer
curl -I https://lms.yourdomain.com

# Test session persistence
curl -c cookies.txt https://lms.yourdomain.com
curl -b cookies.txt https://lms.yourdomain.com
```

### Expected Results

- ✅ Load balancer running
- ✅ Backend servers healthy
- ✅ Session persistence working
- ✅ Failover configured
- ✅ Monitoring active

## 🚨 Troubleshooting

### Common Issues

**1. Backend servers not responding**
```bash
# Check server status
curl -I http://192.168.1.10/health

# Check Nginx configuration
sudo nginx -t

# Check logs
sudo tail -f /var/log/nginx/error.log
```

**2. Session persistence not working**
```bash
# Check IP hash configuration
grep -A 10 "upstream moodle_backend" /etc/nginx/sites-available/load-balancer

# Test session persistence
curl -c cookies.txt https://lms.yourdomain.com
curl -b cookies.txt https://lms.yourdomain.com
```

**3. Load balancer performance issues**
```bash
# Check active connections
netstat -an | grep :443 | grep ESTABLISHED | wc -l

# Check backend server performance
htop
```

## 📝 Next Steps

Setelah load balancing setup selesai, lanjutkan ke:
- [04-maintenance-procedures.md](04-maintenance-procedures.md) - Setup maintenance procedures

## 📚 References

- [Nginx Load Balancing](https://nginx.org/en/docs/http/load_balancing.html)
- [Nginx Upstream Module](https://nginx.org/en/docs/http/ngx_http_upstream_module.html)
- [Load Balancing Algorithms](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
