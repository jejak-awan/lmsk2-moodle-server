package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"lms-manager/config"
	"lms-manager/models"
	"lms-manager/utils"
)

// MonitorService handles system monitoring
type MonitorService struct {
	config     config.MonitoringConfig
	db         *sql.DB
	stats      *models.SystemStats
	mu         sync.RWMutex
	stopChan   chan bool
	running    bool
}

// NewMonitorService creates a new monitor service
func NewMonitorService(cfg config.MonitoringConfig) *MonitorService {
	return &MonitorService{
		config:   cfg,
		stats:    &models.SystemStats{},
		stopChan: make(chan bool),
		running:  false,
	}
}

// SetDatabase sets the database connection
func (m *MonitorService) SetDatabase(db *sql.DB) {
	m.db = db
}

// Start starts the monitoring service
func (m *MonitorService) Start() {
	if m.running {
		return
	}

	m.running = true
	go m.monitorLoop()
	utils.Info("Monitoring service started")
}

// Stop stops the monitoring service
func (m *MonitorService) Stop() {
	if !m.running {
		return
	}

	m.running = false
	m.stopChan <- true
	utils.Info("Monitoring service stopped")
}

// GetStats returns the current system stats
func (m *MonitorService) GetStats() *models.SystemStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// monitorLoop runs the monitoring loop
func (m *MonitorService) monitorLoop() {
	ticker := time.NewTicker(time.Duration(m.config.UpdateInterval) * time.Second)
	defer ticker.Stop()

	// Initial update
	m.updateStats()

	for {
		select {
		case <-ticker.C:
			m.updateStats()
		case <-m.stopChan:
			return
		}
	}
}

// updateStats updates the system statistics
func (m *MonitorService) updateStats() {
	stats := &models.SystemStats{
		Timestamp: time.Now(),
	}

	// Get CPU usage
	if cpuUsage, err := utils.GetCPUUsage(); err == nil {
		stats.CPUUsage = cpuUsage
	}

	// Get memory usage
	if memoryUsage, err := utils.GetMemoryUsage(); err == nil {
		stats.MemoryUsage = memoryUsage
	}

	// Get disk usage
	if diskUsage, err := utils.GetDiskUsage("/"); err == nil {
		stats.DiskUsage = diskUsage
	}

	// Get uptime
	if uptime, err := utils.GetUptime(); err == nil {
		stats.Uptime = uptime
	}

	// Get load average
	if loadAvg, err := utils.GetLoadAverage(); err == nil {
		stats.LoadAvg = models.LoadAvgStats{
			Load1:  loadAvg[0],
			Load5:  loadAvg[1],
			Load15: loadAvg[2],
		}
	}

	// Get network stats
	if networkStats, err := utils.GetNetworkStats(); err == nil {
		stats.NetworkIO = models.NetworkStats{
			BytesReceived:   networkStats["bytes_received"].(uint64),
			BytesSent:       networkStats["bytes_sent"].(uint64),
			PacketsReceived: 0, // Not available from /proc/net/dev
			PacketsSent:     0, // Not available from /proc/net/dev
		}
	}

	// Update stats
	m.mu.Lock()
	m.stats = stats
	m.mu.Unlock()

	// Check for alerts
	m.checkAlerts(stats)

	// Log stats to database
	m.logStats(stats)
}

