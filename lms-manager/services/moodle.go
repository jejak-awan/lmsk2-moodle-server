package services

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"lms-manager/config"
	"lms-manager/models"
	"lms-manager/utils"
)

// MoodleService handles Moodle management
type MoodleService struct {
	config config.MoodleConfig
	db     *sql.DB
	status *models.MoodleStatus
}

// NewMoodleService creates a new Moodle service
func NewMoodleService(cfg config.MoodleConfig) *MoodleService {
	return &MoodleService{
		config: cfg,
		status: &models.MoodleStatus{},
	}
}

// SetDatabase sets the database connection
func (m *MoodleService) SetDatabase(db *sql.DB) {
	m.db = db
}

// GetStatus returns the current Moodle status
func (m *MoodleService) GetStatus() *models.MoodleStatus {
	status := &models.MoodleStatus{
		LastCheck: time.Now(),
	}

	// Check if Moodle is running
	running, err := m.isMoodleRunning()
	if err != nil {
		status.Error = err.Error()
		return status
	}

	status.Running = running

	if running {
		// Get version
		version, err := m.getMoodleVersion()
		if err != nil {
			status.Error = err.Error()
		} else {
			status.Version = version
		}

		// Get process ID
		pid, err := m.getMoodlePID()
		if err != nil {
			status.Error = err.Error()
		} else {
			status.ProcessID = pid
		}

		// Get uptime
		uptime, err := m.getMoodleUptime()
		if err != nil {
			status.Error = err.Error()
		} else {
			status.Uptime = uptime
		}
	}

	return status
}

// Start starts Moodle
func (m *MoodleService) Start() error {
	// Check if Moodle is already running
	if running, _ := m.isMoodleRunning(); running {
		return fmt.Errorf("Moodle is already running")
	}

	// Check if Moodle directory exists
	if !utils.FileExists(m.config.Path) {
		return fmt.Errorf("Moodle directory does not exist: %s", m.config.Path)
	}

	// Check if config file exists
	if !utils.FileExists(m.config.ConfigPath) {
		return fmt.Errorf("Moodle config file does not exist: %s", m.config.ConfigPath)
	}

	// Start web server (assuming Nginx)
	cmd := exec.Command("systemctl", "start", "nginx")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start Nginx: %v", err)
	}

	// Start PHP-FPM
	cmd = exec.Command("systemctl", "start", "php8.1-fpm")
	if err := cmd.Run(); err != nil {
		// Try different PHP versions
		cmd = exec.Command("systemctl", "start", "php8.0-fpm")
		if err := cmd.Run(); err != nil {
			cmd = exec.Command("systemctl", "start", "php7.4-fpm")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to start PHP-FPM: %v", err)
			}
		}
	}

	// Start MariaDB/MySQL
	cmd = exec.Command("systemctl", "start", "mariadb")
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("systemctl", "start", "mysql")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start database: %v", err)
		}
	}

	utils.Info("Moodle started successfully")
	return nil
}

// Stop stops Moodle
func (m *MoodleService) Stop() error {
	// Check if Moodle is running
	if running, _ := m.isMoodleRunning(); !running {
		return fmt.Errorf("Moodle is not running")
	}

	// Stop web server
	cmd := exec.Command("systemctl", "stop", "nginx")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop Nginx: %v", err)
	}

	// Stop PHP-FPM
	cmd = exec.Command("systemctl", "stop", "php8.1-fpm")
	if err := cmd.Run(); err != nil {
		// Try different PHP versions
		cmd = exec.Command("systemctl", "stop", "php8.0-fpm")
		if err := cmd.Run(); err != nil {
			cmd = exec.Command("systemctl", "stop", "php7.4-fpm")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to stop PHP-FPM: %v", err)
			}
		}
	}

	utils.Info("Moodle stopped successfully")
	return nil
}

// Restart restarts Moodle
func (m *MoodleService) Restart() error {
	// Stop first
	if err := m.Stop(); err != nil {
		utils.Warn("Failed to stop Moodle: %v", err)
	}

	// Wait a bit
	time.Sleep(2 * time.Second)

	// Start again
	if err := m.Start(); err != nil {
		return fmt.Errorf("failed to restart Moodle: %v", err)
	}

	utils.Info("Moodle restarted successfully")
	return nil
}

// isMoodleRunning checks if Moodle is running
func (m *MoodleService) isMoodleRunning() (bool, error) {
	// Check if Nginx is running
	cmd := exec.Command("systemctl", "is-active", "nginx")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	if strings.TrimSpace(string(output)) != "active" {
		return false, nil
	}

	// Check if PHP-FPM is running
	cmd = exec.Command("systemctl", "is-active", "php8.1-fpm")
	output, err = cmd.Output()
	if err != nil {
		// Try different PHP versions
		cmd = exec.Command("systemctl", "is-active", "php8.0-fpm")
		output, err = cmd.Output()
		if err != nil {
			cmd = exec.Command("systemctl", "is-active", "php7.4-fpm")
			output, err = cmd.Output()
			if err != nil {
				return false, err
			}
		}
	}

	if strings.TrimSpace(string(output)) != "active" {
		return false, nil
	}

	// Check if database is running
	cmd = exec.Command("systemctl", "is-active", "mariadb")
	output, err = cmd.Output()
	if err != nil {
		cmd = exec.Command("systemctl", "is-active", "mysql")
		output, err = cmd.Output()
		if err != nil {
			return false, err
		}
	}

	if strings.TrimSpace(string(output)) != "active" {
		return false, nil
	}

	return true, nil
}

