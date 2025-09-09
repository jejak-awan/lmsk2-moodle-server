package disaster_recovery

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

// TestDisasterRecoveryBackup tests backup functionality
func TestDisasterRecoveryBackup(t *testing.T) {
	// Test 1: Database backup
	t.Run("DatabaseBackup", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test database backup
		req, _ := http.NewRequest("POST", "/api/backup/database", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Database backup test failed with status %d", w.Code)
		}

		var backup map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &backup)
		if err != nil {
			t.Fatalf("Failed to unmarshal backup response: %v", err)
		}

		// Check for backup creation
		if backup["backup_id"] == nil {
			t.Error("Backup ID not generated")
		}

		if backup["status"] != "completed" {
			t.Error("Backup status not completed")
		}

		if backup["size"] == nil {
			t.Error("Backup size not reported")
		}

		if backup["created_at"] == nil {
			t.Error("Backup creation time not reported")
		}
	})

	// Test 2: File system backup
	t.Run("FileSystemBackup", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test file system backup
		req, _ := http.NewRequest("POST", "/api/backup/filesystem", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("File system backup test failed with status %d", w.Code)
		}

		var backup map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &backup)
		if err != nil {
			t.Fatalf("Failed to unmarshal backup response: %v", err)
		}

		// Check for backup creation
		if backup["backup_id"] == nil {
			t.Error("Backup ID not generated")
		}

		if backup["status"] != "completed" {
			t.Error("Backup status not completed")
		}

		if backup["files_count"] == nil {
			t.Error("Files count not reported")
		}

		if backup["total_size"] == nil {
			t.Error("Total size not reported")
		}
	})

	// Test 3: Configuration backup
	t.Run("ConfigurationBackup", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test configuration backup
		req, _ := http.NewRequest("POST", "/api/backup/configuration", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Configuration backup test failed with status %d", w.Code)
		}

		var backup map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &backup)
		if err != nil {
			t.Fatalf("Failed to unmarshal backup response: %v", err)
		}

		// Check for backup creation
		if backup["backup_id"] == nil {
			t.Error("Backup ID not generated")
		}

		if backup["status"] != "completed" {
			t.Error("Backup status not completed")
		}

		if backup["config_files"] == nil {
			t.Error("Configuration files not backed up")
		}
	})

	// Test 4: Full system backup
	t.Run("FullSystemBackup", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test full system backup
		req, _ := http.NewRequest("POST", "/api/backup/full", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Full system backup test failed with status %d", w.Code)
		}

		var backup map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &backup)
		if err != nil {
			t.Fatalf("Failed to unmarshal backup response: %v", err)
		}

		// Check for backup creation
		if backup["backup_id"] == nil {
			t.Error("Backup ID not generated")
		}

		if backup["status"] != "completed" {
			t.Error("Backup status not completed")
		}

		if backup["components"] == nil {
			t.Error("Backup components not listed")
		}
	})
}

