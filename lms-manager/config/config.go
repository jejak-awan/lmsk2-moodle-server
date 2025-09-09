package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `json:"server"`
	Moodle    MoodleConfig    `json:"moodle"`
	Security  SecurityConfig  `json:"security"`
	Monitoring MonitoringConfig `json:"monitoring"`
}

// ServerConfig contains server configuration
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
	Debug bool  `json:"debug"`
}

// MoodleConfig contains Moodle configuration
type MoodleConfig struct {
	Path       string `json:"path"`
	ConfigPath string `json:"config_path"`
	DataPath   string `json:"data_path"`
}

// SecurityConfig contains security configuration
type SecurityConfig struct {
	JWTSecret    string   `json:"jwt_secret"`
	SessionTimeout int    `json:"session_timeout"`
	RateLimit    int      `json:"rate_limit"`
	AllowedIPs   []string `json:"allowed_ips"`
}

// MonitoringConfig contains monitoring configuration
type MonitoringConfig struct {
	UpdateInterval int                    `json:"update_interval"`
	LogRetention   int                    `json:"log_retention"`
	AlertThresholds AlertThresholdsConfig `json:"alert_thresholds"`
}

// AlertThresholdsConfig contains alert threshold configuration
type AlertThresholdsConfig struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Disk   float64 `json:"disk"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:  8080,
			Host:  "0.0.0.0",
			Debug: false,
		},
		Moodle: MoodleConfig{
			Path:       "/var/www/moodle",
			ConfigPath: "/var/www/moodle/config.php",
			DataPath:   "/var/www/moodledata",
		},
		Security: SecurityConfig{
			JWTSecret:     "your-secret-key-change-this",
			SessionTimeout: 3600,
			RateLimit:     100,
			AllowedIPs:   []string{"127.0.0.1", "192.168.1.0/24"},
		},
		Monitoring: MonitoringConfig{
			UpdateInterval: 30,
			LogRetention:   7,
			AlertThresholds: AlertThresholdsConfig{
				CPU:    80.0,
				Memory: 85.0,
				Disk:   90.0,
			},
		},
	}
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file
		config := DefaultConfig()
		if err := SaveConfig(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %v", err)
		}
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse config
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Write config file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// ValidateConfig validates configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Moodle.Path == "" {
		return fmt.Errorf("moodle path is required")
	}

	if c.Security.JWTSecret == "" || c.Security.JWTSecret == "your-secret-key-change-this" {
		return fmt.Errorf("jwt secret must be set and changed from default")
	}

	if c.Security.SessionTimeout <= 0 {
		return fmt.Errorf("session timeout must be positive")
	}

	if c.Security.RateLimit <= 0 {
		return fmt.Errorf("rate limit must be positive")
	}

	if c.Monitoring.UpdateInterval <= 0 {
		return fmt.Errorf("update interval must be positive")
	}

	return nil
}
