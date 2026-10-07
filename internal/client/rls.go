package client

import (
	"context"
	"fmt"
)

// GetRLS retrieves an RLS rule by ID
func (c *Client) GetRLS(ctx context.Context, id int64) (*RowLevelSecurityRule, error) {
	path := fmt.Sprintf("/api/v1/rowlevelsecurity/%d", id)
	var resp RLSResponse
	if err := c.Request(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	rls := resp.Result
	rls.ID = id
	return &rls, nil
}

// CreateRLS creates a new Row-Level Security rule
func (c *Client) CreateRLS(ctx context.Context, req *CreateRLSRequest) (*RowLevelSecurityRule, error) {
	path := "/api/v1/rowlevelsecurity/"
	var resp struct {
		ID     int64                `json:"id"`
		Result RowLevelSecurityRule `json:"result"`
	}

	if err := c.Request(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}

	if resp.ID != 0 {
		return c.GetRLS(ctx, resp.ID)
	}
	return &resp.Result, nil
}

// UpdateRLS updates an existing Row-Level Security rule
func (c *Client) UpdateRLS(ctx context.Context, id int64, req *UpdateRLSRequest) (*RowLevelSecurityRule, error) {
	path := fmt.Sprintf("/api/v1/rowlevelsecurity/%d", id)
	var resp struct {
		ID     int64                `json:"id"`
		Result RowLevelSecurityRule `json:"result"`
	}

	if err := c.Request(ctx, "PUT", path, req, &resp); err != nil {
		return nil, err
	}

	return c.GetRLS(ctx, id)
}

// DeleteRLS deletes an RLS rule by ID
func (c *Client) DeleteRLS(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/api/v1/rowlevelsecurity/%d", id)
	return c.Request(ctx, "DELETE", path, nil, nil)
}
