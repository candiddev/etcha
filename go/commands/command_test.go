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
		check        bool
		env          types.EnvVars
		execOverride bool
		mockErrs     []error
		remove       bool
		wantErr      error
		wantEnv      types.EnvVars
		wantInputs   []cli.RunMockInput
		wantOutput   Output
	}{
		"remove_check_error": {
			check: true,
			cmd: Command{
				Check:     "check",
				EnvPrefix: "a",
				ID:        "a",
				Remove:    "remove",
			},
			mockErrs: []error{
				ErrCommandsSelfTarget,
			},
			remove: true,
			wantEnv: types.EnvVars{
				"a":           "output",
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
		"remove_no_remove": {
			cmd: Command{
				Check:     "check",
				EnvPrefix: "a",
				ID:        "a",
			},
			mockErrs: []error{
				ErrCommandsIDRequired,
			},
			remove: true,
			wantEnv: types.EnvVars{
				"a":           "output",
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
		"remove_errors": {
			cmd: Command{
				Check:     "check",
				EnvPrefix: "a",
				ID:        "a",
				Remove:    "remove",
			},
			mockErrs: []error{
				nil,
				ErrCommandsIDRequired,
			},
			remove: true,
			wantEnv: types.EnvVars{
				"a":            "output",
				"a_CHECK":      "1",
				"a_CHECK_OUT":  "output",
				"a_REMOVE":     "1",
				"a_REMOVE_OUT": "output2",
			},
			wantErr: ErrCommandsIDRequired,
			wantInputs: []cli.RunMockInput{
				{
					Exec: "check",
				},
				{
					Environment: []string{"a=output", "a_CHECK=1", "a_CHECK_OUT=output"},
					Exec:        "remove",
				},
			},
			wantOutput: Output{
				Check:           "output",
				Checked:         true,
				CheckFailRemove: true,
				ID:              "a",
				Remove:          "output2",
				Removed:         true,
				RemoveFail:      true,
			},
		},
		"remove_removed": {
			cmd: Command{
				Check:     "check",
				EnvPrefix: "a",
				ID:        "a",
				Remove:    "remove",
			},
			remove: true,
			wantEnv: types.EnvVars{
				"a":            "output",
				"a_CHECK":      "1",
				"a_CHECK_OUT":  "output",
				"a_REMOVE":     "0",
				"a_REMOVE_OUT": "output2",
			},
			wantInputs: []cli.RunMockInput{
				{
					Exec: "check",
				},
				{
					Environment: []string{"a=output", "a_CHECK=1", "a_CHECK_OUT=output"},
					Exec:        "remove",
				},
			},
			wantOutput: Output{
				Check:           "output",
				Checked:         true,
				CheckFailRemove: true,
				ID:              "a",
				Remove:          "output2",
				Removed:         true,
			},
		},
		"remove_always": {
			cmd: Command{
				Always: true,
				ID:     "a",
				Remove: "remove",
			},
			remove: true,
			wantEnv: types.EnvVars{
				"_CHECK":      "1",
				"_REMOVE":     "0",
				"_REMOVE_OUT": "output",
			},
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{"_CHECK=1"},
					Exec:        "remove",
				},
			},
			wantOutput: Output{
				CheckFailRemove: true,
				ID:              "a",
				Remove:          "output",
				Removed:         true,
			},
		},
		"remove_removeBy": {
			cmd: Command{
				Always:    true,
				ID:        "a",
				Remove:    "remove",
				RemovedBy: []string{"b"},
			},
			remove: true,
			wantEnv: types.EnvVars{
				"_CHECK":      "1",
				"_REMOVE":     "0",
				"_REMOVE_OUT": "output",
			},
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{"_CHECK=1"},
					Exec:        "remove",
				},
			},
			wantOutput: Output{
				ID:              "a",
				CheckFailRemove: true,
				Remove:          "output",
				Removed:         true,
			},
		},
		"skip": {
			check: true,
			cmd: Command{
				ID:     "a",
				Remove: "remove",
			},
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
			wantEnv: types.EnvVars{
				"a":           "output",
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
			check: true,
			cmd: Command{
				Change:    "change",
				Check:     "check",
				EnvPrefix: "a",
				ID:        "a",
			},
			mockErrs: []error{
				ErrCommandsSelfTarget,
			},
			wantEnv: types.EnvVars{
				"a":           "output",
				"a_CHECK":     "1",
				"a_CHECK_OUT": "output",
			},
			wantInputs: []cli.RunMockInput{
				{
					Exec: "check",
				},
			},
			wantOutput: Output{
				Check:           "output",
				Checked:         true,
				CheckFailChange: true,
				ID:              "a",
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
			wantEnv: types.EnvVars{
				"a":           "output",
				"a_CHECK":     "1",
				"a_CHECK_OUT": "output",
			},
			wantInputs: []cli.RunMockInput{
				{
					Exec: "check",
				},
			},
			wantOutput: Output{
				Check:           "output",
				Checked:         true,
				CheckFailChange: true,
				ID:              "a",
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
			wantEnv: types.EnvVars{
				"a":            "output",
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
					Environment: []string{"a=output", "a_CHECK=1", "a_CHECK_OUT=output"},
					Exec:        "change",
				},
			},
			wantOutput: Output{
				Change:          "output2",
				Changed:         true,
				ChangeFail:      true,
				Check:           "output",
				Checked:         true,
				CheckFailChange: true,
				ID:              "a",
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
			wantEnv: types.EnvVars{
				"a":            "output",
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
					Environment: []string{"a=output", "a_CHECK=1", "a_CHECK_OUT=output"},
					Exec:        "change",
				},
			},
			wantOutput: Output{
				Change:          "output2",
				Changed:         true,
				Check:           "output",
				Checked:         true,
				CheckFailChange: true,
				ID:              "a",
			},
		},
		"change_always": {
			cmd: Command{
				Always: true,
				Change: "change",
				ID:     "a",
			},
			wantEnv: types.EnvVars{
				"_CHECK":      "1",
				"_CHANGE":     "0",
				"_CHANGE_OUT": "output",
			},
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{"_CHECK=1"},
					Exec:        "change",
				},
			},
			wantOutput: Output{
				CheckFailChange: true,
				Change:          "output",
				Changed:         true,
				ID:              "a",
			},
		},
		"change_changedBy": {
			cmd: Command{
				Always:    true,
				Change:    "change",
				ChangedBy: []string{"b"},
				ID:        "a",
			},
			wantEnv: types.EnvVars{
				"_CHECK":      "1",
				"_CHANGE":     "0",
				"_CHANGE_OUT": "output",
			},
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{"_CHECK=1"},
					Exec:        "change",
				},
			},
			wantOutput: Output{
				Change:          "output",
				Changed:         true,
				ID:              "a",
				CheckFailChange: true,
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
			}, tc.check, tc.remove)

			assert.Equal(t, out, &tc.wantOutput)
			assert.Equal(t, env, tc.wantEnv)
			assert.Equal(t, c.RunMockInputs(), tc.wantInputs)
			assert.HasErr(t, err, tc.wantErr)
		})
	}
}
