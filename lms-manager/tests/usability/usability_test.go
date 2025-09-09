package usability

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"lms-manager/config"
	"lms-manager/handlers"
	"lms-manager/services"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// TestUsabilityUserInterface tests user interface usability
func TestUsabilityUserInterface(t *testing.T) {
	// Test 1: Interface clarity
	t.Run("InterfaceClarity", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)
		token := getUsabilityAuthToken(t, router)

		// Test dashboard clarity
		req, _ := http.NewRequest("GET", "/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Dashboard page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for clear navigation
		if !strings.Contains(body, "Dashboard") {
			t.Error("Dashboard page missing clear title")
		}

		// Check for clear sections
		if !strings.Contains(body, "System Status") && !strings.Contains(body, "Statistics") {
			t.Error("Dashboard missing clear section headers")
		}

		// Check for clear actions
		if !strings.Contains(body, "Start") && !strings.Contains(body, "Stop") && !strings.Contains(body, "Restart") {
			t.Error("Dashboard missing clear action buttons")
		}
	})

	// Test 2: Information hierarchy
	t.Run("InformationHierarchy", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)
		token := getUsabilityAuthToken(t, router)

		// Test information hierarchy
		req, _ := http.NewRequest("GET", "/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Dashboard page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for proper heading hierarchy
		if strings.Contains(body, "<h1>") {
			if !strings.Contains(body, "<h2>") {
				t.Error("Missing h2 headings after h1")
			}
		}

		// Check for logical grouping
		if strings.Contains(body, "System") {
			if !strings.Contains(body, "Status") && !strings.Contains(body, "Health") {
				t.Error("System information not properly grouped")
			}
		}

		// Check for priority information
		if !strings.Contains(body, "CPU") && !strings.Contains(body, "Memory") && !strings.Contains(body, "Disk") {
			t.Error("Critical system information not prominently displayed")
		}
	})

	// Test 3: Visual feedback
	t.Run("VisualFeedback", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)
		token := getUsabilityAuthToken(t, router)

		// Test visual feedback
		req, _ := http.NewRequest("GET", "/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Dashboard page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for status indicators
		if !strings.Contains(body, "status") && !strings.Contains(body, "indicator") {
			t.Error("Missing status indicators")
		}

		// Check for progress indicators
		if strings.Contains(body, "progress") {
			if !strings.Contains(body, "bar") && !strings.Contains(body, "percentage") {
				t.Error("Progress indicators not properly implemented")
			}
		}

		// Check for color coding
		if strings.Contains(body, "error") || strings.Contains(body, "warning") {
			if !strings.Contains(body, "red") && !strings.Contains(body, "yellow") && !strings.Contains(body, "green") {
				t.Error("Missing color coding for status")
			}
		}
	})
}

// TestUsabilityNavigation tests navigation usability
func TestUsabilityNavigation(t *testing.T) {
	// Test 1: Navigation consistency
	t.Run("NavigationConsistency", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)
		token := getUsabilityAuthToken(t, router)

		// Test navigation consistency
		req, _ := http.NewRequest("GET", "/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Dashboard page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for consistent navigation
		if strings.Contains(body, "<nav") {
			if !strings.Contains(body, "<ul") && !strings.Contains(body, "<ol") {
				t.Error("Navigation missing proper list structure")
			}
		}

		// Check for breadcrumbs
		if !strings.Contains(body, "breadcrumb") && !strings.Contains(body, "path") {
			t.Error("Missing breadcrumb navigation")
		}

		// Check for active states
		if strings.Contains(body, "active") {
			if !strings.Contains(body, "current") && !strings.Contains(body, "selected") {
				t.Error("Active navigation states not properly implemented")
			}
		}
	})

	// Test 2: Navigation efficiency
	t.Run("NavigationEfficiency", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)
		token := getUsabilityAuthToken(t, router)

		// Test navigation efficiency
		req, _ := http.NewRequest("GET", "/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Dashboard page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for quick access
		if !strings.Contains(body, "shortcut") && !strings.Contains(body, "quick") {
			t.Error("Missing quick access features")
		}

		// Check for search functionality
		if !strings.Contains(body, "search") && !strings.Contains(body, "filter") {
			t.Error("Missing search/filter functionality")
		}

		// Check for keyboard shortcuts
		if !strings.Contains(body, "accesskey") && !strings.Contains(body, "shortcut") {
			t.Error("Missing keyboard shortcuts")
		}
	})

	// Test 3: Navigation depth
	t.Run("NavigationDepth", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)
		token := getUsabilityAuthToken(t, router)

		// Test navigation depth
		req, _ := http.NewRequest("GET", "/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Dashboard page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for reasonable navigation depth
		if strings.Contains(body, "submenu") || strings.Contains(body, "dropdown") {
			if !strings.Contains(body, "level") && !strings.Contains(body, "depth") {
				t.Error("Navigation depth not properly managed")
			}
		}

		// Check for back navigation
		if !strings.Contains(body, "back") && !strings.Contains(body, "previous") {
			t.Error("Missing back navigation")
		}

		// Check for home navigation
		if !strings.Contains(body, "home") && !strings.Contains(body, "main") {
			t.Error("Missing home navigation")
		}
	})
}

