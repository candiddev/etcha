package run

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/handlers"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/crypto"
	"github.com/candiddev/shared/go/get"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

func TestRun(t *testing.T) {
	logger.UseTestLogger(t)

	_, pub, _ := crypto.NewEd25519()

	ctx := context.Background()
	c := config.Default()
	c.CLI.RunMock()
	c.JWT.PublicKeys = crypto.Ed25519PublicKeys{
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
	c.Exec.Override = true
	c.Handlers = handlers.Handlers{
		"test": {
			Exec: commands.Exec{
				Command: "handler",
			},
			EventExit: true,
			Events: []string{
				"test",
			},
		},
	}
	c.Run.StateDir = "testdata"
	c.Sources = map[string]*config.Source{
		"etcha": {
			Exec: commands.Exec{
				Override: true,
			},
		},
	}
	s := newState(c)

	tests := []struct {
		alwaysCheck bool
		check       bool
		j           *pattern.JWT
		mockErrors  []error
		name        string
		wantErr     error
		wantInputs  []cli.RunMockInput
		wantResult  *PushResult
		wantJWT     string
	}{
		{
			name:    "nil_jwt",
			wantErr: ErrNilJWT,
			wantResult: &PushResult{
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
			wantResult: &PushResult{
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
				{Exec: " checkA"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT="}, Exec: " changeA"},
			},
			wantResult: &PushResult{
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
				{Environment: []string{"hello=world"}, Exec: " checkA"},
			},
			wantResult: &PushResult{
				Changed: []string{"a"},
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
				{Exec: " checkA"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT="}, Exec: " changeA"},
			},
			wantJWT: "hello",
			wantResult: &PushResult{
				Changed: []string{"a"},
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
			wantResult: &PushResult{},
		},
		{
			name: "good_alwayscheck",
			j: &pattern.JWT{
				EtchaPattern: &jsonnet.Imports{
					Entrypoint: "/main.jsonnet",
					Files: map[string]string{
						"/main.jsonnet": `{run:[{change:"changeA",check:"checkA",id:"a",remove:"removeA"}]}`,
					},
				},
				Raw: "anew2",
			},
			wantInputs: []cli.RunMockInput{
				{Exec: " checkA"},
			},
			wantJWT:    "anew2",
			wantResult: &PushResult{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s.Config.CLI.RunMockErrors(tc.mockErrors)
			s.Config.Sources["etcha"].RunAlwaysCheck = tc.alwaysCheck

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
	c := config.Default()
	c.CLI.RunMock()
	c.Run.StateDir = "testdata"
	c.Sources = map[string]*config.Source{
		"1": {
			CheckOnly: true,
		},
		"2": {
			CheckOnly: true,
		},
	}

	prv, pub, _ := crypto.NewEd25519()
	c.JWT.PrivateKey = prv
	c.JWT.PublicKeys = crypto.Ed25519PublicKeys{
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
		Audience: "2",
		Imports: &jsonnet.Imports{
			Entrypoint: "/main.jsonnet",
			Files: map[string]string{
				"/main.jsonnet": `{run:[{check:"checkA",id:"a"}]}`,
			},
		},
	}
	j, _ = jwt2.Sign(ctx, c, "", nil)
	os.WriteFile("testdata/2.jwt", []byte(j), 0600)

	s := newState(c)

	s.loadExecJWTs(ctx)

	assert.Equal(t, s.JWTs["2"].Audience[0], "2")
	assert.Equal(t, s.Patterns["2"].Run[0].ID, "a")
	assert.Equal(t, s.Config.CLI.RunMockInputs(), []cli.RunMockInput{{Exec: " checkA"}})

	os.RemoveAll("testdata")
}

func TestStateRunSource(t *testing.T) {
	logger.UseTestLogger(t)

	os.MkdirAll("testdata", 0700)

	ctx := context.Background()
	c := config.Default()
	s := newState(c)
	c.CLI.RunMock()
	c.Run.StateDir = "testdata"
	prv, pub, _ := crypto.NewEd25519()
	c.JWT.PrivateKey = prv
	c.JWT.PublicKeys = crypto.Ed25519PublicKeys{
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
	s.JWTs["1"] = j
	s.Patterns["1"] = p

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

	c.CLI.RunMockErrors([]error{
		ErrNilJWT,
		ErrNilJWT,
	})

	assert.HasErr(t, s.runSource(ctx, "1"), ErrNilJWT)
	assert.Equal(t, c.CLI.RunMockInputs(), []cli.RunMockInput{
		{Exec: " checkA"},
		{Environment: []string{"_CHECK=1", "_CHECK_OUT="}, Exec: " changeA"},
	})
	assert.HasErr(t, s.runSource(ctx, "1"), nil)
	assert.Equal(t, c.CLI.RunMockInputs(), []cli.RunMockInput{
		{Exec: " checkA"},
	})
	assert.HasErr(t, s.runSource(ctx, "1"), nil)
	assert.Equal(t, c.CLI.RunMockInputs(), nil)
	c.Sources["1"].RunAlwaysCheck = true

	assert.HasErr(t, s.runSource(ctx, "1"), nil)
	assert.Equal(t, c.CLI.RunMockInputs(), []cli.RunMockInput{
		{Exec: " checkA"},
	})

	j3, _ := os.ReadFile("testdata/1.jwt")
	assert.Equal(t, string(j3), j2)

	os.RemoveAll("testdata")
}
