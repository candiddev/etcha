package pattern

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cryptolib"
	"github.com/candiddev/shared/go/get"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/jwt"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

func TestParseJWT(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	prv, pub, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)

	p := Pattern{}

	tests := map[string]struct {
		key     cryptolib.Key[cryptolib.KeyProviderPublic]
		wantErr error
		wantOut string
	}{
		"missing keys": {
			wantErr: jwt.ErrParseNoPublicKeys,
			wantOut: "hello",
		},
		"good": {
			key:     pub,
			wantOut: "hello",
		},
	}

	// ParseJWT
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c := config.Default()
			c.Build.SigningKey = prv.String()

			content, _ := p.Sign(ctx, c, "hello", nil)

			if !tc.key.IsNil() {
				c.Sources = map[string]*config.Source{
					"etcha": {
						VerifyKeys: cryptolib.Keys[cryptolib.KeyProviderPublic]{
							tc.key,
						},
					},
				}
			}

			j, err := ParseJWT(ctx, c, content, "etcha")
			assert.HasErr(t, err, tc.wantErr)
			assert.Equal(t, j.EtchaBuildManifest, tc.wantOut)
		})
	}

	c := config.Default()
	c.Build.SigningKey = prv.String()

	jwt1, _ := p.Sign(ctx, c, "1", nil)
	jwt2, _ := p.Sign(ctx, c, "2", nil)
	jwt3, _ := p.Sign(ctx, c, "2", nil)

	os.MkdirAll("testdata/cache", 0700)
	os.WriteFile("testdata/1.jwt", []byte(jwt1), 0600)
	os.WriteFile("testdata/2.jwt", []byte(jwt2), 0600)
	os.WriteFile("testdata/3.jwt", []byte(jwt3), 0600)

	ts := get.NewHTTPMock([]string{"/test.jwt"}, []byte(jwt1), time.Time{})

	// ParseJWTFromSources
	c.Run.StateDir = "testdata/cache"
	c.Sources = map[string]*config.Source{
		"1": {
			PullPaths: []string{
				"1.jwt",
				"/1.jwt",
				ts.URL() + "/test.jwt",
			},
			VerifyKeys: cryptolib.Keys[cryptolib.KeyProviderPublic]{
				pub,
			},
		},
		"2": {
			NoRestore: true,
			PullPaths: []string{
				"testdata/2.jwt",
			},
			VerifyKeys: cryptolib.Keys[cryptolib.KeyProviderPublic]{
				pub,
			},
		},
		"3": {
			PullPaths: []string{
				"testdata/3.jwt",
			},
		},
	}
	j := ParseJWTFromSource(ctx, "1", c)
	assert.Equal(t, j.EtchaBuildManifest, "1")

	f, _ := os.ReadFile("testdata/cache/1.jwt")
	assert.Equal(t, string(f), jwt1)

	j = ParseJWTFromSource(ctx, "3", c)
	assert.Equal(t, j, nil)

	os.RemoveAll("testdata")
}

func TestJWTEqual(t *testing.T) {
	j1 := JWT{
		EtchaBuildManifest: "build",
		EtchaPattern: &jsonnet.Imports{
			Entrypoint: "hello",
		},
		EtchaVersion: "version",
	}
	j2 := j1
	j2.EtchaPattern = &jsonnet.Imports{
		Entrypoint: "hello",
	}

	assert.HasErr(t, j1.Equal(&j2, false), nil)

	j2.EtchaVersion = "2"
	assert.HasErr(t, j1.Equal(&j2, true), nil)
	assert.HasErr(t, j1.Equal(&j2, false), ErrEqualVersion)

	j2.EtchaVersion = j1.EtchaVersion
	j2.EtchaBuildManifest = "2"
	assert.HasErr(t, j1.Equal(&j2, false), ErrEqualBuildManifest)

	j2.EtchaBuildManifest = j1.EtchaBuildManifest
	j2.EtchaPattern.Entrypoint = "2"
	assert.HasErr(t, j1.Equal(&j2, false), ErrEqualPattern)
}

func TestJWTPattern(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.Exec.AllowOverride = true
	c.Sources = map[string]*config.Source{
		"etcha": {
			Exec: &commands.Exec{
				Command: "hello",
			},
		},
	}

	j := JWT{
		EtchaRunEnv: map[string]string{
			"hello": "world",
		},
		EtchaPattern: &jsonnet.Imports{
			Entrypoint: "/main.jsonnet",
			Files: map[string]string{
				"/main.jsonnet": `{}`,
			},
		},
		Raw: "raw",
	}

	p, err := j.Pattern(ctx, c, "etcha")
	assert.HasErr(t, err, commands.ErrCommandsEmpty)
	assert.Equal(t, p, nil)

	j.EtchaPattern.Files = map[string]string{
		"/main.jsonnet": `
		{
			run: [
				{
					id: "id"
				}
			],
			runEnv: {
				world: "hello",
			},
		}
		`,
	}

	p, err = j.Pattern(ctx, c, "etcha")
	assert.HasErr(t, err, nil)
	assert.Equal(t, p.RunExec.Command, "hello")
	assert.Equal(t, p.RunExec.Command, "hello")
	assert.Equal(t, p.JWT, "raw")
	assert.Equal(t, p.Run[0].ID, "id")
	assert.Equal(t, p.RunEnv, types.EnvVars{
		"hello": "world",
		"world": "hello",
	})
}
