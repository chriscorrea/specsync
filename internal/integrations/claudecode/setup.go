package claudecode

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Result contains paths to all files created/modified during setup
type Result struct {
	HookScriptPath   string
	SettingsPath     string
	RulesPath        string
	ClaudeMDAppended bool
	FilesCreated     []string
}

// Setup creates all Claude Code integration files in projectRoot
func Setup(projectRoot string) (*Result, error) {
	result := &Result{
		FilesCreated: make([]string, 0),
	}

	// create directories
	hooksDir := filepath.Join(projectRoot, ".claude", "hooks")
	rulesDir := filepath.Join(projectRoot, ".claude", "rules")

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create hooks directory: %w", err)
	}
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create rules directory: %w", err)
	}

	// write hook script
	hookPath := filepath.Join(projectRoot, HookScriptPath)
	if err := os.WriteFile(hookPath, []byte(HookScript), 0755); err != nil {
		return nil, fmt.Errorf("failed to write hook script: %w", err)
	}
	result.HookScriptPath = hookPath
	result.FilesCreated = append(result.FilesCreated, hookPath)

	// read/merge/write settings.json
	settingsPath := filepath.Join(projectRoot, SettingsPath)
	var existing []byte
	if data, err := os.ReadFile(settingsPath); err == nil {
		existing = data
	}

	merged, err := MergeHooks(existing)
	if err != nil {
		return nil, fmt.Errorf("failed to merge settings: %w", err)
	}

	if err := SaveSettings(settingsPath, merged); err != nil {
		return nil, fmt.Errorf("failed to save settings: %w", err)
	}
	result.SettingsPath = settingsPath
	result.FilesCreated = append(result.FilesCreated, settingsPath)

	// write rules file
	rulesPath := filepath.Join(projectRoot, RulesPath)
	if err := os.WriteFile(rulesPath, []byte(RulesContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write rules file: %w", err)
	}
	result.RulesPath = rulesPath
	result.FilesCreated = append(result.FilesCreated, rulesPath)

	// append to CLAUDE.md (only if missing or section doesn't exist)
	claudeMDPath := filepath.Join(projectRoot, "CLAUDE.md")
	claudeMDContent, err := os.ReadFile(claudeMDPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read CLAUDE.md: %w", err)
	}

	// check if section already exists in file
	if !strings.Contains(string(claudeMDContent), "## SpecSync") {
		// append it
		newContent := string(claudeMDContent) + ClaudeMDSection
		if err := os.WriteFile(claudeMDPath, []byte(newContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write CLAUDE.md: %w", err)
		}
		result.ClaudeMDAppended = true
		result.FilesCreated = append(result.FilesCreated, claudeMDPath)
	}

	return result, nil
}
