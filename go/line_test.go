package main

import (
	"context"
	"os"
	"testing"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/logger"
)

func TestLine(t *testing.T) {
	ctx := context.Background()
	ctx = logger.SetNoColor(ctx, true)
	c := config.Default()

	os.WriteFile("test", []byte(`hello
world
replaceME
MEtoo`), 0600)

	tests := []struct {
		name        string
		mode        string
		path        string
		stdin       string
		match       string
		replace     string
		wantContent string
		wantErr     bool
	}{
		{
			name:        "bad mode",
			mode:        "no",
			wantErr:     true,
			wantContent: "ERROR unrecognized mode: no\n",
		},
		{
			name:        "remove",
			mode:        "remove",
			wantErr:     true,
			wantContent: "ERROR remove is not supported, use change\n",
		},
		{
			name:        "bad_file",
			mode:        "check",
			path:        "/root/no",
			wantErr:     true,
			wantContent: "ERROR open /root/no: permission denied\n",
		},
		{
			name:        "bad_regexp",
			match:       `[1\]`,
			mode:        "change",
			path:        "test",
			wantErr:     true,
			wantContent: "ERROR error parsing regexp: missing closing ]: `[1\\]`\n",
		},
		{
			name:        "fail_check_stdin",
			match:       "you",
			mode:        "check",
			stdin:       "hello=world",
			path:        "test",
			replace:     "you",
			wantErr:     true,
			wantContent: "ERROR replacement text not found\n",
		},
		{
			name:    "pass_check_stdin",
			match:   "you",
			mode:    "check",
			stdin:   "hello=world",
			path:    "test",
			replace: "world",
		},
		{
			name:        "fail_check_path",
			match:       `^me`,
			mode:        "check",
			path:        "test",
			replace:     "you",
			wantErr:     true,
			wantContent: "ERROR replacement text not found\n",
		},
		{
			name:    "change_path",
			match:   `\nME`,
			mode:    "change",
			path:    "test",
			replace: "\nyou",
			wantContent: `hello
world
replaceME
youtoo`,
		},
		{
			name:    "pass_check_path",
			match:   `^ME`,
			mode:    "check",
			path:    "test",
			replace: "\nyou",
			wantContent: `hello
world
replaceME
youtoo`,
		},
		{
			name:  "erase",
			match: `(?m)world\n`,
			mode:  "change",
			path:  "-",
			stdin: `hello
world
replaceME
`,
			wantContent: `hello
replaceME`,
		},
		{
			name:  "ssh",
			match: "(?m)^#?PermitRootLogin.*",
			mode:  "change",
			path:  "-",
			stdin: `#LoginGraceTime 2m
#PermitRootLogin prohibit-password
#StrictModes yes
`,
			replace: "PermitRootLogin yes",
			wantContent: `#LoginGraceTime 2m
PermitRootLogin yes
#StrictModes yes`,
		},
		{
			name:  "append-check",
			match: "(?m)^#?PermitRootLogin.*",
			mode:  "check",
			path:  "-",
			stdin: `#LoginGraceTime 2m
#StrictModes yes
`,
			replace:     "PermitRootLogin yes",
			wantErr:     true,
			wantContent: "ERROR replacement text not found\n",
		},
		{
			name:  "append-change",
			match: "(?m)^#?PermitRootLogin.*",
			mode:  "change",
			path:  "-",
			stdin: `#LoginGraceTime 2m
#StrictModes yes
`,
			replace: "PermitRootLogin yes",
			wantContent: `#LoginGraceTime 2m
#StrictModes yes
PermitRootLogin yes
`,
		},
		{
			name:  "no_change",
			match: "(?m)^#?PermitRootLogin.*",
			mode:  "change",
			path:  "-",
			stdin: `#LoginGraceTime 2m
#StrictModes yes
PermitRootLogin yes`,
			replace: "PermitRootLogin yes$$",
			wantContent: `#LoginGraceTime 2m
#StrictModes yes
PermitRootLogin yes$$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cli.SetStdin(tc.stdin)

			logger.SetStd()
			assert.Equal(t, line.Run(ctx, []string{
				"line",
				tc.mode,
				tc.path,
				tc.match,
				tc.replace,
			}, c) != nil, tc.wantErr)

			content := logger.ReadStd()
			if !tc.wantErr && tc.stdin == "" {
				o, _ := os.ReadFile("test")
				content = string(o)
			}

			assert.Equal(t, content, tc.wantContent)
		})
	}

	os.Remove("test")
}
