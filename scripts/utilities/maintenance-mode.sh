#!/bin/bash

# =============================================================================
# LMSK2 Moodle Server - Maintenance Mode Script
# =============================================================================
# Description: Enable/disable maintenance mode for Moodle
# Author: jejakawan007
# Version: 1.0
# Date: September 2025
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Load configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/config/installer.conf"

if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
else
    echo "❌ Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Script configuration
SCRIPT_NAME="$(basename "$0")"
LOG_FILE="/var/log/lmsk2/${SCRIPT_NAME%.*}.log"
MOODLE_DIR="/var/www/moodle"
MAINTENANCE_MESSAGE="This site is currently being upgraded and is not available. Please try again later."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# =============================================================================
# Logging Functions
# =============================================================================

log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case "$level" in
        "INFO")  echo -e "${GREEN}[INFO]${NC} $message" ;;
        "WARN")  echo -e "${YELLOW}[WARN]${NC} $message" ;;
        "ERROR") echo -e "${RED}[ERROR]${NC} $message" ;;
        "DEBUG") echo -e "${BLUE}[DEBUG]${NC} $message" ;;
    esac
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

# =============================================================================
# Utility Functions
# =============================================================================

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log "ERROR" "This script must be run as root"
        exit 1
    fi
}

check_moodle_installed() {
    if [[ ! -d "$MOODLE_DIR" ]]; then
        log "ERROR" "Moodle directory not found: $MOODLE_DIR"
        exit 1
    fi
    
    if [[ ! -f "$MOODLE_DIR/config.php" ]]; then
        log "ERROR" "Moodle config.php not found"
        exit 1
    fi
}

# =============================================================================
# Maintenance Mode Functions
# =============================================================================

enable_maintenance_mode() {
    log "INFO" "Enabling maintenance mode..."
    
    cd "$MOODLE_DIR"
    
    # Enable maintenance mode via CLI
    if sudo -u www-data php admin/cli/maintenance.php --enable; then
        log "INFO" "Maintenance mode enabled successfully"
    else
        log "ERROR" "Failed to enable maintenance mode"
        return 1
    fi
    
    # Set maintenance message
    if sudo -u www-data php admin/cli/cfg.php --name=maintenance_message --set="$MAINTENANCE_MESSAGE"; then
        log "INFO" "Maintenance message set successfully"
    else
        log "WARN" "Failed to set maintenance message"
    fi
    
    # Create maintenance page
    create_maintenance_page
    
    log "INFO" "Maintenance mode enabled with custom message"
}

disable_maintenance_mode() {
    log "INFO" "Disabling maintenance mode..."
    
    cd "$MOODLE_DIR"
    
    # Disable maintenance mode via CLI
    if sudo -u www-data php admin/cli/maintenance.php --disable; then
        log "INFO" "Maintenance mode disabled successfully"
    else
        log "ERROR" "Failed to disable maintenance mode"
        return 1
    fi
    
    # Remove maintenance page
    remove_maintenance_page
    
    log "INFO" "Maintenance mode disabled"
}

create_maintenance_page() {
    log "INFO" "Creating custom maintenance page..."
    
    # Create maintenance page HTML
    cat > "$MOODLE_DIR/maintenance.html" << EOF
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Site Maintenance - ${SITE_FULLNAME:-LMS K2NET}</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            margin: 0;
            padding: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            color: white;
        }
        .maintenance-container {
            text-align: center;
            background: rgba(255, 255, 255, 0.1);
            padding: 3rem;
            border-radius: 20px;
            backdrop-filter: blur(10px);
            box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.37);
            border: 1px solid rgba(255, 255, 255, 0.18);
            max-width: 600px;
            margin: 2rem;
        }
        .maintenance-icon {
            font-size: 4rem;
            margin-bottom: 1rem;
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0% { transform: scale(1); }
            50% { transform: scale(1.1); }
            100% { transform: scale(1); }
        }
        h1 {
            font-size: 2.5rem;
            margin-bottom: 1rem;
            font-weight: 300;
        }
        p {
            font-size: 1.2rem;
            line-height: 1.6;
            margin-bottom: 2rem;
            opacity: 0.9;
        }
        .status {
            background: rgba(255, 255, 255, 0.2);
            padding: 1rem;
            border-radius: 10px;
            margin-top: 2rem;
        }
        .progress-bar {
            width: 100%;
            height: 4px;
            background: rgba(255, 255, 255, 0.3);
            border-radius: 2px;
            overflow: hidden;
            margin-top: 1rem;
        }
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #4CAF50, #8BC34A);
            width: 0%;
            animation: progress 3s ease-in-out infinite;
        }
        @keyframes progress {
            0% { width: 0%; }
            50% { width: 70%; }
            100% { width: 100%; }
        }
    </style>
</head>
<body>
    <div class="maintenance-container">
        <div class="maintenance-icon">🔧</div>
        <h1>Site Maintenance</h1>
        <p>$MAINTENANCE_MESSAGE</p>
        <div class="status">
            <p><strong>${SITE_FULLNAME:-LMS K2NET}</strong></p>
            <p>We're working hard to improve your experience</p>
            <div class="progress-bar">
                <div class="progress-fill"></div>
            </div>
        </div>
    </div>
</body>
</html>
EOF

    # Set proper permissions
    chown www-data:www-data "$MOODLE_DIR/maintenance.html"
    chmod 644 "$MOODLE_DIR/maintenance.html"
    
    log "INFO" "Custom maintenance page created"
}

remove_maintenance_page() {
    log "INFO" "Removing maintenance page..."
    
    if [[ -f "$MOODLE_DIR/maintenance.html" ]]; then
        rm "$MOODLE_DIR/maintenance.html"
        log "INFO" "Maintenance page removed"
    fi
}

