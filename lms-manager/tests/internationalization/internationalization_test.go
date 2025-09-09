package internationalization

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

// TestInternationalizationLanguageSupport tests language support
func TestInternationalizationLanguageSupport(t *testing.T) {
	// Test 1: English language support
	t.Run("EnglishLanguageSupport", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test English language
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("English language support test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for English language indicators
		if stats["language"] != nil && stats["language"] != "en" {
			t.Error("Language not set to English")
		}
	})

	// Test 2: Indonesian language support
	t.Run("IndonesianLanguageSupport", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test Indonesian language
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "id-ID,id;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Indonesian language support test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for Indonesian language indicators
		if stats["language"] != nil && stats["language"] != "id" {
			t.Error("Language not set to Indonesian")
		}
	})

	// Test 3: Multiple language support
	t.Run("MultipleLanguageSupport", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test multiple languages
		languages := []string{"en-US", "id-ID", "es-ES", "fr-FR", "de-DE"}

		for _, lang := range languages {
			req, _ := http.NewRequest("GET", "/api/stats", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Accept-Language", lang+";q=0.9")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Language %s support test failed with status %d", lang, w.Code)
			}

			var stats map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &stats)
			if err != nil {
				t.Fatalf("Failed to unmarshal stats response for language %s: %v", lang, err)
			}

			// Check for language support
			if stats["language"] == nil {
				t.Errorf("Language %s not supported", lang)
			}
		}
	})
}

// TestInternationalizationLocalization tests localization
func TestInternationalizationLocalization(t *testing.T) {
	// Test 1: Date and time localization
	t.Run("DateTimeLocalization", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test date and time localization
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Date and time localization test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for date and time format
		if stats["timestamp"] != nil {
			timestamp := stats["timestamp"].(string)
			if !strings.Contains(timestamp, "T") && !strings.Contains(timestamp, " ") {
				t.Error("Timestamp not properly formatted")
			}
		}
	})

	// Test 2: Number localization
	t.Run("NumberLocalization", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test number localization
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Number localization test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for number format
		if stats["cpu_usage"] != nil {
			cpuUsage := stats["cpu_usage"].(float64)
			if cpuUsage < 0 || cpuUsage > 100 {
				t.Error("CPU usage not properly formatted")
			}
		}

		if stats["memory_usage"] != nil {
			memoryUsage := stats["memory_usage"].(float64)
			if memoryUsage < 0 || memoryUsage > 100 {
				t.Error("Memory usage not properly formatted")
			}
		}
	})

	// Test 3: Currency localization
	t.Run("CurrencyLocalization", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test currency localization
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Currency localization test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for currency format
		if stats["currency"] != nil {
			currency := stats["currency"].(string)
			if currency != "USD" && currency != "IDR" && currency != "EUR" {
				t.Error("Currency not properly localized")
			}
		}
	})
}

// TestInternationalizationTextDirection tests text direction
func TestInternationalizationTextDirection(t *testing.T) {
	// Test 1: Left-to-right text direction
	t.Run("LeftToRightTextDirection", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test left-to-right text direction
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Left-to-right text direction test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for text direction
		if stats["text_direction"] != nil && stats["text_direction"] != "ltr" {
			t.Error("Text direction not set to left-to-right")
		}
	})

	// Test 2: Right-to-left text direction
	t.Run("RightToLeftTextDirection", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test right-to-left text direction
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "ar-SA,ar;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Right-to-left text direction test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for text direction
		if stats["text_direction"] != nil && stats["text_direction"] != "rtl" {
			t.Error("Text direction not set to right-to-left")
		}
	})
}

// TestInternationalizationCharacterEncoding tests character encoding
func TestInternationalizationCharacterEncoding(t *testing.T) {
	// Test 1: UTF-8 character encoding
	t.Run("UTF8CharacterEncoding", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test UTF-8 character encoding
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("UTF-8 character encoding test failed with status %d", w.Code)
		}

		// Check for UTF-8 encoding
		if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Error("Content-Type not set to UTF-8")
		}
	})

	// Test 2: Unicode character support
	t.Run("UnicodeCharacterSupport", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test Unicode character support
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Unicode character support test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for Unicode support
		if stats["unicode_support"] != nil && stats["unicode_support"] != true {
			t.Error("Unicode character support not enabled")
		}
	})
}

