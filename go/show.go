package main

import (
	"context"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

func showCommands(ctx context.Context, args []string, c *config.Config) errs.Err {
	var p *pattern.Pattern

	// Try JWT
	j, err := pattern.ParseJWTFromPath(logger.SetLevel(ctx, logger.LevelNone), c, args[1], "")
	if err != nil || j.ID == "" {
		p, err = pattern.ParsePatternFromPath(ctx, c, "", args[1])
	} else {
		p, err = j.Pattern(ctx, c, "")
	}

	if p != nil {
		logger.Raw(types.JSONToString(p) + "\n")
	}

	return err
}

func showJWT(ctx context.Context, args []string, c *config.Config) errs.Err {
	j, _ := pattern.ParseJWTFromPath(ctx, c, "", args[1])

	if j != nil {
		logger.Raw(types.JSONToString(j) + "\n")
	}

	return nil
}
