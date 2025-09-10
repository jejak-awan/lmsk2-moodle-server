# 🎓 Phase 2: Moodle Installation Scripts

## Overview
Scripts untuk instalasi Moodle dengan berbagai versi dan verifikasi.

## Scripts

### 01-moodle-3.11-lts-install.sh
- Download dan extract Moodle 3.11 LTS
- File permissions dan ownership setup
- Database configuration
- Web server configuration
- Moodle installation via CLI
- Post-installation configuration
- Initial settings setup

### 02-moodle-4.0-install.sh
- Download dan extract Moodle 4.0
- File permissions dan ownership setup
- Database configuration
- Web server configuration
- Moodle installation via CLI
- Redis session configuration
- OPcache configuration
- Post-installation configuration

### 03-moodle-verification.sh
- Moodle version verification
- Web interface accessibility test
- SSL certificate validation
- Database connectivity test
- File permissions verification
- Cron job testing
- PHP extensions check
- System resources monitoring
- Security audit
- Performance testing
- Functional testing

## Usage
```bash
# Install Moodle 3.11 LTS
./01-moodle-3.11-lts-install.sh

# Install Moodle 4.0
./02-moodle-4.0-install.sh

# Verify installation
./03-moodle-verification.sh

# Run with options
./03-moodle-verification.sh --security
./03-moodle-verification.sh --performance
```

## Dependencies
- Phase 1 scripts completed
- Database configured
- Web server configured
- SSL certificate installed

## Output
- Moodle installed dan accessible
- Database connectivity working
- All verifications passed
- Ready untuk optimization
