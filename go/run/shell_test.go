package run

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cryptolib"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

func TestShell(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	ctx = logger.SetNoColor(ctx, true)
	prv1, pub1, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)
	prv2, _, _ := cryptolib.NewKeysAsymmetric(cryptolib.AlgorithmBest)
	c := config.Default()
	c.CLI.NoColor = true
	c.Run.VerifyKeys = cryptolib.Keys[cryptolib.KeyProviderPublic]{
		pub1,
	}
	c.Sources["shell"] = &config.Source{
		Shell: "/bin/bash",
	}
	c.Sources["noshell"] = &config.Source{}
	s, _ := newState(ctx, c)
	m, _ := s.newMux(ctx)
	ts := httptest.NewServer(m)
	u, _ := url.Parse(ts.URL)

	tests := []struct {
		hostname   string
		name       string
		signingKey cryptolib.Key[cryptolib.KeyProviderPrivate]
		source     string
		stdin      string
		wantErr    error
		wantOut    string
	}{
		{
			name:    "no signing key",
			wantErr: config.ErrMissingBuildKey,
		},
		{
			name:       "no target",
			signingKey: prv1,
			wantErr:    errs.ErrReceiver,
		},
		{
			name:       "no source",
			hostname:   "localhost",
			signingKey: prv1,
			wantErr:    errs.ErrReceiver,
		},
		{
			name:       "wrong source",
			hostname:   "localhost",
			source:     "noshell",
			signingKey: prv1,
			wantErr:    errs.ErrReceiver,
			wantOut:    "unknown source",
		},
		{
			name:       "bad signing key",
			hostname:   "localhost",
			source:     "shell",
			signingKey: prv2,
			wantErr:    errs.ErrReceiver,
			wantOut:    cryptolib.ErrVerify.Error(),
		},
		{
			name:       "good signing key",
			hostname:   "localhost",
			stdin:      "ls run*",
			source:     "shell",
			signingKey: prv1,
			wantOut:    "run.go\nrun_test.go",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger.SetStd()
			logger.SetStdin(tc.stdin + "&& exit\n\n")

			c.Build = config.Build{
				SigningKey: tc.signingKey.String(),
			}

			assert.HasErr(t, Shell(ctx, c, config.Target{
				Hostname:  tc.hostname,
				Insecure:  true,
				PathShell: "/etcha/v1/shell",
				Port:      u.Port(),
			}, tc.source), tc.wantErr)

			if tc.wantOut != "" {
				assert.Contains(t, logger.ReadStd(), tc.wantOut)
			}
		})
	}
}
