package contract

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

// TestContractTestingAuthentication tests authentication contract
func TestContractTestingAuthentication(t *testing.T) {
	// Test 1: Login contract
	t.Run("LoginContract", func(t *testing.T) {
		router := setupContractTestRouter(t)

		// Test valid login contract
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
			t.Errorf("Login contract failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		// Verify login contract
		if response["token"] == nil {
			t.Error("Login contract missing token field")
		}

		if response["user"] == nil {
			t.Error("Login contract missing user field")
		}

		if response["expires_at"] == nil {
			t.Error("Login contract missing expires_at field")
		}

		// Verify token format
		token := response["token"].(string)
		if len(token) < 100 {
			t.Error("Login contract token too short")
		}

		// Verify user object structure
		user := response["user"].(map[string]interface{})
		if user["id"] == nil {
			t.Error("Login contract user missing id field")
		}

		if user["username"] == nil {
			t.Error("Login contract user missing username field")
		}

		if user["email"] == nil {
			t.Error("Login contract user missing email field")
		}

		if user["role"] == nil {
			t.Error("Login contract user missing role field")
		}
	})

	// Test 2: Logout contract
	t.Run("LogoutContract", func(t *testing.T) {
		router := setupContractTestRouter(t)
		token := getContractAuthToken(t, router)

		// Test logout contract
		req, _ := http.NewRequest("POST", "/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Logout contract failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal logout response: %v", err)
		}

		// Verify logout contract
		if response["message"] == nil {
			t.Error("Logout contract missing message field")
		}

		if response["logged_out"] == nil {
			t.Error("Logout contract missing logged_out field")
		}

		if response["logged_out"] != true {
			t.Error("Logout contract logged_out should be true")
		}
	})

	// Test 3: Token validation contract
	t.Run("TokenValidationContract", func(t *testing.T) {
		router := setupContractTestRouter(t)
		token := getContractAuthToken(t, router)

		// Test token validation contract
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Token validation contract failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Verify token validation contract
		if response["cpu_usage"] == nil {
			t.Error("Token validation contract missing cpu_usage field")
		}

		if response["memory_usage"] == nil {
			t.Error("Token validation contract missing memory_usage field")
		}

		if response["disk_usage"] == nil {
			t.Error("Token validation contract missing disk_usage field")
		}

		if response["uptime"] == nil {
			t.Error("Token validation contract missing uptime field")
		}
	})
}

