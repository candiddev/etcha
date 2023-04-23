package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/crypto"
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

func TestConfigParse(t *testing.T) {
	logger.UseTestLogger(t)

	os.Unsetenv("ETCHA_RUN_DIR")

	c := Default()
	ctx := context.Background()

	os.WriteFile("config.jsonnet", []byte(`{
  jwt: {
    publicKeys: [
      'a',
      'b',
    ],
  },
  run: {
    stateDir: 'run',
  },
}`), 0600)

	wd, _ := os.Getwd()

	assert.Equal(t, c.Parse(ctx, "", "/notreal") != nil, true)
	assert.Equal(t, c.Parse(ctx, "", "config.jsonnet") == nil, true)
	assert.Equal(t, c.JWT.PublicKeys, crypto.Ed25519PublicKeys{
		"a",
		"b",
	})
	assert.Equal(t, c.Run.StateDir, filepath.Join(wd, "run"))

	os.Remove("config.jsonnet")
	os.Remove("run")
}

func TestParseJWTFile(t *testing.T) {
	os.MkdirAll("testdata", 0700)

	prv1, pub1, _ := crypto.NewEd25519()
	_, pub2, _ := crypto.NewEd25519()
	prv3, _, _ := crypto.NewEd25519()

	c := Default()
	c.JWT.PublicKeys = crypto.Ed25519PublicKeys{
		pub2,
	}

	jwt1, _ := jwt.SignJWT(prv1, &j{
		A: "prv1",
	}, time.Time{}, "", "", "")
	jwt3, _ := jwt.SignJWT(prv3, &j{
		A: "prv3",
	}, time.Time{}, "", "", "")

	os.WriteFile("testdata/jwt1.jwt", []byte(jwt1), 0600)
	os.WriteFile("testdata/jwt3.jwt", []byte(jwt3), 0600)

	ctx := context.Background()
	out := &j{}

	key, _, err := c.ParseJWTFile(ctx, out, "testdata/jwt4.jwt", crypto.Ed25519PublicKeys{pub1})
	assert.HasErr(t, err, errs.ErrReceiver)
	assert.Equal(t, key, "")
	assert.Equal(t, out.A, "")

	key, _, err = c.ParseJWTFile(ctx, out, "testdata/jwt3.jwt", crypto.Ed25519PublicKeys{pub1})
	assert.HasErr(t, err, errs.ErrReceiver)
	assert.Equal(t, key, "")
	assert.Equal(t, out.A, "prv3")

	key, _, err = c.ParseJWTFile(ctx, out, "testdata/jwt1.jwt", crypto.Ed25519PublicKeys{pub1})
	assert.HasErr(t, err, nil)
	assert.Equal(t, key, pub1)
	assert.Equal(t, out.A, "prv1")

	os.RemoveAll("testdata")
}
