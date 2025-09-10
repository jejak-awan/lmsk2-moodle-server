#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: Phase 4 - SSL Certificate Script
# =============================================================================
# Description: SSL/TLS certificate management for LMSK2-Moodle-Server
# Version: 1.0
# Author: jejakawan007
# Date: September 9, 2025
# =============================================================================

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Script configuration
SCRIPT_NAME="LMSK2-Moodle-Server SSL Certificate"
SCRIPT_VERSION="1.0"
LOG_FILE="/var/log/lmsk2-ssl-certificate.log"
CONFIG_DIR="/opt/lmsk2-moodle-server/scripts/config"
SSL_DIR="/etc/ssl/lmsk2"

# Load configuration
if [ -f "$CONFIG_DIR/ssl.conf" ]; then
    source "$CONFIG_DIR/ssl.conf"
else
    echo -e "${YELLOW}Warning: SSL configuration file not found. Using defaults.${NC}"
fi

# Default configuration
SSL_PROVIDER=${SSL_PROVIDER:-"letsencrypt"}
SSL_EMAIL=${SSL_EMAIL:-"admin@localhost"}
SSL_DOMAIN=${SSL_DOMAIN:-"localhost"}
SSL_RENEWAL=${SSL_RENEWAL:-"true"}
SSL_STRONG_CIPHERS=${SSL_STRONG_CIPHERS:-"true"}
SSL_HSTS=${SSL_HSTS:-"true"}
SSL_OCSP=${SSL_OCSP:-"true"}

# =============================================================================
# Utility Functions
# =============================================================================

# Logging function
log() {
    local level=$1
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case $level in
        "INFO")
            echo -e "${GREEN}[INFO]${NC} $message" | tee -a "$LOG_FILE"
            ;;
        "WARN")
            echo -e "${YELLOW}[WARN]${NC} $message" | tee -a "$LOG_FILE"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message" | tee -a "$LOG_FILE"
            ;;
        "DEBUG")
            if [ "$LOG_LEVEL" = "debug" ]; then
                echo -e "${BLUE}[DEBUG]${NC} $message" | tee -a "$LOG_FILE"
            fi
            ;;
    esac
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# Error handling
handle_error() {
    local exit_code=$1
    local error_message=$2
    
    if [ $exit_code -ne 0 ]; then
        log "ERROR" "$error_message"
        log "ERROR" "SSL certificate setup failed. Exit code: $exit_code"
        exit $exit_code
    fi
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log "ERROR" "This script must be run as root"
        exit 1
    fi
}

# =============================================================================
# SSL Configuration Setup
# =============================================================================

# Create SSL configuration
create_ssl_config() {
    log "INFO" "Creating SSL configuration..."
    
    cat > "$CONFIG_DIR/ssl.conf" << EOF
# LMSK2 SSL Configuration

# SSL Provider
SSL_PROVIDER=letsencrypt
SSL_EMAIL=admin@yourdomain.com
SSL_DOMAIN=yourdomain.com

# SSL Settings
SSL_RENEWAL=true
SSL_STRONG_CIPHERS=true
SSL_HSTS=true
SSL_OCSP=true
SSL_SESSION_CACHE=true
SSL_SESSION_TIMEOUT=1d

# Let's Encrypt Settings
LETSENCRYPT_STAGING=false
LETSENCRYPT_RSA_KEY_SIZE=4096
LETSENCRYPT_ECDSA_KEY_SIZE=384

# Certificate Paths
SSL_CERT_PATH=/etc/letsencrypt/live
SSL_KEY_PATH=/etc/letsencrypt/live
SSL_CHAIN_PATH=/etc/letsencrypt/live

# Nginx SSL Configuration
NGINX_SSL_CONFIG=/etc/nginx/snippets/ssl-lmsk2.conf
NGINX_SSL_PARAMS=/etc/nginx/snippets/ssl-params.conf

# Security Headers
SECURITY_HEADERS=true
CSP_ENABLE=true
HSTS_MAX_AGE=31536000
HSTS_INCLUDE_SUBDOMAINS=true
HSTS_PRELOAD=true

# OCSP Stapling
OCSP_STAPLING=true
OCSP_RESPONDER_TIMEOUT=5s
OCSP_RESPONDER_VERIFY=on

# SSL Monitoring
SSL_MONITORING=true
SSL_EXPIRY_ALERT_DAYS=30
SSL_RENEWAL_ALERT_DAYS=7
EOF

    log "INFO" "SSL configuration created"
}

