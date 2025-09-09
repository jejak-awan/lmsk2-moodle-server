package property_based

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"lms-manager/config"
	"lms-manager/handlers"
	"lms-manager/services"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// TestPropertyBasedAuthentication tests authentication properties
func TestPropertyBasedAuthentication(t *testing.T) {
	// Test 1: Login property - valid credentials should always return token
	t.Run("LoginProperty", func(t *testing.T) {
		router := setupPropertyBasedTestRouter(t)

		// Test with valid credentials
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
			t.Errorf("Valid login should always return 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal login response: %v", err)
		}

		// Property: valid login should always return token
		if response["token"] == nil {
			t.Error("Property violated: valid login should always return token")
		}

		// Property: token should always be a string
		if _, ok := response["token"].(string); !ok {
			t.Error("Property violated: token should always be a string")
		}

		// Property: token should always be non-empty
		token := response["token"].(string)
		if len(token) == 0 {
			t.Error("Property violated: token should always be non-empty")
		}

		// Property: token should always be longer than 50 characters
		if len(token) < 50 {
			t.Error("Property violated: token should always be longer than 50 characters")
		}

		// Property: user object should always be present
		if response["user"] == nil {
			t.Error("Property violated: user object should always be present")
		}

		// Property: user object should always have required fields
		user := response["user"].(map[string]interface{})
		requiredFields := []string{"id", "username", "email", "role"}
		for _, field := range requiredFields {
			if user[field] == nil {
				t.Errorf("Property violated: user object should always have %s field", field)
			}
		}
	})

	// Test 2: Token validation property - valid token should always allow access
	t.Run("TokenValidationProperty", func(t *testing.T) {
		router := setupPropertyBasedTestRouter(t)
		token := getPropertyBasedAuthToken(t, router)

		// Property: valid token should always allow access to protected endpoints
		endpoints := []string{"/api/stats", "/api/users", "/api/alerts", "/api/logs"}

		for _, endpoint := range endpoints {
			req, _ := http.NewRequest("GET", endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Property violated: valid token should always allow access to %s, got %d", endpoint, w.Code)
			}
		}
	})

	// Test 3: Authentication failure property - invalid credentials should always return 401
	t.Run("AuthenticationFailureProperty", func(t *testing.T) {
		router := setupPropertyBasedTestRouter(t)

		// Test with invalid credentials
		invalidCredentials := []map[string]string{
			{"username": "admin", "password": "wrongpassword"},
			{"username": "wronguser", "password": "admin123"},
			{"username": "wronguser", "password": "wrongpassword"},
			{"username": "", "password": "admin123"},
			{"username": "admin", "password": ""},
			{"username": "", "password": ""},
		}

		for _, creds := range invalidCredentials {
			jsonData, _ := json.Marshal(creds)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Property: invalid credentials should always return 401 or 400
			if w.Code != http.StatusUnauthorized && w.Code != http.StatusBadRequest {
				t.Errorf("Property violated: invalid credentials should always return 401 or 400, got %d", w.Code)
			}

			// Property: invalid credentials should never return token
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err == nil && response["token"] != nil {
				t.Error("Property violated: invalid credentials should never return token")
			}
		}
	})
}

