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

var errLint = errors.New("found linting errors")

var lint = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"path",
	},
	ArgumentsOptional: []string{
		"check formatting, default: no",
	},
	Flags: cli.Flags{
		"f": {
			Usage: "Check formatting",
		},
	},
	Run: func(ctx context.Context, args []string, flags cli.Flags, c *config.Config) errs.Err {
		_, check := flags.Value("f")

		l, err := pattern.Lint(ctx, c, args[1], check)
		if err != nil {
			return err
		}

		if len(l) > 0 {
			return logger.Error(ctx, errs.ErrReceiver.Wrap(errLint), strings.Join(l.Show(), "\n"))
		}

		return nil
	},
	Usage: "Lint a pattern or directory",
}
