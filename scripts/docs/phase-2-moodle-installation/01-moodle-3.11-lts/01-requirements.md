# 📋 Moodle 3.11 LTS Requirements

## 📋 Overview

Dokumen ini menjelaskan requirements dan dependencies yang diperlukan untuk instalasi Moodle 3.11 LTS (Long Term Support). Versi ini adalah versi stabil yang direkomendasikan untuk production environment.

## 🎯 Objectives

- [ ] Memahami system requirements untuk Moodle 3.11 LTS
- [ ] Verifikasi PHP version dan extensions
- [ ] Konfigurasi database yang kompatibel
- [ ] Setup web server yang optimal
- [ ] Persiapan hardware yang memadai

## 📋 System Requirements

### 🖥️ Hardware Requirements

#### Minimum Requirements
- **CPU**: 1 GHz single core
- **RAM**: 512 MB
- **Storage**: 200 MB (kode) + 1 GB (konten)
- **Network**: 1 Mbps

#### Recommended Requirements
- **CPU**: 2 GHz dual-core
- **RAM**: 2 GB
- **Storage**: 5 GB SSD
- **Network**: 10 Mbps

#### Production Requirements (100+ users)
- **CPU**: 4 cores @ 2.5 GHz
- **RAM**: 8 GB
- **Storage**: 50 GB SSD
- **Network**: 100 Mbps

### 🐧 Operating System

| OS | Version | Status | Notes |
|----|---------|--------|-------|
| **Ubuntu** | 20.04 LTS, 22.04 LTS | ✅ Recommended | Primary support |
| **Debian** | 10 (Buster), 11 (Bullseye) | ✅ Supported | Good alternative |
| **CentOS** | 8, 9 | ✅ Supported | Enterprise option |
| **RHEL** | 8, 9 | ✅ Supported | Enterprise option |

### 🌐 Web Server

| Server | Version | Status | Notes |
|--------|---------|--------|-------|
| **Apache** | 2.4+ | ✅ Recommended | Best compatibility |
| **Nginx** | 1.18+ | ✅ Supported | High performance |

### 🐘 PHP Requirements

#### PHP Version
- **Minimum**: PHP 7.4.0
- **Maximum**: PHP 8.1.x
- **Recommended**: PHP 8.0.x
- **Not Supported**: PHP 8.2+

#### Required PHP Extensions

```bash
# Core Extensions
php7.4-fpm php7.4-cli php7.4-common
php7.4-mysql php7.4-zip php7.4-gd
php7.4-mbstring php7.4-curl php7.4-xml
php7.4-bcmath php7.4-intl php7.4-soap
php7.4-ldap php7.4-imagick php7.4-xmlrpc
php7.4-openssl php7.4-json php7.4-dom
php7.4-fileinfo php7.4-iconv php7.4-simplexml
php7.4-tokenizer php7.4-xmlreader php7.4-xmlwriter
```

#### PHP Configuration

```ini
# Memory and execution time
memory_limit = 256M
max_execution_time = 300
max_input_time = 300

# File uploads
upload_max_filesize = 100M
post_max_size = 100M
max_file_uploads = 20

# Session configuration
session.gc_maxlifetime = 1440
session.cookie_httponly = 1
session.use_strict_mode = 1

# OPcache configuration
opcache.enable = 1
opcache.memory_consumption = 128
opcache.interned_strings_buffer = 8
opcache.max_accelerated_files = 5000
opcache.revalidate_freq = 2
opcache.fast_shutdown = 1

# Error reporting
display_errors = Off
log_errors = On
error_log = /var/log/php_errors.log
```

### 🗄️ Database Requirements

#### MySQL/MariaDB
- **MySQL**: 5.7.33+ atau 8.0+
- **MariaDB**: 10.3+ atau 10.6+
- **Character Set**: utf8mb4
- **Collation**: utf8mb4_unicode_ci

#### PostgreSQL
- **Version**: 10+ atau 12+
- **Character Set**: UTF8

#### Microsoft SQL Server
- **Version**: 2017+ atau 2019+

#### Oracle Database
- **Version**: 19c+

### 🔧 Additional Software

#### Required Tools
```bash
# Image processing
imagemagick ghostscript

# Document processing
unoconv libreoffice

# Compression
zip unzip gzip

# Network tools
curl wget

# Version control
git
```

#### Optional Tools
```bash
# Development tools
composer nodejs npm

# Monitoring
htop iotop nethogs

# Backup tools
rsync tar
```

## 🔍 Pre-Installation Verification

### System Check Script

