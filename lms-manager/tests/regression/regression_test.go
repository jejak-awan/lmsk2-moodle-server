package regression

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

// TestRegressionAuthentication tests authentication regression
func TestRegressionAuthentication(t *testing.T) {
	router := setupRegressionTestRouter(t)

	// Test 1: Valid login should work
	t.Run("ValidLogin", func(t *testing.T) {
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
			t.Errorf("Valid login should return 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		if response["token"] == nil {
			t.Error("Valid login should return token")
		}
	})

	// Test 2: Invalid login should fail
	t.Run("InvalidLogin", func(t *testing.T) {
		loginData := map[string]string{
			"username": "admin",
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

	// Test 3: Empty credentials should fail
	t.Run("EmptyCredentials", func(t *testing.T) {
		loginData := map[string]string{
			"username": "",
			"password": "",
		}

		jsonData, _ := json.Marshal(loginData)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Empty credentials should return 400, got %d", w.Code)
		}
	})

	// Test 4: Token validation should work
	t.Run("TokenValidation", func(t *testing.T) {
		token := getRegressionAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Valid token should allow access, got %d", w.Code)
		}
	})

	// Test 5: Invalid token should be rejected
	t.Run("InvalidToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Invalid token should return 401, got %d", w.Code)
		}
	})

	// Test 6: Missing token should be rejected
	t.Run("MissingToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/stats", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Missing token should return 401, got %d", w.Code)
		}
	})
}

// TestRegressionUserManagement tests user management regression
func TestRegressionUserManagement(t *testing.T) {
	router := setupRegressionTestRouter(t)
	token := getRegressionAuthToken(t, router)

	// Test 1: Get users should work
	t.Run("GetUsers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get users should return 200, got %d", w.Code)
		}

		var users []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &users)
		if err != nil {
			t.Fatalf("Failed to unmarshal users response: %v", err)
		}

		if len(users) < 1 {
			t.Error("Should have at least one user")
		}
	})

	// Test 2: Create user should work
	t.Run("CreateUser", func(t *testing.T) {
		userData := map[string]interface{}{
			"username": "regressionuser",
			"email":    "regression@example.com",
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
			t.Errorf("Create user should return 201, got %d", w.Code)
		}

		var user map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &user)
		if err != nil {
			t.Fatalf("Failed to unmarshal user response: %v", err)
		}

		if user["username"] != "regressionuser" {
			t.Errorf("Expected username 'regressionuser', got %v", user["username"])
		}
	})

	// Test 3: Duplicate username should fail
	t.Run("DuplicateUsername", func(t *testing.T) {
		userData := map[string]interface{}{
			"username": "admin", // Already exists
			"email":    "duplicate@example.com",
			"password": "TestPass123!",
			"role":     "viewer",
		}

		jsonData, _ := json.Marshal(userData)
		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Duplicate username should return 400, got %d", w.Code)
		}
	})

	// Test 4: Invalid user data should fail
	t.Run("InvalidUserData", func(t *testing.T) {
		userData := map[string]interface{}{
			"username": "ab", // Too short
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
			t.Errorf("Invalid user data should return 400, got %d", w.Code)
		}
	})

	// Test 5: Get user stats should work
	t.Run("GetUserStats", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/users/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get user stats should return 200, got %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal user stats response: %v", err)
		}

		if stats["total_users"] == nil {
			t.Error("User stats should contain total_users")
		}
	})
}

// TestRegressionSystemMonitoring tests system monitoring regression
func TestRegressionSystemMonitoring(t *testing.T) {
	router := setupRegressionTestRouter(t)
	token := getRegressionAuthToken(t, router)

	// Test 1: Get system stats should work
	t.Run("GetSystemStats", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get system stats should return 200, got %d", w.Code)
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
				t.Errorf("System stats should contain %s", field)
			}
		}

		// Verify values are within reasonable ranges
		if cpuUsage, ok := stats["cpu_usage"].(float64); ok {
			if cpuUsage < 0 || cpuUsage > 100 {
				t.Errorf("CPU usage should be between 0 and 100, got %f", cpuUsage)
			}
		}
	})

	// Test 2: Get alerts should work
	t.Run("GetAlerts", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/alerts", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get alerts should return 200, got %d", w.Code)
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
	})

	// Test 3: Get system logs should work
	t.Run("GetSystemLogs", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/logs?limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get system logs should return 200, got %d", w.Code)
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
	})

	// Test 4: Stats should be consistent over time
	t.Run("StatsConsistency", func(t *testing.T) {
		var prevStats map[string]interface{}
		
		for i := 0; i < 3; i++ {
			req, _ := http.NewRequest("GET", "/api/stats", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Get stats request %d failed with status %d", i, w.Code)
			}

			var stats map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &stats)
			if err != nil {
				t.Fatalf("Failed to unmarshal stats response: %v", err)
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
}

// TestRegressionMoodleManagement tests Moodle management regression
func TestRegressionMoodleManagement(t *testing.T) {
	router := setupRegressionTestRouter(t)
	token := getRegressionAuthToken(t, router)

	// Test 1: Get Moodle status should work
	t.Run("GetMoodleStatus", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/moodle/status", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get Moodle status should return 200, got %d", w.Code)
		}

		var status map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &status)
		if err != nil {
			t.Fatalf("Failed to unmarshal Moodle status response: %v", err)
		}

		// Check required fields
		requiredFields := []string{"running", "version", "uptime", "last_check"}
		for _, field := range requiredFields {
			if status[field] == nil {
				t.Errorf("Moodle status should contain %s", field)
			}
		}
	})

	// Test 2: Moodle operations should handle gracefully
	t.Run("MoodleOperations", func(t *testing.T) {
		operations := []string{
			"/api/moodle/start",
			"/api/moodle/stop",
			"/api/moodle/restart",
		}

		for _, operation := range operations {
			req, _ := http.NewRequest("POST", operation, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Accept both success and failure (Moodle might not be installed)
			if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
				t.Errorf("Moodle operation %s should return 200 or 500, got %d", operation, w.Code)
			}
		}
	})
}

