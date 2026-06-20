package middleware

import (
	"net"
	"net/http"
	"strconv"

	"teamtask/internal/cache"
)

type RateLimitKeyFunc func(r *http.Request) string

func KeyByIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func KeyByUserID(r *http.Request) string {
	if id, ok := UserIDFromContext(r.Context()); ok {
		return strconv.FormatInt(id, 10)
	}
	return KeyByIP(r)
}

func RateLimit(limiter *cache.RateLimiter, keyFn RateLimitKeyFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, err := limiter.Allow(r.Context(), keyFn(r))
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if !allowed {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
