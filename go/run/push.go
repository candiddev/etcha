package run

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/semaphore"
)

var ErrPushDecode = errors.New("error decoding response")
var ErrPushRateLimit = errors.New("push destination is rate limiting")
var ErrPushRead = errors.New("error reading response")
var ErrPushRequest = errors.New("error performing request")
var ErrPushSourceMismatch = errors.New("push didn't match any sources")

// PushOpts are options for Push.
type PushOpts struct {
	Check          bool
	ParentIDFilter *regexp.Regexp
	TargetFilter   *regexp.Regexp
}

// PushTargets pushes a cmd to a bunch of targets.
func PushTargets(ctx context.Context, c *config.Config, targets map[string]config.PushTarget, source, cmd string, opts PushOpts) ([]string, errs.Err) { //nolint:gocognit,revive
	t := []string{}

	for k := range targets {
		if opts.TargetFilter == nil || opts.TargetFilter.String() == "" || opts.TargetFilter.MatchString(k) {
			for s := range targets[k].SourcePatterns {
				if s == source {
					t = append(t, k)

					break
				}
			}
		}
	}

	sort.Strings(t)

	l := sync.Mutex{}
	r := types.Results{}
	s := semaphore.NewWeighted(int64(c.Build.PushMaxWorkers))

	var terr errs.Err

	for i := range t {
		if err := s.Acquire(ctx, 1); err != nil {
			return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error acquiring semaphore: %w", err)))
		}

		go func(target string) {
			logger.Debug(ctx, fmt.Sprintf("Pushing config to %s...", target))

			rcmd := cmd

			var err errs.Err

			var p *pattern.Pattern

			var path bool

			var res *Result

			o := []string{}

			if targets[target].SourcePatterns[source] != "" {
				rcmd = targets[target].SourcePatterns[source]
			}

			p, path, err = getPushPattern(ctx, c, rcmd)
			if err == nil {
				var buildManifest string

				var runVars map[string]any

				buildManifest, runVars, err = p.BuildRun(ctx, c)
				if err == nil {
					var dest string

					var jwt string

					dest, jwt, err = getPushDestJWT(ctx, c, targets[target], p, buildManifest, source, runVars, opts)
					if err == nil {
						res, err = pushTarget(ctx, c, dest, jwt)
					}
				}
			}

			if res == nil {
				res = &Result{}
			}

			if err != nil {
				terr = err

				if res.Err == "" {
					res.Err = err.Error()
				}
			}

			if res.Err != "" {
				e := "ERROR: " + res.Err
				if !logger.GetNoColor(ctx) {
					e = fmt.Sprintf("%s%s%s", logger.ColorRed, e, logger.ColorReset)
				}

				o = append(o, e)
			}

			if path {
				if len(res.ChangedIDs) == 0 && len(res.RemovedIDs) == 0 {
					o = append(o, "No changes")
				} else {
					if len(res.ChangedIDs) > 0 {
						o = append(o, fmt.Sprintf("Changed %d: %s", len(res.ChangedIDs), strings.Join(res.ChangedIDs, ", ")))
					}

					if len(res.RemovedIDs) > 0 {
						o = append(o, fmt.Sprintf("Removed %d: %s", len(res.RemovedIDs), strings.Join(res.RemovedIDs, ", ")))
					}
				}
			} else {
				o = append(o, res.ChangedOutputs...)
			}

			l.Lock()
			r[target] = o
			l.Unlock()
			s.Release(1)
		}(t[i])
	}

	if err := s.Acquire(ctx, int64(c.Build.PushMaxWorkers)); err != nil {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error acquiring all semaphores: %w", err)))
	}

	return r.Show(), terr
}

func getPushPattern(ctx context.Context, c *config.Config, cmd string) (*pattern.Pattern, bool, errs.Err) {
	var err errs.Err

	var p *pattern.Pattern

	path := false

	if strings.HasSuffix(cmd, ".jsonnet") {
		path = true

		p, err = pattern.ParsePatternFromPath(ctx, c, "", cmd)
		if err != nil {
			return nil, path, logger.Error(ctx, err)
		}
	} else {
		p = &pattern.Pattern{
			Imports: &jsonnet.Imports{
				Entrypoint: "/main.jsonnet",
				Files: map[string]string{
					"/main.jsonnet": fmt.Sprintf(`{run:[{always: true, change: "%s", id: "etcha push"}]}`, cmd),
				},
			},
		}
	}

	return p, path, err
}

