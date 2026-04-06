package files

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		patterns []string
		want     bool
	}{
		{
			name:     "matches docs glob",
			path:     "docs/auth.md",
			patterns: []string{"docs/**/*.md"},
			want:     true,
		},
		{
			name:     "matches nested path",
			path:     "docs/api/v1/spec.md",
			patterns: []string{"docs/**/*.md"},
			want:     true,
		},
		{
			name:     "no match different dir",
			path:     "random/file.md",
			patterns: []string{"docs/**/*.md"},
			want:     false,
		},
		{
			name:     "no match different extension",
			path:     "docs/spec.json",
			patterns: []string{"docs/**/*.md"},
			want:     false,
		},
		{
			name:     "matches one of multiple patterns",
			path:     "specs/api.md",
			patterns: []string{"docs/**/*.md", "specs/**/*.md"},
			want:     true,
		},
		{
			name:     "matches with dot-slash prefix",
			path:     "./docs/security/security-review.md",
			patterns: []string{"docs/**/*.md"},
			want:     true,
		},
		{
			name:     "matches kiro specs",
			path:     ".kiro/specs/flow.md",
			patterns: []string{".kiro/specs/**/*.md"},
			want:     true,
		},
		{
			name:     "matches speckit",
			path:     ".speckit/design.md",
			patterns: []string{".speckit/**/*.md"},
			want:     true,
		},
		{
			name:     "matches cursor specs",
			path:     ".cursor/specs/review.md",
			patterns: []string{".cursor/specs/**/*.md"},
			want:     true,
		},
		{
			name:     "matches claude specs",
			path:     ".claude/specs/plan.md",
			patterns: []string{".claude/specs/**/*.md"},
			want:     true,
		},
		{
			name:     "matches docs/specs nested",
			path:     "docs/specs/api.md",
			patterns: []string{"docs/specs/**/*.md"},
			want:     true,
		},
		{
			name:     "matches specs top-level",
			path:     "specs/auth.md",
			patterns: []string{"specs/*.md"},
			want:     true,
		},
		{
			name:     "please don't match top-level readme",
			path:     "README.md",
			patterns: []string{"docs/**/*.md", "specs/**/*.md"},
			want:     false,
		},
		{
			name:     "do not match non-markdown in dot dir",
			path:     ".claude/settings.json",
			patterns: []string{".claude/specs/**/*.md"},
			want:     false,
		},
		{
			name:     "empty patterns no match",
			path:     "docs/spec.md",
			patterns: []string{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesPattern(tt.path, tt.patterns)
			if got != tt.want {
				t.Errorf("MatchesPattern(%q, %v) = %v, want %v", tt.path, tt.patterns, got, tt.want)
			}
		})
	}
}

func TestFindMostRecent(t *testing.T) {
	// create temp project structure
	projectRoot, err := os.MkdirTemp("", "specsync-files-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(projectRoot)

	docsDir := filepath.Join(projectRoot, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}

	oldFile := filepath.Join(docsDir, "old.md")
	newFile := filepath.Join(docsDir, "new.md")

	if err := os.WriteFile(oldFile, []byte("old"), 0644); err != nil {
		t.Fatalf("failed to create old.md: %v", err)
	}

	// mtime to an hour ago
	oldTime := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set mtime: %v", err)
	}

	// create new file (will be current mtime)
	if err := os.WriteFile(newFile, []byte("new"), 0644); err != nil {
		t.Fatalf("failed to create new.md: %v", err)
	}

	// change to proj root for relative paths
	origDir, _ := os.Getwd()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("failed to chdir to project root: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	got, err := FindMostRecent([]string{"docs/**/*.md"}, []string{})
	if err != nil {
		t.Fatalf("FindMostRecent() error = %v", err)
	}

	if got != "docs/new.md" {
		t.Errorf("FindMostRecent() = %q, want %q", got, "docs/new.md")
	}
}

func TestFindMostRecent_NoMatches(t *testing.T) {
	projectRoot, err := os.MkdirTemp("", "specsync-files-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(projectRoot)

	origDir, _ := os.Getwd()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("failed to chdir to project root: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	_, err = FindMostRecent([]string{"docs/**/*.md"}, []string{})
	if err == nil {
		t.Error("FindMostRecent() expected error for no matches")
	}
}

func TestFindMostRecent_ExcludePattern(t *testing.T) {
	projectRoot, err := os.MkdirTemp("", "specsync-files-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(projectRoot)

	docsDir := filepath.Join(projectRoot, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}

	// create included file (older)
	includedFile := filepath.Join(docsDir, "included.md")
	if err := os.WriteFile(includedFile, []byte("included"), 0644); err != nil {
		t.Fatalf("failed to create included.md: %v", err)
	}
	oldTime := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(includedFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set mtime: %v", err)
	}

	// create excluded file (newer)
	excludedFile := filepath.Join(docsDir, "excluded.md")
	if err := os.WriteFile(excludedFile, []byte("excluded"), 0644); err != nil {
		t.Fatalf("failed to create excluded.md: %v", err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("failed to chdir to project root: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	got, err := FindMostRecent([]string{"docs/**/*.md"}, []string{"**/excluded.md"})
	if err != nil {
		t.Fatalf("FindMostRecent() error = %v", err)
	}

	if got != "docs/included.md" {
		t.Errorf("FindMostRecent() = %q, want %q (excluded file should be skipped)", got, "docs/included.md")
	}
}
