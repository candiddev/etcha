package pattern

import (
	"context"
	"regexp"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

// DiffRunOpts are options used by DiffRun.
type DiffRunOpts struct {
	/* Run in Check mode */
	Check bool

	/* Never remove */
	NoRemove bool

	/* Source name */
	Source string

	/* ParentIDFilter to limit Commands */
	ParentIDFilter *regexp.Regexp
}

// DiffRun performs a diff against two patterns and runs the changes.
func (p *Pattern) DiffRun(ctx context.Context, c *config.Config, old *Pattern, opts DiffRunOpts) (commands.Outputs, errs.Err) {
	var removeBefore commands.Commands

	var removeAfter commands.Commands

	if old != nil && !opts.NoRemove {
		removeBefore, removeAfter = p.Run.Diff(old.Run)
	}

	o := commands.Outputs{}
	env := types.EnvVars{}

	var err errs.Err

	out, err := removeBefore.Run(ctx, c.CLI, p.RunExec, commands.CommandsRunOpts{
		Check:          opts.Check,
		Env:            env,
		ParentID:       opts.Source,
		ParentIDFilter: opts.ParentIDFilter,
		Remove:         true,
	})
	o = append(o, out...)

	if err != nil {
		return o, logger.Error(ctx, err)
	}

	out, err = p.Run.Run(ctx, c.CLI, p.RunExec, commands.CommandsRunOpts{
		Check:          opts.Check,
		Env:            env,
		ParentID:       opts.Source,
		ParentIDFilter: opts.ParentIDFilter,
	})
	o = append(o, out...)

	if err != nil {
		return o, logger.Error(ctx, err)
	}

	out, err = removeAfter.Run(ctx, c.CLI, p.RunExec, commands.CommandsRunOpts{
		Check:          opts.Check,
		Env:            env,
		ParentID:       opts.Source,
		ParentIDFilter: opts.ParentIDFilter,
		Remove:         true,
	})
	o = append(o, out...)

	if err != nil {
		return o, logger.Error(ctx, err)
	}

	return o, logger.Error(ctx, nil)
}