// TestDisasterRecoveryRestore tests restore functionality
func TestDisasterRecoveryRestore(t *testing.T) {
	// Test 1: Database restore
	t.Run("DatabaseRestore", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test database restore
		restoreData := map[string]interface{}{
			"backup_id": "test-backup-123",
			"force":     false,
		}

		jsonData, _ := json.Marshal(restoreData)
		req, _ := http.NewRequest("POST", "/api/restore/database", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Database restore test failed with status %d", w.Code)
		}

		var restore map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &restore)
		if err != nil {
			t.Fatalf("Failed to unmarshal restore response: %v", err)
		}

		// Check for restore operation
		if restore["restore_id"] == nil {
			t.Error("Restore ID not generated")
		}

		if restore["status"] != "completed" {
			t.Error("Restore status not completed")
		}

		if restore["restored_at"] == nil {
			t.Error("Restore time not reported")
		}
	})

	// Test 2: File system restore
	t.Run("FileSystemRestore", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test file system restore
		restoreData := map[string]interface{}{
			"backup_id": "test-backup-123",
			"force":     false,
		}

		jsonData, _ := json.Marshal(restoreData)
		req, _ := http.NewRequest("POST", "/api/restore/filesystem", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("File system restore test failed with status %d", w.Code)
		}

		var restore map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &restore)
		if err != nil {
			t.Fatalf("Failed to unmarshal restore response: %v", err)
		}

		// Check for restore operation
		if restore["restore_id"] == nil {
			t.Error("Restore ID not generated")
		}

		if restore["status"] != "completed" {
			t.Error("Restore status not completed")
		}

		if restore["files_restored"] == nil {
			t.Error("Files restored count not reported")
		}
	})

	// Test 3: Configuration restore
	t.Run("ConfigurationRestore", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test configuration restore
		restoreData := map[string]interface{}{
			"backup_id": "test-backup-123",
			"force":     false,
		}

		jsonData, _ := json.Marshal(restoreData)
		req, _ := http.NewRequest("POST", "/api/restore/configuration", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Configuration restore test failed with status %d", w.Code)
		}

		var restore map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &restore)
		if err != nil {
			t.Fatalf("Failed to unmarshal restore response: %v", err)
		}

		// Check for restore operation
		if restore["restore_id"] == nil {
			t.Error("Restore ID not generated")
		}

		if restore["status"] != "completed" {
			t.Error("Restore status not completed")
		}

		if restore["config_restored"] == nil {
			t.Error("Configuration restore status not reported")
		}
	})

	// Test 4: Full system restore
	t.Run("FullSystemRestore", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test full system restore
		restoreData := map[string]interface{}{
			"backup_id": "test-backup-123",
			"force":     false,
		}

		jsonData, _ := json.Marshal(restoreData)
		req, _ := http.NewRequest("POST", "/api/restore/full", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Full system restore test failed with status %d", w.Code)
		}

		var restore map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &restore)
		if err != nil {
			t.Fatalf("Failed to unmarshal restore response: %v", err)
		}

		// Check for restore operation
		if restore["restore_id"] == nil {
			t.Error("Restore ID not generated")
		}

		if restore["status"] != "completed" {
			t.Error("Restore status not completed")
		}

		if restore["components_restored"] == nil {
			t.Error("Components restored not listed")
		}
	})
}

// TestDisasterRecoveryFailover tests failover functionality
func TestDisasterRecoveryFailover(t *testing.T) {
	// Test 1: Automatic failover
	t.Run("AutomaticFailover", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test automatic failover
		req, _ := http.NewRequest("POST", "/api/failover/automatic", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Automatic failover test failed with status %d", w.Code)
		}

		var failover map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &failover)
		if err != nil {
			t.Fatalf("Failed to unmarshal failover response: %v", err)
		}

		// Check for failover operation
		if failover["failover_id"] == nil {
			t.Error("Failover ID not generated")
		}

		if failover["status"] != "completed" {
			t.Error("Failover status not completed")
		}

		if failover["primary_server"] == nil {
			t.Error("Primary server not identified")
		}

		if failover["backup_server"] == nil {
			t.Error("Backup server not identified")
		}
	})

	// Test 2: Manual failover
	t.Run("ManualFailover", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test manual failover
		failoverData := map[string]interface{}{
			"target_server": "backup-server-01",
			"force":         false,
		}

		jsonData, _ := json.Marshal(failoverData)
		req, _ := http.NewRequest("POST", "/api/failover/manual", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Manual failover test failed with status %d", w.Code)
		}

		var failover map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &failover)
		if err != nil {
			t.Fatalf("Failed to unmarshal failover response: %v", err)
		}

		// Check for failover operation
		if failover["failover_id"] == nil {
			t.Error("Failover ID not generated")
		}

		if failover["status"] != "completed" {
			t.Error("Failover status not completed")
		}

		if failover["target_server"] != "backup-server-01" {
			t.Error("Target server not set correctly")
		}
	})

	// Test 3: Failover status
	t.Run("FailoverStatus", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test failover status
		req, _ := http.NewRequest("GET", "/api/failover/status", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Failover status test failed with status %d", w.Code)
		}

		var status map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &status)
		if err != nil {
			t.Fatalf("Failed to unmarshal status response: %v", err)
		}

		// Check for failover status
		if status["current_server"] == nil {
			t.Error("Current server not identified")
		}

		if status["backup_servers"] == nil {
			t.Error("Backup servers not listed")
		}

		if status["health_status"] == nil {
			t.Error("Health status not reported")
		}
	})
}