// TestUsabilityForms tests form usability
func TestUsabilityForms(t *testing.T) {
	// Test 1: Form clarity
	t.Run("FormClarity", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test form clarity
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for clear form labels
		if strings.Contains(body, "<input") {
			if !strings.Contains(body, "<label") {
				t.Error("Form inputs missing clear labels")
			}
		}

		// Check for form instructions
		if strings.Contains(body, "<form") {
			if !strings.Contains(body, "instruction") && !strings.Contains(body, "help") {
				t.Error("Form missing instructions or help text")
			}
		}

		// Check for required field indicators
		if strings.Contains(body, "required") {
			if !strings.Contains(body, "asterisk") && !strings.Contains(body, "required") {
				t.Error("Required fields not clearly indicated")
			}
		}
	})

	// Test 2: Form validation
	t.Run("FormValidation", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test form validation
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for real-time validation
		if strings.Contains(body, "<input") {
			if !strings.Contains(body, "onblur") && !strings.Contains(body, "onchange") {
				t.Error("Form inputs missing real-time validation")
			}
		}

		// Check for error messages
		if strings.Contains(body, "error") {
			if !strings.Contains(body, "message") && !strings.Contains(body, "alert") {
				t.Error("Error messages not properly implemented")
			}
		}

		// Check for success feedback
		if strings.Contains(body, "success") {
			if !strings.Contains(body, "confirmation") && !strings.Contains(body, "checkmark") {
				t.Error("Success feedback not properly implemented")
			}
		}
	})

	// Test 3: Form efficiency
	t.Run("FormEfficiency", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test form efficiency
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for auto-focus
		if strings.Contains(body, "<input") {
			if !strings.Contains(body, "autofocus") {
				t.Error("Form missing auto-focus on first input")
			}
		}

		// Check for auto-complete
		if strings.Contains(body, "<input") {
			if !strings.Contains(body, "autocomplete") {
				t.Error("Form inputs missing auto-complete")
			}
		}

		// Check for form submission
		if strings.Contains(body, "<form") {
			if !strings.Contains(body, "onsubmit") && !strings.Contains(body, "submit") {
				t.Error("Form submission not properly handled")
			}
		}
	})
}

