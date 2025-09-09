package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"lms-manager/config"
	"lms-manager/handlers"
	"lms-manager/services"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// TestEndToEndWorkflow tests the complete user workflow
func TestEndToEndWorkflow(t *testing.T) {
	router := setupE2ETestRouter(t)

	// Step 1: Test health check
	t.Run("HealthCheck", func(t *testing.T) {
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
	})

	// Step 2: Test login
	t.Run("Login", func(t *testing.T) {
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

		if response["user"] == nil {
			t.Error("Login response should contain user")
		}
	})

	// Step 3: Test authenticated requests
	t.Run("AuthenticatedRequests", func(t *testing.T) {
		token := getE2EAuthToken(t, router)

		// Test get stats
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get stats failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check required fields
		requiredFields := []string{"cpu_usage", "memory_usage", "disk_usage", "uptime"}
		for _, field := range requiredFields {
			if stats[field] == nil {
				t.Errorf("Stats response should contain %s", field)
			}
		}

		// Test get Moodle status
		req, _ = http.NewRequest("GET", "/api/moodle/status", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get Moodle status failed with status %d", w.Code)
		}

		var moodleStatus map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &moodleStatus)
		if err != nil {
			t.Fatalf("Failed to unmarshal Moodle status response: %v", err)
		}

		// Check required fields
		requiredFields = []string{"running", "version", "uptime", "last_check"}
		for _, field := range requiredFields {
			if moodleStatus[field] == nil {
				t.Errorf("Moodle status response should contain %s", field)
			}
		}

		// Test get users
		req, _ = http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get users failed with status %d", w.Code)
		}

		var users []map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &users)
		if err != nil {
			t.Fatalf("Failed to unmarshal users response: %v", err)
		}

		if len(users) < 1 {
			t.Error("Users response should contain at least one user")
		}

		// Test get alerts
		req, _ = http.NewRequest("GET", "/api/alerts", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get alerts failed with status %d", w.Code)
		}

		var alerts []map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &alerts)
		if err != nil {
			t.Fatalf("Failed to unmarshal alerts response: %v", err)
		}

		// Alerts can be empty, so we just check that it's an array
		if alerts == nil {
			t.Error("Alerts response should be an array")
		}

		// Test get system logs
		req, _ = http.NewRequest("GET", "/api/logs?limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get system logs failed with status %d", w.Code)
		}

		var logs []map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &logs)
		if err != nil {
			t.Fatalf("Failed to unmarshal logs response: %v", err)
		}

		// Logs can be empty, so we just check that it's an array
		if logs == nil {
			t.Error("Logs response should be an array")
		}
	})

	// Step 4: Test Moodle management operations
	t.Run("MoodleManagement", func(t *testing.T) {
		token := getE2EAuthToken(t, router)

		// Test start Moodle (might fail in test environment)
		req, _ := http.NewRequest("POST", "/api/moodle/start", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Accept both success and failure (Moodle might not be installed in test environment)
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("Start Moodle failed with unexpected status %d", w.Code)
		}

		// Test stop Moodle (might fail in test environment)
		req, _ = http.NewRequest("POST", "/api/moodle/stop", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Accept both success and failure (Moodle might not be installed in test environment)
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("Stop Moodle failed with unexpected status %d", w.Code)
		}

		// Test restart Moodle (might fail in test environment)
		req, _ = http.NewRequest("POST", "/api/moodle/restart", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Accept both success and failure (Moodle might not be installed in test environment)
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("Restart Moodle failed with unexpected status %d", w.Code)
		}
	})

	// Step 5: Test user management
	t.Run("UserManagement", func(t *testing.T) {
		token := getE2EAuthToken(t, router)

		// Test create user
		userData := map[string]interface{}{
			"username": "testuser",
			"email":    "test@example.com",
			"password": "TestPass123!",
			"role":     "operator",
		}

		jsonData, _ := json.Marshal(userData)
		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Create user failed with status %d", w.Code)
		}

		var user map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &user)
		if err != nil {
			t.Fatalf("Failed to unmarshal user response: %v", err)
		}

		if user["username"] != "testuser" {
			t.Errorf("Expected username 'testuser', got %v", user["username"])
		}

		// Test get user stats
		req, _ = http.NewRequest("GET", "/api/users/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get user stats failed with status %d", w.Code)
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
	})

	// Step 6: Test logout
	t.Run("Logout", func(t *testing.T) {
		token := getE2EAuthToken(t, router)

		req, _ := http.NewRequest("POST", "/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Logout failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal logout response: %v", err)
		}

		if response["message"] != "Logged out successfully" {
			t.Errorf("Expected logout message, got %v", response["message"])
		}
	})
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	router := setupE2ETestRouter(t)

	// Test invalid login
	t.Run("InvalidLogin", func(t *testing.T) {
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
	})

	// Test unauthorized access
	t.Run("UnauthorizedAccess", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/stats", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Unauthorized access should return 401, got %d", w.Code)
		}
	})

	// Test invalid token
	t.Run("InvalidToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Invalid token should return 401, got %d", w.Code)
		}
	})

	// Test invalid endpoint
	t.Run("InvalidEndpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nonexistent", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Invalid endpoint should return 404, got %d", w.Code)
		}
	})

	// Test invalid method
	t.Run("InvalidMethod", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/health", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Invalid method should return 405, got %d", w.Code)
		}
	})
}

