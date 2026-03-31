package specpress_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/specsync/internal/specpress"
)

func TestCreateProject(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		projName   string
		serverCode int
		serverBody string
		wantID     string
		wantErr    string
	}{
		{
			name:       "success",
			token:      "sk_test_token",
			projName:   "my-project",
			serverCode: 201,
			serverBody: `{"id":"abc-123","name":"my-project"}`,
			wantID:     "abc-123",
		},
		{
			name:       "unauthorized",
			token:      "bad_token",
			projName:   "my-project",
			serverCode: 401,
			serverBody: `{"error":"Unauthorized"}`,
			wantErr:    "authentication failed",
		},
		{
			name:       "name taken",
			token:      "sk_test_token",
			projName:   "taken-name",
			serverCode: 409,
			serverBody: `{"error":"Project name already exists"}`,
			wantErr:    "project name already exists",
		},
		{
			name:       "rate limited",
			token:      "sk_test_token",
			projName:   "my-project",
			serverCode: 429,
			serverBody: `{"error":"Rate limit exceeded"}`,
			wantErr:    "rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// verify auth header
				auth := r.Header.Get("Authorization")
				if !strings.HasPrefix(auth, "Bearer ") {
					w.WriteHeader(401)
					w.Write([]byte(`{"error":"Missing auth"}`))
					return
				}

				// verify method, path
				if r.Method != "POST" || r.URL.Path != "/api/v1/projects" {
					w.WriteHeader(404)
					return
				}

				// verify request body
				var body struct {
					Name string `json:"name"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
					w.WriteHeader(400)
					w.Write([]byte(`{"error":"Name required"}`))
					return
				}

				w.WriteHeader(tt.serverCode)
				w.Write([]byte(tt.serverBody))
			}))
			defer server.Close()

			// create client w/ test server URL
			client := specpress.NewClient(tt.token,
				specpress.WithBaseURL(server.URL),
				specpress.WithHTTPClient(server.Client()),
			)
			id, err := client.CreateProject(tt.projName)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if id != tt.wantID {
				t.Errorf("CreateProject() = %q, want %q", id, tt.wantID)
			}
		})
	}
}

func TestCreateProject_NetworkError(t *testing.T) {
	// use invalid URL to trigger network error
	client := specpress.NewClient("token", specpress.WithBaseURL("http://127.0.0.1:1"))
	_, err := client.CreateProject("test")

	if err == nil {
		t.Error("expected network error, got nil")
	}
}