// TestUsabilityErrorHandling tests error handling usability
func TestUsabilityErrorHandling(t *testing.T) {
	// Test 1: Error clarity
	t.Run("ErrorClarity", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test error clarity
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

		// Check for clear error message
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if response["error"] == nil {
			t.Error("Error response missing clear error message")
		}

		// Check for error context
		if response["context"] == nil && response["details"] == nil {
			t.Error("Error response missing context or details")
		}
	})

	// Test 2: Error recovery
	t.Run("ErrorRecovery", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test error recovery
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

		// Check for recovery suggestions
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if response["suggestion"] == nil && response["help"] == nil {
			t.Error("Error response missing recovery suggestions")
		}

		// Check for retry options
		if response["retry"] == nil && response["retry_after"] == nil {
			t.Error("Error response missing retry options")
		}
	})

	// Test 3: Error prevention
	t.Run("ErrorPrevention", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test error prevention
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for input validation
		if strings.Contains(body, "<input") {
			if !strings.Contains(body, "pattern") && !strings.Contains(body, "minlength") {
				t.Error("Form inputs missing validation patterns")
			}
		}

		// Check for confirmation dialogs
		if strings.Contains(body, "delete") || strings.Contains(body, "remove") {
			if !strings.Contains(body, "confirm") && !strings.Contains(body, "warning") {
				t.Error("Destructive actions missing confirmation")
			}
		}

		// Check for data protection
		if strings.Contains(body, "password") {
			if !strings.Contains(body, "type=\"password\"") {
				t.Error("Password fields not properly protected")
			}
		}
	})
}

// TestUsabilityPerformance tests performance usability
func TestUsabilityPerformance(t *testing.T) {
	// Test 1: Response time
	t.Run("ResponseTime", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test response time
		start := time.Now()
		
		req, _ := http.NewRequest("GET", "/login", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		duration := time.Since(start)
		
		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}
		
		// Check response time is reasonable
		if duration > 500*time.Millisecond {
			t.Errorf("Login page response time too slow: %v", duration)
		}
	})

	// Test 2: Loading indicators
	t.Run("LoadingIndicators", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test loading indicators
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for loading indicators
		if strings.Contains(body, "loading") {
			if !strings.Contains(body, "spinner") && !strings.Contains(body, "progress") {
				t.Error("Loading indicators not properly implemented")
			}
		}

		// Check for progress feedback
		if strings.Contains(body, "progress") {
			if !strings.Contains(body, "percentage") && !strings.Contains(body, "step") {
				t.Error("Progress feedback not properly implemented")
			}
		}

		// Check for timeout handling
		if strings.Contains(body, "timeout") {
			if !strings.Contains(body, "warning") && !strings.Contains(body, "retry") {
				t.Error("Timeout handling not properly implemented")
			}
		}
	})

	// Test 3: Resource optimization
	t.Run("ResourceOptimization", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test resource optimization
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for image optimization
		if strings.Contains(body, "<img") {
			if !strings.Contains(body, "lazy") && !strings.Contains(body, "defer") {
				t.Error("Images not optimized for loading")
			}
		}

		// Check for CSS optimization
		if strings.Contains(body, "<style") {
			if !strings.Contains(body, "minified") && !strings.Contains(body, "compressed") {
				t.Error("CSS not optimized")
			}
		}

		// Check for JavaScript optimization
		if strings.Contains(body, "<script") {
			if !strings.Contains(body, "async") && !strings.Contains(body, "defer") {
				t.Error("JavaScript not optimized for loading")
			}
		}
	})
}

