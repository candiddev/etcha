package pattern

import (
	"context"
	"fmt"
	"strings"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

// Lint will check a jsonnet paths, the format of the files, and optionally check the scripts of a pattern.
func Lint(ctx context.Context, c *config.Config, path string, checkFormat bool) (types.Results, errs.Err) {
	c.Test = true

	r, i, err := jsonnet.Lint(ctx, c, path, checkFormat)
	if err != nil {
		return nil, logger.Error(ctx, err)
	}

	if len(c.Build.Linters) > 0 {
		for path, im := range i {
			if !strings.HasSuffix(path, ".jsonnet") {
				continue
			}

			p, err := ParsePatternFromImports(ctx, c, "", im)
			if err != nil {
				return r, err
			}

			for name, linter := range c.Build.Linters {
				if linter != nil {
					for _, cmd := range append(p.Build, p.Run...) {
						if out, err := linter.Run(ctx, c.CLI, cmd.Check+"\n"+cmd.Change+"\n"+cmd.Remove, ""); err != nil {
							r[path] = append(r[path], fmt.Sprintf("linter %s: %s", name, out.String()))
						}
					}
				}
			}
		}
	}

	return r, nil
}
