package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	visitors sync.Map
	rate     rate.Limit
	burst    int
	ttl      time.Duration
	logger   *zap.Logger
}

func NewRateLimiter(r rate.Limit, b int, ttl time.Duration, logger *zap.Logger) *RateLimiter {
	rl := &RateLimiter{
		rate:   r,
		burst:  b,
		ttl:    ttl,
		logger: logger,
	}
	go rl.cleanupVisitors()
	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	now := time.Now()

	if v, ok := rl.visitors.Load(ip); ok {
		vis := v.(*visitor)
		vis.lastSeen = now
		return vis.limiter
	}

	lim := rate.NewLimiter(rl.rate, rl.burst)
	rl.visitors.Store(ip, &visitor{
		limiter:  lim,
		lastSeen: now,
	})
	return lim
}

func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(1 * time.Minute)
		now := time.Now()

		rl.visitors.Range(func(key, value any) bool {
			vis := value.(*visitor)
			if now.Sub(vis.lastSeen) > rl.ttl {
				rl.visitors.Delete(key)
			}
			return true
		})
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := rl.extractIP(r)

		limiter := rl.getVisitor(ip)
		if !limiter.Allow() {
			rl.logger.Warn("Rate limit exceeded", zap.String("ip", ip))
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) extractIP(r *http.Request) string {
	// Try X-Forwarded-For (in case of proxy/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
