package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/uuid"
	"github.com/urfave/cli/v3"

	"github.com/chriscorrea/specsync/internal/auth"
	"github.com/chriscorrea/specsync/internal/config"
	"github.com/chriscorrea/specsync/internal/integrations/claudecode"
	"github.com/chriscorrea/specsync/internal/specpress"
)

func initCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize a new .specsync.yaml configuration file",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output as JSON (non-interactive mode)",
			},
			&cli.BoolFlag{
				Name:  "claude-code",
				Usage: "Set up Claude Code integration (hooks for automatic spec syncing)",
			},
			&cli.BoolFlag{
				Name:  "specpress",
				Usage: "Configure for spec.press (optionally pass project ID as argument)",
			},
			&cli.StringFlag{
				Name:  "create",
				Usage: "Create new spec.press project (optional: project name)",
			},
			&cli.BoolFlag{
				Name:  "force",
				Usage: "Overwrite existing config without prompting",
			},
			&cli.StringFlag{
				Name:  "paths",
				Usage: "Additional include patterns (comma-separated), appended to defaults",
			},
		},
		Action: runInit,
	}
}

func runInit(ctx context.Context, cmd *cli.Command) error {
	jsonMode := cmd.Bool("json")
	claudeCodeFlag := cmd.Bool("claude-code")
	specpressMode := cmd.Bool("specpress")
	specpressArg := ""
	if cmd.Args().Present() {
		specpressArg = cmd.Args().First()
	}
	createArg := cmd.String("create")
	createMode := cmd.IsSet("create")
	forceMode := cmd.Bool("force")
	configPath := config.ConfigFileName
	cwd, _ := os.Getwd()

	if _, err := os.Stat(configPath); err == nil && !forceMode {
		if jsonMode {
			fmt.Fprintf(os.Stderr, `{"error": "config file already exists"}`+"\n")
			return cli.Exit("", 1)
		}

		existing, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read existing config: %w", err)
		}
		fmt.Println("Existing config:")
		fmt.Println(string(existing))

		var confirm bool
		prompt := &survey.Confirm{
			Message: "Config already exists. Overwrite?",
			Default: false,
		}
		if err := survey.AskOne(prompt, &confirm); err != nil {
			return cli.Exit("Init cancelled.", 0)
		}
		if !confirm {
			fmt.Println("Keeping existing configuration.")
			return nil
		}
	}

	var cfg *config.Config
	var err error

	pathsFlag := cmd.String("paths")

	if specpressMode || createMode {
		cfg, err = collectSpecpressConfig(jsonMode, specpressArg, createMode, createArg)
	} else {
		cfg, err = collectConfig(jsonMode, pathsFlag)
	}
	if err != nil {
		if strings.Contains(err.Error(), "interrupt") {
			return cli.Exit("Init cancelled.", 130)
		}
		return err
	}

	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Claude Code integration
	setupClaudeCode := claudeCodeFlag
	if !setupClaudeCode && !jsonMode && claudecode.DetectClaudeCode(cwd) {
		// prompt user if CC detected
		var confirm bool
		prompt := &survey.Confirm{
			Message: "Claude Code detected. Set up automatic spec syncing?",
			Default: true,
		}
		if err := survey.AskOne(prompt, &confirm); err == nil && confirm {
			setupClaudeCode = true
		}
	}

	var claudeCodeResult *claudecode.Result
	if setupClaudeCode {
		var err error
		claudeCodeResult, err = claudecode.Setup(cwd)
		if err != nil {
			if jsonMode {
				fmt.Fprintf(os.Stderr, `{"error": "claude code setup failed", "message": %q}`+"\n", err.Error())
			}
			return fmt.Errorf("claude code setup failed: %w", err)
		}
	}

	// output results
	if jsonMode {
		claudeConfigured := claudeCodeResult != nil
		claudeFiles := "[]"
		if claudeCodeResult != nil {
			claudeFiles = toJSONArray(claudeCodeResult.FilesCreated)
		}
		fmt.Printf(`{"project_id": %q, "github_repo": %q, "include": %s, "exclude": %s, "api_url": %q, "claude_code_configured": %t, "claude_code_files": %s}`+"\n",
			cfg.ProjectID,
			cfg.GitHubRepo,
			toJSONArray(cfg.Include),
			toJSONArray(cfg.Exclude),
			cfg.APIURL,
			claudeConfigured,
			claudeFiles,
		)
	} else {
		fmt.Printf("Created %s\n", configPath)
		if specpressMode || createMode {
			shortID := cfg.ProjectID
			if len(shortID) > 8 {
				shortID = shortID[:8] + "..."
			}
			fmt.Printf("Configured for spec.press project %s\n", shortID)
		}
		if claudeCodeResult != nil {
			fmt.Println("\nClaude Code integration configured:")
			for _, f := range claudeCodeResult.FilesCreated {
				fmt.Printf("  %s\n", f)
			}
		}
	}

	return nil
}

