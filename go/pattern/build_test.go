package pattern

import (
	"context"
	"os"
	"testing"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cryptolib"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
)

func TestPatternBuildSign(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.CLI.RunMock()

	prv, pub, _ := cryptolib.NewKeysSign()

	p := Pattern{
		Audience: []string{"a"},
		Build: commands.Commands{
			{
				Always: true,
				Change: "a",
				ID:     "a",
				OnChange: []string{
					"etcha:build_manifest",
					"etcha:run_env_hello",
				},
			},
		},
		BuildExec: &commands.Exec{
			Command: "test",
		},
		Imports: &jsonnet.Imports{
			Entrypoint: "/main.jsonnet",
			Files: map[string]string{
				"/main.jsonnet": "hello",
			},
		},
	}

	tests := map[string]struct {
		destination string
		mockErrors  []error
		signingKey  cryptolib.KeySign
		wantErr     error
	}{
		"bad_destination": {
			destination: "/something/somewhere.jwt",
			wantErr:     ErrBuildWriteJWT,
		},
		"bad_run": {
			destination: "test.jwt",
			mockErrors: []error{
				ErrBuildEmpty,
			},
			wantErr: ErrBuildEmpty,
		},
		"bad_sign": {
			destination: "test.jwt",
			wantErr:     ErrPatternMissingKey,
		},
		"good": {
			destination: "test.jwt",
			signingKey:  prv,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.CLI.RunMockInputs()
			c.CLI.RunMockErrors(tc.mockErrors)
			c.CLI.RunMockOutputs([]string{"world"})
			c.Build.SigningKey = tc.signingKey
			c.Run.VerifyKeys = cryptolib.KeysVerify{
				pub,
			}
			assert.HasErr(t, p.BuildSign(ctx, c, tc.destination), tc.wantErr)

			if tc.wantErr == nil {
				j, err := ParseJWTFromPath(ctx, c, "", "test.jwt")
				assert.HasErr(t, err, nil)
				assert.Equal(t, j.EtchaBuildManifest, "world")
				assert.Equal(t, j.EtchaRunEnv, map[string]string{
					"hello": "world",
				})
				assert.Equal(t, j.EtchaPattern, p.Imports)
			}
		})
	}

	os.Remove("test.jwt")
}
