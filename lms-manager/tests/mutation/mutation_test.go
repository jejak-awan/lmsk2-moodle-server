package mutation

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

// TestMutationTestingAuthentication tests authentication mutation
func TestMutationTestingAuthentication(t *testing.T) {
	// Test 1: Login mutation
	t.Run("LoginMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)

		// Test valid login
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
			t.Errorf("Valid login should return 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		if response["token"] == nil {
			t.Error("Valid login should return token")
		}

		// Test invalid login
		invalidLoginData := map[string]string{
			"username": "admin",
			"password": "wrongpassword",
		}

		jsonData, _ = json.Marshal(invalidLoginData)
		req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Invalid login should return 401, got %d", w.Code)
		}

		// Test empty credentials
		emptyLoginData := map[string]string{
			"username": "",
			"password": "",
		}

		jsonData, _ = json.Marshal(emptyLoginData)
		req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Empty credentials should return 400, got %d", w.Code)
		}
	})

	// Test 2: Token validation mutation
	t.Run("TokenValidationMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)
		token := getMutationAuthToken(t, router)

		// Test valid token
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Valid token should allow access, got %d", w.Code)
		}

		// Test invalid token
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Invalid token should return 401, got %d", w.Code)
		}

		// Test missing token
		req, _ = http.NewRequest("GET", "/api/stats", nil)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Missing token should return 401, got %d", w.Code)
		}

		// Test malformed token
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer ")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Malformed token should return 401, got %d", w.Code)
		}
	})

	// Test 3: Session management mutation
	t.Run("SessionManagementMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)
		token := getMutationAuthToken(t, router)

		// Test session creation
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Session creation should succeed, got %d", w.Code)
		}

		// Test session validation
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Session validation should succeed, got %d", w.Code)
		}

		// Test session expiration
		req, _ = http.NewRequest("POST", "/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Session expiration should succeed, got %d", w.Code)
		}

		// Test expired session
		req, _ = http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expired session should return 401, got %d", w.Code)
		}
	})
}

// TestMutationTestingUserManagement tests user management mutation
func TestMutationTestingUserManagement(t *testing.T) {
	// Test 1: User creation mutation
	t.Run("UserCreationMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)
		token := getMutationAuthToken(t, router)

		// Test valid user creation
		userData := map[string]interface{}{
			"username": "testuser",
			"email":    "test@example.com",
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
			t.Errorf("Valid user creation should return 201, got %d", w.Code)
		}

		var user map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &user)
		if err != nil {
			t.Fatalf("Failed to unmarshal user response: %v", err)
		}

		if user["username"] != "testuser" {
			t.Errorf("Expected username 'testuser', got %v", user["username"])
		}

		// Test duplicate username
		duplicateUserData := map[string]interface{}{
			"username": "admin", // Already exists
			"email":    "duplicate@example.com",
			"password": "TestPass123!",
			"role":     "viewer",
		}

		jsonData, _ = json.Marshal(duplicateUserData)
		req, _ = http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Duplicate username should return 400, got %d", w.Code)
		}

		// Test invalid user data
		invalidUserData := map[string]interface{}{
			"username": "ab", // Too short
			"email":    "invalid-email",
			"password": "weak",
			"role":     "invalid-role",
		}

		jsonData, _ = json.Marshal(invalidUserData)
		req, _ = http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Invalid user data should return 400, got %d", w.Code)
		}
	})

	// Test 2: User retrieval mutation
	t.Run("UserRetrievalMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)
		token := getMutationAuthToken(t, router)

		// Test get users
		req, _ := http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get users should return 200, got %d", w.Code)
		}

		var users []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &users)
		if err != nil {
			t.Fatalf("Failed to unmarshal users response: %v", err)
		}

		if len(users) < 1 {
			t.Error("Should have at least one user")
		}

		// Test get user stats
		req, _ = http.NewRequest("GET", "/api/users/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get user stats should return 200, got %d", w.Code)
		}

		var stats map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal user stats response: %v", err)
		}

		if stats["total_users"] == nil {
			t.Error("User stats should contain total_users")
		}
	})

	// Test 3: User update mutation
	t.Run("UserUpdateMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)
		token := getMutationAuthToken(t, router)

		// Test user update
		updateData := map[string]interface{}{
			"email": "updated@example.com",
			"role":  "editor",
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", "/api/users/testuser", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("User update should return 200, got %d", w.Code)
		}

		var user map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &user)
		if err != nil {
			t.Fatalf("Failed to unmarshal user response: %v", err)
		}

		if user["email"] != "updated@example.com" {
			t.Errorf("Expected email 'updated@example.com', got %v", user["email"])
		}

		if user["role"] != "editor" {
			t.Errorf("Expected role 'editor', got %v", user["role"])
		}

		// Test invalid user update
		invalidUpdateData := map[string]interface{}{
			"email": "invalid-email",
			"role":  "invalid-role",
		}

		jsonData, _ = json.Marshal(invalidUpdateData)
		req, _ = http.NewRequest("PUT", "/api/users/testuser", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Invalid user update should return 400, got %d", w.Code)
		}
	})
}