# =============================================================================
# Let's Encrypt Setup
# =============================================================================

# Install and configure Certbot
install_certbot() {
    log "INFO" "Installing Certbot..."
    
    # Update package list
    apt-get update -qq
    
    # Install snapd if not installed
    if ! command -v snap >/dev/null 2>&1; then
        apt-get install -y snapd || handle_error $? "Failed to install snapd"
    fi
    
    # Install certbot via snap
    snap install core; snap refresh core
    snap install --classic certbot
    
    # Create symlink
    ln -sf /snap/bin/certbot /usr/bin/certbot
    
    # Verify installation
    certbot --version || handle_error $? "Certbot installation failed"
    
    log "INFO" "Certbot installed successfully"
}

# Setup Let's Encrypt certificate
setup_letsencrypt_certificate() {
    log "INFO" "Setting up Let's Encrypt certificate..."
    
    # Create webroot directory for ACME challenge
    mkdir -p /var/www/certbot
    
    # Configure Nginx for ACME challenge
    cat > /etc/nginx/sites-available/certbot-challenge << EOF
server {
    listen 80;
    server_name $SSL_DOMAIN;
    
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    location / {
        return 301 https://\$server_name\$request_uri;
    }
}
EOF

    # Enable the site
    ln -sf /etc/nginx/sites-available/certbot-challenge /etc/nginx/sites-enabled/
    nginx -t && systemctl reload nginx
    
    # Obtain certificate
    if [ "$LETSENCRYPT_STAGING" = "true" ]; then
        certbot certonly --webroot -w /var/www/certbot -d "$SSL_DOMAIN" --email "$SSL_EMAIL" --agree-tos --no-eff-email --staging
    else
        certbot certonly --webroot -w /var/www/certbot -d "$SSL_DOMAIN" --email "$SSL_EMAIL" --agree-tos --no-eff-email
    fi
    
    if [ $? -eq 0 ]; then
        log "INFO" "Let's Encrypt certificate obtained successfully"
    else
        log "ERROR" "Failed to obtain Let's Encrypt certificate"
        return 1
    fi
    
    # Setup automatic renewal
    if [ "$SSL_RENEWAL" = "true" ]; then
        setup_certificate_renewal
    fi
    
    log "INFO" "Let's Encrypt certificate setup completed"
}

# Setup certificate renewal
setup_certificate_renewal() {
    log "INFO" "Setting up certificate renewal..."
    
    # Create renewal script
    cat > /opt/lmsk2-moodle-server/scripts/utilities/ssl-renewal.sh << 'EOF'
#!/bin/bash

# SSL Certificate Renewal Script
CONFIG_FILE="/opt/lmsk2-moodle-server/scripts/config/ssl.conf"
LOG_FILE="/var/log/lmsk2-ssl-renewal.log"
ALERT_EMAIL="admin@localhost"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Function to log with timestamp
log_renewal() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Check certificate expiry
check_certificate_expiry() {
    local domain="$1"
    local cert_file="/etc/letsencrypt/live/$domain/fullchain.pem"
    
    if [ ! -f "$cert_file" ]; then
        log_renewal "ERROR: Certificate file not found: $cert_file"
        return 1
    fi
    
    local expiry_date=$(openssl x509 -enddate -noout -in "$cert_file" | cut -d= -f2)
    local expiry_epoch=$(date -d "$expiry_date" +%s)
    local current_epoch=$(date +%s)
    local days_until_expiry=$(( (expiry_epoch - current_epoch) / 86400 ))
    
    log_renewal "Certificate expires in $days_until_expiry days"
    
    if [ "$days_until_expiry" -lt "$SSL_RENEWAL_ALERT_DAYS" ]; then
        log_renewal "WARNING: Certificate expires soon: $days_until_expiry days"
        return 1
    fi
    
    return 0
}

# Renew certificate
renew_certificate() {
    local domain="$1"
    
    log_renewal "Attempting to renew certificate for domain: $domain"
    
    # Try to renew certificate
    certbot renew --cert-name "$domain" --quiet
    
    if [ $? -eq 0 ]; then
        log_renewal "Certificate renewed successfully for domain: $domain"
        
        # Reload Nginx
        systemctl reload nginx
        
        # Send success notification
        if [ -n "$ALERT_EMAIL" ]; then
            echo "SSL certificate renewed successfully for domain: $domain" | mail -s "SSL Certificate Renewed" "$ALERT_EMAIL"
        fi
        
        return 0
    else
        log_renewal "ERROR: Failed to renew certificate for domain: $domain"
        
        # Send failure notification
        if [ -n "$ALERT_EMAIL" ]; then
            echo "SSL certificate renewal failed for domain: $domain" | mail -s "SSL Certificate Renewal Failed" "$ALERT_EMAIL"
        fi
        
        return 1
    fi
}

# Main renewal function
main() {
    log_renewal "Starting SSL certificate renewal check"
    
    # Check if renewal is needed
    if ! check_certificate_expiry "$SSL_DOMAIN"; then
        # Attempt renewal
        if renew_certificate "$SSL_DOMAIN"; then
            log_renewal "SSL certificate renewal process completed successfully"
            exit 0
        else
            log_renewal "SSL certificate renewal process failed"
            exit 1
        fi
    else
        log_renewal "Certificate renewal not needed"
        exit 0
    fi
}

# Run main function
main "$@"
EOF

    chmod +x /opt/lmsk2-moodle-server/scripts/utilities/ssl-renewal.sh
    
    # Setup cron job for renewal
    (crontab -l 2>/dev/null; echo "0 3 * * * /opt/lmsk2-moodle-server/scripts/utilities/ssl-renewal.sh") | crontab -
    
    log "INFO" "Certificate renewal setup completed"
}

