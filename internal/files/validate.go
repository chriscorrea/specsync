package files

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateFilePath ensures filePath resolves to within projectRoot
func ValidateFilePath(filePath, projectRoot string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}

	// resolve project root to absolute
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("invalid project root: %w", err)
	}

	// resolve file path relative to project root
	var absPath string
	if filepath.IsAbs(filePath) {
		absPath = filePath
	} else {
		absPath = filepath.Join(absRoot, filePath)
	}

	// clean the path
	absPath = filepath.Clean(absPath)

	// evaluate symlinks to get real path
	realPath, err := filepath.EvalSymlinks(absPath)
	if err == nil {
		absPath = realPath
	}
	// if file doesn't exist or symlink eval fails, use cleaned path for validation

	// also resolve symlinks in project root for comparison
	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err == nil {
		absRoot = realRoot
	}

	// ensure path is within project root
	if !strings.HasPrefix(absPath, absRoot+string(filepath.Separator)) && absPath != absRoot {
		return fmt.Errorf("file path is outside project directory")
	}

	return nil
}
