package security

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

// TestSQLInjection tests for SQL injection vulnerabilities
func TestSQLInjection(t *testing.T) {
	router := setupSecurityTestRouter(t)
	token := getSecurityTestAuthToken(t, router)

	// Test SQL injection in username field
	sqlInjectionPayloads := []string{
		"admin'; DROP TABLE users; --",
		"admin' OR '1'='1",
		"admin' UNION SELECT * FROM users --",
		"admin' AND 1=1 --",
		"admin' OR 1=1 --",
	}

	for _, payload := range sqlInjectionPayloads {
		loginData := map[string]string{
			"username": payload,
			"password": "admin123",
		}

		jsonData, _ := json.Marshal(loginData)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 401 (unauthorized) not 500 (server error)
		if w.Code == http.StatusInternalServerError {
			t.Errorf("SQL injection vulnerability detected with payload: %s", payload)
		}
	}
}

// TestXSS tests for Cross-Site Scripting vulnerabilities
func TestXSS(t *testing.T) {
	router := setupSecurityTestRouter(t)
	token := getSecurityTestAuthToken(t, router)

	// Test XSS in various fields
	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"javascript:alert('XSS')",
		"<img src=x onerror=alert('XSS')>",
		"<svg onload=alert('XSS')>",
		"';alert('XSS');//",
	}

	for _, payload := range xssPayloads {
		// Test XSS in username field
		loginData := map[string]string{
			"username": payload,
			"password": "admin123",
		}

		jsonData, _ := json.Marshal(loginData)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check if XSS payload is reflected in response
		if bytes.Contains(w.Body.Bytes(), []byte(payload)) {
			t.Errorf("XSS vulnerability detected with payload: %s", payload)
		}
	}
}

// TestCSRF tests for Cross-Site Request Forgery vulnerabilities
func TestCSRF(t *testing.T) {
	router := setupSecurityTestRouter(t)
	token := getSecurityTestAuthToken(t, router)

	// Test CSRF protection on state-changing operations
	csrfEndpoints := []string{
		"/api/moodle/start",
		"/api/moodle/stop",
		"/api/moodle/restart",
	}

	for _, endpoint := range csrfEndpoints {
		req, _ := http.NewRequest("POST", endpoint, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		// Don't set Origin header to simulate CSRF attack

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should still work as we're using JWT tokens (stateless)
		// But we should check for proper CORS headers
		if w.Header().Get("Access-Control-Allow-Origin") == "" {
			t.Errorf("Missing CORS headers for endpoint: %s", endpoint)
		}
	}
}

// TestRateLimiting tests rate limiting functionality
func TestRateLimiting(t *testing.T) {
	router := setupSecurityTestRouter(t)

	// Make multiple requests quickly to trigger rate limiting
	for i := 0; i < 150; i++ { // Exceed the 100 req/min limit
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i >= 100 && w.Code == http.StatusTooManyRequests {
			// Rate limiting is working
			return
		}
	}

	t.Error("Rate limiting not working properly")
}

// TestAuthenticationBypass tests for authentication bypass vulnerabilities
func TestAuthenticationBypass(t *testing.T) {
	router := setupSecurityTestRouter(t)

	// Test various authentication bypass attempts
	bypassAttempts := []struct {
		name        string
		headers     map[string]string
		expectedCode int
	}{
		{
			name: "No Authorization header",
			headers: map[string]string{},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Invalid token format",
			headers: map[string]string{
				"Authorization": "InvalidToken",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Empty token",
			headers: map[string]string{
				"Authorization": "Bearer ",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Malformed JWT",
			headers: map[string]string{
				"Authorization": "Bearer invalid.jwt.token",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "Expired token",
			headers: map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, attempt := range bypassAttempts {
		t.Run(attempt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/stats", nil)
			for key, value := range attempt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != attempt.expectedCode {
				t.Errorf("Expected status %d, got %d for %s", attempt.expectedCode, w.Code, attempt.name)
			}
		})
	}
}

// TestInputValidation tests input validation and sanitization
func TestInputValidation(t *testing.T) {
	router := setupSecurityTestRouter(t)
	token := getSecurityTestAuthToken(t, router)

	// Test various invalid inputs
	invalidInputs := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "Empty username",
			payload: map[string]interface{}{
				"username": "",
				"password": "admin123",
			},
		},
		{
			name: "Empty password",
			payload: map[string]interface{}{
				"username": "admin",
				"password": "",
			},
		},
		{
			name: "Very long username",
			payload: map[string]interface{}{
				"username": string(make([]byte, 1000)),
				"password": "admin123",
			},
		},
		{
			name: "Very long password",
			payload: map[string]interface{}{
				"username": "admin",
				"password": string(make([]byte, 1000)),
			},
		},
		{
			name: "Null values",
			payload: map[string]interface{}{
				"username": nil,
				"password": nil,
			},
		},
	}

	for _, input := range invalidInputs {
		t.Run(input.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(input.payload)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should return 400 (bad request) for invalid input
			if w.Code != http.StatusBadRequest && w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status 400 or 401, got %d for %s", w.Code, input.name)
			}
		})
	}
}

// TestSecurityHeaders tests for proper security headers
func TestSecurityHeaders(t *testing.T) {
	router := setupSecurityTestRouter(t)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check for security headers
	securityHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Content-Security-Policy":   "default-src 'self'",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
	}

	for header, expectedValue := range securityHeaders {
		actualValue := w.Header().Get(header)
		if actualValue == "" {
			t.Errorf("Missing security header: %s", header)
		} else if actualValue != expectedValue {
			t.Errorf("Incorrect value for header %s: expected %s, got %s", header, expectedValue, actualValue)
		}
	}
}

