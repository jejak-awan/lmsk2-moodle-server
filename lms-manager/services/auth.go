package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"lms-manager/models"
	"lms-manager/utils"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService handles authentication
type AuthService struct {
	jwtSecret string
	db        *sql.DB
}

// NewAuthService creates a new auth service
func NewAuthService(jwtSecret string, db *sql.DB) *AuthService {
	return &AuthService{
		jwtSecret: jwtSecret,
		db:        db,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Login authenticates a user
func (a *AuthService) Login(username, password string, ip, userAgent string) (*models.LoginResponse, error) {
	// Get user from database
	user, err := a.getUserByUsername(username)
	if err != nil {
		a.logSecurityEvent("login_failed", fmt.Sprintf("Invalid username: %s", username), ip, userAgent, "warning")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.Active {
		a.logSecurityEvent("login_failed", fmt.Sprintf("Inactive user: %s", username), ip, userAgent, "warning")
		return nil, fmt.Errorf("account is disabled")
	}

	// Verify password
	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		a.logSecurityEvent("login_failed", fmt.Sprintf("Invalid password for user: %s", username), ip, userAgent, "warning")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token, expiresAt, err := a.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// Create session
	sessionID := utils.GenerateID()
	err = a.createSession(sessionID, user.ID, ip, userAgent, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	// Update last login
	err = a.updateLastLogin(user.ID)
	if err != nil {
		utils.Warn("Failed to update last login: %v", err)
	}

	// Log successful login
	a.logSecurityEvent("login_success", fmt.Sprintf("User logged in: %s", username), ip, userAgent, "info")
	a.logUserActivity(user.ID, "login", "authentication", ip, userAgent, true, "")

	return &models.LoginResponse{
		Token:     token,
		User:      *user,
		ExpiresAt: expiresAt,
	}, nil
}

// Logout logs out a user
func (a *AuthService) Logout(userID, ip, userAgent string) error {
	// Log logout activity
	a.logUserActivity(userID, "logout", "authentication", ip, userAgent, true, "")

	// Remove session (optional - for stateless JWT)
	// In a stateless JWT system, we don't need to remove sessions
	// But we can log the logout event

	utils.Info("User logged out: %s", userID)
	return nil
}

// ValidateToken validates a JWT token
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check if user still exists and is active
		user, err := a.getUserByID(claims.UserID)
		if err != nil {
			return nil, fmt.Errorf("user not found")
		}

		if !user.Active {
			return nil, fmt.Errorf("user account is disabled")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetUser returns a user by ID
func (a *AuthService) GetUser(userID string) (*models.User, error) {
	return a.getUserByID(userID)
}

// CreateUser creates a new user
func (a *AuthService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	// Validate input
	if !utils.IsValidUsername(req.Username) {
		return nil, fmt.Errorf("invalid username")
	}

	if !utils.IsValidEmail(req.Email) {
		return nil, fmt.Errorf("invalid email")
	}

	if !utils.IsValidPassword(req.Password) {
		return nil, fmt.Errorf("password must be at least 8 characters with uppercase, lowercase, digit, and special character")
	}

	if !req.Role.IsValid() {
		return nil, fmt.Errorf("invalid role")
	}

	// Check if username already exists
	var count int
	err := a.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", req.Username).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %v", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("username already exists")
	}

	// Check if email already exists
	err = a.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %v", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user
	userID := utils.GenerateID()
	now := time.Now()

	_, err = a.db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, req.Username, req.Email, passwordHash, string(req.Role), true, now, now)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Get created user
	user, err := a.getUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created user: %v", err)
	}

	utils.Info("User created: %s (%s)", user.Username, user.Email)
	return user, nil
}

// UpdateUser updates a user
func (a *AuthService) UpdateUser(userID string, req *models.UpdateUserRequest) (*models.User, error) {
	// Get existing user
	user, err := a.getUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	// Build update query
	setParts := []string{}
	args := []interface{}{}

	if req.Username != nil {
		if !utils.IsValidUsername(*req.Username) {
			return nil, fmt.Errorf("invalid username")
		}
		setParts = append(setParts, "username = ?")
		args = append(args, *req.Username)
	}

	if req.Email != nil {
		if !utils.IsValidEmail(*req.Email) {
			return nil, fmt.Errorf("invalid email")
		}
		setParts = append(setParts, "email = ?")
		args = append(args, *req.Email)
	}

	if req.Role != nil {
		if !req.Role.IsValid() {
			return nil, fmt.Errorf("invalid role")
		}
		setParts = append(setParts, "role = ?")
		args = append(args, string(*req.Role))
	}

	if req.Active != nil {
		setParts = append(setParts, "active = ?")
		args = append(args, *req.Active)
	}

	if len(setParts) == 0 {
		return user, nil // No changes
	}

	// Add updated_at
	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now())

	// Add userID to args
	args = append(args, userID)

	// Execute update
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(setParts, ", "))
	_, err = a.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	// Get updated user
	updatedUser, err := a.getUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user: %v", err)
	}

	utils.Info("User updated: %s", updatedUser.Username)
	return updatedUser, nil
}

