package commands

import (
	"fmt"
	"testing"

	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/logger"
)

func TestExecOverride(t *testing.T) {
	logger.UseTestLogger(t)

	tests := map[string]struct {
		input []*Exec
		want  Exec
	}{
		"empty": {
			want: Exec{
				AllowOverride: true,
			},
		},
		"nil": {
			input: []*Exec{
				nil,
			},
			want: Exec{
				AllowOverride: true,
			},
		},
		"nested": {
			input: []*Exec{
				{
					AllowOverride: true,
				},
				{
					Command: "hello",
				},
			},
			want: Exec{
				Command: "hello",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			e := Exec{
				AllowOverride: true,
			}
			assert.Equal(t, e.Override(tc.input...), &tc.want)
		})
	}
}

func TestExecRun(t *testing.T) {
	logger.UseTestLogger(t)

	c := cli.Config{}
	c.RunMock()
	c.RunMockErrors([]error{ErrCommandsEmpty})
	c.RunMockOutputs([]string{"hello"})

	e := Exec{
		Command:             `a long command "hello world"`,
		ContainerEntrypoint: "d",
		ContainerImage:      "e",
		ContainerPrivileged: true,
		ContainerUser:       "f",
		ContainerVolumes: []string{
			"volume",
		},
		ContainerWorkDir: "work2",
		Env: []string{
			"hello=world",
		},
		WorkDir: "work1",
	}

	out, err := e.Run(ctx, c, "script", "stdin")
	assert.HasErr(t, err, ErrCommandsEmpty)
	assert.Equal(t, out, "hello")

	inputs := c.RunMockInputs()

	assert.Equal(t, inputs[0].Environment, e.Env)

	cr, _ := cli.GetContainerRuntime()

	assert.Equal(t, inputs[0].Exec, fmt.Sprintf("/usr/bin/%s run -i --rm -e hello=world --entrypoint d --privileged -u f -v volume -w work2 e a long command hello world script", cr))
	assert.Equal(t, inputs[0].WorkDir, e.WorkDir)

	e.ContainerImage = ""
	e.Run(ctx, c, "script", "stdin")
	inputs = c.RunMockInputs()

	assert.Equal(t, inputs[0].Environment, e.Env)
	assert.Equal(t, inputs[0].Exec, "a long command hello world script")
	assert.Equal(t, inputs[0].WorkDir, e.WorkDir)
}
