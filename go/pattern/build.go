package pattern

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var ErrBuildEmpty = errors.New("received empty build and runtime commands")
var ErrBuildWriteJWT = errors.New("error writing jwt")

// BuildSign runs the build commands in a template from path and creates a JWT.
func (p *Pattern) BuildSign(ctx context.Context, c *config.Config, destination string) errs.Err {
	buildManifest := ""
	runEnv := map[string]string{}

	f, e := os.Create(destination + ".tmp") //nolint:goconst
	if e != nil {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(ErrBuildWriteJWT, e))
	}

	if len(p.Build) > 0 {
		out, err := p.Build.Run(ctx, c.CLI, nil, p.BuildExec, false, false)
		if err != nil {
			f.Close()
			os.Remove(destination + ".tmp") //nolint:errcheck

			return logger.Error(ctx, err)
		}

		for _, event := range out.Events() {
			if event.Name == "buildManifest" {
				for _, output := range event.Outputs {
					buildManifest += output.Change.String() + "\n"
				}

				continue
			}

			if s := strings.Split(event.Name, "runEnv_"); len(s) == 2 {
				for _, output := range event.Outputs {
					runEnv[s[1]] = output.Change.String()
				}
			}
		}
	}

	out, err := p.Sign(ctx, c, buildManifest, runEnv)
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
