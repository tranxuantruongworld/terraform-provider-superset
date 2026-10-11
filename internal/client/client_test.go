package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientAuthentication(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/security/login" {
			var body LoginRequest
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Username == "admin" && body.Password == "admin" {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(LoginResponse{
					AccessToken: "fake-jwt-token-12345",
				})
				return
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/api/v1/security/csrf_token/" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(CSRFResponse{
				Result: "fake-csrf-token",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "admin", "admin", false)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	ctx := context.Background()
	if err := c.Authenticate(ctx); err != nil {
		t.Fatalf("failed to authenticate: %v", err)
	}

	if c.accessToken != "fake-jwt-token-12345" {
		t.Errorf("expected access token 'fake-jwt-token-12345', got '%s'", c.accessToken)
	}
	if c.csrfToken != "fake-csrf-token" {
		t.Errorf("expected csrf token 'fake-csrf-token', got '%s'", c.csrfToken)
	}
}

func TestUserCRUD(t *testing.T) {
	usersDB := map[int64]User{
		1: {
			ID:        1,
			Username:  "johndoe",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			Active:    true,
			Roles:     []RoleRef{{ID: 2, Name: "Gamma"}},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v1/security/login":
			_ = json.NewEncoder(w).Encode(LoginResponse{AccessToken: "token"})
		case r.URL.Path == "/api/v1/security/csrf_token/":
			_ = json.NewEncoder(w).Encode(CSRFResponse{Result: "csrf"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/security/users/1":
			_ = json.NewEncoder(w).Encode(UserResponse{Result: usersDB[1]})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/security/users/":
			var req CreateUserRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			newUser := User{
				ID:        2,
				Username:  req.Username,
				FirstName: req.FirstName,
				LastName:  req.LastName,
				Email:     req.Email,
				Active:    req.Active,
			}
			usersDB[2] = newUser
			_ = json.NewEncoder(w).Encode(struct {
				ID     int64 `json:"id"`
				Result User  `json:"result"`
			}{ID: 2, Result: newUser})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/security/users/2":
			_ = json.NewEncoder(w).Encode(UserResponse{Result: usersDB[2]})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/security/users/2":
			delete(usersDB, 2)
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "admin", "admin", false)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	ctx := context.Background()

	// Get User
	user, err := c.GetUser(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if user.Username != "johndoe" {
		t.Errorf("expected username 'johndoe', got '%s'", user.Username)
	}

	// Create User
	created, err := c.CreateUser(ctx, &CreateUserRequest{
		Username:  "alice",
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice@example.com",
		Active:    true,
		Password:  "Secret123!",
		Roles:     []int64{2},
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	if created.Username != "alice" {
		t.Errorf("expected username 'alice', got '%s'", created.Username)
	}

	// Delete User
	if err := c.DeleteUser(ctx, 2); err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}
}

func TestRoleCRUD(t *testing.T) {
	rolesDB := map[int64]Role{
		1: {
			ID:   1,
			Name: "DataEngineer",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v1/security/login":
			_ = json.NewEncoder(w).Encode(LoginResponse{AccessToken: "token"})
		case r.URL.Path == "/api/v1/security/csrf_token/":
			_ = json.NewEncoder(w).Encode(CSRFResponse{Result: "csrf"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/security/roles/1":
			_ = json.NewEncoder(w).Encode(RoleResponse{Result: rolesDB[1]})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/security/roles/":
			var req CreateRoleRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			newRole := Role{
				ID:   2,
				Name: req.Name,
			}
			rolesDB[2] = newRole
			_ = json.NewEncoder(w).Encode(struct {
				ID     int64 `json:"id"`
				Result Role  `json:"result"`
			}{ID: 2, Result: newRole})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/security/roles/2":
			_ = json.NewEncoder(w).Encode(RoleResponse{Result: rolesDB[2]})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/security/roles/2":
			delete(rolesDB, 2)
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "admin", "admin", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()

	// Get Role
	role, err := c.GetRole(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get role: %v", err)
	}
	if role.Name != "DataEngineer" {
		t.Errorf("expected role name 'DataEngineer', got '%s'", role.Name)
	}

	// Create Role
	created, err := c.CreateRole(ctx, &CreateRoleRequest{
		Name: "FinancialAnalyst",
	})
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}
	if created.Name != "FinancialAnalyst" {
		t.Errorf("expected role name 'FinancialAnalyst', got '%s'", created.Name)
	}

	// Delete Role
	if err := c.DeleteRole(ctx, 2); err != nil {
		t.Fatalf("failed to delete role: %v", err)
	}
}

func TestPermissionViewMenu(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v1/security/login":
			_ = json.NewEncoder(w).Encode(LoginResponse{AccessToken: "token"})
		case r.URL.Path == "/api/v1/security/permissions-resources/15":
			_ = json.NewEncoder(w).Encode(PermissionViewMenuResponse{
				Result: PermissionViewMenu{
					ID:         15,
					Permission: NamedEntity{ID: 1, Name: "can_read"},
					ViewMenu:   NamedEntity{ID: 2, Name: "Dashboard"},
				},
			})
		case r.URL.Path == "/api/v1/security/permissions-resources/":
			_ = json.NewEncoder(w).Encode(PermissionViewMenuListResponse{
				Count: 1,
				Result: []PermissionViewMenu{
					{
						ID:         15,
						Permission: NamedEntity{ID: 1, Name: "can_read"},
						ViewMenu:   NamedEntity{ID: 2, Name: "Dashboard"},
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "admin", "admin", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()

	pvm, err := c.GetPermissionViewMenu(ctx, "can_read", "Dashboard")
	if err != nil {
		t.Fatalf("failed to get permission view menu: %v", err)
	}
	if pvm.ID != 15 {
		t.Errorf("expected ID 15, got %d", pvm.ID)
	}

	pvmByID, err := c.GetPermissionViewMenuByID(ctx, 15)
	if err != nil {
		t.Fatalf("failed to get permission view menu by ID: %v", err)
	}
	if pvmByID.Permission.Name != "can_read" {
		t.Errorf("expected 'can_read', got '%s'", pvmByID.Permission.Name)
	}
}

func TestRLSCRUD(t *testing.T) {
	rlsDB := map[int64]RowLevelSecurityRule{
		10: {
			ID:          10,
			Name:        "APAC Only",
			FilterType:  "Regular",
			Clause:      "country = 'VN'",
			Tables:      []TableRef{{ID: 1, TableName: "sales"}},
			Roles:       []RoleRef{{ID: 2, Name: "Sales_VN"}},
			Description: "Filter Vietnam sales",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v1/security/login":
			_ = json.NewEncoder(w).Encode(LoginResponse{AccessToken: "token"})
		case r.URL.Path == "/api/v1/security/csrf_token/":
			_ = json.NewEncoder(w).Encode(CSRFResponse{Result: "csrf"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/rowlevelsecurity/10":
			_ = json.NewEncoder(w).Encode(RLSResponse{Result: rlsDB[10]})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/rowlevelsecurity/":
			var req CreateRLSRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			newRLS := RowLevelSecurityRule{
				ID:          11,
				Name:        req.Name,
				FilterType:  req.FilterType,
				Clause:      req.Clause,
				Description: req.Description,
			}
			rlsDB[11] = newRLS
			_ = json.NewEncoder(w).Encode(struct {
				ID     int64                `json:"id"`
				Result RowLevelSecurityRule `json:"result"`
			}{ID: 11, Result: newRLS})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/rowlevelsecurity/11":
			_ = json.NewEncoder(w).Encode(RLSResponse{Result: rlsDB[11]})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/rowlevelsecurity/11":
			delete(rlsDB, 11)
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "admin", "admin", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()

	// Get RLS
	rule, err := c.GetRLS(ctx, 10)
	if err != nil {
		t.Fatalf("failed to get RLS: %v", err)
	}
	if rule.Clause != "country = 'VN'" {
		t.Errorf("expected clause 'country = \\'VN\\'', got '%s'", rule.Clause)
	}

	// Create RLS
	created, err := c.CreateRLS(ctx, &CreateRLSRequest{
		Name:        "EMEA Filter",
		FilterType:  "Regular",
		Clause:      "region = 'EMEA'",
		Tables:      []int64{1},
		Roles:       []int64{3},
		Description: "EMEA region data",
	})
	if err != nil {
		t.Fatalf("failed to create RLS: %v", err)
	}
	if created.Name != "EMEA Filter" {
		t.Errorf("expected name 'EMEA Filter', got '%s'", created.Name)
	}

	// Delete RLS
	if err := c.DeleteRLS(ctx, 11); err != nil {
		t.Fatalf("failed to delete RLS: %v", err)
	}
}
