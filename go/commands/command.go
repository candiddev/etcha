package commands

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

// Command is a configuration to run.
type Command struct {
	Always       bool              `json:"always,omitempty"`
	Change       string            `json:"change,omitempty"`
	ChangeIgnore bool              `json:"changeIgnore,omitempty"`
	ChangedBy    []string          `json:"-"`
	Check        string            `json:"check,omitempty"`
	Commands     Commands          `json:"commands,omitempty"`
	EnvPrefix    string            `json:"envPrefix"`
	Exec         *Exec             `json:"exec,omitempty"`
	ID           string            `json:"id"`
	OnChange     types.SliceString `json:"onChange,omitempty"`
	OnFail       types.SliceString `json:"onFail,omitempty"`
	OnRemove     types.SliceString `json:"onRemove,omitempty"`
	Remove       string            `json:"remove,omitempty"`
	RemoveAfter  bool              `json:"removeAfter,omitempty"`
	RemovedBy    []string          `json:"-"`
	Stdin        string            `json:"stdin"`
}

// CommandRunOpts is options for Command.Run.
type CommandRunOpts struct {
	/* Run in Check mode */
	Check bool

	/* A list of EnvVars to add to Command */
	Env types.EnvVars

	/* ParentID of Command */
	ParentID string

	/* Run in Remove mode */
	Remove bool
}

// Run will run the Command script for the given Mode.
func (cmd *Command) Run(ctx context.Context, c cli.Config, exec Exec, opts CommandRunOpts) (out *Output, newEnv types.EnvVars, err errs.Err) { //nolint:gocognit,gocyclo
	cfg := exec.Override(cmd.Exec)
	ctx = metrics.SetCommandID(ctx, cmd.ID)

	if cfg.Env == nil {
		cfg.Env = types.EnvVars{}
	}

	if opts.Env == nil {
		newEnv = types.EnvVars{}
	} else {
		newEnv = opts.Env
	}

	maps.Copy(cfg.Env, opts.Env)

	out = &Output{
		ID:       cmd.ID,
		ParentID: opts.ParentID,
	}

	id := cmd.ID
	if opts.ParentID != "" {
		id = fmt.Sprintf("%s > %s", opts.ParentID, cmd.ID)
	}

	ctx = metrics.SetCommandMode(ctx, metrics.CommandModeCheck)

	if cmd.Check == "" && !cmd.Always && ((!opts.Remove && len(cmd.ChangedBy) == 0) || (opts.Remove && len(cmd.RemovedBy) == 0)) {
		newEnv[cmd.EnvPrefix+"_CHECK"] = "0" //nolint:goconst

		metrics.CollectCommands(ctx, false)

		return out, newEnv, nil
	}

	ch := fmt.Sprintf("Changing %s...", id)
	if opts.Remove {
		ch = fmt.Sprintf("Removing %s...", id)
	}

	switch {
	case cmd.Always:
		ch = fmt.Sprintf("Always changing %s...", id)
		if opts.Remove {
			ch = fmt.Sprintf("Always removing %s...", id)
		}
	case !opts.Remove && len(cmd.ChangedBy) > 0:
		ch = fmt.Sprintf("Triggering %s via %s...", id, strings.Join(cmd.ChangedBy, ", "))
	case opts.Remove && len(cmd.RemovedBy) > 0:
		ch = fmt.Sprintf("Triggering %s via %s...", id, strings.Join(cmd.RemovedBy, ", "))
	default:
		out.Checked = true

		logger.Debug(ctx, fmt.Sprintf("Checking %s...", id))

		out.Check, err = cfg.Run(ctx, c, cmd.Check, cmd.Stdin)

		newEnv[cmd.EnvPrefix+"_CHECK_OUT"] = out.Check.String()
		if cmd.EnvPrefix != "" {
			newEnv[cmd.EnvPrefix] = out.Check.String()
		}

		if (!opts.Remove && err == nil) || (opts.Remove && err != nil) {
			newEnv[cmd.EnvPrefix+"_CHECK"] = "0"

			metrics.CollectCommands(ctx, false)

			return out, newEnv, nil //nolint:nilerr
		}

		logger.Debug(ctx, out.Check.String())
	}

	if opts.Remove {
		out.CheckFailRemove = true
	} else {
		out.CheckFailChange = true
	}

	cmd.ChangedBy = nil
	cmd.RemovedBy = nil

	metrics.CollectCommands(ctx, true)

	newEnv[cmd.EnvPrefix+"_CHECK"] = "1"

	if opts.Check || (!opts.Remove && cmd.Change == "") || (opts.Remove && cmd.Remove == "") {
		if opts.Check && ((!opts.Remove && cmd.Change != "") || (opts.Remove && cmd.Remove != "")) {
			logger.Info(ctx, "Check mode: "+ch)
		}

		return out, newEnv, nil
	}

	if opts.Remove {
		ctx = metrics.SetCommandMode(ctx, metrics.CommandModeRemove)
		out.Removed = true
	} else {
		ctx = metrics.SetCommandMode(ctx, metrics.CommandModeChange)
		out.Changed = true
	}

	logger.Info(ctx, ch)

	maps.Copy(cfg.Env, newEnv)

	s := cmd.Change
	if opts.Remove {
		s = cmd.Remove
	}

	o, err := cfg.Run(ctx, c, s, cmd.Stdin)
	if err != nil {
		if opts.Remove {
			out.Remove = o
			out.RemoveFail = true
			newEnv[cmd.EnvPrefix+"_REMOVE"] = "1"
			newEnv[cmd.EnvPrefix+"_REMOVE_OUT"] = out.Remove.String()

			err = logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error removing id %s", id)).Wrap(err.Errors()...), out.Remove.String())
		} else {
			out.Change = o
			out.ChangeFail = true
			newEnv[cmd.EnvPrefix+"_CHANGE"] = "1"
			newEnv[cmd.EnvPrefix+"_CHANGE_OUT"] = out.Change.String()

			err = logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error changing id %s", id)).Wrap(err.Errors()...).Wrap(errors.New(out.Change.String())))
		}

		metrics.CollectCommands(ctx, true)

		return out, newEnv, err
	}

	if opts.Remove {
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