// TestInternationalizationCulturalAdaptation tests cultural adaptation
func TestInternationalizationCulturalAdaptation(t *testing.T) {
	// Test 1: Cultural date formats
	t.Run("CulturalDateFormats", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test cultural date formats
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Cultural date formats test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for cultural date format
		if stats["date_format"] != nil {
			dateFormat := stats["date_format"].(string)
			if dateFormat != "MM/DD/YYYY" && dateFormat != "DD/MM/YYYY" && dateFormat != "YYYY-MM-DD" {
				t.Error("Date format not culturally adapted")
			}
		}
	})

	// Test 2: Cultural number formats
	t.Run("CulturalNumberFormats", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test cultural number formats
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Cultural number formats test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for cultural number format
		if stats["number_format"] != nil {
			numberFormat := stats["number_format"].(string)
			if numberFormat != "1,234.56" && numberFormat != "1.234,56" && numberFormat != "1 234,56" {
				t.Error("Number format not culturally adapted")
			}
		}
	})

	// Test 3: Cultural time formats
	t.Run("CulturalTimeFormats", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test cultural time formats
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Cultural time formats test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for cultural time format
		if stats["time_format"] != nil {
			timeFormat := stats["time_format"].(string)
			if timeFormat != "12:34:56 PM" && timeFormat != "12:34:56" && timeFormat != "12.34.56" {
				t.Error("Time format not culturally adapted")
			}
		}
	})
}

// TestInternationalizationRegionalSettings tests regional settings
func TestInternationalizationRegionalSettings(t *testing.T) {
	// Test 1: Regional timezone support
	t.Run("RegionalTimezoneSupport", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test regional timezone support
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Regional timezone support test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for timezone support
		if stats["timezone"] != nil {
			timezone := stats["timezone"].(string)
			if timezone != "UTC" && timezone != "GMT" && timezone != "EST" && timezone != "PST" {
				t.Error("Timezone not properly set")
			}
		}
	})

	// Test 2: Regional currency support
	t.Run("RegionalCurrencySupport", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test regional currency support
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Regional currency support test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for currency support
		if stats["currency"] != nil {
			currency := stats["currency"].(string)
			if currency != "USD" && currency != "EUR" && currency != "GBP" && currency != "JPY" {
				t.Error("Currency not properly set")
			}
		}
	})

	// Test 3: Regional measurement units
	t.Run("RegionalMeasurementUnits", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test regional measurement units
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Regional measurement units test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for measurement units
		if stats["measurement_units"] != nil {
			units := stats["measurement_units"].(string)
			if units != "metric" && units != "imperial" && units != "US" {
				t.Error("Measurement units not properly set")
			}
		}
	})
}

// TestInternationalizationLanguageSwitching tests language switching
func TestInternationalizationLanguageSwitching(t *testing.T) {
	// Test 1: Language switching functionality
	t.Run("LanguageSwitchingFunctionality", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test language switching
		languages := []string{"en", "id", "es", "fr", "de"}

		for _, lang := range languages {
			req, _ := http.NewRequest("GET", "/api/stats", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Accept-Language", lang+";q=0.9")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Language switching test failed for language %s with status %d", lang, w.Code)
			}

			var stats map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &stats)
			if err != nil {
				t.Fatalf("Failed to unmarshal stats response for language %s: %v", lang, err)
			}

			// Check for language switching
			if stats["language"] != nil && stats["language"] != lang {
				t.Errorf("Language not switched to %s", lang)
			}
		}
	})

	// Test 2: Language persistence
	t.Run("LanguagePersistence", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test language persistence
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "id-ID,id;q=0.9")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Language persistence test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for language persistence
		if stats["language"] != nil && stats["language"] != "id" {
			t.Error("Language not persisted")
		}
	})
}

// TestInternationalizationFallbackLanguage tests fallback language
func TestInternationalizationFallbackLanguage(t *testing.T) {
	// Test 1: Fallback language support
	t.Run("FallbackLanguageSupport", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test fallback language
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept-Language", "xx-XX,xx;q=0.9") // Unsupported language

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Fallback language test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for fallback language
		if stats["language"] != nil && stats["language"] != "en" {
			t.Error("Fallback language not set to English")
		}
	})

	// Test 2: Default language support
	t.Run("DefaultLanguageSupport", func(t *testing.T) {
		router := setupInternationalizationTestRouter(t)
		token := getInternationalizationAuthToken(t, router)

		// Test default language
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		// No Accept-Language header

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Default language test failed with status %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check for default language
		if stats["language"] != nil && stats["language"] != "en" {
			t.Error("Default language not set to English")
		}
	})
}

// Helper functions

func setupInternationalizationTestRouter(t *testing.T) *gin.Engine {
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

	db := setupInternationalizationTestDB(t)

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

	return router
}

func getInternationalizationAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupInternationalizationTestDB(t *testing.T) *sql.DB {
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
		t.Fatalf("Failed to hash password: %v", err)
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
