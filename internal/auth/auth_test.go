package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGetToken_EnvVar(t *testing.T) {
	os.Setenv("SPECSYNC_TOKEN", "env-04550479")
	defer os.Unsetenv("SPECSYNC_TOKEN")

	token := GetToken()
	if token != "env-04550479" {
		t.Errorf("GetToken() = %q, want %q", token, "env-04550479")
	}
}

func TestGetToken_CredentialsFile(t *testing.T) {
	os.Unsetenv("SPECSYNC_TOKEN")

	// create temp config dir
	tempDir, err := os.MkdirTemp("", "specsync-auth-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// override config dir for test
	origConfigDir := configDir
	configDir = tempDir
	defer func() { configDir = origConfigDir }()

	// create credentials.json
	creds := map[string]string{"token": "token-04550479"}
	data, _ := json.Marshal(creds)
	credPath := filepath.Join(tempDir, "credentials.json")
	if err := os.WriteFile(credPath, data, 0600); err != nil {
		t.Fatalf("failed to write credentials.json: %v", err)
	}

	token := GetToken()
	if token != "token-04550479" {
		t.Errorf("GetToken() = %q, want %q", token, "token-04550479")
	}
}

func TestGetToken_NoAuth(t *testing.T) {
	os.Unsetenv("SPECSYNC_TOKEN")

	// use non-existent config dir
	origConfigDir := configDir
	configDir = "/path/dont/exist/for/this/test"
	defer func() { configDir = origConfigDir }()

	token := GetToken()
	if token != "" {
		t.Errorf("GetToken() = %q, want empty string", token)
	}
}

func TestGetToken_EnvTakesPrecedence(t *testing.T) {
	os.Setenv("SPECSYNC_TOKEN", "env-wins")
	defer os.Unsetenv("SPECSYNC_TOKEN")

	// create temp config dir with credentials
	tempDir, err := os.MkdirTemp("", "specsync-auth-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origConfigDir := configDir
	configDir = tempDir
	defer func() { configDir = origConfigDir }()

	creds := map[string]string{"token": "file-loses"}
	data, _ := json.Marshal(creds)
	credPath := filepath.Join(tempDir, "credentials.json")
	if err := os.WriteFile(credPath, data, 0600); err != nil {
		t.Fatalf("failed to write credentials.json: %v", err)
	}

	token := GetToken()
	if token != "env-wins" {
		t.Errorf("GetToken() = %q, want %q", token, "env-wins")
	}
}
