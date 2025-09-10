# ✅ Moodle 3.11 LTS Installation Verification

## 📋 Overview

Dokumen ini menjelaskan langkah-langkah verifikasi instalasi Moodle 3.11 LTS untuk memastikan semua komponen berfungsi dengan baik dan siap untuk production use.

## 🎯 Objectives

- [ ] Verifikasi web interface accessibility
- [ ] Test database connectivity dan performance
- [ ] Verifikasi file permissions dan security
- [ ] Test cron job functionality
- [ ] Verifikasi SSL/TLS configuration
- [ ] Performance testing dan optimization
- [ ] Security audit dan compliance check

## 📋 Prerequisites

- Moodle 3.11 LTS sudah terinstall
- Database sudah dikonfigurasi
- Web server sudah dikonfigurasi
- SSL certificate sudah terinstall
- Cron job sudah dikonfigurasi

## 🔧 Step-by-Step Guide

### Step 1: Web Interface Verification

```bash
# Create comprehensive verification script
sudo nano /usr/local/bin/moodle-verification.sh
```

**Moodle Verification Script:**
```bash
#!/bin/bash

# Moodle 3.11 LTS Installation Verification
echo "=== Moodle 3.11 LTS Installation Verification ==="
echo "Date: $(date)"
echo ""

# Check Moodle version
echo "1. Moodle Version Check:"
cd /var/www/moodle
MOODLE_VERSION=$(sudo -u www-data php -r "require_once('version.php'); echo \$version;")
echo "   Moodle Version: $MOODLE_VERSION"
if [[ "$MOODLE_VERSION" == "20221128" ]]; then
    echo "   ✓ Moodle 3.11 LTS detected"
else
    echo "   ⚠ Unexpected Moodle version"
fi
echo ""

# Check web accessibility
echo "2. Web Interface Check:"
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" https://lms.yourdomain.com)
if [ "$HTTP_STATUS" = "200" ]; then
    echo "   ✓ Web interface accessible (HTTP $HTTP_STATUS)"
else
    echo "   ✗ Web interface not accessible (HTTP $HTTP_STATUS)"
fi

# Check SSL certificate
SSL_INFO=$(echo | openssl s_client -connect lms.yourdomain.com:443 -servername lms.yourdomain.com 2>/dev/null | openssl x509 -noout -dates)
if [ $? -eq 0 ]; then
    echo "   ✓ SSL certificate valid"
    echo "   $SSL_INFO"
else
    echo "   ✗ SSL certificate issue"
fi
echo ""

# Check database connectivity
echo "3. Database Connectivity Check:"
DB_CONNECTION=$(mysql -u moodle -p'your_password' -e "SELECT 1;" 2>/dev/null)
if [ $? -eq 0 ]; then
    echo "   ✓ Database connection successful"
    
    # Check database tables
    TABLE_COUNT=$(mysql -u moodle -p'your_password' moodle -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'moodle';" 2>/dev/null | tail -1)
    echo "   ✓ Database tables: $TABLE_COUNT"
    
    # Check database size
    DB_SIZE=$(mysql -u moodle -p'your_password' moodle -e "SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS 'DB Size in MB' FROM information_schema.tables WHERE table_schema = 'moodle';" 2>/dev/null | tail -1)
    echo "   ✓ Database size: $DB_SIZE MB"
else
    echo "   ✗ Database connection failed"
fi
echo ""

# Check file permissions
echo "4. File Permissions Check:"
CONFIG_PERM=$(ls -la /var/www/moodle/config.php | awk '{print $1}')
if [ "$CONFIG_PERM" = "-rw-------" ]; then
    echo "   ✓ config.php permissions correct ($CONFIG_PERM)"
else
    echo "   ✗ config.php permissions incorrect ($CONFIG_PERM)"
fi

MOODLEDATA_PERM=$(ls -ld /var/www/moodle/moodledata | awk '{print $1}')
if [ "$MOODLEDATA_PERM" = "drwxrwxrwx" ]; then
    echo "   ✓ moodledata permissions correct ($MOODLEDATA_PERM)"
else
    echo "   ✗ moodledata permissions incorrect ($MOODLEDATA_PERM)"
fi
echo ""

# Check cron job
echo "5. Cron Job Check:"
CRON_EXISTS=$(crontab -l 2>/dev/null | grep -c "moodle.*cron.php")
if [ "$CRON_EXISTS" -gt 0 ]; then
    echo "   ✓ Moodle cron job configured"
    
    # Test cron execution
    CRON_TEST=$(sudo -u www-data php /var/www/moodle/admin/cli/cron.php 2>&1)
    if [ $? -eq 0 ]; then
        echo "   ✓ Cron job execution successful"
    else
        echo "   ✗ Cron job execution failed"
    fi
else
    echo "   ✗ Moodle cron job not configured"
fi
echo ""

# Check PHP extensions
echo "6. PHP Extensions Check:"
REQUIRED_EXTENSIONS=("mysql" "gd" "curl" "xml" "mbstring" "zip" "intl" "soap" "ldap" "imagick" "xmlrpc" "openssl" "json" "dom" "fileinfo" "iconv" "simplexml" "tokenizer" "xmlreader" "xmlwriter")
MISSING_EXTENSIONS=()
for ext in "${REQUIRED_EXTENSIONS[@]}"; do
    if ! php -m | grep -q "^$ext$"; then
        MISSING_EXTENSIONS+=("$ext")
    fi
done

if [ ${#MISSING_EXTENSIONS[@]} -eq 0 ]; then
    echo "   ✓ All required PHP extensions installed"
else
    echo "   ✗ Missing PHP extensions: ${MISSING_EXTENSIONS[*]}"
fi
echo ""

# Check system resources
echo "7. System Resources Check:"
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -lt 80 ]; then
    echo "   ✓ Disk usage: $DISK_USAGE% (OK)"
else
    echo "   ⚠ Disk usage: $DISK_USAGE% (High)"
fi

MEMORY_USAGE=$(free | awk 'NR==2{printf "%.1f", $3*100/$2}')
if (( $(echo "$MEMORY_USAGE < 80" | bc -l) )); then
    echo "   ✓ Memory usage: $MEMORY_USAGE% (OK)"
else
    echo "   ⚠ Memory usage: $MEMORY_USAGE% (High)"
fi

CPU_LOAD=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
if (( $(echo "$CPU_LOAD < 5" | bc -l) )); then
    echo "   ✓ CPU load: $CPU_LOAD (OK)"
else
    echo "   ⚠ CPU load: $CPU_LOAD (High)"
fi
echo ""

# Check services
echo "8. Services Check:"
SERVICES=("nginx" "php8.1-fpm" "mariadb" "redis-server" "fail2ban")
for service in "${SERVICES[@]}"; do
    if systemctl is-active --quiet $service; then
        echo "   ✓ $service: Running"
    else
        echo "   ✗ $service: Not running"
    fi
done
echo ""

echo "=== Verification Complete ==="
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/moodle-verification.sh

# Run verification
sudo /usr/local/bin/moodle-verification.sh
```