// TestPerformanceBaseline tests basic performance requirements
func TestPerformanceBaseline(t *testing.T) {
	router := setupE2ETestRouter(t)

	// Test response time for health check
	t.Run("HealthCheckPerformance", func(t *testing.T) {
		start := time.Now()
		
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		duration := time.Since(start)
		
		if w.Code != http.StatusOK {
			t.Errorf("Health check failed with status %d", w.Code)
		}
		
		if duration > 100*time.Millisecond {
			t.Errorf("Health check response time too slow: %v", duration)
		}
	})

	// Test response time for authenticated request
	t.Run("AuthenticatedRequestPerformance", func(t *testing.T) {
		token := getE2EAuthToken(t, router)
		
		start := time.Now()
		
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		duration := time.Since(start)
		
		if w.Code != http.StatusOK {
			t.Errorf("Get stats failed with status %d", w.Code)
		}
		
		if duration > 200*time.Millisecond {
			t.Errorf("Get stats response time too slow: %v", duration)
		}
	})
}

// TestDataConsistency tests data consistency across operations
func TestDataConsistency(t *testing.T) {
	router := setupE2ETestRouter(t)
	token := getE2EAuthToken(t, router)

	// Test user creation and retrieval
	t.Run("UserDataConsistency", func(t *testing.T) {
		// Create user
		userData := map[string]interface{}{
			"username": "consistencyuser",
			"email":    "consistency@example.com",
			"password": "TestPass123!",
			"role":     "viewer",
		}

		jsonData, _ := json.Marshal(userData)
		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Create user failed with status %d", w.Code)
		}

		var createdUser map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &createdUser)
		if err != nil {
			t.Fatalf("Failed to unmarshal created user response: %v", err)
		}

		// Get all users and verify the created user exists
		req, _ = http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get users failed with status %d", w.Code)
		}

		var users []map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &users)
		if err != nil {
			t.Fatalf("Failed to unmarshal users response: %v", err)
		}

		// Find the created user
		var foundUser map[string]interface{}
		for _, user := range users {
			if user["username"] == "consistencyuser" {
				foundUser = user
				break
			}
		}

		if foundUser == nil {
			t.Error("Created user not found in users list")
		}

		// Verify data consistency
		if foundUser["email"] != "consistency@example.com" {
			t.Errorf("Email mismatch: expected 'consistency@example.com', got %v", foundUser["email"])
		}

		if foundUser["role"] != "viewer" {
			t.Errorf("Role mismatch: expected 'viewer', got %v", foundUser["role"])
		}
	})
}

// Helper functions

func setupE2ETestRouter(t *testing.T) *gin.Engine {
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

	db := setupE2ETestDB(t)

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
		protected.POST("/moodle/start", apiHandler.StartMoodle)
		protected.POST("/moodle/stop", apiHandler.StopMoodle)
		protected.POST("/moodle/restart", apiHandler.RestartMoodle)
		protected.GET("/users", apiHandler.GetUsers)
		protected.POST("/users", apiHandler.CreateUser)
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

func getE2EAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupE2ETestDB(t *testing.T) *sql.DB {
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