// TestPropertyBasedUserManagement tests user management properties
func TestPropertyBasedUserManagement(t *testing.T) {
	// Test 1: User creation property - valid user data should always create user
	t.Run("UserCreationProperty", func(t *testing.T) {
		router := setupPropertyBasedTestRouter(t)
		token := getPropertyBasedAuthToken(t, router)

		// Test with valid user data
		validUserData := []map[string]interface{}{
			{
				"username": "user1",
				"email":    "user1@example.com",
				"password": "TestPass123!",
				"role":     "viewer",
			},
			{
				"username": "user2",
				"email":    "user2@example.com",
				"password": "TestPass456!",
				"role":     "editor",
			},
			{
				"username": "user3",
				"email":    "user3@example.com",
				"password": "TestPass789!",
				"role":     "admin",
			},
		}

		for i, userData := range validUserData {
			jsonData, _ := json.Marshal(userData)
			req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Property: valid user data should always return 201
			if w.Code != http.StatusCreated {
				t.Errorf("Property violated: valid user data should always return 201, got %d", w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal user creation response: %v", err)
			}

			// Property: created user should always have required fields
			requiredFields := []string{"id", "username", "email", "role", "active", "created_at", "updated_at"}
			for _, field := range requiredFields {
				if response[field] == nil {
					t.Errorf("Property violated: created user should always have %s field", field)
				}
			}

			// Property: created user should always have correct values
			if response["username"] != userData["username"] {
				t.Errorf("Property violated: created user username should match input, expected %v, got %v", userData["username"], response["username"])
			}

			if response["email"] != userData["email"] {
				t.Errorf("Property violated: created user email should match input, expected %v, got %v", userData["email"], response["email"])
			}

			if response["role"] != userData["role"] {
				t.Errorf("Property violated: created user role should match input, expected %v, got %v", userData["role"], response["role"])
			}

			// Property: created user should always be active
			if response["active"] != true {
				t.Errorf("Property violated: created user should always be active, got %v", response["active"])
			}

			// Property: created user should never expose password
			if response["password"] != nil {
				t.Error("Property violated: created user should never expose password")
			}

			if response["password_hash"] != nil {
				t.Error("Property violated: created user should never expose password_hash")
			}

			// Property: created user should always have unique ID
			if response["id"] == nil {
				t.Error("Property violated: created user should always have unique ID")
			}

			// Property: created user should always have timestamps
			if response["created_at"] == nil {
				t.Error("Property violated: created user should always have created_at timestamp")
			}

			if response["updated_at"] == nil {
				t.Error("Property violated: created user should always have updated_at timestamp")
			}

			// Property: created user should always have valid email format
			email := response["email"].(string)
			emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
			if !emailRegex.MatchString(email) {
				t.Errorf("Property violated: created user should always have valid email format, got %s", email)
			}

			// Property: created user should always have valid role
			role := response["role"].(string)
			validRoles := []string{"viewer", "editor", "admin"}
			valid := false
			for _, r := range validRoles {
				if role == r {
					valid = true
					break
				}
			}
			if !valid {
				t.Errorf("Property violated: created user should always have valid role, got %s", role)
			}

			// Property: created user should always have valid username format
			username := response["username"].(string)
			if len(username) < 3 {
				t.Errorf("Property violated: created user should always have username with at least 3 characters, got %s", username)
			}

			if len(username) > 50 {
				t.Errorf("Property violated: created user should always have username with at most 50 characters, got %s", username)
			}

			// Property: created user should always have alphanumeric username
			usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
			if !usernameRegex.MatchString(username) {
				t.Errorf("Property violated: created user should always have alphanumeric username, got %s", username)
			}

			// Property: created user should always have unique username
			if i > 0 {
				// Check that username is unique among created users
				for j := 0; j < i; j++ {
					if username == validUserData[j]["username"] {
						t.Errorf("Property violated: created user should always have unique username, got duplicate %s", username)
					}
				}
			}
		}
	})

	// Test 2: User retrieval property - get users should always return array
	t.Run("UserRetrievalProperty", func(t *testing.T) {
		router := setupPropertyBasedTestRouter(t)
		token := getPropertyBasedAuthToken(t, router)

		// Property: get users should always return 200
		req, _ := http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Property violated: get users should always return 200, got %d", w.Code)
		}

		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal users response: %v", err)
		}

		// Property: get users should always return array
		if response == nil {
			t.Error("Property violated: get users should always return array")
		}

		// Property: get users should always return at least one user (admin)
		if len(response) < 1 {
			t.Error("Property violated: get users should always return at least one user")
		}

		// Property: each user should always have required fields
		for i, user := range response {
			requiredFields := []string{"id", "username", "email", "role", "active", "created_at", "updated_at"}
			for _, field := range requiredFields {
				if user[field] == nil {
					t.Errorf("Property violated: user %d should always have %s field", i, field)
				}
			}

			// Property: each user should never expose password
			if user["password"] != nil {
				t.Errorf("Property violated: user %d should never expose password", i)
			}

			if user["password_hash"] != nil {
				t.Errorf("Property violated: user %d should never expose password_hash", i)
			}

			// Property: each user should always have valid email format
			email := user["email"].(string)
			emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
			if !emailRegex.MatchString(email) {
				t.Errorf("Property violated: user %d should always have valid email format, got %s", i, email)
			}

			// Property: each user should always have valid role
			role := user["role"].(string)
			validRoles := []string{"viewer", "editor", "admin"}
			valid := false
			for _, r := range validRoles {
				if role == r {
					valid = true
					break
				}
			}
			if !valid {
				t.Errorf("Property violated: user %d should always have valid role, got %s", i, role)
			}

			// Property: each user should always have valid username format
			username := user["username"].(string)
			if len(username) < 3 {
				t.Errorf("Property violated: user %d should always have username with at least 3 characters, got %s", i, username)
			}

			if len(username) > 50 {
				t.Errorf("Property violated: user %d should always have username with at most 50 characters, got %s", i, username)
			}

			// Property: each user should always have alphanumeric username
			usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
			if !usernameRegex.MatchString(username) {
				t.Errorf("Property violated: user %d should always have alphanumeric username, got %s", i, username)
			}

			// Property: each user should always have unique username
			for j := i + 1; j < len(response); j++ {
				if username == response[j]["username"] {
					t.Errorf("Property violated: user %d should always have unique username, got duplicate %s", i, username)
				}
			}
		}
	})

	// Test 3: User stats property - get user stats should always return valid stats
	t.Run("UserStatsProperty", func(t *testing.T) {
		router := setupPropertyBasedTestRouter(t)
		token := getPropertyBasedAuthToken(t, router)

		// Property: get user stats should always return 200
		req, _ := http.NewRequest("GET", "/api/users/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Property violated: get user stats should always return 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal user stats response: %v", err)
		}

		// Property: get user stats should always return required fields
		requiredFields := []string{"total_users", "active_users", "inactive_users", "users_by_role", "new_users_today", "new_users_this_week", "new_users_this_month"}
		for _, field := range requiredFields {
			if response[field] == nil {
				t.Errorf("Property violated: user stats should always have %s field", field)
			}
		}

		// Property: user stats should always have valid numeric values
		numericFields := []string{"total_users", "active_users", "inactive_users", "new_users_today", "new_users_this_week", "new_users_this_month"}
		for _, field := range numericFields {
			if value, ok := response[field].(float64); !ok {
				t.Errorf("Property violated: user stats %s should always be a number, got %T", field, response[field])
			} else if value < 0 {
				t.Errorf("Property violated: user stats %s should always be non-negative, got %f", field, value)
			}
		}

		// Property: user stats should always have valid users_by_role object
		usersByRole := response["users_by_role"].(map[string]interface{})
		if usersByRole == nil {
			t.Error("Property violated: user stats users_by_role should always be an object")
		}

		// Property: user stats should always have valid role counts
		validRoles := []string{"viewer", "editor", "admin"}
		for _, role := range validRoles {
			if value, ok := usersByRole[role].(float64); !ok {
				t.Errorf("Property violated: user stats users_by_role[%s] should always be a number, got %T", role, usersByRole[role])
			} else if value < 0 {
				t.Errorf("Property violated: user stats users_by_role[%s] should always be non-negative, got %f", role, value)
			}
		}

		// Property: user stats should always have consistent totals
		totalUsers := response["total_users"].(float64)
		activeUsers := response["active_users"].(float64)
		inactiveUsers := response["inactive_users"].(float64)

		if totalUsers != activeUsers+inactiveUsers {
			t.Errorf("Property violated: user stats total_users should always equal active_users + inactive_users, got %f != %f + %f", totalUsers, activeUsers, inactiveUsers)
		}

		// Property: user stats should always have consistent role totals
		roleTotal := 0.0
		for _, role := range validRoles {
			if value, ok := usersByRole[role].(float64); ok {
				roleTotal += value
			}
		}

		if totalUsers != roleTotal {
			t.Errorf("Property violated: user stats total_users should always equal sum of users_by_role, got %f != %f", totalUsers, roleTotal)
		}
	})
}

