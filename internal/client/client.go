package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Client handles interaction with the Apache Superset REST API
type Client struct {
	BaseURL    string
	Username   string
	Password   string
	Insecure   bool
	HTTPClient *http.Client

	mu          sync.RWMutex
	accessToken string
	csrfToken   string
}

// NewClient initializes a new Superset API client
func NewClient(baseURL, username, password string, insecure bool) (*Client, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base_url '%s': %w", baseURL, err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	httpClient := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   30 * time.Second,
	}

	// Normalize base url: remove trailing slash
	cleanBaseURL := strings.TrimRight(parsedURL.String(), "/")

	c := &Client{
		BaseURL:    cleanBaseURL,
		Username:   username,
		Password:   password,
		Insecure:   insecure,
		HTTPClient: httpClient,
	}

	return c, nil
}

// Authenticate performs login and fetches JWT access token and CSRF token
func (c *Client) Authenticate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	loginPayload := LoginRequest{
		Username: c.Username,
		Password: c.Password,
		Provider: "db",
		Refresh:  true,
	}

	bodyBytes, err := json.Marshal(loginPayload)
	if err != nil {
		return fmt.Errorf("failed to serialize login payload: %w", err)
	}

	loginURL := fmt.Sprintf("%s/api/v1/security/login", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("authentication failed (status %d): %s", resp.StatusCode, string(respBody)),
		}
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	c.accessToken = loginResp.AccessToken

	// Fetch CSRF token if available
	csrfURL := fmt.Sprintf("%s/api/v1/security/csrf_token/", c.BaseURL)
	csrfReq, err := http.NewRequestWithContext(ctx, http.MethodGet, csrfURL, nil)
	if err == nil {
		csrfReq.Header.Set("Authorization", "Bearer "+c.accessToken)
		csrfResp, csrfErr := c.HTTPClient.Do(csrfReq)
		if csrfErr == nil {
			defer csrfResp.Body.Close()
			if csrfResp.StatusCode == http.StatusOK {
				var csrfData CSRFResponse
				if json.NewDecoder(csrfResp.Body).Decode(&csrfData) == nil {
					c.csrfToken = csrfData.Result
				}
			}
		}
	}

	return nil
}

// Request executes an HTTP request with authentication and token retry logic
func (c *Client) Request(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	c.mu.RLock()
	token := c.accessToken
	c.mu.RUnlock()

	if token == "" && c.Username != "" && c.Password != "" {
		if err := c.Authenticate(ctx); err != nil {
			return err
		}
	}

	err := c.execute(ctx, method, path, body, result)
	if err != nil {
		// Retry once on 401 Unauthorized
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == http.StatusUnauthorized {
			if authErr := c.Authenticate(ctx); authErr != nil {
				return authErr
			}
			return c.execute(ctx, method, path, body, result)
		}
		return err
	}
	return nil
}

func (c *Client) execute(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	fullURL := fmt.Sprintf("%s%s", c.BaseURL, path)

	var reqBody io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to serialize request body: %w", err)
		}
		reqBody = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.mu.RLock()
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	if c.csrfToken != "" {
		req.Header.Set("X-CSRFToken", c.csrfToken)
	}
	c.mu.RUnlock()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request %s %s failed: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBytes, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("API request error (%d %s): %s", resp.StatusCode, resp.Status, string(respBytes)),
		}
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

