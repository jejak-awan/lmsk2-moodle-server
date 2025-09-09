package snapshot

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"lms-manager/config"
	"lms-manager/handlers"
	"lms-manager/services"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// TestSnapshotTestingAuthentication tests authentication snapshots
func TestSnapshotTestingAuthentication(t *testing.T) {
	// Test 1: Login response snapshot
	t.Run("LoginResponseSnapshot", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)

		// Test login response
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
			t.Errorf("Login should return 200, got %d", w.Code)
		}

		// Create snapshot of login response
		snapshot := createSnapshot(t, "login_response", w.Body.Bytes())
		
		// Verify snapshot structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		// Verify snapshot contains expected fields
		expectedFields := []string{"token", "user", "expires_at"}
		for _, field := range expectedFields {
			if response[field] == nil {
				t.Errorf("Login snapshot missing %s field", field)
			}
		}

		// Verify user object structure
		user := response["user"].(map[string]interface{})
		userFields := []string{"id", "username", "email", "role"}
		for _, field := range userFields {
			if user[field] == nil {
				t.Errorf("Login snapshot user missing %s field", field)
			}
		}

		// Save snapshot to file
		saveSnapshot(t, "login_response", snapshot)
	})

	// Test 2: Logout response snapshot
	t.Run("LogoutResponseSnapshot", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)
		token := getSnapshotAuthToken(t, router)

		// Test logout response
		req, _ := http.NewRequest("POST", "/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Logout should return 200, got %d", w.Code)
		}

		// Create snapshot of logout response
		snapshot := createSnapshot(t, "logout_response", w.Body.Bytes())
		
		// Verify snapshot structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal logout response: %v", err)
		}

		// Verify snapshot contains expected fields
		expectedFields := []string{"message", "logged_out"}
		for _, field := range expectedFields {
			if response[field] == nil {
				t.Errorf("Logout snapshot missing %s field", field)
			}
		}

		// Save snapshot to file
		saveSnapshot(t, "logout_response", snapshot)
	})

	// Test 3: Authentication error snapshot
	t.Run("AuthenticationErrorSnapshot", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)

		// Test authentication error
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Authentication error should return 401, got %d", w.Code)
		}

		// Create snapshot of authentication error
		snapshot := createSnapshot(t, "auth_error", w.Body.Bytes())
		
		// Verify snapshot structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal auth error response: %v", err)
		}

		// Verify snapshot contains expected fields
		expectedFields := []string{"error", "code", "message", "timestamp"}
		for _, field := range expectedFields {
			if response[field] == nil {
				t.Errorf("Auth error snapshot missing %s field", field)
			}
		}

		// Save snapshot to file
		saveSnapshot(t, "auth_error", snapshot)
	})
}

// TestSnapshotTestingUserManagement tests user management snapshots
func TestSnapshotTestingUserManagement(t *testing.T) {
	// Test 1: User creation snapshot
	t.Run("UserCreationSnapshot", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)
		token := getSnapshotAuthToken(t, router)

		// Test user creation
		userData := map[string]interface{}{
			"username": "snapshotuser",
			"email":    "snapshot@example.com",
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
			t.Errorf("User creation should return 201, got %d", w.Code)
		}

		// Create snapshot of user creation response
		snapshot := createSnapshot(t, "user_creation", w.Body.Bytes())
		
		// Verify snapshot structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal user creation response: %v", err)
		}

		// Verify snapshot contains expected fields
		expectedFields := []string{"id", "username", "email", "role", "active", "created_at", "updated_at"}
		for _, field := range expectedFields {
			if response[field] == nil {
				t.Errorf("User creation snapshot missing %s field", field)
			}
		}

		// Verify password is not exposed
		if response["password"] != nil {
			t.Error("User creation snapshot should not expose password")
		}

		if response["password_hash"] != nil {
			t.Error("User creation snapshot should not expose password_hash")
		}

		// Save snapshot to file
		saveSnapshot(t, "user_creation", snapshot)
	})

	// Test 2: User list snapshot
	t.Run("UserListSnapshot", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)
		token := getSnapshotAuthToken(t, router)

		// Test user list
		req, _ := http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("User list should return 200, got %d", w.Code)
		}

		// Create snapshot of user list response
		snapshot := createSnapshot(t, "user_list", w.Body.Bytes())
		
		// Verify snapshot structure
		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal user list response: %v", err)
		}

		// Verify snapshot is array
		if response == nil {
			t.Error("User list snapshot should be an array")
		}

		// Verify each user has required fields
		if len(response) > 0 {
			user := response[0]
			expectedFields := []string{"id", "username", "email", "role", "active", "created_at", "updated_at"}
			for _, field := range expectedFields {
				if user[field] == nil {
					t.Errorf("User list snapshot user missing %s field", field)
				}
			}

			// Verify password is not exposed
			if user["password"] != nil {
				t.Error("User list snapshot should not expose password")
			}

			if user["password_hash"] != nil {
				t.Error("User list snapshot should not expose password_hash")
			}
		}

		// Save snapshot to file
		saveSnapshot(t, "user_list", snapshot)
	})

	// Test 3: User stats snapshot
	t.Run("UserStatsSnapshot", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)
		token := getSnapshotAuthToken(t, router)

		// Test user stats
		req, _ := http.NewRequest("GET", "/api/users/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("User stats should return 200, got %d", w.Code)
		}

		// Create snapshot of user stats response
		snapshot := createSnapshot(t, "user_stats", w.Body.Bytes())
		
		// Verify snapshot structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal user stats response: %v", err)
		}

		// Verify snapshot contains expected fields
		expectedFields := []string{"total_users", "active_users", "inactive_users", "users_by_role", "new_users_today", "new_users_this_week", "new_users_this_month"}
		for _, field := range expectedFields {
			if response[field] == nil {
				t.Errorf("User stats snapshot missing %s field", field)
			}
		}

		// Verify users_by_role is object
		usersByRole := response["users_by_role"].(map[string]interface{})
		if usersByRole == nil {
			t.Error("User stats snapshot users_by_role should be an object")
		}

		// Save snapshot to file
		saveSnapshot(t, "user_stats", snapshot)
	})
}