# =============================================================================
# SSL Configuration for Nginx
# =============================================================================

# Create SSL configuration for Nginx
create_nginx_ssl_config() {
    log "INFO" "Creating Nginx SSL configuration..."
    
    # Create SSL configuration snippet
    cat > /etc/nginx/snippets/ssl-lmsk2.conf << EOF
# LMSK2 SSL Configuration

# SSL Certificate
ssl_certificate /etc/letsencrypt/live/$SSL_DOMAIN/fullchain.pem;
ssl_certificate_key /etc/letsencrypt/live/$SSL_DOMAIN/privkey.pem;

# SSL Session Configuration
ssl_session_timeout 1d;
ssl_session_cache shared:SSL:50m;
ssl_session_tickets off;

# SSL Protocols
ssl_protocols TLSv1.2 TLSv1.3;

# SSL Ciphers
ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384;
ssl_prefer_server_ciphers off;

# OCSP Stapling
ssl_stapling on;
ssl_stapling_verify on;
ssl_trusted_certificate /etc/letsencrypt/live/$SSL_DOMAIN/chain.pem;
resolver 8.8.8.8 8.8.4.4 valid=300s;
resolver_timeout 5s;
EOF

    # Create SSL parameters snippet
    cat > /etc/nginx/snippets/ssl-params.conf << EOF
# LMSK2 SSL Parameters

# Security Headers
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
add_header X-Frame-Options DENY always;
add_header X-Content-Type-Options nosniff always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;

# Content Security Policy
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none';" always;

# Additional Security Headers
add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;
add_header Cross-Origin-Embedder-Policy "require-corp" always;
add_header Cross-Origin-Opener-Policy "same-origin" always;
add_header Cross-Origin-Resource-Policy "same-origin" always;

# Hide Nginx version
server_tokens off;

# Buffer sizes
client_body_buffer_size 128k;
client_max_body_size 100m;
client_header_buffer_size 1k;
large_client_header_buffers 4 4k;
output_buffers 1 32k;
postpone_output 1460;

# Timeouts
client_header_timeout 3m;
client_body_timeout 3m;
send_timeout 3m;
EOF

    log "INFO" "Nginx SSL configuration created"
}

