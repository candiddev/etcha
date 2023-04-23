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
	"sync"
	"time"

	"github.com/candiddev/etcha/go/commands"
	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/certificates"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/ulule/limiter/v3"
)

type state struct {
	JWTs          map[string]*pattern.JWT
	JWTsMutex     sync.Mutex
	Config        *config.Config
	Patterns      map[string]*pattern.Pattern
	PatternsMutex sync.Mutex
	RateLimiter   *limiter.Limiter
}

func newState(c *config.Config) *state {
	return &state{
		Config:   c,
		JWTs:     map[string]*pattern.JWT{},
		Patterns: map[string]*pattern.Pattern{},
	}
}

var ErrNoPublicKeys = errors.New("error running commands: no public keys specified")
var ErrNilJWT = errors.New("received an empty JWT, this is probably a bug")

// Run starts the Etcha listener.
func Run(ctx context.Context, c *config.Config, once bool) errs.Err {
	s := newState(c)
	s.loadExecJWTs(ctx)

	pubkey := len(c.JWT.PublicKeys) > 0
	if !pubkey {
		for i := range c.Sources {
			if len(c.Sources[i].JWTPublicKeys) > 0 {
				pubkey = true

				break
			}
		}
	}

	if !pubkey {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(ErrNoPublicKeys))
	}

	if len(c.Sources) != 0 {
		if once {
			for source := range c.Sources {
				if err := s.runSource(ctx, source); err != nil {
					logger.Error(ctx, err) //nolint: errcheck
				}
			}
		} else {
			logger.Info(ctx, "Starting source runner...")
			go s.sourceRunner(ctx)
		}
	}

	if !once && s.Config.Run.ListenAddress != "" {
		if err := s.listen(ctx); err != nil {
			return logger.Error(ctx, err)
		}
	}

	return nil
}

func (s *state) diffExec(ctx context.Context, check bool, source string, j *pattern.JWT) (*PushResult, errs.Err) {
	if j == nil {
		return &PushResult{
			Err: ErrNilJWT.Error(),
		}, logger.Error(ctx, errs.ErrReceiver.Wrap(ErrNilJWT))
	}

	p, err := j.Pattern(ctx, s.Config, source)
	if err != nil {
		return &PushResult{
			Err: err.Error(),
		}, logger.Error(ctx, err)
	}

	s.PatternsMutex.Lock()
	defer s.PatternsMutex.Unlock()
	pOld := s.Patterns[source]

	if s.Config.Sources[source].RunAlwaysCheck {
		pOld = nil
	}

	logger.Info(ctx, fmt.Sprintf("Updating config for %s...", source))

	o, err := p.DiffRun(ctx, s.Config, pOld, s.Config.Sources[source].CheckOnly || check)
	if err != nil {
		return &PushResult{
			Err:  err.Error(),
			Exit: s.Config.Handlers.RunEvents(ctx, s.Config.CLI, o),
		}, logger.Error(ctx, err)
	}

	r := PushResult{
		Changed: o.Changed(),
		Removed: o.Removed(),
	}

	if check {
		r.Changed = o.CheckFail()

		return &r, logger.Error(ctx, nil)
	}

	jp := filepath.Join(s.Config.Run.StateDir, source+".jwt")

	if err := os.WriteFile(jp+".tmp", []byte(j.Raw), 0644); err != nil { //nolint:gosec
		return &PushResult{
			Err:  err.Error(),
			Exit: s.Config.Handlers.RunEvents(ctx, s.Config.CLI, o),
		}, logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	if err := os.Rename(jp+".tmp", jp); err != nil {
		return &PushResult{
			Err:  err.Error(),
			Exit: s.Config.Handlers.RunEvents(ctx, s.Config.CLI, o),
		}, logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	s.Patterns[source] = p

	s.JWTsMutex.Lock()
	s.JWTs[source] = j
	s.JWTsMutex.Unlock()

	r.Exit = s.Config.Handlers.RunEvents(ctx, s.Config.CLI, o)

	return &r, logger.Error(ctx, nil)
}

func (s *state) listen(ctx context.Context) errs.Err {
	srv := http.Server{
		Addr:              s.Config.Run.ListenAddress,
		Handler:           s.newMux(ctx),
		ReadHeaderTimeout: 60 * time.Second,
	}

	go func(ctx context.Context, srv *http.Server) {
		<-ctx.Done()
		srv.Shutdown(ctx) //nolint:errcheck
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
	js := pattern.ParseJWTsFromDir(ctx, s.Config)

	s.JWTsMutex.Lock()
	s.PatternsMutex.Lock()
	for n, j := range js {
		p, err := j.Pattern(ctx, s.Config, n)
		if err == nil {
			logger.Info(ctx, fmt.Sprintf("Loading existing config for %s...", n))

			m := commands.ModeChange
			if s.Config.Sources[n].CheckOnly {
				m = commands.ModeCheck
			}

			if _, err := p.Run.Run(ctx, s.Config.CLI, p.RunEnv, p.Exec, m); err == nil {
				s.JWTs[n] = j
				s.Patterns[n] = p
			} else {
				logger.Error(ctx, err) //nolint:errcheck
			}
		} else {
			logger.Error(ctx, err) //nolint:errcheck
		}
	}

	s.JWTsMutex.Unlock()
	s.PatternsMutex.Unlock()
}

func (s *state) runSource(ctx context.Context, source string) errs.Err {
	s.PatternsMutex.Lock()
	oldJ := s.JWTs[source]
	newJ := s.JWTs[source]
	s.PatternsMutex.Unlock()

	diff := false

	if j := pattern.ParseJWTFromSources(ctx, source, s.Config); j != nil && (j.Equal(oldJ, s.Config.Sources[source].PullIgnoreVersion) != nil) {
		diff = true
		newJ = j
	}

	if diff || s.Config.Sources[source].RunAlwaysCheck {
		_, err := s.diffExec(ctx, false, source, newJ)
		if err != nil {
			return logger.Error(ctx, err)
		}
	}

	return nil
}

func (s *state) sourceRunner(ctx context.Context) {
	d := 0

	if s.Config.Run.RandomizedStartDelaySec > 0 {
		if r, err := rand.Int(rand.Reader, big.NewInt(int64(s.Config.Run.RandomizedStartDelaySec))); err == nil {
			d = int(r.Int64())
		}
	}

	time.Sleep(time.Duration(d) * time.Second)

	pull := make(chan string)

	for k, v := range s.Config.Sources {
		if v.RunFrequency > 0 {
			go func(ctx context.Context, id string, frequency int, pull chan string) {
				t := time.NewTicker(time.Duration(frequency) * time.Second)

				for {
					select {
					case <-ctx.Done():
						return
					case <-t.C:
						pull <- id
					}
				}
			}(ctx, k, v.RunFrequency, pull)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case source := <-pull:
			if err := s.runSource(ctx, source); err != nil {
				logger.Error(ctx, err) //nolint: errcheck
			}
		}
	}
}
