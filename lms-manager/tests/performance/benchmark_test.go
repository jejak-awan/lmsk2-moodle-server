package performance

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

// BenchmarkLogin tests the performance of the login endpoint
func BenchmarkLogin(b *testing.B) {
	router := setupBenchmarkRouter(b)
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	jsonData, _ := json.Marshal(loginData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkGetStats tests the performance of the stats endpoint
func BenchmarkGetStats(b *testing.B) {
	router := setupBenchmarkRouter(b)
	token := getBenchmarkAuthToken(b, router)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkGetMoodleStatus tests the performance of the Moodle status endpoint
func BenchmarkGetMoodleStatus(b *testing.B) {
	router := setupBenchmarkRouter(b)
	token := getBenchmarkAuthToken(b, router)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/moodle/status", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkGetUsers tests the performance of the users endpoint
func BenchmarkGetUsers(b *testing.B) {
	router := setupBenchmarkRouter(b)
	token := getBenchmarkAuthToken(b, router)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkGetAlerts tests the performance of the alerts endpoint
func BenchmarkGetAlerts(b *testing.B) {
	router := setupBenchmarkRouter(b)
	token := getBenchmarkAuthToken(b, router)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/alerts", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkGetSystemLogs tests the performance of the system logs endpoint
func BenchmarkGetSystemLogs(b *testing.B) {
	router := setupBenchmarkRouter(b)
	token := getBenchmarkAuthToken(b, router)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/logs?limit=100", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkHealthCheck tests the performance of the health check endpoint
func BenchmarkHealthCheck(b *testing.B) {
	router := setupBenchmarkRouter(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkConcurrentRequests tests concurrent request handling
func BenchmarkConcurrentRequests(b *testing.B) {
	router := setupBenchmarkRouter(b)
	token := getBenchmarkAuthToken(b, router)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/api/stats", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkMemoryUsage tests memory usage under load
func BenchmarkMemoryUsage(b *testing.B) {
	router := setupBenchmarkRouter(b)
	token := getBenchmarkAuthToken(b, router)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkDatabaseOperations tests database operation performance
func BenchmarkDatabaseOperations(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	authService := services.NewAuthService("test-secret", db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test user creation
		_, err := authService.CreateUser(&models.CreateUserRequest{
			Username: "testuser" + string(rune(i)),
			Email:    "test" + string(rune(i)) + "@example.com",
			Password: "TestPass123!",
			Role:     models.RoleViewer,
		})
		if err != nil {
			b.Fatalf("Failed to create user: %v", err)
		}
	}
}

// BenchmarkPasswordHashing tests password hashing performance
func BenchmarkPasswordHashing(b *testing.B) {
	password := "TestPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := utils.HashPassword(password)
		if err != nil {
			b.Fatalf("Failed to hash password: %v", err)
		}
	}
}

// BenchmarkTokenValidation tests JWT token validation performance
func BenchmarkTokenValidation(b *testing.B) {
	authService := services.NewAuthService("test-secret", nil)
	
	// Create a test token
	token, _, err := authService.generateToken(&models.User{
		ID:       "test-id",
		Username: "testuser",
		Role:     "admin",
	})
	if err != nil {
		b.Fatalf("Failed to generate token: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := authService.ValidateToken(token)
		if err != nil {
			b.Fatalf("Failed to validate token: %v", err)
		}
	}
}

// BenchmarkSystemMonitoring tests system monitoring performance
func BenchmarkSystemMonitoring(b *testing.B) {
	monitorService := services.NewMonitorService(config.MonitoringConfig{
		UpdateInterval: 30,
		LogRetention:   7,
		AlertThresholds: config.AlertThresholdsConfig{
			CPU:    80.0,
			Memory: 85.0,
			Disk:   90.0,
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats := monitorService.GetStats()
		if stats == nil {
			b.Fatal("Failed to get stats")
		}
	}
}

// BenchmarkMoodleManagement tests Moodle management performance
func BenchmarkMoodleManagement(b *testing.B) {
	moodleService := services.NewMoodleService(config.MoodleConfig{
		Path:       "/tmp/moodle",
		ConfigPath: "/tmp/moodle/config.php",
		DataPath:   "/tmp/moodledata",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		status := moodleService.GetStatus()
		if status == nil {
			b.Fatal("Failed to get Moodle status")
		}
	}
}

// BenchmarkSecurityOperations tests security operation performance
func BenchmarkSecurityOperations(b *testing.B) {
	securityService := services.NewSecurityService(config.SecurityConfig{
		JWTSecret:     "test-secret",
		SessionTimeout: 3600,
		RateLimit:     100,
		AllowedIPs:   []string{"127.0.0.1"},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats := securityService.GetSecurityStats()
		if stats == nil {
			b.Fatal("Failed to get security stats")
		}
	}
}

// Helper functions for benchmarks

func setupBenchmarkRouter(b *testing.B) *gin.Engine {
	gin.SetMode(gin.ReleaseMode) // Use release mode for benchmarks
	
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

	db := setupBenchmarkDB(b)

	authService := services.NewAuthService(cfg.Security.JWTSecret, db)
	monitorService := services.NewMonitorService(cfg.Monitoring)
	moodleService := services.NewMoodleService(cfg.Moodle)
	securityService := services.NewSecurityService(cfg.Security)

	authHandler := handlers.NewAuthHandler(authService)
	apiHandler := handlers.NewAPIHandler(monitorService, moodleService, securityService)

	router := gin.New()
	router.Use(gin.Recovery())

	router.POST("/login", authHandler.Login)

	protected := router.Group("/api")
	protected.Use(authHandler.AuthMiddleware())
	{
		protected.GET("/stats", apiHandler.GetStats)
		protected.GET("/moodle/status", apiHandler.GetMoodleStatus)
		protected.GET("/users", apiHandler.GetUsers)
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

func getBenchmarkAuthToken(b *testing.B, router *gin.Engine) string {
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
		b.Fatalf("Login failed with status %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		b.Fatalf("Failed to unmarshal login response: %v", err)
	}

	token, ok := response["token"].(string)
	if !ok {
		b.Fatal("Token not found in login response")
	}

	return token
}

func setupBenchmarkDB(b *testing.B) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open test database: %v", err)
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
			b.Fatalf("Failed to create table: %v", err)
		}
	}

	passwordHash, _ := utils.HashPassword("admin123")
	_, err = db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, active)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "admin-id", "admin", "admin@k2net.id", passwordHash, "admin", true)

	if err != nil {
		b.Fatalf("Failed to create admin user: %v", err)
	}

	return db
}
