package commands

import (
	"strings"
	"testing"

	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/logger"
)

func TestExecRun(t *testing.T) {
	logger.UseTestLogger(t)

	c := cli.Config{}
	c.RunMock()
	c.RunMockErrors([]error{ErrCommandsEmpty})
	c.RunMockOutputs([]string{"hello"})

	e := Exec{
		Command:             "c",
		ContainerEntrypoint: "d",
		ContainerImage:      "e",
		ContainerPrivileged: true,
		ContainerUser:       "f",
		ContainerVolumes: []string{
			"volume",
		},
		Environment: []string{
			"hello=world",
		},
		WorkDir: "work",
	}

	out, err := e.Run(ctx, c, "stdin", "script")
	assert.HasErr(t, err, ErrCommandsEmpty)
	assert.Equal(t, out, "hello")

	inputs := c.RunMockInputs()

	assert.Equal(t, inputs[0].Environment, e.Environment)
	assert.Equal(t, strings.Contains(inputs[0].Exec, "-ehello=world --entrypoint d --privileged -u f -v volume -w work e c script"), true)
	assert.Equal(t, inputs[0].WorkDir, e.WorkDir)
}
