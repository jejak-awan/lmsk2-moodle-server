package uat

import (
	"bytes"
	"encoding/json"
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

// TestUserAcceptanceWorkflow tests the complete user acceptance workflow
func TestUserAcceptanceWorkflow(t *testing.T) {
	router, db := setupUATTestSystem(t)
	defer db.Close()

	// Test complete user workflow
	t.Run("CompleteUserWorkflow", func(t *testing.T) {
		// Step 1: System Health Check
		t.Run("SystemHealthCheck", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("System health check failed with status %d", w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal health check response: %v", err)
			}

			if response["status"] != "healthy" {
				t.Errorf("System should be healthy, got status: %v", response["status"])
			}

			if response["version"] == nil {
				t.Error("System should report version information")
			}
		})

		// Step 2: User Authentication
		t.Run("UserAuthentication", func(t *testing.T) {
			// Test successful login
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
				t.Errorf("User login failed with status %d", w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal login response: %v", err)
			}

			if response["token"] == nil {
				t.Error("Login should return authentication token")
			}

			if response["user"] == nil {
				t.Error("Login should return user information")
			}

			user, ok := response["user"].(map[string]interface{})
			if !ok {
				t.Fatal("User information should be an object")
			}

			if user["username"] != "admin" {
				t.Errorf("Expected username 'admin', got %v", user["username"])
			}

			if user["role"] != "admin" {
				t.Errorf("Expected role 'admin', got %v", user["role"])
			}
		})

		// Step 3: Dashboard Access
		t.Run("DashboardAccess", func(t *testing.T) {
			token := getUATAuthToken(t, router)

			// Test system statistics
			req, _ := http.NewRequest("GET", "/api/stats", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("System statistics access failed with status %d", w.Code)
			}

			var stats map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &stats)
			if err != nil {
				t.Fatalf("Failed to unmarshal stats response: %v", err)
			}

			// Verify system statistics are present and reasonable
			requiredFields := []string{"cpu_usage", "memory_usage", "disk_usage", "uptime"}
			for _, field := range requiredFields {
				if stats[field] == nil {
					t.Errorf("System statistics should contain %s", field)
				}
			}

			// Verify CPU usage is within reasonable range
			if cpuUsage, ok := stats["cpu_usage"].(float64); ok {
				if cpuUsage < 0 || cpuUsage > 100 {
					t.Errorf("CPU usage should be between 0 and 100, got %f", cpuUsage)
				}
			}

			// Verify memory usage is within reasonable range
			if memoryUsage, ok := stats["memory_usage"].(float64); ok {
				if memoryUsage < 0 || memoryUsage > 100 {
					t.Errorf("Memory usage should be between 0 and 100, got %f", memoryUsage)
				}
			}

			// Verify disk usage is within reasonable range
			if diskUsage, ok := stats["disk_usage"].(float64); ok {
				if diskUsage < 0 || diskUsage > 100 {
					t.Errorf("Disk usage should be between 0 and 100, got %f", diskUsage)
				}
			}
		})

		// Step 4: Moodle Management
		t.Run("MoodleManagement", func(t *testing.T) {
			token := getUATAuthToken(t, router)

			// Test Moodle status check
			req, _ := http.NewRequest("GET", "/api/moodle/status", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Moodle status check failed with status %d", w.Code)
			}

			var status map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &status)
			if err != nil {
				t.Fatalf("Failed to unmarshal Moodle status response: %v", err)
			}

			// Verify Moodle status information
			requiredFields := []string{"running", "version", "uptime", "last_check"}
			for _, field := range requiredFields {
				if status[field] == nil {
					t.Errorf("Moodle status should contain %s", field)
				}
			}

			// Test Moodle operations (might fail in test environment)
			moodleOperations := []string{
				"/api/moodle/start",
				"/api/moodle/stop",
				"/api/moodle/restart",
			}

			for _, operation := range moodleOperations {
				req, _ = http.NewRequest("POST", operation, nil)
				req.Header.Set("Authorization", "Bearer "+token)

				w = httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Accept both success and failure (Moodle might not be installed)
				if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
					t.Errorf("Moodle operation %s failed with unexpected status %d", operation, w.Code)
				}
			}
		})

		// Step 5: User Management
		t.Run("UserManagement", func(t *testing.T) {
			token := getUATAuthToken(t, router)

			// Test user creation
			userData := map[string]interface{}{
				"username": "testoperator",
				"email":    "operator@example.com",
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
				t.Errorf("User creation failed with status %d", w.Code)
			}

			var createdUser map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &createdUser)
			if err != nil {
				t.Fatalf("Failed to unmarshal user creation response: %v", err)
			}

			if createdUser["username"] != "testoperator" {
				t.Errorf("Expected username 'testoperator', got %v", createdUser["username"])
			}

			if createdUser["email"] != "operator@example.com" {
				t.Errorf("Expected email 'operator@example.com', got %v", createdUser["email"])
			}

			if createdUser["role"] != "operator" {
				t.Errorf("Expected role 'operator', got %v", createdUser["role"])
			}

			// Test user listing
			req, _ = http.NewRequest("GET", "/api/users", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("User listing failed with status %d", w.Code)
			}

			var users []map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &users)
			if err != nil {
				t.Fatalf("Failed to unmarshal users response: %v", err)
			}

			if len(users) < 2 { // Should have at least admin and created user
				t.Errorf("Expected at least 2 users, got %d", len(users))
			}

			// Test user statistics
			req, _ = http.NewRequest("GET", "/api/users/stats", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("User statistics failed with status %d", w.Code)
			}

			var stats map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &stats)
			if err != nil {
				t.Fatalf("Failed to unmarshal user stats response: %v", err)
			}

			// Verify user statistics
			requiredFields := []string{"total_users", "active_users", "online_users", "admin_users"}
			for _, field := range requiredFields {
				if stats[field] == nil {
					t.Errorf("User statistics should contain %s", field)
				}
			}

			if totalUsers, ok := stats["total_users"].(float64); ok {
				if totalUsers < 2 {
					t.Errorf("Expected at least 2 total users, got %f", totalUsers)
				}
			}
		})

		// Step 6: System Monitoring
		t.Run("SystemMonitoring", func(t *testing.T) {
			token := getUATAuthToken(t, router)

			// Test alerts
			req, _ := http.NewRequest("GET", "/api/alerts", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Alerts access failed with status %d", w.Code)
			}

			var alerts []map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &alerts)
			if err != nil {
				t.Fatalf("Failed to unmarshal alerts response: %v", err)
			}

			// Alerts can be empty, just verify the structure
			for _, alert := range alerts {
				if alert["type"] == nil {
					t.Error("Alert should contain type")
				}
				if alert["message"] == nil {
					t.Error("Alert should contain message")
				}
				if alert["severity"] == nil {
					t.Error("Alert should contain severity")
				}
				if alert["timestamp"] == nil {
					t.Error("Alert should contain timestamp")
				}
			}

			// Test system logs
			req, _ = http.NewRequest("GET", "/api/logs?limit=10", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("System logs access failed with status %d", w.Code)
			}

			var logs []map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &logs)
			if err != nil {
				t.Fatalf("Failed to unmarshal logs response: %v", err)
			}

			// Logs can be empty, just verify the structure
			for _, log := range logs {
				if log["level"] == nil {
					t.Error("Log should contain level")
				}
				if log["message"] == nil {
					t.Error("Log should contain message")
				}
				if log["timestamp"] == nil {
					t.Error("Log should contain timestamp")
				}
			}
		})

		// Step 7: User Logout
		t.Run("UserLogout", func(t *testing.T) {
			token := getUATAuthToken(t, router)

			req, _ := http.NewRequest("POST", "/logout", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("User logout failed with status %d", w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal logout response: %v", err)
			}

			if response["message"] != "Logged out successfully" {
				t.Errorf("Expected logout success message, got %v", response["message"])
			}
		})
	})
}

