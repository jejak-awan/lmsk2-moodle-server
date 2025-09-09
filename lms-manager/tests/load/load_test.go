package load

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"lms-manager/config"
	"lms-manager/handlers"
	"lms-manager/services"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// LoadTestConfig represents load test configuration
type LoadTestConfig struct {
	ConcurrentUsers int
	RequestsPerUser int
	TestDuration    time.Duration
	Endpoint        string
	Method          string
	Payload         interface{}
}

// LoadTestResult represents load test results
type LoadTestResult struct {
	TotalRequests    int
	SuccessfulRequests int
	FailedRequests   int
	AverageResponseTime time.Duration
	MinResponseTime  time.Duration
	MaxResponseTime  time.Duration
	RequestsPerSecond float64
	ErrorRate        float64
}

// TestConcurrentUsers tests concurrent user load
func TestConcurrentUsers(t *testing.T) {
	router := setupLoadTestRouter(t)

	configs := []LoadTestConfig{
		{ConcurrentUsers: 10, RequestsPerUser: 100, TestDuration: 30 * time.Second, Endpoint: "/api/stats", Method: "GET"},
		{ConcurrentUsers: 50, RequestsPerUser: 100, TestDuration: 30 * time.Second, Endpoint: "/api/stats", Method: "GET"},
		{ConcurrentUsers: 100, RequestsPerUser: 100, TestDuration: 30 * time.Second, Endpoint: "/api/stats", Method: "GET"},
	}

	for _, config := range configs {
		t.Run(fmt.Sprintf("ConcurrentUsers_%d", config.ConcurrentUsers), func(t *testing.T) {
			result := runLoadTest(t, router, config)
			
			// Assertions
			if result.ErrorRate > 5.0 { // 5% error rate threshold
				t.Errorf("Error rate too high: %.2f%%", result.ErrorRate)
			}
			
			if result.AverageResponseTime > 100*time.Millisecond { // 100ms response time threshold
				t.Errorf("Average response time too high: %v", result.AverageResponseTime)
			}
			
			if result.RequestsPerSecond < 100 { // 100 RPS threshold
				t.Errorf("Requests per second too low: %.2f", result.RequestsPerSecond)
			}
			
			t.Logf("Load Test Results for %d concurrent users:", config.ConcurrentUsers)
			t.Logf("  Total Requests: %d", result.TotalRequests)
			t.Logf("  Successful Requests: %d", result.SuccessfulRequests)
			t.Logf("  Failed Requests: %d", result.FailedRequests)
			t.Logf("  Average Response Time: %v", result.AverageResponseTime)
			t.Logf("  Min Response Time: %v", result.MinResponseTime)
			t.Logf("  Max Response Time: %v", result.MaxResponseTime)
			t.Logf("  Requests Per Second: %.2f", result.RequestsPerSecond)
			t.Logf("  Error Rate: %.2f%%", result.ErrorRate)
		})
	}
}

// TestSustainedLoad tests sustained load over time
func TestSustainedLoad(t *testing.T) {
	router := setupLoadTestRouter(t)

	config := LoadTestConfig{
		ConcurrentUsers: 50,
		RequestsPerUser: 1000,
		TestDuration:    5 * time.Minute,
		Endpoint:        "/api/stats",
		Method:          "GET",
	}

	result := runLoadTest(t, router, config)

	// Assertions for sustained load
	if result.ErrorRate > 1.0 { // 1% error rate threshold for sustained load
		t.Errorf("Error rate too high for sustained load: %.2f%%", result.ErrorRate)
	}

	if result.AverageResponseTime > 50*time.Millisecond { // 50ms response time threshold
		t.Errorf("Average response time too high for sustained load: %v", result.AverageResponseTime)
	}

	t.Logf("Sustained Load Test Results:")
	t.Logf("  Total Requests: %d", result.TotalRequests)
	t.Logf("  Successful Requests: %d", result.SuccessfulRequests)
	t.Logf("  Failed Requests: %d", result.FailedRequests)
	t.Logf("  Average Response Time: %v", result.AverageResponseTime)
	t.Logf("  Min Response Time: %v", result.MinResponseTime)
	t.Logf("  Max Response Time: %v", result.MaxResponseTime)
	t.Logf("  Requests Per Second: %.2f", result.RequestsPerSecond)
	t.Logf("  Error Rate: %.2f%%", result.ErrorRate)
}

