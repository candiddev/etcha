package run

import (
	"context"
	"fmt"
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

func (s *state) newMux(ctx context.Context) http.Handler {
	r := chi.NewRouter()

	rate, err := limiter.NewRateFromFormatted(s.Config.Run.RateLimiterRate)
	if err != nil {
		logger.Error(ctx, errs.ErrReceiver.Wrap(err)) //nolint:errcheck
		panic(err)
	}

	store := memory.NewStore()

	s.RateLimiter = limiter.New(store, rate, limiter.WithTrustForwardHeader(true))

	r.Use(middleware.Recoverer)
	r.Use(s.checkRateLimiter)
	r.Post("/etcha/v1/push/{config}", s.postPush)
	r.Route("/etcha/v1/system", func(r chi.Router) {
		r.Use(s.checkSystemAuth)
		r.Handle("/metrics", promhttp.Handler())
		r.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
		r.Handle("/pprof/heap", pprof.Handler("heap"))
	})

	h, e := s.Config.Handlers.RegisterWebhooks(ctx, s.Config.CLI)
	if e != nil {
		logger.Error(ctx, e) //nolint:errcheck
		panic(err)
	}

	r.NotFound(h.ServeHTTP)

	return r
}

func (s *state) checkRateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.Trace(r.Context())
		key := s.RateLimiter.GetIPKey(r) + chi.RouteContext(ctx).RoutePattern()

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
			logger.Info(ctx, fmt.Sprintf("Rate limiting remote address: %s", key))
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
