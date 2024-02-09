package pattern

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

// Test performs build and run testing against directory or single pattern.
func Test(ctx context.Context, c *config.Config, path string, testBuild bool) (types.Results, errs.Err) {
	l := types.Results{}

	f, e := os.Stat(path)
	if e != nil {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error opening path"), e))
	}

	if f.IsDir() {
		if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.Type().IsDir() {
				if p := filepath.Ext(path); p != ".jsonnet" {
					return nil
				}

				p, err := ParsePatternFromPath(ctx, c, "test", path)
				if err != nil {
					return err
				}

				r := p.Test(ctx, c, testBuild)
				for k, v := range r {
					l[path+":"+k] = v
				}
			}

			if err != nil {
				return errs.ErrReceiver.Wrap(err)
			}

			return nil
		}); err != nil {
			return l, logger.Error(ctx, err.(errs.Err))
		}
	} else {
		p, err := ParsePatternFromPath(ctx, c, "test", path)
		if err != nil {
			return nil, logger.Error(ctx, err)
		}

		r := p.Test(ctx, c, testBuild)
		for k, v := range r {
			l[path+":"+k] = v
		}
	}

	return l, logger.Error(ctx, nil)
}

// Test performs build and run testing against a Pattern.
func (p *Pattern) Test(ctx context.Context, c *config.Config, testBuild bool) types.Results {
	r := types.Results{}

	if logger.GetLevel(ctx) != logger.LevelDebug {
		ctx = logger.SetLevel(ctx, logger.LevelNone)
	}

	if !testBuild {
		l := p.Build.Test(ctx, c.CLI, p.BuildExec, nil)

		for k, v := range l {
			r[k] = v
		}
	}

	l := p.Run.Test(ctx, c.CLI, p.RunExec, nil)

	if len(l) > 0 {
		for k, v := range l {
			r[k] = v
		}
	}

	return r
}