// TestRegressionErrorHandling tests error handling regression
func TestRegressionErrorHandling(t *testing.T) {
	router := setupRegressionTestRouter(t)

	// Test 1: Invalid endpoint should return 404
	t.Run("InvalidEndpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/nonexistent", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Invalid endpoint should return 404, got %d", w.Code)
		}
	})

	// Test 2: Invalid method should return 405
	t.Run("InvalidMethod", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/health", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Invalid method should return 405, got %d", w.Code)
		}
	})

	// Test 3: Malformed JSON should return 400
	t.Run("MalformedJSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Malformed JSON should return 400, got %d", w.Code)
		}
	})

	// Test 4: Missing content type should return 400
	t.Run("MissingContentType", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString("{}"))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Missing content type should return 400, got %d", w.Code)
		}
	})
}

// TestRegressionSecurity tests security regression
func TestRegressionSecurity(t *testing.T) {
	router := setupRegressionTestRouter(t)

	// Test 1: Rate limiting should work
	t.Run("RateLimiting", func(t *testing.T) {
		// Make multiple requests quickly
		for i := 0; i < 150; i++ { // Exceed the 100 req/min limit
			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if i >= 100 && w.Code == http.StatusTooManyRequests {
				// Rate limiting is working
				return
			}
		}

		t.Error("Rate limiting should be triggered")
	})

	// Test 2: SQL injection should be prevented
	t.Run("SQLInjection", func(t *testing.T) {
		loginData := map[string]string{
			"username": "admin'; DROP TABLE users; --",
			"password": "admin123",
		}

		jsonData, _ := json.Marshal(loginData)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 401 (unauthorized) not 500 (server error)
		if w.Code == http.StatusInternalServerError {
			t.Error("SQL injection vulnerability detected")
		}
	})

	// Test 3: XSS should be prevented
	t.Run("XSS", func(t *testing.T) {
		loginData := map[string]string{
			"username": "<script>alert('XSS')</script>",
			"password": "admin123",
		}

		jsonData, _ := json.Marshal(loginData)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check if XSS payload is reflected in response
		if bytes.Contains(w.Body.Bytes(), []byte("<script>")) {
			t.Error("XSS vulnerability detected")
		}
	})
}

// TestRegressionPerformance tests performance regression
func TestRegressionPerformance(t *testing.T) {
	router := setupRegressionTestRouter(t)

	// Test 1: Health check should be fast
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

	// Test 2: Login should be fast
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

	// Test 3: Authenticated requests should be fast
	t.Run("AuthenticatedRequestPerformance", func(t *testing.T) {
		token := getRegressionAuthToken(t, router)
		
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
}

// TestRegressionDataIntegrity tests data integrity regression
func TestRegressionDataIntegrity(t *testing.T) {
	router := setupRegressionTestRouter(t)
	token := getRegressionAuthToken(t, router)

	// Test 1: User creation and retrieval should be consistent
	t.Run("UserDataIntegrity", func(t *testing.T) {
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
			t.Fatalf("Failed to unmarshal user response: %v", err)
		}

		// Verify user appears in user list
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
				t.Errorf("Email mismatch: expected 'integrity@example.com', got %v", foundUser["email"])
			}

			if foundUser["role"] != "viewer" {
				t.Errorf("Role mismatch: expected 'viewer', got %v", foundUser["role"])
			}
		}
	})

	// Test 2: System stats should be consistent
	t.Run("SystemStatsIntegrity", func(t *testing.T) {
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
	})
}

// Helper functions

func setupRegressionTestRouter(t *testing.T) *gin.Engine {
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

	db := setupRegressionTestDB(t)

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

func getRegressionAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupRegressionTestDB(t *testing.T) *sql.DB {
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
