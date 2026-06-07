package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"gateway/pkg/metrics"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

const (
	ScopeGlobal = "global"
	ScopeUser   = "user"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type userEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type UserLimiter struct {
	mu       sync.Mutex
	limit    rate.Limit
	burst    int
	ttl      time.Duration
	users    map[string]*userEntry
	lastGC   time.Time
	disabled bool
}

var (
	globalMu       sync.RWMutex
	globalLimiter  *rate.Limiter
	globalDisabled bool
	users          = NewUserLimiter(20, 20, 10*time.Minute)
)

func Configure(globalRPS, globalBurst, userRPS, userBurst int) {
	globalMu.Lock()
	defer globalMu.Unlock()

	globalDisabled = globalRPS <= 0 || globalBurst <= 0
	if globalDisabled {
		globalLimiter = nil
	} else {
		globalLimiter = rate.NewLimiter(rate.Limit(globalRPS), globalBurst)
	}

	users = NewUserLimiter(userRPS, userBurst, 10*time.Minute)
}

func NewUserLimiter(rps, burst int, ttl time.Duration) *UserLimiter {
	disabled := rps <= 0 || burst <= 0
	return &UserLimiter{
		limit:    rate.Limit(rps),
		burst:    burst,
		ttl:      ttl,
		users:    map[string]*userEntry{},
		lastGC:   time.Now(),
		disabled: disabled,
	}
}

func AllowGlobal() bool {
	globalMu.RLock()
	defer globalMu.RUnlock()

	if globalDisabled || globalLimiter == nil {
		return true
	}
	return globalLimiter.Allow()
}

func AllowUser(userID string) bool {
	return users.Allow(userID)
}

func (l *UserLimiter) Allow(userID string) bool {
	if l == nil || l.disabled || userID == "" {
		return true
	}

	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastGC) > time.Minute {
		l.gc(now)
	}

	entry := l.users[userID]
	if entry == nil {
		entry = &userEntry{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.users[userID] = entry
	}
	entry.lastSeen = now

	return entry.limiter.Allow()
}

func (l *UserLimiter) gc(now time.Time) {
	for userID, entry := range l.users {
		if now.Sub(entry.lastSeen) > l.ttl {
			delete(l.users, userID)
		}
	}
	l.lastGC = now
}

func GlobalMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !AllowGlobal() {
				return Respond(c, ScopeGlobal)
			}
			return next(c)
		}
	}
}

func Respond(c echo.Context, scope string) error {
	metrics.RecordRateLimited(scope)
	return c.JSON(http.StatusTooManyRequests, ErrorResponse{
		Error:   "rate_limit_exceeded",
		Message: "Too many requests. Please retry later.",
	})
}
