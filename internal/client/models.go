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

// NamedEntity represents an entity with ID and Name
type NamedEntity struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// TableRef represents a dataset/table reference in RLS
type TableRef struct {
	ID        int64  `json:"id"`
	TableName string `json:"table_name,omitempty"`
	Schema    string `json:"schema,omitempty"`
}

// PermissionViewMenu represents a permission-resource pair in Superset (FAB)
type PermissionViewMenu struct {
	ID         int64       `json:"id"`
	Permission NamedEntity `json:"permission"`
	ViewMenu   NamedEntity `json:"view_menu"`
}

// PermissionViewMenuListResponse represents response from /api/v1/security/permissions-resources/
type PermissionViewMenuListResponse struct {
	Count  int                  `json:"count"`
	Result []PermissionViewMenu `json:"result"`
}

// PermissionViewMenuResponse represents single response
type PermissionViewMenuResponse struct {
	ID     int64              `json:"id"`
	Result PermissionViewMenu `json:"result"`
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
	ID          int64                `json:"id"`
	Name        string               `json:"name"`
	Permissions []PermissionViewMenu `json:"permissions,omitempty"`
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

// RowLevelSecurityRule represents an RLS filter rule in Superset
type RowLevelSecurityRule struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	FilterType  string     `json:"filter_type"`
	Clause      string     `json:"clause"`
	Tables      []TableRef `json:"tables,omitempty"`
	Roles       []RoleRef  `json:"roles,omitempty"`
}

// CreateRLSRequest represents the payload to create an RLS rule
type CreateRLSRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	FilterType  string  `json:"filter_type"`
	Clause      string  `json:"clause"`
	Tables      []int64 `json:"tables"`
	Roles       []int64 `json:"roles,omitempty"`
}

// UpdateRLSRequest represents the payload to update an RLS rule
type UpdateRLSRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	FilterType  *string `json:"filter_type,omitempty"`
	Clause      *string `json:"clause,omitempty"`
	Tables      []int64 `json:"tables,omitempty"`
	Roles       []int64 `json:"roles,omitempty"`
}

// RLSResponse represents single RLS response
type RLSResponse struct {
	ID     int64                `json:"id,omitempty"`
	Result RowLevelSecurityRule `json:"result"`
}

// RLSListResponse represents list RLS response
type RLSListResponse struct {
	Count  int                    `json:"count"`
	Result []RowLevelSecurityRule `json:"result"`
}

// APIError represents an error response from Superset REST API
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}
