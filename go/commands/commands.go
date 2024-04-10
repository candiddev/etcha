// Package commands contains functions for running and validating Commands.
package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"regexp"
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

// Diff compares two commands and returns the removed Commands.
func (cmds Commands) Diff(old Commands) (removeBefore Commands, removeAfter Commands) {
	removeBefore = Commands{}
	removeAfter = Commands{}

	// Figure out what changed
	for _, cmd := range old {
		match := false

		for _, newV := range cmds {
			if cmd.ID == newV.ID {
				match = newV.ChangeIgnore || cmd.Change == newV.Change
				b, a := newV.Commands.Diff(cmd.Commands)
				removeBefore = append(removeBefore, b...)
				removeAfter = append(removeAfter, a...)
				cmd.Commands = nil

				break
			}
		}

		if !match {
			if len(cmd.Commands) > 0 {
				b, a := Commands{}.Diff(cmd.Commands)
				removeBefore = append(removeBefore, b...)
				removeAfter = append(removeAfter, a...)
			}

			if cmd.Remove != "" {
				if cmd.RemoveAfter {
					removeAfter = append(removeAfter, cmd)
				} else {
					removeBefore = append(removeBefore, cmd)
				}
			}
		}
	}

	return removeBefore, removeAfter
}

// Count returns the total number of Commands.
func (cmds Commands) Count() int {
	c := 0

	for i := range cmds {
		c++
		c += cmds[i].Commands.Count()
	}

	return c
}

// CommandsRunOpts is a list of options for Commands.Run.
type CommandsRunOpts struct { //nolint:revive
	/* Run Commands in Check mode */
	Check bool

	/* A list of EnvVars to add to Command */
	Env types.EnvVars

	/* ParentID of Commands */
	ParentID string

	/* ParentIDFilter to limit Commands being ran */
	ParentIDFilter *regexp.Regexp

	/* Run in Remove mode */
	Remove bool
}

