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
	j, err := pattern.ParseJWTFromPath(ctx, c, args[1], "")
	if err != nil || j.ID == "" {
		p, err = pattern.ParsePatternFromPath(ctx, c, "", args[1])
	} else {
		p, err = j.Pattern(ctx, c, "")
	}

	if err != nil {
		return err
	}

	logger.Info(ctx, types.JSONToString(p))

	return nil
}

func showJWT(ctx context.Context, args []string, c *config.Config) errs.Err {
	j, err := pattern.ParseJWTFromPath(ctx, c, "", args[1])
	if err != nil {
		return err
	}

	logger.Info(ctx, types.JSONToString(j))

	return nil
}
