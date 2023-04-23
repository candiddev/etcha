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
	"github.com/candiddev/etcha/go/handlers"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

func TestCheckAuth(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.Run.SystemMetricsSecret = "test"
	s := newState(c)
	ts := httptest.NewServer(s.newMux(ctx))

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
	s := newState(c)
	ts := httptest.NewServer(s.newMux(ctx))

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
	c.Run.SystemMetricsSecret = "test"
	c.Handlers = handlers.Handlers{
		"test": {
			Exec: commands.Exec{
				Command: "hello",
			},
			WebhookPaths: []string{
				"/test1",
				"/test2",
			},
		},
	}

	s := newState(c)
	ts := httptest.NewServer(s.newMux(ctx))

	c.CLI.RunMockOutputs([]string{"world"})

	res, _ := http.Get(fmt.Sprintf("%s/test3", ts.URL))
	assert.Equal(t, res.StatusCode, http.StatusNotFound)

	res, _ = http.Get(fmt.Sprintf("%s/test1", ts.URL))
	b, _ := io.ReadAll(res.Body)
	assert.Equal(t, string(b), `{"test":"world"}`)
}
