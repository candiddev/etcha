package main

import (
	"context"
	"os"
	"testing"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cryptolib"
	"github.com/candiddev/shared/go/logger"
)

func TestBuild(t *testing.T) {
	logger.UseTestLogger(t)

	c := config.Default()
	c.CLI.RunMock()

	ctx := context.Background()
	ctx = logger.SetNoColor(ctx, true)

	os.MkdirAll("testdata/src", 0700)
	os.MkdirAll("testdata/dst", 0700)
	os.WriteFile("testdata/src/main.jsonnet", []byte(`{build:[{id: 'a',always:true,change:'%s' % std.native('getConfig')().vars.test}],buildExec:{}}`), 0600)

	prv, _, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)
	c.Build.SigningKey = prv.String()

	assert.HasErr(t, build.Run(ctx, []string{"", "testdata/src/main.jsonnet", "testdata/dst/main.jwt"}, nil, c), nil)
	assert.Equal(t, c.CLI.RunMockInputs()[0].Exec, `/usr/bin/false`)

	os.RemoveAll("testdata")
}
