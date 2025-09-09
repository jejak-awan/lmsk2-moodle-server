# 🎨 Customizations for Moodle

## 📋 Overview

Dokumen ini menjelaskan customizations untuk Moodle, termasuk theme customization, branding, custom blocks, dan advanced configurations.

## 🎯 Objectives

- [ ] Setup theme customization
- [ ] Konfigurasi branding dan logo
- [ ] Setup custom blocks
- [ ] Konfigurasi advanced settings
- [ ] Monitor customization performance

## 🔧 Step-by-Step Guide

### Step 1: Theme Customization

```bash
# Create theme customization script
sudo nano /usr/local/bin/customize-theme.sh
```

**Theme Customization Script:**
```bash
#!/bin/bash

# Theme Customization Script
MOODLE_DIR="/var/www/moodle"
THEME_DIR="$MOODLE_DIR/theme"

echo "Customizing Moodle theme..."

# Create custom theme directory
sudo mkdir -p $THEME_DIR/custom
sudo chown -R www-data:www-data $THEME_DIR/custom

# Copy base theme
sudo -u www-data cp -r $THEME_DIR/boost $THEME_DIR/custom/k2net

# Customize theme files
sudo -u www-data nano $THEME_DIR/custom/k2net/config.php
```

**Custom Theme Configuration:**
```php
<?php
// Custom K2NET Theme Configuration
defined('MOODLE_INTERNAL') || die();

$THEME->name = 'k2net';
$THEME->sheets = ['custom'];
$THEME->editor_sheets = [];
$THEME->parents = ['boost'];
$THEME->enable_dock = false;
$THEME->yuicssmodules = array();
$THEME->rendererfactory = 'theme_overridden_renderer_factory';
$THEME->requiredblocks = '';
$THEME->addblockposition = BLOCK_ADDBLOCK_POSITION_FLATNAV;
$THEME->iconsystem = \core\output\icon_system::FONTAWESOME;
$THEME->haseditswitch = true;
$THEME->usefallback = true;
$THEME->scss = function($theme) {
    return theme_boost_get_main_scss_content($theme);
};
```

### Step 2: Branding Configuration

```bash
# Create branding configuration script
sudo nano /usr/local/bin/configure-branding.sh
```

**Branding Configuration Script:**
```bash
#!/bin/bash

# Branding Configuration Script
MOODLE_DIR="/var/www/moodle"

echo "Configuring Moodle branding..."

cd $MOODLE_DIR

# Configure site name and description
sudo -u www-data php admin/cli/cfg.php --name=fullname --set="LMS K2NET"
sudo -u www-data php admin/cli/cfg.php --name=shortname --set="LMSK2"
sudo -u www-data php admin/cli/cfg.php --name=summary --set="Learning Management System by K2NET"

# Configure logo and favicon
sudo -u www-data php admin/cli/cfg.php --name=logo --set="/theme/k2net/pix/logo.png"
sudo -u www-data php admin/cli/cfg.php --name=favicon --set="/theme/k2net/pix/favicon.ico"

# Configure footer
sudo -u www-data php admin/cli/cfg.php --name=footer --set="© 2025 K2NET. All rights reserved."

echo "Branding configuration completed"
```

### Step 3: Custom Blocks

```bash
# Create custom blocks script
sudo nano /usr/local/bin/create-custom-blocks.sh
```

**Custom Blocks Script:**
```bash
#!/bin/bash

# Custom Blocks Creation Script
MOODLE_DIR="/var/www/moodle"
BLOCK_DIR="$MOODLE_DIR/blocks"

echo "Creating custom blocks..."

# Create custom block directory
sudo mkdir -p $BLOCK_DIR/custom
sudo chown -R www-data:www-data $BLOCK_DIR/custom

# Create K2NET Info block
sudo -u www-data mkdir -p $BLOCK_DIR/custom/k2net_info
sudo -u www-data cat > $BLOCK_DIR/custom/k2net_info/version.php << 'EOF'
<?php
defined('MOODLE_INTERNAL') || die();

$plugin->version = 2025090900;
$plugin->requires = 2022112800;
$plugin->component = 'block_k2net_info';
$plugin->maturity = MATURITY_STABLE;
$plugin->release = '1.0';
EOF

sudo -u www-data cat > $BLOCK_DIR/custom/k2net_info/block_k2net_info.php << 'EOF'
<?php
defined('MOODLE_INTERNAL') || die();

class block_k2net_info extends block_base {
    public function init() {
        $this->title = get_string('k2net_info', 'block_k2net_info');
    }
    
    public function get_content() {
        if ($this->content !== null) {
            return $this->content;
        }
        
        $this->content = new stdClass;
        $this->content->text = '<div class="k2net-info">
            <h3>K2NET LMS</h3>
            <p>Powered by K2NET Technology</p>
            <p>Visit: <a href="https://k2net.id">k2net.id</a></p>
        </div>';
        
        return $this->content;
    }
}
EOF

echo "Custom blocks creation completed"
```