func collectConfig(jsonMode bool, pathsFlag string) (*config.Config, error) {
	cfg := config.Defaults()
	defaultUUID := uuid.New().String()
	detectedRepo := config.DetectGitHubRepo()

	if jsonMode {
		cfg.ProjectID = defaultUUID
		cfg.GitHubRepo = detectedRepo
		if pathsFlag != "" {
			cfg.Include = config.MergePatterns(cfg.Include, parsePatterns(pathsFlag))
		}
		return cfg, nil
	}

	var err error

	projectIDPrompt := &survey.Input{
		Message: "Project ID (UUID):",
		Default: defaultUUID,
	}
	if err = survey.AskOne(projectIDPrompt, &cfg.ProjectID); err != nil {
		return nil, err
	}
	if cfg.ProjectID == "" {
		cfg.ProjectID = defaultUUID
	}

	repoPrompt := &survey.Input{
		Message: "GitHub repository (owner/repo):",
		Default: detectedRepo,
	}
	if err = survey.AskOne(repoPrompt, &cfg.GitHubRepo); err != nil {
		return nil, err
	}

	// show defaults, ask if user wants to add custom paths
	if err = promptAdditionalPaths(cfg); err != nil {
		return nil, err
	}

	excludePrompt := &survey.Input{
		Message: "Exclude patterns (comma-separated):",
		Default: "",
	}
	var excludeStr string
	if err = survey.AskOne(excludePrompt, &excludeStr); err != nil {
		return nil, err
	}
	cfg.Exclude = parsePatterns(excludeStr)

	apiPrompt := &survey.Input{
		Message: "API URL:",
		Default: config.DefaultAPIURL,
	}
	if err = survey.AskOne(apiPrompt, &cfg.APIURL); err != nil {
		return nil, err
	}
	if cfg.APIURL == "" {
		cfg.APIURL = config.DefaultAPIURL
	}

	return cfg, nil
}

// collectSpecpressConfig collects config for spec.press integration
func collectSpecpressConfig(jsonMode bool, specpressArg string, createMode bool, createArg string) (*config.Config, error) {
	cfg := config.Defaults()
	cfg.APIURL = config.SpecPressAPIURL

	// check for conflicting options
	if createMode && specpressArg != "" {
		return nil, fmt.Errorf("cannot use --create with a project ID")
	}

	// --create mode: create new project
	if createMode {
		return createSpecpressProject(jsonMode, createArg, cfg)
	}

	// if project ID/URL provided, extract it
	if specpressArg != "" {
		projectID, err := specpress.ExtractProjectID(specpressArg)
		if err != nil {
			return nil, fmt.Errorf("invalid project ID: %w", err)
		}
		cfg.ProjectID = projectID

		// still need include patterns
		if !jsonMode {
			if err := promptIncludePatterns(cfg); err != nil {
				return nil, err
			}
		}

		return cfg, nil
	}

	// no argument: prompt create vs link
	if jsonMode {
		return nil, fmt.Errorf("--specpress requires a project ID in json mode")
	}

	var choice string
	prompt := &survey.Select{
		Message: "Create new project or link existing?",
		Options: []string{"create", "link"},
		Default: "create",
	}
	if err := survey.AskOne(prompt, &choice); err != nil {
		return nil, err
	}

	if choice == "link" {
		// prompt for project ID
		var projectIDInput string
		idPrompt := &survey.Input{
			Message: "Project ID or URL:",
		}
		if err := survey.AskOne(idPrompt, &projectIDInput); err != nil {
			return nil, err
		}
		projectID, err := specpress.ExtractProjectID(projectIDInput)
		if err != nil {
			return nil, fmt.Errorf("invalid project ID: %w", err)
		}
		cfg.ProjectID = projectID

		if err := promptIncludePatterns(cfg); err != nil {
			return nil, err
		}

		return cfg, nil
	}

	// if choice is "create", use createSpecpressProject flow
	return createSpecpressProject(jsonMode, "", cfg)
}

// createSpecpressProject creates a new project on spec.press
func createSpecpressProject(jsonMode bool, projectName string, cfg *config.Config) (*config.Config, error) {
	// require auth token
	token := auth.GetToken()
	if token == "" {
		return nil, fmt.Errorf("authentication required\n\nSet your token using:\n  specsync auth login\n\nGet your token at: https://spec.press/settings")
	}

	// prompt for project name if not provided
	if projectName == "" && !jsonMode {
		namePrompt := &survey.Input{
			Message: "Project name:",
		}
		if err := survey.AskOne(namePrompt, &projectName); err != nil {
			return nil, err
		}
	}
	if projectName == "" {
		return nil, fmt.Errorf("project name is required")
	}

	// create project via API
	client := specpress.NewClient(token)
	projectID, err := client.CreateProject(projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	cfg.ProjectID = projectID
	fmt.Printf("Project created: %s (%s)\n", projectName, projectID)

	// collect include patterns
	if !jsonMode {
		if err := promptIncludePatterns(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// promptAdditionalPaths displays defaults, asks Y/N to add custom paths
func promptAdditionalPaths(cfg *config.Config) error {
	//fmt.Println("\nDefault document locations:")
	//for _, p := range config.DefaultInclude {
	//	fmt.Printf("  %s\n", p)
	//}
	//fmt.Println()

	fmt.Println("Specsync will look for documents in common locations by default.")
	var addCustom bool
	prompt := &survey.Confirm{
		Message: "Do you want to specify your doc locations?",
		Default: false,
	}
	if err := survey.AskOne(prompt, &addCustom); err != nil {
		return err
	}

	if addCustom {
		var customStr string
		input := &survey.Input{
			Message: "Additional patterns (comma-separated):",
		}
		if err := survey.AskOne(input, &customStr); err != nil {
			return err
		}
		cfg.Include = config.MergePatterns(cfg.Include, parsePatterns(customStr))
	}

	return nil
}

// promptIncludePatterns prompts for include/exclude patterns (specpress flow)
func promptIncludePatterns(cfg *config.Config) error {
	if err := promptAdditionalPaths(cfg); err != nil {
		return err
	}

	excludePrompt := &survey.Input{
		Message: "Exclude patterns (comma-separated):",
		Default: "",
	}
	var excludeStr string
	if err := survey.AskOne(excludePrompt, &excludeStr); err != nil {
		return err
	}
	cfg.Exclude = parsePatterns(excludeStr)

	return nil
}

func parsePatterns(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func toJSONArray(ss []string) string {
	if len(ss) == 0 {
		return "[]"
	}
	quoted := make([]string, len(ss))
	for i, s := range ss {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
