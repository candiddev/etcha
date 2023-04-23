// Package handlers contains functions for handling events and webhooks.
package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

// Handler is an event handler.
type Handler struct {
	Exec         commands.Exec `json:"exec"`
	EventExit    bool          `json:"eventExit"`
	Events       []string      `json:"events"`
	WebhookPaths []string      `json:"webhookPaths"`
}

// Handlers are multiple Handler in a map.
type Handlers map[string]*Handler

// RegisterWebhooks sorts and routes paths to defined webhook handlers.
func (h Handlers) RegisterWebhooks(ctx context.Context, c cli.Config) (http.Handler, errs.Err) { //nolint:gocognit
	routes := map[string][]string{}

	for name, handler := range h {
		for _, path := range handler.WebhookPaths {
			if strings.HasPrefix(path, "/etcha") {
				return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("handler %s has webhook path %s.  handlers cannot have a webhook for path for /etcha", name, path)))
			}

			routes[path] = append(routes[path], name)
		}
	}

	for _, path := range routes {
		sort.Strings(path)
	}

	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := map[string]cli.CmdOutput{}

		for route, handlers := range routes {
			if route == r.URL.Path {
				body, e := io.ReadAll(r.Body)
				if e != nil {
					logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error reading body"), e)) //nolint:errcheck

					continue
				}

				for i := range handlers {
					logger.Info(ctx, fmt.Sprintf("Running handler %s for webhook %s...", handlers[i], route))

					headers := []string{}

					for k, v := range r.Header {
						for i := range v {
							headers = append(headers, k+": "+v[i])
						}
					}

					sort.Strings(headers)

					exec := h[handlers[i]].Exec
					exec.Environment = append(exec.Environment, []string{
						"ETCHA_HANDLER_NAME=" + handlers[i],
						"ETCHA_HANDLER_TRIGGER=webhook",
						"ETCHA_WEBHOOK_BODY=" + base64.StdEncoding.EncodeToString(body),
						"ETCHA_WEBHOOK_HEADERS=" + strings.Join(headers, "\n"),
						"ETCHA_WEBHOOK_QUERY=" + r.URL.Query().Encode(),
					}...)

					out, err := exec.Run(ctx, c, "", "")
					if err != nil {
						logger.Error(ctx, err) //nolint:errcheck

						continue
					}

					res[handlers[i]] = out
				}
			}
		}

		if len(res) > 0 {
			b, err := json.Marshal(res)
			if err != nil {
				logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error rendring JSON"), err)) //nolint:errcheck
				w.WriteHeader(errs.ErrReceiver.Status())

				return
			}

			w.Header().Add("contenty-type", "application/json")
			w.Write(b) //nolint:errcheck
		} else {
			w.WriteHeader(errs.ErrSenderNotFound.Status())
			w.Write([]byte("404 page not found")) //nolint:errcheck
		}
	})

	return f, logger.Error(ctx, nil)
}

// RunEvents sorts and execs event handlers.
func (h Handlers) RunEvents(ctx context.Context, c cli.Config, o commands.Outputs) bool {
	events := o.Events()
	exit := false

	for _, sourceEvent := range events {
		handlers := []string{}

		for name, handler := range h {
			for _, event := range handler.Events {
				if sourceEvent.Name == event {
					handlers = append(handlers, name)
				}
			}
		}

		sort.Strings(handlers)

		for _, output := range sourceEvent.Outputs {
			for i := range handlers {
				logger.Info(ctx, fmt.Sprintf("Running handler %s for event %s from ID %s...", handlers[i], sourceEvent.Name, output.ID))

				exec := h[handlers[i]].Exec
				exec.Environment = append(exec.Environment, []string{
					"ETCHA_EVENT_ID=" + output.ID,
					"ETCHA_EVENT_NAME=" + sourceEvent.Name,
					"ETCHA_EVENT_OUTPUT=" + output.Change.String(),
					"ETCHA_HANDLER_NAME=" + handlers[i],
					"ETCHA_HANDLER_TRIGGER=event",
				}...)

				if h[handlers[i]].EventExit && !exit {
					exit = h[handlers[i]].EventExit
				}

				if _, err := exec.Run(ctx, c, "", ""); err != nil {
					logger.Error(ctx, err) //nolint:errcheck
				}
			}
		}
	}

	return exit
}
