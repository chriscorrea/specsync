package specpress

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/doyensec/safeurl"
)

const defaultBaseURL = "https://spec.press"

// httpDoer is an interface for making HTTP requests
type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client for spec.press API
type Client struct {
	httpClient httpDoer
	token      string
	baseURL    string
}

// ClientOption configures the client
type ClientOption func(*Client)

// WithBaseURL sets custom API base URL (for testing)
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client (for testing)
func WithHTTPClient(client httpDoer) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// NewClient creates a new spec.press API client
func NewClient(token string, opts ...ClientOption) *Client {
	config := safeurl.GetConfigBuilder().
		SetTimeout(30*time.Second).
		SetAllowedSchemes("http", "https").
		SetAllowedPorts(80, 443, 8000, 8080, 8443, 3000, 5000).
		SetAllowedIPs("127.0.0.1"). // allow localhost for testing
		EnableIPv6(false).
		Build()

	c := &Client{
		httpClient: safeurl.Client(config),
		token:      token,
		baseURL:    defaultBaseURL,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// createProjectResponse is the API response for project creation
type createProjectResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CreateProject creates a new project, returns project ID
func (c *Client) CreateProject(name string) (string, error) {
	body := map[string]string{"name": name}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/projects", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case 201:
		var result createProjectResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return "", fmt.Errorf("failed to parse response: %w", err)
		}
		return result.ID, nil

	case 401:
		return "", fmt.Errorf("authentication failed: invalid or expired token")

	case 409:
		return "", fmt.Errorf("project name already exists")

	case 429:
		return "", fmt.Errorf("rate limit exceeded: try again later")

	default:
		return "", fmt.Errorf("unexpected error (status %d): %s", resp.StatusCode, string(respBody))
	}
}
