package client

// LoginRequest represents the payload for authentication
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Provider string `json:"provider"`
	Refresh  bool   `json:"refresh"`
}

// LoginResponse represents the response from /api/v1/security/login
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// CSRFResponse represents the response from /api/v1/security/csrf_token/
type CSRFResponse struct {
	Result string `json:"result"`
}

// RoleRef represents a minimal reference to a role
type RoleRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name,omitempty"`
}

// PermissionRef represents a permission view menu reference
type PermissionRef struct {
	ID         int64  `json:"id"`
	Permission string `json:"permission_name,omitempty"`
	ViewMenu   string `json:"view_menu_name,omitempty"`
}

// User represents a Superset user model
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Active    bool      `json:"active"`
	Roles     []RoleRef `json:"roles,omitempty"`
}

// CreateUserRequest represents the payload to create a new user
type CreateUserRequest struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Active    bool    `json:"active"`
	Password  string  `json:"password"`
	Roles     []int64 `json:"roles"`
}

// UpdateUserRequest represents the payload to update an existing user
type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Username  *string `json:"username,omitempty"`
	Email     *string `json:"email,omitempty"`
	Active    *bool   `json:"active,omitempty"`
	Password  *string `json:"password,omitempty"`
	Roles     []int64 `json:"roles,omitempty"`
}

// UserResponse represents standard single user response
type UserResponse struct {
	ID     int64 `json:"id,omitempty"`
	Result User  `json:"result"`
}

// UserListResponse represents list response for users
type UserListResponse struct {
	Count  int    `json:"count"`
	Result []User `json:"result"`
}

// Role represents a Superset role
type Role struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Permissions []PermissionRef `json:"permissions,omitempty"`
}

// CreateRoleRequest represents the payload to create a role
type CreateRoleRequest struct {
	Name        string  `json:"name"`
	Permissions []int64 `json:"permissions,omitempty"`
}

// UpdateRoleRequest represents the payload to update a role
type UpdateRoleRequest struct {
	Name        string  `json:"name,omitempty"`
	Permissions []int64 `json:"permissions,omitempty"`
}

// RoleResponse represents standard single role response
type RoleResponse struct {
	ID     int64 `json:"id,omitempty"`
	Result Role  `json:"result"`
}

// RoleListResponse represents list response for roles
type RoleListResponse struct {
	Count  int    `json:"count"`
	Result []Role `json:"result"`
}

// APIError represents an error response from Superset REST API
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}
