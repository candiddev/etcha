package commands

import (
	"context"
	"strings"
	"testing"

	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

var ctx = context.Background()

var cmdA = &Command{
	Check:  "checkA",
	Change: "changeA",
	Remove: "removeA",
	ID:     "a",
	OnChange: []string{
		"e",
		"etcha:hello",
		"etcha:stderr",
		"etcha:stdout",
	},
	OnFail: []string{
		"f",
		"etcha:fail",
	},
}
var cmdB = &Command{
	Change: "changeB",
	Remove: "removeB",
	ID:     "b",
}
var cmdC = &Command{
	Check:  "checkC",
	Remove: "removeC",
	ID:     "c",
	OnChange: []string{
		"f",
	},
}
var cmdE = &Command{
	Change: "changeE",
	ID:     "e",
}
var cmdF = &Command{
	Exec: &Exec{
		Command: "hello",
	},
	Change: "changeF",
	ID:     "f",
}
var cmdG = &Command{
	Always:    true,
	Change:    "changeG",
	EnvPrefix: "g",
	ID:        "g",
	Remove:    "removeG",
}

func TestCommandsDiff(t *testing.T) {
	a := *cmdA
	b := *cmdB
	c := *cmdC
	g := *cmdG

	change, remove := Commands{
		&a,
		&b,
		&g,
	}.Diff(Commands{
		&a,
		&c,
		&g,
	})
	assert.Equal(t, change, Commands{
		&Command{
			Change: "changeA",
			Remove: "removeA",
			ID:     "a",
			OnChange: []string{
				"e",
				"etcha:hello",
				"etcha:stderr",
				"etcha:stdout",
			},
			OnFail: []string{
				"f",
				"etcha:fail",
			},
		},
		&b,
		&g,
	})
	assert.Equal(t, remove, Commands{
		&c,
	})
}

func TestCommandsRun(t *testing.T) {
	logger.UseTestLogger(t)

	tests := map[string]struct {
		cmds         Commands
		env          types.EnvVars
		execOverride bool
		mockErrs     []error
		mockOutputs  []string
		mode         Mode
		wantErr      error
		wantInputs   []cli.RunMockInput
		wantOut      string
		wantOutputs  Outputs
	}{
		"remove": {
			cmds: Commands{
				cmdA,
				cmdG,
			},
			env: types.EnvVars{
				"hello": "world",
			},
			mockOutputs: []string{"output1", "output2"},
			mode:        ModeRemove,
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{"hello=world"},
					Exec:        "removeG",
				},
				{
					Environment: []string{
						"g_REMOVE=0",
						"g_REMOVE_OUT=output1",
						"hello=world",
					},
					Exec: "removeA",
				},
			},
			wantOut: "INFO  commands/command.go:63\nRemoving g...\nINFO  commands/command.go:63\nRemoving a...\n",
			wantOutputs: Outputs{
				&Output{
					ID:      "g",
					Remove:  "output1",
					Removed: true,
				},
				&Output{
					ID:      "a",
					Remove:  "output2",
					Removed: true,
				},
			},
		},
		"change": {
			cmds: Commands{
				cmdA,
				cmdC,
				cmdE,
				cmdF,
				cmdG,
			},
			mockErrs: []error{
				ErrCommandsEmpty,
			},
			mockOutputs: []string{
				"a",
				"a",
			},
			mode: ModeChange,
			wantInputs: []cli.RunMockInput{
				{
					Exec: "checkA",
				},
				{
					Environment: []string{"_CHECK=1", "_CHECK_OUT=a"},
					Exec:        "changeA",
				},
				{
					Environment: []string{"_CHANGE=0", "_CHANGE_OUT=a", "_CHECK=1", "_CHECK_OUT=a"},
					Exec:        "checkC",
				},
				{
					Environment: []string{"_CHANGE=0", "_CHANGE_OUT=a", "_CHECK=1", "_CHECK_OUT="},
					Exec:        "changeE",
				},
				{
					Environment: []string{"_CHANGE=0", "_CHANGE_OUT=", "_CHECK=0", "_CHECK_OUT=",
						"g_CHECK=1"},
					Exec: "changeG",
				},
			},
			wantOut: "INFO  commands/command.go:131\nChanging a...\naaINFO  commands/command.go:131\nTriggering e via a...\nINFO  commands/command.go:131\nAlways changing g...\n",
			wantOutputs: Outputs{
				&Output{
					Change:    "a",
					Changed:   true,
					Check:     "a",
					CheckFail: true,
					Checked:   true,
					ID:        "a",
					Events:    []string{"hello", "stderr", "stdout"},
				},
				&Output{
					Checked: true,
					ID:      "c",
				},
				&Output{
					Changed: true,
					Checked: false,
					ID:      "e",
				},
				&Output{
					ID: "f",
				},
				&Output{
					Changed: true,
					ID:      "g",
				},
			},
		},
		"change_fail": {
			cmds: Commands{
				cmdA,
				cmdC,
				cmdE,
				cmdF,
				cmdG,
			},
			execOverride: true,
			mockErrs: []error{
				ErrCommandsEmpty,
				ErrCommandsIDRequired,
			},
			mode:    ModeChange,
			wantErr: ErrCommandsIDRequired,
			wantInputs: []cli.RunMockInput{
				{
					Exec: "checkA",
				},
				{
					Environment: []string{"_CHECK=1", "_CHECK_OUT="},
					Exec:        "changeA",
				},
				{
					Environment: []string{"_CHANGE=1", "_CHANGE_OUT=", "_CHECK=1", "_CHECK_OUT="},
					Exec:        "hello changeF",
				},
				{
					Environment: []string{"_CHANGE=1", "_CHANGE_OUT=", "_CHECK=1", "_CHECK_OUT="},
					Exec:        "changeG",
				},
			},
			wantOutputs: Outputs{
				&Output{
					Changed:    true,
					ChangeFail: true,
					CheckFail:  true,
					Checked:    true,
					Events:     []string{"fail"},
					ID:         "a",
				},
				&Output{
					Changed: true,
					ID:      "f",
				},
				&Output{
					Changed: true,
					ID:      "g",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logger.SetStd()
			c := cli.Config{}
			c.RunMock()
			c.RunMockErrors(tc.mockErrs)
			c.RunMockOutputs(tc.mockOutputs)

			out, err := tc.cmds.Run(ctx, c, tc.env, Exec{
				AllowOverride: tc.execOverride,
			}, tc.mode)

			assert.HasErr(t, err, tc.wantErr)
			assert.Equal(t, out, tc.wantOutputs)
			assert.Equal(t, c.RunMockInputs(), tc.wantInputs)

			if tc.wantErr == nil {
				assert.Equal(t, logger.ReadStd(), tc.wantOut)
			}
		})
	}
}

func TestCommandsUnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		input        string
		wantCommands Commands
		wantErr      bool
	}{
		"invalid JSON": {
			input:   `invalid`,
			wantErr: true,
		},
		`good`: {
			input: `[{"id": "1"}, [{"id": "2"}, [{"id": "3"}]]]`,
			wantCommands: Commands{
				{
					ID: "1",
				},
				{
					ID: "2",
				},
				{
					ID: "3",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var c Commands

			assert.Equal(t, c.UnmarshalJSON([]byte(tc.input)) != nil, tc.wantErr)
			assert.Equal(t, c, tc.wantCommands)
		})
	}
}

func TestCommandsValidate(t *testing.T) {
	tests := map[string]struct {
		cmds       Commands
		wantErr    error
		wantOutput string
	}{
		"no_commands": {
			wantErr: ErrCommandsEmpty,
		},
		"empty_command_id": {
			cmds: Commands{
				&Command{
					Check: "a",
				},
			},
			wantErr:    ErrCommandsValidate,
			wantOutput: ErrCommandsIDRequired.Error(),
		},
		"invalid_environment_prefix": {
			cmds: Commands{
				&Command{
					ID:        "a",
					EnvPrefix: "1",
				},
			},
			wantErr:    ErrCommandsValidate,
			wantOutput: "invalid environment",
		},
		"self_target_change": {
			cmds: Commands{
				&Command{
					ID:        "a",
					EnvPrefix: "1",
					OnChange: []string{
						"a",
					},
				},
			},
			wantErr:    ErrCommandsValidate,
			wantOutput: ErrCommandsSelfTarget.Error(),
		},
		"self_target_fail": {
			cmds: Commands{
				&Command{
					ID:        "a",
					EnvPrefix: "1",
					OnFail: []string{
						"a",
					},
				},
			},
			wantErr:    ErrCommandsValidate,
			wantOutput: ErrCommandsSelfTarget.Error(),
		},
		"target_before": {
			cmds: Commands{
				&Command{
					ID: "b",
				},
				&Command{
					ID:        "a",
					EnvPrefix: "1",
					OnFail: []string{
						"b",
					},
				},
			},
			wantErr:    ErrCommandsValidate,
			wantOutput: "has been ran already",
		},
		"target_missing": {
			cmds: Commands{
				&Command{
					ID:        "a",
					EnvPrefix: "1",
					OnFail: []string{
						"b",
					},
				},
			},
			wantErr:    ErrCommandsValidate,
			wantOutput: "does not exist",
		},
		"good": {
			cmds: Commands{
				&Command{
					ID:        "a",
					EnvPrefix: "_A",
					OnFail: []string{
						"etcha:fail",
					},
					OnChange: []string{
						"b",
						"etcha:change",
					},
				},
				&Command{
					ID: "b",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logger.SetStd()
			assert.HasErr(t, tc.cmds.Validate(ctx), tc.wantErr)
			assert.Equal(t, strings.Contains(logger.ReadStd(), tc.wantOutput), true)
		})
	}
}
