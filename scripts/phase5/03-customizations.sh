#!/bin/bash

# =============================================================================
# LMSK2-Moodle-Server: Customizations Script
# =============================================================================
# Description: Advanced customizations for Moodle
# Version: 1.0
# Author: jejakawan007
# Date: September 9, 2025
# =============================================================================

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

# Script information
SCRIPT_NAME="LMSK2 Customizations"
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

# Paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="/var/log/lmsk2"
CONFIG_DIR="${SCRIPT_DIR}/../config"
MOODLE_DIR="/var/www/moodle"
MOODLE_DATA_DIR="/var/www/moodle/moodledata"

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
    print_color $CYAN "=============================================================================="
    print_color $WHITE "  $section"
    print_color $CYAN "=============================================================================="
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
    mkdir -p "$LOG_DIR"
    
    # Write to log file
    echo "[$timestamp] [$level] $message" >> "$LOG_DIR/customizations.log"
    
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
# Customization Functions
# =============================================================================

# Customize theme
customize_theme() {
    print_section "Customizing Theme"
    
    print_info "Installing custom theme..."
    
    # Create custom theme directory
    local custom_theme_dir="$MOODLE_DIR/theme/lmsk2"
    mkdir -p "$custom_theme_dir"
    
    # Create theme configuration
    cat > "$custom_theme_dir/config.php" << 'EOF'
<?php
// LMSK2 Custom Theme Configuration
defined('MOODLE_INTERNAL') || die();

$THEME->name = 'lmsk2';
$THEME->sheets = array('lmsk2');
$THEME->editor_sheets = array('lmsk2');
$THEME->parents = array('boost');
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
EOF

    # Create theme CSS
    cat > "$custom_theme_dir/style/lmsk2.css" << 'EOF'
/* LMSK2 Custom Theme Styles */

/* Brand colors */
:root {
    --primary-color: #2c3e50;
    --secondary-color: #3498db;
    --accent-color: #e74c3c;
    --success-color: #27ae60;
    --warning-color: #f39c12;
    --info-color: #17a2b8;
    --light-color: #ecf0f1;
    --dark-color: #2c3e50;
}

/* Header customization */
.navbar-brand {
    font-weight: bold;
    color: var(--primary-color) !important;
}

/* Navigation customization */
.navbar-nav .nav-link {
    color: var(--dark-color) !important;
    font-weight: 500;
}

.navbar-nav .nav-link:hover {
    color: var(--secondary-color) !important;
}

/* Button customization */
.btn-primary {
    background-color: var(--primary-color);
    border-color: var(--primary-color);
}

.btn-primary:hover {
    background-color: var(--secondary-color);
    border-color: var(--secondary-color);
}

/* Card customization */
.card {
    border: none;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    border-radius: 8px;
}

.card-header {
    background-color: var(--light-color);
    border-bottom: 1px solid #dee2e6;
}

/* Course customization */
.coursebox {
    border: 1px solid #dee2e6;
    border-radius: 8px;
    margin-bottom: 1rem;
    transition: box-shadow 0.3s ease;
}

.coursebox:hover {
    box-shadow: 0 4px 8px rgba(0,0,0,0.15);
}

/* Footer customization */
.footer {
    background-color: var(--dark-color);
    color: white;
    padding: 2rem 0;
    margin-top: 3rem;
}

/* Responsive design */
@media (max-width: 768px) {
    .navbar-brand {
        font-size: 1.2rem;
    }
    
    .card {
        margin-bottom: 1rem;
    }
}

/* Custom animations */
@keyframes fadeIn {
    from { opacity: 0; transform: translateY(20px); }
    to { opacity: 1; transform: translateY(0); }
}

.fade-in {
    animation: fadeIn 0.5s ease-in-out;
}

/* Loading spinner */
.spinner {
    border: 4px solid #f3f3f3;
    border-top: 4px solid var(--primary-color);
    border-radius: 50%;
    width: 40px;
    height: 40px;
    animation: spin 1s linear infinite;
    margin: 20px auto;
}

@keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
}
EOF

    # Create theme layout
    cat > "$custom_theme_dir/layout/columns2.php" << 'EOF'