// TestUsabilityAccessibility tests accessibility usability
func TestUsabilityAccessibility(t *testing.T) {
	// Test 1: Keyboard navigation
	t.Run("KeyboardNavigation", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test keyboard navigation
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for keyboard navigation
		if strings.Contains(body, "<input") {
			if !strings.Contains(body, "tabindex") {
				t.Error("Form inputs missing keyboard navigation")
			}
		}

		// Check for keyboard shortcuts
		if strings.Contains(body, "<form") {
			if !strings.Contains(body, "accesskey") {
				t.Error("Forms missing keyboard shortcuts")
			}
		}

		// Check for focus management
		if strings.Contains(body, "<button") {
			if !strings.Contains(body, "onfocus") && !strings.Contains(body, "onblur") {
				t.Error("Buttons missing focus management")
			}
		}
	})

	// Test 2: Screen reader support
	t.Run("ScreenReaderSupport", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test screen reader support
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for ARIA labels
		if strings.Contains(body, "<input") {
			if !strings.Contains(body, "aria-label") && !strings.Contains(body, "aria-labelledby") {
				t.Error("Form inputs missing ARIA labels")
			}
		}

		// Check for ARIA descriptions
		if strings.Contains(body, "<input") {
			if !strings.Contains(body, "aria-describedby") {
				t.Error("Form inputs missing ARIA descriptions")
			}
		}

		// Check for ARIA roles
		if strings.Contains(body, "<button") {
			if !strings.Contains(body, "role") {
				t.Error("Buttons missing ARIA roles")
			}
		}
	})

	// Test 3: Visual accessibility
	t.Run("VisualAccessibility", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test visual accessibility
		req, _ := http.NewRequest("GET", "/login", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for color contrast
		if strings.Contains(body, "color:") {
			if !strings.Contains(body, "background-color:") {
				t.Error("Text missing background color for contrast")
			}
		}

		// Check for text size
		if strings.Contains(body, "font-size:") {
			if !strings.Contains(body, "rem") && !strings.Contains(body, "em") {
				t.Error("Text size not using relative units")
			}
		}

		// Check for alternative text
		if strings.Contains(body, "<img") {
			if !strings.Contains(body, "alt=") {
				t.Error("Images missing alternative text")
			}
		}
	})
}

// TestUsabilityMobile tests mobile usability
func TestUsabilityMobile(t *testing.T) {
	// Test 1: Mobile layout
	t.Run("MobileLayout", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test mobile layout
		req, _ := http.NewRequest("GET", "/login", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for mobile viewport
		if !strings.Contains(body, "viewport") {
			t.Error("Missing mobile viewport meta tag")
		}

		// Check for responsive design
		if !strings.Contains(body, "@media") {
			t.Error("Missing responsive design media queries")
		}

		// Check for touch targets
		if strings.Contains(body, "<button") {
			if !strings.Contains(body, "min-height: 44px") && !strings.Contains(body, "min-width: 44px") {
				t.Error("Touch targets too small for mobile")
			}
		}
	})

	// Test 2: Mobile navigation
	t.Run("MobileNavigation", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test mobile navigation
		req, _ := http.NewRequest("GET", "/login", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for mobile navigation
		if strings.Contains(body, "<nav") {
			if !strings.Contains(body, "hamburger") && !strings.Contains(body, "menu") {
				t.Error("Mobile navigation not properly implemented")
			}
		}

		// Check for swipe gestures
		if strings.Contains(body, "swipe") {
			if !strings.Contains(body, "touchstart") && !strings.Contains(body, "touchmove") {
				t.Error("Swipe gestures not properly implemented")
			}
		}

		// Check for mobile-specific features
		if strings.Contains(body, "mobile") {
			if !strings.Contains(body, "orientation") && !strings.Contains(body, "landscape") {
				t.Error("Mobile-specific features not properly implemented")
			}
		}
	})

	// Test 3: Mobile performance
	t.Run("MobilePerformance", func(t *testing.T) {
		router := setupUsabilityTestRouter(t)

		// Test mobile performance
		req, _ := http.NewRequest("GET", "/login", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Login page failed with status %d", w.Code)
		}

		body := w.Body.String()

		// Check for mobile optimization
		if strings.Contains(body, "<img") {
			if !strings.Contains(body, "lazy") && !strings.Contains(body, "defer") {
				t.Error("Images not optimized for mobile")
			}
		}

		// Check for mobile-specific CSS
		if strings.Contains(body, "<style") {
			if !strings.Contains(body, "mobile") && !strings.Contains(body, "touch") {
				t.Error("Missing mobile-specific CSS")
			}
		}

		// Check for mobile-specific JavaScript
		if strings.Contains(body, "<script") {
			if !strings.Contains(body, "mobile") && !strings.Contains(body, "touch") {
				t.Error("Missing mobile-specific JavaScript")
			}
		}
	})
}

// Helper functions

func setupUsabilityTestRouter(t *testing.T) *gin.Engine {
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

	db := setupUsabilityTestDB(t)

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

	// Add usability test routes
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

func getUsabilityAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupUsabilityTestDB(t *testing.T) *sql.DB {
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
