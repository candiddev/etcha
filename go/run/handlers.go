package run

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
)

func (s *state) handleEvents(ctx context.Context, o commands.Outputs, source *config.Source) bool {
	ctx = metrics.SetSourceTrigger(ctx, metrics.SourceTriggerEvent)
	exit := false

	for _, event := range o.Events() {
		allow := false
		if source.EventsSend.String() != "" {
			allow = source.EventsSend.MatchString(event.Name)
		}

		if !allow {
			logger.Debug(ctx, fmt.Sprintf("Skipping firing event %s, not allowed for source", event.Name))

			continue
		}

		for _, source := range s.HandlersEvents[event.Name] {
			for i := range event.Outputs {
				ctx = metrics.SetSourceName(ctx, source)

				logger.Info(ctx, fmt.Sprintf("Running source %s for event %s from ID %s...", source, event.Name, event.Outputs[i].ID))

				src := s.Config.Sources[source]
				if src == nil {
					continue
				}

				if !exit && src.EventsReceiveExit {
					exit = true
				}

				p := s.Patterns.Get(source)
				if p == nil || len(p.Run) == 0 {
					logger.Info(ctx, fmt.Sprintf("No pattern with runnable commands for source %s to handle event %s", source, event.Name))

					continue
				}

				env := p.GetRunEnv()
				env["ETCHA_EVENT_ID"] = event.Outputs[i].ID
				env["ETCHA_EVENT_NAME"] = event.Name
				env["ETCHA_EVENT_OUTPUT"] = event.Outputs[i].Change.String()
				env["ETCHA_SOURCE_NAME"] = source
				env["ETCHA_SOURCE_TRIGGER"] = "event"

				_, err := p.Run.Run(ctx, s.Config.CLI, env, s.Config.Exec.Override(s.Config.Sources[source].Exec, p.RunExec), false, false)

				logger.Error(ctx, err) //nolint:errcheck
			}
		}
	}

	return exit
}

func (s *state) initHandlers(ctx context.Context) errs.Err { //nolint:gocognit
	sources := []string{}

	for name, source := range s.Config.Sources {
		if source != nil {
			sources = append(sources, name)
		}
	}

	sort.Strings(sources)

	for _, source := range sources {
		for _, event := range s.Config.Sources[source].EventsReceive {
			s.HandlersEvents[event] = append(s.HandlersEvents[event], source)
		}

		for _, path := range s.Config.Sources[source].WebhookPaths {
			switch {
			case strings.HasPrefix(path, "/etcha"):
				return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("source %s has webhook path %s.  sources cannot have a webhook for path for /etcha", source, path)))
			case s.HandlersRoutes[path] != "":
				return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("source %s tried to claim webhook path %s that was already claimed by %s", source, path, s.HandlersRoutes[path])))
			}

			s.HandlersRoutes[path] = source
		}
	}

	s.WebhookHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = metrics.SetSourceTrigger(ctx, metrics.SourceTriggerWebhook)

		resBody := ""
		resCT := ""

		body, e := io.ReadAll(r.Body)
		if e != nil {
			logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error reading body"), e)) //nolint:errcheck

			w.WriteHeader(errs.ErrSenderNotFound.Status())
			w.Write([]byte("404 page not found")) //nolint:errcheck

			return
		}

		for route, source := range s.HandlersRoutes {
			ctx = metrics.SetSourceName(ctx, source)

			if route == r.URL.Path {
				logger.Info(ctx, fmt.Sprintf("Running source %s for webhook %s...", source, route))

				p := s.Patterns.Get(source)
				if p == nil || len(p.Run) == 0 {
					logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("no pattern with runnable commands for source %s to handle webhook %s", source, route))) //nolint:errcheck

					break
				}

				headers := []string{}

				for k, v := range r.Header {
					for i := range v {
						headers = append(headers, k+": "+v[i])
					}
				}

				sort.Strings(headers)

				out, err := p.Run.Run(ctx, s.Config.CLI, types.EnvVars{
					"ETCHA_SOURCE_NAME":     source,
					"ETCHA_SOURCE_TRIGGER":  "webhook",
					"ETCHA_WEBHOOK_BODY":    base64.StdEncoding.EncodeToString(body),
					"ETCHA_WEBHOOK_HEADERS": strings.Join(headers, "\n"),
					"ETCHA_WEBHOOK_METHOD":  r.Method,
					"ETCHA_WEBHOOK_PATH":    r.URL.Path,
					"ETCHA_WEBHOOK_QUERY":   r.URL.Query().Encode(),
				}, s.Config.Exec.Override(s.Config.Sources[source].Exec, p.RunExec), false, false)
				if err == nil {
					for _, event := range out.Events() {
						if event.Name == "webhookBody" && len(event.Outputs) > 0 {
							resBody = string(event.Outputs[0].Change)

							continue
						}

						if event.Name == "webhookContentType" && len(event.Outputs) > 0 {
							resCT = string(event.Outputs[0].Change)

							continue
						}
					}

					break
				}
				logger.Error(ctx, err) //nolint:errcheck
			}
		}

		if resCT != "" {
			w.Header().Add("content-type", resCT)
		}

		if resBody != "" {
			w.Write([]byte(resBody)) //nolint:errcheck
		} else {
			w.WriteHeader(errs.ErrSenderNotFound.Status())
			w.Write([]byte("404 page not found")) //nolint:errcheck
		}
	})

	return logger.Error(ctx, nil)
}
