# 🔗 Integrations for Moodle

## 📋 Overview

Dokumen ini menjelaskan setup integrations untuk Moodle, termasuk LDAP, OAuth2, API integrations, dan third-party services.

## 🎯 Objectives

- [ ] Setup LDAP integration
- [ ] Konfigurasi OAuth2 authentication
- [ ] Setup API integrations
- [ ] Konfigurasi third-party services
- [ ] Monitor integration performance

## 🔧 Step-by-Step Guide

### Step 1: LDAP Integration

```bash
# Install LDAP packages
sudo apt install -y php8.1-ldap ldap-utils

# Configure LDAP
sudo nano /etc/ldap/ldap.conf
```

**LDAP Configuration:**
```
BASE dc=yourdomain,dc=com
URI ldap://ldap.yourdomain.com
TLS_REQCERT allow
```

### Step 2: OAuth2 Integration

```bash
# Create OAuth2 configuration script
sudo nano /usr/local/bin/configure-oauth2.sh
```

**OAuth2 Configuration Script:**
```bash
#!/bin/bash

# OAuth2 Configuration Script
MOODLE_DIR="/var/www/moodle"

echo "Configuring OAuth2 integration..."

cd $MOODLE_DIR

# Configure OAuth2 settings
sudo -u www-data php admin/cli/cfg.php --name=oauth2_enabled --set=1
sudo -u www-data php admin/cli/cfg.php --name=oauth2_client_id --set="your_client_id"
sudo -u www-data php admin/cli/cfg.php --name=oauth2_client_secret --set="your_client_secret"
sudo -u www-data php admin/cli/cfg.php --name=oauth2_redirect_uri --set="https://lms.yourdomain.com/auth/oauth2/callback.php"

echo "OAuth2 configuration completed"
```

### Step 3: API Integration

```bash
# Create API integration script
sudo nano /usr/local/bin/setup-api-integration.sh
```

**API Integration Script:**
```bash
#!/bin/bash

# API Integration Setup Script
MOODLE_DIR="/var/www/moodle"

echo "Setting up API integration..."

cd $MOODLE_DIR

# Enable web services
sudo -u www-data php admin/cli/cfg.php --name=enablewebservices --set=1
sudo -u www-data php admin/cli/cfg.php --name=enablemobilewebservice --set=1

# Configure API settings
sudo -u www-data php admin/cli/cfg.php --name=webserviceprotocols --set="rest,soap"
sudo -u www-data php admin/cli/cfg.php --name=webserviceallowfileupload --set=1

echo "API integration setup completed"
```

### Step 4: Third-Party Services

```bash
# Create third-party services script
sudo nano /usr/local/bin/setup-third-party-services.sh
```

**Third-Party Services Script:**
```bash
#!/bin/bash

# Third-Party Services Setup Script
MOODLE_DIR="/var/www/moodle"

echo "Setting up third-party services..."

cd $MOODLE_DIR

# Configure Google services
sudo -u www-data php admin/cli/cfg.php --name=google_client_id --set="your_google_client_id"
sudo -u www-data php admin/cli/cfg.php --name=google_client_secret --set="your_google_client_secret"

# Configure Microsoft services
sudo -u www-data php admin/cli/cfg.php --name=microsoft_client_id --set="your_microsoft_client_id"
sudo -u www-data php admin/cli/cfg.php --name=microsoft_client_secret --set="your_microsoft_client_secret"

# Configure Zoom integration
sudo -u www-data php admin/cli/cfg.php --name=zoom_api_key --set="your_zoom_api_key"
sudo -u www-data php admin/cli/cfg.php --name=zoom_api_secret --set="your_zoom_api_secret"

echo "Third-party services setup completed"
```

### Step 5: Integration Monitoring

```bash
# Create integration monitoring script
sudo nano /usr/local/bin/integration-monitor.sh
```

**Integration Monitoring Script:**
```bash
#!/bin/bash

# Integration Monitoring Script
LOG_FILE="/var/log/integration-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Integration monitoring..." >> $LOG_FILE

# Check LDAP connection
if ldapsearch -x -H ldap://ldap.yourdomain.com -b "dc=yourdomain,dc=com" -s base >/dev/null 2>&1; then
    echo "LDAP: Connected" >> $LOG_FILE
else
    echo "LDAP: Connection failed" >> $LOG_FILE
fi

# Check OAuth2 status
if curl -s "https://lms.yourdomain.com/auth/oauth2/test.php" | grep -q "success"; then
    echo "OAuth2: Working" >> $LOG_FILE
else
    echo "OAuth2: Not working" >> $LOG_FILE
fi

# Check API status
if curl -s "https://lms.yourdomain.com/webservice/rest/server.php" | grep -q "Moodle"; then
    echo "API: Working" >> $LOG_FILE
else
    echo "API: Not working" >> $LOG_FILE
fi

echo "---" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/integration-monitor.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# Integration monitoring every 30 minutes
*/30 * * * * /usr/local/bin/integration-monitor.sh
```

## ✅ Verification

### Integration Test

```bash
# Test LDAP connection
ldapsearch -x -H ldap://ldap.yourdomain.com -b "dc=yourdomain,dc=com" -s base

# Test OAuth2
curl -s "https://lms.yourdomain.com/auth/oauth2/test.php"

# Test API
curl -s "https://lms.yourdomain.com/webservice/rest/server.php"

# Test integrations
sudo /usr/local/bin/integration-monitor.sh
```

### Expected Results

- ✅ LDAP integration working
- ✅ OAuth2 authentication working
- ✅ API integration working
- ✅ Third-party services configured
- ✅ Integration monitoring active

## 🚨 Troubleshooting

### Common Issues

**1. LDAP connection failed**
```bash
# Check LDAP server
ldapsearch -x -H ldap://ldap.yourdomain.com -b "dc=yourdomain,dc=com" -s base

# Check LDAP configuration
sudo nano /etc/ldap/ldap.conf
```

**2. OAuth2 not working**
```bash
# Check OAuth2 configuration
cd /var/www/moodle
sudo -u www-data php admin/cli/cfg.php --name=oauth2_*

# Check OAuth2 logs
tail -f /var/log/oauth2.log
```

## 📝 Next Steps

Setelah integrations selesai, lanjutkan ke:
- [03-customizations.md](03-customizations.md) - Setup customizations

## 📚 References

- [Moodle LDAP](https://docs.moodle.org/400/en/LDAP_authentication)
- [Moodle OAuth2](https://docs.moodle.org/400/en/OAuth_2_authentication)
- [Moodle Web Services](https://docs.moodle.org/400/en/Web_services)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
