package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

// MatchesPattern checks if path matches any of the glob patterns
func MatchesPattern(path string, patterns []string) bool {
	path = filepath.Clean(path) // normalize "./docs/file.md" → "docs/file.md"
	for _, pattern := range patterns {
		matched, err := doublestar.Match(pattern, path)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// FindMostRecent identifies the most recently modified matching file
func FindMostRecent(patterns, excludes []string) (string, error) {
	var mostRecent string
	var mostRecentTime time.Time

	// walk directory and update mostRecent, mostRecentTime
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if d.IsDir() {
			return nil
		}

		// check if matches include patterns
		if !MatchesPattern(path, patterns) {
			return nil
		}

		// check if excluded
		if MatchesPattern(path, excludes) {
			return nil
		}

		// get file info for mtime
		info, err := d.Info()
		if err != nil {
			return nil
		}

		// update most recent if this file is newer
		if mostRecent == "" || info.ModTime().After(mostRecentTime) {
			mostRecent = path
			mostRecentTime = info.ModTime()
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if mostRecent == "" {
		return "", errors.New("no files match include patterns")
	}

	return mostRecent, nil
}

// ListMatching returns all files matching patterns
func ListMatching(patterns, excludes []string) ([]string, error) {
	var matches []string

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		if !MatchesPattern(path, patterns) {
			return nil
		}
		if MatchesPattern(path, excludes) {
			return nil
		}

		matches = append(matches, path)
		return nil
	})

	return matches, err
}

// FileExists checks if path exists and is a file
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
