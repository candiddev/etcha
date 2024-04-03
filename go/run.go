package main

import (
	"context"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/run"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
)

var runCmd = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	Flags: cli.Flags{
		"o": {
			Usage: "Run once",
		},
	},
	Run: func(ctx context.Context, _ []string, flags cli.Flags, c *config.Config) errs.Err {
		_, once := flags.Value("o")

		return run.Run(ctx, c, once)
	},
	Usage: "Run Etcha in listening mode, periodically pulling new patterns or receiving new patterns via push",
}
