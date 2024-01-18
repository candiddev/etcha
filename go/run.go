package main

import (
	"context"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/run"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var runCmd = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsOptional: []string{
		"run once, default: no",
	},
	Run: func(ctx context.Context, args []string, c *config.Config) errs.Err {
		once := false

		if len(args) == 3 {
			once = true
		} else if c.CLI.LogFormat == "" {
			c.CLI.LogFormat = logger.FormatKV
			ctx = logger.SetFormat(ctx, logger.FormatKV)
		}

		return run.Run(ctx, c, once)
	},
	Usage: "Run Etcha in listening mode, periodically pulling new patterns or receiving new patterns via push",
}