// TestContractTestingUserManagement tests user management contract
func TestContractTestingUserManagement(t *testing.T) {
	// Test 1: User creation contract
	t.Run("UserCreationContract", func(t *testing.T) {
		router := setupContractTestRouter(t)
		token := getContractAuthToken(t, router)

		// Test user creation contract
		userData := map[string]interface{}{
			"username": "contractuser",
			"email":    "contract@example.com",
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
			t.Errorf("User creation contract failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal user creation response: %v", err)
		}

		// Verify user creation contract
		if response["id"] == nil {
			t.Error("User creation contract missing id field")
		}

		if response["username"] == nil {
			t.Error("User creation contract missing username field")
		}

		if response["email"] == nil {
			t.Error("User creation contract missing email field")
		}

		if response["role"] == nil {
			t.Error("User creation contract missing role field")
		}

		if response["active"] == nil {
			t.Error("User creation contract missing active field")
		}

		if response["created_at"] == nil {
			t.Error("User creation contract missing created_at field")
		}

		if response["updated_at"] == nil {
			t.Error("User creation contract missing updated_at field")
		}

		// Verify field values
		if response["username"] != "contractuser" {
			t.Errorf("User creation contract username mismatch: expected 'contractuser', got %v", response["username"])
		}

		if response["email"] != "contract@example.com" {
			t.Errorf("User creation contract email mismatch: expected 'contract@example.com', got %v", response["email"])
		}

		if response["role"] != "viewer" {
			t.Errorf("User creation contract role mismatch: expected 'viewer', got %v", response["role"])
		}

		if response["active"] != true {
			t.Error("User creation contract active should be true")
		}
	})

	// Test 2: User retrieval contract
	t.Run("UserRetrievalContract", func(t *testing.T) {
		router := setupContractTestRouter(t)
		token := getContractAuthToken(t, router)

		// Test user retrieval contract
		req, _ := http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("User retrieval contract failed with status %d", w.Code)
		}

		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal users response: %v", err)
		}

		// Verify user retrieval contract
		if len(response) < 1 {
			t.Error("User retrieval contract should return at least one user")
		}

		user := response[0]
		if user["id"] == nil {
			t.Error("User retrieval contract missing id field")
		}

		if user["username"] == nil {
			t.Error("User retrieval contract missing username field")
		}

		if user["email"] == nil {
			t.Error("User retrieval contract missing email field")
		}

		if user["role"] == nil {
			t.Error("User retrieval contract missing role field")
		}

		if user["active"] == nil {
			t.Error("User retrieval contract missing active field")
		}

		if user["created_at"] == nil {
			t.Error("User retrieval contract missing created_at field")
		}

		if user["updated_at"] == nil {
			t.Error("User retrieval contract missing updated_at field")
		}

		// Verify password is not exposed
		if user["password"] != nil {
			t.Error("User retrieval contract should not expose password field")
		}

		if user["password_hash"] != nil {
			t.Error("User retrieval contract should not expose password_hash field")
		}
	})

	// Test 3: User stats contract
	t.Run("UserStatsContract", func(t *testing.T) {
		router := setupContractTestRouter(t)
		token := getContractAuthToken(t, router)

		// Test user stats contract
		req, _ := http.NewRequest("GET", "/api/users/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("User stats contract failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal user stats response: %v", err)
		}

		// Verify user stats contract
		if response["total_users"] == nil {
			t.Error("User stats contract missing total_users field")
		}

		if response["active_users"] == nil {
			t.Error("User stats contract missing active_users field")
		}

		if response["inactive_users"] == nil {
			t.Error("User stats contract missing inactive_users field")
		}

		if response["users_by_role"] == nil {
			t.Error("User stats contract missing users_by_role field")
		}

		if response["new_users_today"] == nil {
			t.Error("User stats contract missing new_users_today field")
		}

		if response["new_users_this_week"] == nil {
			t.Error("User stats contract missing new_users_this_week field")
		}

		if response["new_users_this_month"] == nil {
			t.Error("User stats contract missing new_users_this_month field")
		}

		// Verify field types
		if _, ok := response["total_users"].(float64); !ok {
			t.Error("User stats contract total_users should be a number")
		}

		if _, ok := response["active_users"].(float64); !ok {
			t.Error("User stats contract active_users should be a number")
		}

		if _, ok := response["inactive_users"].(float64); !ok {
			t.Error("User stats contract inactive_users should be a number")
		}

		if _, ok := response["users_by_role"].(map[string]interface{}); !ok {
			t.Error("User stats contract users_by_role should be an object")
		}
	})
}

