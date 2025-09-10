# 📋 Moodle 4.0 Requirements

## 📋 Overview

Dokumen ini menjelaskan requirements dan dependencies yang diperlukan untuk instalasi Moodle 4.0 (Next Generation), versi yang membawa UI/UX modern dan performa yang lebih baik.

## 🎯 Objectives

- [ ] Memahami system requirements untuk Moodle 4.0
- [ ] Verifikasi PHP version dan extensions
- [ ] Konfigurasi database yang kompatibel
- [ ] Setup web server yang optimal
- [ ] Persiapan hardware yang memadai

## 📋 System Requirements

### 🖥️ Hardware Requirements

#### Minimum Requirements
- **CPU**: 1.5 GHz single core
- **RAM**: 1 GB
- **Storage**: 500 MB (kode) + 2 GB (konten)
- **Network**: 2 Mbps

#### Recommended Requirements
- **CPU**: 2.5 GHz dual-core
- **RAM**: 4 GB
- **Storage**: 10 GB SSD
- **Network**: 20 Mbps

#### Production Requirements (100+ users)
- **CPU**: 8 cores @ 3.0 GHz
- **RAM**: 16 GB
- **Storage**: 100 GB SSD
- **Network**: 200 Mbps

### 🐧 Operating System

| OS | Version | Status | Notes |
|----|---------|--------|-------|
| **Ubuntu** | 20.04 LTS, 22.04 LTS | ✅ Recommended | Primary support |
| **Debian** | 11 (Bullseye), 12 (Bookworm) | ✅ Supported | Good alternative |
| **CentOS** | 8, 9 | ✅ Supported | Enterprise option |
| **RHEL** | 8, 9 | ✅ Supported | Enterprise option |

### 🌐 Web Server

| Server | Version | Status | Notes |
|--------|---------|--------|-------|
| **Apache** | 2.4+ | ✅ Recommended | Best compatibility |
| **Nginx** | 1.18+ | ✅ Supported | High performance |

### 🐘 PHP Requirements

#### PHP Version
- **Minimum**: PHP 8.0.0
- **Maximum**: PHP 8.2.x
- **Recommended**: PHP 8.1.x
- **Not Supported**: PHP 7.4, PHP 8.3+

#### Required PHP Extensions

```bash
# Core Extensions
php8.1-fpm php8.1-cli php8.1-common
php8.1-mysql php8.1-zip php8.1-gd
php8.1-mbstring php8.1-curl php8.1-xml
php8.1-bcmath php8.1-intl php8.1-soap
php8.1-ldap php8.1-imagick php8.1-redis
php8.1-openssl php8.1-json php8.1-dom
php8.1-fileinfo php8.1-iconv php8.1-simplexml
php8.1-tokenizer php8.1-xmlreader php8.1-xmlwriter
php8.1-exif php8.1-ftp php8.1-gettext
```

#### PHP Configuration

```ini
# Memory and execution time
memory_limit = 512M
max_execution_time = 600
max_input_time = 600

# File uploads
upload_max_filesize = 200M
post_max_size = 200M
max_file_uploads = 20

# Session configuration
session.gc_maxlifetime = 1440
session.cookie_httponly = 1
session.use_strict_mode = 1

# OPcache configuration
opcache.enable = 1
opcache.memory_consumption = 256
opcache.interned_strings_buffer = 8
opcache.max_accelerated_files = 10000
opcache.revalidate_freq = 2
opcache.fast_shutdown = 1

# Error reporting
display_errors = Off
log_errors = On
error_log = /var/log/php_errors.log
```

### 🗄️ Database Requirements

#### MySQL/MariaDB
- **MySQL**: 8.0+ (required)
- **MariaDB**: 10.6+ (required)
- **Character Set**: utf8mb4
- **Collation**: utf8mb4_unicode_ci

#### PostgreSQL
- **Version**: 12+ atau 13+

#### Microsoft SQL Server
- **Version**: 2017+ atau 2019+

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

## 🔍 Pre-Installation Verification

### System Check Script

```bash
#!/bin/bash
# Moodle 4.0 Requirements Check

echo "=== Moodle 4.0 Requirements Check ==="
echo "Date: $(date)"
echo ""

# Check PHP
echo "1. PHP Version:"
if command -v php &> /dev/null; then
    PHP_VERSION=$(php -v | head -n1 | cut -d' ' -f2)
    echo "   PHP Version: $PHP_VERSION"
    if [[ "$PHP_VERSION" == 8.0.* ]] || [[ "$PHP_VERSION" == 8.1.* ]] || [[ "$PHP_VERSION" == 8.2.* ]]; then
        echo "   ✓ PHP version supported"
    else
        echo "   ✗ PHP version not supported (requires 8.0-8.2)"
    fi
else
    echo "   ✗ PHP not installed"
fi
echo ""

# Check PHP Extensions
echo "2. PHP Extensions:"
REQUIRED_EXTENSIONS=("mysql" "gd" "curl" "xml" "mbstring" "zip" "intl" "soap" "ldap" "imagick" "redis" "openssl" "json" "dom" "fileinfo" "iconv" "simplexml" "tokenizer" "xmlreader" "xmlwriter" "exif" "ftp" "gettext")
for ext in "${REQUIRED_EXTENSIONS[@]}"; do
    if php -m | grep -q "^$ext$"; then
        echo "   ✓ $ext"
    else
        echo "   ✗ $ext - Missing"
    fi
done
echo ""

# Check Database
echo "3. Database:"
if command -v mysql &> /dev/null; then
    MYSQL_VERSION=$(mysql --version | cut -d' ' -f3 | cut -d',' -f1)
    echo "   MySQL Version: $MYSQL_VERSION"
    if [[ "$MYSQL_VERSION" == 8.0.* ]]; then
        echo "   ✓ MySQL version supported"
    else
        echo "   ✗ MySQL version not supported (requires 8.0+)"
    fi
elif command -v mariadb &> /dev/null; then
    MARIADB_VERSION=$(mariadb --version | cut -d' ' -f3 | cut -d',' -f1)
    echo "   MariaDB Version: $MARIADB_VERSION"
    if [[ "$MARIADB_VERSION" == 10.6.* ]] || [[ "$MARIADB_VERSION" == 10.7.* ]] || [[ "$MARIADB_VERSION" == 10.8.* ]]; then
        echo "   ✓ MariaDB version supported"
    else
        echo "   ✗ MariaDB version not supported (requires 10.6+)"
    fi
else
    echo "   ✗ No database server found"
fi
echo ""

echo "=== Requirements Check Complete ==="
```

## 🚨 Common Issues

### PHP Version Issues
```bash
# Check current PHP version
php -v

# If PHP 7.4 is installed, upgrade to 8.1
sudo apt install php8.1-fpm php8.1-cli
sudo update-alternatives --set php /usr/bin/php8.1
```

### Database Version Issues
```bash
# Check MariaDB version
mariadb --version

# If version is too old, upgrade to 10.6+
sudo apt update
sudo apt install mariadb-server-10.6
```

## 📝 Next Steps

Setelah requirements terpenuhi, lanjutkan ke:
- [02-database-setup.md](02-database-setup.md) - Setup database untuk Moodle 4.0

## 📚 References

- [Moodle 4.0 Documentation](https://docs.moodle.org/400/en/Main_page)
- [Moodle System Requirements](https://docs.moodle.org/400/en/Installation)
- [PHP 8.1 Documentation](https://www.php.net/manual/en/migration81.php)
- [MariaDB 10.6 Documentation](https://mariadb.org/documentation/mariadb/10.6/)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
