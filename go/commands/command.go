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
	OnRemove  []string `json:"onRemove,omitempty"`
	Remove    string   `json:"remove,omitempty"`
	RemovedBy []string `json:"-"`
}

// Run will run the Command script for the given Mode.
func (cmd *Command) Run(ctx context.Context, c cli.Config, oldEnv types.EnvVars, exec Exec, check bool, remove bool) (out *Output, newEnv types.EnvVars, err errs.Err) { //nolint:revive,gocognit,gocyclo
	cfg := exec.Override(cmd.Exec)
	cfgEnv := cfg.Env
	ctx = metrics.SetCommandID(ctx, cmd.ID)

	if e := oldEnv.GetEnv(); len(e) > 0 {
		cfg.Env = append(e, cfg.Env...)
	}

	if oldEnv == nil {
		newEnv = types.EnvVars{}
	} else {
		newEnv = oldEnv
	}

	out = &Output{
		ID: cmd.ID,
	}

	ctx = metrics.SetCommandMode(ctx, metrics.CommandModeCheck)

	if cmd.Check == "" && !cmd.Always && ((!remove && len(cmd.ChangedBy) == 0) || (remove && len(cmd.RemovedBy) == 0)) {
		newEnv[cmd.EnvPrefix+"_CHECK"] = "0"

		metrics.CollectCommands(ctx, false)

		return out, newEnv, nil
	}

	ch := fmt.Sprintf("Changing %s...", cmd.ID)
	if remove {
		ch = fmt.Sprintf("Removing %s...", cmd.ID)
	}

	switch {
	case cmd.Always:
		ch = fmt.Sprintf("Always changing %s...", cmd.ID)
		if remove {
			ch = fmt.Sprintf("Always removing %s...", cmd.ID)
		}
	case !remove && len(cmd.ChangedBy) > 0:
		ch = fmt.Sprintf("Triggering %s via %s...", cmd.ID, strings.Join(cmd.ChangedBy, ", "))
	case remove && len(cmd.RemovedBy) > 0:
		ch = fmt.Sprintf("Triggering %s via %s...", cmd.ID, strings.Join(cmd.RemovedBy, ", "))
	default:
		out.Checked = true

		logger.Debug(ctx, fmt.Sprintf("Checking %s...", cmd.ID))

		out.Check, err = cfg.Run(ctx, c, cmd.Check, "")
		if (!remove && err == nil) || (remove && err != nil) {
			newEnv[cmd.EnvPrefix+"_CHECK"] = "0"
			newEnv[cmd.EnvPrefix+"_CHECK_OUT"] = out.Check.String()

			metrics.CollectCommands(ctx, false)

			return out, newEnv, nil //nolint:nilerr
		}

		newEnv[cmd.EnvPrefix+"_CHECK_OUT"] = out.Check.String()
	}

	if remove {
		out.CheckFailRemove = true
	} else {
		out.CheckFailChange = true
	}

	cmd.ChangedBy = nil
	cmd.RemovedBy = nil

	metrics.CollectCommands(ctx, true)

	newEnv[cmd.EnvPrefix+"_CHECK"] = "1"

	if check || (!remove && cmd.Change == "") || (remove && cmd.Remove == "") {
		if check && ((!remove && cmd.Change != "") || (remove && cmd.Remove != "")) {
			logger.Info(ctx, "Check mode: "+ch)
		}

		return out, newEnv, nil
	}

	if remove {
		ctx = metrics.SetCommandMode(ctx, metrics.CommandModeRemove)
		out.Removed = true
	} else {
		ctx = metrics.SetCommandMode(ctx, metrics.CommandModeChange)
		out.Changed = true
	}

	logger.Info(ctx, ch)

	cfg.Env = append(cfgEnv, newEnv.GetEnv()...) //nolint:gocritic

	s := cmd.Change
	if remove {
		s = cmd.Remove
	}

	o, err := cfg.Run(ctx, c, s, "")
	if err != nil {
		if remove {
			out.Remove = o
			out.RemoveFail = true
			newEnv[cmd.EnvPrefix+"_REMOVE"] = "1"
			newEnv[cmd.EnvPrefix+"_REMOVE_OUT"] = out.Remove.String()

			err = logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error removing id %s", cmd.ID)).Wrap(err.Errors()...), out.Remove.String())
		} else {
			out.Change = o
			out.ChangeFail = true
			newEnv[cmd.EnvPrefix+"_CHANGE"] = "1"
			newEnv[cmd.EnvPrefix+"_CHANGE_OUT"] = out.Change.String()

			err = logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error changing id %s", cmd.ID)).Wrap(err.Errors()...).Wrap(errors.New(out.Change.String())))
		}

		metrics.CollectCommands(ctx, true)

		return out, newEnv, err
	}

	if remove {
		out.Remove = o
		newEnv[cmd.EnvPrefix+"_REMOVE"] = "0"
		newEnv[cmd.EnvPrefix+"_REMOVE_OUT"] = out.Remove.String()
	} else {
		out.Change = o
		newEnv[cmd.EnvPrefix+"_CHANGE"] = "0"
		newEnv[cmd.EnvPrefix+"_CHANGE_OUT"] = out.Change.String()
	}

	metrics.CollectCommands(ctx, false)

	return out, newEnv, logger.Error(ctx, nil)
}
