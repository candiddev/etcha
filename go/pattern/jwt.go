package pattern

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

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
	EtchaBuildManifest string            `json:"etchaBuildManifest"`
	EtchaPattern       *jsonnet.Imports  `json:"etchaPattern"`
	EtchaRunEnv        map[string]string `json:"etchaRunEnv"`
	EtchaVersion       string            `json:"etchaVersion,omitempty"`
	Raw                string            `json:"-"`

	jwt.RegisteredClaims
}

// ParseJWT renders a JWT from content.
func ParseJWT(ctx context.Context, c *config.Config, token, source string) (*JWT, errs.Err) {
	j := JWT{
		Raw: token,
	}

	if _, err := c.ParseJWT(ctx, &j, token, source); err != nil {
		return &j, logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	return &j, nil
}

// ParseJWTFromPath reads a path and parse a JWT.
func ParseJWTFromPath(ctx context.Context, c *config.Config, configSource, path string) (*JWT, errs.Err) {
	ca := filepath.Join(c.Run.StateDir, configSource+".jwt")

	b := bytes.Buffer{}

	if err := get.FileCache(ctx, path, &b, ca); err != nil {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error reading JWT: %w", err)))
	}

	return ParseJWT(ctx, c, b.String(), configSource)
}

// ParseJWTFromSources gathers and renders JWTs from the sources in a config.
func ParseJWTFromSources(ctx context.Context, source string, c *config.Config) *JWT {
	if s, ok := c.Sources[source]; ok {
		for _, target := range s.PullPaths {
			j, err := ParseJWTFromPath(ctx, c, source, target)
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

// ParseJWTsFromDir finds and renders all JWTs in the StateDir.
func ParseJWTsFromDir(ctx context.Context, c *config.Config) map[string]*JWT {
	j := map[string]*JWT{}

	for k := range c.Sources {
		path := filepath.Join(c.Run.StateDir, k+".jwt")

		if jw, err := ParseJWTFromPath(ctx, c, k, path); err == nil && jw != nil {
			j[k] = jw
		}
	}

	return j
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

// GetRegisteredClaims satisfies the JWT interface.
func (j *JWT) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &j.RegisteredClaims
}

// Pattern returns a Pattern from the JWT.
func (j *JWT) Pattern(ctx context.Context, c *config.Config, configSource string) (*Pattern, errs.Err) {
	p, err := ParsePatternFromImports(ctx, c, configSource, j.EtchaPattern)
	if err != nil {
		return nil, logger.Error(ctx, err)
	}

	for k, v := range j.EtchaRunEnv {
		p.RunEnv[k] = v
	}

	p.JWT = j.Raw

	return p, logger.Error(ctx, nil)
}

// RunEnv parses RunEnv into environment values.
func (j *JWT) RunEnv() []string {
	env := []string{}

	for k, v := range j.EtchaRunEnv {
		env = append(env, strings.ToLower("etcha_run_"+k)+"="+v)
	}

	sort.Strings(env)

	return env
}

func (*JWT) Valid() error {
	return nil
}
