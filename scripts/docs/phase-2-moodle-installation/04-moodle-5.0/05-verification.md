# ✅ Moodle 5.0 Installation Verification

## 📋 Overview

Dokumen ini menjelaskan proses verifikasi instalasi Moodle 5.0, termasuk testing functionality, performance, dan security.

## 🎯 Objectives

- [ ] Verifikasi web interface accessibility
- [ ] Test database connectivity dan performance
- [ ] Verifikasi file permissions dan security
- [ ] Test cron job functionality
- [ ] Performance testing dan optimization

## 🔧 Step-by-Step Guide

### Step 1: Web Interface Verification

```bash
# Test web interface accessibility
curl -I https://lms.yourdomain.com

# Test specific pages
curl -I https://lms.yourdomain.com/login/index.php
curl -I https://lms.yourdomain.com/course/
curl -I https://lms.yourdomain.com/admin/
```

**Expected Results:**
- HTTP 200 OK responses
- SSL certificate working
- All pages accessible

### Step 2: Database Connectivity Test

```bash
# Test database connection
mysql -u moodle -p moodle -e "SELECT VERSION();"

# Test Moodle database tables
mysql -u moodle -p moodle -e "SHOW TABLES LIKE 'mdl_%';"

# Test database performance
mysql -u moodle -p moodle -e "SELECT COUNT(*) FROM mdl_user;"
mysql -u moodle -p moodle -e "SELECT COUNT(*) FROM mdl_course;"
```

**Expected Results:**
- Database connection successful
- All Moodle tables present
- Performance queries working

### Step 3: File Permissions Verification

```bash
# Check Moodle directory permissions
ls -la /var/www/moodle/

# Check config.php permissions
ls -la /var/www/moodle/config.php

# Check data directory permissions
ls -la /var/www/moodledata/

# Check ownership
sudo -u www-data touch /var/www/moodle/test_write.txt
sudo -u www-data rm /var/www/moodle/test_write.txt
```

**Expected Results:**
- Proper ownership (www-data:www-data)
- Correct permissions (755 for directories, 644 for files)
- Write access working

### Step 4: Cron Job Testing

```bash
# Test cron job manually
sudo -u www-data php /var/www/moodle/admin/cli/cron.php

# Check cron job logs
sudo tail -f /var/log/moodle-cron.log

# Verify cron job in crontab
sudo crontab -l
```

**Expected Results:**
- Cron job runs without errors
- Logs show successful execution
- Crontab entry present

### Step 5: Redis Session Testing

```bash
# Test Redis connection
redis-cli ping

# Check Redis memory usage
redis-cli info memory

# Test session storage
redis-cli keys "moodle_session_*"
```

**Expected Results:**
- Redis responding
- Memory usage reasonable
- Session keys present

### Step 6: PHP Configuration Verification

```bash
# Check PHP version
php -v

# Check PHP extensions
php -m | grep -E "(mysql|gd|curl|xml|mbstring|zip|intl|soap|ldap|imagick|redis|openssl|json|dom|fileinfo|iconv|simplexml|tokenizer|xmlreader|xmlwriter|exif|ftp|gettext|sodium|hash|filter)"

# Check PHP configuration
php -i | grep -E "(memory_limit|max_execution_time|upload_max_filesize|post_max_size|max_file_uploads)"
```

**Expected Results:**
- PHP 8.2.x version
- All required extensions present
- Configuration values correct

### Step 7: Security Verification

```bash
# Check SSL certificate
openssl s_client -connect lms.yourdomain.com:443 -servername lms.yourdomain.com

# Check security headers
curl -I https://lms.yourdomain.com | grep -E "(X-Frame-Options|X-XSS-Protection|X-Content-Type-Options|Referrer-Policy|Content-Security-Policy)"

# Check file access restrictions
curl -I https://lms.yourdomain.com/config.php
curl -I https://lms.yourdomain.com/install/
```

**Expected Results:**
- SSL certificate valid
- Security headers present
- Restricted files not accessible

### Step 8: Performance Testing

```bash
# Create performance test script
sudo nano /usr/local/bin/moodle-performance-test.sh
```

**Performance Test Script:**
```bash
#!/bin/bash

# Moodle 5.0 Performance Test
echo "=== Moodle 5.0 Performance Test ==="
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

# Test Redis performance
echo "4. Redis Performance Test:"
REDIS_TEST=$(redis-cli --latency -h 127.0.0.1 -p 6379 -c 10 | tail -1)
echo "   Redis latency: $REDIS_TEST"
echo ""

echo "=== Performance Test Complete ==="
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/moodle-performance-test.sh

# Run performance test
sudo /usr/local/bin/moodle-performance-test.sh
```

### Step 9: Functional Testing

```bash
# Test Moodle functionality
curl -s https://lms.yourdomain.com/ | grep -i "moodle"
curl -s https://lms.yourdomain.com/login/index.php | grep -i "login"
curl -s https://lms.yourdomain.com/course/ | grep -i "course"
```

**Expected Results:**
- Moodle interface loading
- Login page accessible
- Course page accessible

### Step 10: Log Verification

```bash
# Check Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# Check PHP-FPM logs
sudo tail -f /var/log/php8.2-fpm.log

# Check Moodle logs
sudo tail -f /var/www/moodledata/moodle.log
```

**Expected Results:**
- No critical errors in logs
- Access logs showing requests
- Error logs clean

## ✅ Verification Checklist

### System Requirements
- [ ] PHP 8.2.x installed and configured
- [ ] MariaDB/MySQL 8.0+ running
- [ ] Nginx web server configured
- [ ] Redis session storage working
- [ ] SSL certificate installed

### Moodle Installation
- [ ] Moodle 5.0 files extracted and configured
- [ ] Database connection working
- [ ] File permissions correct
- [ ] Config.php properly configured
- [ ] Data directory accessible

### Functionality
- [ ] Web interface accessible
- [ ] Login page working
- [ ] Course page accessible
- [ ] Admin panel accessible
- [ ] Cron job running

### Performance
- [ ] Response time < 3 seconds
- [ ] Database queries < 100ms
- [ ] Redis latency < 1ms
- [ ] Concurrent connections handled
- [ ] Memory usage reasonable

### Security
- [ ] SSL certificate valid
- [ ] Security headers present
- [ ] File access restricted
- [ ] Database credentials secure
- [ ] Firewall configured

## 🚨 Troubleshooting

### Common Issues

**1. Web interface not accessible**
```bash
# Check Nginx status
sudo systemctl status nginx

# Check PHP-FPM status
sudo systemctl status php8.2-fpm

# Check file permissions
sudo chown -R www-data:www-data /var/www/moodle
```

**2. Database connection error**
```bash
# Check database status
sudo systemctl status mariadb

# Test database connection
mysql -u moodle -p moodle -e "SELECT 1;"

# Check config.php
sudo nano /var/www/moodle/config.php
```

**3. Performance issues**
```bash
# Check system resources
htop
df -h
free -h

# Check PHP memory limit
php -i | grep memory_limit

# Check database performance
mysql -u root -p -e "SHOW PROCESSLIST;"
```

## 📝 Next Steps

Setelah verifikasi selesai, lanjutkan ke:
- [Phase 3: Optimization](../../phase-3-optimization/README.md)

## 📚 References

- [Moodle 5.0 Documentation](https://docs.moodle.org/500/en/Main_page)
- [Moodle Troubleshooting](https://docs.moodle.org/500/en/Troubleshooting)
- [Moodle Performance](https://docs.moodle.org/500/en/Performance)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
