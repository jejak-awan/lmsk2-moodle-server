package models

import (
	"time"
)

// SystemStats represents system statistics
type SystemStats struct {
	CPUUsage    float64       `json:"cpu_usage"`
	MemoryUsage float64       `json:"memory_usage"`
	DiskUsage   float64       `json:"disk_usage"`
	NetworkIO   NetworkStats  `json:"network_io"`
	Uptime      int64         `json:"uptime"`
	LoadAvg     LoadAvgStats  `json:"load_avg"`
	Timestamp   time.Time     `json:"timestamp"`
}

// NetworkStats represents network I/O statistics
type NetworkStats struct {
	BytesReceived uint64 `json:"bytes_received"`
	BytesSent     uint64 `json:"bytes_sent"`
	PacketsReceived uint64 `json:"packets_received"`
	PacketsSent   uint64 `json:"packets_sent"`
}

// LoadAvgStats represents load average statistics
type LoadAvgStats struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

// UserSession represents a user session
type UserSession struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	LoginTime time.Time `json:"login_time"`
	IP        string    `json:"ip"`
	Active    bool      `json:"active"`
	LastSeen  time.Time `json:"last_seen"`
	UserAgent string    `json:"user_agent"`
}

// MoodleStatus represents Moodle service status
type MoodleStatus struct {
	Running    bool      `json:"running"`
	Version    string    `json:"version"`
	Uptime     int64     `json:"uptime"`
	LastCheck  time.Time `json:"last_check"`
	Error      string    `json:"error,omitempty"`
	ProcessID  int       `json:"process_id,omitempty"`
}

// SecurityEvent represents a security event
type SecurityEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
	Severity  string    `json:"severity"`
}

// Alert represents a system alert
type Alert struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
	Resolved  bool      `json:"resolved"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// PerformanceMetrics represents performance metrics
type PerformanceMetrics struct {
	ResponseTime    float64 `json:"response_time"`
	RequestCount    int64   `json:"request_count"`
	ErrorCount      int64   `json:"error_count"`
	ActiveConnections int   `json:"active_connections"`
	MemoryUsage     float64 `json:"memory_usage"`
	CPUUsage        float64 `json:"cpu_usage"`
	Timestamp       time.Time `json:"timestamp"`
}

// DatabaseStats represents database statistics
type DatabaseStats struct {
	Connections    int     `json:"connections"`
	MaxConnections int     `json:"max_connections"`
	QueriesPerSec  float64 `json:"queries_per_sec"`
	SlowQueries    int     `json:"slow_queries"`
	Uptime         int64   `json:"uptime"`
	Timestamp      time.Time `json:"timestamp"`
}

// LogEntry represents a log entry
type LogEntry struct {
	ID        string    `json:"id"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}
