// Package config contains configuration structs for Etcha.
package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
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
var ErrMissingBuildKey = errors.New("missing build.signingKey")
var ErrSignJWT = errors.New("error signing jwt")

// Config contains main application configurations.
type Config struct {
	Build      Build              `json:"build"`
	CLI        cli.Config         `json:"cli"`
	Exec       commands.Exec      `json:"exec"`
	License    License            `json:"license"`
	LicenseKey string             `json:"licenseKey"`
	Lint       Lint               `json:"lint"`
	Run        Run                `json:"run"`
	Sources    map[string]*Source `json:"sources"`
	Targets    map[string]Target  `json:"targets"`
	Vars       map[string]any     `json:"vars"`
}

// Build configures Etcha's build behavior.
type Build struct {
	PushMaxWorkers  int               `json:"pushMaxWorkers"`
	SigningCommands commands.Commands `json:"signingCommands"`
	SigningExec     *commands.Exec    `json:"signingExec,omitempty"`

	/* This is a string as it may be a KDF key or plaintext key */
	SigningKey    string `json:"signingKey"`
	TLSSkipVerify bool   `json:"tlsSkipVerify"`

	key cryptolib.Key[cryptolib.KeyProviderPrivate]
}

// License configures licensing.
type License struct {
	Commands     int    `json:"etchaCommands"`
	Organization string `json:"sub"`
	Sources      int    `json:"etchaSources"`
	Targets      int    `json:"etchaTargets"`
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

// Target is a remote Etcha instance.
type Target struct {
	Hostname       string            `json:"hostname"`
	Insecure       bool              `json:"insecure"`
	PathPush       string            `json:"pathPush"`
	PathShell      string            `json:"pathShell"`
	Port           string            `json:"string"`
	SourcePatterns map[string]string `json:"sourcePatterns"`
	Vars           map[string]any    `json:"vars"`
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
	RunFrequencySec   int                                         `json:"runFrequencySec"`
	RunMulti          bool                                        `json:"runMulti"`
	Shell             string                                      `json:"shell"`
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
		Build: Build{
			PushMaxWorkers: runtime.NumCPU(),
		},
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
		License: License{
			Commands: 10,
			Sources:  1,
			Targets:  3,
		},
		Sources: map[string]*Source{},
		Run: Run{
			ListenAddress:   ":4000",
			RateLimiterRate: "10-M",
		},
		Targets: map[string]Target{},
		Vars: map[string]any{
			"sysinfo": GetSysInfo(),
		},
	}
}

