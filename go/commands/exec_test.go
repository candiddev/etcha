package commands

import (
	"strings"
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
		Command:             "a long command",
		ContainerEntrypoint: "d",
		ContainerImage:      "e",
		ContainerPrivileged: true,
		ContainerUser:       "f",
		ContainerVolumes: []string{
			"volume",
		},
		ContainerWorkDir: "work2",
		Environment: []string{
			"hello=world",
		},
		WorkDir: "work1",
	}

	out, err := e.Run(ctx, c, "script", "stdin")
	assert.HasErr(t, err, ErrCommandsEmpty)
	assert.Equal(t, out, "hello")

	inputs := c.RunMockInputs()

	assert.Equal(t, inputs[0].Environment, e.Environment)
	assert.Equal(t, strings.Contains(inputs[0].Exec, "-ehello=world --entrypoint d --privileged -u f -v volume -w work2 e a long command script"), true)
	assert.Equal(t, inputs[0].WorkDir, e.WorkDir)

	e.ContainerImage = ""
	e.Run(ctx, c, "script", "stdin")
	inputs = c.RunMockInputs()

	assert.Equal(t, inputs[0].Environment, e.Environment)
	assert.Equal(t, inputs[0].Exec, "a long command script")
	assert.Equal(t, inputs[0].WorkDir, e.WorkDir)
}
