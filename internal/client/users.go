package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetUser retrieves a user by ID
func (c *Client) GetUser(ctx context.Context, id int64) (*User, error) {
	path := fmt.Sprintf("/api/v1/security/users/%d", id)
	var resp UserResponse
	if err := c.Request(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	user := resp.Result
	user.ID = id
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	// Query using Superset Rison format: (filters:!((col:username,opr:eq,val:'<username>')))
	risonQuery := fmt.Sprintf("(filters:!((col:username,opr:eq,val:'%s')))", username)
	path := fmt.Sprintf("/api/v1/security/users/?q=%s", url.QueryEscape(risonQuery))

	var listResp UserListResponse
	if err := c.Request(ctx, "GET", path, nil, &listResp); err != nil {
		return nil, err
	}

	for _, u := range listResp.Result {
		if u.Username == username {
			return c.GetUser(ctx, u.ID)
		}
	}

	return nil, &APIError{
		StatusCode: 404,
		Message:    fmt.Sprintf("user with username '%s' not found", username),
	}
}

// CreateUser creates a new user in Superset
func (c *Client) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	path := "/api/v1/security/users/"
	var resp struct {
		ID     int64 `json:"id"`
		Result User  `json:"result"`
	}

	if err := c.Request(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}

	if resp.ID != 0 {
		return c.GetUser(ctx, resp.ID)
	}
	return &resp.Result, nil
}

// UpdateUser updates an existing user in Superset
func (c *Client) UpdateUser(ctx context.Context, id int64, req *UpdateUserRequest) (*User, error) {
	path := fmt.Sprintf("/api/v1/security/users/%d", id)
	var resp struct {
		ID     int64 `json:"id"`
		Result User  `json:"result"`
	}

	if err := c.Request(ctx, "PUT", path, req, &resp); err != nil {
		return nil, err
	}

	return c.GetUser(ctx, id)
}

// DeleteUser deletes a user by ID
func (c *Client) DeleteUser(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/api/v1/security/users/%d", id)
	return c.Request(ctx, "DELETE", path, nil, nil)
}
