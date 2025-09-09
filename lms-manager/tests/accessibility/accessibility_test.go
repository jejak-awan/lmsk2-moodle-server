package accessibility

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"lms-manager/config"
	"lms-manager/handlers"
	"lms-manager/services"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// TestAccessibilityWebContent tests web content accessibility
func TestAccessibilityWebContent(t *testing.T) {
	// Test 1: HTML structure accessibility
	t.Run("HTMLStructureAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)
		token := getAccessibilityAuthToken(t, router)

		// Test dashboard page
		req, _ := http.NewRequest("GET", "/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Dashboard page failed with status %d", w.Code)
		}

		// Check for proper HTML structure
		body := w.Body.String()
		
		// Check for DOCTYPE
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Error("Missing DOCTYPE declaration")
		}

		// Check for html lang attribute
		if !strings.Contains(body, `<html lang="en">`) {
			t.Error("Missing html lang attribute")
		}

		// Check for proper heading structure
		if !strings.Contains(body, "<h1>") {
			t.Error("Missing h1 heading")
		}

		// Check for proper form structure
		if strings.Contains(body, "<form") {
			if !strings.Contains(body, `method="post"`) {
				t.Error("Form missing method attribute")
			}
		}
	})

	// Test 2: Form accessibility
	t.Run("FormAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test login form
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for form labels
		if strings.Contains(body, `<input type="text"`) {
			if !strings.Contains(body, `<label`) {
				t.Error("Text input missing label")
			}
		}

		// Check for form labels
		if strings.Contains(body, `<input type="password"`) {
			if !strings.Contains(body, `<label`) {
				t.Error("Password input missing label")
			}
		}

		// Check for form validation
		if strings.Contains(body, `<input`) {
			if !strings.Contains(body, `required`) {
				t.Error("Form inputs missing required attribute")
			}
		}

		// Check for form submission
		if strings.Contains(body, `<form`) {
			if !strings.Contains(body, `<button`) && !strings.Contains(body, `<input type="submit"`) {
				t.Error("Form missing submit button")
			}
		}
	})

	// Test 3: Navigation accessibility
	t.Run("NavigationAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)
		token := getAccessibilityAuthToken(t, router)

		// Test dashboard navigation
		req, _ := http.NewRequest("GET", "/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Dashboard page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for navigation structure
		if strings.Contains(body, `<nav`) {
			// Navigation should have proper structure
			if !strings.Contains(body, `<ul`) {
				t.Error("Navigation missing ul structure")
			}
		}

		// Check for skip links
		if !strings.Contains(body, `href="#main"`) && !strings.Contains(body, `href="#content"`) {
			t.Error("Missing skip links for navigation")
		}

		// Check for proper link structure
		if strings.Contains(body, `<a href`) {
			if !strings.Contains(body, `title=`) && !strings.Contains(body, `aria-label=`) {
				t.Error("Links missing title or aria-label")
			}
		}
	})
}

// TestAccessibilityKeyboardNavigation tests keyboard navigation
func TestAccessibilityKeyboardNavigation(t *testing.T) {
	// Test 1: Tab order accessibility
	t.Run("TabOrderAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test login form tab order
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for tabindex attributes
		if strings.Contains(body, `<input`) {
			if !strings.Contains(body, `tabindex=`) {
				t.Error("Form inputs missing tabindex")
			}
		}

		// Check for proper tab order
		if strings.Contains(body, `tabindex="1"`) {
			if !strings.Contains(body, `tabindex="2"`) {
				t.Error("Incomplete tab order sequence")
			}
		}
	})

	// Test 2: Focus management accessibility
	t.Run("FocusManagementAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test focus management
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for focus indicators
		if strings.Contains(body, `<input`) {
			if !strings.Contains(body, `:focus`) && !strings.Contains(body, `focus:`) {
				t.Error("Missing focus indicators")
			}
		}

		// Check for focus management
		if strings.Contains(body, `<button`) {
			if !strings.Contains(body, `onfocus=`) && !strings.Contains(body, `onblur=`) {
				t.Error("Missing focus event handlers")
			}
		}
	})

	// Test 3: Keyboard shortcuts accessibility
	t.Run("KeyboardShortcutsAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test keyboard shortcuts
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for keyboard shortcuts
		if strings.Contains(body, `<form`) {
			if !strings.Contains(body, `accesskey=`) {
				t.Error("Missing keyboard shortcuts")
			}
		}

		// Check for proper shortcut documentation
		if strings.Contains(body, `accesskey=`) {
			if !strings.Contains(body, `title=`) {
				t.Error("Keyboard shortcuts missing title documentation")
			}
		}
	})
}

