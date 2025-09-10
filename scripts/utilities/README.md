# 🛠️ Utility Scripts

## Overview
Scripts utilitas untuk maintenance, backup, dan monitoring.

## Scripts

### backup-restore.sh
- Full backup functionality
- Incremental backup functionality
- Restore functionality
- Backup management
- Database backup
- Files backup
- Configuration backup

### maintenance-mode.sh
- Enable/disable maintenance mode
- Custom maintenance page
- Scheduled maintenance
- Maintenance message configuration
- Maintenance status check

### system-verification.sh
- System information check
- Service status verification
- Port status check
- Resource usage check
- Network configuration check
- Security status check

## Usage
```bash
# Backup system
./backup-restore.sh --backup-full

# Restore system
./backup-restore.sh --restore-full

# Enable maintenance mode
./maintenance-mode.sh --enable

# Disable maintenance mode
./maintenance-mode.sh --disable

# System verification
./system-verification.sh
```

## Dependencies
- Moodle installed
- Database configured
- Backup directory configured

## Output
- Backup files created
- Maintenance mode managed
- System status verified
