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
	r, err := run.Push(ctx, c, args[2], args[1])
	if r == nil && err != nil {
		return err
	}

	if r == nil {
		return nil
	}

	if len(r.Changed) != 0 {
		logger.Info(ctx, fmt.Sprintf("Changed: %s", strings.Join(r.Changed, ", ")))
	}

	if len(r.Removed) != 0 {
		logger.Info(ctx, fmt.Sprintf("Removed: %s", strings.Join(r.Removed, ", ")))
	}

	return err
}
