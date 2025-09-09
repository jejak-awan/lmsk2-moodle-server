# 🌟 Moodle 3.11 LTS Installation Guide

## 📋 Overview

Moodle 3.11 LTS (Long Term Support) adalah versi stabil yang direkomendasikan untuk production environment. Versi ini memberikan stabilitas maksimal dengan dukungan jangka panjang hingga November 2025.

## 🎯 Objectives

- [ ] Install Moodle 3.11 LTS dengan konfigurasi optimal
- [ ] Setup database yang kompatibel dan performant
- [ ] Konfigurasi web server untuk performa maksimal
- [ ] Verifikasi instalasi dan functionality
- [ ] Siap untuk production deployment

## 📚 Installation Steps

### 1. [Requirements](01-requirements.md)
- System requirements untuk Moodle 3.11 LTS
- PHP version dan extensions
- Database requirements
- Hardware recommendations
- Pre-installation verification

### 2. [Database Setup](02-database-setup.md)
- Install dan konfigurasi MariaDB/MySQL
- Buat database dan user untuk Moodle
- Konfigurasi database untuk performa optimal
- Setup backup dan recovery
- Verifikasi koneksi database

### 3. [Web Server Config](03-web-server-config.md)
- Konfigurasi Nginx virtual host
- Setup PHP-FPM integration
- Konfigurasi SSL/TLS certificates
- Optimasi performa dan caching
- Setup security headers

### 4. [Moodle Installation](04-moodle-installation.md)
- Download dan extract Moodle 3.11 LTS
- Setup file permissions dan ownership
- Konfigurasi database connection
- Jalankan Moodle installation wizard
- Konfigurasi initial settings

### 5. [Verification](05-verification.md)
- Verifikasi web interface accessibility
- Test database connectivity dan performance
- Verifikasi file permissions dan security
- Test cron job functionality
- Performance testing dan optimization

## ⏱️ Timeline

| Step | Duration | Dependencies |
|------|----------|--------------|
| Requirements | 30 minutes | Phase 1 complete |
| Database Setup | 1-2 hours | Requirements verified |
| Web Server Config | 1-2 hours | Database ready |
| Moodle Installation | 1-2 hours | Web server ready |
| Verification | 1 hour | Installation complete |
| **Total** | **4-6 hours** | - |

## 🔧 System Requirements

### Minimum Requirements
- **CPU**: 1 GHz single core
- **RAM**: 512 MB
- **Storage**: 200 MB (kode) + 1 GB (konten)
- **Network**: 1 Mbps

### Recommended Requirements
- **CPU**: 2 GHz dual-core
- **RAM**: 2 GB
- **Storage**: 5 GB SSD
- **Network**: 10 Mbps

### Production Requirements (100+ users)
- **CPU**: 4 cores @ 2.5 GHz
- **RAM**: 8 GB
- **Storage**: 50 GB SSD
- **Network**: 100 Mbps

## 🐘 PHP Requirements

- **Version**: PHP 7.4.0 - 8.1.x
- **Recommended**: PHP 8.0.x
- **Not Supported**: PHP 8.2+

### Required Extensions
```bash
php8.1-fpm php8.1-cli php8.1-common
php8.1-mysql php8.1-zip php8.1-gd
php8.1-mbstring php8.1-curl php8.1-xml
php8.1-bcmath php8.1-intl php8.1-soap
php8.1-ldap php8.1-imagick php8.1-xmlrpc
php8.1-openssl php8.1-json php8.1-dom
php8.1-fileinfo php8.1-iconv php8.1-simplexml
php8.1-tokenizer php8.1-xmlreader php8.1-xmlwriter
```

## 🗄️ Database Requirements

- **MySQL**: 5.7.33+ atau 8.0+
- **MariaDB**: 10.3+ atau 10.6+
- **PostgreSQL**: 10+ atau 12+
- **Character Set**: utf8mb4
- **Collation**: utf8mb4_unicode_ci

## 🌐 Web Server Requirements

- **Apache**: 2.4+ (recommended)
- **Nginx**: 1.18+ (high performance)
- **SSL/TLS**: Required for production

## ✅ Completion Checklist

- [ ] Requirements verified
- [ ] Database configured
- [ ] Web server configured
- [ ] Moodle 3.11 LTS installed
- [ ] Installation verified
- [ ] Performance tested
- [ ] Security audited
- [ ] Functional testing passed

## 🚀 Next Steps

Setelah Moodle 3.11 LTS installation selesai, lanjutkan ke:
- [Phase 3: Optimization](../../phase-3-optimization/README.md)

## 📞 Support

Jika mengalami masalah selama instalasi, silakan:
1. Cek troubleshooting section di setiap dokumen
2. Review requirements dan dependencies
3. Hubungi tim support

## 📚 References

- [Moodle 3.11 LTS Documentation](https://docs.moodle.org/311/en/Main_page)
- [Moodle Installation Guide](https://docs.moodle.org/311/en/Installation)
- [Moodle System Requirements](https://docs.moodle.org/311/en/Installation)
- [Moodle Troubleshooting](https://docs.moodle.org/311/en/Troubleshooting)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
