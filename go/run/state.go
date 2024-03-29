package run

import (
	"context"
	"net/http"
	"sync"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/pattern"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
	"github.com/ulule/limiter/v3"
)

type state struct {
	Config         *config.Config
	HandlersEvents map[string][]string
	HandlersRoutes map[string]string
	JWTs           *types.MapLock[pattern.JWT]
	Patterns       *types.MapLock[pattern.Pattern]
	PatternLocks   map[string]*sync.Mutex
	RateLimiter    *limiter.Limiter
	WaitGroup      sync.WaitGroup
	WebhookHandler http.Handler
}

func newState(ctx context.Context, c *config.Config) (*state, errs.Err) {
	s := &state{
		Config:         c,
		HandlersEvents: map[string][]string{},
		HandlersRoutes: map[string]string{},
		JWTs:           types.NewMapLock[pattern.JWT](),
		Patterns:       types.NewMapLock[pattern.Pattern](),
		PatternLocks:   map[string]*sync.Mutex{},
	}

	return s, logger.Error(ctx, s.initHandlers(ctx))
}
