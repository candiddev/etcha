package run

import (
	"context"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"

	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func (s *state) newMux(ctx context.Context) (http.Handler, errs.Err) {
	r := chi.NewRouter()

	rate, err := limiter.NewRateFromFormatted(s.Config.Run.RateLimiterRate)
	if err != nil {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	store := memory.NewStore()

	s.RateLimiter = limiter.New(store, rate, limiter.WithTrustForwardHeader(true))

	r.Use(middleware.Recoverer)
	r.Use(s.setContext)
	r.Use(s.checkRateLimiter)
	r.Post("/etcha/v1/push/{config}", s.postPush)
	r.Route("/etcha/v1/system", func(r chi.Router) {
		r.Use(s.checkSystemAuth)

		if s.Config.Run.SystemMetricsSecret != "" {
			r.Handle("/metrics", promhttp.Handler())
		}

		if s.Config.Run.SystemPprofSecret != "" {
			r.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
			r.Handle("/pprof/heap", pprof.Handler("heap"))
		}
	})
	r.NotFound(s.WebhookHandler.ServeHTTP)

	return r, nil
}

func (s *state) checkRateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.Trace(r.Context())
		ip := s.RateLimiter.GetIPKey(r)
		ctx = logger.SetAttribute(ctx, "path", r.URL.Path)
		ctx = logger.SetAttribute(ctx, "sourceAddress", ip)
		key := ip + chi.RouteContext(ctx).RoutePattern()

		limit, err := s.RateLimiter.Get(ctx, key)
		if err != nil {
			logger.Error(ctx, errs.ErrReceiver.Wrap(err)) //nolint:errcheck
			w.WriteHeader(errs.ErrReceiver.Status())

			return
		}

		w.Header().Add("x-rate-limit-limit", strconv.Itoa(int(limit.Limit)))
		w.Header().Add("x-rate-limit-remaining", strconv.Itoa(int(limit.Remaining)))
		w.Header().Add("x-rate-limit-reset", strconv.Itoa(int(limit.Reset)))

		if limit.Reached {
			logger.Info(ctx, "Rate limiting remote address: "+key)
			w.WriteHeader(errs.ErrSenderTooManyRequest.Status())

			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *state) checkSystemAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")

		var allow bool

		if (r.URL.Path == "/etcha/v1/system/metrics" && s.Config.Run.SystemMetricsSecret != "" && key == s.Config.Run.SystemMetricsSecret) ||
			(strings.HasPrefix(r.URL.Path, "/etcha/v1/system/pprof") && s.Config.Run.SystemPprofSecret != "" && key == s.Config.Run.SystemPprofSecret) {
			allow = true
		}

		if !allow {
			w.WriteHeader(errs.ErrSenderForbidden.Status())

			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *state) setContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = logger.SetFormat(ctx, s.Config.CLI.LogFormat)
		ctx = logger.SetLevel(ctx, s.Config.CLI.LogLevel)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
