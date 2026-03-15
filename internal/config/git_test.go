package config

import (
	"testing"
)

func TestParseGitHubRepo(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
		wantErr  bool
	}{
		{
			name:     "SSH format",
			url:      "git@github.com:owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "SSH format without .git",
			url:      "git@github.com:owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "HTTPS format",
			url:      "https://github.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "HTTPS format without .git",
			url:      "https://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "owner with hyphen",
			url:      "git@github.com:my-org/my-repo.git",
			expected: "my-org/my-repo",
		},
		{
			name:     "owner with underscore",
			url:      "git@github.com:my_org/my_repo.git",
			expected: "my_org/my_repo",
		},
		{
			name:     "owner with dot",
			url:      "git@github.com:my.org/my.repo.git",
			expected: "my.org/my.repo",
		},
		{
			name:     "non-github URL",
			url:      "git@gitlab.com:owner/repo.git",
			expected: "",
		},
		{
			name:     "invalid format",
			url:      "not-a-url",
			expected: "",
		},
		{
			name:     "empty string",
			url:      "",
			expected: "",
		},
		{
			name:     "injection attempt with semicolon",
			url:      "git@github.com:owner;rm -rf/repo.git",
			expected: "",
		},
		{
			name:     "injection attempt with backticks",
			url:      "git@github.com:owner/`whoami`.git",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseGitHubRepo(tt.url)
			if got != tt.expected {
				t.Errorf("ParseGitHubRepo(%q) = %q, want %q", tt.url, got, tt.expected)
			}
		})
	}
}

func TestValidateOwnerRepo(t *testing.T) {
	tests := []struct {
		ownerRepo string
		valid     bool
	}{
		{"owner/repo", true},
		{"my-org/my-repo", true},
		{"my_org/my_repo", true},
		{"my.org/my.repo", true},
		{"Owner123/Repo456", true},
		{"", false},
		{"owner", false},
		{"owner/", false},
		{"/repo", false},
		{"owner/repo/extra", false},
		{"owner;cmd/repo", false},
		{"owner/repo`cmd`", false},
		{"owner/repo$(cmd)", false},
		{"owner repo/test", false},
	}

	for _, tt := range tests {
		t.Run(tt.ownerRepo, func(t *testing.T) {
			got := ValidateOwnerRepo(tt.ownerRepo)
			if got != tt.valid {
				t.Errorf("ValidateOwnerRepo(%q) = %v, want %v", tt.ownerRepo, got, tt.valid)
			}
		})
	}
}
