package pattern

import (
	"context"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

// DiffRun performs a diff against two patterns and runs the changes.
func (p *Pattern) DiffRun(ctx context.Context, c *config.Config, old *Pattern, check, noRemove, runAll bool) (commands.Outputs, errs.Err) { //nolint:revive
	var change commands.Commands

	var remove commands.Commands

	if old != nil {
		change, remove = p.Run.Diff(old.Run)
	}

	if noRemove {
		remove = commands.Commands{}
	}

	if runAll || old == nil {
		change = p.Run
	}

	diff := false
	o := commands.Outputs{}

	for i := range change {
		if change[i].Check != "" || change[i].Always {
			diff = true

			break
		}
	}

	if !diff && len(remove) == 0 {
		return o, nil
	}

	var err errs.Err

	o, err = change.Run(ctx, c.CLI, p.RunEnv, p.RunExec, check, false)

	if err != nil {
		return o, logger.Error(ctx, err)
	}

	removeOut, err := remove.Run(ctx, c.CLI, p.RunEnv, p.RunExec, check, true)

	o = append(o, removeOut...)

	if err != nil {
		return o, logger.Error(ctx, err)
	}

	return o, logger.Error(ctx, nil)
}
