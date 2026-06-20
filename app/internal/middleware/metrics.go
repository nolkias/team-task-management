package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"teamtask/internal/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func PrometheusMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		path := routePattern(r)
		status := strconv.Itoa(rec.status)

		metrics.RequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		metrics.RequestDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
		if rec.status >= 400 {
			metrics.ErrorsTotal.WithLabelValues(r.Method, path, status).Inc()
		}
	})
}

func routePattern(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.RoutePattern() != "" {
		return rctx.RoutePattern()
	}
	return r.URL.Path
}
