package integration

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

// TestFullSystemIntegration tests the complete system integration
func TestFullSystemIntegration(t *testing.T) {
	// Setup complete system
	router, db := setupIntegrationTestSystem(t)
	defer db.Close()

	// Test complete workflow
	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Step 1: Health check
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Health check failed with status %d", w.Code)
		}

		// Step 2: Login
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

		var loginResponse map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
		if err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		token, ok := loginResponse["token"].(string)
		if !ok {
			t.Fatal("Token not found in login response")
		}

		// Step 3: Access protected endpoints
		protectedEndpoints := []string{
			"/api/stats",
			"/api/moodle/status",
			"/api/users",
			"/api/alerts",
			"/api/logs",
		}

		for _, endpoint := range protectedEndpoints {
			req, _ = http.NewRequest("GET", endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w = httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Protected endpoint %s failed with status %d", endpoint, w.Code)
			}
		}

		// Step 4: Create user
		userData := map[string]interface{}{
			"username": "testuser",
			"email":    "test@example.com",
			"password": "TestPass123!",
			"role":     "operator",
		}

		jsonData, _ = json.Marshal(userData)
		req, _ = http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Create user failed with status %d", w.Code)
		}

		// Step 5: Logout
		req, _ = http.NewRequest("POST", "/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Logout failed with status %d", w.Code)
		}
	})
}

// TestDatabaseIntegration tests database integration
func TestDatabaseIntegration(t *testing.T) {
	router, db := setupIntegrationTestSystem(t)
	defer db.Close()

	t.Run("DatabaseOperations", func(t *testing.T) {
		// Test user creation and retrieval
		authService := services.NewAuthService("test-secret", db)

		// Create user
		user, err := authService.CreateUser(&models.CreateUserRequest{
			Username: "dbuser",
			Email:    "db@example.com",
			Password: "TestPass123!",
			Role:     models.RoleViewer,
		})
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Retrieve user
		retrievedUser, err := authService.GetUser(user.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve user: %v", err)
		}

		if retrievedUser.Username != user.Username {
			t.Errorf("Username mismatch: expected %s, got %s", user.Username, retrievedUser.Username)
		}

		// Update user
		updatedUser, err := authService.UpdateUser(user.ID, &models.UpdateUserRequest{
			Email: stringPtr("updated@example.com"),
		})
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		if updatedUser.Email != "updated@example.com" {
			t.Errorf("Email not updated: expected updated@example.com, got %s", updatedUser.Email)
		}

		// Get all users
		users, err := authService.GetUsers()
		if err != nil {
			t.Fatalf("Failed to get users: %v", err)
		}

		if len(users) < 2 { // Should have at least admin and created user
			t.Errorf("Expected at least 2 users, got %d", len(users))
		}

		// Get user stats
		stats, err := authService.GetUserStats()
		if err != nil {
			t.Fatalf("Failed to get user stats: %v", err)
		}

		if stats.TotalUsers < 2 {
			t.Errorf("Expected at least 2 total users, got %d", stats.TotalUsers)
		}
	})
}

// TestMonitoringIntegration tests monitoring system integration
func TestMonitoringIntegration(t *testing.T) {
	router, db := setupIntegrationTestSystem(t)
	defer db.Close()

	t.Run("MonitoringSystem", func(t *testing.T) {
		monitorService := services.NewMonitorService(config.MonitoringConfig{
			UpdateInterval: 1,
			LogRetention:   7,
			AlertThresholds: config.AlertThresholdsConfig{
				CPU:    80.0,
				Memory: 85.0,
				Disk:   90.0,
			},
		})

		monitorService.SetDatabase(db)

		// Start monitoring
		monitorService.Start()
		defer monitorService.Stop()

		// Wait for updates
		time.Sleep(3 * time.Second)

		// Get stats
		stats := monitorService.GetStats()
		if stats == nil {
			t.Fatal("Stats should not be nil")
		}

		// Verify stats are reasonable
		if stats.CPUUsage < 0 || stats.CPUUsage > 100 {
			t.Errorf("CPU usage should be between 0 and 100, got %f", stats.CPUUsage)
		}

		if stats.MemoryUsage < 0 || stats.MemoryUsage > 100 {
			t.Errorf("Memory usage should be between 0 and 100, got %f", stats.MemoryUsage)
		}

		if stats.DiskUsage < 0 || stats.DiskUsage > 100 {
			t.Errorf("Disk usage should be between 0 and 100, got %f", stats.DiskUsage)
		}

		// Get alerts
		alerts, err := monitorService.GetAlerts()
		if err != nil {
			t.Fatalf("Failed to get alerts: %v", err)
		}

		// Alerts can be empty, just verify the call works
		_ = alerts

		// Get logs
		logs, err := monitorService.GetSystemLogs(10)
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		// Logs can be empty, just verify the call works
		_ = logs
	})
}

