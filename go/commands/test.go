package commands

import (
	"context"

	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/types"
)

const (
	testChangeError = "change had errors"
	testRemoveError = "remove had errors"
	testCheckChange = "check still failed after change"
	testCheckRemove = "check did not fail after remove"
)

// Test runs Commands in Test mode.
func (cmds Commands) Test(ctx context.Context, c cli.Config, exec *Exec, runEnv types.EnvVars) types.Results {
	r := types.Results{}

	// Run everything
	outChange, _ := cmds.Run(ctx, c, runEnv, exec, ModeChange)
	changes := []string{}

	for _, o := range outChange {
		if o.Changed && o.Checked {
			changes = append(changes, o.ID)

			if o.ChangeFail {
				r[o.ID] = append(r[o.ID], testChangeError)
			}
		}
	}

	// Run check
	outCheck, _ := cmds.Run(ctx, c, runEnv, exec, ModeCheck)

	for _, o := range outCheck {
		for j := range changes {
			if o.ID == changes[j] && o.CheckFail {
				r[o.ID] = append(r[o.ID], testCheckChange)
			}
		}
	}

	// Run remove
	outRemove, _ := cmds.Run(ctx, c, runEnv, exec, ModeRemove)

	for _, o := range outRemove {
		if o.RemoveFail {
			r[o.ID] = append(r[o.ID], testRemoveError)
		}
	}

	// Run check again
	outCheck, _ = cmds.Run(ctx, c, runEnv, exec, ModeCheck)

	for _, o := range outCheck {
		for j := range changes {
			if changes[j] == o.ID && !o.CheckFail {
				r[o.ID] = append(r[o.ID], testCheckRemove)
			}
		}
	}

	return r
}
