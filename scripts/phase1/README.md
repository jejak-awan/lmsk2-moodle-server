# 📋 Phase 1: Server Preparation Scripts

## Overview
Scripts untuk persiapan server dasar sebelum instalasi Moodle.

## Scripts

### 01-server-preparation.sh
- System update dan upgrade
- Network configuration
- Hostname dan timezone setup
- User management (moodle user)
- Storage preparation
- Basic firewall configuration
- System optimization

### 02-software-installation.sh
- Nginx installation dan configuration
- PHP 8.1 installation dengan extensions
- MariaDB installation dan secure setup
- Redis installation
- Additional tools (Composer, Node.js)
- Nginx virtual host untuk Moodle

### 03-security-hardening.sh
- Advanced firewall configuration (UFW)
- Fail2ban setup untuk brute force protection
- SSL/TLS certificate setup (Let's Encrypt)
- File permissions dan ownership security
- MariaDB security hardening
- PHP security configuration
- Log monitoring setup

### 04-basic-configuration.sh
- Kernel optimization untuk performance
- System limits configuration
- Cron jobs setup untuk maintenance
- System health monitoring
- Backup system configuration
- Log management setup
- Performance monitoring

## Usage
```bash
# Run individual script
./01-server-preparation.sh

# Run with options
./03-security-hardening.sh --verify
./04-basic-configuration.sh --dry-run
```

## Dependencies
- Ubuntu 22.04/24.04 LTS
- Root privileges
- Internet connection
- Domain name (untuk SSL)

## Output
- Configured server ready untuk Moodle installation
- Security hardening applied
- Performance optimization completed
- Monitoring systems active