// TestSnapshotTestingSystemMonitoring tests system monitoring snapshots
func TestSnapshotTestingSystemMonitoring(t *testing.T) {
	// Test 1: System stats snapshot
	t.Run("SystemStatsSnapshot", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)
		token := getSnapshotAuthToken(t, router)

		// Test system stats
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("System stats should return 200, got %d", w.Code)
		}

		// Create snapshot of system stats response
		snapshot := createSnapshot(t, "system_stats", w.Body.Bytes())
		
		// Verify snapshot structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal system stats response: %v", err)
		}

		// Verify snapshot contains expected fields
		expectedFields := []string{"cpu_usage", "memory_usage", "disk_usage", "uptime", "load_average", "network_io", "disk_io", "timestamp"}
		for _, field := range expectedFields {
			if response[field] == nil {
				t.Errorf("System stats snapshot missing %s field", field)
			}
		}

		// Verify load_average is array
		loadAverage := response["load_average"].([]interface{})
		if loadAverage == nil {
			t.Error("System stats snapshot load_average should be an array")
		}

		// Verify network_io is object
		networkIO := response["network_io"].(map[string]interface{})
		if networkIO == nil {
			t.Error("System stats snapshot network_io should be an object")
		}

		// Verify disk_io is object
		diskIO := response["disk_io"].(map[string]interface{})
		if diskIO == nil {
			t.Error("System stats snapshot disk_io should be an object")
		}

		// Save snapshot to file
		saveSnapshot(t, "system_stats", snapshot)
	})

	// Test 2: Alerts snapshot
	t.Run("AlertsSnapshot", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)
		token := getSnapshotAuthToken(t, router)

		// Test alerts
		req, _ := http.NewRequest("GET", "/api/alerts", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Alerts should return 200, got %d", w.Code)
		}

		// Create snapshot of alerts response
		snapshot := createSnapshot(t, "alerts", w.Body.Bytes())
		
		// Verify snapshot structure
		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal alerts response: %v", err)
		}

		// Verify snapshot is array
		if response == nil {
			t.Error("Alerts snapshot should be an array")
		}

		// Verify each alert has required fields (if any alerts exist)
		if len(response) > 0 {
			alert := response[0]
			expectedFields := []string{"id", "type", "message", "severity", "resolved", "created_at", "resolved_at"}
			for _, field := range expectedFields {
				if alert[field] == nil {
					t.Errorf("Alerts snapshot alert missing %s field", field)
				}
			}
		}

		// Save snapshot to file
		saveSnapshot(t, "alerts", snapshot)
	})

	// Test 3: System logs snapshot
	t.Run("SystemLogsSnapshot", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)
		token := getSnapshotAuthToken(t, router)

		// Test system logs
		req, _ := http.NewRequest("GET", "/api/logs?limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("System logs should return 200, got %d", w.Code)
		}

		// Create snapshot of system logs response
		snapshot := createSnapshot(t, "system_logs", w.Body.Bytes())
		
		// Verify snapshot structure
		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal system logs response: %v", err)
		}

		// Verify snapshot is array
		if response == nil {
			t.Error("System logs snapshot should be an array")
		}

		// Verify each log has required fields (if any logs exist)
		if len(response) > 0 {
			log := response[0]
			expectedFields := []string{"id", "level", "message", "source", "created_at"}
			for _, field := range expectedFields {
				if log[field] == nil {
					t.Errorf("System logs snapshot log missing %s field", field)
				}
			}
		}

		// Save snapshot to file
		saveSnapshot(t, "system_logs", snapshot)
	})
}

