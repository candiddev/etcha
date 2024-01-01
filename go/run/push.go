package run

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
	"github.com/go-chi/chi/v5"
)

var ErrPushDecode = errors.New("error decoding response")
var ErrPushRateLimit = errors.New("push destination is rate limiting")
var ErrPushRead = errors.New("error reading response")
var ErrPushRequest = errors.New("error performing request")
var ErrPushSourceMismatch = errors.New("push didn't match any sources")

// Result is a list of changed and removed IDs.
type Result struct {
	ChangedIDs     []string `json:"changedIDs"`
	ChangedOutputs []string `json:"changedOutputs"`
	Err            string   `json:"err"`
	Exit           bool     `json:"exit"`
	RemovedIDs     []string `json:"removedIDs"`
	RemovedOutputs []string `json:"removedOutputs"`
}

// Push sends content to the dest.
func Push(ctx context.Context, c *config.Config, dest, cmd, path string) (r *Result, err errs.Err) {
	logger.Debug(ctx, fmt.Sprintf("Pushing config to %s...", dest))

	r = &Result{}

	var p *pattern.Pattern

	if path == "" {
		p = &pattern.Pattern{
			Imports: &jsonnet.Imports{
				Entrypoint: "/main.jsonnet",
				Files: map[string]string{
					"/main.jsonnet": fmt.Sprintf(`{run:[{always: true, change: "%s", id: "etcha push"}]}`, cmd),
				},
			},
		}
	} else {
		p, err = pattern.ParsePatternFromPath(ctx, c, "", path)
		if err != nil {
			return r, logger.Error(ctx, err)
		}
	}

	jwt, err := p.Sign(ctx, c, "", nil)
	if err != nil {
		return r, logger.Error(ctx, err)
	}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, dest, bytes.NewReader([]byte(jwt)))
	if e != nil {
		return r, logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error creating request"), e))
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.Build.PushTLSSkipVerify, //nolint:gosec
			},
		},
	}

	res, e := client.Do(req)
	if e != nil {
		return r, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPushRequest, e))
	}

	if res.StatusCode == http.StatusNotFound {
		return r, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPushSourceMismatch))
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		return r, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPushRateLimit))
	}

	if res.StatusCode == http.StatusNotModified {
		logger.Info(ctx, "No changes")

		return r, nil
	}

	b, e := io.ReadAll(res.Body)
	if e != nil {
		return r, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPushRead, e))
	}

	if e := json.Unmarshal(b, r); e != nil {
		return r, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPushDecode, e))
	}

	if r.Err != "" {
		return r, logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New(r.Err)))
	}

	return r, logger.Error(ctx, nil)
}

func (s *state) postPush(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	c := chi.URLParam(r, "config")
	ctx = metrics.SetSourceName(ctx, c)

	src, ok := s.Config.Sources[c]
	if !ok || !src.AllowPush {
		logger.Error(ctx, errs.ErrSenderBadRequest.Wrap(fmt.Errorf("unknown config: %s", c))) //nolint: errcheck
		w.WriteHeader(http.StatusNotFound)

		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error(ctx, errs.ErrSenderBadRequest.Wrap(errors.New("error reading body"), err)) //nolint: errcheck
		w.WriteHeader(http.StatusNotFound)

		return
	}

	push, err := pattern.ParseJWT(ctx, s.Config, string(body), c)
	if err != nil {
		logger.Error(ctx, errs.ErrSenderBadRequest.Wrap(errors.New("error parsing JWT"), err)) //nolint: errcheck
		w.WriteHeader(http.StatusNotFound)

		return
	}

	if old := s.JWTs.Get(c); old == nil || old.Equal(push, src.PullIgnoreVersion) != nil || src.RunAll {
		ctx = metrics.SetSourceTrigger(ctx, metrics.SourceTriggerPush)

		r, err := s.diffExec(ctx, r.URL.Query().Has("check"), c, push, false)
		if err != nil {
			w.WriteHeader(errs.ErrReceiver.Status())
			logger.Error(ctx, err) //nolint:errcheck
		}

		j, e := json.MarshalIndent(r, "", "  ")
		if e != nil {
			logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error rending JSON"), e)) //nolint:errcheck

			return
		}

		w.Header().Set("Content-Type", "application/json")

		if _, e := w.Write(j); e != nil {
			logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error writing JSON"), e)) //nolint:errcheck

			return
		}

		return
	}

	w.WriteHeader(http.StatusNotModified)
}
