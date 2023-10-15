package commands

import (
	"context"
	"errors"
	"strings"

	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var (
	ErrRunEmpty = errors.New("error running blank script")
)

// Exec configures top-level Exec options.
type Exec struct {
	AllowOverride       bool     `json:"allowOverride"`
	Command             string   `json:"command"`
	ContainerEntrypoint string   `json:"containerEntrypoint"`
	ContainerImage      string   `json:"containerImage"`
	ContainerPrivileged bool     `json:"containerPrivileged"`
	ContainerPull       string   `json:"containerPull"`
	ContainerUser       string   `json:"containerUser"`
	ContainerVolumes    []string `json:"containerVolumes"`
	Environment         []string `json:"environment"`
	Group               string   `json:"group"`
	User                string   `json:"user"`
	WorkDir             string   `json:"workDir"`
}

// Override will return the absolute Exec from an ordered list of Execs.
func (e Exec) Override(o ...*Exec) Exec {
	out := e

	if e.AllowOverride {
		for i := range o {
			if o[i] == nil || !out.AllowOverride {
				break
			}

			out = *o[i]
		}
	}

	return out
}

// Run will run a script using the Exec.
func (e *Exec) Run(ctx context.Context, c cli.Config, script, stdin string) (cli.CmdOutput, errs.Err) {
	var s []string

	if e.Command == "" {
		s = strings.Split(script, " ")
	} else {
		s = strings.Split(e.Command, " ")
	}

	if len(s) == 0 {
		return "", logger.Error(ctx, errs.ErrReceiver.Wrap(ErrCommandsEmpty))
	}

	command := s[0]

	args := []string{}
	if len(s) > 1 {
		args = s[1:]
	}

	if e.Command != "" {
		args = append(args, script)
	}

	return c.Run(ctx, cli.RunOpts{
		Args:                args,
		Command:             command,
		ContainerEntrypoint: e.ContainerEntrypoint,
		ContainerImage:      e.ContainerImage,
		ContainerPrivileged: e.ContainerPrivileged,
		ContainerUser:       e.ContainerUser,
		ContainerVolumes:    e.ContainerVolumes,
		Environment:         e.Environment,
		Group:               e.Group,
		NoErrorLog:          true,
		Stdin:               stdin,
		User:                e.User,
		WorkDir:             e.WorkDir,
	})
}