// SysInfo is various details about the underlying system.
type SysInfo struct {
	CPULogical  int    `json:"cpuLogical,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
	MemoryTotal int    `json:"memoryTotal,omitempty"`
	OSType      string `json:"osType,omitempty"`
	RuntimeArch string `json:"runtimeArch,omitempty"`
}

// GetSysInfo gathers SysInfo.
func GetSysInfo() *SysInfo {
	s := SysInfo{
		CPULogical:  runtime.NumCPU(),
		OSType:      runtime.GOOS,
		RuntimeArch: runtime.GOARCH,
	}

	hostname, err := os.Hostname()
	if err == nil {
		s.Hostname = hostname
	}

	if s.OSType == "linux" {
		// MemoryMB
		if f, err := os.ReadFile("/proc/meminfo"); err == nil {
			r := regexp.MustCompile(`^MemTotal:\s+(\S+)`).FindStringSubmatch(string(f))
			if len(r) == 2 {
				if n, err := strconv.Atoi(r[1]); err == nil {
					s.MemoryTotal = n / 1024
				}
			}
		}
	}

	return &s
}

func (c *Config) Parse(ctx context.Context, configArgs []string) errs.Err {
	if err := config.Parse(ctx, c, configArgs, "ETCHA", c.CLI.ConfigPath); err != nil {
		return logger.Error(ctx, err)
	}

	l := License{}

	if _, err := cli.ParseLicense(ctx, c.LicenseKey, &l); err == nil {
		c.License = l
	} else {
		logger.Error(ctx, err) //nolint:errcheck
	}

	wd, err := os.Getwd()
	if err != nil {
		return logger.Error(ctx, errors.New("couldn't get working directory"))
	}

	if !strings.HasPrefix(c.Run.StateDir, "/") {
		c.Run.StateDir = filepath.Join(wd, c.Run.StateDir)
	}

	if len(c.Sources) > c.License.Sources {
		return logger.Error(ctx, fmt.Errorf("number of sources (%d) exceeds the license amount (%d), please upgrade your license or reduce the number of sources", len(c.Sources), c.License.Targets))
	}

	for k, v := range c.Sources {
		if ct := v.Commands.Count(); ct > c.License.Commands {
			return logger.Error(ctx, fmt.Errorf("number of commands (%d) in source %s exceeds the license amount (%d), please upgrade your license or reduce the number of commands", ct, k, c.License.Commands))
		}
	}

	if len(c.Targets) > c.License.Targets {
		return logger.Error(ctx, fmt.Errorf("number of targets (%d) exceeds the license amount (%d), please upgrade your license or reduce the number of targets", len(c.Targets), c.License.Targets))
	}

	for k, v := range c.Targets {
		if v.Hostname == "" {
			v.Hostname = k
		}

		if v.PathPush == "" {
			v.PathPush = "/etcha/v1/push"
		}

		if v.PathShell == "" {
			v.PathShell = "/etcha/v1/shell"
		}

		if v.Port == "" {
			v.Port = "4000"
		}

		c.Targets[k] = v
	}

	return logger.Error(ctx, nil)
}

// ParseJWT parses a JWT token into a customClaims.
func (c *Config) ParseJWT(ctx context.Context, customClaims any, token string, source string) (key cryptolib.Key[cryptolib.KeyProviderPublic], r *jwt.RegisteredClaims, err error) {
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
		ve.Env = types.EnvVars{
			"ETCHA_JWT": token,
		}

		out, e := vc.Run(ctx, c.CLI, ve, commands.CommandsRunOpts{
			ParentID: source + " > verifyCommands",
		})
		if e != nil {
			return key, r, e
		}

		token = ""

		for _, event := range out.Events() {
			if event.Name == "jwt" && len(event.Outputs) > 0 {
				token = event.Outputs[0].Change.String()

				break
			}
		}

		if token == "" {
			return key, nil, cryptolib.ErrVerify
		}

		// The above hopefully validated the token.
		t, _, _ = jwt.Parse(token, nil)
	} else {
		t, key, err = jwt.Parse(token, keys)
	}

	if t != nil {
		r, payloadErr = t.ParsePayload(customClaims, "", "", "")
	}

	if err == nil && payloadErr != nil {
		err = payloadErr
	}

	return key, r, err
}

// ParseJWTFile reads a JWT file path and parses the token into customClaims.
func (c *Config) ParseJWTFile(ctx context.Context, customClaims any, path, source string) (key cryptolib.Key[cryptolib.KeyProviderPublic], r *jwt.RegisteredClaims, err errs.Err) {
	s, e := os.ReadFile(path)
	if e != nil {
		return key, r, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrParseJWT, e))
	}

	if key, r, e = c.ParseJWT(ctx, customClaims, string(s), source); e != nil {
		return key, r, logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error parsing JWT %s", path), e))
	}

	return key, r, logger.Error(ctx, nil)
}

func (c *Config) SignJWT(ctx context.Context, j *jwt.Token) (string, errs.Err) {
	if c.Build.key.IsNil() {
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

		c.Build.key = key
	}

	if c.Build.key.IsNil() && len(c.Build.SigningCommands) == 0 {
		return "", logger.Error(ctx, errs.ErrReceiver.Wrap(ErrMissingBuildKey))
	}

	if len(c.Build.SigningCommands) > 0 {
		e := types.EnvVars{
			"ETCHA_PAYLOAD": j.PayloadBase64,
		}

		out, err := c.Build.SigningCommands.Run(ctx, c.CLI, c.Exec.Override(c.Build.SigningExec), commands.CommandsRunOpts{
			Env:      e,
			ParentID: "signingCommands",
		})
		if err != nil {
			return "", logger.Error(ctx, err)
		}

		for _, event := range out.Events() {
			if event.Name == "jwt" && len(event.Outputs) > 0 {
				return string(event.Outputs[0].Change), logger.Error(ctx, nil)
			}
		}

		return "", logger.Error(ctx, errs.ErrReceiver.Wrap(ErrSignJWT, errors.New("no token returned from signingCommands")))
	}

	if err := j.Sign(c.Build.key); err != nil {
		return "", logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	return j.String(), logger.Error(ctx, nil)
}
