package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()

	// ProjectID must be provided or generated, empty by default
	if cfg.ProjectID != "" {
		t.Errorf("expected empty ProjectID, got %q", cfg.ProjectID)
	}

	// GitHubRepo is optional, empty by default
	if cfg.GitHubRepo != "" {
		t.Errorf("expected empty GitHubRepo, got %q", cfg.GitHubRepo)
	}

	if len(cfg.Include) != len(DefaultInclude) {
		t.Fatalf("expected %d include patterns, got %d", len(DefaultInclude), len(cfg.Include))
	}
	for i, v := range DefaultInclude {
		if cfg.Include[i] != v {
			t.Errorf("include[%d]: expected %q, got %q", i, v, cfg.Include[i])
		}
	}

	if len(cfg.Exclude) != 0 {
		t.Errorf("expected empty exclude, got %v", cfg.Exclude)
	}

}

func TestConfigYAMLTags(t *testing.T) {
	// verify struct can be created with expected field names
	cfg := Config{
		ProjectID:  "test-uuid",
		GitHubRepo: "torvalds/linux",
		Include:    []string{"*.md"},
		Exclude:    []string{"temp/*"},
		APIURL:     "http://app.slack.com/events",
	}

	if cfg.ProjectID != "test-uuid" {
		t.Errorf("ProjectID mismatch")
	}
	if cfg.GitHubRepo != "torvalds/linux" {
		t.Errorf("GitHubRepo mismatch")
	}
}

func TestLoadConfig_CompleteFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".specsync.yaml")

	content := `project_id: "550e8400-e29b-41d4-a716-446655404550"
github_repo: "torvalds/linux"
include:
  - "custom/**/*.md"
exclude:
  - "draft/*"
api_url: "http://app.slack.com/events"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ProjectID != "550e8400-e29b-41d4-a716-446655404550" {
		t.Errorf("ProjectID = %q, want %q", cfg.ProjectID, "550e8400-e29b-41d4-a716-446655404550")
	}
	if cfg.GitHubRepo != "torvalds/linux" {
		t.Errorf("GitHubRepo = %q, want %q", cfg.GitHubRepo, "torvalds/linux")
	}
	if len(cfg.Include) != 1 || cfg.Include[0] != "custom/**/*.md" {
		t.Errorf("Include = %v, want [custom/**/*.md]", cfg.Include)
	}
	if len(cfg.Exclude) != 1 || cfg.Exclude[0] != "draft/*" {
		t.Errorf("Exclude = %v, want [draft/*]", cfg.Exclude)
	}
	if cfg.APIURL != "http://app.slack.com/events" {
		t.Errorf("APIURL = %q, want %q", cfg.APIURL, "http://app.slack.com/events")
	}
}

func TestLoadConfig_DefaultsApplied(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".specsync.yaml")

	// only project_id provided
	content := `project_id: "550e8400-e29b-41d4-a716-446655404550"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// project_id should be from file
	if cfg.ProjectID != "550e8400-e29b-41d4-a716-446655404550" {
		t.Errorf("ProjectID = %q, want from file", cfg.ProjectID)
	}

	// optinal fields should have defaults
	if len(cfg.Include) != len(DefaultInclude) {
		t.Errorf("Include = %v, want defaults %v", cfg.Include, DefaultInclude)
	}
	if len(cfg.Exclude) != 0 {
		t.Errorf("Exclude = %v, want empty", cfg.Exclude)
	}

}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/.specsync.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "specsync init") {
		t.Errorf("error should mention 'specsync init', got: %v", err)
	}
}

func TestLoadConfig_MissingProjectID(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".specsync.yaml")

	content := `github_repo: "owner/repo"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for missing project_id")
	}
	if !strings.Contains(err.Error(), "project_id") {
		t.Errorf("error should mention 'project_id', got: %v", err)
	}
}

func TestLoadConfig_InvalidUUID(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".specsync.yaml")

	content := `project_id: "def-not_a^valid-uuid"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
	if !strings.Contains(err.Error(), "UUID") {
		t.Errorf("error should mention 'UUID', got: %v", err)
	}
}

func TestLoadConfig_FileTooLarge(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".specsync.yaml")

	// create a file larger than 1MB
	largeContent := make([]byte, 1024*1024+7)
	for i := range largeContent {
		largeContent[i] = 'x'
	}
	if err := os.WriteFile(configPath, largeContent, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected error for file too large")
	}
	if !strings.Contains(err.Error(), "size") {
		t.Errorf("error should mention 'size', got: %v", err)
	}
}

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".specsync.yaml")

	cfg := &Config{
		ProjectID:  "550e8400-e29b-41d4-a716-446655042080",
		GitHubRepo: "myorg/myrepo",
		Include:    []string{"docs/**/*.md"},
		Exclude:    []string{"temp/*"},
		APIURL:     "http://localhost:8000/api/v1",
	}

	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("failed to stat saved file: %v", err)
	}
	// check perms (0644 = rw-r--r--)
	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("file permissions = %o, want 0644", perm)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if loaded.ProjectID != cfg.ProjectID {
		t.Errorf("ProjectID = %q, want %q", loaded.ProjectID, cfg.ProjectID)
	}
	if loaded.GitHubRepo != cfg.GitHubRepo {
		t.Errorf("GitHubRepo = %q, want %q", loaded.GitHubRepo, cfg.GitHubRepo)
	}
	if len(loaded.Include) != 1 || loaded.Include[0] != "docs/**/*.md" {
		t.Errorf("Include = %v, want [docs/**/*.md]", loaded.Include)
	}
}

func TestSaveConfig_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".specsync.yaml")

	cfg := &Config{
		ProjectID: "550e8400-e29b-41d4-a716-446655042080",
		Include:   DefaultInclude,
		Exclude:   []string{},
		APIURL:    DefaultAPIURL,
	}

	// save should not scatter temp files about
	if err := Save(cfg, configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// no temp files left in directory?
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".specsync.yaml.tmp") {
			t.Errorf("temp file left behind: %s", entry.Name())
		}
	}
}

func TestMergePatterns(t *testing.T) {
	tests := []struct {
		name     string
		defaults []string
		custom   []string
		want     []string
	}{
		{
			name:     "empty custom appends nothing",
			defaults: []string{"docs/**/*.md", "specs/**/*.md"},
			custom:   []string{},
			want:     []string{"docs/**/*.md", "specs/**/*.md"},
		},
		{
			name:     "new patterns appended",
			defaults: []string{"docs/**/*.md"},
			custom:   []string{"placebo_payload/**/*.md", "other/*.md"},
			want:     []string{"docs/**/*.md", "placebo_payload/**/*.md", "other/*.md"},
		},
		{
			name:     "duplicate custom patterns are removed/deduped",
			defaults: []string{"docs/**/*.md", "specs/**/*.md"},
			custom:   []string{"docs/**/*.md", "new_and_different/**/*.md"},
			want:     []string{"docs/**/*.md", "specs/**/*.md", "new_and_different/**/*.md"},
		},
		{
			name:     "all duplicates",
			defaults: []string{"docs/**/*.md"},
			custom:   []string{"docs/**/*.md"},
			want:     []string{"docs/**/*.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergePatterns(tt.defaults, tt.custom)
			if len(got) != len(tt.want) {
				t.Fatalf("MergePatterns() = %v, want %v", got, tt.want)
			}
			for i, v := range tt.want {
				if got[i] != v {
					t.Errorf("MergePatterns()[%d] = %q, want %q", i, got[i], v)
				}
			}
		})
	}
}
