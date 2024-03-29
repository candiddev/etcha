package run

import (
	"context"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/cryptolib"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
)

func TestPush(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.Exec.AllowOverride = true
	c.CLI.RunMock()

	s, _ := newState(ctx, c)
	m, _ := s.newMux(ctx)
	ts := httptest.NewServer(m)

	prv1, pub1, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)
	prv2, _, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)

	os.MkdirAll("testdata/state", 0700)

	c.Run.StateDir = "testdata/state"

	os.WriteFile("testdata/good1.jsonnet", []byte(`
{
	build: [
		{
			id: "1",
			always: true,
			change: "change2",
			onChange: [
				"etcha:runVar_a",
			],
		},
	],
	run: [
		{
			change: "change1",
			check: std.get(std.native('getConfig')().vars, 'check'),
			id: "1",
			remove: "remove1",
		}
	],
	runVars: {
		check: 'check1',
	},
}
`), 0600)

	os.WriteFile("testdata/good2.jsonnet", []byte(`
{
	run: [
		{
			change: "change2",
			check: "check2",
			id: "2",
			remove: "remove2",
		}
	]
}
`), 0600)

	c.Sources = map[string]*config.Source{
		"etcha": {
			AllowPush: true,
			Exec: &commands.Exec{
				Command: "",
			},
			VerifyKeys: cryptolib.Keys[cryptolib.KeyProviderPublic]{
				pub1,
			},
		},
		"denied": {},
	}

	tests := []struct {
		check       bool
		command     string
		destination string
		filter      *regexp.Regexp
		mockErrors  []error
		name        string
		path        string
		signingKey  cryptolib.Key[cryptolib.KeyProviderPrivate]
		wantErr     error
		wantInputs  []cli.RunMockInput
		wantResult  *Result
	}{
		{
			name:       "bad_path",
			path:       "testdata/not.jsonnet",
			wantErr:    jsonnet.ErrImport,
			wantResult: &Result{},
		},
		{
			name:       "bad_sign",
			path:       "testdata/good1.jsonnet",
			wantErr:    pattern.ErrPatternMissingKey,
			wantResult: &Result{},
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{
						"_CHECK=1",
					},
					Exec: "/usr/bin/bash -e -o pipefail -c change2",
				},
			},
		},
		{
			name:        "no_source",
			destination: ts.URL + "/etcha/v1/push/nowhere",
			signingKey:  prv2,
			path:        "testdata/good1.jsonnet",
			wantErr:     ErrPushSourceMismatch,
			wantResult:  &Result{},
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{
						"_CHECK=1",
					},
					Exec: "/usr/bin/bash -e -o pipefail -c change2",
				},
			},
		},
		{
			name:        "denied_source",
			destination: ts.URL + "/etcha/v1/push/denied",
			signingKey:  prv2,
			path:        "testdata/good1.jsonnet",
			wantErr:     ErrPushSourceMismatch,
			wantResult:  &Result{},
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{
						"_CHECK=1",
					},
					Exec: "/usr/bin/bash -e -o pipefail -c change2",
				},
			},
		},
		{
			name:        "bad_private_key",
			destination: ts.URL + "/etcha/v1/push/etcha",
			signingKey:  prv2,
			path:        "testdata/good1.jsonnet",
			wantErr:     ErrPushSourceMismatch,
			wantResult:  &Result{},
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{
						"_CHECK=1",
					},
					Exec: "/usr/bin/bash -e -o pipefail -c change2",
				},
			},
		},
		{
			name:        "error_build",
			destination: ts.URL + "/etcha/v1/push/etcha",
			mockErrors: []error{
				ErrPushSourceMismatch,
			},
			signingKey: prv1,
			path:       "testdata/good1.jsonnet",
			wantErr:    ErrPushSourceMismatch,
			wantResult: &Result{},
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{
						"_CHECK=1",
					},
					Exec: "/usr/bin/bash -e -o pipefail -c change2",
				},
			},
		},
		{
			name:        "error_exec",
			destination: ts.URL + "/etcha/v1/push/etcha",
			mockErrors: []error{
				nil,
				ErrNoVerifyKeys,
				ErrNoVerifyKeys,
			},
			signingKey: prv1,
			path:       "testdata/good1.jsonnet",
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{
						"_CHECK=1",
					},
					Exec: "/usr/bin/bash -e -o pipefail -c change2",
				},
				{
					Exec: "check1",
				},
				{
					Environment: []string{
						"_CHECK=1",
						"_CHECK_OUT=a",
					},
					Exec: "change1",
				},
			},
			wantResult: &Result{
				Err: "error changing id etcha > 1: error running commands: error running commands: no verify keys specified: b",
			},
			wantErr: errs.ErrReceiver,
		},
		{
			name:        "good",
			destination: ts.URL + "/etcha/v1/push/etcha",
			mockErrors: []error{
				nil,
				ErrNoVerifyKeys,
			},
			signingKey: prv1,
			path:       "testdata/good1.jsonnet",
			wantInputs: []cli.RunMockInput{
				{
					Environment: []string{
						"_CHECK=1",
					},
					Exec: "/usr/bin/bash -e -o pipefail -c change2",
				},
				{Exec: "check1"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT=a"}, Exec: "change1"},
			},
			wantResult: &Result{
				ChangedIDs:     []string{"1"},
				ChangedOutputs: []string{"b"},
			},
		},
		{
			name:        "good-check",
			check:       true,
			destination: ts.URL + "/etcha/v1/push/etcha",
			mockErrors: []error{
				ErrNoVerifyKeys,
			},
			signingKey: prv1,
			path:       "testdata/good2.jsonnet",
			wantInputs: []cli.RunMockInput{
				{Exec: "check2"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT=1"}, Exec: "check1"},
			},
			wantResult: &Result{
				ChangedIDs: []string{"2"},
				RemovedIDs: []string{"1"},
			},
		},
		{
			name:        "good-2",
			destination: ts.URL + "/etcha/v1/push/etcha",
			signingKey:  prv1,
			path:        "testdata/good2.jsonnet",
			wantInputs: []cli.RunMockInput{
				{Exec: "check2"},
				{Environment: []string{"_CHECK=0", "_CHECK_OUT=1"}, Exec: "check1"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT=a"}, Exec: "remove1"},
			},
			wantResult: &Result{
				RemovedIDs:     []string{"1"},
				RemovedOutputs: []string{"b"},
			},
		},
		{
			name:        "good-command",
			command:     "ls",
			destination: ts.URL + "/etcha/v1/push/etcha",
			signingKey:  prv1,
			wantInputs: []cli.RunMockInput{
				{Environment: []string{"_CHECK=1"}, Exec: "/usr/bin/ls"},
				{Environment: []string{"_CHANGE=0", "_CHANGE_OUT=1", "_CHECK=1"}, Exec: "check2"},
				{
					Environment: []string{"_CHANGE=0", "_CHANGE_OUT=1", "_CHECK=1", "_CHECK_OUT=a"},
					Exec:        "remove2",
				},
			},
			wantResult: &Result{
				ChangedIDs:     []string{"etcha push"},
				ChangedOutputs: []string{"1"},
				RemovedIDs:     []string{"2"},
				RemovedOutputs: []string{"b"},
			},
		},
		{
			name:        "good-filter",
			command:     "ls",
			destination: ts.URL + "/etcha/v1/push/etcha",
			filter:      regexp.MustCompile("^123$"),
			signingKey:  prv1,
			wantResult:  &Result{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c.Build.SigningKey = tc.signingKey.String()
			c.CLI.RunMockErrors(tc.mockErrors)
			c.CLI.RunMockOutputs([]string{"1", "a", "b", "c", "d", "e"})

			r, err := Push(ctx, c, tc.destination, tc.command, tc.path, PushOpts{
				Check:          tc.check,
				ParentIDFilter: tc.filter,
			})
			assert.HasErr(t, err, tc.wantErr)
			assert.Equal(t, r, tc.wantResult)
			assert.Equal(t, c.CLI.RunMockInputs(), tc.wantInputs)
		})
	}

	os.RemoveAll("testdata")
}
