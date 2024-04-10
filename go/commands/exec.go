package commands

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

var (
	ErrRunEmpty = errors.New("error running blank script")
)

// Exec configures top-level Exec options.
type Exec struct {
	AllowOverride       bool          `json:"allowOverride"`
	Command             string        `json:"command"`
	ContainerEntrypoint string        `json:"containerEntrypoint"`
	ContainerImage      string        `json:"containerImage"`
	ContainerNetwork    string        `json:"containerNetwork"`
	ContainerPrivileged bool          `json:"containerPrivileged"`
	ContainerPull       string        `json:"containerPull"`
	ContainerUser       string        `json:"containerUser"`
	ContainerVolumes    []string      `json:"containerVolumes"`
	ContainerWorkDir    string        `json:"containerWorkDir"`
	Env                 types.EnvVars `json:"env"`
	EnvInherit          bool          `json:"envInherit"`
	Group               string        `json:"group"`
	Stdin               io.Reader     `json:"-"`
	Stderr              io.Writer     `json:"-"`
	Stdout              io.Writer     `json:"-"`
	Sudo                bool          `json:"sudo"`
	User                string        `json:"user"`
	WorkDir             string        `json:"workDir"`
}

// Override will return the absolute Exec from an ordered list of Execs.
func (e Exec) Override(o ...*Exec) *Exec {
	out := e

	if e.AllowOverride {
		for i := range o {
			if o[i] == nil || !out.AllowOverride {
				break
			}

			out = *o[i]
		}
	}

	return &out
}

func (e *Exec) RunOpts(ctx context.Context, script string) (cli.RunOpts, errs.Err) {
	var s []string

	if e.Command == "" {
		s = strings.Split(script, " ")
	} else {
		s = strings.Split(e.Command, " ")
	}

	if len(s) == 0 {
		return cli.RunOpts{}, logger.Error(ctx, ErrCommandsEmpty)
	}

	command := s[0]

	args := []string{}
	if len(s) > 1 {
		args = s[1:]
	}

	if e.Command != "" && script != "" {
		args = append(args, script)
	}

	return cli.RunOpts{
		Args:                args,
		Command:             command,
		ContainerEntrypoint: e.ContainerEntrypoint,
		ContainerImage:      e.ContainerImage,
		ContainerNetwork:    e.ContainerNetwork,
		ContainerPrivileged: e.ContainerPrivileged,
		ContainerUser:       e.ContainerUser,
		ContainerVolumes:    e.ContainerVolumes,
		ContainerWorkDir:    e.ContainerWorkDir,
		Environment:         e.Env.GetEnv(),
		EnvironmentInherit:  e.EnvInherit,
		Group:               e.Group,
		NoErrorLog:          true,
		Sudo:                e.Sudo,
		Stderr:              e.Stderr,
		Stdout:              e.Stdout,
		Stdin:               e.Stdin,
		User:                e.User,
		WorkDir:             e.WorkDir,
	}, nil
}

// Run will run a script using the Exec.
func (e *Exec) Run(ctx context.Context, c cli.Config, script string) (cli.CmdOutput, errs.Err) {
	opts, err := e.RunOpts(ctx, script)
	if err != nil {
		return "", logger.Error(ctx, err)
	}

	return c.Run(ctx, opts)
}