# Update Nginx configuration for SSL
update_nginx_ssl_config() {
    log "INFO" "Updating Nginx configuration for SSL..."
    
    # Create main SSL site configuration
    cat > /etc/nginx/sites-available/lmsk2-ssl << EOF
# LMSK2 SSL Site Configuration

# HTTP to HTTPS redirect
server {
    listen 80;
    server_name $SSL_DOMAIN;
    
    # Let's Encrypt challenge
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    # Redirect all other traffic to HTTPS
    location / {
        return 301 https://\$server_name\$request_uri;
    }
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name $SSL_DOMAIN;
    
    # SSL Configuration
    include /etc/nginx/snippets/ssl-lmsk2.conf;
    include /etc/nginx/snippets/ssl-params.conf;
    
    # Document root
    root /var/www/moodle;
    index index.php index.html index.htm;
    
    # Security
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
    
    location ~ /(config|install|lib|lang|locale|pix|theme|userpix|backup|temp|cache|sessions|trashdir|upgrade|admin|course|mod|blocks|filter|repository|user|files|search|tag|message|notification|badges|calendar|completion|plagiarism|report|stats|tool|webservice|auth|enrol|format|grade|local|qbehaviour|qtype|qformat|question|quiz|scorm|workshop|assignment|book|chat|choice|data|feedback|forum|glossary|hotpot|imscp|lesson|lti|page|resource|survey|url|wiki|workshop)/ {
        try_files \$uri \$uri/ /index.php?\$query_string;
    }
    
    # PHP processing
    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php8.1-fpm.sock;
        fastcgi_param SCRIPT_FILENAME \$document_root\$fastcgi_script_name;
        include fastcgi_params;
        
        # Security
        fastcgi_hide_header X-Powered-By;
        fastcgi_read_timeout 300;
        fastcgi_connect_timeout 300;
        fastcgi_send_timeout 300;
    }
    
    # Static files caching
    location ~* \.(css|js|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        add_header Vary Accept-Encoding;
        access_log off;
    }
    
    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;
}
EOF

    # Enable the SSL site
    ln -sf /etc/nginx/sites-available/lmsk2-ssl /etc/nginx/sites-enabled/
    
    # Test Nginx configuration
    nginx -t || handle_error $? "Nginx configuration test failed"
    
    # Reload Nginx
    systemctl reload nginx
    
    log "INFO" "Nginx SSL configuration updated"
}

# =============================================================================
# SSL Monitoring
# =============================================================================

