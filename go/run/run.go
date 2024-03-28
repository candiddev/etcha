// Package run contains functions for pushes and pulling data for Etcha.
package run

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/certificates"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

var ErrNoVerifyKeys = errors.New("error running commands: no verify keys specified")
var ErrNilJWT = errors.New("received an empty JWT, this is probably a bug")

// Run starts the Etcha listener.
func Run(ctx context.Context, c *config.Config, once bool) errs.Err {
	s, err := newState(ctx, c)
	if err != nil {
		return logger.Error(ctx, err)
	}

	pubkey := len(c.Run.VerifyKeys) > 0
	if !pubkey {
		for i := range c.Sources {
			if len(c.Sources[i].VerifyKeys) > 0 {
				pubkey = true

				break
			}
		}
	}

	if !pubkey {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(ErrNoVerifyKeys))
	}

	s.loadExecJWTs(ctx)

	if once {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer s.WaitGroup.Wait()

	if len(c.Sources) != 0 {
		logger.Info(ctx, "Starting source runners...")

		s.sourceRunner(ctx, cancel)
	}

	if !once && s.Config.Run.ListenAddress != "" {
		if err := s.listen(ctx); err != nil {
			return logger.Error(ctx, err)
		}
	}

	return nil
}

type diffExecOpts struct {
	check          bool
	init           bool
	parentIDFilter *regexp.Regexp
}

// Result is a list of changed and removed IDs.
type Result struct {
	ChangedIDs     []string `json:"changedIDs"`
	ChangedOutputs []string `json:"changedOutputs"`
	Err            string   `json:"err"`
	Exit           bool     `json:"exit"`
	RemovedIDs     []string `json:"removedIDs"`
	RemovedOutputs []string `json:"removedOutputs"`
}

func (s *state) diffExec(ctx context.Context, source string, j *pattern.JWT, opts diffExecOpts) (*Result, errs.Err) {
	ctx = metrics.SetSourceName(ctx, source)

	if !opts.init && j == nil {
		metrics.CollectSources(ctx, true)

		return &Result{
			Err: ErrNilJWT.Error(),
		}, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrNilJWT))
	}

	var err errs.Err

	var pOld *pattern.Pattern

	p := s.Patterns.Get(source)

	if j != nil && j.EtchaPattern != nil {
		pOld = p

		p, err = j.Pattern(ctx, s.Config, source)
		if err != nil {
			metrics.CollectSources(ctx, true)

			return &Result{
				Err: err.Error(),
			}, logger.Error(ctx, err)
		}
	}

	if (p == nil || len(p.Run) == 0) && pOld == nil {
		return nil, logger.Error(ctx, nil)
	}

	src := s.Config.Sources[source]

	var o commands.Outputs

	var r Result

	if !src.TriggerOnly {
		logger.Info(ctx, fmt.Sprintf("Diffing Pattern for %s...", source))

		l := s.PatternLocks[source]
		if l == nil {
			s.PatternLocks[source] = &sync.Mutex{}
			l = s.PatternLocks[source]
		}

		if !src.RunMulti {
			l.Lock()
			defer l.Unlock()
		}

		o, err = p.DiffRun(ctx, s.Config, pOld, pattern.DiffRunOpts{
			Check:          src.CheckOnly || opts.check,
			NoRemove:       src.NoRemove,
			ParentIDFilter: opts.parentIDFilter,
			Source:         source,
		})
		if err != nil {
			metrics.CollectSources(ctx, true)

			return &Result{
				Err:  err.Error(),
				Exit: s.handleEvents(ctx, o, src),
			}, logger.Error(ctx, err)
		}

		cID, cOut := o.Changed()
		rID, rOut := o.Removed()

		r = Result{
			ChangedIDs:     cID,
			ChangedOutputs: cOut,
			RemovedIDs:     rID,
			RemovedOutputs: rOut,
		}

		metrics.CollectSources(ctx, false)
		metrics.CollectSourcesCommands(metrics.SetCommandMode(ctx, metrics.CommandModeChange), len(r.ChangedIDs))
		metrics.CollectSourcesCommands(metrics.SetCommandMode(ctx, metrics.CommandModeRemove), len(r.RemovedIDs))

		if opts.check {
			r.ChangedIDs = o.CheckFail(false)
			r.RemovedIDs = o.CheckFail(true)

			return &r, logger.Error(ctx, nil)
		}
	}

	if !src.NoRestore && j != nil && j.Raw != "" {
		logger.Info(ctx, fmt.Sprintf("Saving JWT for %s...", source))

		jp := filepath.Join(s.Config.Run.StateDir, source+".jwt")

		if err := os.WriteFile(jp+".tmp", []byte(j.Raw), 0644); err != nil { //nolint:gosec
			metrics.CollectSources(ctx, true)

			return &Result{
				Err:  err.Error(),
				Exit: s.handleEvents(ctx, o, src),
			}, logger.Error(ctx, errs.ErrReceiver.Wrap(err))
		}

		if err := os.Rename(jp+".tmp", jp); err != nil {
			metrics.CollectSources(ctx, true)

			return &Result{
				Err:  err.Error(),
				Exit: s.handleEvents(ctx, o, src),
			}, logger.Error(ctx, errs.ErrReceiver.Wrap(err))
		}

		s.JWTs.Set(source, j)
	}

	s.Patterns.Set(source, p)
	metrics.CollectSources(ctx, false)

	r.Exit = s.handleEvents(ctx, o, src)

	return &r, logger.Error(ctx, nil)
}