check_maintenance_status() {
    log "INFO" "Checking maintenance mode status..."
    
    cd "$MOODLE_DIR"
    
    # Check maintenance mode status
    local maintenance_status=$(sudo -u www-data php admin/cli/cfg.php --name=maintenance_enabled --get 2>/dev/null || echo "0")
    
    if [[ "$maintenance_status" == "1" ]]; then
        log "INFO" "Maintenance mode is ENABLED"
        return 0
    else
        log "INFO" "Maintenance mode is DISABLED"
        return 1
    fi
}

show_maintenance_info() {
    log "INFO" "Maintenance mode information:"
    
    cd "$MOODLE_DIR"
    
    # Get maintenance settings
    local maintenance_enabled=$(sudo -u www-data php admin/cli/cfg.php --name=maintenance_enabled --get 2>/dev/null || echo "0")
    local maintenance_message=$(sudo -u www-data php admin/cli/cfg.php --name=maintenance_message --get 2>/dev/null || echo "Not set")
    
    echo "  Status: $([ "$maintenance_enabled" == "1" ] && echo "ENABLED" || echo "DISABLED")"
    echo "  Message: $maintenance_message"
    echo "  Site: ${SITE_FULLNAME:-LMS K2NET}"
    echo "  URL: https://${MOODLE_DOMAIN:-lms.example.com}"
}

# =============================================================================
# Scheduled Maintenance Functions
# =============================================================================

schedule_maintenance() {
    local start_time="$1"
    local duration="$2"
    local message="${3:-$MAINTENANCE_MESSAGE}"
    
    log "INFO" "Scheduling maintenance for $start_time (duration: $duration)"
    
    # Create maintenance script
    cat > /usr/local/bin/scheduled-maintenance.sh << EOF
#!/bin/bash

# Scheduled maintenance script
LOG_FILE="/var/log/lmsk2/scheduled-maintenance.log"
DATE=\$(date '+%Y-%m-%d %H:%M:%S')

echo "[\$DATE] Starting scheduled maintenance..." >> \$LOG_FILE

# Enable maintenance mode
cd $MOODLE_DIR
sudo -u www-data php admin/cli/maintenance.php --enable
sudo -u www-data php admin/cli/cfg.php --name=maintenance_message --set="$message"

echo "[\$DATE] Maintenance mode enabled" >> \$LOG_FILE

# Wait for specified duration
sleep $duration

# Disable maintenance mode
sudo -u www-data php admin/cli/maintenance.php --disable

echo "[\$DATE] Maintenance mode disabled" >> \$LOG_FILE
echo "[\$DATE] Scheduled maintenance completed" >> \$LOG_FILE
EOF

    # Make script executable
    chmod +x /usr/local/bin/scheduled-maintenance.sh
    
    # Schedule maintenance
    (crontab -l 2>/dev/null; echo "$start_time /usr/local/bin/scheduled-maintenance.sh") | crontab -
    
    log "INFO" "Maintenance scheduled for $start_time"
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    log "INFO" "Starting maintenance mode management..."
    
    # Check prerequisites
    check_root
    check_moodle_installed
    
    # Show current status
    show_maintenance_info
    
    log "INFO" "Maintenance mode management completed"
}

# =============================================================================
# Script Execution
# =============================================================================

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo "Options:"
        echo "  --help, -h           Show this help message"
        echo "  --enable             Enable maintenance mode"
        echo "  --disable            Disable maintenance mode"
        echo "  --status             Show maintenance mode status"
        echo "  --info               Show maintenance information"
        echo "  --schedule TIME DURATION [MESSAGE]  Schedule maintenance"
        echo "  --message MESSAGE    Set custom maintenance message"
        echo ""
        echo "Examples:"
        echo "  $0 --enable"
        echo "  $0 --disable"
        echo "  $0 --status"
        echo "  $0 --schedule '0 2 * * *' '3600' 'Scheduled maintenance'"
        echo "  $0 --message 'Custom maintenance message'"
        exit 0
        ;;
    --enable)
        check_root
        check_moodle_installed
        enable_maintenance_mode
        log "INFO" "Maintenance mode enabled"
        ;;
    --disable)
        check_root
        check_moodle_installed
        disable_maintenance_mode
        log "INFO" "Maintenance mode disabled"
        ;;
    --status)
        check_root
        check_moodle_installed
        if check_maintenance_status; then
            echo "Maintenance mode is ENABLED"
            exit 0
        else
            echo "Maintenance mode is DISABLED"
            exit 1
        fi
        ;;
    --info)
        check_root
        check_moodle_installed
        show_maintenance_info
        ;;
    --schedule)
        if [[ -z "${2:-}" || -z "${3:-}" ]]; then
            log "ERROR" "Usage: $0 --schedule TIME DURATION [MESSAGE]"
            exit 1
        fi
        check_root
        check_moodle_installed
        schedule_maintenance "$2" "$3" "${4:-$MAINTENANCE_MESSAGE}"
        ;;
    --message)
        if [[ -z "${2:-}" ]]; then
            log "ERROR" "Usage: $0 --message MESSAGE"
            exit 1
        fi
        check_root
        check_moodle_installed
        cd "$MOODLE_DIR"
        sudo -u www-data php admin/cli/cfg.php --name=maintenance_message --set="$2"
        log "INFO" "Maintenance message updated: $2"
        ;;
    "")
        main
        ;;
    *)
        log "ERROR" "Unknown option: $1"
        log "INFO" "Use --help for usage information"
        exit 1
        ;;
esac