# Setup SSL monitoring
setup_ssl_monitoring() {
    log "INFO" "Setting up SSL monitoring..."
    
    # Create SSL monitoring script
    cat > /opt/lmsk2-moodle-server/scripts/utilities/ssl-monitor.sh << 'EOF'
#!/bin/bash

# SSL Monitoring Script
CONFIG_FILE="/opt/lmsk2-moodle-server/scripts/config/ssl.conf"
LOG_FILE="/var/log/lmsk2-ssl-monitor.log"
ALERT_EMAIL="admin@localhost"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Function to log with timestamp
log_monitor() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Check certificate expiry
check_certificate_expiry() {
    local domain="$1"
    local cert_file="/etc/letsencrypt/live/$domain/fullchain.pem"
    
    if [ ! -f "$cert_file" ]; then
        log_monitor "ERROR: Certificate file not found: $cert_file"
        return 1
    fi
    
    local expiry_date=$(openssl x509 -enddate -noout -in "$cert_file" | cut -d= -f2)
    local expiry_epoch=$(date -d "$expiry_date" +%s)
    local current_epoch=$(date +%s)
    local days_until_expiry=$(( (expiry_epoch - current_epoch) / 86400 ))
    
    log_monitor "Certificate expires in $days_until_expiry days"
    
    if [ "$days_until_expiry" -lt "$SSL_EXPIRY_ALERT_DAYS" ]; then
        log_monitor "WARNING: Certificate expires soon: $days_until_expiry days"
        
        # Send alert
        if [ -n "$ALERT_EMAIL" ]; then
            echo "SSL certificate expires in $days_until_expiry days for domain: $domain" | mail -s "SSL Certificate Expiry Alert" "$ALERT_EMAIL"
        fi
        
        return 1
    fi
    
    return 0
}

# Check certificate chain
check_certificate_chain() {
    local domain="$1"
    local cert_file="/etc/letsencrypt/live/$domain/fullchain.pem"
    
    if [ ! -f "$cert_file" ]; then
        log_monitor "ERROR: Certificate file not found: $cert_file"
        return 1
    fi
    
    # Check certificate chain
    local chain_check=$(openssl verify -CAfile /etc/letsencrypt/live/$domain/chain.pem "$cert_file" 2>&1)
    
    if echo "$chain_check" | grep -q "OK"; then
        log_monitor "Certificate chain is valid"
        return 0
    else
        log_monitor "ERROR: Certificate chain validation failed: $chain_check"
        
        # Send alert
        if [ -n "$ALERT_EMAIL" ]; then
            echo "SSL certificate chain validation failed for domain: $domain" | mail -s "SSL Certificate Chain Error" "$ALERT_EMAIL"
        fi
        
        return 1
    fi
}

# Check SSL configuration
check_ssl_configuration() {
    local domain="$1"
    
    # Test SSL connection
    local ssl_test=$(echo | openssl s_client -servername "$domain" -connect "$domain:443" 2>/dev/null | openssl x509 -noout -text 2>/dev/null)
    
    if [ -n "$ssl_test" ]; then
        log_monitor "SSL connection test successful for domain: $domain"
        
        # Check SSL grade (simplified)
        local ssl_protocol=$(echo | openssl s_client -servername "$domain" -connect "$domain:443" 2>/dev/null | grep "Protocol" | awk '{print $3}')
        log_monitor "SSL Protocol: $ssl_protocol"
        
        return 0
    else
        log_monitor "ERROR: SSL connection test failed for domain: $domain"
        
        # Send alert
        if [ -n "$ALERT_EMAIL" ]; then
            echo "SSL connection test failed for domain: $domain" | mail -s "SSL Connection Error" "$ALERT_EMAIL"
        fi
        
        return 1
    fi
}

# Check OCSP stapling
check_ocsp_stapling() {
    local domain="$1"
    
    # Test OCSP stapling
    local ocsp_test=$(echo | openssl s_client -servername "$domain" -connect "$domain:443" -status 2>/dev/null | grep "OCSP Response Status")
    
    if [ -n "$ocsp_test" ]; then
        log_monitor "OCSP stapling is working for domain: $domain"
        return 0
    else
        log_monitor "WARNING: OCSP stapling is not working for domain: $domain"
        return 1
    fi
}

# Main monitoring function
main() {
    log_monitor "Starting SSL monitoring for domain: $SSL_DOMAIN"
    
    local monitoring_failed=0
    
    # Check certificate expiry
    if ! check_certificate_expiry "$SSL_DOMAIN"; then
        monitoring_failed=1
    fi
    
    # Check certificate chain
    if ! check_certificate_chain "$SSL_DOMAIN"; then
        monitoring_failed=1
    fi
    
    # Check SSL configuration
    if ! check_ssl_configuration "$SSL_DOMAIN"; then
        monitoring_failed=1
    fi
    
    # Check OCSP stapling
    if ! check_ocsp_stapling "$SSL_DOMAIN"; then
        monitoring_failed=1
    fi
    
    if [ "$monitoring_failed" -eq 0 ]; then
        log_monitor "SSL monitoring completed successfully"
        exit 0
    else
        log_monitor "SSL monitoring completed with issues"
        exit 1
    fi
}

# Run main function
main "$@"
EOF

    chmod +x /opt/lmsk2-moodle-server/scripts/utilities/ssl-monitor.sh
    
    # Setup cron job for SSL monitoring
    (crontab -l 2>/dev/null; echo "0 6 * * * /opt/lmsk2-moodle-server/scripts/utilities/ssl-monitor.sh") | crontab -
    
    log "INFO" "SSL monitoring setup completed"
}

# =============================================================================
# SSL Security Headers
# =============================================================================

# Setup additional security headers
setup_security_headers() {
    log "INFO" "Setting up additional security headers..."
    
    # Create security headers configuration
    cat > /etc/nginx/snippets/security-headers.conf << EOF
# LMSK2 Security Headers

# HSTS (HTTP Strict Transport Security)
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;

# X-Frame-Options
add_header X-Frame-Options "DENY" always;

# X-Content-Type-Options
add_header X-Content-Type-Options "nosniff" always;

# X-XSS-Protection
add_header X-XSS-Protection "1; mode=block" always;

# Referrer-Policy
add_header Referrer-Policy "strict-origin-when-cross-origin" always;

# Content Security Policy
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none';" always;

# Permissions Policy
add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;

# Cross-Origin Policies
add_header Cross-Origin-Embedder-Policy "require-corp" always;
add_header Cross-Origin-Opener-Policy "same-origin" always;
add_header Cross-Origin-Resource-Policy "same-origin" always;

# Server information hiding
server_tokens off;
more_clear_headers 'Server';
more_clear_headers 'X-Powered-By';
EOF

    log "INFO" "Security headers configuration created"
}