// TestContractTestingSystemMonitoring tests system monitoring contract
func TestContractTestingSystemMonitoring(t *testing.T) {
	// Test 1: System stats contract
	t.Run("SystemStatsContract", func(t *testing.T) {
		router := setupContractTestRouter(t)
		token := getContractAuthToken(t, router)

		// Test system stats contract
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("System stats contract failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal system stats response: %v", err)
		}

		// Verify system stats contract
		if response["cpu_usage"] == nil {
			t.Error("System stats contract missing cpu_usage field")
		}

		if response["memory_usage"] == nil {
			t.Error("System stats contract missing memory_usage field")
		}

		if response["disk_usage"] == nil {
			t.Error("System stats contract missing disk_usage field")
		}

		if response["uptime"] == nil {
			t.Error("System stats contract missing uptime field")
		}

		if response["load_average"] == nil {
			t.Error("System stats contract missing load_average field")
		}

		if response["network_io"] == nil {
			t.Error("System stats contract missing network_io field")
		}

		if response["disk_io"] == nil {
			t.Error("System stats contract missing disk_io field")
		}

		if response["timestamp"] == nil {
			t.Error("System stats contract missing timestamp field")
		}

		// Verify field types
		if _, ok := response["cpu_usage"].(float64); !ok {
			t.Error("System stats contract cpu_usage should be a number")
		}

		if _, ok := response["memory_usage"].(float64); !ok {
			t.Error("System stats contract memory_usage should be a number")
		}

		if _, ok := response["disk_usage"].(float64); !ok {
			t.Error("System stats contract disk_usage should be a number")
		}

		if _, ok := response["uptime"].(float64); !ok {
			t.Error("System stats contract uptime should be a number")
		}

		if _, ok := response["load_average"].([]interface{}); !ok {
			t.Error("System stats contract load_average should be an array")
		}

		if _, ok := response["network_io"].(map[string]interface{}); !ok {
			t.Error("System stats contract network_io should be an object")
		}

		if _, ok := response["disk_io"].(map[string]interface{}); !ok {
			t.Error("System stats contract disk_io should be an object")
		}

		// Verify value ranges
		if cpuUsage, ok := response["cpu_usage"].(float64); ok {
			if cpuUsage < 0 || cpuUsage > 100 {
				t.Errorf("System stats contract cpu_usage out of range: %f", cpuUsage)
			}
		}

		if memoryUsage, ok := response["memory_usage"].(float64); ok {
			if memoryUsage < 0 || memoryUsage > 100 {
				t.Errorf("System stats contract memory_usage out of range: %f", memoryUsage)
			}
		}

		if diskUsage, ok := response["disk_usage"].(float64); ok {
			if diskUsage < 0 || diskUsage > 100 {
				t.Errorf("System stats contract disk_usage out of range: %f", diskUsage)
			}
		}
	})

	// Test 2: Alerts contract
	t.Run("AlertsContract", func(t *testing.T) {
		router := setupContractTestRouter(t)
		token := getContractAuthToken(t, router)

		// Test alerts contract
		req, _ := http.NewRequest("GET", "/api/alerts", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Alerts contract failed with status %d", w.Code)
		}

		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal alerts response: %v", err)
		}

		// Verify alerts contract structure
		if response == nil {
			t.Error("Alerts contract should return an array")
		}

		// If there are alerts, verify their structure
		if len(response) > 0 {
			alert := response[0]
			if alert["id"] == nil {
				t.Error("Alerts contract missing id field")
			}

			if alert["type"] == nil {
				t.Error("Alerts contract missing type field")
			}

			if alert["message"] == nil {
				t.Error("Alerts contract missing message field")
			}

			if alert["severity"] == nil {
				t.Error("Alerts contract missing severity field")
			}

			if alert["resolved"] == nil {
				t.Error("Alerts contract missing resolved field")
			}

			if alert["created_at"] == nil {
				t.Error("Alerts contract missing created_at field")
			}

			if alert["resolved_at"] == nil {
				t.Error("Alerts contract missing resolved_at field")
			}

			// Verify field types
			if _, ok := alert["resolved"].(bool); !ok {
				t.Error("Alerts contract resolved should be a boolean")
			}

			// Verify severity values
			severity := alert["severity"].(string)
			validSeverities := []string{"low", "medium", "high", "critical"}
			valid := false
			for _, s := range validSeverities {
				if severity == s {
					valid = true
					break
				}
			}
			if !valid {
				t.Errorf("Alerts contract invalid severity: %s", severity)
			}
		}
	})

	// Test 3: System logs contract
	t.Run("SystemLogsContract", func(t *testing.T) {
		router := setupContractTestRouter(t)
		token := getContractAuthToken(t, router)

		// Test system logs contract
		req, _ := http.NewRequest("GET", "/api/logs?limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("System logs contract failed with status %d", w.Code)
		}

		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal system logs response: %v", err)
		}

		// Verify system logs contract structure
		if response == nil {
			t.Error("System logs contract should return an array")
		}

		// If there are logs, verify their structure
		if len(response) > 0 {
			log := response[0]
			if log["id"] == nil {
				t.Error("System logs contract missing id field")
			}

			if log["level"] == nil {
				t.Error("System logs contract missing level field")
			}

			if log["message"] == nil {
				t.Error("System logs contract missing message field")
			}

			if log["source"] == nil {
				t.Error("System logs contract missing source field")
			}

			if log["created_at"] == nil {
				t.Error("System logs contract missing created_at field")
			}

			// Verify field types
			if _, ok := log["message"].(string); !ok {
				t.Error("System logs contract message should be a string")
			}

			// Verify log levels
			level := log["level"].(string)
			validLevels := []string{"debug", "info", "warn", "error", "fatal"}
			valid := false
			for _, l := range validLevels {
				if level == l {
					valid = true
					break
				}
			}
			if !valid {
				t.Errorf("System logs contract invalid level: %s", level)
			}
		}
	})
}

