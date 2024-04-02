package main

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"strconv"
	"strings"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/run"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var push = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"source name",
	},
	ArgumentsOptional: []string{
		"command or pattern path",
	},
	Flags: cli.Flags{
		"c": {
			Usage: "Check mode",
		},
		"f": {
			Placeholder: "regexp",
			Usage:       "Filter parent Command IDs",
		},
		"h": {
			Placeholder: "host",
			Usage:       "Push to a specific host address",
		},
		"n": {
			Placeholder: "regexp",
			Usage:       "Filter PushTarget",
		},
		"p": {
			Default:     []string{"4000"},
			Placeholder: "port",
			Usage:       "Push to a specific host port",
		},
		"u": {
			Default:     []string{"/etcha/v1/push"},
			Placeholder: "path",
			Usage:       "Push to a specific host path",
		},
	},
	Run: func(ctx context.Context, args []string, flags cli.Flags, c *config.Config) errs.Err {
		source := args[1]
		cmd := strings.Join(args[2:], " ")
		re, _ := flags.Value("f")
		f, e := regexp.Compile(re)
		if e != nil {
			return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error parsing filter: %w", e)))
		}

		re, _ = flags.Value("t")
		t, e := regexp.Compile(re)
		if e != nil {
			return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error parsing filter: %w", e)))
		}

		_, check := flags.Value("c")

		targets := maps.Clone(c.Build.PushTargets)

		if host, ok := flags.Value("h"); ok {
			po, _ := flags.Value("p")
			p, err := strconv.Atoi(po)
			if err != nil {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error parsing host port: %w", err)))
			}

			u, _ := flags.Value("u")

			targets = map[string]config.PushTarget{
				host: {
					Hostname: host,
					Path:     u,
					Port:     p,
					SourcePatterns: map[string]string{
						source: "",
					},
				},
			}
		}

		r, err := run.PushTargets(ctx, c, targets, source, cmd, run.PushOpts{
			Check:          check,
			ParentIDFilter: f,
			TargetFilter:   t,
		})
		if r == nil && err != nil {
			return err
		}

		if r == nil {
			return nil
		}

		logger.Raw(r...)
		logger.Raw("\n")

		return err
	},
	Usage: "Push signed commands or patterns to a destination URL.  Must specify a Source, which will push Commands or Patterns specified in build.pushTargets to targets with that Source.  Can optionally specify filters and custom targets.  May also specify a Pattern or Command to run against the Sources, otherwise the Pattern or Command specified in sourcePatterns will be used.",
}
