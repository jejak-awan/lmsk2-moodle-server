# 📋 Moodle 5.0 Requirements

## 📋 Overview

Dokumen ini menjelaskan requirements dan dependencies yang diperlukan untuk instalasi Moodle 5.0 (Future), versi yang membawa fitur-fitur cutting-edge dan arsitektur masa depan.

## 🎯 Objectives

- [ ] Memahami system requirements untuk Moodle 5.0
- [ ] Verifikasi PHP version dan extensions
- [ ] Konfigurasi database yang kompatibel
- [ ] Setup web server yang optimal
- [ ] Persiapan hardware yang memadai

## 📋 System Requirements

### 🖥️ Hardware Requirements

#### Minimum Requirements
- **CPU**: 2 GHz single core
- **RAM**: 2 GB
- **Storage**: 1 GB (kode) + 5 GB (konten)
- **Network**: 5 Mbps

#### Recommended Requirements
- **CPU**: 3 GHz quad-core
- **RAM**: 8 GB
- **Storage**: 20 GB NVMe SSD
- **Network**: 50 Mbps

#### Production Requirements (100+ users)
- **CPU**: 16 cores @ 3.5 GHz
- **RAM**: 32 GB
- **Storage**: 200 GB NVMe SSD
- **Network**: 500 Mbps

### 🐧 Operating System

| OS | Version | Status | Notes |
|----|---------|--------|-------|
| **Ubuntu** | 22.04 LTS, 24.04 LTS | ✅ Recommended | Primary support |
| **Debian** | 12 (Bookworm), 13 (Trixie) | ✅ Supported | Good alternative |
| **CentOS** | 9 | ✅ Supported | Enterprise option |
| **RHEL** | 9 | ✅ Supported | Enterprise option |

### 🌐 Web Server

| Server | Version | Status | Notes |
|--------|---------|--------|-------|
| **Apache** | 2.4+ | ✅ Recommended | Best compatibility |
| **Nginx** | 1.20+ | ✅ Supported | High performance |

### 🐘 PHP Requirements

#### PHP Version
- **Minimum**: PHP 8.1.0
- **Maximum**: PHP 8.3.x
- **Recommended**: PHP 8.2.x
- **Not Supported**: PHP 8.0, PHP 8.4+

#### Required PHP Extensions

```bash
# Core Extensions
php8.2-fpm php8.2-cli php8.2-common
php8.2-mysql php8.2-zip php8.2-gd
php8.2-mbstring php8.2-curl php8.2-xml
php8.2-bcmath php8.2-intl php8.2-soap
php8.2-ldap php8.2-imagick php8.2-redis
php8.2-openssl php8.2-json php8.2-dom
php8.2-fileinfo php8.2-iconv php8.2-simplexml
php8.2-tokenizer php8.2-xmlreader php8.2-xmlwriter
php8.2-exif php8.2-ftp php8.2-gettext
php8.2-sodium php8.2-hash php8.2-filter
```

#### PHP Configuration

```ini
# Memory and execution time
memory_limit = 1G
max_execution_time = 900
max_input_time = 900

# File uploads
upload_max_filesize = 500M
post_max_size = 500M
max_file_uploads = 20

# Session configuration
session.gc_maxlifetime = 1440
session.cookie_httponly = 1
session.use_strict_mode = 1

# OPcache configuration
opcache.enable = 1
opcache.memory_consumption = 512
opcache.interned_strings_buffer = 16
opcache.max_accelerated_files = 20000
opcache.revalidate_freq = 2
opcache.fast_shutdown = 1
opcache.validate_timestamps = 0

# Error reporting
display_errors = Off
log_errors = On
error_log = /var/log/php_errors.log
```

### 🗄️ Database Requirements

#### MySQL/MariaDB
- **MySQL**: 8.0+ (required)
- **MariaDB**: 10.8+ (required)
- **Character Set**: utf8mb4
- **Collation**: utf8mb4_unicode_ci