<?php
// LMSK2 Custom Layout
defined('MOODLE_INTERNAL') || die();

echo $OUTPUT->doctype() ?>
<html <?php echo $OUTPUT->htmlattributes(); ?>>
<head>
    <title><?php echo $OUTPUT->page_title(); ?></title>
    <link rel="shortcut icon" href="<?php echo $OUTPUT->favicon(); ?>" />
    <?php echo $OUTPUT->standard_head_html() ?>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>

<body <?php echo $OUTPUT->body_attributes(); ?>>
<?php echo $OUTPUT->standard_top_of_body_html() ?>

<div id="page" class="container-fluid">
    <header id="page-header" class="row">
        <div class="col-12 p-a-1">
            <div class="card">
                <div class="card-block">
                    <?php echo $OUTPUT->context_header_settings_menu() ?>
                    <?php echo $OUTPUT->heading() ?>
                </div>
            </div>
        </div>
    </header>

    <div id="page-content" class="row">
        <div id="region-main-box" class="col-12">
            <section id="region-main" class="col-12">
                <?php echo $OUTPUT->course_content_header(); ?>
                <?php echo $OUTPUT->main_content(); ?>
                <?php echo $OUTPUT->course_content_footer(); ?>
            </section>
        </div>
    </div>

    <footer id="page-footer" class="footer">
        <div class="container">
            <div class="row">
                <div class="col-md-6">
                    <p>&copy; <?php echo date('Y'); ?> LMSK2 Moodle Server. All rights reserved.</p>
                </div>
                <div class="col-md-6 text-right">
                    <p>Powered by Moodle <?php echo $CFG->release; ?></p>
                </div>
            </div>
        </div>
    </footer>
</div>

<?php echo $OUTPUT->standard_end_of_body_html() ?>
</body>
</html>
EOF

    # Set proper permissions
    chown -R www-data:www-data "$custom_theme_dir"
    chmod -R 755 "$custom_theme_dir"
    
    # Enable custom theme
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=theme --set="lmsk2"
    
    print_success "Custom theme created and enabled"
    log_message "SUCCESS" "Theme customization completed"
}

