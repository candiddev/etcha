package run

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/cryptolib"
	"github.com/candiddev/shared/go/get"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

func TestRun(t *testing.T) {
	logger.UseTestLogger(t)

	_, pub, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)

	ctx := context.Background()
	c := config.Default()
	c.CLI.RunMock()
	c.Run.VerifyKeys = cryptolib.Keys[cryptolib.KeyProviderPublic]{
		pub,
	}
	c.Run.ListenAddress = ""
	c.Sources = map[string]*config.Source{
		"etcha": {},
	}

	tests := []struct {
		address string
		check   bool
		name    string
		once    bool
	}{
		{
			name:  "check",
			check: true,
		},
		{
			name: "once",
			once: true,
		},
		{
			address: ":4001",
			name:    "listen",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctx)
			g := runtime.NumGoroutine()
			c.Sources["etcha"].CheckOnly = true
			c.Run.ListenAddress = tc.address

			go Run(ctx, c, tc.once)

			switch {
			case tc.address != "":
				time.Sleep(500 * time.Millisecond)
				assert.Equal(t, runtime.NumGoroutine(), g+4) // pulltargets + shutdown watch + http.server + rate limiter
			case tc.check:
				time.Sleep(700 * time.Millisecond)
				assert.Equal(t, runtime.NumGoroutine(), g+1)
			default:
				time.Sleep(500 * time.Millisecond)
				assert.Equal(t, runtime.NumGoroutine(), g)
			}

			cancel()

			if tc.address != "" {
				time.Sleep(500 * time.Millisecond)
				assert.Equal(t, runtime.NumGoroutine(), g+1) // rate limiter
			} else {
				time.Sleep(700 * time.Millisecond)
				assert.Equal(t, runtime.NumGoroutine(), g) // rate limiter
			}
		})
	}
}