// DeleteUser deletes a user
func (a *AuthService) DeleteUser(userID string) error {
	// Get user to check if exists
	user, err := a.getUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	// Don't allow deleting admin users
	if user.Role == "admin" {
		return fmt.Errorf("cannot delete admin user")
	}

	// Delete user
	_, err = a.db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	// Delete user sessions
	_, err = a.db.Exec("DELETE FROM user_sessions WHERE user_id = ?", userID)
	if err != nil {
		utils.Warn("Failed to delete user sessions: %v", err)
	}

	utils.Info("User deleted: %s", user.Username)
	return nil
}

// GetUsers returns all users
func (a *AuthService) GetUsers() ([]models.User, error) {
	rows, err := a.db.Query(`
		SELECT id, username, email, role, active, created_at, updated_at, last_login
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %v", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var lastLogin sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.Active,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLogin,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}

		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}

		users = append(users, user)
	}

	return users, nil
}

// ChangePassword changes a user's password
func (a *AuthService) ChangePassword(userID, currentPassword, newPassword string) error {
	// Get user
	user, err := a.getUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	// Verify current password
	if !utils.CheckPasswordHash(currentPassword, user.PasswordHash) {
		return fmt.Errorf("current password is incorrect")
	}

	// Validate new password
	if !utils.IsValidPassword(newPassword) {
		return fmt.Errorf("new password must be at least 8 characters with uppercase, lowercase, digit, and special character")
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %v", err)
	}

	// Update password in database
	_, err = a.db.Exec(`
		UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?
	`, newPasswordHash, time.Now(), userID)

	if err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	utils.Info("Password changed for user: %s", user.Username)
	return nil
}

// GetUserStats returns user statistics
func (a *AuthService) GetUserStats() (*models.UserStats, error) {
	stats := &models.UserStats{}

	// Total users
	err := a.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %v", err)
	}

	// Active users
	err = a.db.QueryRow("SELECT COUNT(*) FROM users WHERE active = 1").Scan(&stats.ActiveUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %v", err)
	}

	// Online users (sessions not expired)
	err = a.db.QueryRow(`
		SELECT COUNT(DISTINCT user_id) FROM user_sessions 
		WHERE expires_at > ?
	`, time.Now()).Scan(&stats.OnlineUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get online users: %v", err)
	}

	// Users by role
	err = a.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&stats.AdminUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin users: %v", err)
	}

	err = a.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'operator'").Scan(&stats.OperatorUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator users: %v", err)
	}

	err = a.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'viewer'").Scan(&stats.ViewerUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get viewer users: %v", err)
	}

	// Users by time
	now := time.Now()
	last24Hours := now.Add(-24 * time.Hour)
	last7Days := now.Add(-7 * 24 * time.Hour)
	last30Days := now.Add(-30 * 24 * time.Hour)

	err = a.db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at > ?", last24Hours).Scan(&stats.Last24Hours)
	if err != nil {
		return nil, fmt.Errorf("failed to get last 24 hours users: %v", err)
	}

	err = a.db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at > ?", last7Days).Scan(&stats.Last7Days)
	if err != nil {
		return nil, fmt.Errorf("failed to get last 7 days users: %v", err)
	}

	err = a.db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at > ?", last30Days).Scan(&stats.Last30Days)
	if err != nil {
		return nil, fmt.Errorf("failed to get last 30 days users: %v", err)
	}

	return stats, nil
}

// Helper methods

// getUserByUsername gets a user by username
func (a *AuthService) getUserByUsername(username string) (*models.User, error) {
	var user models.User
	var lastLogin sql.NullTime

	err := a.db.QueryRow(`
		SELECT id, username, email, password_hash, role, active, created_at, updated_at, last_login
		FROM users
		WHERE username = ?
	`, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

// getUserByID gets a user by ID
func (a *AuthService) getUserByID(userID string) (*models.User, error) {
	var user models.User
	var lastLogin sql.NullTime

	err := a.db.QueryRow(`
		SELECT id, username, email, role, active, created_at, updated_at, last_login
		FROM users
		WHERE id = ?
	`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

// generateToken generates a JWT token
func (a *AuthService) generateToken(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hours

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// createSession creates a user session
func (a *AuthService) createSession(sessionID, userID, ip, userAgent string, expiresAt time.Time) error {
	_, err := a.db.Exec(`
		INSERT INTO user_sessions (id, user_id, ip_address, user_agent, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, userID, ip, userAgent, expiresAt)

	return err
}

// updateLastLogin updates the last login time
func (a *AuthService) updateLastLogin(userID string) error {
	_, err := a.db.Exec(`
		UPDATE users SET last_login = ? WHERE id = ?
	`, time.Now(), userID)

	return err
}

// logSecurityEvent logs a security event
func (a *AuthService) logSecurityEvent(eventType, message, ip, userAgent, severity string) {
	_, err := a.db.Exec(`
		INSERT INTO security_events (id, type, message, ip_address, user_agent, severity)
		VALUES (?, ?, ?, ?, ?, ?)
	`, utils.GenerateID(), eventType, message, ip, userAgent, severity)

	if err != nil {
		utils.Error("Failed to log security event: %v", err)
	}
}

// logUserActivity logs user activity
func (a *AuthService) logUserActivity(userID, action, resource, ip, userAgent string, success bool, message string) {
	_, err := a.db.Exec(`
		INSERT INTO user_activities (id, user_id, action, resource, ip_address, user_agent, success, message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, utils.GenerateID(), userID, action, resource, ip, userAgent, success, message)

	if err != nil {
		utils.Error("Failed to log user activity: %v", err)
	}
}
