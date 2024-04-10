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

	prv, pub, _ := cryptolib.NewKeysAsymmetric(cryptolib.BestEncryptionAsymmetric)
	cli.LicensePublicKeys = pub.String()

	assert.Equal(t, c.License, License{
		Commands: 10,
		Sources:  1,
		Targets:  3,
	})

	c.Sources = map[string]*Source{
		"a": {},
		"b": {},
	}
	claims := License{
		Sources: 2,
	}

	assert.Contains(t, c.Parse(ctx, nil).Error(), "number of sources")

	token, _, _ := jwt.New(&claims, time.Now().Add(10*time.Second), nil, "", "", "A")
	token.Sign(prv)
	c.LicenseKey = token.String()

	assert.HasErr(t, c.Parse(ctx, nil), nil)

	c.Targets = map[string]Target{
		"a": {},
		"b": {},
		"c": {},
		"d": {},
	}
	claims.Targets = 4

	assert.Contains(t, c.Parse(ctx, nil).Error(), "number of targets")

	token, _, _ = jwt.New(&claims, time.Now().Add(10*time.Second), nil, "", "", "A")
	token.Sign(prv)
	c.LicenseKey = token.String()

	assert.HasErr(t, c.Parse(ctx, nil), nil)

	c.Sources["a"].Commands = []*commands.Command{
		{
			ID: "a",
		},
		{
			ID: "a",
		},
		{
			ID: "a",
		},
		{
			ID: "a",
		},
		{
			ID: "a",
		},
		{
			ID: "a",
		},
		{
			ID: "a",
		},
		{
			ID: "a",
		},
		{
			ID: "a",
		},
		{
			ID: "a",
		},
		{
			ID: "a",
		},
	}
	claims.Commands = 11

	assert.Contains(t, c.Parse(ctx, nil).Error(), "number of commands")

	token, _, _ = jwt.New(&claims, time.Now().Add(10*time.Second), nil, "", "", "A")
	token.Sign(prv)
	c.LicenseKey = token.String()

	assert.HasErr(t, c.Parse(ctx, nil), nil)

	os.Remove("config.jsonnet")
	os.Remove("run")
}

func TestGetSysInfo(t *testing.T) {
	s := GetSysInfo()

	assert.Equal(t, s.CPULogical != 0, true)
	assert.Equal(t, s.Hostname != "", true)
	assert.Equal(t, s.MemoryTotal != 0, true)
	assert.Equal(t, s.OSType != "", true)
	assert.Equal(t, s.RuntimeArch != "", true)
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

func TestSignJWT(t *testing.T) {
	ctx := context.Background()
	c := Default()
	j := &jwt.Token{
		HeaderBase64:  "header",
		PayloadBase64: "payload",
	}

	// No keys
	out, err := c.SignJWT(ctx, j)
	assert.HasErr(t, err, ErrMissingBuildKey)
	assert.Equal(t, out == "", true)

	prv, pub, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)
	c.Build.SigningKey = prv.String()
	c.Run.VerifyKeys = cryptolib.Keys[cryptolib.KeyProviderPublic]{
		pub,
	}

	out, err = c.SignJWT(ctx, j)
	assert.HasErr(t, err, nil)
	assert.Equal(t, out == "", false)

	// With key
	c.Build.SigningKey = prv.String()

	out, err = c.SignJWT(ctx, j)
	assert.HasErr(t, err, nil)
	assert.Equal(t, out == "", false)

	// With encrypted key password
	logger.SetStdin("password\npassword\n")

	ev, _ := cryptolib.KDFSet(cryptolib.Argon2ID, "123", []byte(prv.String()), cryptolib.EncryptionBest)
	c.Build.SigningKey = ev.String()

	logger.SetStdin("password")

	out, err = c.SignJWT(ctx, j)
	assert.HasErr(t, err, nil)
	assert.Equal(t, out == "", false)

	// With wrong encrypted key password
	logger.SetStdin("wrong")

	c.Build.key = cryptolib.Key[cryptolib.KeyProviderPrivate]{}

	out, err = c.SignJWT(ctx, j)
	assert.HasErr(t, err, ErrMissingBuildKey)
	assert.Equal(t, out == "", true)

	c.CLI.RunMock()
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
	c.CLI.RunMockOutputs([]string{
		"output",
	})

	out, err = c.SignJWT(ctx, j)
	assert.HasErr(t, err, nil)
	assert.Equal(t, out == "", false)

	in := c.CLI.RunMockInputs()
	assert.Equal(t, len(in), 1)
	assert.Equal(t, in[0].Exec, "hello changeA")
	assert.Equal(t, out, "output")
}
