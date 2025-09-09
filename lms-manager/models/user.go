package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Hidden from JSON
	Role         string    `json:"role"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
}

// UserRole represents user roles
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleOperator UserRole = "operator"
	RoleViewer   UserRole = "viewer"
)

// IsValidRole checks if the role is valid
func (r UserRole) IsValid() bool {
	switch r {
	case RoleAdmin, RoleOperator, RoleViewer:
		return true
	default:
		return false
	}
}

// HasPermission checks if the role has permission for an action
func (r UserRole) HasPermission(action string) bool {
	switch r {
	case RoleAdmin:
		return true // Admin has all permissions
	case RoleOperator:
		// Operator can perform most actions except user management
		operatorActions := []string{
			"view_dashboard",
			"manage_moodle",
			"view_logs",
			"manage_backups",
			"view_system_stats",
		}
		for _, allowedAction := range operatorActions {
			if action == allowedAction {
				return true
			}
		}
		return false
	case RoleViewer:
		// Viewer can only view
		viewerActions := []string{
			"view_dashboard",
			"view_logs",
			"view_system_stats",
		}
		for _, allowedAction := range viewerActions {
			if action == allowedAction {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// UserActivity represents user activity
type UserActivity struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	User      User      `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ChangePasswordRequest represents a change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// CreateUserRequest represents a create user request
type CreateUserRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=50"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8"`
	Role     UserRole `json:"role" binding:"required"`
}

// UpdateUserRequest represents an update user request
type UpdateUserRequest struct {
	Username *string   `json:"username,omitempty"`
	Email    *string   `json:"email,omitempty"`
	Role     *UserRole `json:"role,omitempty"`
	Active   *bool     `json:"active,omitempty"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers     int `json:"total_users"`
	ActiveUsers    int `json:"active_users"`
	OnlineUsers    int `json:"online_users"`
	AdminUsers     int `json:"admin_users"`
	OperatorUsers  int `json:"operator_users"`
	ViewerUsers    int `json:"viewer_users"`
	Last24Hours    int `json:"last_24_hours"`
	Last7Days      int `json:"last_7_days"`
	Last30Days     int `json:"last_30_days"`
}
