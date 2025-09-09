# 🖥️ Server Preparation

## 📋 Overview

Dokumen ini menjelaskan langkah-langkah persiapan server untuk instalasi LMS Moodle. Server preparation adalah tahap fundamental yang menentukan stabilitas dan performa sistem LMS.

## 🎯 Objectives

- [ ] Memastikan server dalam kondisi optimal
- [ ] Update sistem operasi ke versi terbaru
- [ ] Konfigurasi network dan firewall
- [ ] Setup user dan permissions
- [ ] Persiapan storage dan backup

## 📋 Prerequisites

- Server Ubuntu 22.04 LTS (fresh installation)
- Root access atau sudo privileges
- Koneksi internet yang stabil
- Minimal 4GB RAM, 20GB storage
- IP address yang sudah dikonfigurasi

## 🔧 Step-by-Step Guide

### Step 1: System Update

```bash
# Update package list
sudo apt update

# Upgrade system packages
sudo apt upgrade -y

# Install essential tools
sudo apt install -y curl wget git vim nano htop tree unzip
```

### Step 2: Network Configuration

```bash
# Check network configuration
ip addr show
ip route show

# Configure static IP (if needed)
sudo nano /etc/netplan/00-installer-config.yaml
```

**Example netplan configuration:**
```yaml
network:
  version: 2
  ethernets:
    eth0:
      dhcp4: false
      addresses:
        - 192.168.88.14/24
      gateway4: 192.168.88.1
      nameservers:
        addresses:
          - 8.8.8.8
          - 8.8.4.4
```

```bash
# Apply network configuration
sudo netplan apply
```

### Step 3: Hostname Configuration

```bash
# Set hostname
sudo hostnamectl set-hostname lms-server

# Update /etc/hosts
sudo nano /etc/hosts
```

**Add to /etc/hosts:**
```
127.0.0.1 localhost
192.168.88.14 lms-server lms
```

### Step 4: Timezone Configuration

```bash
# Set timezone
sudo timedatectl set-timezone Asia/Jakarta

# Enable NTP
sudo timedatectl set-ntp true

# Verify timezone
timedatectl status
```

### Step 5: User Management

```bash
# Create moodle user
sudo useradd -m -s /bin/bash moodle
sudo usermod -aG www-data moodle

# Set password for moodle user
sudo passwd moodle

# Create moodle directory
sudo mkdir -p /var/www/moodle
sudo chown -R moodle:www-data /var/www/moodle
sudo chmod -R 755 /var/www/moodle
```

### Step 6: Storage Preparation

```bash
# Check disk space
df -h

# Create additional partitions if needed
sudo fdisk -l

# Mount additional storage (if available)
sudo mkdir -p /mnt/moodle-data
sudo chown moodle:www-data /mnt/moodle-data
```

### Step 7: Firewall Configuration

```bash
# Install UFW
sudo apt install -y ufw

# Configure firewall rules
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow SSH
sudo ufw allow ssh

# Allow HTTP and HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable firewall
sudo ufw enable

# Check firewall status
sudo ufw status verbose
```

### Step 8: System Optimization

```bash
# Configure swap (if needed)
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Make swap permanent
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# Optimize kernel parameters
sudo nano /etc/sysctl.conf
```

**Add to /etc/sysctl.conf:**
```
# Network optimizations
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 65536 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216

# File system optimizations
fs.file-max = 65536
vm.swappiness = 10
```

```bash
# Apply kernel parameters
sudo sysctl -p
```

## ✅ Verification

### System Check

```bash
# Check system information
uname -a
lsb_release -a
free -h
df -h

# Check network
ping -c 4 8.8.8.8
nslookup google.com

# Check firewall
sudo ufw status

# Check timezone
timedatectl status
```

### Expected Results

- ✅ System updated to latest packages
- ✅ Network connectivity working
- ✅ Hostname set to `lms-server`
- ✅ Timezone set to `Asia/Jakarta`
- ✅ User `moodle` created
- ✅ Firewall configured and enabled
- ✅ Storage prepared for Moodle

## 🚨 Troubleshooting

### Common Issues

**1. Network not working after netplan apply**
```bash
# Revert to DHCP temporarily
sudo nano /etc/netplan/00-installer-config.yaml
# Set dhcp4: true
sudo netplan apply
```

**2. Firewall blocking SSH**
```bash
# Allow SSH from specific IP
sudo ufw allow from 192.168.88.0/24 to any port 22
```

**3. Timezone not updating**
```bash
# Reconfigure timezone
sudo dpkg-reconfigure tzdata
```

## 📝 Next Steps

Setelah server preparation selesai, lanjutkan ke:
- [02-software-installation.md](02-software-installation.md) - Install Nginx, PHP, MariaDB, Redis

## 📚 References

- [Ubuntu Server Guide](https://ubuntu.com/server/docs)
- [Netplan Configuration](https://netplan.io/examples/)
- [UFW Firewall](https://help.ubuntu.com/community/UFW)

---

**Last Updated:** September 9, 2025  
**Version:** 1.0  
**Author:** jejakawan007
