# 🚀 LMSK2-Moodle-Server Automation Scripts

## 📋 Overview

Script automation untuk LMSK2-Moodle-Server yang mengotomatisasi seluruh proses instalasi, konfigurasi, dan maintenance berdasarkan dokumentasi yang telah dibuat. Scripts ini dirancang untuk memudahkan deployment dan maintenance sistem LMS Moodle.

## 🎯 Objectives

- [ ] Otomatisasi instalasi server dan software
- [ ] Otomatisasi konfigurasi keamanan dan optimasi
- [ ] Otomatisasi instalasi dan konfigurasi Moodle
- [ ] Otomatisasi monitoring dan maintenance
- [ ] Otomatisasi backup dan disaster recovery

## 📚 Script Structure

### **Phase 1: Server Preparation Scripts**
- `01-server-preparation.sh` - Setup dasar server
- `02-software-installation.sh` - Install Nginx, PHP, MariaDB, Redis
- `03-security-hardening.sh` - Konfigurasi keamanan
- `04-basic-configuration.sh` - Konfigurasi dasar sistem

### **Phase 2: Moodle Installation Scripts**
- `01-moodle-3.11-lts-install.sh` - Install Moodle 3.11 LTS
- `02-moodle-4.0-install.sh` - Install Moodle 4.0 (future)
- `03-moodle-verification.sh` - Verifikasi instalasi

### **Phase 3: Optimization Scripts**
- `01-performance-tuning.sh` - Optimasi performa
- `02-caching-setup.sh` - Setup caching system
- `03-monitoring-setup.sh` - Setup monitoring
- `04-backup-strategy.sh` - Setup backup

### **Phase 4: Production Deployment Scripts**
- `01-production-setup.sh` - Setup production
- `02-ssl-certificate.sh` - Setup SSL
- `03-load-balancing.sh` - Setup load balancing
- `04-maintenance-procedures.sh` - Setup maintenance

### **Phase 5: Advanced Features Scripts**
- `01-plugins-management.sh` - Manage plugins
- `02-integrations.sh` - Setup integrations
- `03-customizations.sh` - Setup customizations
- `04-advanced-features.sh` - Setup advanced features

### **Utility Scripts**
- `system-verification.sh` - Verifikasi sistem
- `performance-benchmark.sh` - Benchmark performa
- `backup-restore.sh` - Backup dan restore
- `maintenance-mode.sh` - Mode maintenance

### **Monitoring Scripts**
- `system-health-check.sh` - Health check
- `performance-monitor.sh` - Monitor performa
- `security-monitor.sh` - Monitor keamanan
- `cache-monitor.sh` - Monitor cache

## 🚀 Master Installation Script

### **lmsk2-moodle-installer.sh**
Master script yang dapat menjalankan seluruh proses secara otomatis:

```bash
# Master installation script
./lmsk2-moodle-installer.sh --phase=all --version=3.11-lts

# Atau per fase
./lmsk2-moodle-installer.sh --phase=1  # Server preparation
./lmsk2-moodle-installer.sh --phase=2  # Moodle installation
./lmsk2-moodle-installer.sh --phase=3  # Optimization
./lmsk2-moodle-installer.sh --phase=4  # Production deployment
./lmsk2-moodle-installer.sh --phase=5  # Advanced features
```

## 📊 Script Features

### **Core Features**
1. **Interactive Mode** - User dapat memilih fase yang ingin dijalankan
2. **Dry Run Mode** - Test mode tanpa eksekusi
3. **Logging System** - Log detail setiap operasi
4. **Error Handling** - Rollback jika terjadi error
5. **Progress Indicator** - Progress bar untuk operasi panjang
6. **Configuration Validation** - Validasi konfigurasi sebelum eksekusi
7. **Backup Before Changes** - Backup otomatis sebelum perubahan
8. **Email Notifications** - Notifikasi via email
9. **Health Checks** - Verifikasi sistem setelah instalasi
10. **Performance Benchmarking** - Benchmark otomatis

### **Advanced Features**
- **Multi-version Support** - Support Moodle 3.11 LTS, 4.0, 4.1, 5.0
- **Environment Detection** - Auto-detect Ubuntu version dan hardware
- **Configuration Templates** - Template konfigurasi yang dapat disesuaikan
- **Rollback Capability** - Rollback ke state sebelumnya jika error
- **Parallel Execution** - Eksekusi parallel untuk operasi yang tidak bergantung
- **Resource Monitoring** - Monitor resource usage selama instalasi
- **Network Testing** - Test konektivitas dan performa network
- **Security Scanning** - Scan keamanan setelah instalasi

## 🎯 Development Priorities

### **High Priority (Phase 1)**
1. **Phase 1 Scripts** - Server preparation dan software installation
2. **System Verification Scripts** - Verifikasi sistem dan dependencies
3. **Backup/Restore Scripts** - Backup dan restore functionality
4. **Master Installer Script** - Master script untuk koordinasi

### **Medium Priority (Phase 2-3)**
1. **Phase 2 Scripts** - Moodle installation dan configuration
2. **Phase 3 Scripts** - Performance optimization dan caching
3. **Monitoring Scripts** - System health dan performance monitoring
4. **Performance Benchmarking** - Automated performance testing