// TestMutationTestingSystemMonitoring tests system monitoring mutation
func TestMutationTestingSystemMonitoring(t *testing.T) {
	// Test 1: System stats mutation
	t.Run("SystemStatsMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)
		token := getMutationAuthToken(t, router)

		// Test get system stats
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get system stats should return 200, got %d", w.Code)
		}

		var stats map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &stats)
		if err != nil {
			t.Fatalf("Failed to unmarshal stats response: %v", err)
		}

		// Check required fields
		requiredFields := []string{"cpu_usage", "memory_usage", "disk_usage", "uptime"}
		for _, field := range requiredFields {
			if stats[field] == nil {
				t.Errorf("System stats should contain %s", field)
			}
		}

		// Verify values are within reasonable ranges
		if cpuUsage, ok := stats["cpu_usage"].(float64); ok {
			if cpuUsage < 0 || cpuUsage > 100 {
				t.Errorf("CPU usage should be between 0 and 100, got %f", cpuUsage)
			}
		}

		if memoryUsage, ok := stats["memory_usage"].(float64); ok {
			if memoryUsage < 0 || memoryUsage > 100 {
				t.Errorf("Memory usage should be between 0 and 100, got %f", memoryUsage)
			}
		}

		if diskUsage, ok := stats["disk_usage"].(float64); ok {
			if diskUsage < 0 || diskUsage > 100 {
				t.Errorf("Disk usage should be between 0 and 100, got %f", diskUsage)
			}
		}
	})

	// Test 2: Alerts mutation
	t.Run("AlertsMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)
		token := getMutationAuthToken(t, router)

		// Test get alerts
		req, _ := http.NewRequest("GET", "/api/alerts", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get alerts should return 200, got %d", w.Code)
		}

		var alerts []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &alerts)
		if err != nil {
			t.Fatalf("Failed to unmarshal alerts response: %v", err)
		}

		// Alerts can be empty, just verify it's an array
		if alerts == nil {
			t.Error("Alerts response should be an array")
		}

		// Test create alert
		alertData := map[string]interface{}{
			"type":     "cpu_high",
			"message":  "CPU usage is high",
			"severity": "warning",
		}

		jsonData, _ := json.Marshal(alertData)
		req, _ = http.NewRequest("POST", "/api/alerts", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Create alert should return 201, got %d", w.Code)
		}

		var alert map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &alert)
		if err != nil {
			t.Fatalf("Failed to unmarshal alert response: %v", err)
		}

		if alert["type"] != "cpu_high" {
			t.Errorf("Expected alert type 'cpu_high', got %v", alert["type"])
		}

		if alert["message"] != "CPU usage is high" {
			t.Errorf("Expected alert message 'CPU usage is high', got %v", alert["message"])
		}

		if alert["severity"] != "warning" {
			t.Errorf("Expected alert severity 'warning', got %v", alert["severity"])
		}
	})

	// Test 3: System logs mutation
	t.Run("SystemLogsMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)
		token := getMutationAuthToken(t, router)

		// Test get system logs
		req, _ := http.NewRequest("GET", "/api/logs?limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get system logs should return 200, got %d", w.Code)
		}

		var logs []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &logs)
		if err != nil {
			t.Fatalf("Failed to unmarshal logs response: %v", err)
		}

		// Logs can be empty, just verify it's an array
		if logs == nil {
			t.Error("Logs response should be an array")
		}

		// Test create log entry
		logData := map[string]interface{}{
			"level":   "info",
			"message": "Test log entry",
			"source":  "mutation_test",
		}

		jsonData, _ := json.Marshal(logData)
		req, _ = http.NewRequest("POST", "/api/logs", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Create log entry should return 201, got %d", w.Code)
		}

		var log map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &log)
		if err != nil {
			t.Fatalf("Failed to unmarshal log response: %v", err)
		}

		if log["level"] != "info" {
			t.Errorf("Expected log level 'info', got %v", log["level"])
		}

		if log["message"] != "Test log entry" {
			t.Errorf("Expected log message 'Test log entry', got %v", log["message"])
		}

		if log["source"] != "mutation_test" {
			t.Errorf("Expected log source 'mutation_test', got %v", log["source"])
		}
	})
}

