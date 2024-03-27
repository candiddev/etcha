package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
)

var local = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"pattern path (ending in .jsonnet) or Jsonnet to render",
	},
	Flags: cli.Flags{
		"f": {
			Placeholder: "regexp",
			Usage:       "Filter parent Command IDs",
		},
		"r": {
			Usage: "Set mode to Remove (default: Change)",
		},
		"s": {
			Default: []string{
				"local",
			},
			Placeholder: "name",
			Usage:       "Source name to use",
		},
	},
	Run: func(ctx context.Context, args []string, flags cli.Flags, c *config.Config) errs.Err {
		_, remove := flags.Value("r")
		source, _ := flags.Value("s")

		r, _ := flags.Value("f")
		reg, e := regexp.Compile(r)
		if e != nil {
			return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error parsing filter: %w", e)))
		}

		var err errs.Err

		var p *pattern.Pattern

		if strings.HasSuffix(args[1], ".jsonnet") {
			p, err = pattern.ParsePatternFromPath(ctx, c, source, args[1])
		} else {
			var i *jsonnet.Imports

			j := jsonnet.NewRender(ctx, nil)

			i, err = j.GetString(ctx, fmt.Sprintf("{run: [%s]}", args[1]))
			if err != nil {
				return logger.Error(ctx, err)
			}

			p, err = pattern.ParsePatternFromImports(ctx, c, source, i, nil)
		}

		if err != nil {
			return logger.Error(ctx, err)
		}

		_, runVars, err := p.BuildRun(ctx, c)
		if err != nil {
			return logger.Error(ctx, err)
		}

		p, err = pattern.ParsePatternFromImports(ctx, c, source, p.Imports, runVars)
		if err != nil {
			return err
		}

		s := c.Sources[source]

		_, err = p.Run.Run(ctx, c.CLI, p.RunExec, commands.CommandsRunOpts{
			Check:          s != nil && s.CheckOnly,
			ParentID:       source,
			ParentIDFilter: reg,
			Remove:         remove,
		})

		return err
	},
	Usage: "Run commands in a pattern locally",
}
