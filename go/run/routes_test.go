package run

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

func TestCheckAuth(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.Run.SystemMetricsSecret = "test"
	s, _ := newState(ctx, c)
	m, _ := s.newMux(ctx)
	ts := httptest.NewServer(m)

	tests := map[string]struct {
		key        string
		statusCode int
	}{
		"pass": {
			key:        "test",
			statusCode: http.StatusOK,
		},
		"fail": {
			key:        "test!",
			statusCode: http.StatusForbidden,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			res, _ := http.Get(fmt.Sprintf("%s/etcha/v1/system/metrics?key=%s", ts.URL, tc.key))

			assert.Equal(t, res.StatusCode, tc.statusCode)
		})
	}
}

func TestCheckRateLimiter(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.Run.SystemMetricsSecret = "test"
	s, _ := newState(ctx, c)
	m, _ := s.newMux(ctx)
	ts := httptest.NewServer(m)

	var res *http.Response

	for i := 0; i < 11; i++ {
		res, _ = http.Get(fmt.Sprintf("%s/etcha/v1/system/metrics?key=12345", ts.URL))
	}

	assert.Equal(t, res.StatusCode, errs.ErrSenderTooManyRequest.Status())
}

func TestHandlers(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.CLI.RunMock()
	c.Exec.AllowOverride = true
	c.Sources = map[string]*config.Source{
		"1": {
			WebhookPaths: []string{
				"/test2",
			},
		},
		"2": {
			WebhookPaths: []string{
				"/test3",
			},
		},
		"3": {
			Exec: &commands.Exec{
				AllowOverride: true,
			},
			WebhookPaths: []string{
				"/test1",
			},
		},
	}

	s, _ := newState(ctx, c)
	m, _ := s.newMux(ctx)
	ts := httptest.NewServer(m)

	s.Patterns.Set("2", &pattern.Pattern{
		Run: commands.Commands{},
	})
	s.Patterns.Set("3", &pattern.Pattern{
		RunExec: commands.Exec{
			Command: "hello",
		},
		Run: commands.Commands{
			{
				Always: true,
				Change: "a",
				ID:     "a",
				OnChange: []string{
					"etcha:webhook_body",
				},
			},
			{
				Always: true,
				Change: "b",
				ID:     "b",
				OnChange: []string{
					"etcha:webhook_content_type",
				},
			},
		},
	})

	c.CLI.RunMockOutputs([]string{`{"test":"world"}`, "application/json"})

	res, _ := http.Get(fmt.Sprintf("%s/test3", ts.URL))
	assert.Equal(t, res.StatusCode, http.StatusNotFound)

	res, _ = http.Get(fmt.Sprintf("%s/test1", ts.URL))
	b, _ := io.ReadAll(res.Body)
	assert.Equal(t, string(b), `{"test":"world"}`)
	assert.Equal(t, res.Header.Get("content-type"), "application/json")
	assert.Equal(t, c.CLI.RunMockInputs(), []cli.RunMockInput{
		{
			Environment: []string{
				"ETCHA_SOURCE_NAME=3",
				"ETCHA_SOURCE_TRIGGER=webhook",
				"ETCHA_WEBHOOK_BODY=",
				"ETCHA_WEBHOOK_HEADERS=Accept-Encoding: gzip\nUser-Agent: Go-http-client/1.1",
				"ETCHA_WEBHOOK_METHOD=GET",
				"ETCHA_WEBHOOK_PATH=/test1",
				"ETCHA_WEBHOOK_QUERY=",
				"_CHECK=1",
			},
			Exec: "hello a",
		},
		{
			Environment: []string{
				"ETCHA_SOURCE_NAME=3", "ETCHA_SOURCE_TRIGGER=webhook", "ETCHA_WEBHOOK_BODY=",
				"ETCHA_WEBHOOK_HEADERS=Accept-Encoding: gzip\nUser-Agent: Go-http-client/1.1",
				"ETCHA_WEBHOOK_METHOD=GET",
				"ETCHA_WEBHOOK_PATH=/test1",
				"ETCHA_WEBHOOK_QUERY=",
				"_CHANGE=0",
				`_CHANGE_OUT={"test":"world"}`,
				"_CHECK=1",
			},
			Exec: "hello b",
		},
	})
}
