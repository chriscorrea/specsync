package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var Version = "dev" // set at build time

func main() {
	app := &cli.Command{
		Name:    "specsync",
		Usage:   "Sync markdown spec files",
		Version: Version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output as JSON",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview without actually making changes",
			},
			&cli.BoolFlag{
				Name:  "silent",
				Usage: "Suppress non-error output",
			},
		},
		Commands: []*cli.Command{
			initCommand(),
			pushCommand(),
			authCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
