package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

// Mode is the operation that will be performed during a run.
type Mode int

// Modes are the operation that will be performed during a run.
const (
	ModeChange Mode = iota
	ModeCheck
	ModeRemove
)

// Command is a configuration to run.
type Command struct {
	Always    bool     `json:"always,omitempty"`
	Change    string   `json:"change,omitempty"`
	Check     string   `json:"check,omitempty"`
	ChangedBy []string `json:"-"`
	EnvPrefix string   `json:"envPrefix"`
	Exec      *Exec    `json:"exec,omitempty"`
	ID        string   `json:"id"`
	OnChange  []string `json:"onChange,omitempty"`
	OnFail    []string `json:"onFail,omitempty"`
	Remove    string   `json:"remove,omitempty"`
}

// Run will run the Command script for the given Mode.
func (cmd *Command) Run(ctx context.Context, c cli.Config, oldEnv types.EnvVars, exec Exec, m Mode) (out *Output, newEnv types.EnvVars, err errs.Err) {
	cfg := exec.Override(cmd.Exec)
	cfgEnv := cfg.Environment
	ctx = metrics.SetCommandID(ctx, cmd.ID)

	if e := oldEnv.GetEnv(); len(e) > 0 {
		cfg.Environment = append(e, cfg.Environment...)
	}

	if oldEnv == nil {
		newEnv = types.EnvVars{}
	} else {
		newEnv = oldEnv
	}

	out = &Output{
		ID: cmd.ID,
	}

	if m == ModeRemove && cmd.Remove != "" {
		ctx = metrics.SetCommandMode(ctx, metrics.CommandModeRemove)

		logger.Info(ctx, fmt.Sprintf("Removing %s...", cmd.ID))

		out.Removed = true

		if out.Remove, err = cfg.Run(ctx, c, cmd.Remove, ""); err != nil {
			out.RemoveFail = true
			err := logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error removing id %s", cmd.ID)).Wrap(err.Errors()...), out.Remove.String())

			newEnv[cmd.EnvPrefix+"_REMOVE"] = "1"
			newEnv[cmd.EnvPrefix+"_REMOVE_OUT"] = out.Remove.String()

			metrics.CollectCommands(ctx, true)

			return out, newEnv, err
		}

		newEnv[cmd.EnvPrefix+"_REMOVE"] = "0"
		newEnv[cmd.EnvPrefix+"_REMOVE_OUT"] = out.Remove.String()

		metrics.CollectCommands(ctx, false)
	} else {
		ctx = metrics.SetCommandMode(ctx, metrics.CommandModeCheck)

		if cmd.Check == "" && !cmd.Always && len(cmd.ChangedBy) == 0 {
			newEnv[cmd.EnvPrefix+"_CHECK"] = "0"
			metrics.CollectCommands(ctx, false)

			return out, newEnv, nil
		}

		switch {
		case cmd.Always:
			logger.Info(ctx, fmt.Sprintf("Always changing %s...", cmd.ID))
		case len(cmd.ChangedBy) > 0:
			logger.Info(ctx, fmt.Sprintf("Triggering %s via %s...", cmd.ID, strings.Join(cmd.ChangedBy, ", ")))
		default:
			out.Checked = true

			logger.Debug(ctx, fmt.Sprintf("Checking %s...", cmd.ID))

			out.Check, err = cfg.Run(ctx, c, cmd.Check, "")
			if err == nil {
				newEnv[cmd.EnvPrefix+"_CHECK"] = "0"
				newEnv[cmd.EnvPrefix+"_CHECK_OUT"] = out.Check.String()

				metrics.CollectCommands(ctx, false)

				return out, newEnv, nil
			}

			newEnv[cmd.EnvPrefix+"_CHECK_OUT"] = out.Check.String()
			out.CheckFail = true
		}

		cmd.ChangedBy = nil

		metrics.CollectCommands(ctx, true)
		newEnv[cmd.EnvPrefix+"_CHECK"] = "1"

		if m == ModeCheck || cmd.Change == "" {
			return out, newEnv, nil
		}

		ctx = metrics.SetCommandMode(ctx, metrics.CommandModeChange)
		out.Changed = true

		logger.Info(ctx, fmt.Sprintf("Changing %s...", cmd.ID))

		cfg.Environment = append(cfgEnv, newEnv.GetEnv()...) //nolint:gocritic

		if out.Change, err = cfg.Run(ctx, c, cmd.Change, ""); err != nil {
			metrics.CollectCommands(ctx, true)
			newEnv[cmd.EnvPrefix+"_CHANGE"] = "1"
			newEnv[cmd.EnvPrefix+"_CHANGE_OUT"] = out.Change.String()

			out.ChangeFail = true
			err := logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error changing id %s", cmd.ID)).Wrap(err.Errors()...).Wrap(errors.New(out.Change.String())))

			return out, newEnv, err
		}

		newEnv[cmd.EnvPrefix+"_CHANGE"] = "0"
		newEnv[cmd.EnvPrefix+"_CHANGE_OUT"] = out.Change.String()
		metrics.CollectCommands(ctx, false)
	}

	return out, newEnv, logger.Error(ctx, nil)
}
