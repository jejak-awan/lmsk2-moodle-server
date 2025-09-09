package handlers

import (
	"net/http"
	"strconv"

	"lms-manager/services"

	"github.com/gin-gonic/gin"
)

// DashboardHandler handles dashboard requests
type DashboardHandler struct {
	monitorService *services.MonitorService
	moodleService  *services.MoodleService
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(monitorService *services.MonitorService, moodleService *services.MoodleService) *DashboardHandler {
	return &DashboardHandler{
		monitorService: monitorService,
		moodleService:  moodleService,
	}
}

// Dashboard renders the main dashboard
func (h *DashboardHandler) Dashboard(c *gin.Context) {
	// Get system stats
	stats := h.monitorService.GetStats()
	
	// Get Moodle status
	moodleStatus := h.moodleService.GetStatus()

	// Prepare dashboard data
	dashboardData := gin.H{
		"title":         "LMS Manager - K2NET",
		"system_stats":  stats,
		"moodle_status": moodleStatus,
		"timestamp":     stats.Timestamp,
	}

	// Render dashboard template
	c.HTML(http.StatusOK, "dashboard.html", dashboardData)
}

// GetDashboardData returns dashboard data as JSON
func (h *DashboardHandler) GetDashboardData(c *gin.Context) {
	// Get system stats
	stats := h.monitorService.GetStats()
	
	// Get Moodle status
	moodleStatus := h.moodleService.GetStatus()

	// Prepare response
	response := gin.H{
		"system_stats":  stats,
		"moodle_status": moodleStatus,
		"timestamp":     stats.Timestamp,
	}

	c.JSON(http.StatusOK, response)
}

// GetSystemStats returns system statistics
func (h *DashboardHandler) GetSystemStats(c *gin.Context) {
	stats := h.monitorService.GetStats()
	c.JSON(http.StatusOK, stats)
}

// GetMoodleStatus returns Moodle status
func (h *DashboardHandler) GetMoodleStatus(c *gin.Context) {
	status := h.moodleService.GetStatus()
	c.JSON(http.StatusOK, status)
}

// GetAlerts returns system alerts
func (h *DashboardHandler) GetAlerts(c *gin.Context) {
	alerts, err := h.monitorService.GetAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get alerts",
		})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetSystemLogs returns system logs
func (h *DashboardHandler) GetSystemLogs(c *gin.Context) {
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
func (h *DashboardHandler) GetMoodleInfo(c *gin.Context) {
	info, err := h.moodleService.GetMoodleInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get Moodle info",
		})
		return
	}

	c.JSON(http.StatusOK, info)
}
