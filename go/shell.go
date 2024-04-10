package main

import (
	"context"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/run"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var shell = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"source name",
		"target name or host address",
	},
	Flags: cli.Flags{
		"p": {
			Default:     []string{"4000"},
			Placeholder: "port",
			Usage:       "Push to a specific host port",
		},
		"u": {
			Default:     []string{"/etcha/v1/shell"},
			Placeholder: "path",
			Usage:       "Push to a specific host path",
		},
	},
	Run: func(ctx context.Context, args []string, flags cli.Flags, c *config.Config) errs.Err {
		source := args[1]
		target := c.Targets[args[2]]

		target.Port, _ = flags.Value("p")

		if u, ok := flags.Value("u"); ok {
			target.PathShell = u
		}

		return logger.Error(ctx, run.Shell(ctx, c, target, source))
	},
	Usage: "Open an interactive shell for the Target and Source.  Shell will use signed commands pushed to URL and receive responses using Server-Sent Events (SSE).  Must specify a Source.  Can specify a Target name from he configuration, or individual target details.",
}
