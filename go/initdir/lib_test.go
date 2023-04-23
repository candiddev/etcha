package initdir

import (
	"context"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

var c = config.Default()

var ctx = context.Background()

func TestLib(t *testing.T) {
	logger.UseTestLogger(t)

	i := jsonnet.Imports{}

	f, err := os.ReadFile("patterns/lib_test.jsonnet")
	assert.HasErr(t, err, nil)

	i.Entrypoint = "/initdir/patterns/lib_test.jsonnet"
	i.Files = map[string]string{
		"/initdir/patterns/lib_test.jsonnet":  string(f),
		"/initdir/lib/etcha/native.libsonnet": jsonnet.Native,
	}

	err = fs.WalkDir(lib, "lib", func(path string, d fs.DirEntry, err error) error {
		if d != nil && !d.Type().IsDir() {
			f, err := lib.ReadFile(path)
			if err != nil {
				return err
			}

			i.Files[strings.Replace(path, "lib", "/initdir/lib", 1)] = string(f)
		}

		return nil
	})

	assert.HasErr(t, err, nil)

	r := jsonnet.NewRender(ctx, c)
	r.Import(&i)
	assert.HasErr(t, r.Fmt(ctx), nil)

	p, err := pattern.ParsePatternFromImports(ctx, c, "", &i)
	assert.HasErr(t, err, nil)

	os.Mkdir("etcha", 0700)

	res := p.Test(ctx, c, false)
	assert.Equal(t, res, types.Results{})

	os.RemoveAll("etcha")
}
