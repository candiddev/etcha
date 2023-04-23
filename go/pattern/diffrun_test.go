package pattern

import (
	"context"
	"testing"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/logger"
)

func TestPatternDiffRun(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.CLI.RunMock()

	pNew := Pattern{
		Run: commands.Commands{
			{
				Change: "changeA",
				Check:  "checkA",
				ID:     "a",
				Remove: "removeA",
			},
			{
				Change: "changeB",
				Check:  "checkB",
				ID:     "b",
				Remove: "removeB",
			},
		},
	}

	pOld := Pattern{
		Run: commands.Commands{
			{
				Change: "changeb",
				Check:  "checkb",
				ID:     "b",
				Remove: "removeb",
			},
			{
				Change: "changeC",
				Check:  "checkC",
				ID:     "c",
				Remove: "removeC",
			},
		},
	}

	tests := map[string]struct {
		check       bool
		mockError   []error
		wantErr     error
		wantInputs  []cli.RunMockInput
		wantOutputs commands.Outputs
	}{
		"change_error": {
			mockError: []error{
				ErrBuildEmpty,
				ErrBuildEmpty,
			},
			wantInputs: []cli.RunMockInput{
				{
					Exec: " checkA",
				},
				{
					Environment: []string{"_CHECK=1", "_CHECK_OUT="}, Exec: " changeA",
				},
			},
			wantOutputs: commands.Outputs{
				{
					Changed:    true,
					ChangeFail: true,
					Checked:    true,
					CheckFail:  true,
					ID:         "a",
				},
			},
			wantErr: ErrBuildEmpty,
		},
		"remove_error": {
			mockError: []error{
				nil,
				nil,
				ErrBuildEmpty,
			},
			wantErr: ErrBuildEmpty,
			wantInputs: []cli.RunMockInput{
				{
					Exec: " checkA",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="}, Exec: " checkB",
				},
				{
					Exec: " removeC",
				},
			},
			wantOutputs: commands.Outputs{
				{
					Checked: true,
					ID:      "a",
				},
				{
					Checked: true,
					ID:      "b",
				},
				{
					ID:         "c",
					Removed:    true,
					RemoveFail: true,
				},
			},
		},
		"good_check": {
			check: true,
			wantInputs: []cli.RunMockInput{
				{
					Exec: " checkA",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="}, Exec: " checkB",
				},
			},
			wantOutputs: commands.Outputs{
				{
					Checked: true,
					ID:      "a",
				},
				{
					Checked: true,
					ID:      "b",
				},
			},
		},
		"good": {
			wantInputs: []cli.RunMockInput{
				{
					Exec: " checkA",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="}, Exec: " checkB",
				},
				{
					Exec: " removeC",
				},
			},
			wantOutputs: commands.Outputs{
				{
					Checked: true,
					ID:      "a",
				},
				{
					Checked: true,
					ID:      "b",
				},
				{
					ID:      "c",
					Removed: true,
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.CLI.RunMockErrors(tc.mockError)

			o, err := pNew.DiffRun(ctx, c, &pOld, tc.check)

			assert.HasErr(t, err, tc.wantErr)
			assert.Equal(t, o, tc.wantOutputs)
			assert.Equal(t, c.CLI.RunMockInputs(), tc.wantInputs)
		})
	}
}
