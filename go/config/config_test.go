package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/cryptolib"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jwt"
	"github.com/candiddev/shared/go/logger"
)

type j struct {
	A string

	jwt.RegisteredClaims
}

// GetRegisteredClaims satisfies the JWT interface.
func (j *j) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &j.RegisteredClaims
}

func (*j) Valid() error {
	return nil
}

func TestConfigParse(t *testing.T) {
	logger.UseTestLogger(t)

	_, pub1, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)
	_, pub2, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)

	os.Unsetenv("ETCHA_RUN_DIR")

	c := Default()
	ctx := context.Background()

	os.WriteFile("config.jsonnet", []byte(fmt.Sprintf(`{
  run: {
    stateDir: 'run',
    verifyKeys: [
      '%s',
      '%s',
    ],
  },
}`, pub1, pub2)), 0600)

	wd, _ := os.Getwd()

	c.CLI.ConfigPath = "/notreal"
	assert.Equal(t, c.Parse(ctx, []string{}) != nil, true)
	c.CLI.ConfigPath = "./config.jsonnet"
	assert.Equal(t, c.Parse(ctx, []string{}) == nil, true)
	assert.Equal(t, c.Run.VerifyKeys, cryptolib.Keys[cryptolib.KeyProviderPublic]{
		pub1,
		pub2,
	})
	assert.Equal(t, c.Run.StateDir, filepath.Join(wd, "run"))

	os.Remove("config.jsonnet")
	os.Remove("run")
}

func TestParseJWTFile(t *testing.T) {
	logger.UseTestLogger(t)

	os.MkdirAll("testdata", 0700)

	prv1, _, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)
	_, pub2, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)
	prv3, _, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)

	c := Default()
	c.CLI.RunMock()
	c.Run.VerifyExec = &commands.Exec{
		AllowOverride: true,
	}
	c.Run.VerifyKeys = cryptolib.Keys[cryptolib.KeyProviderPublic]{
		pub2,
	}

	e := time.Now().Add(10 * time.Second)

	t1, _, _ := jwt.New(&j{
		A: "prv1",
	}, e, []string{}, "", "", "")
	t1.Sign(prv1)

	t3, _, _ := jwt.New(&j{
		A: "prv3",
	}, e, []string{}, "", "", "")
	t3.Sign(prv3)

	c.CLI.RunMockErrors([]error{
		ErrParseJWT,
	})
	c.CLI.RunMockOutputs([]string{
		"",
		t1.String(),
	})

	c.Sources["etcha"] = &Source{
		VerifyExec: &commands.Exec{
			Command: "hello",
		},
		VerifyCommands: commands.Commands{
			{
				Change: "getJWT",
				Check:  "check",
				OnChange: []string{
					"etcha:jwt",
				},
			},
		},
	}

	os.WriteFile("testdata/jwt1.jwt", []byte(t1.String()), 0600)
	os.WriteFile("testdata/jwt3.jwt", []byte(t3.String()), 0600)

	ctx := context.Background()
	out := &j{}

	key, _, err := c.ParseJWTFile(ctx, out, "testdata/jwt4.jwt", "")
	assert.HasErr(t, err, errs.ErrReceiver)
	assert.Equal(t, key.IsNil(), true)
	assert.Equal(t, out.A, "")

	key, _, err = c.ParseJWTFile(ctx, out, "testdata/jwt3.jwt", "")
	assert.HasErr(t, err, errs.ErrReceiver)
	assert.Equal(t, key.ID, "")
	assert.Equal(t, out.A, "prv3")

	_, _, err = c.ParseJWTFile(ctx, out, "testdata/jwt1.jwt", "etcha")
	assert.HasErr(t, err, nil)
	assert.Equal(t, out.A, "prv1")
	assert.Equal(t, c.CLI.RunMockInputs(), []cli.RunMockInput{
		{
			Environment: []string{"ETCHA_JWT=" + t1.String()},
			Exec:        "hello check",
		},
		{
			Environment: []string{
				"ETCHA_JWT=" + t1.String(),
				"_CHECK=1", "_CHECK_OUT=",
			},
			Exec: "hello getJWT",
		},
	})

	_, _, err = c.ParseJWTFile(ctx, out, "testdata/jwt1.jwt", "etcha")
	assert.HasErr(t, err, cryptolib.ErrVerify)
	assert.Equal(t, c.CLI.RunMockInputs(), []cli.RunMockInput{
		{
			Environment: []string{"ETCHA_JWT=" + t1.String()},
			Exec:        "hello check",
		},
	})

	os.RemoveAll("testdata")
}
