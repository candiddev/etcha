package main

import (
	"context"
	"errors"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var local = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"mode [change,remove]",
		"pattern path",
	},
	ArgumentsOptional: []string{
		"source name, default: local",
	},
	Run: func(ctx context.Context, args []string, c *config.Config) errs.Err {
		mode, err := parseMode(args[1], false)
		if err != nil {
			return logger.Error(ctx, err)
		}

		if mode == metrics.CommandModeCheck {
			return logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("check is not a supported mode, use a checkOnly source")))
		}

		source := "local"
		if len(args) == 4 {
			source = args[3]
		}

		p, err := pattern.ParsePatternFromPath(ctx, c, source, args[2])
		if err != nil {
			return err
		}

		_, runVars, err := p.BuildRun(ctx, c)
		if err != nil {
			return logger.Error(ctx, err)
		}

		p, err = pattern.ParsePatternFromImports(ctx, c, source, p.Imports, runVars)
		if err != nil {
			return err
		}

		s := c.Sources[source]

		_, err = p.Run.Run(ctx, c.CLI, nil, p.RunExec, s != nil && s.CheckOnly, mode == "remove")

		return err
	},
	Usage: "Run commands in a pattern locally",
}
