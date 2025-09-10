#!/bin/bash

# =============================================================================
# Phase 1: Server Preparation
# =============================================================================
# Version: 1.0
# Author: jejakawan007
# Description: Server preparation untuk LMSK2-Moodle-Server
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Script information
SCRIPT_NAME="Phase 1: Server Preparation"
SCRIPT_VERSION="1.0"
SCRIPT_AUTHOR="jejakawan007"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# =============================================================================
# Utility Functions
# =============================================================================

# Print colored output
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Print section header
print_section() {
    local section=$1
    echo
    print_color $YELLOW ">>> $section"
    print_color $YELLOW "=============================================================================="
}

# Print success message
print_success() {
    print_color $GREEN "✓ $1"
}

# Print error message
print_error() {
    print_color $RED "✗ $1"
}

# Print warning message
print_warning() {
    print_color $YELLOW "⚠ $1"
}

# Print info message
print_info() {
    print_color $BLUE "ℹ $1"
}

# Log message
log_message() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Create log directory if it doesn't exist
    mkdir -p "/var/log/lmsk2"
    
    # Write to log file
    echo "[$timestamp] [$level] $message" >> "/var/log/lmsk2/phase1.log"
    
    # Print to console based on log level
    case $level in
        "ERROR")
            print_error "$message"
            ;;
        "WARNING")
            print_warning "$message"
            ;;
        "INFO")
            print_info "$message"
            ;;
        "SUCCESS")
            print_success "$message"
            ;;
    esac
}

# =============================================================================
# Main Functions
# =============================================================================

# System update
system_update() {
    print_section "System Update"
    
    print_info "Updating package list..."
    apt update
    
    print_info "Upgrading system packages..."
    apt upgrade -y
    
    print_info "Installing essential tools..."
    apt install -y curl wget git vim nano htop tree unzip zip gzip tar \
        build-essential make cmake software-properties-common \
        net-tools dnsutils iputils-ping
    
    print_success "System update completed"
    log_message "SUCCESS" "System update completed"
}

# Network configuration
network_configuration() {
    print_section "Network Configuration"
    
    print_info "Checking network configuration..."
    
    # Display current network configuration
    print_info "Current network configuration:"
    ip addr show
    echo
    ip route show
    echo
    
    # Check if we need to configure static IP
    local current_ip=$(ip route get 8.8.8.8 | awk '{print $7; exit}')
    print_info "Current IP address: $current_ip"
    
    # Test internet connectivity
    if ping -c 4 8.8.8.8 >/dev/null 2>&1; then
        print_success "Internet connectivity: OK"
    else
        print_error "Internet connectivity: Failed"
        return 1
    fi
    
    # Test DNS resolution
    if nslookup google.com >/dev/null 2>&1; then
        print_success "DNS resolution: OK"
    else
        print_warning "DNS resolution: Failed"
    fi
    
    log_message "SUCCESS" "Network configuration checked"
}

# Hostname configuration
hostname_configuration() {
    print_section "Hostname Configuration"
    
    local current_hostname=$(hostname)
    print_info "Current hostname: $current_hostname"
    
    # Set hostname to lms-server if not already set
    if [[ "$current_hostname" != "lms-server" ]]; then
        print_info "Setting hostname to lms-server..."
        hostnamectl set-hostname lms-server
        
        # Update /etc/hosts
        print_info "Updating /etc/hosts..."
        local current_ip=$(ip route get 8.8.8.8 | awk '{print $7; exit}')
        
        # Backup original hosts file
        cp /etc/hosts /etc/hosts.backup.$(date +%Y%m%d_%H%M%S)
        
        # Update hosts file
        cat > /etc/hosts << EOF
127.0.0.1 localhost
$current_ip lms-server lms

# The following lines are desirable for IPv6 capable hosts
::1     localhost ip6-localhost ip6-loopback
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
EOF
        
        print_success "Hostname configured: lms-server"
    else
        print_success "Hostname already configured: lms-server"
    fi
    
    log_message "SUCCESS" "Hostname configuration completed"
}