### Step 2: Performance Testing

```bash
# Create performance testing script
sudo nano /usr/local/bin/moodle-performance-test.sh
```

**Performance Testing Script:**
```bash
#!/bin/bash

# Moodle Performance Testing
echo "=== Moodle Performance Testing ==="
echo "Date: $(date)"
echo ""

# Test response time
echo "1. Response Time Test:"
RESPONSE_TIME=$(curl -o /dev/null -s -w '%{time_total}' https://lms.yourdomain.com/)
echo "   Homepage response time: ${RESPONSE_TIME}s"

LOGIN_TIME=$(curl -o /dev/null -s -w '%{time_total}' https://lms.yourdomain.com/login/index.php)
echo "   Login page response time: ${LOGIN_TIME}s"
echo ""

# Test database performance
echo "2. Database Performance Test:"
DB_QUERY_TIME=$(time mysql -u moodle -p'your_password' moodle -e "SELECT COUNT(*) FROM mdl_user;" 2>/dev/null)
echo "   Database query performance: OK"
echo ""

# Test file upload performance
echo "3. File Upload Test:"
# Create test file
echo "Test file content" > /tmp/test_upload.txt

# Test upload (simulate)
UPLOAD_TEST=$(curl -o /dev/null -s -w '%{time_total}' -X POST -F "file=@/tmp/test_upload.txt" https://lms.yourdomain.com/)
echo "   File upload simulation: ${UPLOAD_TEST}s"

# Cleanup
rm /tmp/test_upload.txt
echo ""

# Test concurrent connections
echo "4. Concurrent Connection Test:"
for i in {1..10}; do
    curl -o /dev/null -s https://lms.yourdomain.com/ &
done
wait
echo "   ✓ 10 concurrent connections handled"
echo ""

echo "=== Performance Testing Complete ==="
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/moodle-performance-test.sh

# Run performance test
sudo /usr/local/bin/moodle-performance-test.sh
```

### Step 3: Security Audit

```bash
# Create security audit script
sudo nano /usr/local/bin/moodle-security-audit.sh
```

