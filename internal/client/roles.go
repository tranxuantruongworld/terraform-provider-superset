package client

import (
	"context"
	"fmt"
	"net/url"
)

// GetRole retrieves a role by ID
func (c *Client) GetRole(ctx context.Context, id int64) (*Role, error) {
	path := fmt.Sprintf("/api/v1/security/roles/%d", id)
	var resp RoleResponse
	if err := c.Request(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	role := resp.Result
	role.ID = id
	return &role, nil
}

// GetRoleByName retrieves a role by its name
func (c *Client) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	risonQuery := fmt.Sprintf("(filters:!((col:name,opr:eq,val:'%s')))", name)
	path := fmt.Sprintf("/api/v1/security/roles/?q=%s", url.QueryEscape(risonQuery))

	var listResp RoleListResponse
	if err := c.Request(ctx, "GET", path, nil, &listResp); err != nil {
		return nil, err
	}

	for _, r := range listResp.Result {
		if r.Name == name {
			return c.GetRole(ctx, r.ID)
		}
	}

	return nil, &APIError{
		StatusCode: 404,
		Message:    fmt.Sprintf("role with name '%s' not found", name),
	}
}

// CreateRole creates a new role in Superset
func (c *Client) CreateRole(ctx context.Context, req *CreateRoleRequest) (*Role, error) {
	path := "/api/v1/security/roles/"
	var resp struct {
		ID     int64 `json:"id"`
		Result Role  `json:"result"`
	}

	if err := c.Request(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}

	if resp.ID != 0 {
		return c.GetRole(ctx, resp.ID)
	}
	return &resp.Result, nil
}

// UpdateRole updates an existing role in Superset
func (c *Client) UpdateRole(ctx context.Context, id int64, req *UpdateRoleRequest) (*Role, error) {
	path := fmt.Sprintf("/api/v1/security/roles/%d", id)
	var resp struct {
		ID     int64 `json:"id"`
		Result Role  `json:"result"`
	}

	if err := c.Request(ctx, "PUT", path, req, &resp); err != nil {
		return nil, err
	}

	return c.GetRole(ctx, id)
}

// DeleteRole deletes a role by ID
func (c *Client) DeleteRole(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/api/v1/security/roles/%d", id)
	return c.Request(ctx, "DELETE", path, nil, nil)
}
