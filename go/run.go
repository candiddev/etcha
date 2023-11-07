package main

import (
	"context"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/etcha/go/run"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

func runCommands(ctx context.Context, args []string, c *config.Config) errs.Err {
	if args[0] == "local-change" || args[0] == "local-check" || args[0] == "local-remove" {
		source := "etcha"
		if len(args) == 3 {
			source = args[2]
		}

		p, err := pattern.ParsePatternFromPath(ctx, c, source, args[1])
		if err != nil {
			return err
		}

		check := false
		if s := c.Sources[source]; s != nil {
			check = s.CheckOnly
		}

		_, err = p.Run.Run(ctx, c.CLI, p.RunEnv, p.RunExec, check, args[0] == "local-remove")

		return err
	}

	once := false

	if args[0] == "run-once" {
		once = true
	} else if c.CLI.LogFormat == "" {
		c.CLI.LogFormat = logger.FormatKV
		ctx = logger.SetFormat(ctx, logger.FormatKV)
	}

	return run.Run(ctx, c, once)
}