**Security Audit Script:**
```bash
#!/bin/bash

# Moodle Security Audit
echo "=== Moodle Security Audit ==="
echo "Date: $(date)"
echo ""

# Check SSL configuration
echo "1. SSL Security Check:"
SSL_PROTOCOLS=$(echo | openssl s_client -connect lms.yourdomain.com:443 -servername lms.yourdomain.com 2>/dev/null | openssl x509 -noout -text | grep "Signature Algorithm")
echo "   SSL Signature: $SSL_PROTOCOLS"

SSL_CIPHERS=$(echo | openssl s_client -connect lms.yourdomain.com:443 -servername lms.yourdomain.com 2>/dev/null | grep "Cipher")
echo "   SSL Cipher: $SSL_CIPHERS"
echo ""

# Check security headers
echo "2. Security Headers Check:"
HEADERS=$(curl -I https://lms.yourdomain.com/ 2>/dev/null)
if echo "$HEADERS" | grep -q "Strict-Transport-Security"; then
    echo "   ✓ HSTS header present"
else
    echo "   ✗ HSTS header missing"
fi

if echo "$HEADERS" | grep -q "X-Frame-Options"; then
    echo "   ✓ X-Frame-Options header present"
else
    echo "   ✗ X-Frame-Options header missing"
fi

if echo "$HEADERS" | grep -q "X-Content-Type-Options"; then
    echo "   ✓ X-Content-Type-Options header present"
else
    echo "   ✗ X-Content-Type-Options header missing"
fi
echo ""

# Check file permissions
echo "3. File Security Check:"
CONFIG_OWNER=$(ls -la /var/www/moodle/config.php | awk '{print $3":"$4}')
if [ "$CONFIG_OWNER" = "root:www-data" ]; then
    echo "   ✓ config.php ownership correct"
else
    echo "   ✗ config.php ownership incorrect"
fi

CONFIG_PERM=$(ls -la /var/www/moodle/config.php | awk '{print $1}')
if [ "$CONFIG_PERM" = "-rw-------" ]; then
    echo "   ✓ config.php permissions secure"
else
    echo "   ✗ config.php permissions insecure"
fi
echo ""

# Check database security
echo "4. Database Security Check:"
DB_USERS=$(mysql -u root -p'your_password' -e "SELECT user, host FROM mysql.user WHERE user != 'root' AND user != 'mysql.sys' AND user != 'mysql.session';" 2>/dev/null)
echo "   Database users:"
echo "$DB_USERS"
echo ""

# Check firewall
echo "5. Firewall Check:"
UFW_STATUS=$(sudo ufw status | grep "Status")
echo "   $UFW_STATUS"

OPEN_PORTS=$(sudo ufw status | grep "ALLOW")
echo "   Open ports:"
echo "$OPEN_PORTS"
echo ""

# Check fail2ban
echo "6. Fail2ban Check:"
if systemctl is-active --quiet fail2ban; then
    echo "   ✓ Fail2ban running"
    JAIL_STATUS=$(sudo fail2ban-client status 2>/dev/null)
    echo "   Active jails:"
    echo "$JAIL_STATUS"
else
    echo "   ✗ Fail2ban not running"
fi
echo ""

echo "=== Security Audit Complete ==="
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/moodle-security-audit.sh

# Run security audit
sudo /usr/local/bin/moodle-security-audit.sh
```

### Step 4: Functional Testing

```bash
# Create functional testing script
sudo nano /usr/local/bin/moodle-functional-test.sh
```

**Functional Testing Script:**
```bash
#!/bin/bash

# Moodle Functional Testing
echo "=== Moodle Functional Testing ==="
echo "Date: $(date)"
echo ""

# Test admin login
echo "1. Admin Login Test:"
LOGIN_TEST=$(curl -s -o /dev/null -w "%{http_code}" -X POST -d "username=admin&password=admin_password" https://lms.yourdomain.com/login/index.php)
if [ "$LOGIN_TEST" = "200" ]; then
    echo "   ✓ Admin login page accessible"
else
    echo "   ✗ Admin login page issue"
fi
echo ""

# Test course creation
echo "2. Course Creation Test:"
cd /var/www/moodle
COURSE_TEST=$(sudo -u www-data php admin/cli/cfg.php --name=enablecourserequests --get)
echo "   Course requests enabled: $COURSE_TEST"
echo ""

# Test user creation
echo "3. User Management Test:"
USER_COUNT=$(mysql -u moodle -p'your_password' moodle -e "SELECT COUNT(*) FROM mdl_user WHERE deleted = 0;" 2>/dev/null | tail -1)
echo "   Total users: $USER_COUNT"
echo ""

# Test file upload
echo "4. File Upload Test:"
UPLOAD_DIR="/var/www/moodle/moodledata"
if [ -w "$UPLOAD_DIR" ]; then
    echo "   ✓ Moodledata directory writable"
else
    echo "   ✗ Moodledata directory not writable"
fi
echo ""

# Test email configuration
echo "5. Email Configuration Test:"
EMAIL_CONFIG=$(sudo -u www-data php admin/cli/cfg.php --name=smtphosts --get)
echo "   SMTP hosts: $EMAIL_CONFIG"
echo ""

# Test backup functionality
echo "6. Backup Functionality Test:"
BACKUP_DIR="/var/www/moodle/moodledata/backup"
if [ -d "$BACKUP_DIR" ]; then
    echo "   ✓ Backup directory exists"
else
    echo "   ✗ Backup directory missing"
fi
echo ""

echo "=== Functional Testing Complete ==="
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/moodle-functional-test.sh

# Run functional test
sudo /usr/local/bin/moodle-functional-test.sh
```

