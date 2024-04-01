package pattern

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var ErrBuildEmpty = errors.New("received empty build and runtime commands")
var ErrBuildWriteJWT = errors.New("error writing jwt")

// BuildRun runs build commands.
func (p *Pattern) BuildRun(ctx context.Context, c *config.Config) (buildManifest string, runVars map[string]any, err errs.Err) {
	runVars = p.RunVars
	if runVars == nil {
		runVars = map[string]any{}
	}

	if len(p.Build) > 0 {
		out, err := p.Build.Run(ctx, c.CLI, p.BuildExec, commands.CommandsRunOpts{
			ParentID: "build",
		})
		if err != nil {
			return "", nil, logger.Error(ctx, err)
		}

		for _, event := range out.Events() {
			if event.Name == "buildManifest" {
				for _, output := range event.Outputs {
					buildManifest += output.Change.String() + "\n"
				}

				continue
			}

			if s := strings.Split(event.Name, "runVar_"); len(s) == 2 {
				for _, output := range event.Outputs {
					runVars[s[1]] = output.Change.String()
				}
			}
		}
	}

	return buildManifest, runVars, nil
}

// BuildSign runs the build commands in a template from path and creates a JWT.
func (p *Pattern) BuildSign(ctx context.Context, c *config.Config, destination string) errs.Err {
	f, e := os.Create(destination + ".tmp")
	if e != nil {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(ErrBuildWriteJWT, e))
	}

	buildManifest, runVars, err := p.BuildRun(ctx, c)
	if err != nil {
		f.Close()
		os.Remove(destination + ".tmp") //nolint:errcheck

		return logger.Error(ctx, err)
	}

	out, _, err := p.Sign(ctx, c, buildManifest, runVars)
	if err != nil {
		f.Close()
		os.Remove(destination + ".tmp") //nolint:errcheck

		return logger.Error(ctx, err)
	}

	if _, err := f.WriteString(out); err != nil {
		f.Close()
		os.Remove(destination + ".tmp") //nolint:errcheck

		return logger.Error(ctx, errs.ErrReceiver.Wrap(ErrBuildWriteJWT, err))
	}

	f.Close()

	if err := os.Rename(destination+".tmp", destination); err != nil {
		os.Remove(destination + ".tmp") //nolint:errcheck

		return logger.Error(ctx, errs.ErrReceiver.Wrap(ErrBuildWriteJWT, err))
	}

	return logger.Error(ctx, nil)
}
