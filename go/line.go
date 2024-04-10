package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"regexp"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var line = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"mode [check,change]",
		"path or - for stdin",
		"match regexp",
		"replacement text",
	},
	Run: func(ctx context.Context, args []string, _ cli.Flags, _ *config.Config) errs.Err {
		mode, e := parseMode(args[1])
		if e != nil {
			return logger.Error(ctx, e)
		}

		if mode == metrics.CommandModeRemove {
			return logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("remove is not supported, use change")))
		}

		var content []byte

		var err error

		if args[2] == "-" {
			content = logger.ReadStdin()
		} else {
			content, err = os.ReadFile(args[2])
			if err != nil {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
			}
		}

		reg := args[3]
		rep := args[4]

		switch mode {
		case metrics.CommandModeChange:
			r, err := regexp.Compile(reg)
			if err != nil {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
			}

			o := r.ReplaceAllLiteral(content, []byte(rep))
			if !bytes.Contains(o, []byte(rep)) {
				o = append(append(o, append([]byte("\n"), rep...)...), []byte("\n")...)
			}

			if args[2] == "-" {
				logger.Raw(string(o))
			} else if !bytes.Equal(o, content) {
				if err := os.WriteFile(args[2], o, 0644); err != nil { //nolint:gosec
					return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
				}
			}
		case metrics.CommandModeCheck:
			if !bytes.Contains(content, []byte(rep)) {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("replacement text not found")))
			}
		case metrics.CommandModeRemove:
		}

		return nil
	},
	Usage: "Manage a line in text using regex.  If path is -, the contents will be read from stdin and output to stdout.  Otherwise the path specified will be checked/changed/removed with the line.",
}