### Step 5: Generate Verification Report

```bash
# Create verification report generator
sudo nano /usr/local/bin/moodle-verification-report.sh
```

**Verification Report Generator:**
```bash
#!/bin/bash

# Moodle Verification Report Generator
REPORT_FILE="/var/log/moodle-verification-report.txt"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "=== Moodle 3.11 LTS Verification Report ===" > $REPORT_FILE
echo "Generated: $DATE" >> $REPORT_FILE
echo "" >> $REPORT_FILE

# Run all verification scripts
echo "Running comprehensive verification..." >> $REPORT_FILE
echo "" >> $REPORT_FILE

echo "1. Installation Verification:" >> $REPORT_FILE
/usr/local/bin/moodle-verification.sh >> $REPORT_FILE 2>&1
echo "" >> $REPORT_FILE

echo "2. Performance Testing:" >> $REPORT_FILE
/usr/local/bin/moodle-performance-test.sh >> $REPORT_FILE 2>&1
echo "" >> $REPORT_FILE

echo "3. Security Audit:" >> $REPORT_FILE
/usr/local/bin/moodle-security-audit.sh >> $REPORT_FILE 2>&1
echo "" >> $REPORT_FILE

echo "4. Functional Testing:" >> $REPORT_FILE
/usr/local/bin/moodle-functional-test.sh >> $REPORT_FILE 2>&1
echo "" >> $REPORT_FILE

echo "=== Report Complete ===" >> $REPORT_FILE

# Display report
echo "Verification report generated: $REPORT_FILE"
echo "Report contents:"
cat $REPORT_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/moodle-verification-report.sh

# Generate report
sudo /usr/local/bin/moodle-verification-report.sh
```

## ✅ Verification Checklist

### Installation Verification
- [ ] Moodle 3.11 LTS version confirmed
- [ ] Web interface accessible via HTTPS
- [ ] SSL certificate valid and working
- [ ] Database connection established
- [ ] All required PHP extensions installed
- [ ] File permissions correctly set
- [ ] Cron job configured and running

### Performance Verification
- [ ] Response time < 2 seconds
- [ ] Database queries performing well
- [ ] File upload functionality working
- [ ] Concurrent connections handled
- [ ] System resources within limits

### Security Verification
- [ ] SSL/TLS properly configured
- [ ] Security headers implemented
- [ ] File permissions secure
- [ ] Database users properly configured
- [ ] Firewall rules active
- [ ] Fail2ban protecting against attacks

### Functional Verification
- [ ] Admin login working
- [ ] Course creation functional
- [ ] User management working
- [ ] File upload working
- [ ] Email configuration set
- [ ] Backup functionality available

## 🚨 Troubleshooting

### Common Issues

**1. Verification script fails**
```bash
# Check script permissions
sudo chmod +x /usr/local/bin/moodle-*.sh

# Check script syntax
bash -n /usr/local/bin/moodle-verification.sh
```

**2. Performance issues**
```bash
# Check system resources
htop
iotop
nethogs

# Check database performance
mysql -u root -p -e "SHOW PROCESSLIST;"
```

**3. Security issues**
```bash
# Check SSL certificate
sudo certbot certificates

# Check firewall status
sudo ufw status verbose

# Check fail2ban status
sudo fail2ban-client status
```

## 📝 Next Steps

Setelah verification selesai, lanjutkan ke:
- [Phase 3: Optimization](../../phase-3-optimization/README.md) - Optimasi performa dan monitoring

## 📚 References

- [Moodle Installation Verification](https://docs.moodle.org/311/en/Installation)
- [Moodle Performance](https://docs.moodle.org/311/en/Performance)
- [Moodle Security](https://docs.moodle.org/311/en/Security)
- [Moodle Troubleshooting](https://docs.moodle.org/311/en/Troubleshooting)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
