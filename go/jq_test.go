package main

import (
	"context"
	"testing"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/logger"
)

func TestJQ(t *testing.T) {
	c := config.Default()
	ctx := context.Background()

	logger.NoColor()

	tests := map[string]struct {
		args    []string
		wantOut string
		wantErr error
	}{
		"raw": {
			args: []string{
				"",
				"-r",
				".nested[0].string",
			},
			wantOut: "value\n",
		},
		"notRaw": {
			args: []string{
				"",
				".nested[0].string",
			},
			wantOut: `"value"` + "\n",
		},
		"int": {
			args: []string{
				"",
				".nested[0].int",
			},
			wantOut: "10\n",
		},
		"bool": {
			args: []string{
				"",
				".nested[0].bool",
			},
			wantOut: "true\n",
		},
		"array": {
			args: []string{
				"",
				".nested",
			},
			wantOut: `[
  {
    "bool": true,
    "int": 10,
    "string": "value"
  }
]
`,
		},
		"invalid": {
			args: []string{
				"",
				"oops",
			},
			wantErr: errJQ,
			wantOut: "ERROR go/jq.go:56\nerror querying JSON: function not defined: oops/0\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cli.SetStdin(`{"nested":[{"string":"value","int":10,"bool":true}]}`)
			logger.SetStd()
			assert.HasErr(t, jq(ctx, tc.args, c), tc.wantErr)
			assert.Equal(t, logger.ReadStd(), tc.wantOut)
		})
	}
}
