package main

import (
	"context"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/etcha/go/run"
	"github.com/candiddev/shared/go/errs"
)

func runCommands(ctx context.Context, args []string, c *config.Config) errs.Err {
	if args[0] == "check" || args[0] == "run-commands" || args[0] == "remove" {
		source := "etcha"
		if len(args) == 3 {
			source = args[2]
		}

		p, err := pattern.ParsePatternFromPath(ctx, c, source, args[1])
		if err != nil {
			return err
		}

		m := commands.ModeChange

		switch args[0] {
		case "check":
			m = commands.ModeChange
		case "remove":
			m = commands.ModeRemove
		}

		_, err = p.Run.Run(ctx, c.CLI, p.RunEnv, p.Exec, m)

		return err
	}

	once := false

	if args[0] == "run-once" {
		once = true
	}

	return run.Run(ctx, c, once)
}