#### PostgreSQL
- **Version**: 13+ atau 14+

#### Microsoft SQL Server
- **Version**: 2019+ atau 2022+

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

# Additional for Moodle 5.0
nodejs npm
composer
```

## 🔍 Pre-Installation Verification

### System Check Script

```bash
#!/bin/bash
# Moodle 5.0 Requirements Check

echo "=== Moodle 5.0 Requirements Check ==="
echo "Date: $(date)"
echo ""

# Check PHP
echo "1. PHP Version:"
if command -v php &> /dev/null; then
    PHP_VERSION=$(php -v | head -n1 | cut -d' ' -f2)
    echo "   PHP Version: $PHP_VERSION"
    if [[ "$PHP_VERSION" == 8.1.* ]] || [[ "$PHP_VERSION" == 8.2.* ]] || [[ "$PHP_VERSION" == 8.3.* ]]; then
        echo "   ✓ PHP version supported"
    else
        echo "   ✗ PHP version not supported (requires 8.1-8.3)"
    fi
else
    echo "   ✗ PHP not installed"
fi
echo ""

# Check PHP Extensions
echo "2. PHP Extensions:"
REQUIRED_EXTENSIONS=("mysql" "gd" "curl" "xml" "mbstring" "zip" "intl" "soap" "ldap" "imagick" "redis" "openssl" "json" "dom" "fileinfo" "iconv" "simplexml" "tokenizer" "xmlreader" "xmlwriter" "exif" "ftp" "gettext" "sodium" "hash" "filter")
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
    if [[ "$MARIADB_VERSION" == 10.8.* ]] || [[ "$MARIADB_VERSION" == 10.9.* ]] || [[ "$MARIADB_VERSION" == 10.10.* ]]; then
        echo "   ✓ MariaDB version supported"
    else
        echo "   ✗ MariaDB version not supported (requires 10.8+)"
    fi
else
    echo "   ✗ No database server found"
fi
echo ""

# Check Node.js
echo "4. Node.js:"
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version)
    echo "   Node.js Version: $NODE_VERSION"
    echo "   ✓ Node.js available"
else
    echo "   ✗ Node.js not installed"
fi
echo ""

# Check Composer
echo "5. Composer:"
if command -v composer &> /dev/null; then
    COMPOSER_VERSION=$(composer --version | head -n1)
    echo "   Composer: $COMPOSER_VERSION"
    echo "   ✓ Composer available"
else
    echo "   ✗ Composer not installed"
fi
echo ""

echo "=== Requirements Check Complete ==="
```

## 🚨 Common Issues

### PHP Version Issues
```bash
# Check current PHP version
php -v

# If PHP 8.0 is installed, upgrade to 8.2
sudo apt install php8.2-fpm php8.2-cli
sudo update-alternatives --set php /usr/bin/php8.2
```

### Database Version Issues
```bash
# Check MariaDB version
mariadb --version

# If version is too old, upgrade to 10.8+
sudo apt update
sudo apt install mariadb-server-10.8
```

### Node.js Installation
```bash
# Install Node.js 18+
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt install -y nodejs
```

### Composer Installation
```bash
# Install Composer
curl -sS https://getcomposer.org/installer | php
sudo mv composer.phar /usr/local/bin/composer
sudo chmod +x /usr/local/bin/composer
```

## 📝 Next Steps

Setelah requirements terpenuhi, lanjutkan ke:
- [02-database-setup.md](02-database-setup.md) - Setup database untuk Moodle 5.0

## 📚 References

- [Moodle 5.0 Documentation](https://docs.moodle.org/500/en/Main_page)
- [Moodle System Requirements](https://docs.moodle.org/500/en/Installation)
- [PHP 8.2 Documentation](https://www.php.net/manual/en/migration82.php)
- [MariaDB 10.8 Documentation](https://mariadb.org/documentation/mariadb/10.8/)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
