package client

import (
	"context"
	"fmt"
	"net/url"
)

// ListPermissionViewMenus retrieves all permission-resource combinations
func (c *Client) ListPermissionViewMenus(ctx context.Context) ([]PermissionViewMenu, error) {
	// Request with page_size=1000 to get all available permissions
	path := "/api/v1/security/permissions-resources/?q=(page:0,page_size:1000)"
	var resp PermissionViewMenuListResponse
	if err := c.Request(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Result, nil
}

// GetPermissionViewMenuByID retrieves a PermissionViewMenu by its ID
func (c *Client) GetPermissionViewMenuByID(ctx context.Context, id int64) (*PermissionViewMenu, error) {
	path := fmt.Sprintf("/api/v1/security/permissions-resources/%d", id)
	var resp PermissionViewMenuResponse
	if err := c.Request(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	pvm := resp.Result
	pvm.ID = id
	return &pvm, nil
}

// GetPermissionViewMenu retrieves a PermissionViewMenu by permission name and view_menu name
func (c *Client) GetPermissionViewMenu(ctx context.Context, permissionName, viewMenuName string) (*PermissionViewMenu, error) {
	// Try Rison filter first, fallback to listing if needed
	risonQuery := fmt.Sprintf("(filters:!((col:permission_name,opr:eq,val:'%s'),(col:view_menu_name,opr:eq,val:'%s')))", permissionName, viewMenuName)
	path := fmt.Sprintf("/api/v1/security/permissions-resources/?q=%s", url.QueryEscape(risonQuery))

	var resp PermissionViewMenuListResponse
	err := c.Request(ctx, "GET", path, nil, &resp)
	if err == nil && len(resp.Result) > 0 {
		return &resp.Result[0], nil
	}

	// Fallback to scanning list
	all, listErr := c.ListPermissionViewMenus(ctx)
	if listErr != nil {
		return nil, listErr
	}

	for _, item := range all {
		if item.Permission.Name == permissionName && item.ViewMenu.Name == viewMenuName {
			return &item, nil
		}
	}

	return nil, &APIError{
		StatusCode: 404,
		Message:    fmt.Sprintf("permission view menu with permission '%s' on view '%s' not found", permissionName, viewMenuName),
	}
}
