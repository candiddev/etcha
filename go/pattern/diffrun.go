package pattern

import (
	"context"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

// DiffRun performs a diff against two patterns and runs the changes.
func (p *Pattern) DiffRun(ctx context.Context, c *config.Config, old *Pattern, check bool) (commands.Outputs, errs.Err) {
	var change commands.Commands

	var remove commands.Commands

	if old == nil {
		change = p.Run
	} else {
		change, remove = p.Run.Diff(old.Run)
	}

	diff := false
	o := commands.Outputs{}

	for i := range change {
		if change[i].Check != "" {
			diff = true

			break
		}
	}

	if !diff && len(remove) == 0 {
		return o, nil
	}

	if old != nil {
		change, remove = p.Run.Diff(old.Run)
	}

	m := commands.ModeChange
	if check {
		m = commands.ModeCheck
	}

	var err errs.Err

	o, err = change.Run(ctx, c.CLI, p.RunEnv, p.Exec, m)

	if err != nil {
		return o, logger.Error(ctx, err)
	}

	if !check {
		removeOut, err := remove.Run(ctx, c.CLI, p.RunEnv, p.Exec, commands.ModeRemove)

		o = append(o, removeOut...)

		if err != nil {
			return o, logger.Error(ctx, err)
		}
	}

	return o, logger.Error(ctx, nil)
}