// TestUserAcceptanceErrorScenarios tests error scenarios from user perspective
func TestUserAcceptanceErrorScenarios(t *testing.T) {
	router, db := setupUATTestSystem(t)
	defer db.Close()

	t.Run("ErrorScenarios", func(t *testing.T) {
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

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal error response: %v", err)
			}

			if response["error"] == nil {
				t.Error("Error response should contain error message")
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

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal error response: %v", err)
			}

			if response["error"] == nil {
				t.Error("Error response should contain error message")
			}
		})

		// Test invalid user creation
		t.Run("InvalidUserCreation", func(t *testing.T) {
			token := getUATAuthToken(t, router)

			// Test with invalid data
			userData := map[string]interface{}{
				"username": "", // Empty username
				"email":    "invalid-email",
				"password": "weak",
				"role":     "invalid-role",
			}

			jsonData, _ := json.Marshal(userData)
			req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Invalid user creation should return 400, got %d", w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal error response: %v", err)
			}

			if response["error"] == nil {
				t.Error("Error response should contain error message")
			}
		})
	})
}

// TestUserAcceptancePerformance tests performance from user perspective
func TestUserAcceptancePerformance(t *testing.T) {
	router, db := setupUATTestSystem(t)
	defer db.Close()

	t.Run("PerformanceRequirements", func(t *testing.T) {
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

		// Test response time for login
		t.Run("LoginPerformance", func(t *testing.T) {
			start := time.Now()
			
			loginData := map[string]string{
				"username": "admin",
				"password": "admin123",
			}

			jsonData, _ := json.Marshal(loginData)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			duration := time.Since(start)
			
			if w.Code != http.StatusOK {
				t.Errorf("Login failed with status %d", w.Code)
			}
			
			if duration > 500*time.Millisecond {
				t.Errorf("Login response time too slow: %v", duration)
			}
		})

		// Test response time for authenticated requests
		t.Run("AuthenticatedRequestPerformance", func(t *testing.T) {
			token := getUATAuthToken(t, router)
			
			endpoints := []string{
				"/api/stats",
				"/api/moodle/status",
				"/api/users",
				"/api/alerts",
			}

			for _, endpoint := range endpoints {
				start := time.Now()
				
				req, _ := http.NewRequest("GET", endpoint, nil)
				req.Header.Set("Authorization", "Bearer "+token)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				duration := time.Since(start)
				
				if w.Code != http.StatusOK {
					t.Errorf("Endpoint %s failed with status %d", endpoint, w.Code)
				}
				
				if duration > 200*time.Millisecond {
					t.Errorf("Endpoint %s response time too slow: %v", endpoint, duration)
				}
			}
		})
	})
}

