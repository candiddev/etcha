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
	Flags: cli.Flags{
		"o": {
			Usage: "Run once",
		},
	},
	Run: func(ctx context.Context, args []string, flags cli.Flags, c *config.Config) errs.Err {
		_, once := flags.Value("o")

		if c.CLI.LogFormat == "" {
			c.CLI.LogFormat = logger.FormatKV
			ctx = logger.SetFormat(ctx, logger.FormatKV)
		}

		return run.Run(ctx, c, once)
	},
	Usage: "Run Etcha in listening mode, periodically pulling new patterns or receiving new patterns via push",
}
