package pattern

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/cryptolib"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

func TestParsePatternFromImports(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.Exec.Command = "0"
	c.Sources = map[string]*config.Source{
		"1": {
			Exec: &commands.Exec{
				AllowOverride: true,
				Command:       "1",
			},
			Vars: map[string]any{
				"hello": "person",
			},
		},
		"2": {
			Exec: &commands.Exec{
				Command: "2",
			},
		},
	}

	c.Vars = map[string]any{
		"hello": "world",
		"int":   1,
		"bool":  true,
	}

	cValues := `
local config = std.native('getConfig')();

{
	run: [
		{
			id: '1',
		}
	],
	runEnv: {
		bool: '%s' % config.bool,
		int: '%s' % config.int,
		string: config.hello,
		test: '%s' % std.get(config, 'test', 'false'),
	}
}
`

	tests := map[string]struct {
		file                 string
		override             bool
		source               string
		wantErr              bool
		wantBuildExecCommand string
		wantRunExecCommand   string
		wantRunEnv           types.EnvVars
	}{
		"bad_render": {
			wantErr: true,
		},
		"no_commands": {
			file: `{
				run: [],
			}`,
			wantErr: true,
		},
		"build_fail": {
			file: `{
				build: [
					{
						id: "a",
						onChange: ["b"]
					}
				]
			}`,
			wantErr: true,
		},
		"run_fail": {
			file: `{
				run: [
					{
						id: "a",
						onChange: ["b"]
					}
				]
			}`,
			wantErr: true,
		},
		"source_override": {
			file: `{
				run: [
					{
						id: "a"
					}
				]
			}`,
			override:             true,
			source:               "2",
			wantBuildExecCommand: "2",
			wantRunExecCommand:   "2",
		},
		"pattern_override": {
			file: `{
				buildExec: {
					command: "3",
				},
				run: [
					{
						id: "a"
					}
				],
				runExec: {
					command: "4",
				},
			}`,
			override:             true,
			source:               "1",
			wantBuildExecCommand: "3",
			wantRunExecCommand:   "4",
		},
		"no_override": {
			file: `{
				buildExec: {
					command: "3",
				},
				run: [
					{
						id: "a"
					}
				],
				runExec: {
					command: "4",
				},
			}`,
			override:             false,
			source:               "1",
			wantBuildExecCommand: "0",
			wantRunExecCommand:   "0",
		},
		"override_nothing": {
			file: `{
				run: [
					{
						id: "a"
					}
				],
			}`,
			override:             true,
			source:               "1",
			wantBuildExecCommand: "1",
			wantRunExecCommand:   "1",
		},
		"config values": {
			file:                 cValues,
			source:               "1",
			wantBuildExecCommand: "0",
			wantRunExecCommand:   "0",
			wantRunEnv: types.EnvVars{
				"bool":   "true",
				"int":    "1",
				"string": "person",
				"test":   "false",
			},
		},
		"test": {
			file:                 cValues,
			source:               "test",
			wantBuildExecCommand: "0",
			wantRunExecCommand:   "0",
			wantRunEnv: types.EnvVars{
				"bool":   "true",
				"int":    "1",
				"string": "world",
				"test":   "true",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.Exec.AllowOverride = tc.override
			p, err := ParsePatternFromImports(ctx, c, tc.source, &jsonnet.Imports{
				Entrypoint: "/main.jsonnet",
				Files: map[string]string{
					"/main.jsonnet": tc.file,
				},
			})
			assert.Equal(t, err != nil, tc.wantErr)
			if !tc.wantErr {
				assert.Equal(t, p.BuildExec.Command, tc.wantBuildExecCommand)
				assert.Equal(t, p.RunExec.Command, tc.wantRunExecCommand)
				assert.Equal(t, len(p.Run), 1)

				if tc.wantRunEnv != nil {
					assert.Equal(t, p.RunEnv, tc.wantRunEnv)
				}
			}
		})
	}
}

