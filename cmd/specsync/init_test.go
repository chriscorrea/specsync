package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitClaudeCode_Integration(t *testing.T) {
	// first, build the binary
	binPath := filepath.Join(t.TempDir(), "specsync")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = "."
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}

	// create temp project dir
	projectDir := t.TempDir()

	// run init --json --claude-code
	cmd := exec.Command(binPath, "init", "--json", "--claude-code")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init --claude-code failed: %v\n%s", err, output)
	}

	// parse JSON output
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\n%s", err, output)
	}

	// verify claude_code_configured
	if configured, ok := result["claude_code_configured"].(bool); !ok || !configured {
		t.Errorf("claude_code_configured = %v, want true", result["claude_code_configured"])
	}

	// verify claude_code_files isn't empty
	files, ok := result["claude_code_files"].([]any)
	if !ok || len(files) == 0 {
		t.Errorf("claude_code_files = %v, want non-empty array", result["claude_code_files"])
	}

	// verify hook script exists & is executable
	hookPath := filepath.Join(projectDir, ".claude", "hooks", "specsync-push.sh")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Errorf("hook script not created: %v", err)
	} else if info.Mode()&0111 == 0 {
		t.Errorf("hook script not executable: mode = %o", info.Mode())
	}

	// verify settings.json has PostToolUse
	settingsPath := filepath.Join(projectDir, ".claude", "settings.json")
	settingsContent, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Errorf("settings.json not created: %v", err)
	} else if !strings.Contains(string(settingsContent), "PostToolUse") {
		t.Error("settings.json missing PostToolUse hook")
	}

	// verify rules file exists
	rulesPath := filepath.Join(projectDir, ".claude", "rules", "specsync.md")
	if _, err := os.Stat(rulesPath); err != nil {
		t.Errorf("rules file not created: %v", err)
	}

	// verify .specsync.yaml exists
	configPath := filepath.Join(projectDir, ".specsync.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf(".specsync.yaml not created: %v", err)
	}
}

func TestInitSpecpress_UUID(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "specsync")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build: %v\n%s", err, out)
	}

	projectDir := t.TempDir()
	testUUID := "8f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c"

	cmd := exec.Command(binPath, "init", "--specpress", testUUID, "--json")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init --specpress failed: %v\n%s", err, output)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\n%s", err, output)
	}

	if result["project_id"] != testUUID {
		t.Errorf("project_id = %v, want %s", result["project_id"], testUUID)
	}
	if result["api_url"] != "https://spec.press/api/v1" {
		t.Errorf("api_url = %v, want spec.press URL", result["api_url"])
	}
}

func TestInitSpecpress_URL(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "specsync")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build: %v\n%s", err, out)
	}

	projectDir := t.TempDir()
	testUUID := "8f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c"
	specpressURL := "https://spec.press/p/" + testUUID

	cmd := exec.Command(binPath, "init", "--specpress", specpressURL, "--json")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init --specpress (URL) failed: %v\n%s", err, output)
	}

	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\n%s", err, output)
	}

	if result["project_id"] != testUUID {
		t.Errorf("project_id = %v, want %s", result["project_id"], testUUID)
	}
}

func TestInitSpecpress_Create_NoToken(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "specsync")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build: %v\n%s", err, out)
	}

	projectDir := t.TempDir()

	cmd := exec.Command(binPath, "init", "--create", "test-project", "--json")
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "SPECSYNC_TOKEN=") // clear token
	output, err := cmd.CombinedOutput()

	// should fail
	if err == nil {
		t.Fatal("expected error when no token, got none")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "authentication required") {
		t.Errorf("expected 'authentication required' error, got: %s", outputStr)
	}
}

func TestInitSpecpress_Create_Conflict(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "specsync")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build: %v\n%s", err, out)
	}

	projectDir := t.TempDir()
	testUUID := "8f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c"

	// --specpress <id> --create should fail
	cmd := exec.Command(binPath, "init", "--specpress", testUUID, "--create", "test", "--json")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected conflict error, got none")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "cannot use --create") {
		t.Errorf("expected 'cannot use --create' error, got: %s", outputStr)
	}
}

func TestInitForce_OverwritesConfig(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "specsync")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build: %v\n%s", err, out)
	}

	projectDir := t.TempDir()
	configPath := filepath.Join(projectDir, ".specsync.yaml")

	// create existing config
	oldConfig := "project_id: old-uuid\n"
	if err := os.WriteFile(configPath, []byte(oldConfig), 0644); err != nil {
		t.Fatalf("failed to create existing config: %v", err)
	}

	testUUID := "8f3a2b1c-4d5e-6f7a-8b9c-0d1e2f3a4b5c"

	// run with --force
	cmd := exec.Command(binPath, "init", "--specpress", testUUID, "--force", "--json")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init --force failed: %v\n%s", err, output)
	}

	// verify new config written
	newConfig, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if !strings.Contains(string(newConfig), testUUID) {
		t.Errorf("config not overwritten, got: %s", newConfig)
	}
}

func TestInitSpecpress_Create_WithMockServer(t *testing.T) {
	// mock spec.press server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/projects" && r.Method == "POST" {
			w.WriteHeader(201)
			_, _ = w.Write([]byte(`{"id":"created-uuid-123","name":"test-project"}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	// this test validates the CLI flag parsing works
	t.Skip("mock server injection requires refactoring; covered by client_test.go")
}
