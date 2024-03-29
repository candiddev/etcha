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
			{
				Change: "changeD",
				Check:  "checkD",
				ID:     "d",
				Remove: "removeD",
			},
		},
	}

	pOld := Pattern{
		Run: commands.Commands{
			{
				Change:      "changeb",
				Check:       "checkb",
				ID:          "b",
				Remove:      "removeb",
				RemoveAfter: true,
			},
			{
				Change: "changeC",
				Check:  "checkC",
				ID:     "c",
				Remove: "removeC",
			},
			{
				Change: "changeD",
				Check:  "checkD",
				ID:     "d",
				Remove: "removeD",
			},
		},
	}

	tests := map[string]struct {
		check       bool
		noRemove    bool
		mockError   []error
		wantErr     error
		wantInputs  []cli.RunMockInput
		wantOutputs commands.Outputs
	}{
		"change_error": {
			mockError: []error{
				ErrBuildEmpty,
				ErrBuildEmpty,
				ErrBuildEmpty,
			},
			wantInputs: []cli.RunMockInput{
				{
					Exec: "checkC",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="}, Exec: "checkA",
				},
				{
					Environment: []string{"_CHECK=1", "_CHECK_OUT="}, Exec: "changeA",
				},
			},
			wantOutputs: commands.Outputs{
				{
					Checked: true,
					ID:      "c",
				},
				{
					Changed:         true,
					ChangeFail:      true,
					Checked:         true,
					CheckFailChange: true,
					ID:              "a",
				},
			},
			wantErr: ErrBuildEmpty,
		},
		"remove_error": {
			mockError: []error{
				ErrBuildEmpty,
				nil,
				nil,
				nil,
				nil,
				ErrBuildEmpty,
				ErrBuildEmpty,
			},
			wantErr: ErrBuildEmpty,
			wantInputs: []cli.RunMockInput{
				{
					Exec: "checkC",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="},
					Exec:        "checkA",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="},
					Exec:        "checkB",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="},
					Exec:        "checkD",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="},
					Exec:        "checkb",
				},
				{
					Environment: []string{"_CHECK=1", "_CHECK_OUT="},
					Exec:        "removeb",
				},
			},
			wantOutputs: commands.Outputs{
				{
					Checked: true,
					ID:      "c",
				},
				{
					Checked: true,
					ID:      "a",
				},
				{
					Checked: true,
					ID:      "b",
				},
				{
					Checked: true,
					ID:      "d",
				},
				{
					Checked:         true,
					CheckFailRemove: true,
					ID:              "b",
					Removed:         true,
					RemoveFail:      true,
				},
			},
		},
		"good_check": {
			check: true,
			wantInputs: []cli.RunMockInput{
				{
					Exec: "checkC",
				},
				{
					Environment: []string{"_CHECK=1", "_CHECK_OUT="},
					Exec:        "checkA",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="},
					Exec:        "checkB",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="},
					Exec:        "checkD",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="},
					Exec:        "checkb",
				},
			},
			wantOutputs: commands.Outputs{
				{
					Checked:         true,
					CheckFailRemove: true,
					ID:              "c",
				},
				{
					Checked: true,
					ID:      "a",
				},
				{
					Checked: true,
					ID:      "b",
				},
				{
					Checked: true,
					ID:      "d",
				},
				{
					Checked:         true,
					CheckFailRemove: true,
					ID:              "b",
				},
			},
		},
		"good": {
			wantInputs: []cli.RunMockInput{
				{
					Exec: "checkC",
				},
				{
					Environment: []string{"_CHECK=1", "_CHECK_OUT="},
					Exec:        "removeC",
				},
				{
					Environment: []string{"_CHECK=1", "_CHECK_OUT=", "_REMOVE=0", "_REMOVE_OUT="},
					Exec:        "checkA",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT=", "_REMOVE=0", "_REMOVE_OUT="},
					Exec:        "checkB",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT=", "_REMOVE=0", "_REMOVE_OUT="},
					Exec:        "checkD",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT=", "_REMOVE=0", "_REMOVE_OUT="},
					Exec:        "checkb",
				},
				{
					Environment: []string{"_CHECK=1", "_CHECK_OUT=", "_REMOVE=0", "_REMOVE_OUT="},
					Exec:        "removeb",
				},
			},
			wantOutputs: commands.Outputs{
				{
					ID:              "c",
					Checked:         true,
					CheckFailRemove: true,
					Removed:         true,
				},
				{
					Checked: true,
					ID:      "a",
				},
				{
					Checked: true,
					ID:      "b",
				},
				{
					Checked: true,
					ID:      "d",
				},
				{
					ID:              "b",
					Checked:         true,
					CheckFailRemove: true,
					Removed:         true,
				},
			},
		},
		"good_noRemove": {
			noRemove: true,
			wantInputs: []cli.RunMockInput{
				{
					Exec: "checkA",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="}, Exec: "checkB",
				},
				{
					Environment: []string{"_CHECK=0", "_CHECK_OUT="}, Exec: "checkD",
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
					Checked: true,
					ID:      "d",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.CLI.RunMockErrors(tc.mockError)

			o, err := pNew.DiffRun(ctx, c, &pOld, DiffRunOpts{
				Check:    tc.check,
				NoRemove: tc.noRemove,
			})

			assert.HasErr(t, err, tc.wantErr)
			assert.Equal(t, o, tc.wantOutputs)
			assert.Equal(t, c.CLI.RunMockInputs(), tc.wantInputs)
		})
	}
}