// TestContractTestingMoodleManagement tests Moodle management contract
func TestContractTestingMoodleManagement(t *testing.T) {
	// Test 1: Moodle status contract
	t.Run("MoodleStatusContract", func(t *testing.T) {
		router := setupContractTestRouter(t)
		token := getContractAuthToken(t, router)

		// Test Moodle status contract
		req, _ := http.NewRequest("GET", "/api/moodle/status", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Moodle status contract failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal Moodle status response: %v", err)
		}

		// Verify Moodle status contract
		if response["running"] == nil {
			t.Error("Moodle status contract missing running field")
		}

		if response["version"] == nil {
			t.Error("Moodle status contract missing version field")
		}

		if response["uptime"] == nil {
			t.Error("Moodle status contract missing uptime field")
		}

		if response["last_check"] == nil {
			t.Error("Moodle status contract missing last_check field")
		}

		if response["database_status"] == nil {
			t.Error("Moodle status contract missing database_status field")
		}

		if response["cache_status"] == nil {
			t.Error("Moodle status contract missing cache_status field")
		}

		if response["plugins_status"] == nil {
			t.Error("Moodle status contract missing plugins_status field")
		}

		if response["maintenance_mode"] == nil {
			t.Error("Moodle status contract missing maintenance_mode field")
		}

		// Verify field types
		if _, ok := response["running"].(bool); !ok {
			t.Error("Moodle status contract running should be a boolean")
		}

		if _, ok := response["uptime"].(float64); !ok {
			t.Error("Moodle status contract uptime should be a number")
		}

		if _, ok := response["maintenance_mode"].(bool); !ok {
			t.Error("Moodle status contract maintenance_mode should be a boolean")
		}

		// Verify version format
		version := response["version"].(string)
		if len(version) < 3 {
			t.Error("Moodle status contract version should be at least 3 characters")
		}
	})

	// Test 2: Moodle operations contract
	t.Run("MoodleOperationsContract", func(t *testing.T) {
		router := setupContractTestRouter(t)
		token := getContractAuthToken(t, router)

		// Test Moodle start contract
		req, _ := http.NewRequest("POST", "/api/moodle/start", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("Moodle start contract failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal Moodle start response: %v", err)
		}

		// Verify Moodle start contract
		if response["status"] == nil {
			t.Error("Moodle start contract missing status field")
		}

		if response["message"] == nil {
			t.Error("Moodle start contract missing message field")
		}

		if response["timestamp"] == nil {
			t.Error("Moodle start contract missing timestamp field")
		}

		// Test Moodle stop contract
		req, _ = http.NewRequest("POST", "/api/moodle/stop", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("Moodle stop contract failed with status %d", w.Code)
		}

		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal Moodle stop response: %v", err)
		}

		// Verify Moodle stop contract
		if response["status"] == nil {
			t.Error("Moodle stop contract missing status field")
		}

		if response["message"] == nil {
			t.Error("Moodle stop contract missing message field")
		}

		if response["timestamp"] == nil {
			t.Error("Moodle stop contract missing timestamp field")
		}

		// Test Moodle restart contract
		req, _ = http.NewRequest("POST", "/api/moodle/restart", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("Moodle restart contract failed with status %d", w.Code)
		}

		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal Moodle restart response: %v", err)
		}

		// Verify Moodle restart contract
		if response["status"] == nil {
			t.Error("Moodle restart contract missing status field")
		}

		if response["message"] == nil {
			t.Error("Moodle restart contract missing message field")
		}

		if response["timestamp"] == nil {
			t.Error("Moodle restart contract missing timestamp field")
		}
	})
}