// TestSnapshotTestingMoodleManagement tests Moodle management snapshots
func TestSnapshotTestingMoodleManagement(t *testing.T) {
	// Test 1: Moodle status snapshot
	t.Run("MoodleStatusSnapshot", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)
		token := getSnapshotAuthToken(t, router)

		// Test Moodle status
		req, _ := http.NewRequest("GET", "/api/moodle/status", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Moodle status should return 200, got %d", w.Code)
		}

		// Create snapshot of Moodle status response
		snapshot := createSnapshot(t, "moodle_status", w.Body.Bytes())
		
		// Verify snapshot structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal Moodle status response: %v", err)
		}

		// Verify snapshot contains expected fields
		expectedFields := []string{"running", "version", "uptime", "last_check", "database_status", "cache_status", "plugins_status", "maintenance_mode"}
		for _, field := range expectedFields {
			if response[field] == nil {
				t.Errorf("Moodle status snapshot missing %s field", field)
			}
		}

		// Save snapshot to file
		saveSnapshot(t, "moodle_status", snapshot)
	})

	// Test 2: Moodle operations snapshot
	t.Run("MoodleOperationsSnapshot", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)
		token := getSnapshotAuthToken(t, router)

		// Test Moodle start
		req, _ := http.NewRequest("POST", "/api/moodle/start", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Accept both success and failure (Moodle might not be installed)
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("Moodle start should return 200 or 500, got %d", w.Code)
		}

		// Create snapshot of Moodle start response
		snapshot := createSnapshot(t, "moodle_start", w.Body.Bytes())
		
		// Verify snapshot structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal Moodle start response: %v", err)
		}

		// Verify snapshot contains expected fields
		expectedFields := []string{"status", "message", "timestamp"}
		for _, field := range expectedFields {
			if response[field] == nil {
				t.Errorf("Moodle start snapshot missing %s field", field)
			}
		}

		// Save snapshot to file
		saveSnapshot(t, "moodle_start", snapshot)
	})
}

// TestSnapshotTestingErrorHandling tests error handling snapshots
func TestSnapshotTestingErrorHandling(t *testing.T) {
	// Test 1: HTTP error snapshots
	t.Run("HTTPErrorSnapshots", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)

		// Test 404 error
		req, _ := http.NewRequest("GET", "/api/nonexistent", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("404 error should return 404, got %d", w.Code)
		}

		// Create snapshot of 404 error
		snapshot := createSnapshot(t, "error_404", w.Body.Bytes())
		
		// Verify snapshot structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal 404 error response: %v", err)
		}

		// Verify snapshot contains expected fields
		expectedFields := []string{"error", "code", "message", "timestamp"}
		for _, field := range expectedFields {
			if response[field] == nil {
				t.Errorf("404 error snapshot missing %s field", field)
			}
		}

		// Save snapshot to file
		saveSnapshot(t, "error_404", snapshot)

		// Test 405 error
		req, _ = http.NewRequest("DELETE", "/health", nil)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("405 error should return 405, got %d", w.Code)
		}

		// Create snapshot of 405 error
		snapshot = createSnapshot(t, "error_405", w.Body.Bytes())
		
		// Verify snapshot structure
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal 405 error response: %v", err)
		}

		// Verify snapshot contains expected fields
		for _, field := range expectedFields {
			if response[field] == nil {
				t.Errorf("405 error snapshot missing %s field", field)
			}
		}

		// Save snapshot to file
		saveSnapshot(t, "error_405", snapshot)
	})

	// Test 2: Validation error snapshots
	t.Run("ValidationErrorSnapshots", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)
		token := getSnapshotAuthToken(t, router)

		// Test validation error
		invalidUserData := map[string]interface{}{
			"username": "ab", // Too short
			"email":    "invalid-email",
			"password": "weak",
			"role":     "invalid-role",
		}

		jsonData, _ := json.Marshal(invalidUserData)
		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Validation error should return 400, got %d", w.Code)
		}

		// Create snapshot of validation error
		snapshot := createSnapshot(t, "validation_error", w.Body.Bytes())
		
		// Verify snapshot structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal validation error response: %v", err)
		}

		// Verify snapshot contains expected fields
		expectedFields := []string{"error", "code", "message", "details", "timestamp"}
		for _, field := range expectedFields {
			if response[field] == nil {
				t.Errorf("Validation error snapshot missing %s field", field)
			}
		}

		// Verify details is object
		details := response["details"].(map[string]interface{})
		if details == nil {
			t.Error("Validation error snapshot details should be an object")
		}

		// Save snapshot to file
		saveSnapshot(t, "validation_error", snapshot)
	})
}