// TestAccessibilityScreenReader tests screen reader accessibility
func TestAccessibilityScreenReader(t *testing.T) {
	// Test 1: ARIA labels accessibility
	t.Run("ARIALabelsAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test ARIA labels
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for ARIA labels
		if strings.Contains(body, `<input`) {
			if !strings.Contains(body, `aria-label=`) && !strings.Contains(body, `aria-labelledby=`) {
				t.Error("Form inputs missing ARIA labels")
			}
		}

		// Check for ARIA descriptions
		if strings.Contains(body, `<input`) {
			if !strings.Contains(body, `aria-describedby=`) {
				t.Error("Form inputs missing ARIA descriptions")
			}
		}

		// Check for ARIA roles
		if strings.Contains(body, `<button`) {
			if !strings.Contains(body, `role=`) {
				t.Error("Buttons missing ARIA roles")
			}
		}
	})

	// Test 2: Semantic HTML accessibility
	t.Run("SemanticHTMLAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)
		token := getAccessibilityAuthToken(t, router)

		// Test semantic HTML
		req, _ := http.NewRequest("GET", "/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Dashboard page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for semantic elements
		if !strings.Contains(body, `<main`) && !strings.Contains(body, `<article`) {
			t.Error("Missing semantic main content element")
		}

		// Check for proper heading hierarchy
		if strings.Contains(body, `<h1>`) {
			if !strings.Contains(body, `<h2>`) {
				t.Error("Missing h2 heading after h1")
			}
		}

		// Check for proper list structure
		if strings.Contains(body, `<ul`) {
			if !strings.Contains(body, `<li`) {
				t.Error("Unordered list missing list items")
			}
		}

		// Check for proper table structure
		if strings.Contains(body, `<table`) {
			if !strings.Contains(body, `<thead`) && !strings.Contains(body, `<th`) {
				t.Error("Table missing proper header structure")
			}
		}
	})

	// Test 3: Live regions accessibility
	t.Run("LiveRegionsAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test live regions
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for live regions
		if strings.Contains(body, `<div`) {
			if !strings.Contains(body, `aria-live=`) && !strings.Contains(body, `aria-atomic=`) {
				t.Error("Missing live regions for dynamic content")
			}
		}

		// Check for status messages
		if strings.Contains(body, `<div`) {
			if !strings.Contains(body, `role="status"`) && !strings.Contains(body, `role="alert"`) {
				t.Error("Missing status message roles")
			}
		}
	})
}

// TestAccessibilityColorContrast tests color contrast accessibility
func TestAccessibilityColorContrast(t *testing.T) {
	// Test 1: Color contrast accessibility
	t.Run("ColorContrastAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test color contrast
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for color contrast information
		if strings.Contains(body, `color:`) {
			if !strings.Contains(body, `background-color:`) {
				t.Error("Missing background color for text")
			}
		}

		// Check for high contrast mode support
		if strings.Contains(body, `<style`) {
			if !strings.Contains(body, `@media (prefers-contrast: high)`) {
				t.Error("Missing high contrast mode support")
			}
		}
	})

	// Test 2: Color independence accessibility
	t.Run("ColorIndependenceAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test color independence
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for color independence
		if strings.Contains(body, `color: red`) || strings.Contains(body, `color: green`) {
			if !strings.Contains(body, `text-decoration:`) && !strings.Contains(body, `font-weight:`) {
				t.Error("Color-dependent information missing alternative indicators")
			}
		}

		// Check for proper error indication
		if strings.Contains(body, `error`) {
			if !strings.Contains(body, `aria-invalid=`) {
				t.Error("Error states missing ARIA invalid attribute")
			}
		}
	})
}