# Timezone configuration
timezone_configuration() {
    print_section "Timezone Configuration"
    
    local current_timezone=$(timedatectl show --property=Timezone --value)
    print_info "Current timezone: $current_timezone"
    
    # Set timezone to Asia/Jakarta
    if [[ "$current_timezone" != "Asia/Jakarta" ]]; then
        print_info "Setting timezone to Asia/Jakarta..."
        timedatectl set-timezone Asia/Jakarta
        
        # Enable NTP
        print_info "Enabling NTP..."
        timedatectl set-ntp true
        
        print_success "Timezone configured: Asia/Jakarta"
    else
        print_success "Timezone already configured: Asia/Jakarta"
    fi
    
    # Verify timezone
    print_info "Timezone status:"
    timedatectl status
    
    log_message "SUCCESS" "Timezone configuration completed"
}

# User management
user_management() {
    print_section "User Management"
    
    # Create moodle user if it doesn't exist
    if ! id "moodle" &>/dev/null; then
        print_info "Creating moodle user..."
        useradd -m -s /bin/bash moodle
        usermod -aG www-data moodle
        
        # Set password for moodle user (if not in non-interactive mode)
        if [[ "${LMSK2_INTERACTIVE:-false}" == "true" ]]; then
            print_info "Setting password for moodle user..."
            passwd moodle
        else
            print_warning "Moodle user created without password. Set password manually if needed."
        fi
        
        print_success "Moodle user created"
    else
        print_success "Moodle user already exists"
    fi
    
    # Create moodle directory
    print_info "Creating moodle directory..."
    mkdir -p /var/www/moodle
    chown -R moodle:www-data /var/www/moodle
    chmod -R 755 /var/www/moodle
    
    print_success "Moodle directory created: /var/www/moodle"
    
    log_message "SUCCESS" "User management completed"
}

# Storage preparation
storage_preparation() {
    print_section "Storage Preparation"
    
    print_info "Checking disk space..."
    df -h
    
    # Check if we have enough space
    local available_space=$(df / | awk 'NR==2 {print $4}' | sed 's/G//')
    if [[ $available_space -lt 20 ]]; then
        print_warning "Available disk space: ${available_space}GB (Recommended: 20GB+)"
    else
        print_success "Available disk space: ${available_space}GB"
    fi
    
    # Create additional directories
    print_info "Creating additional directories..."
    mkdir -p /mnt/moodle-data
    chown moodle:www-data /mnt/moodle-data
    chmod 755 /mnt/moodle-data
    
    # Create backup directory
    mkdir -p /backup/moodle
    chown root:root /backup/moodle
    chmod 700 /backup/moodle
    
    print_success "Storage preparation completed"
    log_message "SUCCESS" "Storage preparation completed"
}

