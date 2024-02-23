package main

import (
	"context"
	"path/filepath"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
)

var build = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"pattern path",
		"destination path",
	},
	Run: func(ctx context.Context, args []string, _ cli.Flags, c *config.Config) errs.Err {
		source := args[1]
		destination := args[2]

		c.Vars["buildDir"] = filepath.Dir(source)
		c.Vars["buildPath"] = source

		p, err := pattern.ParsePatternFromPath(ctx, c, "", source)
		if err != nil {
			return err
		}

		return p.BuildSign(ctx, c, destination)
	},
	Usage: "Build a pattern",
}