// TestAccessibilityResponsiveDesign tests responsive design accessibility
func TestAccessibilityResponsiveDesign(t *testing.T) {
	// Test 1: Mobile accessibility
	t.Run("MobileAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test mobile viewport
		req, _ := http.NewRequest("GET", "/login", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for mobile viewport
		if !strings.Contains(body, `viewport`) {
			t.Error("Missing mobile viewport meta tag")
		}

		// Check for touch targets
		if strings.Contains(body, `<button`) {
			if !strings.Contains(body, `min-height: 44px`) && !strings.Contains(body, `min-width: 44px`) {
				t.Error("Touch targets too small for mobile")
			}
		}
	})

	// Test 2: Zoom accessibility
	t.Run("ZoomAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test zoom support
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for zoom support
		if strings.Contains(body, `viewport`) {
			if strings.Contains(body, `user-scalable=no`) {
				t.Error("Zoom disabled - accessibility issue")
			}
		}

		// Check for responsive design
		if strings.Contains(body, `<style`) {
			if !strings.Contains(body, `@media`) {
				t.Error("Missing responsive design media queries")
			}
		}
	})
}

// TestAccessibilityErrorHandling tests error handling accessibility
func TestAccessibilityErrorHandling(t *testing.T) {
	// Test 1: Error message accessibility
	t.Run("ErrorMessageAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test error handling
		loginData := map[string]string{
			"username": "invalid",
			"password": "wrong",
		}

		jsonData, _ := json.Marshal(loginData)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return error
		if w.Code == http.StatusOK {
			t.Error("Invalid login should return error")
		}

		// Check for proper error response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		// Check for error message
		if response["error"] == nil {
			t.Error("Error response missing error message")
		}

		// Check for error code
		if response["code"] == nil {
			t.Error("Error response missing error code")
		}
	})

	// Test 2: Form validation accessibility
	t.Run("FormValidationAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test form validation
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for form validation
		if strings.Contains(body, `<input`) {
			if !strings.Contains(body, `aria-invalid=`) {
				t.Error("Form inputs missing validation attributes")
			}
		}

		// Check for validation messages
		if strings.Contains(body, `<input`) {
			if !strings.Contains(body, `aria-describedby=`) {
				t.Error("Form inputs missing validation message references")
			}
		}
	})
}

// TestAccessibilityPerformance tests performance accessibility
func TestAccessibilityPerformance(t *testing.T) {
	// Test 1: Loading performance accessibility
	t.Run("LoadingPerformanceAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test loading performance
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for loading indicators
		if strings.Contains(body, `<img`) {
			if !strings.Contains(body, `alt=`) {
				t.Error("Images missing alt text")
			}
		}

		// Check for loading states
		if strings.Contains(body, `<button`) {
			if !strings.Contains(body, `aria-busy=`) {
				t.Error("Buttons missing loading state indicators")
			}
		}
	})

	// Test 2: Timeout accessibility
	t.Run("TimeoutAccessibility", func(t *testing.T) {
		router := setupAccessibilityTestRouter(t)

		// Test timeout handling
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for timeout warnings
		if strings.Contains(body, `<form`) {
			if !strings.Contains(body, `aria-live=`) {
				t.Error("Forms missing timeout warnings")
			}
		}

		// Check for session management
		if strings.Contains(body, `<script`) {
			if !strings.Contains(body, `setTimeout`) && !strings.Contains(body, `setInterval`) {
				t.Error("Missing session timeout management")
			}
		}
	})
}

// Helper functions

func setupAccessibilityTestRouter(t *testing.T) *gin.Engine {
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

	db := setupAccessibilityTestDB(t)

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

	// Add accessibility test routes
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login - LMS Manager",
		})
	})

	router.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"title": "Dashboard - LMS Manager",
		})
	})

	return router
}

func getAccessibilityAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupAccessibilityTestDB(t *testing.T) *sql.DB {
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
