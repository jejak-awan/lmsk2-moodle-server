package handlers

import (
	"net/http"
	"strconv"

	"lms-manager/services"

	"github.com/gin-gonic/gin"
)

// APIHandler handles API requests
type APIHandler struct {
	monitorService  *services.MonitorService
	moodleService   *services.MoodleService
	securityService *services.SecurityService
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(monitorService *services.MonitorService, moodleService *services.MoodleService, securityService *services.SecurityService) *APIHandler {
	return &APIHandler{
		monitorService:  monitorService,
		moodleService:   moodleService,
		securityService: securityService,
	}
}

// GetStats returns system statistics
func (h *APIHandler) GetStats(c *gin.Context) {
	stats := h.monitorService.GetStats()
	c.JSON(http.StatusOK, stats)
}

// GetUsers returns user information
func (h *APIHandler) GetUsers(c *gin.Context) {
	// This would typically get user information from the auth service
	// For now, return basic user info
	users := []map[string]interface{}{
		{
			"id":       "1",
			"username": "admin",
			"role":     "admin",
			"active":   true,
		},
	}

	c.JSON(http.StatusOK, users)
}

// GetMoodleStatus returns Moodle status
func (h *APIHandler) GetMoodleStatus(c *gin.Context) {
	status := h.moodleService.GetStatus()
	c.JSON(http.StatusOK, status)
}

// StartMoodle starts Moodle
func (h *APIHandler) StartMoodle(c *gin.Context) {
	err := h.moodleService.Start()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to start Moodle",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Moodle started successfully",
		"status":  "running",
	})
}

// StopMoodle stops Moodle
func (h *APIHandler) StopMoodle(c *gin.Context) {
	err := h.moodleService.Stop()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to stop Moodle",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Moodle stopped successfully",
		"status":  "stopped",
	})
}

// RestartMoodle restarts Moodle
func (h *APIHandler) RestartMoodle(c *gin.Context) {
	err := h.moodleService.Restart()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to restart Moodle",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Moodle restarted successfully",
		"status":  "running",
	})
}

// GetUserStats returns user statistics
func (h *APIHandler) GetUserStats(c *gin.Context) {
	// This would typically get user statistics from the auth service
	// For now, return mock data
	stats := map[string]interface{}{
		"total_users":     1,
		"active_users":    1,
		"online_users":    1,
		"admin_users":     1,
		"operator_users":  0,
		"viewer_users":    0,
		"last_24_hours":   0,
		"last_7_days":     0,
		"last_30_days":    0,
	}

	c.JSON(http.StatusOK, stats)
}

// CreateUser creates a new user
func (h *APIHandler) CreateUser(c *gin.Context) {
	// This would typically create a user via the auth service
	// For now, return a success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "User creation endpoint - implementation needed",
	})
}

// UpdateUser updates a user
func (h *APIHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	// This would typically update a user via the auth service
	// For now, return a success response
	c.JSON(http.StatusOK, gin.H{
		"message": "User update endpoint - implementation needed",
		"user_id": userID,
	})
}

// DeleteUser deletes a user
func (h *APIHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	// This would typically delete a user via the auth service
	// For now, return a success response
	c.JSON(http.StatusOK, gin.H{
		"message": "User deletion endpoint - implementation needed",
		"user_id": userID,
	})
}

// GetSecurityEvents returns security events
func (h *APIHandler) GetSecurityEvents(c *gin.Context) {
	// This would typically get security events from the database
	// For now, return mock data
	events := []map[string]interface{}{
		{
			"id":        "1",
			"type":      "login_success",
			"message":   "User logged in successfully",
			"ip":        "127.0.0.1",
			"user_agent": "Mozilla/5.0...",
			"severity":  "info",
			"timestamp": "2025-01-09T10:00:00Z",
		},
	}

	c.JSON(http.StatusOK, events)
}

// GetAlerts returns system alerts
func (h *APIHandler) GetAlerts(c *gin.Context) {
	alerts, err := h.monitorService.GetAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get alerts",
		})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// ResolveAlert resolves an alert
func (h *APIHandler) ResolveAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alert ID is required",
		})
		return
	}

	err := h.monitorService.ResolveAlert(alertID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to resolve alert",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Alert resolved successfully",
		"alert_id": alertID,
	})
}

// GetSystemLogs returns system logs
func (h *APIHandler) GetSystemLogs(c *gin.Context) {
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	logs, err := h.monitorService.GetSystemLogs(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get system logs",
		})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetMoodleInfo returns detailed Moodle information
func (h *APIHandler) GetMoodleInfo(c *gin.Context) {
	info, err := h.moodleService.GetMoodleInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get Moodle info",
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

// BackupMoodle creates a backup of Moodle
func (h *APIHandler) BackupMoodle(c *gin.Context) {
	backupPath := c.Query("path")
	if backupPath == "" {
		backupPath = "/tmp/moodle-backups"
	}

	err := h.moodleService.BackupMoodle(backupPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create backup",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Backup created successfully",
		"backup_path": backupPath,
	})
}

// RestoreMoodle restores Moodle from backup
func (h *APIHandler) RestoreMoodle(c *gin.Context) {
	backupFile := c.Query("file")
	if backupFile == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Backup file path is required",
		})
		return
	}

	err := h.moodleService.RestoreMoodle(backupFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to restore backup",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Moodle restored successfully",
		"backup_file": backupFile,
	})
}

// GetSecurityStats returns security statistics
func (h *APIHandler) GetSecurityStats(c *gin.Context) {
	stats := h.securityService.GetSecurityStats()
	c.JSON(http.StatusOK, stats)
}

// GetBlacklist returns the IP blacklist
func (h *APIHandler) GetBlacklist(c *gin.Context) {
	blacklist := h.securityService.GetBlacklist()
	c.JSON(http.StatusOK, blacklist)
}

// GetWhitelist returns the IP whitelist
func (h *APIHandler) GetWhitelist(c *gin.Context) {
	whitelist := h.securityService.GetWhitelist()
	c.JSON(http.StatusOK, whitelist)
}

// UpdateWhitelist updates the IP whitelist
func (h *APIHandler) UpdateWhitelist(c *gin.Context) {
	var req struct {
		IPs []string `json:"ips" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	h.securityService.UpdateWhitelist(req.IPs)

	c.JSON(http.StatusOK, gin.H{
		"message": "Whitelist updated successfully",
		"ips":     req.IPs,
	})
}

// ClearBlacklist clears the IP blacklist
func (h *APIHandler) ClearBlacklist(c *gin.Context) {
	h.securityService.ClearBlacklist()

	c.JSON(http.StatusOK, gin.H{
		"message": "Blacklist cleared successfully",
	})
}