# Create custom blocks
create_custom_blocks() {
    print_section "Creating Custom Blocks"
    
    # Create custom welcome block
    local welcome_block_dir="$MOODLE_DIR/blocks/lmsk2_welcome"
    mkdir -p "$welcome_block_dir"
    
    # Create block configuration
    cat > "$welcome_block_dir/version.php" << 'EOF'
<?php
defined('MOODLE_INTERNAL') || die();

$plugin->version   = 2025090900;
$plugin->requires  = 2022041900;
$plugin->component = 'block_lmsk2_welcome';
$plugin->maturity  = MATURITY_STABLE;
$plugin->release   = '1.0';
EOF

    cat > "$welcome_block_dir/block_lmsk2_welcome.php" << 'EOF'
<?php
defined('MOODLE_INTERNAL') || die();

class block_lmsk2_welcome extends block_base {
    
    public function init() {
        $this->title = get_string('welcome', 'block_lmsk2_welcome');
    }
    
    public function get_content() {
        global $USER, $OUTPUT;
        
        if ($this->content !== null) {
            return $this->content;
        }
        
        $this->content = new stdClass;
        $this->content->text = '';
        $this->content->footer = '';
        
        if (isloggedin()) {
            $this->content->text .= html_writer::tag('h3', get_string('welcome_user', 'block_lmsk2_welcome', fullname($USER)));
            $this->content->text .= html_writer::tag('p', get_string('welcome_message', 'block_lmsk2_welcome'));
            
            // Add user avatar
            $this->content->text .= $OUTPUT->user_picture($USER, array('size' => 100));
            
            // Add quick links
            $this->content->text .= html_writer::start_tag('ul');
            $this->content->text .= html_writer::tag('li', html_writer::link(new moodle_url('/my/'), get_string('mycourses')));
            $this->content->text .= html_writer::tag('li', html_writer::link(new moodle_url('/calendar/'), get_string('calendar')));
            $this->content->text .= html_writer::tag('li', html_writer::link(new moodle_url('/message/'), get_string('messages')));
            $this->content->text .= html_writer::end_tag('ul');
        } else {
            $this->content->text .= html_writer::tag('p', get_string('welcome_guest', 'block_lmsk2_welcome'));
            $this->content->text .= html_writer::link(new moodle_url('/login/'), get_string('login'), array('class' => 'btn btn-primary'));
        }
        
        return $this->content;
    }
    
    public function applicable_formats() {
        return array('all' => true);
    }
    
    public function has_config() {
        return true;
    }
}
EOF

    # Create language file
    mkdir -p "$welcome_block_dir/lang/en"
    cat > "$welcome_block_dir/lang/en/block_lmsk2_welcome.php" << 'EOF'
<?php
defined('MOODLE_INTERNAL') || die();

$string['welcome'] = 'Welcome';
$string['welcome_user'] = 'Welcome, {$a}!';
$string['welcome_message'] = 'Welcome to our learning management system. Explore your courses and start learning today!';
$string['welcome_guest'] = 'Welcome to our learning platform. Please log in to access your courses.';
$string['pluginname'] = 'LMSK2 Welcome Block';
EOF

    # Set proper permissions
    chown -R www-data:www-data "$welcome_block_dir"
    chmod -R 755 "$welcome_block_dir"
    
    print_success "Custom welcome block created"
    
    # Create custom statistics block
    local stats_block_dir="$MOODLE_DIR/blocks/lmsk2_stats"
    mkdir -p "$stats_block_dir"
    
    cat > "$stats_block_dir/version.php" << 'EOF'
<?php
defined('MOODLE_INTERNAL') || die();

$plugin->version   = 2025090900;
$plugin->requires  = 2022041900;
$plugin->component = 'block_lmsk2_stats';
$plugin->maturity  = MATURITY_STABLE;
$plugin->release   = '1.0';
EOF

    cat > "$stats_block_dir/block_lmsk2_stats.php" << 'EOF'
<?php
defined('MOODLE_INTERNAL') || die();

class block_lmsk2_stats extends block_base {
    
    public function init() {
        $this->title = get_string('statistics', 'block_lmsk2_stats');
    }
    
    public function get_content() {
        global $DB;
        
        if ($this->content !== null) {
            return $this->content;
        }
        
        $this->content = new stdClass;
        $this->content->text = '';
        $this->content->footer = '';
        
        // Get statistics
        $total_courses = $DB->count_records('course', array('id' => 1, 'visible' => 1));
        $total_users = $DB->count_records('user', array('deleted' => 0));
        $total_activities = $DB->count_records('course_modules');
        
        $this->content->text .= html_writer::tag('div', 
            html_writer::tag('strong', $total_courses) . ' ' . get_string('courses', 'block_lmsk2_stats'),
            array('class' => 'stat-item')
        );
        
        $this->content->text .= html_writer::tag('div', 
            html_writer::tag('strong', $total_users) . ' ' . get_string('users', 'block_lmsk2_stats'),
            array('class' => 'stat-item')
        );
        
        $this->content->text .= html_writer::tag('div', 
            html_writer::tag('strong', $total_activities) . ' ' . get_string('activities', 'block_lmsk2_stats'),
            array('class' => 'stat-item')
        );
        
        return $this->content;
    }
    
    public function applicable_formats() {
        return array('site' => true, 'course' => true);
    }
}
EOF

    # Create language file for stats block
    mkdir -p "$stats_block_dir/lang/en"
    cat > "$stats_block_dir/lang/en/block_lmsk2_stats.php" << 'EOF'
<?php
defined('MOODLE_INTERNAL') || die();

$string['statistics'] = 'Statistics';
$string['courses'] = 'Courses';
$string['users'] = 'Users';
$string['activities'] = 'Activities';
$string['pluginname'] = 'LMSK2 Statistics Block';
EOF

    # Set proper permissions
    chown -R www-data:www-data "$stats_block_dir"
    chmod -R 755 "$stats_block_dir"
    
    print_success "Custom statistics block created"
    log_message "SUCCESS" "Custom blocks creation completed"
}

