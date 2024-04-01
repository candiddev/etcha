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
	PushMaxWorkers    int                   `json:"pushMaxWorkers"`
	PushTLSSkipVerify bool                  `json:"pushTLSSkipVerify"`
	PushTargets       map[string]PushTarget `json:"pushTargets"`
	SigningCommands   commands.Commands     `json:"signingCommands"`
	SigningExec       *commands.Exec        `json:"signingExec,omitempty"`

	/* This is a string as it may be a KDF key or plaintext key */
	SigningKey string `json:"signingKey"`
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

// PushTarget is a target that can be pushed.
type PushTarget struct {
	Hostname string         `json:"hostname"`
	Insecure bool           `json:"insecure"`
	Path     string         `json:"path"`
	Port     int            `json:"port"`
	Sources  []string       `json:"sources"`
	Vars     map[string]any `json:"vars"`
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
		Sources: map[string]*Source{},
		Run: Run{
			ListenAddress:   ":4000",
			RateLimiterRate: "10-M",
		},
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

	wd, err := os.Getwd()
	if err != nil {
		return errs.ErrReceiver.Wrap(errors.New("couldn't get working directory"))
	}

	if !strings.HasPrefix(c.Run.StateDir, "/") {
		c.Run.StateDir = filepath.Join(wd, c.Run.StateDir)
	}

	for k, v := range c.Build.PushTargets {
		if v.Hostname == "" {
			v.Hostname = k
		}

		if v.Path == "" {
			v.Path = "/etcha/v1/push"
		}

		if v.Port == 0 {
			v.Port = 4000
		}

		c.Build.PushTargets[k] = v
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
