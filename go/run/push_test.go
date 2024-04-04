package run

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	"github.com/candiddev/shared/go/types"
)

func TestPushTargets(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()

	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		j, _ := json.MarshalIndent(Result{
			ChangedIDs:     []string{"2"},
			ChangedOutputs: []string{"a"},
			RemovedIDs:     []string{"1"},
		}, "", "")

		w.Write(j)
	}))
	p1, _ := strconv.Atoi(strings.Split(ts1.URL, ":")[2])

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		j, _ := json.MarshalIndent(Result{
			Err: "error",
		}, "", "")

		w.Write(j)
	}))
	p2, _ := strconv.Atoi(strings.Split(ts2.URL, ":")[2])

	ts3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		j, _ := json.MarshalIndent(Result{
			ChangedIDs:     []string{"3"},
			ChangedOutputs: []string{"b"},
			RemovedIDs:     []string{"2"},
		}, "", "")

		w.Write(j)
	}))
	p3, _ := strconv.Atoi(strings.Split(ts3.URL, ":")[2])

	c := config.Default()

	cmd := ""

	ts4 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		j, _ := io.ReadAll(r.Body)
		o, _, _ := pattern.ParseJWT(ctx, c, string(j), "")

		p, _ := o.Pattern(ctx, c, "")

		cmd = p.Run[0].Change

		if cmd != "test" {
			w.WriteHeader(http.StatusBadGateway)
		} else {
			w.Write([]byte(types.JSONToString(&Result{})))
		}
	}))
	p4, _ := strconv.Atoi(strings.Split(ts4.URL, ":")[2])

	os.WriteFile("a.jsonnet", []byte(`{
		run: [
			{
				id: 'a',
				change: 'a',
				check: 'a',
			}
		]
	}`), 0600)

	c.Build.PushMaxWorkers = 2
	prv, _, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)
	c.Build.SigningKey = prv.String()
	tgt := map[string]config.Target{
		"1": {
			Hostname: "127.0.0.1",
			Insecure: true,
			PathPush: "/etcha/v1/push",
			Port:     p1,
			SourcePatterns: map[string]string{
				"a": "",
				"b": "",
			},
			Vars: map[string]any{
				"1": 2,
			},
		},
		"2": {
			Hostname: "127.0.0.1",
			Insecure: true,
			PathPush: "/etcha/v1/push",
			Port:     p2,
			SourcePatterns: map[string]string{
				"a": "",
				"c": "",
			},
		},
		"3": {
			Hostname: "localhost",
			Insecure: true,
			PathPush: "/etcha/v1/push",
			Port:     p3,
			SourcePatterns: map[string]string{
				"a": "",
				"c": "",
			},
		},
		"4": {
			Hostname: "localhost",
			Insecure: true,
			PathPush: "/etcha/v1/push",
			Port:     p4,
			SourcePatterns: map[string]string{
				"a": "",
				"d": "test",
			},
		},
	}

	// PushTargets
	s, err := PushTargets(ctx, c, tgt, "a", "ls", PushOpts{})
	assert.HasErr(t, err, ErrPushDecode)
	assert.Equal(t, s, []string{
		"1:\n    a",
		fmt.Sprintf("2:\n    %sERROR: error%s", logger.ColorRed, logger.ColorReset),
		"3:\n    b",
		fmt.Sprintf("4:\n    %sERROR: error decoding response: unexpected end of JSON input%s", logger.ColorRed, logger.ColorReset),
	})

	ctx = logger.SetNoColor(ctx, true)

	s, err = PushTargets(ctx, c, tgt, "c", "a.jsonnet", PushOpts{})
	assert.Equal(t, err.Error(), "error")
	assert.Equal(t, s, []string{
		"2:\n    ERROR: error\n    No changes",
		"3:\n    Changed 1: 3\n    Removed 1: 2",
	})

	s, err = PushTargets(ctx, c, tgt, "c", "a.jsonnet", PushOpts{
		TargetFilter: regexp.MustCompile("^[1|3]$"),
	})
	assert.HasErr(t, err, nil)
	assert.Equal(t, s, []string{
		"3:\n    Changed 1: 3\n    Removed 1: 2",
	})

	_, err = PushTargets(ctx, c, tgt, "d", "", PushOpts{})
	assert.HasErr(t, err, nil)
	assert.Equal(t, cmd, "test")

	_, err = PushTargets(ctx, c, tgt, "d", "a.jsonnet", PushOpts{})
	assert.HasErr(t, err, nil)
	assert.Equal(t, cmd, "test")

	// getPushDestJWT
	d, j, err := getPushDestJWT(ctx, c, tgt["1"], &pattern.Pattern{}, "", "test", map[string]any{
		"1": 1,
		"2": 2,
	}, PushOpts{
		Check:          true,
		ParentIDFilter: regexp.MustCompile("^1$"),
	})
	assert.HasErr(t, err, nil)
	assert.Equal(t, d, ts1.URL+"/etcha/v1/push/test")

	jw, _, _ := pattern.ParseJWT(ctx, c, j, "")
	assert.Equal(t, jw.EtchaRunVars, map[string]any{
		"1": float64(2),
		"2": float64(2),
	})

	d, _, _ = getPushDestJWT(ctx, c, config.Target{
		Hostname: "a",
		PathPush: "/b",
		Port:     123,
	}, &pattern.Pattern{}, "", "test", map[string]any{
		"1": 1,
		"2": 2,
	}, PushOpts{
		Check:          true,
		ParentIDFilter: regexp.MustCompile("^1$"),
	})
	assert.Equal(t, d, "https://a:123/b/test")

	os.Remove("a.jsonnet")
}