# Configure advanced settings
configure_advanced_settings() {
    print_section "Configuring Advanced Settings"
    
    print_info "Configuring advanced Moodle settings..."
    
    # Performance settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enablecompletion --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enablebadges --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enableportfolios --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enableblogs --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=enablemessaging --set=1
    
    # Security settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=passwordpolicy --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=passwordreuselimit --set=5
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=minpasswordlength --set=8
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=minpassworddigits --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=minpasswordlower --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=minpasswordupper --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=minpasswordspecial --set=1
    
    # Session settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=session_handler_class --set="\core\session\redis"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=session_redis_host --set="127.0.0.1"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=session_redis_port --set="6379"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=session_redis_database --set="0"
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=session_redis_prefix --set="moodle_session_"
    
    # Cache settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=cachejs --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=cachecss --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=cachetemplates --set=1
    
    # File settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=maxbytes --set=268435456  # 256MB
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=maxareabytes --set=1073741824  # 1GB
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=repositorycacheexpire --set=120
    
    # Logging settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=loglifetime --set=30
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=logguests --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=logreader_standardlog --set=1
    
    print_success "Advanced settings configured"
    log_message "SUCCESS" "Advanced settings configuration completed"
}

# Create custom language pack
create_custom_language_pack() {
    print_section "Creating Custom Language Pack"
    
    print_info "Creating custom language pack..."
    
    # Create custom language directory
    local custom_lang_dir="$MOODLE_DIR/lang/lmsk2"
    mkdir -p "$custom_lang_dir"
    
    # Create language configuration
    cat > "$custom_lang_dir/langconfig.php" << 'EOF'
<?php
defined('MOODLE_INTERNAL') || die();

$string['thislanguage'] = 'LMSK2 Custom';
$string['thislanguageint'] = 'LMSK2 Custom';
$string['parentlanguage'] = 'en';
$string['locale'] = 'en_US.UTF-8';
$string['localewin'] = 'English_United States.1252';
$string['localewincharset'] = 'UTF-8';
$string['direction'] = 'ltr';
$string['fullname'] = 'LMSK2 Custom Language Pack';
$string['alphabet'] = 'A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z';
$string['listsep'] = ',';
$string['yes'] = 'Yes';
$string['no'] = 'No';
$string['am'] = 'AM';
$string['pm'] = 'PM';
$string['firstdayofweek'] = '1';
$string['calendar'] = 'gregorian';
$string['labelsep'] = ': ';
$string['decimalsep'] = '.';
$string['thousandssep'] = ',';
$string['decsep'] = '.';
$string['listsep'] = ',';
$string['iso6391'] = 'en';
$string['iso6392'] = 'eng';
EOF

    # Create custom strings
    cat > "$custom_lang_dir/moodle.php" << 'EOF'
<?php
defined('MOODLE_INTERNAL') || die();

// Custom LMSK2 strings
$string['lmsk2_welcome'] = 'Welcome to LMSK2 Learning Platform';
$string['lmsk2_custom_message'] = 'This is a customized learning management system';
$string['lmsk2_brand_name'] = 'LMSK2';
$string['lmsk2_support_email'] = 'support@lmsk2.com';
$string['lmsk2_help_desk'] = 'Help Desk';
$string['lmsk2_quick_links'] = 'Quick Links';
$string['lmsk2_announcements'] = 'Announcements';
$string['lmsk2_news'] = 'News & Updates';
$string['lmsk2_contact_us'] = 'Contact Us';
$string['lmsk2_about'] = 'About LMSK2';
$string['lmsk2_privacy_policy'] = 'Privacy Policy';
$string['lmsk2_terms_of_service'] = 'Terms of Service';
$string['lmsk2_copyright'] = 'Copyright © {$a} LMSK2. All rights reserved.';
EOF

    # Set proper permissions
    chown -R www-data:www-data "$custom_lang_dir"
    chmod -R 755 "$custom_lang_dir"
    
    print_success "Custom language pack created"
    log_message "SUCCESS" "Custom language pack creation completed"
}

