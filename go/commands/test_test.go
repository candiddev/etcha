package commands

import (
	"context"
	"testing"

	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

func TestTest(t *testing.T) {
	logger.UseTestLogger(t)

	c := cli.Config{}
	c.RunMock()

	ctx := context.Background()

	cmds := Commands{
		&Command{
			Change: "a",
			Check:  "b",
			ID:     "a",
			Remove: "c",
		},
	}

	tests := map[string][]error{
		testChangeError: {
			ErrCommandsEmpty, // check
			ErrCommandsEmpty, // change
			nil,              // check
			nil,              // remove
			ErrCommandsEmpty, // check
		},
		testCheckChange: {
			ErrCommandsEmpty, // check
			nil,              // change
			ErrCommandsEmpty, // check
			nil,              // remove
			ErrCommandsEmpty, // check
		},
		testRemoveError: {
			ErrCommandsEmpty, // check
			nil,              // change
			nil,              // check
			ErrCommandsEmpty, // remove
			ErrCommandsEmpty, // check
		},
		testCheckRemove: {
			ErrCommandsEmpty, // check
			nil,              // change
			nil,              // check
			nil,              // remove
			nil,              // check
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.RunMockErrors(tc)
			assert.Equal(t, cmds.Test(ctx, c, &Exec{}, types.EnvVars{}), types.Results{
				"a": []string{name},
			})
		})
	}
}
