package unit

import (
	"testing"
	"time"

	"lms-manager/models"
	"lms-manager/services"
	"lms-manager/utils"

	_ "github.com/mattn/go-sqlite3"
)

func TestAuthService_Login(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	// Create auth service
	authService := services.NewAuthService("test-secret", db)

	// Create test user
	user, err := authService.CreateUser(&models.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "TestPass123!",
		Role:     models.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test successful login
	response, err := authService.Login("testuser", "TestPass123!", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if response.Token == "" {
		t.Error("Token should not be empty")
	}

	if response.User.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", response.User.Username)
	}

	// Test invalid password
	_, err = authService.Login("testuser", "wrongpassword", "127.0.0.1", "test-agent")
	if err == nil {
		t.Error("Login should fail with invalid password")
	}

	// Test invalid username
	_, err = authService.Login("nonexistent", "TestPass123!", "127.0.0.1", "test-agent")
	if err == nil {
		t.Error("Login should fail with invalid username")
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	// Create auth service
	authService := services.NewAuthService("test-secret", db)

	// Create test user
	user, err := authService.CreateUser(&models.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "TestPass123!",
		Role:     models.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Login to get token
	response, err := authService.Login("testuser", "TestPass123!", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Test valid token
	claims, err := authService.ValidateToken(response.Token)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID '%s', got '%s'", user.ID, claims.UserID)
	}

	if claims.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", claims.Username)
	}

	// Test invalid token
	_, err = authService.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Token validation should fail with invalid token")
	}
}

func TestAuthService_CreateUser(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	// Create auth service
	authService := services.NewAuthService("test-secret", db)

	// Test valid user creation
	user, err := authService.CreateUser(&models.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "NewPass123!",
		Role:     models.RoleOperator,
	})
	if err != nil {
		t.Fatalf("User creation failed: %v", err)
	}

	if user.Username != "newuser" {
		t.Errorf("Expected username 'newuser', got '%s'", user.Username)
	}

	if user.Email != "newuser@example.com" {
		t.Errorf("Expected email 'newuser@example.com', got '%s'", user.Email)
	}

	if user.Role != string(models.RoleOperator) {
		t.Errorf("Expected role 'operator', got '%s'", user.Role)
	}

	// Test duplicate username
	_, err = authService.CreateUser(&models.CreateUserRequest{
		Username: "newuser",
		Email:    "another@example.com",
		Password: "AnotherPass123!",
		Role:     models.RoleViewer,
	})
	if err == nil {
		t.Error("User creation should fail with duplicate username")
	}

	// Test duplicate email
	_, err = authService.CreateUser(&models.CreateUserRequest{
		Username: "anotheruser",
		Email:    "newuser@example.com",
		Password: "AnotherPass123!",
		Role:     models.RoleViewer,
	})
	if err == nil {
		t.Error("User creation should fail with duplicate email")
	}

	// Test invalid password
	_, err = authService.CreateUser(&models.CreateUserRequest{
		Username: "weakuser",
		Email:    "weak@example.com",
		Password: "weak",
		Role:     models.RoleViewer,
	})
	if err == nil {
		t.Error("User creation should fail with weak password")
	}
}

func TestAuthService_GetUsers(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	// Create auth service
	authService := services.NewAuthService("test-secret", db)

	// Create test users
	_, err := authService.CreateUser(&models.CreateUserRequest{
		Username: "user1",
		Email:    "user1@example.com",
		Password: "Pass123!",
		Role:     models.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	_, err = authService.CreateUser(&models.CreateUserRequest{
		Username: "user2",
		Email:    "user2@example.com",
		Password: "Pass123!",
		Role:     models.RoleOperator,
	})
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Get all users
	users, err := authService.GetUsers()
	if err != nil {
		t.Fatalf("Failed to get users: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestAuthService_GetUserStats(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	// Create auth service
	authService := services.NewAuthService("test-secret", db)

	// Create test users
	_, err := authService.CreateUser(&models.CreateUserRequest{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "Pass123!",
		Role:     models.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	_, err = authService.CreateUser(&models.CreateUserRequest{
		Username: "operator",
		Email:    "operator@example.com",
		Password: "Pass123!",
		Role:     models.RoleOperator,
	})
	if err != nil {
		t.Fatalf("Failed to create operator user: %v", err)
	}

	// Get user stats
	stats, err := authService.GetUserStats()
	if err != nil {
		t.Fatalf("Failed to get user stats: %v", err)
	}

	if stats.TotalUsers != 2 {
		t.Errorf("Expected 2 total users, got %d", stats.TotalUsers)
	}

	if stats.AdminUsers != 1 {
		t.Errorf("Expected 1 admin user, got %d", stats.AdminUsers)
	}

	if stats.OperatorUsers != 1 {
		t.Errorf("Expected 1 operator user, got %d", stats.OperatorUsers)
	}
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
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

	return db
}
