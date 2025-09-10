# 🚀 LMSK2-Moodle-Server Scripts - Quick Start Guide

## 📋 Overview

Script automation untuk LMSK2-Moodle-Server yang telah dibuat dan siap digunakan. Scripts ini mengotomatisasi seluruh proses instalasi, konfigurasi, dan maintenance sistem LMS Moodle.

## 🎯 Scripts yang Tersedia

### **1. Master Installer Script**
- **File**: `lmsk2-moodle-installer.sh`
- **Fungsi**: Master script untuk menjalankan seluruh proses instalasi
- **Status**: ✅ Ready

### **2. Phase 1 Scripts**
- **File**: `01-server-preparation.sh`
- **Fungsi**: Server preparation dan system setup
- **Status**: ✅ Ready

- **File**: `02-software-installation.sh`
- **Fungsi**: Install Nginx, PHP, MariaDB, Redis
- **Status**: ✅ Ready

### **3. Utility Scripts**
- **File**: `system-verification.sh`
- **Fungsi**: Verifikasi sistem dan health check
- **Status**: ✅ Ready

- **File**: `backup-restore.sh`
- **Fungsi**: Backup dan restore sistem
- **Status**: ✅ Ready

### **4. Configuration Templates**
- **File**: `config/installer.conf`
- **Fungsi**: Konfigurasi utama installer
- **Status**: ✅ Ready

- **File**: `config/phase1.conf`
- **Fungsi**: Konfigurasi Phase 1
- **Status**: ✅ Ready

## 🚀 Quick Start

### **1. Persiapan**
```bash
# Masuk ke direktori scripts
cd /opt/lmsk2-moodle-server/scripts

# Pastikan semua script executable
chmod +x *.sh

# Cek help untuk melihat opsi yang tersedia
./lmsk2-moodle-installer.sh --help
```

### **2. Test Mode (Dry Run)**
```bash
# Test Phase 1 tanpa perubahan
./lmsk2-moodle-installer.sh --phase=1 --dry-run --domain=lms.example.com

# Test dengan interactive mode
./lmsk2-moodle-installer.sh --phase=1 --interactive --dry-run
```

### **3. Instalasi Phase 1 (Server Preparation)**
```bash
# Jalankan Phase 1 dengan konfigurasi default
./lmsk2-moodle-installer.sh --phase=1 --domain=lms.example.com --email=admin@example.com

# Atau dengan interactive mode
./lmsk2-moodle-installer.sh --phase=1 --interactive
```

### **4. Instalasi Phase 2 (Software Installation)**
```bash
# Jalankan Phase 2
./lmsk2-moodle-installer.sh --phase=2 --domain=lms.example.com --email=admin@example.com
```

### **5. Verifikasi Sistem**
```bash
# Jalankan system verification
./system-verification.sh --full-check

# Atau jalankan manual
./system-verification.sh
```

### **6. Backup System**
```bash
# Buat full backup
./backup-restore.sh backup full

# Buat incremental backup
./backup-restore.sh backup incremental

# List backup yang tersedia
./backup-restore.sh list

# Cleanup backup lama (30 hari)
./backup-restore.sh cleanup 30
```

## 📊 Status Implementasi

### **✅ Completed (Ready to Use)**
- [x] Master installer script
- [x] Phase 1: Server preparation script
- [x] Phase 1: Software installation script
- [x] System verification script
- [x] Backup and restore script
- [x] Configuration templates
- [x] Documentation

### **📋 In Progress (Next Phase)**
- [ ] Phase 1: Security hardening script
- [ ] Phase 1: Basic configuration script
- [ ] Phase 2: Moodle installation scripts
- [ ] Phase 3: Optimization scripts
- [ ] Phase 4: Production deployment scripts
- [ ] Phase 5: Advanced features scripts

## 🔧 Configuration

