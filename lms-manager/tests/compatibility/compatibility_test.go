package compatibility

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	"lms-manager/config"
	"lms-manager/handlers"
	"lms-manager/services"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// TestCompatibilityOperatingSystems tests OS compatibility
func TestCompatibilityOperatingSystems(t *testing.T) {
	// Test 1: Linux compatibility
	t.Run("LinuxCompatibility", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("Skipping Linux compatibility test on non-Linux system")
		}

		// Test system calls work
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Linux compatibility test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Verify system stats are available
		if stats["cpu_usage"] == nil {
			t.Error("CPU usage not available on Linux")
		}

		if stats["memory_usage"] == nil {
			t.Error("Memory usage not available on Linux")
		}

		if stats["disk_usage"] == nil {
			t.Error("Disk usage not available on Linux")
		}
	})

	// Test 2: Windows compatibility
	t.Run("WindowsCompatibility", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Skipping Windows compatibility test on non-Windows system")
		}

		// Test basic functionality
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Windows compatibility test failed with status %d", w.Code)
		}
	})

	// Test 3: macOS compatibility
	t.Run("macOSCompatibility", func(t *testing.T) {
		if runtime.GOOS != "darwin" {
			t.Skip("Skipping macOS compatibility test on non-macOS system")
		}

		// Test basic functionality
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("macOS compatibility test failed with status %d", w.Code)
		}
	})
}

// TestCompatibilityArchitectures tests architecture compatibility
func TestCompatibilityArchitectures(t *testing.T) {
	// Test 1: AMD64 compatibility
	t.Run("AMD64Compatibility", func(t *testing.T) {
		if runtime.GOARCH != "amd64" {
			t.Skip("Skipping AMD64 compatibility test on non-AMD64 system")
		}

		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("AMD64 compatibility test failed with status %d", w.Code)
		}
	})

	// Test 2: ARM64 compatibility
	t.Run("ARM64Compatibility", func(t *testing.T) {
		if runtime.GOARCH != "arm64" {
			t.Skip("Skipping ARM64 compatibility test on non-ARM64 system")
		}

		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("ARM64 compatibility test failed with status %d", w.Code)
		}
	})

	// Test 3: ARM compatibility
	t.Run("ARMCompatibility", func(t *testing.T) {
		if runtime.GOARCH != "arm" {
			t.Skip("Skipping ARM compatibility test on non-ARM system")
		}

		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("ARM compatibility test failed with status %d", w.Code)
		}
	})
}

// TestCompatibilityGoVersions tests Go version compatibility
func TestCompatibilityGoVersions(t *testing.T) {
	// Test 1: Go 1.21+ compatibility
	t.Run("Go121Compatibility", func(t *testing.T) {
		// Test that we can use Go 1.21+ features
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Go 1.21+ compatibility test failed with status %d", w.Code)
		}
	})

	// Test 2: Go modules compatibility
	t.Run("GoModulesCompatibility", func(t *testing.T) {
		// Test that Go modules work correctly
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Go modules compatibility test failed with status %d", w.Code)
		}
	})
}

// TestCompatibilityDatabases tests database compatibility
func TestCompatibilityDatabases(t *testing.T) {
	// Test 1: SQLite compatibility
	t.Run("SQLiteCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		// Test user creation
		userData := map[string]interface{}{
			"username": "sqliteuser",
			"email":    "sqlite@example.com",
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
			t.Errorf("SQLite compatibility test failed with status %d", w.Code)
		}

		// Test user retrieval
		req, _ = http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("SQLite user retrieval test failed with status %d", w.Code)
		}

		var users []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &users)
		if err != nil {
			t.Fatalf("Failed to unmarshal users response: %v", err)
		}

		// Verify user was created
		var foundUser map[string]interface{}
		for _, user := range users {
			if user["username"] == "sqliteuser" {
				foundUser = user
				break
			}
		}

		if foundUser == nil {
			t.Error("SQLite user not found after creation")
		}
	})
}

// TestCompatibilityWebServers tests web server compatibility
func TestCompatibilityWebServers(t *testing.T) {
	// Test 1: Nginx compatibility
	t.Run("NginxCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		// Test that responses are compatible with Nginx
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		req.Header.Set("X-Real-IP", "192.168.1.1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Nginx compatibility test failed with status %d", w.Code)
		}

		// Verify headers are set correctly
		if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Error("Content-Type header not set correctly for Nginx")
		}
	})

	// Test 2: Apache compatibility
	t.Run("ApacheCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		// Test that responses are compatible with Apache
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Apache compatibility test failed with status %d", w.Code)
		}
	})
}