// TestSnapshotTestingHealthCheck tests health check snapshots
func TestSnapshotTestingHealthCheck(t *testing.T) {
	// Test 1: Health check snapshot
	t.Run("HealthCheckSnapshot", func(t *testing.T) {
		router := setupSnapshotTestRouter(t)

		// Test health check
		req, _ := http.NewRequest("GET", "/health", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Health check should return 200, got %d", w.Code)
		}

		// Create snapshot of health check response
		snapshot := createSnapshot(t, "health_check", w.Body.Bytes())
		
		// Verify snapshot structure
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal health check response: %v", err)
		}

		// Verify snapshot contains expected fields
		expectedFields := []string{"status", "timestamp", "version"}
		for _, field := range expectedFields {
			if response[field] == nil {
				t.Errorf("Health check snapshot missing %s field", field)
			}
		}

		// Verify status is healthy
		if response["status"] != "healthy" {
			t.Errorf("Health check snapshot status should be 'healthy', got %v", response["status"])
		}

		// Save snapshot to file
		saveSnapshot(t, "health_check", snapshot)
	})
}

// Helper functions

func setupSnapshotTestRouter(t *testing.T) *gin.Engine {
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

	db := setupSnapshotTestDB(t)

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

func getSnapshotAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupSnapshotTestDB(t *testing.T) *sql.DB {
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

// Snapshot utility functions

func createSnapshot(t *testing.T, name string, data []byte) map[string]interface{} {
	var snapshot map[string]interface{}
	err := json.Unmarshal(data, &snapshot)
	if err != nil {
		t.Fatalf("Failed to create snapshot %s: %v", name, err)
	}
	return snapshot
}

func saveSnapshot(t *testing.T, name string, snapshot map[string]interface{}) {
	// Create snapshots directory if it doesn't exist
	snapshotsDir := "snapshots"
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		t.Fatalf("Failed to create snapshots directory: %v", err)
	}

	// Create snapshot file path
	snapshotFile := filepath.Join(snapshotsDir, name+".json")

	// Marshal snapshot to JSON
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal snapshot %s: %v", name, err)
	}

	// Write snapshot to file
	if err := os.WriteFile(snapshotFile, data, 0644); err != nil {
		t.Fatalf("Failed to save snapshot %s: %v", name, err)
	}

	t.Logf("Snapshot %s saved to %s", name, snapshotFile)
}

func loadSnapshot(t *testing.T, name string) map[string]interface{} {
	// Create snapshot file path
	snapshotFile := filepath.Join("snapshots", name+".json")

	// Read snapshot file
	data, err := os.ReadFile(snapshotFile)
	if err != nil {
		t.Fatalf("Failed to load snapshot %s: %v", name, err)
	}

	// Unmarshal snapshot
	var snapshot map[string]interface{}
	err = json.Unmarshal(data, &snapshot)
	if err != nil {
		t.Fatalf("Failed to unmarshal snapshot %s: %v", name, err)
	}

	return snapshot
}

func compareSnapshots(t *testing.T, name string, current, expected map[string]interface{}) {
	// Convert to JSON for comparison
	currentJSON, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal current snapshot %s: %v", name, err)
	}

	expectedJSON, err := json.MarshalIndent(expected, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal expected snapshot %s: %v", name, err)
	}

	// Compare snapshots
	if string(currentJSON) != string(expectedJSON) {
		t.Errorf("Snapshot %s mismatch:\nCurrent:\n%s\nExpected:\n%s", name, string(currentJSON), string(expectedJSON))
	}
}