### **Environment Variables**
```bash
# Set konfigurasi utama
export LMSK2_VERSION="3.11-lts"
export LMSK2_DOMAIN="lms.example.com"
export LMSK2_EMAIL="admin@example.com"
export LMSK2_DB_PASSWORD="strong_password_here"
export LMSK2_ADMIN_PASSWORD="admin_password_here"

# Set mode
export LMSK2_DEBUG="false"
export LMSK2_VERBOSE="false"
export LMSK2_DRY_RUN="false"
export LMSK2_INTERACTIVE="true"
```

### **Configuration Files**
```bash
# Edit konfigurasi utama
nano config/installer.conf

# Edit konfigurasi Phase 1
nano config/phase1.conf
```

## 📝 Usage Examples

### **Full Installation (All Phases)**
```bash
# Jalankan semua phase
./lmsk2-moodle-installer.sh --phase=all --version=3.11-lts --domain=lms.example.com --email=admin@example.com --interactive
```

### **Phase by Phase Installation**
```bash
# Phase 1: Server preparation
./lmsk2-moodle-installer.sh --phase=1 --domain=lms.example.com --email=admin@example.com

# Phase 2: Software installation
./lmsk2-moodle-installer.sh --phase=2 --domain=lms.example.com --email=admin@example.com

# Dan seterusnya...
```

### **Maintenance Operations**
```bash
# System verification
./system-verification.sh --full-check

# Backup system
./backup-restore.sh backup full --compress

# Restore system
./backup-restore.sh restore full_20240101_120000

# Cleanup old backups
./backup-restore.sh cleanup 7
```

## 🚨 Troubleshooting

### **Common Issues**
1. **Permission Denied**
   ```bash
   # Pastikan script executable
   chmod +x *.sh
   
   # Jalankan sebagai root
   sudo ./lmsk2-moodle-installer.sh --phase=1
   ```

2. **Configuration Issues**
   ```bash
   # Cek konfigurasi
   cat config/installer.conf
   
   # Edit konfigurasi
   nano config/installer.conf
   ```

3. **Log Files**
   ```bash
   # Cek log installer
   tail -f /var/log/lmsk2/installer.log
   
   # Cek log Phase 1
   tail -f /var/log/lmsk2/phase1.log
   ```

### **Debug Mode**
```bash
# Enable debug mode
./lmsk2-moodle-installer.sh --phase=1 --debug --verbose

# Dry run mode
./lmsk2-moodle-installer.sh --phase=1 --dry-run
```

## 📚 Documentation

### **Available Documentation**
- [README.md](README.md) - Dokumentasi lengkap
- [DEVELOPMENT-PLAN.md](DEVELOPMENT-PLAN.md) - Rencana pengembangan
- [IMPLEMENTATION-ROADMAP.md](IMPLEMENTATION-ROADMAP.md) - Roadmap implementasi
- [QUICK-START.md](QUICK-START.md) - Panduan cepat (file ini)

### **Configuration Files**
- [config/installer.conf](config/installer.conf) - Konfigurasi utama
- [config/phase1.conf](config/phase1.conf) - Konfigurasi Phase 1

## 🎯 Next Steps

### **Immediate Actions**
1. **Test Scripts**: Jalankan test mode untuk memastikan scripts berfungsi
2. **Configure**: Edit konfigurasi sesuai kebutuhan
3. **Deploy**: Jalankan instalasi Phase 1 untuk setup server

### **Future Development**
1. **Phase 2**: Moodle installation scripts
2. **Phase 3**: Performance optimization scripts
3. **Phase 4**: Production deployment scripts
4. **Phase 5**: Advanced features scripts

## 📞 Support

### **Getting Help**
1. Check troubleshooting section
2. Review log files
3. Check system requirements
4. Contact support team

### **Log Files Location**
- `/var/log/lmsk2/installer.log` - Main installer log
- `/var/log/lmsk2/phase1.log` - Phase 1 log
- `/var/log/lmsk2/backup-restore.log` - Backup/restore log
- `/var/log/lmsk2/verification-report-*.txt` - Verification reports

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