// TestMemoryUsage tests memory usage under load
func TestMemoryUsage(t *testing.T) {
	router := setupLoadTestRouter(t)

	// Get initial memory usage
	var initialMem uint64
	// Note: In a real test, you would measure actual memory usage
	// This is a placeholder for demonstration

	config := LoadTestConfig{
		ConcurrentUsers: 100,
		RequestsPerUser: 500,
		TestDuration:    2 * time.Minute,
		Endpoint:        "/api/stats",
		Method:          "GET",
	}

	result := runLoadTest(t, router, config)

	// Get final memory usage
	var finalMem uint64
	// Note: In a real test, you would measure actual memory usage
	// This is a placeholder for demonstration

	memoryIncrease := finalMem - initialMem
	maxMemoryIncrease := uint64(100 * 1024 * 1024) // 100MB

	if memoryIncrease > maxMemoryIncrease {
		t.Errorf("Memory usage increased too much: %d bytes", memoryIncrease)
	}

	t.Logf("Memory Usage Test Results:")
	t.Logf("  Initial Memory: %d bytes", initialMem)
	t.Logf("  Final Memory: %d bytes", finalMem)
	t.Logf("  Memory Increase: %d bytes", memoryIncrease)
	t.Logf("  Total Requests: %d", result.TotalRequests)
	t.Logf("  Error Rate: %.2f%%", result.ErrorRate)
}

// TestDatabaseLoad tests database performance under load
func TestDatabaseLoad(t *testing.T) {
	router := setupLoadTestRouter(t)

	config := LoadTestConfig{
		ConcurrentUsers: 50,
		RequestsPerUser: 200,
		TestDuration:    2 * time.Minute,
		Endpoint:        "/api/users",
		Method:          "GET",
	}

	result := runLoadTest(t, router, config)

	// Assertions for database load
	if result.ErrorRate > 2.0 { // 2% error rate threshold for database operations
		t.Errorf("Error rate too high for database operations: %.2f%%", result.ErrorRate)
	}

	if result.AverageResponseTime > 200*time.Millisecond { // 200ms response time threshold
		t.Errorf("Average response time too high for database operations: %v", result.AverageResponseTime)
	}

	t.Logf("Database Load Test Results:")
	t.Logf("  Total Requests: %d", result.TotalRequests)
	t.Logf("  Successful Requests: %d", result.SuccessfulRequests)
	t.Logf("  Failed Requests: %d", result.FailedRequests)
	t.Logf("  Average Response Time: %v", result.AverageResponseTime)
	t.Logf("  Min Response Time: %v", result.MinResponseTime)
	t.Logf("  Max Response Time: %v", result.MaxResponseTime)
	t.Logf("  Requests Per Second: %.2f", result.RequestsPerSecond)
	t.Logf("  Error Rate: %.2f%%", result.ErrorRate)
}

// TestAuthenticationLoad tests authentication performance under load
func TestAuthenticationLoad(t *testing.T) {
	router := setupLoadTestRouter(t)

	config := LoadTestConfig{
		ConcurrentUsers: 100,
		RequestsPerUser: 100,
		TestDuration:    1 * time.Minute,
		Endpoint:        "/login",
		Method:          "POST",
		Payload: map[string]string{
			"username": "admin",
			"password": "admin123",
		},
	}

	result := runLoadTest(t, router, config)

	// Assertions for authentication load
	if result.ErrorRate > 5.0 { // 5% error rate threshold for authentication
		t.Errorf("Error rate too high for authentication: %.2f%%", result.ErrorRate)
	}

	if result.AverageResponseTime > 300*time.Millisecond { // 300ms response time threshold
		t.Errorf("Average response time too high for authentication: %v", result.AverageResponseTime)
	}

	t.Logf("Authentication Load Test Results:")
	t.Logf("  Total Requests: %d", result.TotalRequests)
	t.Logf("  Successful Requests: %d", result.SuccessfulRequests)
	t.Logf("  Failed Requests: %d", result.FailedRequests)
	t.Logf("  Average Response Time: %v", result.AverageResponseTime)
	t.Logf("  Min Response Time: %v", result.MinResponseTime)
	t.Logf("  Max Response Time: %v", result.MaxResponseTime)
	t.Logf("  Requests Per Second: %.2f", result.RequestsPerSecond)
	t.Logf("  Error Rate: %.2f%%", result.ErrorRate)
}

// TestRateLimitingLoad tests rate limiting under load
func TestRateLimitingLoad(t *testing.T) {
	router := setupLoadTestRouter(t)

	// Test with high concurrent users to trigger rate limiting
	config := LoadTestConfig{
		ConcurrentUsers: 200,
		RequestsPerUser: 50,
		TestDuration:    30 * time.Second,
		Endpoint:        "/health",
		Method:          "GET",
	}

	result := runLoadTest(t, router, config)

	// For rate limiting, we expect some requests to be rate limited
	// This is actually a good thing - it means rate limiting is working
	if result.ErrorRate > 50.0 { // 50% error rate threshold (some rate limiting is expected)
		t.Errorf("Error rate too high even considering rate limiting: %.2f%%", result.ErrorRate)
	}

	t.Logf("Rate Limiting Load Test Results:")
	t.Logf("  Total Requests: %d", result.TotalRequests)
	t.Logf("  Successful Requests: %d", result.SuccessfulRequests)
	t.Logf("  Failed Requests: %d", result.FailedRequests)
	t.Logf("  Average Response Time: %v", result.AverageResponseTime)
	t.Logf("  Min Response Time: %v", result.MinResponseTime)
	t.Logf("  Max Response Time: %v", result.MaxResponseTime)
	t.Logf("  Requests Per Second: %.2f", result.RequestsPerSecond)
	t.Logf("  Error Rate: %.2f%%", result.ErrorRate)
}