// This is pretty close to an end-to-end test.
func TestStateDiffExec(t *testing.T) {
	logger.UseTestLogger(t)

	os.MkdirAll("testdata", 0700)

	ctx := context.Background()
	c := config.Default()
	c.CLI.RunMock()
	c.Exec.AllowOverride = true
	c.Run.StateDir = "testdata"
	c.Sources = map[string]*config.Source{
		"etcha": {
			Exec: &commands.Exec{
				AllowOverride: true,
			},
		},
	}
	s, _ := newState(ctx, c)

	tests := []struct {
		check       bool
		j           *pattern.JWT
		mockErrors  []error
		name        string
		noRestore   bool
		runAll      bool
		triggerOnly bool
		wantErr     error
		wantInputs  []cli.RunMockInput
		wantResult  *Result
		wantJWT     string
	}{
		{
			name:    "nil_jwt",
			wantErr: ErrNilJWT,
			wantResult: &Result{
				Err: ErrNilJWT.Error(),
			},
		},
		{
			name: "bad_pattern",
			j: &pattern.JWT{
				EtchaPattern: &jsonnet.Imports{
					Entrypoint: "/main.jsonnet",
				},
			},
			wantErr: commands.ErrCommandsEmpty,
			wantResult: &Result{
				Err: commands.ErrCommandsEmpty.Error(),
			},
		},
		{
			name: "err_run",
			mockErrors: []error{
				ErrNilJWT,
				ErrNilJWT,
			},
			j: &pattern.JWT{
				EtchaPattern: &jsonnet.Imports{
					Entrypoint: "/main.jsonnet",
					Files: map[string]string{
						"/main.jsonnet": `{run:[{change:"changeA",check:"checkA",id:"a",remove:"removeA"}]}`,
					},
				},
			},
			wantErr: ErrNilJWT,
			wantInputs: []cli.RunMockInput{
				{Exec: "checkA"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT="}, Exec: "changeA"},
			},
			wantResult: &Result{
				Err: "error changing id a: error running commands: " + ErrNilJWT.Error() + ": ",
			},
		},
		{
			name:  "check",
			check: true,
			mockErrors: []error{
				ErrNilJWT,
			},
			j: &pattern.JWT{
				EtchaPattern: &jsonnet.Imports{
					Entrypoint: "/main.jsonnet",
					Files: map[string]string{
						"/main.jsonnet": `{run:[{change:"changeA",check:"checkA",id:"a",remove:"removeA"}]}`,
					},
				},
				EtchaRunEnv: types.EnvVars{
					"hello": "world",
				},
				Raw: "hello",
			},
			wantInputs: []cli.RunMockInput{
				{Environment: []string{"hello=world"}, Exec: "checkA"},
			},
			wantResult: &Result{
				ChangedIDs: []string{"a"},
			},
		},
		{
			name: "good",
			mockErrors: []error{
				ErrNilJWT,
			},
			j: &pattern.JWT{
				EtchaPattern: &jsonnet.Imports{
					Entrypoint: "/main.jsonnet",
					Files: map[string]string{
						"/main.jsonnet": `{run:[{change:"changeA",check:"checkA",id:"a",remove:"removeA"}]}`,
					},
				},
				Raw: "hello",
			},
			wantInputs: []cli.RunMockInput{
				{Exec: "checkA"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT="}, Exec: "changeA"},
			},
			wantJWT: "hello",
			wantResult: &Result{
				ChangedIDs:     []string{"a"},
				ChangedOutputs: []string{""},
			},
		},
		{
			name: "good_nochange",
			j: &pattern.JWT{
				EtchaPattern: &jsonnet.Imports{
					Entrypoint: "/main.jsonnet",
					Files: map[string]string{
						"/main.jsonnet": `{run:[{change:"changeA",check:"checkA",id:"a",remove:"removeA"}]}`,
					},
				},
				Raw: "anew",
			},
			wantJWT:    "anew",
			wantResult: &Result{},
		},
		{
			name: "good_runAll",
			j: &pattern.JWT{
				EtchaPattern: &jsonnet.Imports{
					Entrypoint: "/main.jsonnet",
					Files: map[string]string{
						"/main.jsonnet": `{run:[{change:"changeA",check:"checkA",id:"a",remove:"removeA"}]}`,
					},
				},
				Raw: "anew2",
			},
			runAll: true,
			wantInputs: []cli.RunMockInput{
				{Exec: "checkA"},
			},
			wantJWT:    "anew2",
			wantResult: &Result{},
		},
		{
			name: "good_triggerOnly",
			j: &pattern.JWT{
				EtchaPattern: &jsonnet.Imports{
					Entrypoint: "/main.jsonnet",
					Files: map[string]string{
						"/main.jsonnet": `{run:[{change:"changeA",check:"checkA",id:"a",remove:"removeA"}]}`,
					},
				},
				Raw: "anew2",
			},
			triggerOnly: true,
			wantJWT:     "anew2",
			wantResult:  &Result{},
		},
		{
			name: "good_noRestore",
			j: &pattern.JWT{
				EtchaPattern: &jsonnet.Imports{
					Entrypoint: "/main.jsonnet",
					Files: map[string]string{
						"/main.jsonnet": `{run:[{change:"changeB",check:"checkB",id:"a",remove:"removeB"}]}`,
					},
				},
				Raw: "anew3",
			},
			noRestore: true,
			wantJWT:   "anew2",
			wantInputs: []cli.RunMockInput{
				{
					Exec: "checkB",
				},
			},
			wantResult: &Result{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s.Config.CLI.RunMockErrors(tc.mockErrors)
			s.Config.Sources["etcha"].NoRestore = tc.noRestore
			s.Config.Sources["etcha"].RunAll = tc.runAll
			s.Config.Sources["etcha"].TriggerOnly = tc.triggerOnly

			r, err := s.diffExec(ctx, tc.check, "etcha", tc.j)

			assert.HasErr(t, err, tc.wantErr)
			assert.Equal(t, r, tc.wantResult)
			assert.Equal(t, s.Config.CLI.RunMockInputs(), tc.wantInputs)

			j, _ := os.ReadFile("testdata/etcha.jwt")
			assert.Equal(t, string(j), tc.wantJWT)
		})
	}

	os.RemoveAll("testdata")
}

func TestStateLoadExecJWTs(t *testing.T) {
	logger.UseTestLogger(t)

	os.MkdirAll("testdata", 0700)

	ctx := context.Background()
	ctx = logger.SetFormat(ctx, logger.FormatKV)
	c := config.Default()
	c.CLI.RunMock()
	c.Exec.Command = ""
	c.Exec.WorkDir = ""
	c.Run.StateDir = "testdata"
	c.Sources = map[string]*config.Source{
		"1": {
			CheckOnly: true,
		},
		"2": {
			CheckOnly: true,
			Commands: commands.Commands{
				{
					ID:    "d",
					Check: "checkD",
				},
			},
		},
		"3": {
			CheckOnly: true,
			Commands: commands.Commands{
				{
					ID:    "b",
					Check: "checkB",
				},
			},
		},
		"4": {
			Commands: commands.Commands{
				{
					ID:    "z",
					Check: "checkZ",
				},
			},
			TriggerOnly: true,
		},
	}

	prv, pub, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)
	c.Build.SigningKey = prv.String()
	c.Run.VerifyKeys = cryptolib.Keys[cryptolib.KeyProviderPublic]{
		pub,
	}

	jwt1 := pattern.Pattern{
		Imports: &jsonnet.Imports{
			Entrypoint: "/main.jsonnet",
		},
	}
	j, _ := jwt1.Sign(ctx, c, "", nil)
	os.WriteFile("testdata/1.jwt", []byte(j), 0600)

	jwt2 := pattern.Pattern{
		Audience: []string{
			"2",
		},
		Imports: &jsonnet.Imports{
			Entrypoint: "/main.jsonnet",
			Files: map[string]string{
				"/main.jsonnet": `{run:[{check:"checkA",id:"a"}]}`,
			},
		},
	}
	j, _ = jwt2.Sign(ctx, c, "", nil)
	os.WriteFile("testdata/2.jwt", []byte(j), 0600)

	s, _ := newState(ctx, c)

	s.loadExecJWTs(ctx)

	assert.Equal(t, s.JWTs.Get("2").Audience[0], "2")
	assert.Equal(t, s.JWTs.Keys(), []string{"2"})
	assert.Equal(t, s.Patterns.Get("2").Run[0].ID, "a")
	assert.Equal(t, s.Patterns.Keys(), []string{"2", "3", "4"})
	assert.Equal(t, s.Config.CLI.RunMockInputs(), []cli.RunMockInput{{Exec: "checkA"}, {Exec: "checkB"}})

	os.RemoveAll("testdata")
}

func TestStateRunSource(t *testing.T) {
	logger.UseTestLogger(t)

	os.MkdirAll("testdata", 0700)

	ctx := context.Background()
	c := config.Default()
	s, _ := newState(ctx, c)
	c.CLI.RunMock()
	c.Exec.Command = ""
	c.Exec.WorkDir = ""
	c.Run.StateDir = "testdata"
	prv, pub, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)
	c.Build.SigningKey = prv.String()
	c.Run.VerifyKeys = cryptolib.Keys[cryptolib.KeyProviderPublic]{
		pub,
	}

	jwt1 := pattern.Pattern{
		Imports: &jsonnet.Imports{
			Entrypoint: "/main.jsonnet",
			Files: map[string]string{
				"/main.jsonnet": `{run:[{check:"checkA",id:"a"}]}`,
			},
		},
	}
	cli.BuildVersion = "v2023.10.02"
	j1, _ := jwt1.Sign(ctx, c, "", nil)

	j, _ := pattern.ParseJWT(ctx, c, j1, "1")
	p, _ := j.Pattern(ctx, c, "1")
	s.JWTs.Set("1", j)
	s.Patterns.Set("1", p)

	jwt2 := pattern.Pattern{
		Imports: &jsonnet.Imports{
			Entrypoint: "/main.jsonnet",
			Files: map[string]string{
				"/main.jsonnet": `{run:[{change:"changeA",check:"checkA",id:"b"}]}`,
			},
		},
	}
	cli.BuildVersion = "v2023.10.03"
	j2, _ := jwt2.Sign(ctx, c, "", nil)

	ts := get.NewHTTPMock([]string{"/1.jwt"}, []byte(j2), time.Time{})

	c.Sources = map[string]*config.Source{
		"1": {
			PullIgnoreVersion: true,
			PullPaths:         []string{ts.URL() + "/1.jwt"},
		},
	}

	tests := []struct {
		mockErrors  []error
		mockInputs  []cli.RunMockInput
		name        string
		wantErr     error
		wantResults *Result
	}{
		{
			name: "errors",
			mockErrors: []error{
				ErrNilJWT,
				ErrNilJWT,
			},
			mockInputs: []cli.RunMockInput{
				{Exec: "checkA"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT=a"}, Exec: "changeA"},
			},
			wantErr: ErrNilJWT,
			wantResults: &Result{
				Err: "error changing id b: error running commands: received an empty JWT, this is probably a bug: b",
			},
		},
		{
			name: "no_error",
			mockErrors: []error{
				ErrNilJWT,
			},
			mockInputs: []cli.RunMockInput{
				{Exec: "checkA"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT=a"}, Exec: "changeA"},
			},
			wantResults: &Result{
				ChangedIDs: []string{
					"b",
				},
				ChangedOutputs: []string{"b"},
			},
		},
		{
			name:        "no_diff",
			wantResults: &Result{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c.CLI.RunMockErrors(tc.mockErrors)
			c.CLI.RunMockOutputs([]string{"a", "b", "c", "d", "e"})

			r, err := s.runSource(ctx, "1")
			assert.HasErr(t, err, tc.wantErr)
			assert.Equal(t, c.CLI.RunMockInputs(), tc.mockInputs)
			assert.Equal(t, r, tc.wantResults)
		})
	}

	j3, _ := os.ReadFile("testdata/1.jwt")
	assert.Equal(t, string(j3), j2)

	os.RemoveAll("testdata")
}
