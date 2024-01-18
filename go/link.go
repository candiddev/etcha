package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var link = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"mode [check,change,remove]",
		"src",
		"dst",
	},
	Run: func(ctx context.Context, args []string, config *config.Config) errs.Err {
		mode, e := parseMode(args[1], true)
		if e != nil {
			return logger.Error(ctx, e)
		}

		src := args[2]
		dst := args[3]

		_, err := os.Lstat(dst)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && mode == metrics.CommandModeChange {
				if err := os.Symlink(src, dst); err != nil {
					return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
				}
			} else if mode == metrics.CommandModeRemove {
				return nil
			} else {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
			}
		}

		s, err := os.Readlink(dst)
		if err != nil {
			if mode == metrics.CommandModeRemove {
				return nil
			} else if mode == metrics.CommandModeChange { //nolint:revive
				if err := os.RemoveAll(dst); err != nil {
					return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
				}
			} else {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
			}
		}

		if mode == metrics.CommandModeRemove {
			if err := os.Remove(dst); err != nil {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
			}

			return nil
		}

		if s != src {
			if mode == metrics.CommandModeCheck {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("symlink has wrong src: %s", s)))
			}

			if err := os.Symlink(src, dst); err != nil {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
			}
		}

		return nil
	},
	Usage: "Manage symlinks for src and dst",
}
