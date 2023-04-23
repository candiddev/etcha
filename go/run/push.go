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
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/go-chi/chi/v5"
)

var ErrPushDecode = errors.New("error decoding response")
var ErrPushRateLimit = errors.New("push destination is rate limiting")
var ErrPushRead = errors.New("error reading response")
var ErrPushRequest = errors.New("error performing request")
var ErrPushSourceMismatch = errors.New("push didn't match any sources")

// PushResult is a list of changed and removed IDs.
type PushResult struct {
	Changed []string `json:"changed"`
	Err     string   `json:"err"`
	Exit    bool     `json:"exit"`
	Removed []string `json:"removed"`
}

// Push sends content to the dest.
func Push(ctx context.Context, c *config.Config, dest, path string) (push *PushResult, err errs.Err) {
	logger.Info(ctx, fmt.Sprintf("Pushing config to %s...", dest))

	p, err := pattern.ParsePatternFromPath(ctx, c, "", path)
	if err != nil {
		return nil, logger.Error(ctx, err)
	}

	jwt, err := p.Sign(ctx, c, "", nil)
	if err != nil {
		return nil, logger.Error(ctx, err)
	}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, dest, bytes.NewReader([]byte(jwt)))
	if e != nil {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error creating request"), e))
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.Run.PushTLSSkipVerify, //nolint:gosec
			},
		},
	}

	res, e := client.Do(req)
	if e != nil {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPushRequest, e))
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPushSourceMismatch))
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPushRateLimit))
	}

	if res.StatusCode == http.StatusNotModified {
		logger.Info(ctx, "No changes")

		return nil, nil
	}

	push = &PushResult{}

	b, e := io.ReadAll(res.Body)
	if e != nil {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPushRead, e))
	}

	if e := json.Unmarshal(b, push); e != nil {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrPushDecode, e))
	}

	if push.Err != "" {
		return push, logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New(push.Err)))
	}

	return push, logger.Error(ctx, nil)
}

func (s *state) postPush(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	c := chi.URLParam(r, "config")

	if s, ok := s.Config.Sources[c]; !ok || !s.AllowPush {
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

	if old, ok := s.JWTs[c]; !ok || old.Equal(push, s.Config.Sources[c].PullIgnoreVersion) != nil {
		r, err := s.diffExec(ctx, r.URL.Query().Has("check"), c, push)
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