// TestPropertyBasedSystemMonitoring tests system monitoring properties
func TestPropertyBasedSystemMonitoring(t *testing.T) {
	// Test 1: System stats property - get stats should always return valid stats
	t.Run("SystemStatsProperty", func(t *testing.T) {
		router := setupPropertyBasedTestRouter(t)
		token := getPropertyBasedAuthToken(t, router)

		// Property: get system stats should always return 200
		req, _ := http.NewRequest("GET", "/api/stats", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Property violated: get system stats should always return 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal system stats response: %v", err)
		}

		// Property: get system stats should always return required fields
		requiredFields := []string{"cpu_usage", "memory_usage", "disk_usage", "uptime", "load_average", "network_io", "disk_io", "timestamp"}
		for _, field := range requiredFields {
			if response[field] == nil {
				t.Errorf("Property violated: system stats should always have %s field", field)
			}
		}

		// Property: system stats should always have valid numeric values
		numericFields := []string{"cpu_usage", "memory_usage", "disk_usage", "uptime"}
		for _, field := range numericFields {
			if value, ok := response[field].(float64); !ok {
				t.Errorf("Property violated: system stats %s should always be a number, got %T", field, response[field])
			} else if value < 0 {
				t.Errorf("Property violated: system stats %s should always be non-negative, got %f", field, value)
			}
		}

		// Property: system stats should always have valid percentage values
		percentageFields := []string{"cpu_usage", "memory_usage", "disk_usage"}
		for _, field := range percentageFields {
			if value, ok := response[field].(float64); ok {
				if value < 0 || value > 100 {
					t.Errorf("Property violated: system stats %s should always be between 0 and 100, got %f", field, value)
				}
			}
		}

		// Property: system stats should always have valid uptime
		uptime := response["uptime"].(float64)
		if uptime < 0 {
			t.Errorf("Property violated: system stats uptime should always be non-negative, got %f", uptime)
		}

		// Property: system stats should always have valid load average
		loadAverage := response["load_average"].([]interface{})
		if loadAverage == nil {
			t.Error("Property violated: system stats load_average should always be an array")
		}

		if len(loadAverage) != 3 {
			t.Errorf("Property violated: system stats load_average should always have 3 elements, got %d", len(loadAverage))
		}

		for i, load := range loadAverage {
			if value, ok := load.(float64); !ok {
				t.Errorf("Property violated: system stats load_average[%d] should always be a number, got %T", i, load)
			} else if value < 0 {
				t.Errorf("Property violated: system stats load_average[%d] should always be non-negative, got %f", i, value)
			}
		}

		// Property: system stats should always have valid network I/O
		networkIO := response["network_io"].(map[string]interface{})
		if networkIO == nil {
			t.Error("Property violated: system stats network_io should always be an object")
		}

		networkIOFields := []string{"bytes_sent", "bytes_recv", "packets_sent", "packets_recv"}
		for _, field := range networkIOFields {
			if value, ok := networkIO[field].(float64); !ok {
				t.Errorf("Property violated: system stats network_io[%s] should always be a number, got %T", field, networkIO[field])
			} else if value < 0 {
				t.Errorf("Property violated: system stats network_io[%s] should always be non-negative, got %f", field, value)
			}
		}

		// Property: system stats should always have valid disk I/O
		diskIO := response["disk_io"].(map[string]interface{})
		if diskIO == nil {
			t.Error("Property violated: system stats disk_io should always be an object")
		}

		diskIOFields := []string{"read_bytes", "write_bytes", "read_count", "write_count"}
		for _, field := range diskIOFields {
			if value, ok := diskIO[field].(float64); !ok {
				t.Errorf("Property violated: system stats disk_io[%s] should always be a number, got %T", field, diskIO[field])
			} else if value < 0 {
				t.Errorf("Property violated: system stats disk_io[%s] should always be non-negative, got %f", field, value)
			}
		}

		// Property: system stats should always have valid timestamp
		timestamp := response["timestamp"].(string)
		if timestamp == "" {
			t.Error("Property violated: system stats timestamp should always be non-empty")
		}

		// Property: system stats should always have valid timestamp format
		timestampRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)
		if !timestampRegex.MatchString(timestamp) {
			t.Errorf("Property violated: system stats timestamp should always have valid format, got %s", timestamp)
		}
	})

	// Test 2: Alerts property - get alerts should always return array
	t.Run("AlertsProperty", func(t *testing.T) {
		router := setupPropertyBasedTestRouter(t)
		token := getPropertyBasedAuthToken(t, router)

		// Property: get alerts should always return 200
		req, _ := http.NewRequest("GET", "/api/alerts", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Property violated: get alerts should always return 200, got %d", w.Code)
		}

		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal alerts response: %v", err)
		}

		// Property: get alerts should always return array
		if response == nil {
			t.Error("Property violated: get alerts should always return array")
		}

		// Property: each alert should always have required fields
		for i, alert := range response {
			requiredFields := []string{"id", "type", "message", "severity", "resolved", "created_at", "resolved_at"}
			for _, field := range requiredFields {
				if alert[field] == nil {
					t.Errorf("Property violated: alert %d should always have %s field", i, field)
				}
			}

			// Property: each alert should always have valid severity
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
				t.Errorf("Property violated: alert %d should always have valid severity, got %s", i, severity)
			}

			// Property: each alert should always have valid resolved status
			resolved := alert["resolved"].(bool)
			if resolved != true && resolved != false {
				t.Errorf("Property violated: alert %d should always have valid resolved status, got %v", i, resolved)
			}

			// Property: each alert should always have valid message
			message := alert["message"].(string)
			if message == "" {
				t.Errorf("Property violated: alert %d should always have non-empty message", i)
			}

			// Property: each alert should always have valid type
			alertType := alert["type"].(string)
			if alertType == "" {
				t.Errorf("Property violated: alert %d should always have non-empty type", i)
			}

			// Property: each alert should always have valid ID
			id := alert["id"].(string)
			if id == "" {
				t.Errorf("Property violated: alert %d should always have non-empty ID", i)
			}

			// Property: each alert should always have unique ID
			for j := i + 1; j < len(response); j++ {
				if id == response[j]["id"] {
					t.Errorf("Property violated: alert %d should always have unique ID, got duplicate %s", i, id)
				}
			}

			// Property: each alert should always have valid timestamps
			createdAt := alert["created_at"].(string)
			if createdAt == "" {
				t.Errorf("Property violated: alert %d should always have non-empty created_at", i)
			}

			// Property: each alert should always have valid timestamp format
			timestampRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)
			if !timestampRegex.MatchString(createdAt) {
				t.Errorf("Property violated: alert %d should always have valid created_at format, got %s", i, createdAt)
			}
		}
	})

	// Test 3: System logs property - get logs should always return array
	t.Run("SystemLogsProperty", func(t *testing.T) {
		router := setupPropertyBasedTestRouter(t)
		token := getPropertyBasedAuthToken(t, router)

		// Property: get system logs should always return 200
		req, _ := http.NewRequest("GET", "/api/logs?limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Property violated: get system logs should always return 200, got %d", w.Code)
		}

		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal system logs response: %v", err)
		}

		// Property: get system logs should always return array
		if response == nil {
			t.Error("Property violated: get system logs should always return array")
		}

		// Property: each log should always have required fields
		for i, log := range response {
			requiredFields := []string{"id", "level", "message", "source", "created_at"}
			for _, field := range requiredFields {
				if log[field] == nil {
					t.Errorf("Property violated: log %d should always have %s field", i, field)
				}
			}

			// Property: each log should always have valid level
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
				t.Errorf("Property violated: log %d should always have valid level, got %s", i, level)
			}

			// Property: each log should always have valid message
			message := log["message"].(string)
			if message == "" {
				t.Errorf("Property violated: log %d should always have non-empty message", i)
			}

			// Property: each log should always have valid source
			source := log["source"].(string)
			if source == "" {
				t.Errorf("Property violated: log %d should always have non-empty source", i)
			}

			// Property: each log should always have valid ID
			id := log["id"].(string)
			if id == "" {
				t.Errorf("Property violated: log %d should always have non-empty ID", i)
			}

			// Property: each log should always have unique ID
			for j := i + 1; j < len(response); j++ {
				if id == response[j]["id"] {
					t.Errorf("Property violated: log %d should always have unique ID, got duplicate %s", i, id)
				}
			}

			// Property: each log should always have valid timestamp
			createdAt := log["created_at"].(string)
			if createdAt == "" {
				t.Errorf("Property violated: log %d should always have non-empty created_at", i)
			}

			// Property: each log should always have valid timestamp format
			timestampRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)
			if !timestampRegex.MatchString(createdAt) {
				t.Errorf("Property violated: log %d should always have valid created_at format, got %s", i, createdAt)
			}
		}
	})
}

// Helper functions

func setupPropertyBasedTestRouter(t *testing.T) *gin.Engine {
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

	db := setupPropertyBasedTestDB(t)

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

func getPropertyBasedAuthToken(t *testing.T, router *gin.Engine) string {
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

func setupPropertyBasedTestDB(t *testing.T) *sql.DB {
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
