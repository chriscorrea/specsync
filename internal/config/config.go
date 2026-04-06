package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

const MaxConfigSize = 1 << 20 // 1MB limit
const ConfigFileName = ".specsync.yaml"

var ErrConfigNotFound = errors.New("config file not found")

type Config struct {
	ProjectID  string   `yaml:"project_id"`
	GitHubRepo string   `yaml:"github_repo,omitempty"`
	Include    []string `yaml:"include,omitempty"`
	Exclude    []string `yaml:"exclude,omitempty"`
	APIURL     string   `yaml:"api_url,omitempty"`
}

var DefaultInclude = []string{
	"specs/**/*.md",
	"specs/*.md",
	"docs/**/*.md",
	"docs/*.md",
	".kiro/specs/**/*.md",
	".speckit/**/*.md",
	".cursor/specs/**/*.md",
	".claude/specs/**/*.md",
}

const DefaultAPIURL = "http://localhost:8000/api/v1"
const SpecPressAPIURL = "https://spec.press/api/v1"

func Defaults() *Config {
	return &Config{
		ProjectID:  "",
		GitHubRepo: "",
		Include:    append([]string{}, DefaultInclude...),
		Exclude:    []string{},
		APIURL:     DefaultAPIURL,
	}
}

func Load(path string) (*Config, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("No %s found. Run `specsync init` to create one.", ConfigFileName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat config file: %w", err)
	}

	// enforce size limit (A08)
	if info.Size() > MaxConfigSize {
		return nil, fmt.Errorf("config file size exceeds maximum allowed (1MB)")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML in config file: %w", err)
	}

	if cfg.ProjectID == "" {
		return nil, errors.New("project_id is required in config file")
	}
	if _, err := uuid.Parse(cfg.ProjectID); err != nil {
		return nil, fmt.Errorf("project_id must be a valid UUID: %w", err)
	}

	// apply defaults
	if len(cfg.Include) == 0 {
		cfg.Include = append([]string{}, DefaultInclude...)
	}
	if cfg.Exclude == nil {
		cfg.Exclude = []string{}
	}
	if cfg.APIURL == "" {
		cfg.APIURL = DefaultAPIURL
	}

	return &cfg, nil
}

// Save writes config atomically via temp file, rename
func Save(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".specsync.yaml.tmp*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// cleanup on error
	success := false
	defer func() {
		if !success {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write config: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, 0644); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
	}

	success = true
	return nil
}

// MergePatterns deduplicates, merges default, custom patterns
func MergePatterns(defaults, custom []string) []string {
	seen := make(map[string]bool, len(defaults))
	for _, p := range defaults {
		seen[p] = true
	}
	merged := append([]string{}, defaults...)
	for _, p := range custom {
		if !seen[p] {
			seen[p] = true
			merged = append(merged, p)
		}
	}
	return merged
}