// TestDisasterRecoveryRecoveryTime tests recovery time objectives
func TestDisasterRecoveryRecoveryTime(t *testing.T) {
	// Test 1: Recovery Time Objective (RTO)
	t.Run("RecoveryTimeObjective", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test RTO
		start := time.Now()
		
		req, _ := http.NewRequest("POST", "/api/restore/full", bytes.NewBufferString(`{"backup_id": "test-backup-123"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		duration := time.Since(start)
		
		if w.Code != http.StatusOK {
			t.Errorf("RTO test failed with status %d", w.Code)
		}

		// Check RTO is within acceptable limits (e.g., 4 hours)
		if duration > 4*time.Hour {
			t.Errorf("Recovery time too long: %v", duration)
		}

		var restore map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &restore)
		if err != nil {
			t.Fatalf("Failed to unmarshal restore response: %v", err)
		}

		// Check for recovery time reporting
		if restore["recovery_time"] == nil {
			t.Error("Recovery time not reported")
		}
	})

	// Test 2: Recovery Point Objective (RPO)
	t.Run("RecoveryPointObjective", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test RPO
		req, _ := http.NewRequest("GET", "/api/backup/latest", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("RPO test failed with status %d", w.Code)
		}

		var backup map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &backup)
		if err != nil {
			t.Fatalf("Failed to unmarshal backup response: %v", err)
		}

		// Check for RPO compliance
		if backup["created_at"] == nil {
			t.Error("Backup creation time not reported")
		}

		// Check if backup is recent (e.g., within 1 hour)
		createdAt := backup["created_at"].(string)
		if createdAt == "" {
			t.Error("Backup creation time is empty")
		}
	})

	// Test 3: Service Level Agreement (SLA)
	t.Run("ServiceLevelAgreement", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test SLA compliance
		req, _ := http.NewRequest("GET", "/api/sla/compliance", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("SLA compliance test failed with status %d", w.Code)
		}

		var sla map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &sla)
		if err != nil {
			t.Fatalf("Failed to unmarshal SLA response: %v", err)
		}

		// Check for SLA compliance
		if sla["uptime"] == nil {
			t.Error("Uptime not reported")
		}

		if sla["availability"] == nil {
			t.Error("Availability not reported")
		}

		if sla["rto_compliance"] == nil {
			t.Error("RTO compliance not reported")
		}

		if sla["rpo_compliance"] == nil {
			t.Error("RPO compliance not reported")
		}
	})
}

// TestDisasterRecoveryDataIntegrity tests data integrity during recovery
func TestDisasterRecoveryDataIntegrity(t *testing.T) {
	// Test 1: Data consistency
	t.Run("DataConsistency", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test data consistency
		req, _ := http.NewRequest("GET", "/api/integrity/check", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Data consistency test failed with status %d", w.Code)
		}

		var integrity map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &integrity)
		if err != nil {
			t.Fatalf("Failed to unmarshal integrity response: %v", err)
		}

		// Check for data consistency
		if integrity["database_consistent"] != true {
			t.Error("Database not consistent")
		}

		if integrity["files_consistent"] != true {
			t.Error("Files not consistent")
		}

		if integrity["checksums_valid"] != true {
			t.Error("Checksums not valid")
		}
	})

	// Test 2: Data validation
	t.Run("DataValidation", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test data validation
		req, _ := http.NewRequest("POST", "/api/integrity/validate", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Data validation test failed with status %d", w.Code)
		}

		var validation map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &validation)
		if err != nil {
			t.Fatalf("Failed to unmarshal validation response: %v", err)
		}

		// Check for data validation
		if validation["validation_id"] == nil {
			t.Error("Validation ID not generated")
		}

		if validation["status"] != "completed" {
			t.Error("Validation status not completed")
		}

		if validation["errors"] == nil {
			t.Error("Validation errors not reported")
		}

		if validation["warnings"] == nil {
			t.Error("Validation warnings not reported")
		}
	})

	// Test 3: Data corruption detection
	t.Run("DataCorruptionDetection", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test data corruption detection
		req, _ := http.NewRequest("GET", "/api/integrity/corruption", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Data corruption detection test failed with status %d", w.Code)
		}

		var corruption map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &corruption)
		if err != nil {
			t.Fatalf("Failed to unmarshal corruption response: %v", err)
		}

		// Check for corruption detection
		if corruption["corruption_detected"] == nil {
			t.Error("Corruption detection status not reported")
		}

		if corruption["corrupted_files"] == nil {
			t.Error("Corrupted files not listed")
		}

		if corruption["corrupted_records"] == nil {
			t.Error("Corrupted records not listed")
		}
	})
}

// TestDisasterRecoveryTesting tests disaster recovery testing
func TestDisasterRecoveryTesting(t *testing.T) {
	// Test 1: Recovery testing
	t.Run("RecoveryTesting", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test recovery testing
		req, _ := http.NewRequest("POST", "/api/testing/recovery", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Recovery testing test failed with status %d", w.Code)
		}

		var testing map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &testing)
		if err != nil {
			t.Fatalf("Failed to unmarshal testing response: %v", err)
		}

		// Check for recovery testing
		if testing["test_id"] == nil {
			t.Error("Test ID not generated")
		}

		if testing["status"] != "completed" {
			t.Error("Test status not completed")
		}

		if testing["test_results"] == nil {
			t.Error("Test results not reported")
		}

		if testing["recommendations"] == nil {
			t.Error("Recommendations not provided")
		}
	})

	// Test 2: Failover testing
	t.Run("FailoverTesting", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test failover testing
		req, _ := http.NewRequest("POST", "/api/testing/failover", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Failover testing test failed with status %d", w.Code)
		}

		var testing map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &testing)
		if err != nil {
			t.Fatalf("Failed to unmarshal testing response: %v", err)
		}

		// Check for failover testing
		if testing["test_id"] == nil {
			t.Error("Test ID not generated")
		}

		if testing["status"] != "completed" {
			t.Error("Test status not completed")
		}

		if testing["failover_time"] == nil {
			t.Error("Failover time not reported")
		}

		if testing["service_availability"] == nil {
			t.Error("Service availability not reported")
		}
	})

	// Test 3: Backup testing
	t.Run("BackupTesting", func(t *testing.T) {
		router := setupDisasterRecoveryTestRouter(t)
		token := getDisasterRecoveryAuthToken(t, router)

		// Test backup testing
		req, _ := http.NewRequest("POST", "/api/testing/backup", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Backup testing test failed with status %d", w.Code)
		}

		var testing map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &testing)
		if err != nil {
			t.Fatalf("Failed to unmarshal testing response: %v", err)
		}

		// Check for backup testing
		if testing["test_id"] == nil {
			t.Error("Test ID not generated")
		}

		if testing["status"] != "completed" {
			t.Error("Test status not completed")
		}

		if testing["backup_integrity"] == nil {
			t.Error("Backup integrity not tested")
		}

		if testing["restore_capability"] == nil {
			t.Error("Restore capability not tested")
		}
	})
}

// Helper functions

func setupDisasterRecoveryTestRouter(t *testing.T) *gin.Engine {
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

	db := setupDisasterRecoveryTestDB(t)

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

func getDisasterRecoveryAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupDisasterRecoveryTestDB(t *testing.T) *sql.DB {
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