// Run the commands, either as change (default) or remove, and optionally as check only.
func (cmds Commands) Run(ctx context.Context, c cli.Config, exe *Exec, opts CommandsRunOpts) (out Outputs, err errs.Err) { //nolint:gocognit,gocyclo
	cout := Outputs{}

	var exec Exec
	if exe != nil {
		exec = *exe
	}

	if opts.Remove {
		c := Commands{}

		for i := range cmds {
			c = append(c, cmds[len(cmds)-1-i])
		}

		cmds = c
	}

	for i, cmd := range cmds {
		if len(cmd.Commands) > 0 {
			var o Outputs

			if opts.ParentID == "" {
				opts.ParentID = cmd.ID
			} else {
				opts.ParentID = fmt.Sprintf("%s > %s", opts.ParentID, cmd.ID)
			}

			ctx = metrics.SetCommandParentID(ctx, opts.ParentID)

			o, err = cmd.Commands.Run(ctx, c, exe, opts)

			cout = append(cout, o...)

			if err != nil {
				return cout, logger.Error(ctx, err)
			}
		} else {
			if opts.ParentIDFilter != nil && opts.ParentIDFilter.String() != "" && !opts.ParentIDFilter.MatchString(opts.ParentID) {
				continue
			}

			var out *Output

			out, opts.Env, err = cmd.Run(ctx, c, exec, CommandRunOpts{
				Check:    opts.Check,
				Env:      opts.Env,
				ParentID: opts.ParentID,
				Remove:   opts.Remove,
			})

			cout = append(cout, out)

			if err != nil {
				if !opts.Check && !opts.Remove {
					// Parse events
					for k := range cmd.OnFail {
						if strings.HasPrefix(cmd.OnFail[k], "etcha:") {
							out.Events = append(out.Events, strings.ReplaceAll(cmd.OnFail[k], "etcha:", ""))

							continue
						}
					}

					// Match commands
					for j := i + 1; j < len(cmds); j++ {
						cfg := exec.Override(cmds[j].Exec)
						if cfg.Env == nil {
							cfg.Env = types.EnvVars{}
						}

						maps.Copy(cfg.Env, opts.Env)

						run := cmds[j].Always
						if run {
							logger.Info(ctx, "Always changing "+cmds[j].ID)
						} else {
							for k := range cmd.OnFail {
								r, err := regexp.Compile(cmd.OnFail[k])
								if err != nil { //nolint:revive
									return cout, logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error parsing onFail %s for %s: %w", cmd.OnFail[k], cmd.ID, err)))
								}

								if r.MatchString(cmds[j].ID) { //nolint:revive
									run = true

									logger.Info(ctx, fmt.Sprintf("Triggering %s via %s.onFail", cmds[j].ID, cmd.ID))

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

							if out.Change, e = cfg.Run(ctx, c, cmds[j].Change); e != nil { //nolint:revive
								logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error changing id %s", cmds[j].ID)).Wrap(e.Unwrap()...), out.Change.String()) //nolint:errcheck
							}
						}
					}
				}

				return cout, logger.Error(ctx, err)
			}

			if !opts.Check {
				if !opts.Remove && out.Changed {
					for _, id := range cmd.OnChange {
						if strings.HasPrefix(id, "etcha:") {
							switch id {
							case "etcha:stderr":
								fmt.Fprintln(logger.Stderr, out.Change) //nolint:forbidigo
							case "etcha:stdout":
								fmt.Fprintln(logger.Stdout, out.Change) //nolint:forbidigo
							}

							out.Events = append(out.Events, strings.ReplaceAll(id, "etcha:", ""))

							continue
						}

						r, err := regexp.Compile(id)
						if err != nil {
							return cout, logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error parsing onChange %s for %s: %w", id, cmd.ID, err)))
						}

						for i := range cmds {
							if r.MatchString(cmds[i].ID) {
								cmds[i].ChangedBy = append(cmds[i].ChangedBy, cmd.ID)
							}
						}
					}
				}

				if opts.Remove && out.Removed {
					for _, id := range cmd.OnRemove {
						if strings.HasPrefix(id, "etcha:") {
							switch id {
							case "etcha:stderr":
								fmt.Fprint(logger.Stderr, out.Remove) //nolint:forbidigo
							case "etcha:stdout":
								fmt.Fprint(logger.Stdout, out.Remove) //nolint:forbidigo
							}

							out.Events = append(out.Events, strings.ReplaceAll(id, "etcha:", ""))

							continue
						}

						r, err := regexp.Compile(id)
						if err != nil {
							return cout, logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error parsing onRemove %s for %s: %w", id, cmd.ID, err)))
						}

						for i := range cmds {
							if r.MatchString(cmds[i].ID) {
								cmds[i].RemovedBy = append(cmds[i].RemovedBy, cmd.ID)
							}
						}
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

	return cmds.validate()
}

// Validate checks a list of Commands for formatting errors.
func (cmds Commands) validate() error { //nolint:gocognit
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

		for _, id := range append(append(cmd.OnChange, append([]string{"fail"}, cmd.OnFail...)...), append([]string{"remove"}, cmd.OnRemove...)...) {
			if id == "fail" {
				target = "onFail"

				continue
			} else if id == "remove" {
				target = "onRemove" //nolint:goconst

				continue
			}

			if cmd.ID == id {
				r[cmd.ID] = append(r[cmd.ID], ErrCommandsSelfTarget.Error())
			}

			if strings.HasPrefix(id, "etcha:") {
				continue
			}

			match := false

			reg, err := regexp.Compile(id)
			if err != nil {
				r[cmd.ID] = append(r[cmd.ID], fmt.Sprintf("error compiling target %s: %s", id, err))

				continue
			}

			for j, cmd := range cmds {
				if reg.MatchString(cmd.ID) {
					if (j < i && target != "onRemove") || (i > j && target == "onRemove") {
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
		return fmt.Errorf("%w: %s", ErrCommandsValidate, strings.Join(r.Show(), "\n"))
	}

	return nil
}
