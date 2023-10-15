package commands

import (
	"testing"

	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

func TestCommandRun(t *testing.T) {
	logger.UseTestLogger(t)

	tests := map[string]struct {
		cmd          Command
		env          types.EnvVars
		execOverride bool
		mockErrs     []error
		mode         Mode
		wantErr      error
		wantEnv      types.EnvVars
		wantInputs   []cli.RunMockInput
		wantOutput   Output
	}{
		"remove_error": {
			cmd: Command{
				EnvPrefix: "a",
				Exec: &Exec{
					Command: "c",
				},
				ID:     "a",
				Remove: "remove",
			},
			execOverride: true,
			mockErrs: []error{
				ErrCommandsSelfTarget,
			},
			mode: ModeRemove,
			wantEnv: types.EnvVars{
				"a_REMOVE":     "1",
				"a_REMOVE_OUT": "output",
			},
			wantInputs: []cli.RunMockInput{
				{
					Exec: "c remove",
				},
			},
			wantOutput: Output{
				ID:         "a",
				Remove:     "output",
				Removed:    true,
				RemoveFail: true,
			},
			wantErr: ErrCommandsSelfTarget,
		},
		"remove": {
			cmd: Command{
				EnvPrefix: "a",
				ID:        "a",
				Remove:    "remove",
			},
			env:  types.EnvVars{"hello": "world"},
			mode: ModeRemove,
			wantEnv: types.EnvVars{
				"a_REMOVE":     "0",
				"a_REMOVE_OUT": "output",
				"hello":        "world",
			},
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{"hello=world"},
					Exec:        "remove",
				},
			},
			wantOutput: Output{
				ID:      "a",
				Remove:  cli.CmdOutput("output"),
				Removed: true,
			},
		},
		"skip": {
			cmd: Command{
				ID:     "a",
				Remove: "remove",
			},
			mode: ModeCheck,
			wantEnv: types.EnvVars{
				"_CHECK": "0",
			},
			wantOutput: Output{
				ID: "a",
			},
		},
		"check": {
			cmd: Command{
				Check:     "check",
				EnvPrefix: "a",
				ID:        "a",
			},
			mode: ModeChange,
			wantEnv: types.EnvVars{
				"a_CHECK":     "0",
				"a_CHECK_OUT": "output",
			},
			wantInputs: []cli.RunMockInput{
				{
					Exec: "check",
				},
			},
			wantOutput: Output{
				Check:   "output",
				Checked: true,
				ID:      "a",
			},
		},
		"check_error_check_only": {
			cmd: Command{
				Change:    "change",
				Check:     "check",
				EnvPrefix: "a",
				ID:        "a",
			},
			mockErrs: []error{
				ErrCommandsSelfTarget,
			},
			mode: ModeCheck,
			wantEnv: types.EnvVars{
				"a_CHECK":     "1",
				"a_CHECK_OUT": "output",
			},
			wantInputs: []cli.RunMockInput{
				{
					Exec: "check",
				},
			},
			wantOutput: Output{
				Check:     "output",
				Checked:   true,
				CheckFail: true,
				ID:        "a",
			},
		},
		"check_error_no_change": {
			cmd: Command{
				Check:     "check",
				EnvPrefix: "a",
				ID:        "a",
			},
			mockErrs: []error{
				ErrCommandsSelfTarget,
			},
			mode: ModeChange,
			wantEnv: types.EnvVars{
				"a_CHECK":     "1",
				"a_CHECK_OUT": "output",
			},
			wantInputs: []cli.RunMockInput{
				{
					Exec: "check",
				},
			},
			wantOutput: Output{
				Check:     "output",
				Checked:   true,
				CheckFail: true,
				ID:        "a",
			},
		},
		"check_error_change_error": {
			cmd: Command{
				Change:    "change",
				Check:     "check",
				EnvPrefix: "a",
				ID:        "a",
			},
			mockErrs: []error{
				ErrCommandsSelfTarget,
				ErrCommandsIDRequired,
			},
			mode: ModeChange,
			wantEnv: types.EnvVars{
				"a_CHANGE":     "1",
				"a_CHANGE_OUT": "output2",
				"a_CHECK":      "1",
				"a_CHECK_OUT":  "output",
			},
			wantErr: ErrCommandsIDRequired,
			wantInputs: []cli.RunMockInput{
				{
					Exec: "check",
				},
				{
					Environment: []string{"a_CHECK=1", "a_CHECK_OUT=output"},
					Exec:        "change",
				},
			},
			wantOutput: Output{
				Change:     "output2",
				Changed:    true,
				ChangeFail: true,
				Check:      "output",
				Checked:    true,
				CheckFail:  true,
				ID:         "a",
			},
		},
		"check_error_change": {
			cmd: Command{
				Change:    "change",
				Check:     "check",
				EnvPrefix: "a",
				ID:        "a",
			},
			mockErrs: []error{
				ErrCommandsSelfTarget,
			},
			mode: ModeChange,
			wantEnv: types.EnvVars{
				"a_CHANGE":     "0",
				"a_CHANGE_OUT": "output2",
				"a_CHECK":      "1",
				"a_CHECK_OUT":  "output",
			},
			wantInputs: []cli.RunMockInput{
				{
					Exec: "check",
				},
				{
					Environment: []string{"a_CHECK=1", "a_CHECK_OUT=output"},
					Exec:        "change",
				},
			},
			wantOutput: Output{
				Change:    "output2",
				Changed:   true,
				Check:     "output",
				Checked:   true,
				CheckFail: true,
				ID:        "a",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c := cli.Config{}
			c.RunMock()
			c.RunMockErrors(tc.mockErrs)
			c.RunMockOutputs([]string{"output", "output2"})

			out, env, err := tc.cmd.Run(ctx, c, tc.env, Exec{
				AllowOverride: tc.execOverride,
			}, tc.mode)

			assert.Equal(t, out, &tc.wantOutput)
			assert.Equal(t, env, tc.wantEnv)
			assert.Equal(t, c.RunMockInputs(), tc.wantInputs)
			assert.HasErr(t, err, tc.wantErr)
		})
	}
}
