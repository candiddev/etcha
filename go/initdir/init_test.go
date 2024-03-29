package initdir

import (
	"os"
	"strings"
	"testing"

	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

func TestInit(t *testing.T) {
	logger.UseTestLogger(t)

	cli.BuildVersion = "v2023.09.28"

	assert.HasErr(t, Init(ctx, "/"), errs.ErrReceiver)
	assert.HasErr(t, Init(ctx, "init"), nil)
	assert.HasErr(t, Init(ctx, "init"), nil)

	r, _ := os.ReadFile("init/lib/etcha/native.libsonnet")

	assert.Equal(t, strings.HasPrefix(string(r), "// Generated by Etcha v2023.09.28"), true)

	f, _ := os.ReadFile("patterns/lib_test.jsonnet")
	os.WriteFile("init/patterns/lib_test.jsonnet", f, 0755)

	d, _ := os.ReadDir("lib/etcha")

	p, err := pattern.ParsePatternFromPath(ctx, c, "", "init/patterns/lib_test.jsonnet")
	assert.HasErr(t, err, nil)
	assert.Equal(t, len(p.Run) > 30, true)
	assert.Equal(t, len(p.Imports.Files), len(d)+2) // 2=lib_test.jsonnet, native.libsonnet

	os.Remove("etcha")
	os.RemoveAll("init")
}
