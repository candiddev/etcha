package commands

import (
	"context"
	"regexp"
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
		"[e|f]",
		"etcha:hello",
		"etcha:stderr",
		"etcha:stdout",
	},
	OnFail: []string{
		"f",
		"etcha:fail",
	},
	OnRemove: []string{
		"e",
		"etcha:hello",
		"etcha:stderr",
		"etcha:stdout",
	},
}
var cmdB = &Command{
	Change: "changeB",
	Remove: "removeB",
	ID:     "b",
}
var cmdC = &Command{
	Check:       "checkC",
	Remove:      "removeC",
	RemoveAfter: true,
	ID:          "c",
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
var cmdH = &Command{
	Commands: Commands{
		{
			Commands: Commands{
				{
					ID:     "j",
					Change: "changeJ",
					Check:  "checkJ",
					Remove: "removeJ",
				},
			},
			ID: "i",
		},
	},
	ID: "h",
}

func TestCommandsDiff(t *testing.T) {
	before, after := Commands{
		cmdA,
		cmdB,
		cmdG,
		{
			ID:     "k",
			Check:  "checkL",
			Change: "changeL",
			Remove: "removeL",
		},
		{
			ID: "l",
			Commands: Commands{
				{
					ID:     "m",
					Always: true,
					Change: "removeN",
					Remove: "removeN",
				},
			},
		},
	}.Diff(Commands{
		cmdA,
		cmdC,
		cmdG,
		cmdH,
		{
			ID:     "k",
			Check:  "checkK",
			Change: "changeK",
			Remove: "removeL",
		},
		{
			ID: "l",
			Commands: Commands{
				{
					ID:     "m",
					Always: true,
					Remove: "removeM",
				},
			},
		},
	})
	assert.Equal(t, before, Commands{
		cmdH.Commands[0].Commands[0],
		{
			ID:     "k",
			Check:  "checkK",
			Change: "changeK",
			Remove: "removeL",
		},
		{
			ID:     "m",
			Always: true,
			Remove: "removeM",
		},
	})
	assert.Equal(t, after, Commands{
		cmdC,
	})
}

func TestCommandsRun(t *testing.T) {
	logger.UseTestLogger(t)

	tests := map[string]struct {
		check          bool
		cmds           Commands
		env            types.EnvVars
		execOverride   bool
		mockErrs       []error
		mockOutputs    []string
		parentID       string
		parentIDFilter *regexp.Regexp
		remove         bool
		wantErr        error
		wantInputs     []cli.RunMockInput
		wantOut        string
		wantOutputs    Outputs
	}{
		"remove": {
			cmds: Commands{
				cmdA,
				cmdG,
			},
			env: types.EnvVars{
				"hello": "world",
			},
			mockOutputs: []string{"output1", "", "output2"},
			remove:      true,
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{"g_CHECK=1", "hello=world"},
					Exec:        "removeG",
				},
				{
					Environment: []string{
						"g_CHECK=1",
						"g_REMOVE=0",
						"g_REMOVE_OUT=output1",
						"hello=world",
					},
					Exec: "checkA",
				},
				{
					Environment: []string{
						"_CHECK=1",
						"_CHECK_OUT=",
						"g_CHECK=1",
						"g_REMOVE=0",
						"g_REMOVE_OUT=output1",
						"hello=world",
					},
					Exec: "removeA",
				},
			},
			wantOut: "INFO  Always removing g...\nINFO  Removing a...\noutput2output2",
			wantOutputs: Outputs{
				&Output{
					ID:              "g",
					CheckFailRemove: true,
					Remove:          "output1",
					Removed:         true,
				},
				&Output{
					ID:              "a",
					Checked:         true,
					CheckFailRemove: true,
					Events:          []string{"hello", "stderr", "stdout"},
					Remove:          "output2",
					Removed:         true,
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
					Environment: []string{"_CHANGE=0", "_CHANGE_OUT=", "_CHECK=1", "_CHECK_OUT="},
					Exec:        "changeF",
				},
				{
					Environment: []string{"_CHANGE=0", "_CHANGE_OUT=", "_CHECK=1", "_CHECK_OUT=",
						"g_CHECK=1"},
					Exec: "changeG",
				},
			},
			wantOut: "INFO  Changing a...\na\na\nINFO  Triggering e via a...\nINFO  Triggering f via a...\nINFO  Always changing g...\n",
			wantOutputs: Outputs{
				&Output{
					Change:          "a",
					Changed:         true,
					Check:           "a",
					CheckFailChange: true,
					Checked:         true,
					ID:              "a",
					Events:          []string{"hello", "stderr", "stdout"},
				},
				&Output{
					Checked: true,
					ID:      "c",
				},
				&Output{
					Changed:         true,
					CheckFailChange: true,
					ID:              "e",
				},
				&Output{
					Changed:         true,
					CheckFailChange: true,
					ID:              "f",
				},
				&Output{
					CheckFailChange: true,
					Changed:         true,
					ID:              "g",
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
					Changed:         true,
					ChangeFail:      true,
					CheckFailChange: true,
					Checked:         true,
					Events:          []string{"fail"},
					ID:              "a",
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
		"nested_no_filter": {
			cmds: Commands{
				cmdA,
				cmdH,
			},
			wantInputs: []cli.RunMockInput{
				{
					Exec: "checkA",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="},
					Exec:        "checkJ",
				},
			},
			wantOutputs: Outputs{
				&Output{
					Checked: true,
					ID:      "a",
				},
				&Output{
					Checked:  true,
					ID:       "j",
					ParentID: "h > i",
				},
			},
		},
		"nested_filter1": {
			cmds: Commands{
				cmdA,
				cmdH,
			},
			parentID:       "k",
			parentIDFilter: regexp.MustCompile("^k$"),
			wantInputs: []cli.RunMockInput{
				{
					Exec: "checkA",
				},
			},
			wantOutputs: Outputs{
				{
					Checked:  true,
					ID:       "a",
					ParentID: "k",
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

			out, err := tc.cmds.Run(ctx, c, &Exec{
				AllowOverride: tc.execOverride,
			}, CommandsRunOpts{
				Check:          tc.check,
				Env:            tc.env,
				ParentID:       tc.parentID,
				ParentIDFilter: tc.parentIDFilter,
				Remove:         tc.remove,
			})

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
			err := tc.cmds.validate()
			assert.HasErr(t, err, tc.wantErr)

			if tc.wantErr != nil {
				assert.Equal(t, strings.Contains(err.Error(), tc.wantOutput), true)
			}
		})
	}
}
