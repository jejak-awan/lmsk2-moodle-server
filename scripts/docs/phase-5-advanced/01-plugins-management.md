# 🔌 Plugins Management for Moodle

## 📋 Overview

Dokumen ini menjelaskan management plugins untuk Moodle, termasuk instalasi, konfigurasi, update, dan maintenance plugins untuk enhanced functionality.

## 🎯 Objectives

- [ ] Setup plugin management system
- [ ] Install essential plugins
- [ ] Konfigurasi plugin settings
- [ ] Setup plugin updates dan maintenance
- [ ] Monitor plugin performance

## 🔧 Step-by-Step Guide

### Step 1: Essential Plugins Installation

```bash
# Create plugin installation script
sudo nano /usr/local/bin/install-essential-plugins.sh
```

**Essential Plugins Installation Script:**
```bash
#!/bin/bash

# Essential Moodle Plugins Installation Script
MOODLE_DIR="/var/www/moodle"
PLUGIN_DIR="$MOODLE_DIR/local"

echo "Installing essential Moodle plugins..."

# Create plugin directories
sudo mkdir -p $PLUGIN_DIR
sudo chown -R www-data:www-data $PLUGIN_DIR

# Install Local Mobile plugin
echo "Installing Local Mobile plugin..."
cd $PLUGIN_DIR
sudo -u www-data git clone https://github.com/moodlehq/moodle-local_mobile.git mobile

# Install Local Boost plugin
echo "Installing Local Boost plugin..."
sudo -u www-data git clone https://github.com/moodlehq/moodle-local_boost.git boost

# Install Local Code Checker plugin
echo "Installing Local Code Checker plugin..."
sudo -u www-data git clone https://github.com/moodlehq/moodle-local_codechecker.git codechecker

# Install Local Moodle Check plugin
echo "Installing Local Moodle Check plugin..."
sudo -u www-data git clone https://github.com/moodlehq/moodle-local_moodlecheck.git moodlecheck

# Set permissions
sudo chown -R www-data:www-data $PLUGIN_DIR
sudo chmod -R 755 $PLUGIN_DIR

echo "Essential plugins installation completed"
```

### Step 2: Third-Party Plugins Installation

```bash
# Create third-party plugins installation script
sudo nano /usr/local/bin/install-third-party-plugins.sh
```

**Third-Party Plugins Installation Script:**
```bash
#!/bin/bash

# Third-Party Moodle Plugins Installation Script
MOODLE_DIR="/var/www/moodle"
PLUGIN_DIR="$MOODLE_DIR/local"

echo "Installing third-party Moodle plugins..."

# Install Local Advanced Notifications plugin
echo "Installing Local Advanced Notifications plugin..."
cd $PLUGIN_DIR
sudo -u www-data git clone https://github.com/catalyst/moodle-local_advancednotifications.git advancednotifications

# Install Local LDAP plugin
echo "Installing Local LDAP plugin..."
sudo -u www-data git clone https://github.com/catalyst/moodle-local_ldap.git ldap

# Install Local OAuth2 plugin
echo "Installing Local OAuth2 plugin..."
sudo -u www-data git clone https://github.com/catalyst/moodle-local_oauth2.git oauth2

# Install Local Redis plugin
echo "Installing Local Redis plugin..."
sudo -u www-data git clone https://github.com/catalyst/moodle-local_redis.git redis

# Install Local Solr plugin
echo "Installing Local Solr plugin..."
sudo -u www-data git clone https://github.com/catalyst/moodle-local_solr.git solr

# Set permissions
sudo chown -R www-data:www-data $PLUGIN_DIR
sudo chmod -R 755 $PLUGIN_DIR

echo "Third-party plugins installation completed"
```

### Step 3: Plugin Configuration

```bash
# Create plugin configuration script
sudo nano /usr/local/bin/configure-plugins.sh
```

