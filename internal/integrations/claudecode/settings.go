package claudecode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// hook represents a single hook handler
type hook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// matcherGroup represents a matcher with its hooks
type matcherGroup struct {
	Matcher string `json:"matcher"`
	Hooks   []hook `json:"hooks"`
}

// settings represent .claude/settings.json structure
type settings struct {
	Hooks       map[string][]matcherGroup `json:"hooks,omitempty"`
	Permissions map[string]any            `json:"permissions,omitempty"`
	// preserve any other fields
	Other map[string]any `json:"-"`
}

// MergeHooks adds the specsync PostToolUse hook to existing settings
func MergeHooks(existing []byte) ([]byte, error) {
	var s settings

	if len(existing) > 0 {
		// unmarshal to get known fields
		if err := json.Unmarshal(existing, &s); err != nil {
			return nil, fmt.Errorf("invalid settings JSON: %w", err)
		}
		// unmarshal to map to preserve unknown fields
		var raw map[string]any
		if err := json.Unmarshal(existing, &raw); err != nil {
			return nil, fmt.Errorf("invalid settings JSON: %w", err)
		}
		s.Other = raw
	}

	// ensure hooks map exists
	if s.Hooks == nil {
		s.Hooks = make(map[string][]matcherGroup)
	}

	// create our hook
	specsyncHook := matcherGroup{
		Matcher: HookMatcher,
		Hooks: []hook{
			{
				Type:    "command",
				Command: `"$CLAUDE_PROJECT_DIR"/.claude/hooks/specsync-push.sh`,
			},
		},
	}

	// add/append to PostToolUse
	s.Hooks["PostToolUse"] = append(s.Hooks["PostToolUse"], specsyncHook)

	// build output, w/ care to preserve other fields
	output := make(map[string]any)
	for k, v := range s.Other {
		output[k] = v
	}
	output["hooks"] = s.Hooks
	if s.Permissions != nil {
		output["permissions"] = s.Permissions
	}

	return json.MarshalIndent(output, "", "  ")
}

// SaveSettings writes settings.json atomically via temp file + rename
func SaveSettings(path string, data []byte) error {
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, "settings.json.tmp*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	success := false
	defer func() {
		if !success {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write settings: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, 0644); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	success = true
	return nil
}