### **Low Priority (Phase 4-5)**
1. **Phase 4 Scripts** - Production deployment dan SSL setup
2. **Phase 5 Scripts** - Advanced features dan integrations
3. **Advanced Monitoring** - Advanced monitoring dan alerting
4. **Customization Scripts** - Theme dan plugin customization

## 📋 Script Requirements

### **System Requirements**
- Ubuntu 22.04 LTS (primary support)
- Ubuntu 24.04 LTS (planned support)
- Debian 12 (planned support)
- Root access atau sudo privileges
- Internet connection yang stabil
- Minimal 4GB RAM, 20GB storage

### **Dependencies**
- `curl` - Download files
- `wget` - Download files
- `git` - Version control
- `jq` - JSON processing
- `bc` - Calculator
- `htop` - System monitoring
- `iotop` - I/O monitoring
- `nethogs` - Network monitoring

## 🔧 Configuration

### **Configuration Files**
- `config/installer.conf` - Main configuration
- `config/phase1.conf` - Phase 1 configuration
- `config/phase2.conf` - Phase 2 configuration
- `config/phase3.conf` - Phase 3 configuration
- `config/phase4.conf` - Phase 4 configuration
- `config/phase5.conf` - Phase 5 configuration

### **Environment Variables**
```bash
# Main configuration
export LMSK2_VERSION="3.11-lts"
export LMSK2_DOMAIN="lms.yourdomain.com"
export LMSK2_EMAIL="admin@yourdomain.com"
export LMSK2_DB_PASSWORD="strong_password_here"

# Phase specific
export LMSK2_PHASE1_ENABLE="true"
export LMSK2_PHASE2_ENABLE="true"
export LMSK2_PHASE3_ENABLE="true"
export LMSK2_PHASE4_ENABLE="false"
export LMSK2_PHASE5_ENABLE="false"
```

## 📝 Usage Examples

### **Full Installation**
```bash
# Download dan setup
git clone https://github.com/jejakawan007/lmsk2-moodle-server.git
cd lmsk2-moodle-server/scripts

# Make executable
chmod +x *.sh

# Run full installation
./lmsk2-moodle-installer.sh --phase=all --version=3.11-lts --domain=lms.yourdomain.com
```

### **Phase by Phase Installation**
```bash
# Phase 1: Server preparation
./lmsk2-moodle-installer.sh --phase=1 --interactive

# Phase 2: Moodle installation
./lmsk2-moodle-installer.sh --phase=2 --version=3.11-lts

# Phase 3: Optimization
./lmsk2-moodle-installer.sh --phase=3 --performance-mode=high
```

### **Maintenance Operations**
```bash
# System verification
./system-verification.sh --full-check

# Performance benchmark
./performance-benchmark.sh --output=report.html

# Backup system
./backup-restore.sh --backup --type=full

# Health check
./system-health-check.sh --alert-email=admin@yourdomain.com
```

## 🚨 Troubleshooting

### **Common Issues**
1. **Permission Denied** - Pastikan script executable dan user memiliki sudo access
2. **Network Issues** - Check internet connection dan firewall settings
3. **Disk Space** - Pastikan minimal 20GB free space
4. **Memory Issues** - Pastikan minimal 4GB RAM available
5. **Service Conflicts** - Check jika ada service yang conflict

### **Debug Mode**
```bash
# Enable debug mode
export LMSK2_DEBUG="true"
./lmsk2-moodle-installer.sh --phase=1 --debug

# Verbose logging
./lmsk2-moodle-installer.sh --phase=1 --verbose --log-level=debug
```

## 📚 Documentation

### **Script Documentation**
- [Phase 1 Scripts](docs/phase1-scripts.md)
- [Phase 2 Scripts](docs/phase2-scripts.md)
- [Phase 3 Scripts](docs/phase3-scripts.md)
- [Phase 4 Scripts](docs/phase4-scripts.md)
- [Phase 5 Scripts](docs/phase5-scripts.md)
- [Utility Scripts](docs/utility-scripts.md)
- [Monitoring Scripts](docs/monitoring-scripts.md)

### **Configuration Guide**
- [Configuration Guide](docs/configuration-guide.md)
- [Environment Setup](docs/environment-setup.md)
- [Troubleshooting Guide](docs/troubleshooting-guide.md)

## 🤝 Contributing

### **Development Guidelines**
1. Follow bash scripting best practices
2. Include error handling dan logging
3. Add configuration validation
4. Include progress indicators
5. Add rollback capabilities
6. Test on multiple Ubuntu versions
7. Document all functions dan parameters

### **Testing**
```bash
# Run tests
./tests/run-tests.sh

# Test specific phase
./tests/test-phase1.sh

# Integration tests
./tests/integration-tests.sh
```

## 📞 Support

### **Getting Help**
1. Check troubleshooting guide
2. Review log files
3. Check system requirements
4. Contact support team

### **Log Files**
- `/var/log/lmsk2-installer.log` - Main installer log
- `/var/log/lmsk2-phase1.log` - Phase 1 log
- `/var/log/lmsk2-phase2.log` - Phase 2 log
- `/var/log/lmsk2-phase3.log` - Phase 3 log
- `/var/log/lmsk2-phase4.log` - Phase 4 log
- `/var/log/lmsk2-phase5.log` - Phase 5 log

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
