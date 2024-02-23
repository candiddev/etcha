package main

import (
	"context"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

var jwt = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"jwt path",
	},
	Run: func(ctx context.Context, args []string, _ cli.Flags, c *config.Config) errs.Err {
		j, err := pattern.ParseJWTFromPath(ctx, c, "", args[1])

		if j != nil {
			logger.Raw(types.JSONToString(j) + "\n")
		}

		return err
	},
	Usage: "Show the contents of a JWT",
}