// TestSecurityIntegration tests security system integration
func TestSecurityIntegration(t *testing.T) {
	router, db := setupIntegrationTestSystem(t)
	defer db.Close()

	t.Run("SecuritySystem", func(t *testing.T) {
		securityService := services.NewSecurityService(config.SecurityConfig{
			JWTSecret:     "test-secret",
			SessionTimeout: 3600,
			RateLimit:     100,
			AllowedIPs:   []string{"127.0.0.1", "192.168.1.0/24"},
		})

		// Test IP whitelist
		if !securityService.isIPAllowed("127.0.0.1") {
			t.Error("127.0.0.1 should be allowed")
		}

		if !securityService.isIPAllowed("192.168.1.100") {
			t.Error("192.168.1.100 should be allowed")
		}

		if securityService.isIPAllowed("10.0.0.1") {
			t.Error("10.0.0.1 should not be allowed")
		}

		// Test rate limiting
		rateLimiter := services.NewRateLimiter(10) // 10 requests per minute

		// Test rate limiting
		for i := 0; i < 15; i++ {
			allowed := rateLimiter.Allow("127.0.0.1")
			if i < 10 && !allowed {
				t.Errorf("Request %d should be allowed", i)
			}
			if i >= 10 && allowed {
				t.Errorf("Request %d should be rate limited", i)
			}
		}

		// Test security stats
		stats := securityService.GetSecurityStats()
		if stats == nil {
			t.Fatal("Security stats should not be nil")
		}

		if stats["rate_limit"] != 100 {
			t.Errorf("Expected rate limit 100, got %v", stats["rate_limit"])
		}
	})
}

// TestMoodleIntegration tests Moodle management integration
func TestMoodleIntegration(t *testing.T) {
	router, db := setupIntegrationTestSystem(t)
	defer db.Close()

	t.Run("MoodleManagement", func(t *testing.T) {
		moodleService := services.NewMoodleService(config.MoodleConfig{
			Path:       "/tmp/moodle",
			ConfigPath: "/tmp/moodle/config.php",
			DataPath:   "/tmp/moodledata",
		})

		// Get Moodle status
		status := moodleService.GetStatus()
		if status == nil {
			t.Fatal("Moodle status should not be nil")
		}

		// Status fields should be present
		if status.LastCheck.IsZero() {
			t.Error("LastCheck should not be zero")
		}

		// Test Moodle info
		info, err := moodleService.GetMoodleInfo()
		if err != nil {
			t.Fatalf("Failed to get Moodle info: %v", err)
		}

		if info == nil {
			t.Fatal("Moodle info should not be nil")
		}

		// Check required fields
		requiredFields := []string{"path", "config_path", "data_path", "path_exists", "config_exists", "data_exists", "running"}
		for _, field := range requiredFields {
			if info[field] == nil {
				t.Errorf("Moodle info should contain %s", field)
			}
		}
	})
}

// TestAuthenticationIntegration tests authentication system integration
func TestAuthenticationIntegration(t *testing.T) {
	router, db := setupIntegrationTestSystem(t)
	defer db.Close()

	t.Run("AuthenticationSystem", func(t *testing.T) {
		authService := services.NewAuthService("test-secret", db)

		// Test user creation
		user, err := authService.CreateUser(&models.CreateUserRequest{
			Username: "authuser",
			Email:    "auth@example.com",
			Password: "TestPass123!",
			Role:     models.RoleAdmin,
		})
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Test login
		response, err := authService.Login("authuser", "TestPass123!", "127.0.0.1", "test-agent")
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		if response.Token == "" {
			t.Error("Login response should contain token")
		}

		if response.User.Username != "authuser" {
			t.Errorf("Expected username 'authuser', got %s", response.User.Username)
		}

		// Test token validation
		claims, err := authService.ValidateToken(response.Token)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if claims.UserID != user.ID {
			t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
		}

		// Test logout
		err = authService.Logout(user.ID, "127.0.0.1", "test-agent")
		if err != nil {
			t.Fatalf("Failed to logout: %v", err)
		}
	})
}

// TestErrorHandlingIntegration tests error handling integration
func TestErrorHandlingIntegration(t *testing.T) {
	router, db := setupIntegrationTestSystem(t)
	defer db.Close()

	t.Run("ErrorHandling", func(t *testing.T) {
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
	})
}

// TestPerformanceIntegration tests performance integration
func TestPerformanceIntegration(t *testing.T) {
	router, db := setupIntegrationTestSystem(t)
	defer db.Close()

	t.Run("Performance", func(t *testing.T) {
		// Test response times
		endpoints := []string{
			"/health",
			"/login",
		}

		for _, endpoint := range endpoints {
			start := time.Now()
			
			var req *http.Request
			if endpoint == "/login" {
				loginData := map[string]string{
					"username": "admin",
					"password": "admin123",
				}
				jsonData, _ := json.Marshal(loginData)
				req, _ = http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = http.NewRequest("GET", endpoint, nil)
			}
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			duration := time.Since(start)
			
			if w.Code != http.StatusOK {
				t.Errorf("Endpoint %s failed with status %d", endpoint, w.Code)
			}
			
			if duration > 500*time.Millisecond {
				t.Errorf("Endpoint %s response time too slow: %v", endpoint, duration)
			}
		}
	})
}

// Helper functions

func setupIntegrationTestSystem(t *testing.T) (*gin.Engine, *sql.DB) {
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

	db := setupIntegrationTestDB(t)

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
		protected.GET("/users", apiHandler.GetUsers)
		protected.POST("/users", apiHandler.CreateUser)
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

func setupIntegrationTestDB(t *testing.T) *sql.DB {
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

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
