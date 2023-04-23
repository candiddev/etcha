package pattern

import (
	"context"
	"testing"
	"time"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/crypto"
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
			Exec: commands.Exec{
				Command:  "1",
				Override: true,
			},
		},
		"2": {
			Exec: commands.Exec{
				Command: "2",
			},
		},
	}

	tests := map[string]struct {
		file            string
		override        bool
		source          string
		wantErr         bool
		wantExecCommand string
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
			override:        true,
			source:          "2",
			wantExecCommand: "2",
		},
		"pattern_override": {
			file: `{
				exec: {
					command: "3",
				},
				run: [
					{
						id: "a"
					}
				]
			}`,
			override:        true,
			source:          "1",
			wantExecCommand: "3",
		},
		"no_override": {
			file: `{
				exec: {
					command: "3",
				},
				run: [
					{
						id: "a"
					}
				]
			}`,
			override:        false,
			source:          "1",
			wantExecCommand: "0",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.Exec.Override = tc.override
			p, err := ParsePatternFromImports(ctx, c, tc.source, &jsonnet.Imports{
				Entrypoint: "/main.jsonnet",
				Files: map[string]string{
					"/main.jsonnet": tc.file,
				},
			})
			assert.Equal(t, err != nil, tc.wantErr)
			if !tc.wantErr {
				assert.Equal(t, p.Exec.Command, tc.wantExecCommand)
				assert.Equal(t, len(p.Run), 1)
			}
		})
	}
}

func TestPatternSign(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()

	prv, pub, _ := crypto.NewEd25519()

	p := Pattern{
		Audience:     "audience!",
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
	c.JWT.PrivateKey = prv
	c.JWT.PublicKeys = crypto.Ed25519PublicKeys{
		pub,
	}

	j, err = p.Sign(ctx, c, "build", map[string]string{"hello": "world"})
	assert.HasErr(t, err, nil)
	assert.Equal(t, j != "", true)

	jw, err := ParseJWT(ctx, c, j, "")
	assert.HasErr(t, err, nil)
	assert.Equal(t, jw.Audience[0], p.Audience)
	assert.Equal(t, jw.EtchaBuildManifest, "build")
	assert.Equal(t, jw.EtchaPattern, p.Imports)
	assert.Equal(t, jw.EtchaRunEnv, map[string]string{"hello": "world", "world": "hello"})
	assert.Equal(t, jw.EtchaVersion, "v2023.10.02")
	assert.Equal(t, jw.ExpiresAt.Time.Before(time.Now().Add(1*time.Minute)), true)
	assert.Equal(t, jw.Issuer, p.Issuer)
	assert.Equal(t, jw.Subject, p.Subject)
}