### Step 4: Advanced Settings

```bash
# Create advanced settings script
sudo nano /usr/local/bin/configure-advanced-settings.sh
```

**Advanced Settings Script:**
```bash
#!/bin/bash

# Advanced Settings Configuration Script
MOODLE_DIR="/var/www/moodle"

echo "Configuring advanced settings..."

cd $MOODLE_DIR

# Configure advanced features
sudo -u www-data php admin/cli/cfg.php --name=enablecompletion --set=1
sudo -u www-data php admin/cli/cfg.php --name=enablebadges --set=1
sudo -u www-data php admin/cli/cfg.php --name=enableportfolios --set=1

# Configure performance settings
sudo -u www-data php admin/cli/cfg.php --name=enableajax --set=1
sudo -u www-data php admin/cli/cfg.php --name=enablemobileapp --set=1
sudo -u www-data php admin/cli/cfg.php --name=enablemobilewebservice --set=1

# Configure security settings
sudo -u www-data php admin/cli/cfg.php --name=passwordpolicy --set=1
sudo -u www-data php admin/cli/cfg.php --name=passwordreuselimit --set=5
sudo -u www-data php admin/cli/cfg.php --name=passwordexpirytime --set=90

echo "Advanced settings configuration completed"
```

### Step 5: Customization Monitoring

```bash
# Create customization monitoring script
sudo nano /usr/local/bin/customization-monitor.sh
```

**Customization Monitoring Script:**
```bash
#!/bin/bash

# Customization Monitoring Script
LOG_FILE="/var/log/customization-monitor.log"
DATE=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$DATE] Customization monitoring..." >> $LOG_FILE

# Check theme status
cd /var/www/moodle
THEME_STATUS=$(sudo -u www-data php admin/cli/cfg.php --name=theme)
echo "Current theme: $THEME_STATUS" >> $LOG_FILE

# Check custom blocks
CUSTOM_BLOCKS=$(ls -la /var/www/moodle/blocks/custom/ | wc -l)
echo "Custom blocks: $CUSTOM_BLOCKS" >> $LOG_FILE

# Check branding settings
SITE_NAME=$(sudo -u www-data php admin/cli/cfg.php --name=fullname)
echo "Site name: $SITE_NAME" >> $LOG_FILE

# Check performance
PERFORMANCE=$(sudo -u www-data php admin/cli/performance_report.php | grep "Custom")
echo "Performance: $PERFORMANCE" >> $LOG_FILE

echo "---" >> $LOG_FILE
```

```bash
# Make script executable
sudo chmod +x /usr/local/bin/customization-monitor.sh

# Add to crontab
sudo crontab -e
```

**Add to crontab:**
```
# Customization monitoring daily at 7 AM
0 7 * * * /usr/local/bin/customization-monitor.sh
```

## ✅ Verification

### Customization Test

```bash
# Test theme customization
sudo /usr/local/bin/customize-theme.sh

# Test branding configuration
sudo /usr/local/bin/configure-branding.sh

# Test custom blocks
sudo /usr/local/bin/create-custom-blocks.sh

# Test advanced settings
sudo /usr/local/bin/configure-advanced-settings.sh

# Test customization monitoring
sudo /usr/local/bin/customization-monitor.sh
```

### Expected Results

- ✅ Theme customization working
- ✅ Branding configuration applied
- ✅ Custom blocks created
- ✅ Advanced settings configured
- ✅ Customization monitoring active

## 🚨 Troubleshooting

### Common Issues

**1. Theme not working**
```bash
# Check theme files
ls -la /var/www/moodle/theme/custom/

# Check theme configuration
cd /var/www/moodle
sudo -u www-data php admin/cli/cfg.php --name=theme
```

**2. Custom blocks not showing**
```bash
# Check block files
ls -la /var/www/moodle/blocks/custom/

# Check block configuration
cd /var/www/moodle
sudo -u www-data php admin/cli/cfg.php --name=block_*
```

## 📝 Next Steps

Setelah customizations selesai, lanjutkan ke:
- [04-advanced-features.md](04-advanced-features.md) - Setup advanced features

## 📚 References

- [Moodle Themes](https://docs.moodle.org/400/en/Themes)
- [Moodle Blocks](https://docs.moodle.org/400/en/Blocks)
- [Moodle Customization](https://docs.moodle.org/400/en/Customisation)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