```bash
#!/bin/bash
# Moodle 3.11 LTS Requirements Check

echo "=== Moodle 3.11 LTS Requirements Check ==="
echo "Date: $(date)"
echo ""

# Check OS
echo "1. Operating System:"
if [ -f /etc/os-release ]; then
    . /etc/os-release
    echo "   OS: $NAME $VERSION"
    if [[ "$NAME" == "Ubuntu" && "$VERSION_ID" == "22.04" ]]; then
        echo "   ✓ Ubuntu 22.04 LTS - Supported"
    elif [[ "$NAME" == "Ubuntu" && "$VERSION_ID" == "20.04" ]]; then
        echo "   ✓ Ubuntu 20.04 LTS - Supported"
    else
        echo "   ⚠ $NAME $VERSION - Check compatibility"
    fi
else
    echo "   ✗ Cannot determine OS"
fi
echo ""

# Check PHP
echo "2. PHP Version:"
if command -v php &> /dev/null; then
    PHP_VERSION=$(php -v | head -n1 | cut -d' ' -f2)
    echo "   PHP Version: $PHP_VERSION"
    if [[ "$PHP_VERSION" == 7.4.* ]] || [[ "$PHP_VERSION" == 8.0.* ]] || [[ "$PHP_VERSION" == 8.1.* ]]; then
        echo "   ✓ PHP version supported"
    else
        echo "   ✗ PHP version not supported"
    fi
else
    echo "   ✗ PHP not installed"
fi
echo ""

# Check PHP Extensions
echo "3. PHP Extensions:"
REQUIRED_EXTENSIONS=("mysql" "gd" "curl" "xml" "mbstring" "zip" "intl" "soap" "ldap" "imagick" "xmlrpc" "openssl" "json" "dom" "fileinfo" "iconv" "simplexml" "tokenizer" "xmlreader" "xmlwriter")
for ext in "${REQUIRED_EXTENSIONS[@]}"; do
    if php -m | grep -q "^$ext$"; then
        echo "   ✓ $ext"
    else
        echo "   ✗ $ext - Missing"
    fi
done
echo ""

# Check Database
echo "4. Database:"
if command -v mysql &> /dev/null; then
    MYSQL_VERSION=$(mysql --version | cut -d' ' -f3 | cut -d',' -f1)
    echo "   MySQL Version: $MYSQL_VERSION"
    if [[ "$MYSQL_VERSION" == 5.7.* ]] || [[ "$MYSQL_VERSION" == 8.0.* ]]; then
        echo "   ✓ MySQL version supported"
    else
        echo "   ⚠ MySQL version - Check compatibility"
    fi
elif command -v mariadb &> /dev/null; then
    MARIADB_VERSION=$(mariadb --version | cut -d' ' -f3 | cut -d',' -f1)
    echo "   MariaDB Version: $MARIADB_VERSION"
    if [[ "$MARIADB_VERSION" == 10.3.* ]] || [[ "$MARIADB_VERSION" == 10.6.* ]]; then
        echo "   ✓ MariaDB version supported"
    else
        echo "   ⚠ MariaDB version - Check compatibility"
    fi
else
    echo "   ✗ No database server found"
fi
echo ""

# Check Web Server
echo "5. Web Server:"
if command -v nginx &> /dev/null; then
    NGINX_VERSION=$(nginx -v 2>&1 | cut -d' ' -f3 | cut -d'/' -f2)
    echo "   Nginx Version: $NGINX_VERSION"
    if [[ "$NGINX_VERSION" == 1.18.* ]] || [[ "$NGINX_VERSION" == 1.20.* ]]; then
        echo "   ✓ Nginx version supported"
    else
        echo "   ⚠ Nginx version - Check compatibility"
    fi
elif command -v apache2 &> /dev/null; then
    APACHE_VERSION=$(apache2 -v | head -n1 | cut -d' ' -f3 | cut -d'/' -f2)
    echo "   Apache Version: $APACHE_VERSION"
    if [[ "$APACHE_VERSION" == 2.4.* ]]; then
        echo "   ✓ Apache version supported"
    else
        echo "   ⚠ Apache version - Check compatibility"
    fi
else
    echo "   ✗ No web server found"
fi
echo ""

# Check Disk Space
echo "6. Disk Space:"
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
DISK_AVAILABLE=$(df -h / | awk 'NR==2 {print $4}')
echo "   Available: $DISK_AVAILABLE"
echo "   Usage: $DISK_USAGE%"
if [ $DISK_USAGE -lt 80 ]; then
    echo "   ✓ Sufficient disk space"
else
    echo "   ⚠ Disk space low"
fi
echo ""

# Check Memory
echo "7. Memory:"
MEMORY_TOTAL=$(free -h | awk 'NR==2{print $2}')
MEMORY_AVAILABLE=$(free -h | awk 'NR==2{print $7}')
echo "   Total: $MEMORY_TOTAL"
echo "   Available: $MEMORY_AVAILABLE"
if [ $(free -m | awk 'NR==2{print $7}') -gt 1024 ]; then
    echo "   ✓ Sufficient memory"
else
    echo "   ⚠ Memory may be insufficient"
fi
echo ""

echo "=== Requirements Check Complete ==="
```

### Save and Run Check

```bash
# Save the script
sudo nano /usr/local/bin/moodle-requirements-check.sh
sudo chmod +x /usr/local/bin/moodle-requirements-check.sh

# Run the check
sudo /usr/local/bin/moodle-requirements-check.sh
```

## 🚨 Common Issues

### PHP Version Issues
```bash
# Check current PHP version
php -v

# If PHP 8.2+ is installed, downgrade to 8.1
sudo apt install php8.1-fpm php8.1-cli
sudo update-alternatives --set php /usr/bin/php8.1
```

### Missing PHP Extensions
```bash
# Install missing extensions
sudo apt install php8.1-mysql php8.1-gd php8.1-curl php8.1-xml php8.1-mbstring php8.1-zip php8.1-intl php8.1-soap php8.1-ldap php8.1-imagick
```

### Database Version Issues
```bash
# Check MariaDB version
mariadb --version

# If version is too old, upgrade
sudo apt update
sudo apt install mariadb-server-10.6
```

## 📝 Next Steps

Setelah requirements terpenuhi, lanjutkan ke:
- [02-database-setup.md](02-database-setup.md) - Setup database untuk Moodle 3.11 LTS

## 📚 References

- [Moodle 3.11 LTS Documentation](https://docs.moodle.org/311/en/Main_page)
- [Moodle System Requirements](https://docs.moodle.org/311/en/Installation)
- [PHP 8.1 Documentation](https://www.php.net/manual/en/migration81.php)
- [MariaDB 10.6 Documentation](https://mariadb.org/documentation/)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
