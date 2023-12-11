package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/cryptolib"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
)

func TestCompare(t *testing.T) {
	os.MkdirAll("testdata", 0700)
	os.MkdirAll("testdata", 0700)

	ctx := context.Background()
	c := config.Default()
	c.Run.StateDir = "testdata"

	prv, pub, _ := cryptolib.NewKeysEncryptAsymmetric(cryptolib.AlgorithmBest)

	c.Build.SigningKey = prv.String()
	c.Run.VerifyKeys = cryptolib.Keys[cryptolib.KeyProviderPublic]{
		pub,
	}

	p1 := pattern.Pattern{
		Imports: &jsonnet.Imports{
			Entrypoint: "/main.jsonnet",
			Files: map[string]string{
				"/main.jsonnet": `{run:[{change:"changeA",check:"checkA",id:"b"}]}`,
			},
		},
	}
	j1, _ := p1.Sign(ctx, c, "", nil)
	os.WriteFile("testdata/1.jwt", []byte(j1), 0700)

	m1, _ := p1.Sign(ctx, c, "manifest", nil)
	os.WriteFile("testdata/m.jwt", []byte(m1), 0700)

	cli.BuildVersion = "1.1"
	v1, _ := p1.Sign(ctx, c, "", nil)
	os.WriteFile("testdata/v.jwt", []byte(v1), 0700)

	cli.BuildVersion = ""
	p2 := p1
	p2.Imports.Files["/main.jsonnet"] = `{run:[{change:"changeA",check:"checkA",id:"c"}]}`
	j2, _ := p2.Sign(ctx, c, "", nil)
	os.WriteFile("testdata/2.jwt", []byte(j2), 0700)

	tests := map[string]struct {
		args    []string
		wantErr bool
		wantOut string
	}{
		"bad_1": {
			args: []string{
				"",
				"testdata/bad.jwt",
				"testdata/1.jwt",
			},
			wantErr: true,
			wantOut: "error opening src",
		},
		"bad_2": {
			args: []string{
				"",
				"testdata/1.jwt",
				"testdata/bad.jwt",
			},
			wantErr: true,
			wantOut: "error opening src",
		},
		"manifest": {
			args: []string{
				"",
				"testdata/1.jwt",
				"testdata/m.jwt",
			},
			wantErr: true,
			wantOut: "old etchaManifest",
		},
		"pattern": {
			args: []string{
				"",
				"testdata/1.jwt",
				"testdata/2.jwt",
			},
			wantErr: true,
			wantOut: "old etchaPattern",
		},
		"version": {
			args: []string{
				"",
				"testdata/1.jwt",
				"testdata/v.jwt",
			},
			wantErr: true,
			wantOut: "old etchaVersion",
		},
		"version_ignore": {
			args: []string{
				"",
				"testdata/1.jwt",
				"testdata/v.jwt",
				"yes",
			},
		},
		"good": {
			args: []string{
				"",
				"testdata/1.jwt",
				"testdata/1.jwt",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logger.SetStd()
			assert.Equal(t, compare(ctx, tc.args, c) != nil, tc.wantErr)

			if tc.wantOut != "" {
				assert.Equal(t, strings.Contains(logger.ReadStd(), tc.wantOut), true)
			} else {
				assert.Equal(t, logger.ReadStd(), "")
			}
		})
	}

	os.RemoveAll("testdata")
}