**Plugin Configuration Script:**
```bash
#!/bin/bash

# Plugin Configuration Script
MOODLE_DIR="/var/www/moodle"

echo "Configuring Moodle plugins..."

cd $MOODLE_DIR

# Configure Local Mobile plugin
echo "Configuring Local Mobile plugin..."
sudo -u www-data php admin/cli/cfg.php --name=local_mobile_enabled --set=1
sudo -u www-data php admin/cli/cfg.php --name=local_mobile_app_id --set="com.moodle.moodlemobile"
sudo -u www-data php admin/cli/cfg.php --name=local_mobile_app_secret --set="your_app_secret_here"

# Configure Local Boost plugin
echo "Configuring Local Boost plugin..."
sudo -u www-data php admin/cli/cfg.php --name=local_boost_enabled --set=1
sudo -u www-data php admin/cli/cfg.php --name=local_boost_cache_enabled --set=1
sudo -u www-data php admin/cli/cfg.php --name=local_boost_cache_ttl --set=3600

# Configure Local Redis plugin
echo "Configuring Local Redis plugin..."
sudo -u www-data php admin/cli/cfg.php --name=local_redis_enabled --set=1
sudo -u www-data php admin/cli/cfg.php --name=local_redis_host --set="127.0.0.1"
sudo -u www-data php admin/cli/cfg.php --name=local_redis_port --set=6379
sudo -u www-data php admin/cli/cfg.php --name=local_redis_database --set=1

# Configure Local OAuth2 plugin
echo "Configuring Local OAuth2 plugin..."
sudo -u www-data php admin/cli/cfg.php --name=local_oauth2_enabled --set=1
sudo -u www-data php admin/cli/cfg.php --name=local_oauth2_client_id --set="your_client_id"
sudo -u www-data php admin/cli/cfg.php --name=local_oauth2_client_secret --set="your_client_secret"

# Clear caches
echo "Clearing caches..."
sudo -u www-data php admin/cli/purge_caches.php

echo "Plugin configuration completed"
```

### Step 4: Plugin Management System

```bash
# Create plugin management script
sudo nano /usr/local/bin/plugin-manager.sh
```

**Plugin Management Script:**
```bash
#!/bin/bash

# Plugin Management Script
MOODLE_DIR="/var/www/moodle"
PLUGIN_DIR="$MOODLE_DIR/local"
LOG_FILE="/var/log/plugin-manager.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Plugin management started..." >> $LOG_FILE

# Function to install plugin
install_plugin() {
    local plugin_name=$1
    local plugin_url=$2
    
    echo "Installing plugin: $plugin_name" >> $LOG_FILE
    
    cd $PLUGIN_DIR
    if [ -d "$plugin_name" ]; then
        echo "Plugin $plugin_name already exists" >> $LOG_FILE
        return 1
    fi
    
    sudo -u www-data git clone $plugin_url $plugin_name
    if [ $? -eq 0 ]; then
        echo "Plugin $plugin_name installed successfully" >> $LOG_FILE
        sudo chown -R www-data:www-data $PLUGIN_DIR/$plugin_name
        sudo chmod -R 755 $PLUGIN_DIR/$plugin_name
        return 0
    else
        echo "Failed to install plugin $plugin_name" >> $LOG_FILE
        return 1
    fi
}

# Function to update plugin
update_plugin() {
    local plugin_name=$1
    
    echo "Updating plugin: $plugin_name" >> $LOG_FILE
    
    if [ -d "$PLUGIN_DIR/$plugin_name" ]; then
        cd $PLUGIN_DIR/$plugin_name
        sudo -u www-data git pull origin main
        if [ $? -eq 0 ]; then
            echo "Plugin $plugin_name updated successfully" >> $LOG_FILE
            sudo chown -R www-data:www-data $PLUGIN_DIR/$plugin_name
            sudo chmod -R 755 $PLUGIN_DIR/$plugin_name
            return 0
        else
            echo "Failed to update plugin $plugin_name" >> $LOG_FILE
            return 1
        fi
    else
        echo "Plugin $plugin_name not found" >> $LOG_FILE
        return 1
    fi
}

# Function to remove plugin
remove_plugin() {
    local plugin_name=$1
    
    echo "Removing plugin: $plugin_name" >> $LOG_FILE
    
    if [ -d "$PLUGIN_DIR/$plugin_name" ]; then
        sudo rm -rf $PLUGIN_DIR/$plugin_name
        echo "Plugin $plugin_name removed successfully" >> $LOG_FILE
        return 0
    else
        echo "Plugin $plugin_name not found" >> $LOG_FILE
        return 1
    fi
}

# Function to list plugins
list_plugins() {
    echo "Installed plugins:" >> $LOG_FILE
    ls -la $PLUGIN_DIR/ | grep "^d" | awk '{print $9}' >> $LOG_FILE
}

# Main menu
case "$1" in
    install)
        install_plugin "$2" "$3"
        ;;
    update)
        update_plugin "$2"
        ;;
    remove)
        remove_plugin "$2"
        ;;
    list)
        list_plugins
        ;;
    *)
        echo "Usage: $0 {install|update|remove|list} [plugin_name] [plugin_url]"
        exit 1
        ;;
esac

echo "[$DATE] Plugin management completed" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/plugin-manager.sh
```

### Step 5: Plugin Monitoring

```bash
# Create plugin monitoring script
sudo nano /usr/local/bin/plugin-monitor.sh
```

