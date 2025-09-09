package compliance

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

// TestComplianceGDPR tests GDPR compliance
func TestComplianceGDPR(t *testing.T) {
	// Test 1: Data protection
	t.Run("DataProtection", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test data protection
		req, _ := http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Data protection test failed with status %d", w.Code)
		}

		var users []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &users)
		if err != nil {
			t.Fatalf("Failed to unmarshal users response: %v", err)
		}

		// Check for data protection
		for _, user := range users {
			// Check for sensitive data protection
			if user["password"] != nil {
				t.Error("Password data exposed in user list")
			}

			// Check for data minimization
			if user["ssn"] != nil || user["credit_card"] != nil {
				t.Error("Unnecessary sensitive data exposed")
			}

			// Check for data anonymization
			if user["email"] != nil {
				email := user["email"].(string)
				if !strings.Contains(email, "@") {
					t.Error("Email data not properly protected")
				}
			}
		}
	})

	// Test 2: Right to access
	t.Run("RightToAccess", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test right to access
		req, _ := http.NewRequest("GET", "/api/users/profile", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Right to access test failed with status %d", w.Code)
		}

		var profile map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &profile)
		if err != nil {
			t.Fatalf("Failed to unmarshal profile response: %v", err)
		}

		// Check for data access
		if profile["username"] == nil {
			t.Error("User data not accessible")
		}

		if profile["email"] == nil {
			t.Error("Email data not accessible")
		}

		if profile["created_at"] == nil {
			t.Error("Account creation date not accessible")
		}
	})

	// Test 3: Right to rectification
	t.Run("RightToRectification", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test right to rectification
		updateData := map[string]interface{}{
			"email": "updated@example.com",
			"name":  "Updated Name",
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", "/api/users/profile", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Right to rectification test failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal update response: %v", err)
		}

		// Check for data update
		if response["email"] != "updated@example.com" {
			t.Error("Email data not updated")
		}

		if response["name"] != "Updated Name" {
			t.Error("Name data not updated")
		}
	})

	// Test 4: Right to erasure
	t.Run("RightToErasure", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test right to erasure
		req, _ := http.NewRequest("DELETE", "/api/users/profile", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Right to erasure test failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal deletion response: %v", err)
		}

		// Check for data deletion
		if response["deleted"] != true {
			t.Error("User data not deleted")
		}

		if response["confirmation"] == nil {
			t.Error("Deletion confirmation not provided")
		}
	})

	// Test 5: Data portability
	t.Run("DataPortability", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test data portability
		req, _ := http.NewRequest("GET", "/api/users/export", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Data portability test failed with status %d", w.Code)
		}

		// Check for data export
		if w.Header().Get("Content-Type") != "application/json" {
			t.Error("Data export not in JSON format")
		}

		var exportData map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &exportData)
		if err != nil {
			t.Fatalf("Failed to unmarshal export response: %v", err)
		}

		// Check for complete data export
		if exportData["user_data"] == nil {
			t.Error("User data not exported")
		}

		if exportData["export_date"] == nil {
			t.Error("Export date not provided")
		}
	})
}

// TestComplianceCCPA tests CCPA compliance
func TestComplianceCCPA(t *testing.T) {
	// Test 1: Right to know
	t.Run("RightToKnow", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test right to know
		req, _ := http.NewRequest("GET", "/api/users/data-collection", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Right to know test failed with status %d", w.Code)
		}

		var dataCollection map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &dataCollection)
		if err != nil {
			t.Fatalf("Failed to unmarshal data collection response: %v", err)
		}

		// Check for data collection disclosure
		if dataCollection["categories"] == nil {
			t.Error("Data categories not disclosed")
		}

		if dataCollection["purposes"] == nil {
			t.Error("Data collection purposes not disclosed")
		}

		if dataCollection["third_parties"] == nil {
			t.Error("Third-party sharing not disclosed")
		}
	})

	// Test 2: Right to delete
	t.Run("RightToDelete", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test right to delete
		req, _ := http.NewRequest("DELETE", "/api/users/personal-data", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Right to delete test failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal deletion response: %v", err)
		}

		// Check for data deletion
		if response["deleted"] != true {
			t.Error("Personal data not deleted")
		}

		if response["categories_deleted"] == nil {
			t.Error("Data categories not specified for deletion")
		}
	})

	// Test 3: Right to opt-out
	t.Run("RightToOptOut", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test right to opt-out
		optOutData := map[string]interface{}{
			"data_sharing": false,
			"marketing":    false,
			"analytics":    false,
		}

		jsonData, _ := json.Marshal(optOutData)
		req, _ := http.NewRequest("POST", "/api/users/opt-out", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Right to opt-out test failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal opt-out response: %v", err)
		}

		// Check for opt-out confirmation
		if response["opt_out_confirmed"] != true {
			t.Error("Opt-out not confirmed")
		}

		if response["effective_date"] == nil {
			t.Error("Opt-out effective date not provided")
		}
	})

	// Test 4: Non-discrimination
	t.Run("NonDiscrimination", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test non-discrimination
		req, _ := http.NewRequest("GET", "/api/users/services", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Non-discrimination test failed with status %d", w.Code)
		}

		var services map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &services)
		if err != nil {
			t.Fatalf("Failed to unmarshal services response: %v", err)
		}

		// Check for service availability
		if services["available"] != true {
			t.Error("Services not available after opt-out")
		}

		if services["quality"] == nil {
			t.Error("Service quality not maintained")
		}
	})
}

