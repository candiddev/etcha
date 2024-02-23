package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/run"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var push = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"destination URL",
		"command or pattern path",
	},
	Run: func(ctx context.Context, args []string, _ cli.Flags, c *config.Config) errs.Err {
		var cmd string

		var path string

		if strings.HasSuffix(args[2], ".jsonnet") {
			path = args[2]
		} else {
			cmd = args[2]
		}

		r, err := run.Push(ctx, c, args[1], cmd, path)
		if r == nil && err != nil {
			return err
		}

		if r == nil {
			return nil
		}

		if cmd != "" {
			logger.Raw(r.ChangedOutputs...)
			logger.Raw("\n")

			return nil
		}

		if len(r.ChangedIDs) != 0 {
			logger.Info(ctx, fmt.Sprintf("Changed: %s", strings.Join(r.ChangedIDs, ", ")))
		}

		if len(r.RemovedIDs) != 0 {
			logger.Info(ctx, fmt.Sprintf("Removed: %s", strings.Join(r.RemovedIDs, ", ")))
		}

		return err
	},
	Usage: "Push signed commands or patterns to a destination URL.",
}
