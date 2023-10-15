package run

import (
	"context"
	"regexp"
	"testing"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

func TestStateHandleEvents(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()
	c.CLI.RunMock()
	c.Exec.AllowOverride = true
	c.Sources = map[string]*config.Source{
		"1": {
			Exec: &commands.Exec{
				AllowOverride: true,
			},
			EventsReceive: []string{
				"a",
				"b",
				"c",
			},
			EventsSend: *regexp.MustCompile(`a|b`),
		},
		"2": {
			Exec: &commands.Exec{
				AllowOverride: true,
			},
			EventsSend: *regexp.MustCompile("[^b]"),
		},
	}

	s, _ := newState(ctx, c)

	s.Patterns.Set("1", &pattern.Pattern{
		RunExec: commands.Exec{
			Command: "exec1",
		},
		Run: commands.Commands{
			{
				Always: true,
				Change: "change1",
				ID:     "1",
			},
		},
	})
	s.Patterns.Set("2", &pattern.Pattern{
		RunExec: commands.Exec{
			Command: "exec2",
		},
		Run: commands.Commands{
			{
				Always: true,
				Change: "change2",
				ID:     "2",
			},
		},
	})

	s.handleEvents(ctx, commands.Outputs{
		{
			Change: "change1",
			Events: []string{
				"a",
				"b",
				"c",
			},
			ID: "1",
		},
	}, c.Sources["1"])

	assert.Equal(t, c.CLI.RunMockInputs(), []cli.RunMockInput{
		{
			Environment: []string{
				"ETCHA_EVENT_ID=1",
				"ETCHA_EVENT_NAME=a",
				"ETCHA_EVENT_OUTPUT=change1",
				"ETCHA_SOURCE_NAME=1",
				"ETCHA_SOURCE_TRIGGER=event",
				"_CHECK=1",
			},
			Exec: "exec1 change1",
		},
		{
			Environment: []string{
				"ETCHA_EVENT_ID=1",
				"ETCHA_EVENT_NAME=b",
				"ETCHA_EVENT_OUTPUT=change1",
				"ETCHA_SOURCE_NAME=1",
				"ETCHA_SOURCE_TRIGGER=event",
				"_CHECK=1",
			},
			Exec: "exec1 change1",
		},
	})

	s.handleEvents(ctx, commands.Outputs{
		{
			Change: "change1",
			Events: []string{
				"a",
				"b",
				"c",
			},
			ID: "1",
		},
	}, c.Sources["2"])

	assert.Equal(t, c.CLI.RunMockInputs(), []cli.RunMockInput{
		{
			Environment: []string{
				"ETCHA_EVENT_ID=1",
				"ETCHA_EVENT_NAME=a",
				"ETCHA_EVENT_OUTPUT=change1",
				"ETCHA_SOURCE_NAME=1",
				"ETCHA_SOURCE_TRIGGER=event",
				"_CHECK=1",
			},
			Exec: "exec1 change1",
		},
		{
			Environment: []string{
				"ETCHA_EVENT_ID=1",
				"ETCHA_EVENT_NAME=c",
				"ETCHA_EVENT_OUTPUT=change1",
				"ETCHA_SOURCE_NAME=1",
				"ETCHA_SOURCE_TRIGGER=event",
				"_CHECK=1",
			},
			Exec: "exec1 change1",
		},
	})
}

func TestStateInitHandlers(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	c := config.Default()

	s, _ := newState(ctx, c)

	c.Sources = map[string]*config.Source{
		"etcha": {
			WebhookPaths: []string{
				"/etcha/oops",
			},
		},
	}

	assert.HasErr(t, s.initHandlers(ctx), errs.ErrReceiver)

	c.Sources = map[string]*config.Source{
		"1": {
			WebhookPaths: []string{
				"/1",
				"/2",
			},
		},
		"2": {
			WebhookPaths: []string{
				"/2",
			},
		},
	}

	s.HandlersEvents = map[string][]string{}
	s.HandlersRoutes = map[string]string{}

	assert.HasErr(t, s.initHandlers(ctx), errs.ErrReceiver)

	c.Sources = map[string]*config.Source{
		"1": {
			EventsReceive: []string{
				"a",
				"b",
			},
			WebhookPaths: []string{
				"/1",
			},
		},
		"2": {
			EventsReceive: []string{
				"a",
				"b",
			},
			WebhookPaths: []string{
				"/2",
			},
		},
	}

	s.HandlersEvents = map[string][]string{}
	s.HandlersRoutes = map[string]string{}

	assert.HasErr(t, s.initHandlers(ctx), nil)
	assert.Equal(t, s.HandlersEvents, map[string][]string{
		"a": {
			"1",
			"2",
		},
		"b": {
			"1",
			"2",
		},
	})
	assert.Equal(t, s.HandlersRoutes, map[string]string{
		"/1": "1",
		"/2": "2",
	})
}