// TestComplianceHIPAA tests HIPAA compliance
func TestComplianceHIPAA(t *testing.T) {
	// Test 1: Administrative safeguards
	t.Run("AdministrativeSafeguards", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test administrative safeguards
		req, _ := http.NewRequest("GET", "/api/security/policies", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Administrative safeguards test failed with status %d", w.Code)
		}

		var policies map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &policies)
		if err != nil {
			t.Fatalf("Failed to unmarshal policies response: %v", err)
		}

		// Check for security policies
		if policies["access_control"] == nil {
			t.Error("Access control policies not implemented")
		}

		if policies["workforce_training"] == nil {
			t.Error("Workforce training policies not implemented")
		}

		if policies["incident_response"] == nil {
			t.Error("Incident response policies not implemented")
		}
	})

	// Test 2: Physical safeguards
	t.Run("PhysicalSafeguards", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test physical safeguards
		req, _ := http.NewRequest("GET", "/api/security/physical", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Physical safeguards test failed with status %d", w.Code)
		}

		var physical map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &physical)
		if err != nil {
			t.Fatalf("Failed to unmarshal physical response: %v", err)
		}

		// Check for physical security
		if physical["facility_access"] == nil {
			t.Error("Facility access controls not implemented")
		}

		if physical["workstation_security"] == nil {
			t.Error("Workstation security not implemented")
		}

		if physical["device_controls"] == nil {
			t.Error("Device controls not implemented")
		}
	})

	// Test 3: Technical safeguards
	t.Run("TechnicalSafeguards", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test technical safeguards
		req, _ := http.NewRequest("GET", "/api/security/technical", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Technical safeguards test failed with status %d", w.Code)
		}

		var technical map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &technical)
		if err != nil {
			t.Fatalf("Failed to unmarshal technical response: %v", err)
		}

		// Check for technical security
		if technical["access_control"] == nil {
			t.Error("Access control not implemented")
		}

		if technical["audit_controls"] == nil {
			t.Error("Audit controls not implemented")
		}

		if technical["integrity"] == nil {
			t.Error("Data integrity controls not implemented")
		}

		if technical["transmission_security"] == nil {
			t.Error("Transmission security not implemented")
		}
	})
}

// TestComplianceSOX tests SOX compliance
func TestComplianceSOX(t *testing.T) {
	// Test 1: Internal controls
	t.Run("InternalControls", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test internal controls
		req, _ := http.NewRequest("GET", "/api/controls/internal", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Internal controls test failed with status %d", w.Code)
		}

		var controls map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &controls)
		if err != nil {
			t.Fatalf("Failed to unmarshal controls response: %v", err)
		}

		// Check for internal controls
		if controls["financial_reporting"] == nil {
			t.Error("Financial reporting controls not implemented")
		}

		if controls["data_integrity"] == nil {
			t.Error("Data integrity controls not implemented")
		}

		if controls["access_management"] == nil {
			t.Error("Access management controls not implemented")
		}
	})

	// Test 2: Audit trail
	t.Run("AuditTrail", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test audit trail
		req, _ := http.NewRequest("GET", "/api/audit/trail", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Audit trail test failed with status %d", w.Code)
		}

		var auditTrail []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &auditTrail)
		if err != nil {
			t.Fatalf("Failed to unmarshal audit trail response: %v", err)
		}

		// Check for audit trail
		if len(auditTrail) == 0 {
			t.Error("Audit trail is empty")
		}

		for _, entry := range auditTrail {
			if entry["timestamp"] == nil {
				t.Error("Audit entry missing timestamp")
			}

			if entry["user"] == nil {
				t.Error("Audit entry missing user")
			}

			if entry["action"] == nil {
				t.Error("Audit entry missing action")
			}

			if entry["resource"] == nil {
				t.Error("Audit entry missing resource")
			}
		}
	})

	// Test 3: Data retention
	t.Run("DataRetention", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test data retention
		req, _ := http.NewRequest("GET", "/api/data/retention", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Data retention test failed with status %d", w.Code)
		}

		var retention map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &retention)
		if err != nil {
			t.Fatalf("Failed to unmarshal retention response: %v", err)
		}

		// Check for data retention policies
		if retention["policies"] == nil {
			t.Error("Data retention policies not implemented")
		}

		if retention["automated_deletion"] == nil {
			t.Error("Automated deletion not implemented")
		}

		if retention["backup_retention"] == nil {
			t.Error("Backup retention not implemented")
		}
	})
}