// getMoodleVersion gets the Moodle version
func (m *MoodleService) getMoodleVersion() (string, error) {
	// Read version.php file
	versionFile := filepath.Join(m.config.Path, "version.php")
	if !utils.FileExists(versionFile) {
		return "", fmt.Errorf("version.php not found")
	}

	// Read file content
	content, err := os.ReadFile(versionFile)
	if err != nil {
		return "", err
	}

	// Parse version
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "$version") {
			// Extract version number
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				version := strings.TrimSpace(parts[1])
				version = strings.Trim(version, ";")
				version = strings.Trim(version, " ")
				return version, nil
			}
		}
	}

	return "", fmt.Errorf("version not found in version.php")
}

// getMoodlePID gets the Moodle process ID
func (m *MoodleService) getMoodlePID() (int, error) {
	// Get Nginx PID
	cmd := exec.Command("pgrep", "nginx")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	pidStr := strings.TrimSpace(string(output))
	pid, err := utils.ParseInt(pidStr)
	if err != nil {
		return 0, err
	}

	return pid, nil
}

// getMoodleUptime gets the Moodle uptime
func (m *MoodleService) getMoodleUptime() (int64, error) {
	// Get Nginx uptime
	cmd := exec.Command("systemctl", "show", "nginx", "--property=ActiveEnterTimestamp")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Parse timestamp
	line := strings.TrimSpace(string(output))
	if !strings.HasPrefix(line, "ActiveEnterTimestamp=") {
		return 0, fmt.Errorf("invalid timestamp format")
	}

	timestampStr := strings.TrimPrefix(line, "ActiveEnterTimestamp=")
	timestamp, err := time.Parse("Mon 2006-01-02 15:04:05 MST", timestampStr)
	if err != nil {
		return 0, err
	}

	uptime := time.Since(timestamp)
	return int64(uptime.Seconds()), nil
}

// GetMoodleInfo gets detailed Moodle information
func (m *MoodleService) GetMoodleInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Basic info
	info["path"] = m.config.Path
	info["config_path"] = m.config.ConfigPath
	info["data_path"] = m.config.DataPath

	// Check if directories exist
	info["path_exists"] = utils.FileExists(m.config.Path)
	info["config_exists"] = utils.FileExists(m.config.ConfigPath)
	info["data_exists"] = utils.FileExists(m.config.DataPath)

	// Get sizes
	if utils.FileExists(m.config.Path) {
		size, err := utils.GetDirectorySize(m.config.Path)
		if err == nil {
			info["path_size"] = utils.FormatBytes(size)
		}
	}

	if utils.FileExists(m.config.DataPath) {
		size, err := utils.GetDirectorySize(m.config.DataPath)
		if err == nil {
			info["data_size"] = utils.FormatBytes(size)
		}
	}

	// Get version
	if version, err := m.getMoodleVersion(); err == nil {
		info["version"] = version
	}

	// Get status
	status := m.GetStatus()
	info["running"] = status.Running
	info["uptime"] = utils.FormatDuration(status.Uptime)
	info["process_id"] = status.ProcessID

	return info, nil
}

// BackupMoodle creates a backup of Moodle
func (m *MoodleService) BackupMoodle(backupPath string) error {
	if !utils.FileExists(m.config.Path) {
		return fmt.Errorf("Moodle directory does not exist")
	}

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Create tar backup
	backupFile := filepath.Join(backupPath, fmt.Sprintf("moodle-backup-%s.tar.gz", time.Now().Format("2006-01-02-15-04-05")))
	
	cmd := exec.Command("tar", "-czf", backupFile, "-C", filepath.Dir(m.config.Path), filepath.Base(m.config.Path))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	utils.Info("Moodle backup created: %s", backupFile)
	return nil
}

// RestoreMoodle restores Moodle from backup
func (m *MoodleService) RestoreMoodle(backupFile string) error {
	if !utils.FileExists(backupFile) {
		return fmt.Errorf("backup file does not exist: %s", backupFile)
	}

	// Stop Moodle first
	if err := m.Stop(); err != nil {
		utils.Warn("Failed to stop Moodle: %v", err)
	}

	// Remove existing Moodle directory
	if utils.FileExists(m.config.Path) {
		if err := os.RemoveAll(m.config.Path); err != nil {
			return fmt.Errorf("failed to remove existing Moodle directory: %v", err)
		}
	}

	// Extract backup
	cmd := exec.Command("tar", "-xzf", backupFile, "-C", filepath.Dir(m.config.Path))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract backup: %v", err)
	}

	// Start Moodle
	if err := m.Start(); err != nil {
		return fmt.Errorf("failed to start Moodle after restore: %v", err)
	}

	utils.Info("Moodle restored from backup: %s", backupFile)
	return nil
}
