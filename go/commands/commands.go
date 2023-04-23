// Package commands contains functions for running and validating Commands.
package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

// Commands errors.
var (
	ErrCommandsIDRequired = errors.New("id is a required property for a command")
	ErrCommandsEmpty      = errors.New("received empty commands")
	ErrCommandsSelfTarget = errors.New("change cannot target self")
	ErrCommandsValidate   = errors.New("error validating commands")
)

// Commands are multiple Commands.
type Commands []*Command

// Diff compares two commands and returns the differences.
func (cmds Commands) Diff(old Commands) (change Commands, remove Commands) {
	remove = Commands{}

	// Figure out what changed
	for _, cmd := range old {
		match := false

		for _, newV := range cmds {
			if cmd.ID == newV.ID {
				if cmd.Change == newV.Change && cmd.Check == newV.Check && !newV.Always {
					newV.Check = ""
				}

				match = true

				break
			}
		}

		if !match && cmd.Remove != "" {
			remove = append(remove, cmd)
		}
	}

	return cmds, remove
}

// Run the commands, either as change (default) or remove, and optionally as check only.
func (cmds Commands) Run(ctx context.Context, c cli.Config, env types.EnvVars, exec Exec, m Mode) (out Outputs, err errs.Err) { //nolint:gocognit
	cout := Outputs{}

	if m == ModeRemove {
		c := Commands{}

		for i := range cmds {
			c = append(c, cmds[len(cmds)-1-i])
		}

		cmds = c
	}

	for i, cmd := range cmds {
		var out *Output

		out, env, err = cmd.Run(ctx, c, env, exec, m)
		cout = append(cout, out)

		if err != nil {
			if m == ModeChange {
				// Parse events
				for k := range cmd.OnFail {
					if strings.HasPrefix(cmd.OnFail[k], "etcha:") {
						out.Events = append(out.Events, strings.ReplaceAll(cmd.OnFail[k], "etcha:", ""))

						continue
					}
				}

				// Match commands
				for j := i + 1; j < len(cmds); j++ {
					cfg := exec
					if cfg.Override && cmds[j].Exec != nil {
						cfg = *cmds[j].Exec
					}

					cfg.Environment = append(env.GetEnv(), cfg.Environment...)

					run := cmds[j].Always
					if run {
						logger.Info(ctx, fmt.Sprintf("Always changing %s...", cmds[j].ID))
					} else {
						for k := range cmd.OnFail {
							if cmds[j].ID == cmd.OnFail[k] {
								run = true
								logger.Info(ctx, fmt.Sprintf("Triggering %s via %s.onFail...", cmds[j].ID, cmd.ID))

								break
							}
						}
					}

					if run {
						var e errs.Err

						out := &Output{
							Changed: true,
							ID:      cmds[j].ID,
						}

						cout = append(cout, out)

						if out.Change, e = cfg.Run(ctx, c, "", cmds[j].Change); e != nil {
							logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error changing id %s", cmds[j].ID)).Wrap(err.Errors()...), out.Change.String()) //nolint:errcheck
						}

						metrics.CommandsChanged.WithLabelValues(cmds[j].ID, "0")
					}
				}
			}

			return cout, logger.Error(ctx, err)
		}

		if m == ModeChange && out.Changed {
			for _, id := range cmd.OnChange {
				if strings.HasPrefix(id, "etcha:") {
					out.Events = append(out.Events, strings.ReplaceAll(id, "etcha:", ""))

					continue
				}

				for i := range cmds {
					if cmds[i].ID == id {
						cmds[i].ChangedBy = append(cmds[i].ChangedBy, cmd.ID)
					}
				}
			}
		}
	}

	return cout, logger.Error(ctx, nil)
}

func (cmds *Commands) UnmarshalJSON(v []byte) error {
	type tmpCmds Commands

	a := []any{}
	if err := json.Unmarshal(v, &a); err != nil {
		return err
	}

	a = types.ArrayFlatten(a)

	v, err := json.Marshal(&a)
	if err != nil {
		return err
	}

	tmp := tmpCmds{}
	if err := json.Unmarshal(v, &tmp); err != nil {
		return err
	}

	*cmds = Commands(tmp)

	return nil
}

// Validate checks a list of Commands for formatting errors.
func (cmds Commands) Validate(ctx context.Context) errs.Err { //nolint:gocognit
	if len(cmds) == 0 {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(ErrCommandsEmpty))
	}

	r := types.Results{}

	// Loop through Commands and sanity check them
	for i, cmd := range cmds {
		if cmd.ID == "" {
			r[cmd.ID] = append(r[cmd.ID], fmt.Sprintf("%s:\n%s", ErrCommandsIDRequired, types.JSONToString(cmd)))
		}

		if e := types.EnvValidate(cmd.EnvPrefix); cmd.EnvPrefix != "" && e != nil {
			r[cmd.ID] = append(r[cmd.ID], fmt.Sprintf("invalid environment prefix: %s: %s", cmd.EnvPrefix, e))
		}

		target := "onChange"

		for _, id := range append(cmd.OnChange, append([]string{"fail"}, cmd.OnFail...)...) {
			if id == "fail" {
				target = "onFail"

				continue
			}

			if cmd.ID == id {
				r[cmd.ID] = append(r[cmd.ID], ErrCommandsSelfTarget.Error())
			}

			if strings.HasPrefix(id, "etcha:") {
				continue
			}

			match := false

			for j, cmd := range cmds {
				if cmd.ID == id {
					if j < i {
						r[cmd.ID] = append(r[cmd.ID], fmt.Sprintf("%s target %s has been ran already", target, id))
					}

					match = true

					break
				}
			}

			if !match {
				r[cmd.ID] = append(r[cmd.ID], fmt.Sprintf("%s target %s does not exist", target, id))
			}
		}
	}

	if len(r) > 0 {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(ErrCommandsValidate), strings.Join(r.Show(), "\n"))
	}

	return logger.Error(ctx, nil)
}