func TestPatternSign(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()

	prv, pub, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)

	p := Pattern{
		Audience:     []string{"audience!"},
		ExpiresInSec: 59,
		Imports: &jsonnet.Imports{
			Entrypoint: "/main.jsonnet",
			Files: map[string]string{
				"/main.jsonnet": `{
					run: [
						id: "1"
					]
				}`,
			},
		},
		Issuer: "issuer!",
		RunEnv: types.EnvVars{
			"world": "hello",
		},
		Subject: "subject!",
	}

	j, err := p.Sign(ctx, c, "build", map[string]string{"hello": "world"})
	assert.HasErr(t, err, ErrPatternMissingKey)
	assert.Equal(t, j, "")

	cli.BuildVersion = "v2023.10.02"
	c.Build.SigningKey = prv.String()
	c.Run.VerifyKeys = cryptolib.Keys[cryptolib.KeyProviderPublic]{
		pub,
	}

	j, err = p.Sign(ctx, c, "build", map[string]string{"hello": "world"})
	assert.HasErr(t, err, nil)
	assert.Equal(t, j != "", true)

	jw, err := ParseJWT(ctx, c, j, "")
	assert.HasErr(t, err, nil)
	assert.Equal(t, jw.Audience, p.Audience)
	assert.Equal(t, jw.EtchaBuildManifest, "build")
	assert.Equal(t, jw.EtchaPattern, p.Imports)
	assert.Equal(t, jw.EtchaRunEnv, map[string]string{"hello": "world", "world": "hello"})
	assert.Equal(t, jw.EtchaVersion, "v2023.10.02")
	assert.Equal(t, time.Unix(jw.ExpiresAt, 0).Before(time.Now().Add(1*time.Minute)), true)
	assert.Equal(t, jw.Issuer, p.Issuer)
	assert.Equal(t, jw.Subject, p.Subject)

	c.CLI.RunMock()
	c.CLI.RunMockErrors([]error{
		ErrBuildEmpty,
	})
	c.CLI.RunMockOutputs([]string{
		"",
		strings.Split(jw.Raw, ".")[0],
		strings.Split(jw.Raw, ".")[2],
	})

	cli.SetStdin("password\npassword\n")

	ev, _ := cryptolib.KDFSet(cryptolib.Argon2ID, "123", []byte(prv.String()), cryptolib.EncryptionBest)

	c.Build.SigningKey = ev.String()

	cli.SetStdin("password")

	j, err = p.Sign(ctx, c, "build", map[string]string{"hello": "world"})
	assert.HasErr(t, err, nil)
	assert.Equal(t, j != "", true)

	jw, err = ParseJWT(ctx, c, j, "")
	assert.HasErr(t, err, nil)
	assert.Equal(t, jw.EtchaRunEnv, map[string]string{"hello": "world", "world": "hello"})

	cli.SetStdin("wrong")

	j, err = p.Sign(ctx, c, "build", map[string]string{"hello": "world"})
	assert.HasErr(t, err, ErrPatternMissingKey)
	assert.Equal(t, j, "")

	c.Exec.AllowOverride = true
	c.Build.SigningKey = ""
	c.Build.SigningExec = &commands.Exec{
		Command: "hello",
	}
	c.Build.SigningCommands = append(commands.Commands{
		{
			Always: true,
			Change: "changeA",
			ID:     "a",
			OnChange: []string{
				"etcha:jwt",
			},
		},
	}, c.Build.SigningCommands...)
	c.CLI.RunMockErrors([]error{
		nil,
		ErrBuildEmpty,
	})
	c.CLI.RunMockOutputs([]string{
		jw.Raw,
		"",
		strings.Split(jw.Raw, ".")[0],
		strings.Split(jw.Raw, ".")[2],
	})

	out, err := p.Sign(ctx, c, "", nil)
	assert.HasErr(t, err, nil)

	in := c.CLI.RunMockInputs()
	assert.Equal(t, len(in), 1)
	assert.Equal(t, in[0].Exec, "hello changeA")
	assert.Equal(t, out, jw.Raw)
}
