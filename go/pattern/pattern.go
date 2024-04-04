// Package pattern contains functions for building, signing, testing, linting, and running patterns.
package pattern

import (
	"context"
	"time"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/jwt"
	"github.com/candiddev/shared/go/logger"
)

// Pattern is a list of Build and Runtime Commands.
type Pattern struct {
	Audience     []string          `json:"audience"`
	Build        commands.Commands `json:"build"`
	BuildExec    *commands.Exec    `json:"buildExec,omitempty"`
	ExpiresInSec int               `json:"expiresInSec"`
	ID           string            `json:"id"`
	Issuer       string            `json:"issuer"`
	Run          commands.Commands `json:"run"`
	RunExec      *commands.Exec    `json:"runExec,omitempty"`
	RunVars      map[string]any    `json:"runVars"`
	Subject      string            `json:"subject"`

	Imports *jsonnet.Imports `json:"-"`
	JWT     string           `json:"-"`
}

// ParsePatternFromImports returns a new Pattern from a list of imports.
func ParsePatternFromImports(ctx context.Context, c *config.Config, configSource string, imports *jsonnet.Imports, runVars map[string]any) (*Pattern, errs.Err) {
	if imports == nil || len(imports.Files) == 0 {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(commands.ErrCommandsEmpty))
	}

	p := Pattern{
		Imports: imports,
	}

	vars := map[string]any{}

	for k, v := range runVars {
		vars[k] = v
	}

	for k, v := range c.Vars {
		vars[k] = v
	}

	s := c.Sources[configSource]

	if s != nil {
		for k, v := range s.Vars {
			vars[k] = v
		}
	}

	vars["source"] = configSource
	vars["sysinfo"] = config.GetSysInfo()

	if configSource == "test" {
		vars["test"] = true
	} else {
		vars["test"] = false
	}

	exec := &c.Exec

	if s != nil {
		exec = c.Exec.Override(s.Exec)
	}

	r := jsonnet.NewRender(ctx, map[string]any{
		"exec": exec,
		"vars": vars,
	})
	r.Import(imports)

	if !exec.EnvInherit {
		e := exec.Env.GetEnv()
		r.SetEnv(&e)
	}

	if err := r.Render(ctx, &p); err != nil {
		return nil, logger.Error(ctx, err)
	}

	p.BuildExec = exec.Override(p.BuildExec)
	p.RunExec = exec.Override(p.RunExec)

	if len(p.Build) == 0 && len(p.Run) == 0 {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(commands.ErrCommandsEmpty))
	}

	return &p, nil
}

// ParsePatternFromPath returns a new Pattern from a path.
func ParsePatternFromPath(ctx context.Context, c *config.Config, configSource, path string) (*Pattern, errs.Err) {
	r := jsonnet.NewRender(ctx, c)

	i, err := r.GetPath(ctx, path)
	if err != nil {
		return nil, logger.Error(ctx, err)
	}

	return ParsePatternFromImports(ctx, c, configSource, i, nil)
}

// Sign creates a signed JWT.
func (p *Pattern) Sign(ctx context.Context, c *config.Config, buildManifest string, runVars map[string]any) (string, *jwt.RegisteredClaims, errs.Err) {
	e := time.Time{}
	if p.ExpiresInSec != 0 {
		e = time.Now().Add(time.Duration(p.ExpiresInSec) * time.Second)
	}

	j := &JWT{
		EtchaBuildManifest: buildManifest,
		EtchaPattern:       p.Imports,
		EtchaVersion:       cli.BuildVersion,
		EtchaRunVars:       runVars,
	}

	t, r, err := jwt.New(j, e, p.Audience, p.ID, p.Issuer, p.Subject)
	if err != nil {
		return "", r, logger.Error(ctx, errs.ErrReceiver.Wrap(config.ErrSignJWT, err))
	}

	s, er := c.SignJWT(ctx, t)

	return s, r, logger.Error(ctx, er)
}
