package pattern

import (
	"context"
	"os"
	"testing"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

func TestTest(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.CLI.RunMock()
	c.Sources["test"].Exec.Command = "test"

	os.MkdirAll("testdata", 0700)
	os.WriteFile("testdata/main.jsonnet", []byte(`
{
	build: [
		{
			change: "change",
			check: "check",
			remove: "remove",
			id: "a",
		}
	],
	run: [
		{
			change: "change",
			check: "check",
			remove: "remove",
			id: "a",
		}
	]
}
`), 0600)

	c.CLI.RunMockErrors([]error{
		ErrBuildEmpty,
		ErrBuildEmpty,
		ErrBuildEmpty,
	})

	r, err := Test(ctx, c, "/asdfasdjzcbjzxkbjxcb", true)
	assert.HasErr(t, err, errs.ErrReceiver)
	assert.Equal(t, r, nil)

	r, err = Test(ctx, c, "testdata", true)
	assert.HasErr(t, err, nil)
	assert.Equal(t, r, types.Results{
		"testdata/main.jsonnet:a": {
			"change had errors", "check still failed after change",
			"check did not fail after remove",
		},
	})

	r, err = Test(ctx, c, "testdata/main.jsonnet", true)
	assert.HasErr(t, err, nil)
	assert.Equal(t, r, types.Results{})

	os.RemoveAll("testdata")
}