func getPushDestJWT(ctx context.Context, c *config.Config, target config.PushTarget, p *pattern.Pattern, buildManifest, source string, runVars map[string]any, opts PushOpts) (dest, jwt string, err errs.Err) { //nolint:revive
	d := url.URL{
		Host: net.JoinHostPort(target.Hostname, strconv.Itoa(target.Port)),
		Path: target.Path + "/" + source,
	}

	if target.Insecure {
		d.Scheme = "http"
	} else {
		d.Scheme = "https"
	}

	if opts.Check {
		d.Query().Add("check", "")
	}

	if opts.ParentIDFilter != nil && opts.ParentIDFilter.String() != "" {
		d.Query().Add("filter", opts.ParentIDFilter.String())
	}

	vars := maps.Clone(runVars)

	for k, v := range target.Vars {
		vars[k] = v
	}

	jwt, _, err = p.Sign(ctx, c, buildManifest, vars)

	return d.String(), jwt, logger.Error(ctx, err)
}

func pushTarget(ctx context.Context, c *config.Config, dest, jwt string) (r *Result, err errs.Err) {
	r = &Result{}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, dest, bytes.NewReader([]byte(jwt)))
	if e != nil {
		return r, errs.ErrReceiver.Wrap(errors.New("error creating request"))
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
		return r, errs.ErrReceiver.Wrap(ErrPushRequest, e)
	}

	if res.StatusCode == http.StatusNotFound {
		return r, errs.ErrReceiver.Wrap(ErrPushSourceMismatch)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		return r, errs.ErrReceiver.Wrap(ErrPushRateLimit)
	}

	if res.StatusCode == http.StatusNotModified {
		return r, nil
	}

	b, e := io.ReadAll(res.Body)
	if e != nil {
		return r, errs.ErrReceiver.Wrap(ErrPushRead, e)
	}

	if e := json.Unmarshal(b, r); e != nil {
		return r, errs.ErrReceiver.Wrap(ErrPushDecode, e)
	}

	if r.Err != "" {
		return r, errs.ErrReceiver.Wrap(errors.New(r.Err))
	}

	return r, logger.Error(ctx, nil)
}

func (s *state) postPush(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	c := chi.URLParam(r, "config")
	ctx = metrics.SetSourceName(ctx, c)
	ctx = logger.SetAttribute(ctx, "remoteAddr", r.RemoteAddr)

	logger.Info(ctx, "Push configuration request")

	src, ok := s.Config.Sources[c]
	if !ok || !src.AllowPush {
		logger.Error(ctx, errs.ErrSenderBadRequest.Wrap(fmt.Errorf("unknown config: %s", c))) //nolint: errcheck
		w.WriteHeader(http.StatusNotFound)

		return
	}

	body, e := io.ReadAll(r.Body)
	if e != nil {
		logger.Error(ctx, errs.ErrSenderBadRequest.Wrap(errors.New("error reading body"), e)) //nolint: errcheck
		w.WriteHeader(http.StatusNotFound)

		return
	}

	push, _, err := pattern.ParseJWT(ctx, s.Config, string(body), c)
	if err != nil {
		logger.Error(ctx, err) //nolint: errcheck
		w.WriteHeader(http.StatusNotFound)

		return
	}

	ctx = metrics.SetSourceTrigger(ctx, metrics.SourceTriggerPush)

	reg, e := regexp.Compile(r.URL.Query().Get("filter"))
	if e != nil {
		logger.Error(ctx, errs.ErrSenderBadRequest.Wrap(errors.New("error parsing filter"), e)) //nolint:errcheck
		w.WriteHeader(http.StatusBadRequest)
	}

	res, err := s.diffExec(ctx, c, push, diffExecOpts{
		check:          r.URL.Query().Has("check"),
		parentIDFilter: reg,
	})

	if err != nil {
		w.WriteHeader(errs.ErrReceiver.Status())
		logger.Error(ctx, err) //nolint:errcheck
	}

	j, e := json.MarshalIndent(res, "", "  ")
	if e != nil {
		logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error rending JSON"), e)) //nolint:errcheck

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if _, e := w.Write(j); e != nil {
		logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error writing JSON"), e)) //nolint:errcheck

		return
	}
}
