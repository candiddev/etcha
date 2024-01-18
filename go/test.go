package main

import (
	"context"
	"errors"
	"strings"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var errTestFailure = errors.New("test failure")

var test = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"path",
	},
	ArgumentsOptional: []string{
		"test build commands, default: no",
	},
	Run: func(ctx context.Context, args []string, c *config.Config) errs.Err {
		l, err := pattern.Test(ctx, c, args[1], len(args) == 3)
		if err != nil {
			return err
		}

		if len(l) > 0 {
			return logger.Error(ctx, errs.ErrReceiver.Wrap(errTestFailure), strings.Join(l.Show(), "\n"))
		}

		return nil
	},
	Usage: "Test all patterns in path",
}