# =============================================================================
# Main Installation Function
# =============================================================================

# Main installation function
main() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  $SCRIPT_NAME v$SCRIPT_VERSION${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo
    
    log "INFO" "Starting SSL certificate setup..."
    
    # Check prerequisites
    check_root
    
    # Setup SSL
    create_ssl_config
    install_certbot
    setup_letsencrypt_certificate
    create_nginx_ssl_config
    update_nginx_ssl_config
    setup_ssl_monitoring
    setup_security_headers
    
    # Final verification
    log "INFO" "Verifying SSL setup..."
    
    # Check certificate files
    if [ -f "/etc/letsencrypt/live/$SSL_DOMAIN/fullchain.pem" ]; then
        log "INFO" "✓ SSL certificate files found"
    else
        log "ERROR" "✗ SSL certificate files not found"
    fi
    
    # Check Nginx configuration
    if nginx -t >/dev/null 2>&1; then
        log "INFO" "✓ Nginx configuration is valid"
    else
        log "ERROR" "✗ Nginx configuration is invalid"
    fi
    
    # Check SSL monitoring
    if [ -f "/opt/lmsk2-moodle-server/scripts/utilities/ssl-monitor.sh" ]; then
        log "INFO" "✓ SSL monitoring script created"
    else
        log "ERROR" "✗ SSL monitoring script not found"
    fi
    
    # Display summary
    echo
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  SSL Certificate Setup Completed!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo
    echo -e "${WHITE}SSL Components Installed:${NC}"
    echo -e "  • Let's Encrypt certificate"
    echo -e "  • Automatic certificate renewal"
    echo -e "  • SSL monitoring and alerts"
    echo -e "  • Security headers"
    echo -e "  • OCSP stapling"
    echo -e "  • Strong SSL ciphers"
    echo -e "  • HSTS configuration"
    echo
    echo -e "${WHITE}Configuration Files:${NC}"
    echo -e "  • $CONFIG_DIR/ssl.conf"
    echo -e "  • /etc/nginx/snippets/ssl-lmsk2.conf"
    echo -e "  • /etc/nginx/snippets/ssl-params.conf"
    echo -e "  • /etc/nginx/snippets/security-headers.conf"
    echo
    echo -e "${WHITE}Certificate Location:${NC}"
    echo -e "  • /etc/letsencrypt/live/$SSL_DOMAIN/"
    echo
    echo -e "${WHITE}Monitoring:${NC}"
    echo -e "  • SSL expiry monitoring"
    echo -e "  • Certificate chain validation"
    echo -e "  • OCSP stapling verification"
    echo -e "  • Automatic renewal"
    echo
    echo -e "${WHITE}Next Steps:${NC}"
    echo -e "  1. Update domain and email in ssl.conf"
    echo -e "  2. Test SSL configuration"
    echo -e "  3. Verify certificate renewal"
    echo -e "  4. Test security headers"
    echo -e "  5. Monitor SSL status"
    echo
    
    log "INFO" "SSL certificate setup completed successfully"
}

# =============================================================================
# Script Execution
# =============================================================================

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo
            echo "Options:"
            echo "  --help, -h          Show this help message"
            echo "  --version, -v       Show version information"
            echo "  --domain DOMAIN     Set SSL domain"
            echo "  --email EMAIL       Set SSL email"
            echo "  --staging           Use Let's Encrypt staging environment"
            echo "  --renewal           Enable/disable automatic renewal"
            echo "  --monitoring        Enable/disable SSL monitoring"
            echo
            exit 0
            ;;
        --version|-v)
            echo "$SCRIPT_NAME v$SCRIPT_VERSION"
            exit 0
            ;;
        --domain)
            SSL_DOMAIN="$2"
            shift 2
            ;;
        --email)
            SSL_EMAIL="$2"
            shift 2
            ;;
        --staging)
            LETSENCRYPT_STAGING="true"
            shift
            ;;
        --renewal)
            SSL_RENEWAL="$2"
            shift 2
            ;;
        --monitoring)
            SSL_MONITORING="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run main function
main "$@"
