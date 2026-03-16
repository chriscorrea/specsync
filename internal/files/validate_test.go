package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFilePath(t *testing.T) {
	// create temp project root
	projectRoot, err := os.MkdirTemp("", "specsync-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(projectRoot)

	// create docs/auth.md
	docsDir := filepath.Join(projectRoot, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}
	authFile := filepath.Join(docsDir, "auth.md")
	if err := os.WriteFile(authFile, []byte("# Auth"), 0644); err != nil {
		t.Fatalf("failed to create auth.md: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		projectRoot string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid relative path",
			filePath:    "docs/auth.md",
			projectRoot: projectRoot,
			wantErr:     false,
		},
		{
			name:        "valid dot relative path",
			filePath:    "./docs/auth.md",
			projectRoot: projectRoot,
			wantErr:     false,
		},
		{
			name:        "traversal outside project",
			filePath:    "../outside.md",
			projectRoot: projectRoot,
			wantErr:     true,
			errMsg:      "outside",
		},
		{
			name:        "traversal via nested path",
			filePath:    "docs/../../../etc/passwd",
			projectRoot: projectRoot,
			wantErr:     true,
			errMsg:      "outside",
		},
		{
			name:        "empty path",
			filePath:    "",
			projectRoot: projectRoot,
			wantErr:     true,
			errMsg:      "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.filePath, tt.projectRoot)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateFilePath(%q, %q) = nil, want error", tt.filePath, tt.projectRoot)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateFilePath(%q, %q) = %v, want nil", tt.filePath, tt.projectRoot, err)
				}
			}
		})
	}
}

func TestValidateFilePath_Symlink(t *testing.T) {
	// create temp project root
	projectRoot, err := os.MkdirTemp("", "specsync-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(projectRoot)

	// create outside directory and file
	outsideDir, err := os.MkdirTemp("", "specsync-outside-*")
	if err != nil {
		t.Fatalf("failed to create outside dir: %v", err)
	}
	defer os.RemoveAll(outsideDir)

	outsideFile := filepath.Join(outsideDir, "secret.md")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
		t.Fatalf("failed to create secret.md: %v", err)
	}

	// create symlink inside project pointing outside
	symlinkPath := filepath.Join(projectRoot, "evil-link.md")
	if err := os.Symlink(outsideFile, symlinkPath); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	err = ValidateFilePath("evil-link.md", projectRoot)
	if err == nil {
		t.Error("ValidateFilePath should reject symlink pointing outside project")
	}
}