# Firewall configuration
firewall_configuration() {
    print_section "Firewall Configuration"
    
    # Install UFW if not installed
    if ! command -v ufw &> /dev/null; then
        print_info "Installing UFW firewall..."
        apt install -y ufw
    fi
    
    # Configure firewall rules
    print_info "Configuring firewall rules..."
    
    # Reset UFW to defaults
    ufw --force reset
    
    # Set default policies
    ufw default deny incoming
    ufw default allow outgoing
    
    # Allow SSH
    ufw allow ssh
    
    # Allow HTTP and HTTPS
    ufw allow 80/tcp
    ufw allow 443/tcp
    
    # Allow specific IP ranges (adjust as needed)
    local current_network=$(ip route | grep -E '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | head -1 | awk '{print $1}')
    if [[ -n "$current_network" ]]; then
        ufw allow from "$current_network" to any port 22
        ufw allow from "$current_network" to any port 80
        ufw allow from "$current_network" to any port 443
        print_info "Allowed access from network: $current_network"
    fi
    
    # Enable firewall
    ufw --force enable
    
    # Check firewall status
    print_info "Firewall status:"
    ufw status verbose
    
    print_success "Firewall configuration completed"
    log_message "SUCCESS" "Firewall configuration completed"
}

# System optimization
system_optimization() {
    print_section "System Optimization"
    
    # Configure swap if needed
    if [[ ! -f /swapfile ]]; then
        print_info "Creating swap file..."
        fallocate -l 2G /swapfile
        chmod 600 /swapfile
        mkswap /swapfile
        swapon /swapfile
        
        # Make swap permanent
        echo '/swapfile none swap sw 0 0' >> /etc/fstab
        
        print_success "Swap file created: 2GB"
    else
        print_success "Swap file already exists"
    fi
    
    # Optimize kernel parameters
    print_info "Optimizing kernel parameters..."
    
    # Create kernel optimization configuration
    cat > /etc/sysctl.d/99-moodle-optimization.conf << EOF
# Network optimizations
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.core.rmem_default = 262144
net.core.wmem_default = 262144
net.ipv4.tcp_rmem = 4096 65536 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
net.ipv4.tcp_congestion_control = bbr
net.ipv4.tcp_slow_start_after_idle = 0
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.tcp_keepalive_intvl = 30
net.ipv4.tcp_keepalive_probes = 3

# File system optimizations
fs.file-max = 65536
fs.inotify.max_user_watches = 524288
fs.inotify.max_user_instances = 256
fs.inotify.max_queued_events = 32768

# Memory management
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
vm.vfs_cache_pressure = 50

# Process limits
kernel.pid_max = 4194304
kernel.threads-max = 2097152
EOF
    
    # Apply kernel parameters
    sysctl -p /etc/sysctl.d/99-moodle-optimization.conf
    
    print_success "Kernel parameters optimized"
    
    # Configure system limits
    print_info "Configuring system limits..."
    
    cat > /etc/security/limits.d/99-moodle.conf << EOF
# Moodle user limits
moodle soft nofile 65536
moodle hard nofile 65536
moodle soft nproc 32768
moodle hard nproc 32768

# www-data user limits
www-data soft nofile 65536
www-data hard nofile 65536
www-data soft nproc 32768
www-data hard nproc 32768

# Root limits
root soft nofile 65536
root hard nofile 65536
root soft nproc 32768
root hard nproc 32768
EOF
    
    print_success "System limits configured"
    
    log_message "SUCCESS" "System optimization completed"
}

# Verification
verification() {
    print_section "Verification"
    
    print_info "System information:"
    echo "OS: $(lsb_release -d | cut -f2)"
    echo "Kernel: $(uname -r)"
    echo "Uptime: $(uptime -p)"
    echo
    
    print_info "Network configuration:"
    ip addr show | grep -E "(inet |UP)"
    echo
    
    print_info "Hostname: $(hostname)"
    print_info "Timezone: $(timedatectl show --property=Timezone --value)"
    echo
    
    print_info "Disk space:"
    df -h | grep -E "(Filesystem|/dev/)"
    echo
    
    print_info "Memory usage:"
    free -h
    echo
    
    print_info "Swap usage:"
    swapon --show
    echo
    
    print_info "Firewall status:"
    ufw status | head -5
    echo
    
    print_info "User information:"
    id moodle
    echo
    
    print_info "Directory permissions:"
    ls -la /var/www/moodle/
    echo
    
    print_success "Verification completed"
    log_message "SUCCESS" "Verification completed"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    print_color $CYAN "=============================================================================="
    print_color $WHITE "  $SCRIPT_NAME v$SCRIPT_VERSION"
    print_color $CYAN "=============================================================================="
    
    log_message "INFO" "Starting Phase 1: Server Preparation"
    
    # Execute all functions
    system_update
    network_configuration
    hostname_configuration
    timezone_configuration
    user_management
    storage_preparation
    firewall_configuration
    system_optimization
    verification
    
    print_section "Phase 1 Complete"
    print_success "Server preparation completed successfully!"
    print_info "Log file: /var/log/lmsk2/phase1.log"
    
    log_message "SUCCESS" "Phase 1: Server Preparation completed successfully"
}

# Trap errors
trap 'print_error "Script failed at line $LINENO"; log_message "ERROR" "Script failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