// TestMutationTestingMoodleManagement tests Moodle management mutation
func TestMutationTestingMoodleManagement(t *testing.T) {
	// Test 1: Moodle status mutation
	t.Run("MoodleStatusMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)
		token := getMutationAuthToken(t, router)

		// Test get Moodle status
		req, _ := http.NewRequest("GET", "/api/moodle/status", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get Moodle status should return 200, got %d", w.Code)
		}

		var status map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &status)
		if err != nil {
			t.Fatalf("Failed to unmarshal Moodle status response: %v", err)
		}

		// Check required fields
		requiredFields := []string{"running", "version", "uptime", "last_check"}
		for _, field := range requiredFields {
			if status[field] == nil {
				t.Errorf("Moodle status should contain %s", field)
			}
		}

		// Test Moodle start
		req, _ = http.NewRequest("POST", "/api/moodle/start", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Accept both success and failure (Moodle might not be installed)
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("Moodle start should return 200 or 500, got %d", w.Code)
		}

		// Test Moodle stop
		req, _ = http.NewRequest("POST", "/api/moodle/stop", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Accept both success and failure (Moodle might not be installed)
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("Moodle stop should return 200 or 500, got %d", w.Code)
		}

		// Test Moodle restart
		req, _ = http.NewRequest("POST", "/api/moodle/restart", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Accept both success and failure (Moodle might not be installed)
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("Moodle restart should return 200 or 500, got %d", w.Code)
		}
	})

	// Test 2: Moodle configuration mutation
	t.Run("MoodleConfigurationMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)
		token := getMutationAuthToken(t, router)

		// Test get Moodle configuration
		req, _ := http.NewRequest("GET", "/api/moodle/config", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Get Moodle configuration should return 200, got %d", w.Code)
		}

		var config map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &config)
		if err != nil {
			t.Fatalf("Failed to unmarshal Moodle configuration response: %v", err)
		}

		// Check required fields
		requiredFields := []string{"database", "wwwroot", "dataroot", "admin"}
		for _, field := range requiredFields {
			if config[field] == nil {
				t.Errorf("Moodle configuration should contain %s", field)
			}
		}

		// Test update Moodle configuration
		configData := map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
				"port": 3306,
				"name": "moodle",
				"user": "moodle",
			},
			"wwwroot": "http://localhost/moodle",
			"dataroot": "/var/moodledata",
		}

		jsonData, _ := json.Marshal(configData)
		req, _ = http.NewRequest("PUT", "/api/moodle/config", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Update Moodle configuration should return 200, got %d", w.Code)
		}

		var updatedConfig map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &updatedConfig)
		if err != nil {
			t.Fatalf("Failed to unmarshal updated Moodle configuration response: %v", err)
		}

		if updatedConfig["wwwroot"] != "http://localhost/moodle" {
			t.Errorf("Expected wwwroot 'http://localhost/moodle', got %v", updatedConfig["wwwroot"])
		}

		if updatedConfig["dataroot"] != "/var/moodledata" {
			t.Errorf("Expected dataroot '/var/moodledata', got %v", updatedConfig["dataroot"])
		}
	})
}

// TestMutationTestingErrorHandling tests error handling mutation
func TestMutationTestingErrorHandling(t *testing.T) {
	// Test 1: HTTP error handling mutation
	t.Run("HTTPErrorHandlingMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)

		// Test invalid endpoint
		req, _ := http.NewRequest("GET", "/api/nonexistent", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Invalid endpoint should return 404, got %d", w.Code)
		}

		// Test invalid method
		req, _ = http.NewRequest("DELETE", "/health", nil)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Invalid method should return 405, got %d", w.Code)
		}

		// Test malformed JSON
		req, _ = http.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Malformed JSON should return 400, got %d", w.Code)
		}

		// Test missing content type
		req, _ = http.NewRequest("POST", "/login", bytes.NewBufferString("{}"))

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Missing content type should return 400, got %d", w.Code)
		}
	})

	// Test 2: Validation error handling mutation
	t.Run("ValidationErrorHandlingMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)
		token := getMutationAuthToken(t, router)

		// Test invalid user creation
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
			t.Errorf("Invalid user data should return 400, got %d", w.Code)
		}

		var errorResponse map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if errorResponse["error"] == nil {
			t.Error("Error response should contain error message")
		}

		if errorResponse["details"] == nil {
			t.Error("Error response should contain validation details")
		}

		// Test invalid user update
		invalidUpdateData := map[string]interface{}{
			"email": "invalid-email",
			"role":  "invalid-role",
		}

		jsonData, _ = json.Marshal(invalidUpdateData)
		req, _ = http.NewRequest("PUT", "/api/users/testuser", bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Invalid user update should return 400, got %d", w.Code)
		}
	})

	// Test 3: System error handling mutation
	t.Run("SystemErrorHandlingMutation", func(t *testing.T) {
		router := setupMutationTestRouter(t)
		token := getMutationAuthToken(t, router)

		// Test system error simulation
		req, _ := http.NewRequest("GET", "/api/error/simulate", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("System error simulation should return 500, got %d", w.Code)
		}

		var errorResponse map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		if err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if errorResponse["error"] == nil {
			t.Error("Error response should contain error message")
		}

		if errorResponse["code"] == nil {
			t.Error("Error response should contain error code")
		}

		if errorResponse["timestamp"] == nil {
			t.Error("Error response should contain timestamp")
		}
	})
}

// Helper functions

func setupMutationTestRouter(t *testing.T) *gin.Engine {
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

	db := setupMutationTestDB(t)

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

func getMutationAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupMutationTestDB(t *testing.T) *sql.DB {
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
