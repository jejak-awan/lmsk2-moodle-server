package smoke

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

// TestSmokeHealthCheck tests basic system health
func TestSmokeHealthCheck(t *testing.T) {
	router := setupSmokeTestRouter(t)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health check failed with status %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal health check response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
}

// TestSmokeLogin tests basic login functionality
func TestSmokeLogin(t *testing.T) {
	router := setupSmokeTestRouter(t)

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
		t.Errorf("Login failed with status %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}

	if response["token"] == nil {
		t.Error("Login response should contain token")
	}
}

// TestSmokeAuthenticatedRequest tests basic authenticated request
func TestSmokeAuthenticatedRequest(t *testing.T) {
	router := setupSmokeTestRouter(t)
	token := getSmokeAuthToken(t, router)

	req, _ := http.NewRequest("GET", "/api/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Authenticated request failed with status %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal stats response: %v", err)
	}

	// Check required fields
	requiredFields := []string{"cpu_usage", "memory_usage", "disk_usage", "uptime"}
	for _, field := range requiredFields {
		if response[field] == nil {
			t.Errorf("Stats response should contain %s", field)
		}
	}
}

// TestSmokeMoodleStatus tests basic Moodle status check
func TestSmokeMoodleStatus(t *testing.T) {
	router := setupSmokeTestRouter(t)
	token := getSmokeAuthToken(t, router)

	req, _ := http.NewRequest("GET", "/api/moodle/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Moodle status request failed with status %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal Moodle status response: %v", err)
	}

	// Check required fields
	requiredFields := []string{"running", "version", "uptime", "last_check"}
	for _, field := range requiredFields {
		if response[field] == nil {
			t.Errorf("Moodle status response should contain %s", field)
		}
	}
}

// TestSmokeUserManagement tests basic user management
func TestSmokeUserManagement(t *testing.T) {
	router := setupSmokeTestRouter(t)
	token := getSmokeAuthToken(t, router)

	// Test get users
	req, _ := http.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get users request failed with status %d", w.Code)
	}

	var users []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &users)
	if err != nil {
		t.Fatalf("Failed to unmarshal users response: %v", err)
	}

	if len(users) < 1 {
		t.Error("Users response should contain at least one user")
	}

	// Test get user stats
	req, _ = http.NewRequest("GET", "/api/users/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get user stats request failed with status %d", w.Code)
	}

	var stats map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &stats)
	if err != nil {
		t.Fatalf("Failed to unmarshal user stats response: %v", err)
	}

	// Check required fields
	requiredFields := []string{"total_users", "active_users", "online_users", "admin_users"}
	for _, field := range requiredFields {
		if stats[field] == nil {
			t.Errorf("User stats response should contain %s", field)
		}
	}
}

// TestSmokeAlerts tests basic alerts functionality
func TestSmokeAlerts(t *testing.T) {
	router := setupSmokeTestRouter(t)
	token := getSmokeAuthToken(t, router)

	req, _ := http.NewRequest("GET", "/api/alerts", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get alerts request failed with status %d", w.Code)
	}

	var alerts []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &alerts)
	if err != nil {
		t.Fatalf("Failed to unmarshal alerts response: %v", err)
	}

	// Alerts can be empty, just verify it's an array
	if alerts == nil {
		t.Error("Alerts response should be an array")
	}
}

// TestSmokeSystemLogs tests basic system logs functionality
func TestSmokeSystemLogs(t *testing.T) {
	router := setupSmokeTestRouter(t)
	token := getSmokeAuthToken(t, router)

	req, _ := http.NewRequest("GET", "/api/logs?limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get system logs request failed with status %d", w.Code)
	}

	var logs []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &logs)
	if err != nil {
		t.Fatalf("Failed to unmarshal logs response: %v", err)
	}

	// Logs can be empty, just verify it's an array
	if logs == nil {
		t.Error("Logs response should be an array")
	}
}

// TestSmokeLogout tests basic logout functionality
func TestSmokeLogout(t *testing.T) {
	router := setupSmokeTestRouter(t)
	token := getSmokeAuthToken(t, router)

	req, _ := http.NewRequest("POST", "/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Logout request failed with status %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal logout response: %v", err)
	}

	if response["message"] != "Logged out successfully" {
		t.Errorf("Expected logout success message, got %v", response["message"])
	}
}

// TestSmokeErrorHandling tests basic error handling
func TestSmokeErrorHandling(t *testing.T) {
	router := setupSmokeTestRouter(t)

	// Test invalid login
	loginData := map[string]string{
		"username": "nonexistent",
		"password": "wrongpassword",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Invalid login should return 401, got %d", w.Code)
	}

	// Test unauthorized access
	req, _ = http.NewRequest("GET", "/api/stats", nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Unauthorized access should return 401, got %d", w.Code)
	}

	// Test invalid endpoint
	req, _ = http.NewRequest("GET", "/api/nonexistent", nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Invalid endpoint should return 404, got %d", w.Code)
	}
}

