package run

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/crypto"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
)

func TestPush(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.CLI.RunMock()

	s := newState(c)
	ts := httptest.NewServer(s.newMux(ctx))

	prv1, pub1, _ := crypto.NewEd25519()
	prv2, _, _ := crypto.NewEd25519()

	os.MkdirAll("testdata/state", 0700)

	c.Run.StateDir = "testdata/state"

	os.WriteFile("testdata/good1.jsonnet", []byte(`
{
	run: [
		{
			change: "change1",
			check: "check1",
			id: "1",
			remove: "remove1",
		}
	]
}
`), 0600)

	os.WriteFile("testdata/good2.jsonnet", []byte(`
{
	run: [
		{
			change: "change2",
			check: "check2",
			id: "2",
			remove: "remove2",
		}
	]
}
`), 0600)

	c.Sources = map[string]*config.Source{
		"etcha": {
			AllowPush: true,
			JWTPublicKeys: crypto.Ed25519PublicKeys{
				pub1,
			},
		},
		"denied": {},
	}

	tests := []struct {
		destination string
		mockErrors  []error
		name        string
		path        string
		privateKey  crypto.Ed25519PrivateKey
		wantErr     error
		wantInputs  []cli.RunMockInput
		wantResult  *PushResult
	}{
		{
			name:    "bad_path",
			path:    "testdata/not.jsonnet",
			wantErr: jsonnet.ErrImport,
		},
		{
			name:    "bad_sign",
			path:    "testdata/good1.jsonnet",
			wantErr: pattern.ErrPatternMissingKey,
		},
		{
			name:        "no_source",
			destination: ts.URL + "/etcha/v1/push/nowhere",
			privateKey:  prv2,
			path:        "testdata/good1.jsonnet",
			wantErr:     ErrPushSourceMismatch,
		},
		{
			name:        "denied_source",
			destination: ts.URL + "/etcha/v1/push/denied",
			privateKey:  prv2,
			path:        "testdata/good1.jsonnet",
			wantErr:     ErrPushSourceMismatch,
		},
		{
			name:        "bad_private_key",
			destination: ts.URL + "/etcha/v1/push/etcha",
			privateKey:  prv2,
			path:        "testdata/good1.jsonnet",
			wantErr:     ErrPushSourceMismatch,
		},
		{
			name:        "error_exec",
			destination: ts.URL + "/etcha/v1/push/etcha",
			mockErrors: []error{
				ErrNoPublicKeys,
				ErrNoPublicKeys,
			},
			privateKey: prv1,
			path:       "testdata/good1.jsonnet",
			wantInputs: []cli.RunMockInput{
				{Exec: " check1"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT="}, Exec: " change1"},
			},
			wantResult: &PushResult{
				Err: "error changing id 1: error running commands: error running commands: no public keys specified: ",
			},
			wantErr: errs.ErrReceiver,
		},
		{
			name:        "good",
			destination: ts.URL + "/etcha/v1/push/etcha",
			mockErrors: []error{
				ErrNoPublicKeys,
			},
			privateKey: prv1,
			path:       "testdata/good1.jsonnet",
			wantInputs: []cli.RunMockInput{
				{Exec: " check1"},
				{Environment: []string{"_CHECK=1", "_CHECK_OUT="}, Exec: " change1"},
			},
			wantResult: &PushResult{
				Changed: []string{"1"},
			},
		},
		{
			name:        "good-check",
			destination: ts.URL + "/etcha/v1/push/etcha?check",
			mockErrors: []error{
				ErrNoPublicKeys,
			},
			privateKey: prv1,
			path:       "testdata/good2.jsonnet",
			wantInputs: []cli.RunMockInput{
				{Exec: " check2"},
			},
			wantResult: &PushResult{
				Changed: []string{"2"},
			},
		},
		{
			name:        "good-2",
			destination: ts.URL + "/etcha/v1/push/etcha",
			privateKey:  prv1,
			path:        "testdata/good2.jsonnet",
			wantInputs: []cli.RunMockInput{
				{Exec: " check2"},
				{Environment: []string{"_CHECK=0", "_CHECK_OUT="}, Exec: " remove1"},
			},
			wantResult: &PushResult{
				Removed: []string{"1"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c.JWT.PrivateKey = tc.privateKey
			c.CLI.RunMockErrors(tc.mockErrors)

			r, err := Push(ctx, c, tc.destination, tc.path)
			assert.HasErr(t, err, tc.wantErr)
			assert.Equal(t, r, tc.wantResult)
			assert.Equal(t, c.CLI.RunMockInputs(), tc.wantInputs)
		})
	}

	os.RemoveAll("testdata")
}