func (s *state) listen(ctx context.Context) errs.Err {
	m, e := s.newMux(ctx)
	if e != nil {
		return logger.Error(ctx, e)
	}

	srv := http.Server{
		Addr:              s.Config.Run.ListenAddress,
		Handler:           m,
		ReadHeaderTimeout: 60 * time.Second,
	}

	s.WaitGroup.Add(1)

	go func(ctx context.Context, srv *http.Server) {
		<-ctx.Done()
		srv.Shutdown(ctx) //nolint:errcheck
		s.WaitGroup.Done()
	}(ctx, &srv)

	var c tls.Certificate

	var err error

	switch {
	case s.Config.Run.TLSCertificateBase64 != "" && s.Config.Run.TLSKeyBase64 != "":
		c, err = certificates.GetBase64(s.Config.Run.TLSCertificateBase64, s.Config.Run.TLSKeyBase64)
	case s.Config.Run.TLSCertificatePath == "" || s.Config.Run.TLSCertificateBase64 == "":
		logger.Info(ctx, "Generating self-signed certificate for listener...")

		c, err = certificates.GetSelfSigned("Etcha")
	}

	if err != nil {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	if len(c.Certificate) > 0 {
		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{
				c,
			},
			MinVersion: tls.VersionTLS12,
		}
	}

	logger.Info(ctx, "Starting listener...")

	if err := srv.ListenAndServeTLS(s.Config.Run.TLSCertificatePath, s.Config.Run.TLSKeyPath); err != nil && err != http.ErrServerClosed {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	return logger.Error(ctx, nil)
}

func (s *state) loadExecJWTs(ctx context.Context) {
	ss := []string{}
	for source := range s.Config.Sources {
		ss = append(ss, source)
	}

	sort.Strings(ss)

	for _, k := range ss {
		ctx = metrics.SetSourceName(ctx, k)
		source := s.Config.Sources[k]

		// Load static commands from config.
		if len(source.Commands) > 0 {
			logger.Info(ctx, fmt.Sprintf("Loading config commands for %s...", k))

			p := &pattern.Pattern{
				Run:     source.Commands,
				RunExec: s.Config.Exec.Override(source.Exec),
			}
			s.Patterns.Set(k, p)
		}

		// Load cached JWT from state.
		if !source.NoRestore {
			path := filepath.Join(s.Config.Run.StateDir, k+".jwt")

			if jw, _, err := pattern.ParseJWTFromPath(ctx, s.Config, k, path); err == nil && jw != nil {
				p, err := jw.Pattern(ctx, s.Config, k)
				if err == nil {
					logger.Info(ctx, fmt.Sprintf("Loading cached config for %s...", k))
					s.JWTs.Set(k, jw)
					s.Patterns.Set(k, p)
				} else {
					logger.Error(ctx, err) //nolint:errcheck
				}
			}
		}

		// Run source
		if _, err := s.runSource(ctx, k, true); err != nil {
			logger.Error(ctx, err) //nolint: errcheck
		}
	}
}

func (s *state) runSource(ctx context.Context, source string, init bool) (*Result, errs.Err) {
	var err errs.Err

	j := s.JWTs.Get(source)

	if len(s.Config.Sources[source].PullPaths) > 0 {
		if n := pattern.ParseJWTFromSource(ctx, source, s.Config); j != nil && (j.Equal(n, s.Config.Sources[source].PullIgnoreVersion) != nil) {
			j = n
		}
	}

	if init {
		ctx = metrics.SetSourceTrigger(ctx, metrics.SourceTriggerInit)
	} else {
		ctx = metrics.SetSourceTrigger(ctx, metrics.SourceTriggerPull)
	}

	r, err := s.diffExec(ctx, source, j, diffExecOpts{
		init: init,
	})
	if err != nil {
		return r, logger.Error(ctx, err)
	}

	return r, nil
}

func (s *state) sourceRunner(ctx context.Context, cancel context.CancelFunc) {
	d := 0

	if s.Config.Run.RandomizedStartDelaySec > 0 {
		if r, err := rand.Int(rand.Reader, big.NewInt(int64(s.Config.Run.RandomizedStartDelaySec))); err == nil {
			d = int(r.Int64())
		}
	}

	time.Sleep(time.Duration(d) * time.Second)

	for k, v := range s.Config.Sources {
		if v.RunFrequencySec > 0 {
			s.WaitGroup.Add(1)

			go func(ctx context.Context, id string, frequency int) {
				defer s.WaitGroup.Done()

				t := time.NewTicker(time.Duration(frequency) * time.Second)

				for {
					select {
					case <-ctx.Done():
						return
					case <-t.C:
						r, err := s.runSource(ctx, id, false)
						if err != nil {
							logger.Error(ctx, err) //nolint: errcheck
						}

						if r.Exit {
							cancel()
						}
					}
				}
			}(ctx, k, v.RunFrequencySec)
		}
	}
}
