package handlers

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var ctx = context.Background()

var h = Handlers{
	"50_b": &Handler{
		Events: []string{
			"c",
			"b",
		},
		Exec: commands.Exec{
			Command: "b",
		},
		EventExit: true,
		WebhookPaths: []string{
			"/c",
			"/c/b",
		},
	},
	"30_a": &Handler{
		Events: []string{
			"c",
			"a",
		},
		Exec: commands.Exec{
			Command: "a",
		},
		WebhookPaths: []string{
			"/c",
			"/c/a",
		},
	},
	"40_c": &Handler{
		Events: []string{
			"c",
		},
		Exec: commands.Exec{
			Command: "c",
		},
		WebhookPaths: []string{
			"/c",
			"/c/c",
		},
	},
	"40_b": &Handler{
		Events: []string{
			"d",
		},
		Exec: commands.Exec{
			Command: "d",
		},
		WebhookPaths: []string{
			"/d",
		},
	},
}

func TestHandlersRegisterWebhooks(t *testing.T) {
	logger.UseTestLogger(t)

	c := cli.Config{}
	c.RunMock()

	env := []string{
		"ETCHA_HANDLER_TRIGGER=webhook",
		"ETCHA_WEBHOOK_BODY=aGVsbG8=",
		`ETCHA_WEBHOOK_HEADERS=Accept-Encoding: gzip
Content-Length: 5
Content-Type: application/xml
User-Agent: Go-http-client/1.1`,
		"ETCHA_WEBHOOK_QUERY=a=d&b=c",
	}

	tests := map[string]struct {
		handlers   Handlers
		mockErrors []error
		wantBody   string
		wantErr    error
		wantInputs []cli.RunMockInput
	}{
		"invalid_path": {
			handlers: Handlers{
				"a": &Handler{
					WebhookPaths: []string{
						"/test",
						"/etcha",
					},
				},
			},
			wantErr: errs.ErrReceiver,
		},
		"handlers": {
			handlers: h,
			mockErrors: []error{
				nil,
				errs.ErrSenderConflict,
				nil,
			},
			wantBody: `{"30_a":"hello","50_b":"world"}`,
			wantInputs: []cli.RunMockInput{
				{
					Environment: append([]string{
						"ETCHA_HANDLER_NAME=30_a",
					}, env...),
					Exec: "a",
				},
				{
					Environment: append([]string{
						"ETCHA_HANDLER_NAME=40_c",
					}, env...),
					Exec: "c",
				},
				{
					Environment: append([]string{
						"ETCHA_HANDLER_NAME=50_b",
					}, env...),
					Exec: "b",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.RunMockErrors(tc.mockErrors)
			c.RunMockOutputs([]string{"hello", "", "world"})

			h, err := tc.handlers.RegisterWebhooks(ctx, c)
			if tc.wantErr != nil {
				assert.HasErr(t, err, tc.wantErr)

				return
			}

			ts := httptest.NewServer(h)

			client := ts.Client()
			res, e := client.Post(ts.URL+"/c?b=c&a=d", "application/xml", bytes.NewBufferString("hello"))
			assert.HasErr(t, e, nil)

			defer res.Body.Close()

			b, e := io.ReadAll(res.Body)
			assert.HasErr(t, e, nil)

			assert.Equal(t, string(b), tc.wantBody)
			assert.Equal(t, c.RunMockInputs(), tc.wantInputs)
		})
	}
}

func TestHandlersRunEvents(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := cli.Config{}
	c.RunMock()

	h := Handlers{
		"1": {
			Exec: commands.Exec{
				Command: "1",
			},
			EventExit: true,
			Events: []string{
				"1",
				"3",
			},
		},
		"2": {
			Exec: commands.Exec{
				Command: "2",
			},
			Events: []string{
				"2",
				"3",
			},
		},
		"3": {
			Exec: commands.Exec{
				Command: "3",
			},
			Events: []string{
				"3",
			},
		},
	}

	assert.Equal(t, h.RunEvents(ctx, c, commands.Outputs{
		{
			ID: "a",
		},
		{
			Change: "changeA",
			ID:     "b",
			Events: []string{
				"1",
			},
		},
		{
			Change: "changeB",
			ID:     "c",
			Events: []string{
				"2",
			},
		},
		{
			Change: "changeC",
			ID:     "d",
			Events: []string{
				"3",
			},
		},
	}), true)
	assert.Equal(t, c.RunMockInputs(), []cli.RunMockInput{
		{
			Environment: []string{
				"ETCHA_EVENT_ID=b",
				"ETCHA_EVENT_NAME=1",
				"ETCHA_EVENT_OUTPUT=changeA",
				"ETCHA_HANDLER_NAME=1",
				"ETCHA_HANDLER_TRIGGER=event",
			},
			Exec: "1",
		},
		{
			Environment: []string{
				"ETCHA_EVENT_ID=c",
				"ETCHA_EVENT_NAME=2",
				"ETCHA_EVENT_OUTPUT=changeB",
				"ETCHA_HANDLER_NAME=2",
				"ETCHA_HANDLER_TRIGGER=event",
			},
			Exec: "2",
		},
		{
			Environment: []string{
				"ETCHA_EVENT_ID=d",
				"ETCHA_EVENT_NAME=3",
				"ETCHA_EVENT_OUTPUT=changeC",
				"ETCHA_HANDLER_NAME=1",
				"ETCHA_HANDLER_TRIGGER=event",
			},
			Exec: "1",
		},
		{
			Environment: []string{
				"ETCHA_EVENT_ID=d",
				"ETCHA_EVENT_NAME=3",
				"ETCHA_EVENT_OUTPUT=changeC",
				"ETCHA_HANDLER_NAME=2",
				"ETCHA_HANDLER_TRIGGER=event",
			},
			Exec: "2",
		},
		{
			Environment: []string{
				"ETCHA_EVENT_ID=d",
				"ETCHA_EVENT_NAME=3",
				"ETCHA_EVENT_OUTPUT=changeC",
				"ETCHA_HANDLER_NAME=3",
				"ETCHA_HANDLER_TRIGGER=event",
			},
			Exec: "3",
		},
	})
}
