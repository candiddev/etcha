// Package pattern contains functions for building, signing, testing, linting, and running patterns.
package pattern

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/jwt"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

var (
	ErrPatternMissingKey = errors.New("missing build.signingKey")
	ErrPatternSigningJWT = errors.New("error signing jwt")
)

// Pattern is a list of Build and Runtime Commands.
type Pattern struct {
	Audience     []string          `json:"audience"`
	Build        commands.Commands `json:"build"`
	BuildExec    commands.Exec     `json:"buildExec,omitempty"`
	ExpiresInSec int               `json:"expiresInSec"`
	ID           string            `json:"id"`
	Issuer       string            `json:"issuer"`
	Run          commands.Commands `json:"run"`
	RunEnv       types.EnvVars     `json:"runEnv"`
	RunExec      commands.Exec     `json:"runExec,omitempty"`
	Subject      string            `json:"subject"`

	Imports *jsonnet.Imports `json:"-"`
	JWT     string           `json:"-"`
}

// ParsePatternFromImports returns a new Pattern from a list of imports.
func ParsePatternFromImports(ctx context.Context, c *config.Config, configSource string, imports *jsonnet.Imports) (*Pattern, errs.Err) {
	if imports == nil || len(imports.Files) == 0 {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(commands.ErrCommandsEmpty))
	}

	p := Pattern{
		Imports: imports,
	}
	r := jsonnet.NewRender(ctx, c)
	r.Import(imports)

	if err := r.Render(ctx, &p); err != nil {
		return nil, logger.Error(ctx, err)
	}

	if s, ok := c.Sources[configSource]; ok && s != nil {
		p.BuildExec = c.Exec.Override(s.Exec, &p.BuildExec)
		p.RunExec = c.Exec.Override(s.Exec, &p.RunExec)
	} else if !c.Exec.AllowOverride {
		p.BuildExec = c.Exec
		p.RunExec = c.Exec
	}

	if p.RunEnv == nil {
		p.RunEnv = map[string]string{}
	}

	if len(p.Build) == 0 && len(p.Run) == 0 {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(commands.ErrCommandsEmpty))
	}

	if len(p.Build) != 0 {
		if err := p.Build.Validate(ctx); err != nil {
			return nil, logger.Error(ctx, err)
		}
	}

	if len(p.Run) != 0 {
		if err := p.Run.Validate(ctx); err != nil {
			return nil, logger.Error(ctx, err)
		}
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

	return ParsePatternFromImports(ctx, c, configSource, i)
}

// Sign creates a signed JWT.
func (p *Pattern) Sign(ctx context.Context, c *config.Config, buildManifest string, runEnv map[string]string) (string, errs.Err) {
	if c.Build.SigningKey.IsNil() && len(c.Build.SigningCommands) == 0 {
		return "", logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPatternMissingKey))
	}

	e := time.Time{}
	if p.ExpiresInSec != 0 {
		e = time.Now().Add(time.Duration(p.ExpiresInSec) * time.Second)
	}

	if p.RunEnv == nil {
		p.RunEnv = types.EnvVars{}
	}

	for k, v := range runEnv {
		p.RunEnv[k] = v
	}

	j := &JWT{
		EtchaBuildManifest: buildManifest,
		EtchaPattern:       p.Imports,
		EtchaVersion:       cli.BuildVersion,
		EtchaRunEnv:        p.RunEnv,
	}

	t, err := jwt.New(j, e, p.Audience, p.ID, p.Issuer, p.Subject)
	if err != nil {
		return "", logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPatternSigningJWT, err))
	}

	if len(c.Build.SigningCommands) > 0 {
		e := types.EnvVars{
			"ETCHA_PAYLOAD": t.PayloadBase64,
		}

		out, err := c.Build.SigningCommands.Run(ctx, c.CLI, e, c.Exec.Override(c.Build.SigningCommandsExec), commands.ModeChange)
		if err != nil {
			return "", logger.Error(ctx, err)
		}

		for _, event := range out.Events() {
			if event.Name == "jwt" && len(event.Outputs) > 0 {
				return string(event.Outputs[0].Change), nil
			}
		}

		return "", logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPatternSigningJWT, fmt.Errorf("no token returned from signingCommands")))
	}

	if err := t.Sign(c.Build.SigningKey); err != nil {
		return "", logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	return t.String(), logger.Error(ctx, nil)
}
