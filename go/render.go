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

var render = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"jwt or pattern path",
	},
	Run: func(ctx context.Context, args []string, c *config.Config) errs.Err {
		var p *pattern.Pattern

		// Try JWT
		j, err := pattern.ParseJWTFromPath(logger.SetLevel(ctx, logger.LevelNone), c, "", args[1])
		if err != nil && (j == nil || j.EtchaPattern == nil) {
			p, err = pattern.ParsePatternFromPath(ctx, c, "", args[1])
		} else {
			var e errs.Err

			p, e = j.Pattern(ctx, c, "")
			if e != nil {
				return e
			}
		}

		if p != nil {
			logger.Raw(types.JSONToString(p) + "\n")
		}

		return err
	},
	Usage: "Show the rendered pattern of a JWT or pattern file",
}
