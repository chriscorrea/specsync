package claudecode

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectClaudeCode_WithDir(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.Mkdir(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	if !DetectClaudeCode(dir) {
		t.Error("DetectClaudeCode() = false, want true when .claude/ exists")
	}
}

func TestDetectClaudeCode_WithoutDir(t *testing.T) {
	dir := t.TempDir()

	if DetectClaudeCode(dir) {
		t.Error("DetectClaudeCode() = true, want false when .claude/ missing")
	}
}

func TestDetectClaudeCode_WithFile(t *testing.T) {
	dir := t.TempDir()
	// erroneoulsy create .claude as a file (instead of dir)
	claudeFile := filepath.Join(dir, ".claude")
	if err := os.WriteFile(claudeFile, []byte("not a dir"), 0644); err != nil {
		t.Fatalf("failed to create .claude file: %v", err)
	}

	if DetectClaudeCode(dir) {
		t.Error("DetectClaudeCode() = true, want false when .claude is a file")
	}
}

func TestHookScript_ContainsDefensiveValidation(t *testing.T) {
	if !contains(HookScript, `push -- "$FILE_PATH"`) {
		t.Error("HookScript missing -- argument terminator")
	}
	if !contains(HookScript, "rejected suspicious file path") {
		t.Error("HookScript missing path rejection message")
	}
}

func TestHookScript_ContainsJqCheck(t *testing.T) {
	if !contains(HookScript, "command -v jq") {
		t.Error("HookScript missing jq dependency check")
	}
	// verify cross-platform hints
	if !contains(HookScript, "brew install jq") {
		t.Error("HookScript missing brew install hint")
	}
	if !contains(HookScript, "apt-get install jq") {
		t.Error("HookScript missing apt-get install hint")
	}
}

func TestSettingsHook_ContainsCorrectMatcher(t *testing.T) {
	if !contains(SettingsHook, HookMatcher) {
		t.Errorf("SettingsHook missing matcher %q", HookMatcher)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestMergeHooks_EmptySettings(t *testing.T) {
	result, err := MergeHooks(nil)
	if err != nil {
		t.Fatalf("MergeHooks(nil) error = %v", err)
	}

	// should contain our hook
	if !contains(string(result), "PostToolUse") {
		t.Error("result missing PostToolUse hook")
	}
	if !contains(string(result), HookMatcher) {
		t.Error("result missing matcher")
	}
}

func TestMergeHooks_PreservesExistingPermissions(t *testing.T) {
	existing := []byte(`{
  "permissions": {
    "allow": ["Bash(git:*)"]
  }
}`)

	result, err := MergeHooks(existing)
	if err != nil {
		t.Fatalf("MergeHooks() error = %v", err)
	}

	// should preserve permissions
	if !contains(string(result), "permissions") {
		t.Error("result missing permissions")
	}
	if !contains(string(result), "Bash(git:*)") {
		t.Error("result missing original permission rule")
	}
	// should add our hook
	if !contains(string(result), "PostToolUse") {
		t.Error("result missing PostToolUse hook")
	}
}

func TestMergeHooks_PreservesOtherHooks(t *testing.T) {
	existing := []byte(`{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [{"type": "command", "command": "echo test"}]
      }
    ]
  }
}`)

	result, err := MergeHooks(existing)
	if err != nil {
		t.Fatalf("MergeHooks() error = %v", err)
	}

	// preserve existing hooks
	if !contains(string(result), "PreToolUse") {
		t.Error("result missing PreToolUse hook")
	}
	// ...and add our hook
	if !contains(string(result), "PostToolUse") {
		t.Error("result missing PostToolUse hook")
	}
}

func TestMergeHooks_ExistingPostToolUse(t *testing.T) {
	existing := []byte(`{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [{"type": "command", "command": "echo existing"}]
      }
    ]
  }
}`)

	result, err := MergeHooks(existing)
	if err != nil {
		t.Fatalf("MergeHooks() error = %v", err)
	}

	// preserve existing PostToolUse hook
	if !contains(string(result), "echo existing") {
		t.Error("result missing existing PostToolUse hook command")
	}
	// add our matcher
	if !contains(string(result), HookMatcher) {
		t.Errorf("result missing our matcher %q", HookMatcher)
	}
}

func TestSaveSettings_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	data := []byte(`{"test": true}`)

	if err := SaveSettings(path, data); err != nil {
		t.Fatalf("SaveSettings() error = %v", err)
	}

	// verify file exists w/ expected content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if string(content) != string(data) {
		t.Errorf("content = %q, want %q", string(content), string(data))
	}

	// verify no temp files left
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() != "settings.json" {
			t.Errorf("unexpected file left behind: %s", entry.Name())
		}
	}
}

func TestSetup_CreatesAllFiles(t *testing.T) {
	dir := t.TempDir()

	result, err := Setup(dir)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	// verify hook script exists
	hookPath := filepath.Join(dir, HookScriptPath)
	if _, err := os.Stat(hookPath); err != nil {
		t.Errorf("hook script not created: %v", err)
	}

	// verify settings.json exists
	settingsPath := filepath.Join(dir, SettingsPath)
	if _, err := os.Stat(settingsPath); err != nil {
		t.Errorf("settings.json not created: %v", err)
	}

	// verify rules md file exists
	rulesPath := filepath.Join(dir, RulesPath)
	if _, err := os.Stat(rulesPath); err != nil {
		t.Errorf("rules file not created: %v", err)
	}

	// verify result has correct paths
	if result.HookScriptPath != hookPath {
		t.Errorf("HookScriptPath = %q, want %q", result.HookScriptPath, hookPath)
	}
	if len(result.FilesCreated) < 3 {
		t.Errorf("FilesCreated has %d items, want at least 3", len(result.FilesCreated))
	}
}

func TestSetup_HookScriptExecutable(t *testing.T) {
	dir := t.TempDir()

	_, err := Setup(dir)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	hookPath := filepath.Join(dir, HookScriptPath)
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("failed to stat hook script: %v", err)
	}

	// check executable bit
	mode := info.Mode()
	if mode&0111 == 0 {
		t.Errorf("hook script not executable: mode = %o", mode)
	}
}

func TestSetup_AppendsToClaudeMD(t *testing.T) {
	dir := t.TempDir()

	// create existing CLAUDE.md
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(claudeMD, []byte("# Existing\n\nSome content.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Setup(dir)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	if !result.ClaudeMDAppended {
		t.Error("ClaudeMDAppended = false, want true")
	}

	content, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(string(content), "# Existing") {
		t.Error("original content lost")
	}
	if !contains(string(content), "## SpecSync") {
		t.Error("SpecSync section not appended")
	}
}

func TestSetup_IdempotentClaudeMD(t *testing.T) {
	dir := t.TempDir()

	// run setup twice
	_, err := Setup(dir)
	if err != nil {
		t.Fatalf("Setup() first call error = %v", err)
	}

	result, err := Setup(dir)
	if err != nil {
		t.Fatalf("Setup() second call error = %v", err)
	}

	// second call should not append again
	if result.ClaudeMDAppended {
		t.Error("ClaudeMDAppended = true on second call, section should already exist")
	}

	// verify only one SpecSync section
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	content, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatal(err)
	}

	// count occurrences of "## SpecSync"
	count := 0
	for i := 0; i <= len(content)-11; i++ {
		if string(content[i:i+11]) == "## SpecSync" {
			count++
		}
	}
	if count > 1 {
		t.Errorf("found %d SpecSync sections, want 1", count)
	}
}
