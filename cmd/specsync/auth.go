package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v3"

	"github.com/chriscorrea/specsync/internal/auth"
)

func authCommand() *cli.Command {
	return &cli.Command{
		Name:  "auth",
		Usage: "Manage authentication",
		Commands: []*cli.Command{
			{
				Name:      "login",
				Usage:     "Save authentication token",
				ArgsUsage: "[token]",
				Action:    runAuthLogin,
			},
			{
				Name:   "logout",
				Usage:  "Remove saved authentication token",
				Action: runAuthLogout,
			},
			{
				Name:   "status",
				Usage:  "Show authentication status",
				Action: runAuthStatus,
			},
		},
	}
}

func runAuthLogin(ctx context.Context, cmd *cli.Command) error {
	jsonMode := cmd.Root().Bool("json")

	var token string
	if cmd.NArg() > 0 {
		token = cmd.Args().First()
	} else {
		// interactive prompt
		prompt := &survey.Password{
			Message: "Enter token:",
		}
		if err := survey.AskOne(prompt, &token); err != nil {
			if strings.Contains(err.Error(), "interrupt") {
				return cli.Exit("", 130)
			}
			return cli.Exit(fmt.Sprintf("failed to read token: %v", err), 1)
		}
	}

	token = strings.TrimSpace(token)
	if token == "" {
		if jsonMode {
			fmt.Println(`{"error": "token required"}`)
		}
		return cli.Exit("token required", 2)
	}

	if err := auth.SaveToken(token); err != nil {
		if jsonMode {
			fmt.Printf(`{"error": "failed to save token", "message": %q}`+"\n", err.Error())
		}
		return cli.Exit(fmt.Sprintf("failed to save token: %v", err), 1)
	}

	if jsonMode {
		fmt.Println(`{"status": "authenticated"}`)
	} else {
		fmt.Println("Token saved to ~/.config/specsync/credentials.json")
	}

	return nil
}

func runAuthLogout(ctx context.Context, cmd *cli.Command) error {
	jsonMode := cmd.Root().Bool("json")

	if err := auth.DeleteToken(); err != nil {
		if jsonMode {
			fmt.Printf(`{"error": "failed to remove token", "message": %q}`+"\n", err.Error())
		}
		return cli.Exit(fmt.Sprintf("failed to remove token: %v", err), 1)
	}

	if jsonMode {
		fmt.Println(`{"status": "logged out"}`)
	} else {
		fmt.Println("Logged out")
	}

	return nil
}

func runAuthStatus(ctx context.Context, cmd *cli.Command) error {
	jsonMode := cmd.Root().Bool("json")

	source := auth.GetTokenSource()

	switch source {
	case auth.TokenSourceEnv:
		if jsonMode {
			fmt.Println(`{"authenticated": true, "source": "env"}`)
		} else {
			fmt.Println("authenticated via SPECSYNC_TOKEN environment variable")
		}
	case auth.TokenSourceFile:
		if jsonMode {
			fmt.Println(`{"authenticated": true, "source": "file"}`)
		} else {
			fmt.Println("authenticated via ~/.config/specsync/credentials.json")
		}
	default:
		if jsonMode {
			fmt.Println(`{"authenticated": false}`)
		} else {
			fmt.Println("not authenticated")
		}
	}

	return nil
}