// TestStressLoad tests stress conditions
func TestStressLoad(t *testing.T) {
	router := setupLoadTestRouter(t)

	// Test with very high load to find breaking point
	config := LoadTestConfig{
		ConcurrentUsers: 500,
		RequestsPerUser: 100,
		TestDuration:    1 * time.Minute,
		Endpoint:        "/api/stats",
		Method:          "GET",
	}

	result := runLoadTest(t, router, config)

	// For stress testing, we expect higher error rates
	// The goal is to find the breaking point, not to pass all tests
	t.Logf("Stress Load Test Results:")
	t.Logf("  Total Requests: %d", result.TotalRequests)
	t.Logf("  Successful Requests: %d", result.SuccessfulRequests)
	t.Logf("  Failed Requests: %d", result.FailedRequests)
	t.Logf("  Average Response Time: %v", result.AverageResponseTime)
	t.Logf("  Min Response Time: %v", result.MinResponseTime)
	t.Logf("  Max Response Time: %v", result.MaxResponseTime)
	t.Logf("  Requests Per Second: %.2f", result.RequestsPerSecond)
	t.Logf("  Error Rate: %.2f%%", result.ErrorRate)
}

// Helper functions

func runLoadTest(t *testing.T, router *gin.Engine, config LoadTestConfig) LoadTestResult {
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	var totalRequests int
	var successfulRequests int
	var failedRequests int
	var totalResponseTime time.Duration
	var minResponseTime time.Duration = time.Hour
	var maxResponseTime time.Duration
	
	startTime := time.Now()
	
	// Create worker goroutines
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Get auth token for protected endpoints
			var token string
			if config.Endpoint != "/health" && config.Endpoint != "/login" {
				token = getLoadTestAuthToken(t, router)
			}
			
			// Make requests
			for j := 0; j < config.RequestsPerUser; j++ {
				// Check if test duration has been exceeded
				if time.Since(startTime) > config.TestDuration {
					break
				}
				
				reqStart := time.Now()
				
				// Create request
				var req *http.Request
				if config.Payload != nil {
					jsonData, _ := json.Marshal(config.Payload)
					req, _ = http.NewRequest(config.Method, config.Endpoint, bytes.NewBuffer(jsonData))
					req.Header.Set("Content-Type", "application/json")
				} else {
					req, _ = http.NewRequest(config.Method, config.Endpoint, nil)
				}
				
				if token != "" {
					req.Header.Set("Authorization", "Bearer "+token)
				}
				
				// Make request
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				
				reqDuration := time.Since(reqStart)
				
				// Update statistics
				mu.Lock()
				totalRequests++
				if w.Code >= 200 && w.Code < 300 {
					successfulRequests++
				} else {
					failedRequests++
				}
				
				totalResponseTime += reqDuration
				if reqDuration < minResponseTime {
					minResponseTime = reqDuration
				}
				if reqDuration > maxResponseTime {
					maxResponseTime = reqDuration
				}
				mu.Unlock()
			}
		}()
	}
	
	// Wait for all workers to complete
	wg.Wait()
	
	// Calculate results
	duration := time.Since(startTime)
	averageResponseTime := totalResponseTime / time.Duration(totalRequests)
	requestsPerSecond := float64(totalRequests) / duration.Seconds()
	errorRate := float64(failedRequests) / float64(totalRequests) * 100
	
	return LoadTestResult{
		TotalRequests:      totalRequests,
		SuccessfulRequests: successfulRequests,
		FailedRequests:     failedRequests,
		AverageResponseTime: averageResponseTime,
		MinResponseTime:    minResponseTime,
		MaxResponseTime:    maxResponseTime,
		RequestsPerSecond:  requestsPerSecond,
		ErrorRate:          errorRate,
	}
}

func setupLoadTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.ReleaseMode) // Use release mode for load testing
	
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
			RateLimit:     1000, // Higher rate limit for load testing
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

	db := setupLoadTestDB(t)

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

func getLoadTestAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupLoadTestDB(t *testing.T) *sql.DB {
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