# Setup custom navigation
setup_custom_navigation() {
    print_section "Setting Up Custom Navigation"
    
    print_info "Configuring custom navigation..."
    
    # Configure navigation settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=navshowfullcoursenames --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=navshowcategories --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=navshowmycourses --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=navshowallcourses --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=navsortmycoursessort --set="fullname"
    
    # Configure dashboard settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mycoursesonfrontpage --set=1
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mycoursesperpage --set=12
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=mycoursesmax --set=20
    
    # Configure course settings
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=coursesperpage --set=20
    php "$MOODLE_DIR/admin/cli/cfg.php" --name=courseswithsummarieslimit --set=10
    
    print_success "Custom navigation configured"
    log_message "SUCCESS" "Custom navigation setup completed"
}

# Test customizations
test_customizations() {
    print_section "Testing Customizations"
    
    print_info "Testing custom theme..."
    if [[ -d "$MOODLE_DIR/theme/lmsk2" ]]; then
        print_success "Custom theme directory exists"
    else
        print_warning "Custom theme directory not found"
    fi
    
    print_info "Testing custom blocks..."
    if [[ -d "$MOODLE_DIR/blocks/lmsk2_welcome" ]]; then
        print_success "Custom welcome block exists"
    else
        print_warning "Custom welcome block not found"
    fi
    
    if [[ -d "$MOODLE_DIR/blocks/lmsk2_stats" ]]; then
        print_success "Custom statistics block exists"
    else
        print_warning "Custom statistics block not found"
    fi
    
    print_info "Testing custom language pack..."
    if [[ -d "$MOODLE_DIR/lang/lmsk2" ]]; then
        print_success "Custom language pack exists"
    else
        print_warning "Custom language pack not found"
    fi
    
    print_info "Testing advanced settings..."
    local advanced_settings=("enablecompletion" "enablebadges" "passwordpolicy" "cachejs")
    for setting in "${advanced_settings[@]}"; do
        if php "$MOODLE_DIR/admin/cli/cfg.php" --name="$setting" --get | grep -q "1"; then
            print_success "Advanced setting $setting is enabled"
        else
            print_warning "Advanced setting $setting is not enabled"
        fi
    done
    
    log_message "SUCCESS" "Customizations testing completed"
}

# =============================================================================
# Main Function
# =============================================================================

main() {
    print_color $CYAN "=============================================================================="
    print_color $WHITE "  $SCRIPT_NAME v$SCRIPT_VERSION"
    print_color $CYAN "=============================================================================="
    
    log_message "INFO" "Starting customizations setup"
    
    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root"
        exit 1
    fi
    
    # Check if Moodle is installed
    if [[ ! -d "$MOODLE_DIR" ]]; then
        print_error "Moodle directory not found: $MOODLE_DIR"
        exit 1
    fi
    
    # Parse command line arguments
    local action="all"
    while [[ $# -gt 0 ]]; do
        case $1 in
            --theme)
                action="theme"
                shift
                ;;
            --blocks)
                action="blocks"
                shift
                ;;
            --settings)
                action="settings"
                shift
                ;;
            --language)
                action="language"
                shift
                ;;
            --navigation)
                action="navigation"
                shift
                ;;
            --test)
                action="test"
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --theme       Customize theme"
                echo "  --blocks      Create custom blocks"
                echo "  --settings    Configure advanced settings"
                echo "  --language    Create custom language pack"
                echo "  --navigation  Setup custom navigation"
                echo "  --test        Test customizations"
                echo "  --help        Show this help"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Execute based on action
    case $action in
        "theme")
            customize_theme
            ;;
        "blocks")
            create_custom_blocks
            ;;
        "settings")
            configure_advanced_settings
            ;;
        "language")
            create_custom_language_pack
            ;;
        "navigation")
            setup_custom_navigation
            ;;
        "test")
            test_customizations
            ;;
        "all")
            customize_theme
            create_custom_blocks
            configure_advanced_settings
            create_custom_language_pack
            setup_custom_navigation
            test_customizations
            ;;
    esac
    
    print_section "Customizations Setup Complete"
    print_success "Customizations setup completed successfully!"
    print_info "Log file: $LOG_DIR/customizations.log"
    
    log_message "SUCCESS" "Customizations setup completed successfully"
}

# Trap errors
trap 'print_error "Script failed at line $LINENO"; log_message "ERROR" "Script failed at line $LINENO"; exit 1' ERR

# Run main function
main "$@"
