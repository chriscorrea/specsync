package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	"github.com/urfave/cli/v3"

	"github.com/chriscorrea/specsync/internal/auth"
	"github.com/chriscorrea/specsync/internal/config"
	"github.com/chriscorrea/specsync/internal/doctype"
	"github.com/chriscorrea/specsync/internal/files"
	"github.com/chriscorrea/specsync/internal/push"
)

func pushCommand() *cli.Command {
	return &cli.Command{
		Name:      "push",
		Usage:     "Push a file to the remote endpoint",
		ArgsUsage: "[file]",
		Action:    runPush,
	}
}

type pushOutput struct {
	Path    string `json:"path"`
	Hash    string `json:"hash"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func runPush(ctx context.Context, cmd *cli.Command) error {
	jsonMode := cmd.Root().Bool("json")
	dryRun := cmd.Root().Bool("dry-run")
	silent := cmd.Root().Bool("silent")

	// load config
	cfg, err := config.Load(config.ConfigFileName)
	if err != nil {
		if jsonMode {
			outputJSON(pushOutput{Error: "config error", Message: err.Error()})
		}
		return cli.Exit(err.Error(), 2)
	}

	// resolve file to push
	var filePath string
	if cmd.NArg() > 0 {
		filePath = cmd.Args().First()
	} else {
		// find most recent
		filePath, err = files.FindMostRecent(cfg.Include, cfg.Exclude)
		if err != nil {
			if jsonMode {
				outputJSON(pushOutput{Error: "no files", Message: err.Error()})
			}
			return cli.Exit(err.Error(), 1)
		}
	}

	// validate path stays within project
	cwd, _ := os.Getwd()
	if err := files.ValidateFilePath(filePath, cwd); err != nil {
		if jsonMode {
			outputJSON(pushOutput{Path: filePath, Error: "path error", Message: err.Error()})
		}
		return cli.Exit(fmt.Sprintf("invalid path: %v", err), 1)
	}

	// check file matches include patterns
	if !files.MatchesPattern(filePath, cfg.Include) {
		msg := fmt.Sprintf("file not in include patterns: %s", filePath)
		if jsonMode {
			outputJSON(pushOutput{Path: filePath, Error: "not matched", Message: msg})
		}
		return cli.Exit(msg, 1)
	}

	// check not excluded
	if files.MatchesPattern(filePath, cfg.Exclude) {
		msg := fmt.Sprintf("file is excluded: %s", filePath)
		if jsonMode {
			outputJSON(pushOutput{Path: filePath, Error: "excluded", Message: msg})
		}
		return cli.Exit(msg, 1)
	}

	// read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if jsonMode {
			outputJSON(pushOutput{Path: filePath, Error: "read error", Message: err.Error()})
		}
		return cli.Exit(fmt.Sprintf("failed to read file: %v", err), 1)
	}

	// validate UTF-8
	if !utf8.Valid(content) {
		msg := "file is not valid UTF-8"
		if jsonMode {
			outputJSON(pushOutput{Path: filePath, Error: "encoding error", Message: msg})
		}
		return cli.Exit(msg, 1)
	}

	// compute hash
	hashBytes := sha256.Sum256(content)
	hash := "sha256:" + hex.EncodeToString(hashBytes[:])

	// infer type
	docType := doctype.InferType(string(content))

	// dry run - just output what would happen
	if dryRun {
		out := pushOutput{
			Path:   filePath,
			Hash:   hash,
			Type:   docType,
			Status: "dry-run",
		}
		if jsonMode {
			outputJSON(out)
		} else if !silent {
			fmt.Printf("would push: %s\n", filePath)
			fmt.Printf("  type: %s\n", docType)
			fmt.Printf("  hash: %s\n", hash)
		}
		return nil
	}

	// get auth token
	token := auth.GetToken()

	// push
	req := &push.PushRequest{
		ProjectID: cfg.ProjectID,
		Path:      filePath,
		Content:   string(content),
		Type:      docType,
		Hash:      hash,
		Timestamp: time.Now(),
	}

	_, err = push.Push(ctx, cfg.APIURL, token, req)
	if err != nil {
		if jsonMode {
			outputJSON(pushOutput{Path: filePath, Hash: hash, Type: docType, Error: "push failed", Message: err.Error()})
		}
		return cli.Exit(fmt.Sprintf("push failed: %v", err), 1)
	}

	// success
	out := pushOutput{
		Path:   filePath,
		Hash:   hash,
		Type:   docType,
		Status: "pushed",
	}
	if jsonMode {
		outputJSON(out)
	} else if !silent {
		fmt.Printf("pushed: %s\n", filePath)
	}

	return nil
}

func outputJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