// TestContractTestingErrorHandling tests error handling contract
func TestContractTestingErrorHandling(t *testing.T) {
	// Test 1: HTTP error contract
	t.Run("HTTPErrorContract", func(t *testing.T) {
		router := setupContractTestRouter(t)

		// Test 404 error contract
		req, _ := http.NewRequest("GET", "/api/nonexistent", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("404 error contract failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal 404 error response: %v", err)
		}

		// Verify 404 error contract
		if response["error"] == nil {
			t.Error("404 error contract missing error field")
		}

		if response["code"] == nil {
			t.Error("404 error contract missing code field")
		}

		if response["message"] == nil {
			t.Error("404 error contract missing message field")
		}

		if response["timestamp"] == nil {
			t.Error("404 error contract missing timestamp field")
		}

		// Verify field values
		if response["code"] != float64(404) {
			t.Errorf("404 error contract code mismatch: expected 404, got %v", response["code"])
		}

		// Test 405 error contract
		req, _ = http.NewRequest("DELETE", "/health", nil)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("405 error contract failed with status %d", w.Code)
		}

		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal 405 error response: %v", err)
		}

		// Verify 405 error contract
		if response["error"] == nil {
			t.Error("405 error contract missing error field")
		}

		if response["code"] == nil {
			t.Error("405 error contract missing code field")
		}

		if response["message"] == nil {
			t.Error("405 error contract missing message field")
		}

		if response["timestamp"] == nil {
			t.Error("405 error contract missing timestamp field")
		}

		// Verify field values
		if response["code"] != float64(405) {
			t.Errorf("405 error contract code mismatch: expected 405, got %v", response["code"])
		}
	})

	// Test 2: Validation error contract
	t.Run("ValidationErrorContract", func(t *testing.T) {
		router := setupContractTestRouter(t)
		token := getContractAuthToken(t, router)

		// Test validation error contract
		invalidUserData := map[string]interface{}{
			"username": "ab", // Too short
			"email":    "invalid-email",
			"password": "weak",
			"role":     "invalid-role",
		}

		jsonData, _ := json.Marshal(invalidUserData)
		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Validation error contract failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal validation error response: %v", err)
		}

		// Verify validation error contract
		if response["error"] == nil {
			t.Error("Validation error contract missing error field")
		}

		if response["code"] == nil {
			t.Error("Validation error contract missing code field")
		}

		if response["message"] == nil {
			t.Error("Validation error contract missing message field")
		}

		if response["details"] == nil {
			t.Error("Validation error contract missing details field")
		}

		if response["timestamp"] == nil {
			t.Error("Validation error contract missing timestamp field")
		}

		// Verify field values
		if response["code"] != float64(400) {
			t.Errorf("Validation error contract code mismatch: expected 400, got %v", response["code"])
		}

		// Verify details structure
		details := response["details"].(map[string]interface{})
		if details["username"] == nil {
			t.Error("Validation error contract details missing username field")
		}

		if details["email"] == nil {
			t.Error("Validation error contract details missing email field")
		}

		if details["password"] == nil {
			t.Error("Validation error contract details missing password field")
		}

		if details["role"] == nil {
			t.Error("Validation error contract details missing role field")
		}
	})

	// Test 3: Authentication error contract
	t.Run("AuthenticationErrorContract", func(t *testing.T) {
		router := setupContractTestRouter(t)

		// Test authentication error contract
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Authentication error contract failed with status %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal authentication error response: %v", err)
		}

		// Verify authentication error contract
		if response["error"] == nil {
			t.Error("Authentication error contract missing error field")
		}

		if response["code"] == nil {
			t.Error("Authentication error contract missing code field")
		}

		if response["message"] == nil {
			t.Error("Authentication error contract missing message field")
		}

		if response["timestamp"] == nil {
			t.Error("Authentication error contract missing timestamp field")
		}

		// Verify field values
		if response["code"] != float64(401) {
			t.Errorf("Authentication error contract code mismatch: expected 401, got %v", response["code"])
		}
	})
}

// Helper functions

func setupContractTestRouter(t *testing.T) *gin.Engine {
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

	db := setupContractTestDB(t)

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

func getContractAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupContractTestDB(t *testing.T) *sql.DB {
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
