package main

import (
	"context"
	"errors"
	"os"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/initdir"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var initCmd = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsOptional: []string{
		"directory, default: current directory",
	},
	Run: func(ctx context.Context, args []string, _ cli.Flags, _ *config.Config) errs.Err {
		var path string

		var err error

		if len(args) == 2 {
			path = args[1]
		} else {
			path, err = os.Getwd()
			if err != nil {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error getting current directory path")))
			}
		}

		return initdir.Init(ctx, path)
	},
	Usage: "Initialize a directory for pattern development",
}
