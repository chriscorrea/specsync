package claudecode

import (
	"os"
	"path/filepath"
)

// DetectClaudeCode checks if .claude/ directory exists in projectRoot
// This is simple presence check to determine if Claude Code config is relevant
func DetectClaudeCode(projectRoot string) bool {
	// for now, just check for .claude dir; cuold check for CLAUDE.md etc in future
	claudePath := filepath.Join(projectRoot, ".claude")
	info, err := os.Stat(claudePath)
	if err != nil {
		return false
	}
	return info.IsDir()
}
