package commands

import (
	"context"

	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/types"
)

const (
	testChangeError = "change had errors"
	testRemoveError = "remove had errors"
	testCheck       = "check did not fail after remove"
	testCheckChange = "check still failed after change"
	testCheckRemove = "check still failed after remove"
)

// Test runs Commands in Test mode.
func (cmds Commands) Test(ctx context.Context, c cli.Config, exec *Exec, runEnv types.EnvVars) types.Results { //nolint:gocognit
	r := types.Results{}

	// Run everything
	outChange, _ := cmds.Run(ctx, c, runEnv, exec, false, false)
	changes := []string{}
	removes := []string{}

	for _, o := range outChange {
		if o.Changed && o.Checked {
			changes = append(changes, o.ID)
		}

		if o.ChangeFail {
			r[o.ID] = append(r[o.ID], testChangeError)
		}
	}

	// Run check
	outCheck, _ := cmds.Run(ctx, c, runEnv, exec, true, false)

	for _, o := range outCheck {
		for j := range changes {
			if o.ID == changes[j] && o.CheckFailChange {
				r[o.ID] = append(r[o.ID], testCheckChange)
			}
		}
	}

	// Run remove
	outRemove, _ := cmds.Run(ctx, c, runEnv, exec, false, true)

	for _, o := range outRemove {
		if o.Removed && o.Checked {
			removes = append(removes, o.ID)
		}

		if o.RemoveFail {
			r[o.ID] = append(r[o.ID], testRemoveError)
		}
	}

	// Run check remove
	outCheck, _ = cmds.Run(ctx, c, runEnv, exec, true, true)

	for _, o := range outCheck {
		for j := range removes {
			if o.ID == removes[j] && o.CheckFailRemove {
				r[o.ID] = append(r[o.ID], testCheckRemove)
			}
		}
	}

	// Run check again
	outCheck, _ = cmds.Run(ctx, c, runEnv, exec, true, false)

	for _, o := range outCheck {
		for j := range changes {
			if changes[j] == o.ID && !o.CheckFailChange {
				r[o.ID] = append(r[o.ID], testCheck)
			}
		}
	}

	return r
}
