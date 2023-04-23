// Package config contains configuration structs for Etcha.
package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/handlers"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/config"
	"github.com/candiddev/shared/go/crypto"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jwt"
	"github.com/candiddev/shared/go/logger"
)

var ErrParseJWT = errors.New("error parsing jwt")

// Config contains main application configurations.
type Config struct {
	CLI      cli.Config                `json:"cli"`
	Exec     commands.Exec             `json:"exec"`
	Handlers handlers.Handlers         `json:"handlers"`
	Linters  map[string]*commands.Exec `json:"linters"`
	JWT      JWT                       `json:"jwt"`
	Run      Run                       `json:"run"`
	Sources  map[string]*Source        `json:"sources"`
	Vars     map[string]any            `json:"vars"`
}

// JWT configures top-level keys for Etcha.
type JWT struct {
	PrivateKey crypto.Ed25519PrivateKey `json:"privateKey"`
	PublicKeys crypto.Ed25519PublicKeys `json:"publicKeys"`
}

// Run configures Etcha's runtime behavior.
type Run struct {
	ListenAddress           string `json:"listenDddress"`
	PushTLSSkipVerify       bool   `json:"pushTLSSkipVerify"`
	RandomizedStartDelaySec int    `json:"randomizedStartDelaySec"`
	RateLimiterRate         string `json:"rateLimiterRate"`
	StateDir                string `json:"stateDir"`
	SystemMetricsSecret     string `json:"systemMetricsSecret"`
	SystemPprofSecret       string `json:"systemPprofSecret"`
	TLSCertificateBase64    string `json:"tlsCertificateBase64"`
	TLSCertificatePath      string `json:"tlsCertificatePath"`
	TLSKeyBase64            string `json:"tlsKeyBase64"`
	TLSKeyPath              string `json:"tlsKeyPath"`
}

// Source contains configurations for a source.
type Source struct {
	AllowPush         bool                     `json:"push"`
	CheckOnly         bool                     `json:"checkOnly"`
	Exec              commands.Exec            `json:"exec"`
	JWTPublicKeys     crypto.Ed25519PublicKeys `json:"jwtPublicKeys"`
	PullIgnoreVersion bool                     `json:"pullIgnoreVersion"`
	PullPaths         []string                 `json:"pullPaths"`
	RunAlwaysCheck    bool                     `json:"runAlwaysCheck"`
	RunFrequency      int                      `json:"runFrequency"`
}

func (c *Config) CLIConfig() *cli.Config {
	return &c.CLI
}

func Default() *Config {
	return &Config{
		Exec: commands.Exec{
			Command:  "/usr/bin/bash",
			Flags:    "-e -o pipefail -c",
			Override: true,
			WorkDir:  os.TempDir(),
		},
		Linters: map[string]*commands.Exec{
			"shellcheck": {
				ContainerImage: "koalaman/shellcheck",
				Flags:          "-s bash -e 2154 -",
			},
		},
		Run: Run{
			ListenAddress:   ":4000",
			RateLimiterRate: "10-M",
			StateDir:        "etcha",
		},
		Sources: map[string]*Source{
			"test": {
				Exec: commands.Exec{
					Override: true,
				},
			},
		},
		Vars: map[string]any{},
	}
}

func (c *Config) Parse(ctx context.Context, configArgs, paths string) errs.Err {
	if err := config.Parse(ctx, c, "etcha", "", configArgs, paths); err != nil {
		return logger.Error(ctx, err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return errs.ErrReceiver.Wrap(errors.New("couldn't get working directory"))
	}

	if !strings.HasPrefix(c.Run.StateDir, "/") {
		c.Run.StateDir = filepath.Join(wd, c.Run.StateDir)
	}

	if err := os.MkdirAll(c.Run.StateDir, 0750); err != nil {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("couldn't create %s: %w", c.Run.StateDir, err)))
	}

	return logger.Error(ctx, nil)
}

// ParseJWT parses a JWT token into a customClaims.
func (c *Config) ParseJWT(customClaims jwt.CustomClaims, token string, publicKeys crypto.Ed25519PublicKeys) (key crypto.Ed25519PublicKey, expiresAt time.Time, err error) {
	for _, k := range append(c.JWT.PublicKeys, publicKeys...) {
		expiresAt, err = jwt.VerifyJWT(k, customClaims, token)
		if err == nil {
			key = k

			break
		}
	}

	if key == "" {
		err = jwt.ErrParsingToken
	}

	return key, expiresAt, err
}

// ParseJWTFile reads a JWT file path and parses the token into customClaims.
func (c *Config) ParseJWTFile(ctx context.Context, customClaims jwt.CustomClaims, path string, publicKeys crypto.Ed25519PublicKeys) (key crypto.Ed25519PublicKey, expiresAt time.Time, err errs.Err) {
	s, e := os.ReadFile(path)
	if e != nil {
		return "", time.Time{}, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrParseJWT, e))
	}

	if key, expiresAt, e = c.ParseJWT(customClaims, string(s), publicKeys); e != nil {
		return key, expiresAt, logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error parsing JWT %s", path), e))
	}

	return key, expiresAt, logger.Error(ctx, nil)
}
