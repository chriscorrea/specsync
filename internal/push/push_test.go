package push

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func init() {
	// use standard http.Client in tests (bypasses safeurl restrictions)
	newHTTPClient = func(isLocalhost bool) httpClient {
		return &http.Client{Timeout: 30 * time.Second}
	}
}

func TestPush_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify method and path
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/projects/test-uuid/files" {
			t.Errorf("path = %s, want /projects/test-uuid/files", r.URL.Path)
		}

		// verify headers
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", ct)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("Authorization = %s, want Bearer test-token", auth)
		}

		// verify body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["path"] != "docs/auth.md" {
			t.Errorf("body.path = %v, want docs/auth.md", body["path"])
		}
		if body["type"] != "specification" {
			t.Errorf("body.type = %v, want specification", body["type"])
		}

		// respond
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"hash":       "sha256:abc123",
			"updated_at": "2026-03-14T10:30:00Z",
		})
	}))
	defer server.Close()

	req := &PushRequest{
		ProjectID: "test-uuid",
		Path:      "docs/auth.md",
		Content:   "# Auth",
		Type:      "specification",
		Hash:      "sha256:xyz",
		Timestamp: time.Now(),
	}

	resp, err := Push(context.Background(), server.URL, "test-token", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if resp.Hash != "sha256:abc123" {
		t.Errorf("resp.Hash = %s, want sha256:abc123", resp.Hash)
	}
}

func TestPush_NoToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify no auth header when token empty
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Errorf("Authorization = %s, want empty", auth)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"hash":       "sha256:abc123",
			"updated_at": "2026-03-14T10:30:00Z",
		})
	}))
	defer server.Close()

	req := &PushRequest{
		ProjectID: "test-uuid",
		Path:      "docs/auth.md",
		Content:   "# Auth",
		Type:      "document",
		Hash:      "sha256:xyz",
		Timestamp: time.Now(),
	}

	_, err := Push(context.Background(), server.URL, "", req)
	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}
}

func TestPush_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "internal server error",
		})
	}))
	defer server.Close()

	req := &PushRequest{
		ProjectID: "test-uuid",
		Path:      "docs/sample-spec.md",
		Content:   "# Spec for Push",
		Type:      "specification",
		Hash:      "sha256:xyzabc",
		Timestamp: time.Now(),
	}

	_, err := Push(context.Background(), server.URL, "", req)
	if err == nil {
		t.Error("Push() expected error for 500 response")
	}
}

func TestPush_InvalidURL(t *testing.T) {
	req := &PushRequest{
		ProjectID: "test-uuid",
		Path:      "docs/sample-spec.md",
		Content:   "# Spec for Push",
		Type:      "specification",
		Hash:      "sha256:xyzabc",
		Timestamp: time.Now(),
	}

	_, err := Push(context.Background(), "http://evil.com/api", "", req)
	if err == nil {
		t.Error("Push() expected error for non-HTTPS URL")
	}
}
