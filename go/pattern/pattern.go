// Package pattern contains functions for building, signing, testing, linting, and running patterns.
package pattern

import (
	"context"
	"errors"
	"time"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/cryptolib"
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
	key, err := cryptolib.ParseKey[cryptolib.KeyProviderPrivate](c.Build.SigningKey)
	if err != nil {
		// try to decrypt the key
		if ev, err := cryptolib.ParseEncryptedValue(c.Build.SigningKey); err == nil {
			if s, err := ev.Decrypt(nil); err == nil {
				if k, err := cryptolib.ParseKey[cryptolib.KeyProviderPrivate](string(s)); err == nil {
					key = k
				}
			}
		}
	}

	if key.IsNil() && len(c.Build.SigningCommands) == 0 {
		return "", nil, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPatternMissingKey))
	}

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
		return "", r, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPatternSigningJWT, err))
	}

	if len(c.Build.SigningCommands) > 0 {
		e := types.EnvVars{
			"ETCHA_PAYLOAD": t.PayloadBase64,
		}

		out, err := c.Build.SigningCommands.Run(ctx, c.CLI, c.Exec.Override(c.Build.SigningExec), commands.CommandsRunOpts{
			Env:      e,
			ParentID: "signingCommands",
		})
		if err != nil {
			return "", r, logger.Error(ctx, err)
		}

		for _, event := range out.Events() {
			if event.Name == "jwt" && len(event.Outputs) > 0 {
				return string(event.Outputs[0].Change), r, nil
			}
		}

		return "", r, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPatternSigningJWT, errors.New("no token returned from signingCommands")))
	}

	if err := t.Sign(key); err != nil {
		return "", r, logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	return t.String(), r, logger.Error(ctx, nil)
}