// TestUserAcceptanceDataIntegrity tests data integrity from user perspective
func TestUserAcceptanceDataIntegrity(t *testing.T) {
	router, db := setupUATTestSystem(t)
	defer db.Close()

	t.Run("DataIntegrity", func(t *testing.T) {
		token := getUATAuthToken(t, router)

		// Test user creation and data consistency
		t.Run("UserDataConsistency", func(t *testing.T) {
			// Create user
			userData := map[string]interface{}{
				"username": "integrityuser",
				"email":    "integrity@example.com",
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
				t.Errorf("User creation failed with status %d", w.Code)
			}

			var createdUser map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &createdUser)
			if err != nil {
				t.Fatalf("Failed to unmarshal user creation response: %v", err)
			}

			// Verify created user data
			if createdUser["username"] != "integrityuser" {
				t.Errorf("Username mismatch: expected 'integrityuser', got %v", createdUser["username"])
			}

			if createdUser["email"] != "integrity@example.com" {
				t.Errorf("Email mismatch: expected 'integrity@example.com', got %v", createdUser["email"])
			}

			if createdUser["role"] != "viewer" {
				t.Errorf("Role mismatch: expected 'viewer', got %v", createdUser["role"])
			}

			// Verify user appears in user list
			req, _ = http.NewRequest("GET", "/api/users", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("User listing failed with status %d", w.Code)
			}

			var users []map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &users)
			if err != nil {
				t.Fatalf("Failed to unmarshal users response: %v", err)
			}

			// Find the created user
			var foundUser map[string]interface{}
			for _, user := range users {
				if user["username"] == "integrityuser" {
					foundUser = user
					break
				}
			}

			if foundUser == nil {
				t.Error("Created user not found in user list")
			} else {
				// Verify data consistency
				if foundUser["email"] != "integrity@example.com" {
					t.Errorf("Email inconsistency: expected 'integrity@example.com', got %v", foundUser["email"])
				}

				if foundUser["role"] != "viewer" {
					t.Errorf("Role inconsistency: expected 'viewer', got %v", foundUser["role"])
				}
			}
		})

		// Test system statistics consistency
		t.Run("SystemStatisticsConsistency", func(t *testing.T) {
			// Get stats multiple times and verify consistency
			var prevStats map[string]interface{}
			for i := 0; i < 3; i++ {
				req, _ := http.NewRequest("GET", "/api/stats", nil)
				req.Header.Set("Authorization", "Bearer "+token)

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Stats request failed with status %d", w.Code)
				}

				var stats map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &stats)
				if err != nil {
					t.Fatalf("Failed to unmarshal stats response: %v", err)
				}

				// Verify stats are within reasonable ranges
				if cpuUsage, ok := stats["cpu_usage"].(float64); ok {
					if cpuUsage < 0 || cpuUsage > 100 {
						t.Errorf("CPU usage out of range: %f", cpuUsage)
					}
				}

				if memoryUsage, ok := stats["memory_usage"].(float64); ok {
					if memoryUsage < 0 || memoryUsage > 100 {
						t.Errorf("Memory usage out of range: %f", memoryUsage)
					}
				}

				if diskUsage, ok := stats["disk_usage"].(float64); ok {
					if diskUsage < 0 || diskUsage > 100 {
						t.Errorf("Disk usage out of range: %f", diskUsage)
					}
				}

				// Verify uptime is increasing
				if prevStats != nil {
					if prevUptime, ok := prevStats["uptime"].(float64); ok {
						if currentUptime, ok := stats["uptime"].(float64); ok {
							if currentUptime < prevUptime {
								t.Error("Uptime should be increasing")
							}
						}
					}
				}

				prevStats = stats
				time.Sleep(1 * time.Second)
			}
		})
	})
}

// Helper functions

func setupUATTestSystem(t *testing.T) (*gin.Engine, *sql.DB) {
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

	db := setupUATTestDB(t)

	authService := services.NewAuthService(cfg.Security.JWTSecret, db)
	monitorService := services.NewMonitorService(cfg.Monitoring)
	moodleService := services.NewMoodleService(cfg.Moodle)
	securityService := services.NewSecurityService(cfg.Security)

	authHandler := handlers.NewAuthHandler(authService)
	dashboardHandler := handlers.NewDashboardHandler(monitorService, moodleService)
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

	return router, db
}

func getUATAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupUATTestDB(t *testing.T) *sql.DB {
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
