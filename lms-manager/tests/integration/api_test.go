package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"lms-manager/config"
	"lms-manager/handlers"
	"lms-manager/services"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func TestAPI_Login(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Test login request
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["token"] == nil {
		t.Error("Response should contain token")
	}

	if response["user"] == nil {
		t.Error("Response should contain user")
	}
}

func TestAPI_GetStats(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Get auth token
	token := getAuthToken(t, router)

	// Test get stats request
	req, _ := http.NewRequest("GET", "/api/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check required fields
	requiredFields := []string{"cpu_usage", "memory_usage", "disk_usage", "uptime"}
	for _, field := range requiredFields {
		if response[field] == nil {
			t.Errorf("Response should contain %s", field)
		}
	}
}

func TestAPI_GetMoodleStatus(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Get auth token
	token := getAuthToken(t, router)

	// Test get Moodle status request
	req, _ := http.NewRequest("GET", "/api/moodle/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check required fields
	requiredFields := []string{"running", "version", "uptime", "last_check"}
	for _, field := range requiredFields {
		if response[field] == nil {
			t.Errorf("Response should contain %s", field)
		}
	}
}

func TestAPI_StartMoodle(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Get auth token
	token := getAuthToken(t, router)

	// Test start Moodle request
	req, _ := http.NewRequest("POST", "/api/moodle/start", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This might fail in test environment if Moodle is not installed
	// We're just testing the API endpoint structure
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestAPI_StopMoodle(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Get auth token
	token := getAuthToken(t, router)

	// Test stop Moodle request
	req, _ := http.NewRequest("POST", "/api/moodle/stop", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This might fail in test environment if Moodle is not installed
	// We're just testing the API endpoint structure
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestAPI_RestartMoodle(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Get auth token
	token := getAuthToken(t, router)

	// Test restart Moodle request
	req, _ := http.NewRequest("POST", "/api/moodle/restart", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This might fail in test environment if Moodle is not installed
	// We're just testing the API endpoint structure
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 200 or 500, got %d", w.Code)
	}
}

func TestAPI_GetUsers(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Get auth token
	token := getAuthToken(t, router)

	// Test get users request
	req, _ := http.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should have at least the default admin user
	if len(response) < 1 {
		t.Error("Response should contain at least one user")
	}
}

func TestAPI_GetUserStats(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Get auth token
	token := getAuthToken(t, router)

	// Test get user stats request
	req, _ := http.NewRequest("GET", "/api/users/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check required fields
	requiredFields := []string{"total_users", "active_users", "online_users", "admin_users"}
	for _, field := range requiredFields {
		if response[field] == nil {
			t.Errorf("Response should contain %s", field)
		}
	}
}

func TestAPI_GetAlerts(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Get auth token
	token := getAuthToken(t, router)

	// Test get alerts request
	req, _ := http.NewRequest("GET", "/api/alerts", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Response should be an array (even if empty)
	if response == nil {
		t.Error("Response should be an array")
	}
}

func TestAPI_GetSystemLogs(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Get auth token
	token := getAuthToken(t, router)

	// Test get system logs request
	req, _ := http.NewRequest("GET", "/api/logs?limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Response should be an array (even if empty)
	if response == nil {
		t.Error("Response should be an array")
	}
}

func TestAPI_HealthCheck(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Test health check request
	req, _ := http.NewRequest("GET", "/health", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check required fields
	requiredFields := []string{"status", "timestamp", "version"}
	for _, field := range requiredFields {
		if response[field] == nil {
			t.Errorf("Response should contain %s", field)
		}
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
}

func TestAPI_UnauthorizedAccess(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Test unauthorized access to protected endpoint
	req, _ := http.NewRequest("GET", "/api/stats", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAPI_InvalidToken(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	router := setupTestRouter(t)

	// Test with invalid token
	req, _ := http.NewRequest("GET", "/api/stats", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// Helper functions

func setupTestRouter(t *testing.T) *gin.Engine {
	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:  8080,
			Host:  "0.0.0.0",
			Debug: true,
		},
		Moodle: config.MoodleConfig{
			Path:       "/tmp/moodle",
			ConfigPath: "/tmp/moodle/config.php",
			DataPath:   "/tmp/moodledata",
		},
		Security: config.SecurityConfig{
			JWTSecret:     "test-secret-key",
			SessionTimeout: 3600,
			RateLimit:     100,
			AllowedIPs:   []string{"127.0.0.1"},
		},
		Monitoring: config.MonitoringConfig{
			UpdateInterval: 30,
			LogRetention:   7,
			AlertThresholds: config.AlertThresholdsConfig{
				CPU:    80.0,
				Memory: 85.0,
				Disk:   90.0,
			},
		},
	}

	// Create test database
	db := setupTestDB(t)

	// Initialize services
	authService := services.NewAuthService(cfg.Security.JWTSecret, db)
	monitorService := services.NewMonitorService(cfg.Monitoring)
	moodleService := services.NewMoodleService(cfg.Moodle)
	securityService := services.NewSecurityService(cfg.Security)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	dashboardHandler := handlers.NewDashboardHandler(monitorService, moodleService)
	apiHandler := handlers.NewAPIHandler(monitorService, moodleService, securityService)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// Public routes
	router.POST("/login", authHandler.Login)

	// Protected routes
	protected := router.Group("/api")
	protected.Use(authHandler.AuthMiddleware())
	{
		protected.GET("/stats", apiHandler.GetStats)
		protected.GET("/moodle/status", apiHandler.GetMoodleStatus)
		protected.POST("/moodle/start", apiHandler.StartMoodle)
		protected.POST("/moodle/stop", apiHandler.StopMoodle)
		protected.POST("/moodle/restart", apiHandler.RestartMoodle)
		protected.GET("/users", apiHandler.GetUsers)
		protected.GET("/users/stats", apiHandler.GetUserStats)
		protected.GET("/alerts", apiHandler.GetAlerts)
		protected.GET("/logs", apiHandler.GetSystemLogs)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": "2025-01-09T10:00:00Z",
			"version":   "1.0.0",
		})
	})

	return router
}

func getAuthToken(t *testing.T, router *gin.Engine) string {
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Login failed with status %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}

	token, ok := response["token"].(string)
	if !ok {
		t.Fatal("Token not found in login response")
	}

	return token
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS user_sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			user_agent TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_activities (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			action TEXT NOT NULL,
			resource TEXT,
			ip_address TEXT,
			user_agent TEXT,
			success BOOLEAN DEFAULT 1,
			message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id)
		)`,
		`CREATE TABLE IF NOT EXISTS security_events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			message TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			severity TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			message TEXT NOT NULL,
			severity TEXT NOT NULL,
			resolved BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			resolved_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS system_logs (
			id TEXT PRIMARY KEY,
			level TEXT NOT NULL,
			message TEXT NOT NULL,
			source TEXT,
			data TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	// Create default admin user
	passwordHash, err := utils.HashPassword("admin123")
	if err != nil {
		t.Fatalf("Failed to hash admin password: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, active)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "admin-id", "admin", "admin@k2net.id", passwordHash, "admin", true)

	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	return db
}