func TestPushTargetPostPush(t *testing.T) {
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
		command     string
		destination string
		mockErrors  []error
		name        string
		signingKey  cryptolib.Key[cryptolib.KeyProviderPrivate]
		wantErr     error
		wantInputs  []cli.RunMockInput
		wantResult  *Result
	}{
		{
			name:    "bad_path",
			command: "testdata/not.jsonnet",
			wantErr: jsonnet.ErrImport,
		},
		{
			name:    "bad_sign",
			command: "testdata/good1.jsonnet",
			wantErr: config.ErrMissingBuildKey,
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
			command:     "testdata/good1.jsonnet",
			destination: ts.URL + "/etcha/v1/push/nowhere",
			signingKey:  prv2,
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
			command:     "testdata/good1.jsonnet",
			destination: ts.URL + "/etcha/v1/push/denied",
			signingKey:  prv2,
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
			command:     "testdata/good1.jsonnet",
			destination: ts.URL + "/etcha/v1/push/etcha",
			signingKey:  prv2,
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
			command:     "testdata/good1.jsonnet",
			destination: ts.URL + "/etcha/v1/push/etcha",
			mockErrors: []error{
				ErrPushSourceMismatch,
			},
			signingKey: prv1,
			wantErr:    ErrPushSourceMismatch,
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
			command:     "testdata/good1.jsonnet",
			destination: ts.URL + "/etcha/v1/push/etcha",
			mockErrors: []error{
				nil,
				ErrNoVerifyKeys,
				ErrNoVerifyKeys,
			},
			signingKey: prv1,
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
			command:     "testdata/good1.jsonnet",
			destination: ts.URL + "/etcha/v1/push/etcha",
			mockErrors: []error{
				nil,
				ErrNoVerifyKeys,
			},
			signingKey: prv1,
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
			command:     "testdata/good2.jsonnet",
			destination: ts.URL + "/etcha/v1/push/etcha?check=",
			mockErrors: []error{
				nil,
				ErrNoVerifyKeys,
			},
			signingKey: prv1,
			wantInputs: []cli.RunMockInput{
				{Exec: "check1"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT=1"}, Exec: "check2"},
			},
			wantResult: &Result{
				ChangedIDs: []string{"2"},
				RemovedIDs: []string{"1"},
			},
		},
		{
			name:        "good-2",
			command:     "testdata/good2.jsonnet",
			destination: ts.URL + "/etcha/v1/push/etcha",
			signingKey:  prv1,
			wantInputs: []cli.RunMockInput{
				{Exec: "check1"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT=1"}, Exec: "remove1"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT=1", "_REMOVE=0", "_REMOVE_OUT=a"}, Exec: "check2"},
			},
			wantResult: &Result{
				RemovedIDs:     []string{"1"},
				RemovedOutputs: []string{"a"},
			},
		},
		{
			name:        "good-command",
			command:     "ls",
			destination: ts.URL + "/etcha/v1/push/etcha",
			signingKey:  prv1,
			wantInputs: []cli.RunMockInput{
				{Exec: "check2"},
				{
					Environment: []string{"_CHECK=1", "_CHECK_OUT=1"},
					Exec:        "remove2",
				},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT=1", "_REMOVE=0", "_REMOVE_OUT=a"}, Exec: "/usr/bin/ls"},
			},
			wantResult: &Result{
				ChangedIDs:     []string{"etcha push"},
				ChangedOutputs: []string{"b"},
				RemovedIDs:     []string{"2"},
				RemovedOutputs: []string{"a"},
			},
		},
		{
			name:        "good-filter",
			command:     "ls",
			destination: ts.URL + "/etcha/v1/push/etcha?filter=^123$",
			signingKey:  prv1,
			wantResult:  &Result{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c.Build.SigningKey = tc.signingKey.String()
			c.CLI.RunMockErrors(tc.mockErrors)
			c.CLI.RunMockOutputs([]string{"1", "a", "b", "c", "d", "e"})

			var err errs.Err

			var p *pattern.Pattern

			var r *Result

			p, _, err = getPushPattern(ctx, c, tc.command)
			if err == nil {
				var buildManifest string

				var runVars map[string]any

				buildManifest, runVars, err = p.BuildRun(ctx, c)
				if err == nil {
					var jwt string

					jwt, _, err = p.Sign(ctx, c, buildManifest, runVars)

					if err == nil {
						r, err = pushTarget(ctx, c, tc.destination, jwt)
					}
				}
			}

			assert.HasErr(t, err, tc.wantErr)
			assert.Equal(t, r, tc.wantResult)
			assert.Equal(t, c.CLI.RunMockInputs(), tc.wantInputs)
		})
	}

	os.RemoveAll("testdata")
}