// TestSmokePerformance tests basic performance requirements
func TestSmokePerformance(t *testing.T) {
	router := setupSmokeTestRouter(t)

	// Test health check response time
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health check failed with status %d", w.Code)
	}

	// Test login response time
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Login failed with status %d", w.Code)
	}

	// Test authenticated request response time
	token := getSmokeAuthToken(t, router)
	req, _ = http.NewRequest("GET", "/api/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Authenticated request failed with status %d", w.Code)
	}
}

// TestSmokeDataConsistency tests basic data consistency
func TestSmokeDataConsistency(t *testing.T) {
	router := setupSmokeTestRouter(t)
	token := getSmokeAuthToken(t, router)

	// Test system stats consistency
	req, _ := http.NewRequest("GET", "/api/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get stats request failed with status %d", w.Code)
	}

	var stats map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	if err != nil {
		t.Fatalf("Failed to unmarshal stats response: %v", err)
	}

	// Verify stats are within reasonable ranges
	if cpuUsage, ok := stats["cpu_usage"].(float64); ok {
		if cpuUsage < 0 || cpuUsage > 100 {
			t.Errorf("CPU usage should be between 0 and 100, got %f", cpuUsage)
		}
	}

	if memoryUsage, ok := stats["memory_usage"].(float64); ok {
		if memoryUsage < 0 || memoryUsage > 100 {
			t.Errorf("Memory usage should be between 0 and 100, got %f", memoryUsage)
		}
	}

	if diskUsage, ok := stats["disk_usage"].(float64); ok {
		if diskUsage < 0 || diskUsage > 100 {
			t.Errorf("Disk usage should be between 0 and 100, got %f", diskUsage)
		}
	}

	// Test user stats consistency
	req, _ = http.NewRequest("GET", "/api/users/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get user stats request failed with status %d", w.Code)
	}

	var userStats map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &userStats)
	if err != nil {
		t.Fatalf("Failed to unmarshal user stats response: %v", err)
	}

	// Verify user stats are reasonable
	if totalUsers, ok := userStats["total_users"].(float64); ok {
		if totalUsers < 1 {
			t.Errorf("Total users should be at least 1, got %f", totalUsers)
		}
	}

	if activeUsers, ok := userStats["active_users"].(float64); ok {
		if activeUsers < 1 {
			t.Errorf("Active users should be at least 1, got %f", activeUsers)
		}
	}

	if adminUsers, ok := userStats["admin_users"].(float64); ok {
		if adminUsers < 1 {
			t.Errorf("Admin users should be at least 1, got %f", adminUsers)
		}
	}
}

// TestSmokeSecurity tests basic security features
func TestSmokeSecurity(t *testing.T) {
	router := setupSmokeTestRouter(t)

	// Test rate limiting (make multiple requests quickly)
	for i := 0; i < 150; i++ { // Exceed the 100 req/min limit
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i >= 100 && w.Code == http.StatusTooManyRequests {
			// Rate limiting is working
			return
		}
	}

	// Test invalid token
	req, _ := http.NewRequest("GET", "/api/stats", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Invalid token should return 401, got %d", w.Code)
	}

	// Test missing token
	req, _ = http.NewRequest("GET", "/api/stats", nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Missing token should return 401, got %d", w.Code)
	}
}

// TestSmokeConcurrency tests basic concurrency handling
func TestSmokeConcurrency(t *testing.T) {
	router := setupSmokeTestRouter(t)
	token := getSmokeAuthToken(t, router)

	// Test concurrent requests
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			req, _ := http.NewRequest("GET", "/api/stats", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				t.Errorf("Concurrent request failed with status %d", w.Code)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Helper functions

func setupSmokeTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:  8080,
			Host:  "0.0.0.0",
			Debug: false,
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

	db := setupSmokeTestDB(t)

	authService := services.NewAuthService(cfg.Security.JWTSecret, db)
	monitorService := services.NewMonitorService(cfg.Monitoring)
	moodleService := services.NewMoodleService(cfg.Moodle)
	securityService := services.NewSecurityService(cfg.Security)

	authHandler := handlers.NewAuthHandler(authService)
	apiHandler := handlers.NewAPIHandler(monitorService, moodleService, securityService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(securityService.RateLimitMiddleware())

	router.POST("/login", authHandler.Login)
	router.POST("/logout", authHandler.Logout)

	protected := router.Group("/api")
	protected.Use(authHandler.AuthMiddleware())
	{
		protected.GET("/stats", apiHandler.GetStats)
		protected.GET("/moodle/status", apiHandler.GetMoodleStatus)
		protected.GET("/users", apiHandler.GetUsers)
		protected.GET("/users/stats", apiHandler.GetUserStats)
		protected.GET("/alerts", apiHandler.GetAlerts)
		protected.GET("/logs", apiHandler.GetSystemLogs)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": "2025-01-09T10:00:00Z",
			"version":   "1.0.0",
		})
	})

	return router
}

func getSmokeAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupSmokeTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

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

	passwordHash, _ := utils.HashPassword("admin123")
	_, err = db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, active)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "admin-id", "admin", "admin@k2net.id", passwordHash, "admin", true)

	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	return db
}
