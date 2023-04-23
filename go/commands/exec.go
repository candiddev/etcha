package commands

import (
	"context"
	"strings"

	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
)

// Exec configures top-level Exec options.
type Exec struct {
	Command             string   `json:"command"`
	ContainerEntrypoint string   `json:"containerEntrypoint"`
	ContainerImage      string   `json:"containerImage"`
	ContainerPrivileged bool     `json:"containerPrivileged"`
	ContainerUser       string   `json:"containerUser"`
	ContainerVolumes    []string `json:"containerVolumes"`
	Environment         []string `json:"environment"`
	Flags               string   `json:"flags"`
	Override            bool     `json:"override"`
	Test                bool     `json:"testMode"`
	WorkDir             string   `json:"workDir"`
}

// Run will run a script using the Exec.
func (e *Exec) Run(ctx context.Context, c cli.Config, stdin, script string) (cli.CmdOutput, errs.Err) {
	args := []string{}
	if e.Flags != "" {
		args = strings.Split(e.Flags, " ")
	}

	if script != "" {
		args = append(args, script)
	}

	return c.Run(ctx, cli.RunOpts{
		Args:                args,
		Command:             e.Command,
		ContainerEntrypoint: e.ContainerEntrypoint,
		ContainerImage:      e.ContainerImage,
		ContainerPrivileged: e.ContainerPrivileged,
		ContainerUser:       e.ContainerUser,
		ContainerVolumes:    e.ContainerVolumes,
		Environment:         e.Environment,
		NoErrorLog:          true,
		Stdin:               stdin,
		WorkDir:             e.WorkDir,
	})
}