// checkAlerts checks for system alerts
func (m *MonitorService) checkAlerts(stats *models.SystemStats) {
	alerts := []models.Alert{}

	// CPU alert
	if stats.CPUUsage > m.config.AlertThresholds.CPU {
		alerts = append(alerts, models.Alert{
			ID:        utils.GenerateID(),
			Type:      "cpu_high",
			Message:   fmt.Sprintf("CPU usage is high: %.1f%%", stats.CPUUsage),
			Severity:  "warning",
			Timestamp: time.Now(),
			Resolved:  false,
		})
	}

	// Memory alert
	if stats.MemoryUsage > m.config.AlertThresholds.Memory {
		alerts = append(alerts, models.Alert{
			ID:        utils.GenerateID(),
			Type:      "memory_high",
			Message:   fmt.Sprintf("Memory usage is high: %.1f%%", stats.MemoryUsage),
			Severity:  "warning",
			Timestamp: time.Now(),
			Resolved:  false,
		})
	}

	// Disk alert
	if stats.DiskUsage > m.config.AlertThresholds.Disk {
		alerts = append(alerts, models.Alert{
			ID:        utils.GenerateID(),
			Type:      "disk_high",
			Message:   fmt.Sprintf("Disk usage is high: %.1f%%", stats.DiskUsage),
			Severity:  "critical",
			Timestamp: time.Now(),
			Resolved:  false,
		})
	}

	// Save alerts to database
	for _, alert := range alerts {
		m.saveAlert(alert)
	}
}

// logStats logs system stats to database
func (m *MonitorService) logStats(stats *models.SystemStats) {
	if m.db == nil {
		return
	}

	// Convert stats to JSON
	statsJSON, err := json.Marshal(stats)
	if err != nil {
		utils.Error("Failed to marshal stats: %v", err)
		return
	}

	// Insert into database
	_, err = m.db.Exec(`
		INSERT INTO system_logs (id, level, message, source, data, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, utils.GenerateID(), "INFO", "System stats updated", "monitor", string(statsJSON), time.Now())

	if err != nil {
		utils.Error("Failed to log stats: %v", err)
	}
}

// saveAlert saves an alert to the database
func (m *MonitorService) saveAlert(alert models.Alert) {
	if m.db == nil {
		return
	}

	// Check if alert already exists
	var count int
	err := m.db.QueryRow(`
		SELECT COUNT(*) FROM alerts 
		WHERE type = ? AND resolved = 0
	`, alert.Type).Scan(&count)

	if err != nil {
		utils.Error("Failed to check existing alert: %v", err)
		return
	}

	// Only save if alert doesn't exist
	if count == 0 {
		_, err = m.db.Exec(`
			INSERT INTO alerts (id, type, message, severity, resolved, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, alert.ID, alert.Type, alert.Message, alert.Severity, alert.Resolved, alert.Timestamp)

		if err != nil {
			utils.Error("Failed to save alert: %v", err)
		} else {
			utils.Warn("Alert created: %s", alert.Message)
		}
	}
}

// GetAlerts returns active alerts
func (m *MonitorService) GetAlerts() ([]models.Alert, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := m.db.Query(`
		SELECT id, type, message, severity, resolved, created_at, resolved_at
		FROM alerts
		WHERE resolved = 0
		ORDER BY created_at DESC
		LIMIT 50
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var alert models.Alert
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&alert.ID,
			&alert.Type,
			&alert.Message,
			&alert.Severity,
			&alert.Resolved,
			&alert.Timestamp,
			&resolvedAt,
		)
		if err != nil {
			return nil, err
		}

		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// ResolveAlert resolves an alert
func (m *MonitorService) ResolveAlert(alertID string) error {
	if m.db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := m.db.Exec(`
		UPDATE alerts 
		SET resolved = 1, resolved_at = ?
		WHERE id = ?
	`, time.Now(), alertID)

	return err
}

// GetSystemLogs returns system logs
func (m *MonitorService) GetSystemLogs(limit int) ([]models.LogEntry, error) {
	if m.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if limit <= 0 {
		limit = 100
	}

	rows, err := m.db.Query(`
		SELECT id, level, message, source, data, created_at
		FROM system_logs
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.LogEntry
	for rows.Next() {
		var log models.LogEntry
		var data sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.Level,
			&log.Message,
			&log.Source,
			&data,
			&log.Timestamp,
		)
		if err != nil {
			return nil, err
		}

		if data.Valid {
			json.Unmarshal([]byte(data.String), &log.Data)
		}

		logs = append(logs, log)
	}

	return logs, nil
}
