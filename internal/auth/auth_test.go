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

func TestSaveToken(t *testing.T) {
	os.Unsetenv("SPECSYNC_TOKEN")

	tempDir, err := os.MkdirTemp("", "specsync-auth-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origConfigDir := configDir
	configDir = tempDir
	defer func() { configDir = origConfigDir }()

	// save a token
	if err := SaveToken("mockup-token"); err != nil {
		t.Fatalf("SaveToken() failed: %v", err)
	}

	// verify file exists, can be read
	token := GetToken()
	if token != "mockup-token" {
		t.Errorf("GetToken() = %q, want %q", token, "mockup-token")
	}

	// verify file perms (should be 0600)
	credPath := filepath.Join(tempDir, "credentials.json")
	info, err := os.Stat(credPath)
	if err != nil {
		t.Fatalf("failed to stat credentials.json: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("credentials.json permissions = %o, want 0600", info.Mode().Perm())
	}
}

func TestDeleteToken(t *testing.T) {
	os.Unsetenv("SPECSYNC_TOKEN")

	tempDir, err := os.MkdirTemp("", "specsync-auth-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origConfigDir := configDir
	configDir = tempDir
	defer func() { configDir = origConfigDir }()

	// save, delete
	if err := SaveToken("boxer-token"); err != nil {
		t.Fatalf("SaveToken() failed: %v", err)
	}
	if err := DeleteToken(); err != nil {
		t.Fatalf("DeleteToken() failed: %v", err)
	}

	// verify token is gone
	token := GetToken()
	if token != "" {
		t.Errorf("GetToken() = %q, want empty string", token)
	}
}

func TestDeleteToken_NonExistent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "specsync-auth-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origConfigDir := configDir
	configDir = tempDir
	defer func() { configDir = origConfigDir }()

	// delete should not error
	if err := DeleteToken(); err != nil {
		t.Errorf("DeleteToken() = %v, want nil", err)
	}
}

func TestGetTokenSource_Env(t *testing.T) {
	os.Setenv("SPECSYNC_TOKEN", "squealer-token")
	defer os.Unsetenv("SPECSYNC_TOKEN")

	source := GetTokenSource()
	if source != TokenSourceEnv {
		t.Errorf("GetTokenSource() = %v, want TokenSourceEnv", source)
	}
}

func TestGetTokenSource_File(t *testing.T) {
	os.Unsetenv("SPECSYNC_TOKEN")

	tempDir, err := os.MkdirTemp("", "specsync-auth-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origConfigDir := configDir
	configDir = tempDir
	defer func() { configDir = origConfigDir }()

	if err := SaveToken("johannes-fust-token"); err != nil {
		t.Fatalf("SaveToken() failed: %v", err)
	}

	source := GetTokenSource()
	if source != TokenSourceFile {
		t.Errorf("GetTokenSource() = %v, want TokenSourceFile", source)
	}
}

func TestGetTokenSource_None(t *testing.T) {
	os.Unsetenv("SPECSYNC_TOKEN")

	origConfigDir := configDir
	configDir = "/path/no/exist/for/spec/sync/"
	defer func() { configDir = origConfigDir }()

	source := GetTokenSource()
	if source != TokenSourceNone {
		t.Errorf("GetTokenSource() = %v, want TokenSourceNone", source)
	}
}
