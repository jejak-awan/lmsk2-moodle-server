# 🔥 Moodle 4.1 Installation Guide

## 📋 Overview

Moodle 4.1 (Enhanced) adalah versi yang membawa fitur-fitur enhanced dan performa yang lebih baik. Versi ini memberikan pengalaman pengguna yang lebih baik dengan fitur-fitur terbaru.

## 🎯 Objectives

- [ ] Install Moodle 4.1 dengan konfigurasi optimal
- [ ] Setup database yang kompatibel dan performant
- [ ] Konfigurasi web server untuk performa maksimal
- [ ] Verifikasi instalasi dan functionality
- [ ] Siap untuk production deployment

## 📚 Installation Steps

### 1. [Requirements](01-requirements.md)
- System requirements untuk Moodle 4.1
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
- Download dan extract Moodle 4.1
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
| **Total** | **5-7 hours** | - |

## 🔧 System Requirements

### Minimum Requirements
- **CPU**: 1.5 GHz single core
- **RAM**: 1 GB
- **Storage**: 500 MB (kode) + 2 GB (konten)
- **Network**: 2 Mbps

### Recommended Requirements
- **CPU**: 2.5 GHz dual-core
- **RAM**: 4 GB
- **Storage**: 10 GB SSD
- **Network**: 20 Mbps

### Production Requirements (100+ users)
- **CPU**: 8 cores @ 3.0 GHz
- **RAM**: 16 GB
- **Storage**: 100 GB SSD
- **Network**: 200 Mbps

## 🐘 PHP Requirements

- **Version**: PHP 8.0.0 - 8.2.x
- **Recommended**: PHP 8.1.x
- **Not Supported**: PHP 7.4, PHP 8.3+

## 🗄️ Database Requirements

- **MySQL**: 8.0+ (required)
- **MariaDB**: 10.6+ (required)
- **PostgreSQL**: 12+ atau 13+
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
- [ ] Moodle 4.1 installed
- [ ] Installation verified
- [ ] Performance tested
- [ ] Security audited
- [ ] Functional testing passed

## 🚀 Next Steps

Setelah Moodle 4.1 installation selesai, lanjutkan ke:
- [Phase 3: Optimization](../../phase-3-optimization/README.md)

## 📞 Support

Jika mengalami masalah selama instalasi, silakan:
1. Cek troubleshooting section di setiap dokumen
2. Review requirements dan dependencies
3. Hubungi tim support

## 📚 References

- [Moodle 4.1 Documentation](https://docs.moodle.org/401/en/Main_page)
- [Moodle Installation Guide](https://docs.moodle.org/401/en/Installation)
- [Moodle System Requirements](https://docs.moodle.org/401/en/Installation)
- [Moodle Troubleshooting](https://docs.moodle.org/401/en/Troubleshooting)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