// TestPasswordSecurity tests password security measures
func TestPasswordSecurity(t *testing.T) {
	// Test password hashing
	password := "TestPassword123!"
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Hash should not be the same as password
	if hash == password {
		t.Error("Password hash should not be the same as password")
	}

	// Hash should be different each time (due to salt)
	hash2, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == hash2 {
		t.Error("Password hashes should be different due to salt")
	}

	// Test password verification
	if !utils.CheckPasswordHash(password, hash) {
		t.Error("Password verification should succeed for correct password")
	}

	if utils.CheckPasswordHash("wrongpassword", hash) {
		t.Error("Password verification should fail for incorrect password")
	}
}

// TestJWTTokenSecurity tests JWT token security
func TestJWTTokenSecurity(t *testing.T) {
	authService := services.NewAuthService("test-secret", nil)

	// Test token generation
	user := &models.User{
		ID:       "test-id",
		Username: "testuser",
		Role:     "admin",
	}

	token, expiresAt, err := authService.generateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	if expiresAt.IsZero() {
		t.Error("Expiration time should not be zero")
	}

	// Test token validation
	claims, err := authService.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}

	if claims.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, claims.Username)
	}

	// Test invalid token
	_, err = authService.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Token validation should fail for invalid token")
	}

	// Test token with wrong secret
	wrongAuthService := services.NewAuthService("wrong-secret", nil)
	_, err = wrongAuthService.ValidateToken(token)
	if err == nil {
		t.Error("Token validation should fail with wrong secret")
	}
}

// TestIPWhitelist tests IP whitelist functionality
func TestIPWhitelist(t *testing.T) {
	securityService := services.NewSecurityService(config.SecurityConfig{
		JWTSecret:     "test-secret",
		SessionTimeout: 3600,
		RateLimit:     100,
		AllowedIPs:   []string{"127.0.0.1", "192.168.1.0/24"},
	})

	// Test allowed IP
	if !securityService.isIPAllowed("127.0.0.1") {
		t.Error("127.0.0.1 should be allowed")
	}

	if !securityService.isIPAllowed("192.168.1.100") {
		t.Error("192.168.1.100 should be allowed")
	}

	// Test disallowed IP
	if securityService.isIPAllowed("10.0.0.1") {
		t.Error("10.0.0.1 should not be allowed")
	}

	if securityService.isIPAllowed("8.8.8.8") {
		t.Error("8.8.8.8 should not be allowed")
	}
}

// TestLoggingSecurity tests security event logging
func TestLoggingSecurity(t *testing.T) {
	db := setupSecurityTestDB(t)
	defer db.Close()

	authService := services.NewAuthService("test-secret", db)

	// Test failed login logging
	_, err := authService.Login("nonexistent", "wrongpassword", "192.168.1.100", "test-agent")
	if err == nil {
		t.Error("Login should fail for nonexistent user")
	}

	// Check if security event was logged
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM security_events WHERE type = 'login_failed'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check security events: %v", err)
	}

	if count == 0 {
		t.Error("Security event should be logged for failed login")
	}
}

// Helper functions for security tests

func setupSecurityTestRouter(t *testing.T) *gin.Engine {
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

	db := setupSecurityTestDB(t)

	authService := services.NewAuthService(cfg.Security.JWTSecret, db)
	monitorService := services.NewMonitorService(cfg.Monitoring)
	moodleService := services.NewMoodleService(cfg.Moodle)
	securityService := services.NewSecurityService(cfg.Security)

	authHandler := handlers.NewAuthHandler(authService)
	apiHandler := handlers.NewAPIHandler(monitorService, moodleService, securityService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(securityService.SecurityHeadersMiddleware())
	router.Use(securityService.RateLimitMiddleware())

	router.POST("/login", authHandler.Login)

	protected := router.Group("/api")
	protected.Use(authHandler.AuthMiddleware())
	{
		protected.GET("/stats", apiHandler.GetStats)
		protected.GET("/moodle/status", apiHandler.GetMoodleStatus)
		protected.POST("/moodle/start", apiHandler.StartMoodle)
		protected.POST("/moodle/stop", apiHandler.StopMoodle)
		protected.POST("/moodle/restart", apiHandler.RestartMoodle)
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

func getSecurityTestAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupSecurityTestDB(t *testing.T) *sql.DB {
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
