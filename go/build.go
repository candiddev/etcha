package main

import (
	"context"
	"path/filepath"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/errs"
)

func build(ctx context.Context, args []string, c *config.Config) errs.Err {
	source := args[1]
	destination := args[2]

	configSource := "etcha"
	if len(args) == 4 {
		configSource = args[3]
	}

	c.Vars["buildDir"] = filepath.Dir(source)
	c.Vars["buildPath"] = source

	p, err := pattern.ParsePatternFromPath(ctx, c, configSource, source)
	if err != nil {
		return err
	}

	return p.BuildSign(ctx, c, destination)
}
