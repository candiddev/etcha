// Package config contains configuration structs for Etcha.
package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/config"
	"github.com/candiddev/shared/go/cryptolib"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jwt"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

var ErrParseJWT = errors.New("error parsing jwt")

// Config contains main application configurations.
type Config struct {
	Build   Build              `json:"build"`
	CLI     cli.Config         `json:"cli"`
	Exec    commands.Exec      `json:"exec"`
	Lint    Lint               `json:"lint"`
	Run     Run                `json:"run"`
	Sources map[string]*Source `json:"sources"`
	Vars    map[string]any     `json:"vars"`
}

// Build configures Etcha's build behavior.
type Build struct {
	PushTLSSkipVerify bool              `json:"pushTLSSkipVerify"`
	SigningCommands   commands.Commands `json:"signingCommands"`
	SigningExec       *commands.Exec    `json:"signingExec,omitempty"`
	SigningKey        string            `json:"signingKey"`
}

// Lint are config values for linters.
type Lint struct {
	Exclude regexp.Regexp             `json:"exclude"`
	Linters map[string]*commands.Exec `json:"linters"`
}

// Run configures Etcha's runtime behavior.
type Run struct {
	ListenAddress           string                                      `json:"listenAddress"`
	RandomizedStartDelaySec int                                         `json:"randomizedStartDelaySec"`
	RateLimiterRate         string                                      `json:"rateLimiterRate"`
	StateDir                string                                      `json:"stateDir"`
	SystemMetricsSecret     string                                      `json:"systemMetricsSecret"`
	SystemPprofSecret       string                                      `json:"systemPprofSecret"`
	TLSCertificateBase64    string                                      `json:"tlsCertificateBase64"`
	TLSCertificatePath      string                                      `json:"tlsCertificatePath"`
	TLSKeyBase64            string                                      `json:"tlsKeyBase64"`
	TLSKeyPath              string                                      `json:"tlsKeyPath"`
	VerifyCommands          commands.Commands                           `json:"verifyCommands"`
	VerifyExec              *commands.Exec                              `json:"verifyExec,omitempty"`
	VerifyKeys              cryptolib.Keys[cryptolib.KeyProviderPublic] `json:"verifyKeys"`
}

// Source contains configurations for a source.
type Source struct {
	AllowPush         bool                                        `json:"allowPush"`
	CheckOnly         bool                                        `json:"checkOnly"`
	Commands          commands.Commands                           `json:"commands"`
	EventsReceive     types.SliceString                           `json:"eventsReceive"`
	EventsReceiveExit bool                                        `json:"eventsReceiveExit"`
	EventsSend        regexp.Regexp                               `json:"eventsSend"`
	Exec              *commands.Exec                              `json:"exec,omitempty"`
	NoRemove          bool                                        `json:"noRemove"`
	NoRestore         bool                                        `json:"noRestore"`
	PullIgnoreVersion bool                                        `json:"pullIgnoreVersion"`
	PullPaths         types.SliceString                           `json:"pullPaths"`
	RunAll            bool                                        `json:"runAll"`
	RunFrequencySec   int                                         `json:"runFrequencySec"`
	RunMulti          bool                                        `json:"runMulti"`
	TriggerOnly       bool                                        `json:"triggerOnly"`
	VerifyCommands    commands.Commands                           `json:"verifyCommands"`
	VerifyExec        *commands.Exec                              `json:"verifyExec,omitempty"`
	VerifyKeys        cryptolib.Keys[cryptolib.KeyProviderPublic] `json:"verifyKeys"`
	Vars              map[string]any                              `json:"vars"`
	WebhookPaths      types.SliceString                           `json:"webhookPaths"`
}

func (c *Config) CLIConfig() *cli.Config {
	return &c.CLI
}

func Default() *Config {
	return &Config{
		Lint: Lint{
			Exclude: *regexp.MustCompile("etcha.jsonnet"),
			Linters: map[string]*commands.Exec{
				"shellcheck": {
					Command:          "-s bash -e 2016 -e 2154 -",
					ContainerImage:   "docker.io/koalaman/shellcheck",
					ContainerNetwork: "none",
					EnvInherit:       true,
				},
			},
		},
		Exec: commands.Exec{
			AllowOverride: true,
			Command:       "/usr/bin/bash -e -o pipefail -c",
			EnvInherit:    true,
		},
		Sources: map[string]*Source{},
		Run: Run{
			ListenAddress:   ":4000",
			RateLimiterRate: "10-M",
		},
		Vars: map[string]any{},
	}
}

func (c *Config) Parse(ctx context.Context, configArgs []string) errs.Err {
	if err := config.Parse(ctx, c, configArgs, "ETCHA", c.CLI.ConfigPath); err != nil {
		return logger.Error(ctx, err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return errs.ErrReceiver.Wrap(errors.New("couldn't get working directory"))
	}

	if !strings.HasPrefix(c.Run.StateDir, "/") {
		c.Run.StateDir = filepath.Join(wd, c.Run.StateDir)
	}

	return logger.Error(ctx, nil)
}

// ParseJWT parses a JWT token into a customClaims.
func (c *Config) ParseJWT(ctx context.Context, customClaims jwt.CustomClaims, token string, source string) (key cryptolib.Key[cryptolib.KeyProviderPublic], err error) {
	var payloadErr error

	keys := c.Run.VerifyKeys
	ve := c.Exec.Override(c.Run.VerifyExec)
	vc := c.Run.VerifyCommands

	if s, ok := c.Sources[source]; ok && s != nil {
		keys = append(keys, s.VerifyKeys...)
		ve = ve.Override(s.VerifyExec)

		if len(s.VerifyCommands) > 0 {
			vc = s.VerifyCommands
		}
	}

	var t *jwt.Token

	if len(vc) > 0 {
		out, e := vc.Run(ctx, c.CLI, types.EnvVars{
			"ETCHA_JWT": token,
		}, ve, false, false)
		if e != nil {
			return key, e
		}

		token = ""

		for _, event := range out.Events() {
			if event.Name == "jwt" && len(event.Outputs) > 0 {
				token = event.Outputs[0].Change.String()

				break
			}
		}

		if token == "" {
			return key, cryptolib.ErrVerify
		}

		// The above hopefully validated the token.
		t, _, _ = jwt.Parse(token, nil)
	} else {
		t, key, err = jwt.Parse(token, keys)
	}

	if t != nil {
		payloadErr = t.ParsePayload(customClaims, "", "", "")
	}

	if err == nil && payloadErr != nil {
		err = payloadErr
	}

	return key, err
}

// ParseJWTFile reads a JWT file path and parses the token into customClaims.
func (c *Config) ParseJWTFile(ctx context.Context, customClaims jwt.CustomClaims, path, source string) (key cryptolib.Key[cryptolib.KeyProviderPublic], err errs.Err) {
	s, e := os.ReadFile(path)
	if e != nil {
		return key, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrParseJWT, e))
	}

	if key, e = c.ParseJWT(ctx, customClaims, string(s), source); e != nil {
		return key, logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error parsing JWT %s", path), e))
	}

	return key, logger.Error(ctx, nil)
}