// TestCompliancePCI tests PCI DSS compliance
func TestCompliancePCI(t *testing.T) {
	// Test 1: Network security
	t.Run("NetworkSecurity", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test network security
		req, _ := http.NewRequest("GET", "/api/security/network", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Network security test failed with status %d", w.Code)
		}

		var network map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &network)
		if err != nil {
			t.Fatalf("Failed to unmarshal network response: %v", err)
		}

		// Check for network security
		if network["firewall"] == nil {
			t.Error("Firewall not implemented")
		}

		if network["intrusion_detection"] == nil {
			t.Error("Intrusion detection not implemented")
		}

		if network["network_segmentation"] == nil {
			t.Error("Network segmentation not implemented")
		}
	})

	// Test 2: Data protection
	t.Run("DataProtection", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test data protection
		req, _ := http.NewRequest("GET", "/api/security/data-protection", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Data protection test failed with status %d", w.Code)
		}

		var protection map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &protection)
		if err != nil {
			t.Fatalf("Failed to unmarshal protection response: %v", err)
		}

		// Check for data protection
		if protection["encryption"] == nil {
			t.Error("Data encryption not implemented")
		}

		if protection["key_management"] == nil {
			t.Error("Key management not implemented")
		}

		if protection["data_masking"] == nil {
			t.Error("Data masking not implemented")
		}
	})

	// Test 3: Vulnerability management
	t.Run("VulnerabilityManagement", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test vulnerability management
		req, _ := http.NewRequest("GET", "/api/security/vulnerabilities", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Vulnerability management test failed with status %d", w.Code)
		}

		var vulnerabilities []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &vulnerabilities)
		if err != nil {
			t.Fatalf("Failed to unmarshal vulnerabilities response: %v", err)
		}

		// Check for vulnerability management
		for _, vuln := range vulnerabilities {
			if vuln["severity"] == nil {
				t.Error("Vulnerability missing severity")
			}

			if vuln["status"] == nil {
				t.Error("Vulnerability missing status")
			}

			if vuln["remediation"] == nil {
				t.Error("Vulnerability missing remediation")
			}
		}
	})
}

// TestComplianceISO27001 tests ISO 27001 compliance
func TestComplianceISO27001(t *testing.T) {
	// Test 1: Information security management
	t.Run("InformationSecurityManagement", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test information security management
		req, _ := http.NewRequest("GET", "/api/security/ism", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Information security management test failed with status %d", w.Code)
		}

		var ism map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &ism)
		if err != nil {
			t.Fatalf("Failed to unmarshal ISM response: %v", err)
		}

		// Check for information security management
		if ism["policies"] == nil {
			t.Error("Security policies not implemented")
		}

		if ism["procedures"] == nil {
			t.Error("Security procedures not implemented")
		}

		if ism["risk_assessment"] == nil {
			t.Error("Risk assessment not implemented")
		}
	})

	// Test 2: Risk management
	t.Run("RiskManagement", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test risk management
		req, _ := http.NewRequest("GET", "/api/security/risks", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Risk management test failed with status %d", w.Code)
		}

		var risks []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &risks)
		if err != nil {
			t.Fatalf("Failed to unmarshal risks response: %v", err)
		}

		// Check for risk management
		for _, risk := range risks {
			if risk["level"] == nil {
				t.Error("Risk missing level")
			}

			if risk["impact"] == nil {
				t.Error("Risk missing impact")
			}

			if risk["likelihood"] == nil {
				t.Error("Risk missing likelihood")
			}

			if risk["mitigation"] == nil {
				t.Error("Risk missing mitigation")
			}
		}
	})

	// Test 3: Continuous improvement
	t.Run("ContinuousImprovement", func(t *testing.T) {
		router := setupComplianceTestRouter(t)
		token := getComplianceAuthToken(t, router)

		// Test continuous improvement
		req, _ := http.NewRequest("GET", "/api/security/improvements", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Continuous improvement test failed with status %d", w.Code)
		}

		var improvements []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &improvements)
		if err != nil {
			t.Fatalf("Failed to unmarshal improvements response: %v", err)
		}

		// Check for continuous improvement
		for _, improvement := range improvements {
			if improvement["type"] == nil {
				t.Error("Improvement missing type")
			}

			if improvement["priority"] == nil {
				t.Error("Improvement missing priority")
			}

			if improvement["status"] == nil {
				t.Error("Improvement missing status")
			}

			if improvement["target_date"] == nil {
				t.Error("Improvement missing target date")
			}
		}
	})
}

// Helper functions

func setupComplianceTestRouter(t *testing.T) *gin.Engine {
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

	db := setupComplianceTestDB(t)

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

func getComplianceAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupComplianceTestDB(t *testing.T) *sql.DB {
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