**Plugin Monitoring Script:**
```bash
#!/bin/bash

# Plugin Monitoring Script
MOODLE_DIR="/var/www/moodle"
PLUGIN_DIR="$MOODLE_DIR/local"
LOG_FILE="/var/log/plugin-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Plugin monitoring started..." >> $LOG_FILE

# Check plugin status
cd $MOODLE_DIR
PLUGIN_STATUS=$(sudo -u www-data php admin/cli/plugin_status.php)

echo "Plugin Status:" >> $LOG_FILE
echo "$PLUGIN_STATUS" >> $LOG_FILE

# Check for plugin updates
echo "Checking for plugin updates..." >> $LOG_FILE
cd $PLUGIN_DIR
for plugin in */; do
    if [ -d "$plugin/.git" ]; then
        cd "$plugin"
        PLUGIN_NAME=$(basename "$plugin")
        CURRENT_COMMIT=$(git rev-parse HEAD)
        git fetch origin
        LATEST_COMMIT=$(git rev-parse origin/main)
        
        if [ "$CURRENT_COMMIT" != "$LATEST_COMMIT" ]; then
            echo "Plugin $PLUGIN_NAME has updates available" >> $LOG_FILE
        else
            echo "Plugin $PLUGIN_NAME is up to date" >> $LOG_FILE
        fi
        cd ..
    fi
done

# Check plugin performance
echo "Checking plugin performance..." >> $LOG_FILE
PLUGIN_PERFORMANCE=$(sudo -u www-data php admin/cli/performance_report.php)
echo "$PLUGIN_PERFORMANCE" >> $LOG_FILE

echo "[$DATE] Plugin monitoring completed" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/plugin-monitor.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# Plugin monitoring daily at 6 AM
0 6 * * * /usr/local/bin/plugin-monitor.sh
```

### Step 6: Plugin Backup

```bash
# Create plugin backup script
sudo nano /usr/local/bin/plugin-backup.sh
```

**Plugin Backup Script:**
```bash
#!/bin/bash

# Plugin Backup Script
BACKUP_DIR="/backup/plugins"
DATE=$(date +%Y%m%d_%H%M%S)
MOODLE_DIR="/var/www/moodle"
PLUGIN_DIR="$MOODLE_DIR/local"

echo "Starting plugin backup at $(date)"

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup all plugins
tar -czf $BACKUP_DIR/plugins_$DATE.tar.gz -C $MOODLE_DIR local/

# Backup plugin configuration
cd $MOODLE_DIR
sudo -u www-data php admin/cli/cfg.php --name=local_* > $BACKUP_DIR/plugin_config_$DATE.txt

# Cleanup old backups (keep 30 days)
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete
find $BACKUP_DIR -name "*.txt" -mtime +30 -delete

echo "Plugin backup completed at $(date)"
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/plugin-backup.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# Plugin backup weekly on Sunday at 4 AM
0 4 * * 0 /usr/local/bin/plugin-backup.sh
```

## ✅ Verification

### Plugin Management Test

```bash
# Test plugin installation
sudo /usr/local/bin/install-essential-plugins.sh

# Test plugin configuration
sudo /usr/local/bin/configure-plugins.sh

# Test plugin management
sudo /usr/local/bin/plugin-manager.sh list

# Test plugin monitoring
sudo /usr/local/bin/plugin-monitor.sh

# Test plugin backup
sudo /usr/local/bin/plugin-backup.sh
```

### Expected Results

- ✅ Essential plugins installed
- ✅ Third-party plugins installed
- ✅ Plugin configuration applied
- ✅ Plugin management system working
- ✅ Plugin monitoring active
- ✅ Plugin backup system working

## 🚨 Troubleshooting

### Common Issues

**1. Plugin installation failed**
```bash
# Check permissions
ls -la /var/www/moodle/local/

# Check Git access
sudo -u www-data git clone https://github.com/moodlehq/moodle-local_mobile.git /tmp/test
```

**2. Plugin configuration error**
```bash
# Check Moodle configuration
cd /var/www/moodle
sudo -u www-data php admin/cli/cfg.php --name=local_*

# Check plugin status
sudo -u www-data php admin/cli/plugin_status.php
```

**3. Plugin performance issues**
```bash
# Check plugin performance
cd /var/www/moodle
sudo -u www-data php admin/cli/performance_report.php

# Check plugin logs
tail -f /var/log/plugin-monitor.log
```

## 📝 Next Steps

Setelah plugins management selesai, lanjutkan ke:
- [02-integrations.md](02-integrations.md) - Setup integrations

## 📚 References

- [Moodle Plugins](https://moodle.org/plugins/)
- [Moodle Plugin Development](https://docs.moodle.org/dev/Plugin_types)
- [Moodle Plugin Management](https://docs.moodle.org/400/en/Managing_plugins)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
