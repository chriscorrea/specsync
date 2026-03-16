package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"time"
)

type PushRequest struct {
	ProjectID string
	Path      string
	Content   string
	Type      string
	Hash      string
	Timestamp time.Time
}

type PushResponse struct {
	Hash      string
	UpdatedAt time.Time
}

// wire format for API request
type pushRequestBody struct {
	ProjectID string `json:"project_id"`
	Path      string `json:"path"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	Hash      string `json:"hash"`
	Timestamp string `json:"timestamp"`
}

// wire format for API response
type pushResponseBody struct {
	Hash      string `json:"hash"`
	UpdatedAt string `json:"updated_at"`
	Error     string `json:"error,omitempty"`
}

// httpClient interface for testing
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// newHTTPClient creates the HTTP client; override in tests
var newHTTPClient = func(isLocalhost bool) httpClient {
	return NewSafeClient(isLocalhost)
}

// Push sends a file to the API
func Push(ctx context.Context, apiURL string, token string, req *PushRequest) (*PushResponse, error) {

	// first, validate the URL to prevent misconfig, SSRF, etc.
	if err := ValidateAPIURL(apiURL); err != nil {
		return nil, err
	}

	// build request body
	body := pushRequestBody{
		ProjectID: req.ProjectID,
		Path:      req.Path,
		Content:   req.Content,
		Type:      req.Type,
		Hash:      req.Hash,
		Timestamp: req.Timestamp.UTC().Format(time.RFC3339),
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// build URL
	url := fmt.Sprintf("%s/projects/%s/files", apiURL, req.ProjectID)

	// create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	// create client (allows localhost for dev)
	parsed, _ := neturl.Parse(apiURL)
	host := parsed.Hostname()
	isLocalhost := host == "localhost" || host == "127.0.0.1"
	client := newHTTPClient(isLocalhost)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// parse response
	var respBody pushResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// check status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := respBody.Error
		if errMsg == "" {
			errMsg = http.StatusText(resp.StatusCode)
		}
		return nil, fmt.Errorf("push failed (%d): %s", resp.StatusCode, errMsg)
	}

	// parse updated_at and return response
	updatedAt, _ := time.Parse(time.RFC3339, respBody.UpdatedAt)

	return &PushResponse{
		Hash:      respBody.Hash,
		UpdatedAt: updatedAt,
	}, nil
}
