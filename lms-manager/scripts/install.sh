#!/bin/bash

# LMS Manager Installation Script
# Author: jejakawan007
# Company: K2NET
# Website: https://k2net.id

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/opt/lms-manager"
SERVICE_NAME="lms-manager"
SERVICE_USER="lms-manager"
SERVICE_GROUP="lms-manager"
CONFIG_FILE="$INSTALL_DIR/config/config.json"
SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"

# Functions
print_header() {
    echo -e "${BLUE}"
    echo "=========================================="
    echo "    LMS Manager - K2NET Management System"
    echo "=========================================="
    echo -e "${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root"
        exit 1
    fi
}

check_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
    else
        print_error "Cannot determine OS version"
        exit 1
    fi

    print_info "Detected OS: $OS $VER"
}

install_dependencies() {
    print_info "Installing dependencies..."

    if command -v apt-get &> /dev/null; then
        # Ubuntu/Debian
        apt-get update
        apt-get install -y wget curl git build-essential sqlite3
    elif command -v yum &> /dev/null; then
        # CentOS/RHEL
        yum update -y
        yum install -y wget curl git gcc sqlite
    elif command -v dnf &> /dev/null; then
        # Fedora
        dnf update -y
        dnf install -y wget curl git gcc sqlite
    else
        print_error "Unsupported package manager"
        exit 1
    fi

    print_success "Dependencies installed"
}

install_go() {
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_info "Go is already installed: $GO_VERSION"
        return
    fi

    print_info "Installing Go..."

    # Download Go
    GO_VERSION="1.21.5"
    GO_ARCH="linux-amd64"
    GO_TAR="go$GO_VERSION.$GO_ARCH.tar.gz"

    cd /tmp
    wget "https://golang.org/dl/$GO_TAR"
    tar -C /usr/local -xzf "$GO_TAR"
    rm "$GO_TAR"

    # Add Go to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin

    print_success "Go installed: $GO_VERSION"
}

create_user() {
    if id "$SERVICE_USER" &>/dev/null; then
        print_info "User $SERVICE_USER already exists"
    else
        print_info "Creating user $SERVICE_USER..."
        useradd -r -s /bin/false -d "$INSTALL_DIR" "$SERVICE_USER"
        print_success "User $SERVICE_USER created"
    fi
}

install_application() {
    print_info "Installing LMS Manager..."

    # Create installation directory
    mkdir -p "$INSTALL_DIR"
    cd "$INSTALL_DIR"

    # Copy application files
    cp -r /tmp/lms-manager/* .

    # Build application
    print_info "Building application..."
    go mod tidy
    go build -o lms-manager main.go

    # Set permissions
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
    chmod +x lms-manager

    print_success "Application installed"
}

create_config() {
    print_info "Creating configuration..."

    # Create config directory
    mkdir -p "$INSTALL_DIR/config"

    # Generate JWT secret
    JWT_SECRET=$(openssl rand -hex 32)

    # Create config file
    cat > "$CONFIG_FILE" << EOF
{
  "server": {
    "port": 8080,
    "host": "0.0.0.0",
    "debug": false
  },
  "moodle": {
    "path": "/var/www/moodle",
    "config_path": "/var/www/moodle/config.php",
    "data_path": "/var/www/moodledata"
  },
  "security": {
    "jwt_secret": "$JWT_SECRET",
    "session_timeout": 3600,
    "rate_limit": 100,
    "allowed_ips": ["127.0.0.1", "192.168.1.0/24", "10.0.0.0/8"]
  },
  "monitoring": {
    "update_interval": 30,
    "log_retention": 7,
    "alert_thresholds": {
      "cpu": 80.0,
      "memory": 85.0,
      "disk": 90.0
    }
  }
}
EOF

    chown "$SERVICE_USER:$SERVICE_GROUP" "$CONFIG_FILE"
    chmod 600 "$CONFIG_FILE"

    print_success "Configuration created"
}

create_systemd_service() {
    print_info "Creating systemd service..."

    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=LMS Manager - K2NET Management System
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_GROUP
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/lms-manager
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=lms-manager

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME"

    print_success "Systemd service created"
}

create_directories() {
    print_info "Creating directories..."

    mkdir -p "$INSTALL_DIR/data"
    mkdir -p "$INSTALL_DIR/logs"
    mkdir -p "$INSTALL_DIR/backups"

    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"

    print_success "Directories created"
}

configure_firewall() {
    if command -v ufw &> /dev/null; then
        print_info "Configuring firewall..."
        ufw allow 8080/tcp
        print_success "Firewall configured"
    elif command -v firewall-cmd &> /dev/null; then
        print_info "Configuring firewall..."
        firewall-cmd --permanent --add-port=8080/tcp
        firewall-cmd --reload
        print_success "Firewall configured"
    else
        print_warning "No firewall detected, please configure manually"
    fi
}

start_service() {
    print_info "Starting service..."

    systemctl start "$SERVICE_NAME"
    sleep 3

    if systemctl is-active --quiet "$SERVICE_NAME"; then
        print_success "Service started successfully"
    else
        print_error "Failed to start service"
        systemctl status "$SERVICE_NAME"
        exit 1
    fi
}

show_status() {
    print_info "Service Status:"
    systemctl status "$SERVICE_NAME" --no-pager

    print_info "Service Logs (last 10 lines):"
    journalctl -u "$SERVICE_NAME" -n 10 --no-pager
}

show_instructions() {
    echo -e "${GREEN}"
    echo "=========================================="
    echo "    Installation Completed Successfully!"
    echo "=========================================="
    echo -e "${NC}"
    echo ""
    echo "LMS Manager has been installed and started."
    echo ""
    echo "Access Information:"
    echo "  URL: http://localhost:8080"
    echo "  Default Login: admin / admin123"
    echo ""
    echo "Service Management:"
    echo "  Start:   systemctl start $SERVICE_NAME"
    echo "  Stop:    systemctl stop $SERVICE_NAME"
    echo "  Restart: systemctl restart $SERVICE_NAME"
    echo "  Status:  systemctl status $SERVICE_NAME"
    echo "  Logs:    journalctl -u $SERVICE_NAME -f"
    echo ""
    echo "Configuration:"
    echo "  Config File: $CONFIG_FILE"
    echo "  Install Dir: $INSTALL_DIR"
    echo ""
    echo "Security Notes:"
    echo "  - Change the default admin password immediately"
    echo "  - Configure IP whitelist in config file"
    echo "  - Enable SSL/TLS for production use"
    echo ""
    echo "Support:"
    echo "  Website: https://k2net.id"
    echo "  Email: support@k2net.id"
    echo ""
}

cleanup() {
    print_info "Cleaning up temporary files..."
    rm -rf /tmp/lms-manager
    print_success "Cleanup completed"
}

# Main installation process
main() {
    print_header

    check_root
    check_os
    install_dependencies
    install_go
    create_user
    install_application
    create_config
    create_systemd_service
    create_directories
    configure_firewall
    start_service
    show_status
    show_instructions
    cleanup

    print_success "Installation completed successfully!"
}

# Run main function
main "$@"