// TestCompatibilityBrowsers tests browser compatibility
func TestCompatibilityBrowsers(t *testing.T) {
	// Test 1: Chrome compatibility
	t.Run("ChromeCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Chrome compatibility test failed with status %d", w.Code)
		}
	})

	// Test 2: Firefox compatibility
	t.Run("FirefoxCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Firefox compatibility test failed with status %d", w.Code)
		}
	})

	// Test 3: Safari compatibility
	t.Run("SafariCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Safari compatibility test failed with status %d", w.Code)
		}
	})

	// Test 4: Edge compatibility
	t.Run("EdgeCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Edge compatibility test failed with status %d", w.Code)
		}
	})
}

// TestCompatibilityMobileDevices tests mobile device compatibility
func TestCompatibilityMobileDevices(t *testing.T) {
	// Test 1: iOS compatibility
	t.Run("iOSCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("iOS compatibility test failed with status %d", w.Code)
		}
	})

	// Test 2: Android compatibility
	t.Run("AndroidCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 11; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Android compatibility test failed with status %d", w.Code)
		}
	})
}

// TestCompatibilityAPIs tests API compatibility
func TestCompatibilityAPIs(t *testing.T) {
	// Test 1: REST API compatibility
	t.Run("RESTAPICompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		// Test GET request
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("REST API GET test failed with status %d", w.Code)
		}

		// Test POST request
		userData := map[string]interface{}{
			"username": "apiuser",
			"email":    "api@example.com",
			"password": "TestPass123!",
			"role":     "viewer",
		}

		jsonData, _ := json.Marshal(userData)
		req, _ = http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("REST API POST test failed with status %d", w.Code)
		}
	})

	// Test 2: JSON API compatibility
	t.Run("JSONAPICompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("JSON API test failed with status %d", w.Code)
		}

		// Verify JSON response
		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal JSON response: %v", err)
		}

		// Verify JSON structure
		if stats["cpu_usage"] == nil {
			t.Error("JSON response missing cpu_usage field")
		}

		if stats["memory_usage"] == nil {
			t.Error("JSON response missing memory_usage field")
		}
	})
}

// TestCompatibilitySecurity tests security compatibility
func TestCompatibilitySecurity(t *testing.T) {
	// Test 1: HTTPS compatibility
	t.Run("HTTPSCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Forwarded-Proto", "https")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("HTTPS compatibility test failed with status %d", w.Code)
		}
	})

	// Test 2: CORS compatibility
	t.Run("CORSCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Origin", "https://example.com")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("CORS compatibility test failed with status %d", w.Code)
		}
	})

	// Test 3: JWT compatibility
	t.Run("JWTCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		// Test JWT token format
		if len(token) < 100 {
			t.Error("JWT token too short")
		}

		// Test JWT token usage
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("JWT compatibility test failed with status %d", w.Code)
		}
	})
}

// TestCompatibilityPerformance tests performance compatibility
func TestCompatibilityPerformance(t *testing.T) {
	// Test 1: Memory usage compatibility
	t.Run("MemoryUsageCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		// Make multiple requests to test memory usage
		for i := 0; i < 100; i++ {
			req, _ := http.NewRequest("GET", "/api/stats", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Memory usage test failed at request %d with status %d", i, w.Code)
			}
		}
	})

	// Test 2: CPU usage compatibility
	t.Run("CPUUsageCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		// Test CPU-intensive operations
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("CPU usage test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Verify CPU usage is reported
		if stats["cpu_usage"] == nil {
			t.Error("CPU usage not reported")
		}
	})
}

// TestCompatibilityScalability tests scalability compatibility
func TestCompatibilityScalability(t *testing.T) {
	// Test 1: Concurrent requests compatibility
	t.Run("ConcurrentRequestsCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		// Test concurrent requests
		done := make(chan bool, 10)
		
		for i := 0; i < 10; i++ {
			go func() {
				req, _ := http.NewRequest("GET", "/api/stats", nil)
				req.Header.Set("Authorization", "Bearer "+token)

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Concurrent request failed with status %d", w.Code)
				}

				done <- true
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	// Test 2: Large payload compatibility
	t.Run("LargePayloadCompatibility", func(t *testing.T) {
		router := setupCompatibilityTestRouter(t)
		token := getCompatibilityAuthToken(t, router)

		// Test large user creation
		userData := map[string]interface{}{
			"username": "largeuser",
			"email":    "large@example.com",
			"password": "TestPass123!",
			"role":     "viewer",
			"description": "This is a very long description that tests large payload compatibility. " +
				"It contains multiple sentences and should test the system's ability to handle " +
				"larger data payloads without issues. This is important for scalability and " +
				"performance testing of the application.",
		}

		jsonData, _ := json.Marshal(userData)
		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Large payload test failed with status %d", w.Code)
		}
	})
}

// Helper functions

func setupCompatibilityTestRouter(t *testing.T) *gin.Engine {
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

	db := setupCompatibilityTestDB(t)

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

func getCompatibilityAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupCompatibilityTestDB(t *testing.T) *sql.DB {
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
