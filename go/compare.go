package main

import (
	"context"
	"errors"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/diff"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var compare = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"new jwt path or URL",
		"old jwt path or URL",
	},
	ArgumentsOptional: []string{
		"ignore version, default: no",
	},
	Run: func(ctx context.Context, args []string, c *config.Config) errs.Err {
		j1, err := pattern.ParseJWTFromPath(ctx, c, "", args[1])
		if err != nil {
			return err
		}

		j2, err := pattern.ParseJWTFromPath(ctx, c, "", args[2])
		if err != nil {
			return err
		}

		ignore := false
		if len(args) == 4 && args[3] != "" {
			ignore = true
		}

		if err := j1.Equal(j2, ignore); err != nil {
			var c string

			switch err {
			case pattern.ErrEqualBuildManifest:
				c = string(diff.Diff("old etchaManifest", []byte(j1.EtchaBuildManifest), "new etchaManifest", []byte(j2.EtchaBuildManifest)))
			case pattern.ErrEqualEmpty:
				c = "old JWT is empty"
			case pattern.ErrEqualPattern:
				c = j1.EtchaPattern.Diff("old etchaPattern", "new etchaPattern", j2.EtchaPattern)
			case pattern.ErrEqualVersion:
				c = string(diff.Diff("old etchaVersion", []byte(j2.EtchaVersion), "new etchaVersion", []byte(j1.EtchaVersion)))
			}

			return logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("JWTs do not match"), err), c)
		}

		return nil
	},
	Usage: "Compare two JWTs to see if they have the same etchaBuildManifest, etchaPattern, and optionally etchaVersion",
}
