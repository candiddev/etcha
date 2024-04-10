package pattern

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"time"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/get"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/jwt"
	"github.com/candiddev/shared/go/logger"
)

var ErrEqualBuildManifest = errors.New("etchaBuildManifest does not match")
var ErrEqualEmpty = errors.New("old JWT does not exist")
var ErrEqualPattern = errors.New("etchaPattern does not match")
var ErrEqualVersion = errors.New("etchaVersion does not match")

// JWT is an artifact JWT.
type JWT struct {
	EtchaBuildManifest string           `json:"etchaBuildManifest"`
	EtchaPattern       *jsonnet.Imports `json:"etchaPattern"`
	EtchaRunVars       map[string]any   `json:"etchaRunVars"`
	EtchaVersion       string           `json:"etchaVersion,omitempty"`
	Raw                string           `json:"-"`
}

// ParseJWT renders a JWT from content.
func ParseJWT(ctx context.Context, c *config.Config, token, source string) (*JWT, *jwt.RegisteredClaims, errs.Err) {
	j := JWT{
		Raw: token,
	}

	k, r, err := c.ParseJWT(ctx, &j, token, source)
	if err != nil {
		return &j, r, logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	logger.Info(ctx, fmt.Sprintf("Received JWT for %s from key %s", source, k.ID))

	return &j, r, nil
}

// ParseJWTFromPath reads a path and parse a JWT.
func ParseJWTFromPath(ctx context.Context, c *config.Config, configSource, path string) (*JWT, *jwt.RegisteredClaims, errs.Err) {
	ca := filepath.Join(c.Run.StateDir, configSource+".jwt")

	b := bytes.Buffer{}

	var err error

	if configSource == "" {
		_, err = get.File(ctx, path, &b, time.Time{})
	} else {
		err = get.FileCache(ctx, path, &b, ca)
	}

	if err != nil {
		return nil, nil, logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error reading JWT: %w", err)))
	}

	return ParseJWT(ctx, c, b.String(), configSource)
}

// ParseJWTFromSource gathers and renders JWTs from the source in a config.
func ParseJWTFromSource(ctx context.Context, source string, c *config.Config) *JWT {
	if s, ok := c.Sources[source]; ok {
		for _, target := range s.PullPaths {
			j, _, err := ParseJWTFromPath(ctx, c, source, target)
			if err == nil {
				logger.Error(ctx, nil) //nolint: errcheck

				return j
			}

			logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error parsing JWT for source %s", source), err)) //nolint: errcheck
		}
	}

	logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("no valid targets for source %s", source))) //nolint: errcheck

	return nil
}

// Equal checks if two JWTs are equivalent.
func (j *JWT) Equal(j2 *JWT, ignoreVersion bool) error {
	if j2 == nil {
		return ErrEqualPattern
	}

	switch {
	case j.EtchaBuildManifest != j2.EtchaBuildManifest:
		return ErrEqualBuildManifest
	case !ignoreVersion && j.EtchaVersion != j2.EtchaVersion:
		return ErrEqualVersion
	case j.EtchaPattern.Diff("", "", j2.EtchaPattern) != "":
		return ErrEqualPattern
	}

	return nil
}

// Pattern returns a Pattern from the JWT.
func (j *JWT) Pattern(ctx context.Context, c *config.Config, configSource string) (*Pattern, errs.Err) {
	vars := maps.Clone(j.EtchaRunVars)
	if vars == nil {
		vars = map[string]any{}
	}

	vars["jwt"] = j.Raw

	p, err := ParsePatternFromImports(ctx, c, configSource, j.EtchaPattern, j.EtchaRunVars)
	if err != nil {
		return nil, logger.Error(ctx, err)
	}

	p.JWT = j.Raw
	p.RunVars = j.EtchaRunVars

	return p, logger.Error(ctx, nil)
}
