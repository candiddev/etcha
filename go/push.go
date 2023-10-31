package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/run"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

func push(ctx context.Context, args []string, c *config.Config) errs.Err {
	var cmd string

	var path string

	if args[0] == "push-command" {
		cmd = args[1]
	} else {
		path = args[1]
	}

	r, err := run.Push(ctx, c, args[2], cmd, path)
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
}
