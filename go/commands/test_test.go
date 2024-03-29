package commands

import (
	"context"
	"regexp"
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
			nil,              // check
			nil,              // remove
			ErrCommandsEmpty, // check
			ErrCommandsEmpty, // check
		},
		testCheckChange: {
			ErrCommandsEmpty, // check
			nil,              // change
			ErrCommandsEmpty, // check
			nil,              // check
			nil,              // remove
			ErrCommandsEmpty, // check
			ErrCommandsEmpty, // check
		},
		testRemoveError: {
			ErrCommandsEmpty, // check
			nil,              // change
			nil,              // check
			nil,              // check
			ErrCommandsEmpty, // remove
			ErrCommandsEmpty, // check
			ErrCommandsEmpty, // check
		},
		testCheckRemove: {
			ErrCommandsEmpty, // check
			nil,              // change
			nil,              // check
			nil,              // check
			nil,              // remove
			nil,              // check
			ErrCommandsEmpty, // check
		},
		testCheck: {
			ErrCommandsEmpty, // check
			nil,              // change
			nil,              // check
			nil,              // check
			nil,              // remove
			ErrCommandsEmpty, // check
			nil,              // check
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.RunMockErrors(tc)
			assert.Equal(t, cmds.Test(ctx, c, &Exec{}, nil), types.Results{
				"a": []string{name},
			})
		})
	}

	cmds = Commands{
		{
			ID: "a",
			Commands: Commands{
				{
					ID: "b",
					Commands: Commands{
						{
							ID:     "c",
							Change: "c",
							Check:  "c",
							Remove: "c",
						},
					},
				},
			},
		},
	}

	c.RunMockErrors([]error{
		ErrCommandsEmpty, // check
		nil,              // change
		nil,              // check
		nil,              // check
		nil,              // remove
		ErrCommandsEmpty, // check
		nil,              // check
	})

	assert.Equal(t, cmds.Test(ctx, c, &Exec{}, regexp.MustCompile("^a > b$")), types.Results{
		"c": []string{testCheck},
	})
	assert.Equal(t, cmds.Test(ctx, c, &Exec{}, regexp.MustCompile("a$")), types.Results{})
}
