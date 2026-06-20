package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"teamtask/internal/cache"
	"teamtask/internal/jwtutil"
	custommw "teamtask/internal/middleware"
)

type Handlers struct {
	Auth  *AuthHandler
	Team  *TeamHandler
	Task  *TaskHandler
	Admin *AdminHandler
}

func NewRouter(h Handlers, issuer *jwtutil.Issuer, limiter *cache.RateLimiter) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)
	r.Use(custommw.PrometheusMetrics)
	r.Use(chimw.Logger)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api/v1", func(api chi.Router) {
		api.Group(func(pub chi.Router) {
			pub.Use(custommw.RateLimit(limiter, custommw.KeyByIP))
			pub.Post("/register", h.Auth.Register)
			pub.Post("/login", h.Auth.Login)
		})

		api.Group(func(priv chi.Router) {
			priv.Use(custommw.JWTAuth(issuer))
			priv.Use(custommw.RateLimit(limiter, custommw.KeyByUserID))

			priv.Route("/teams", func(tr chi.Router) {
				tr.Post("/", h.Team.Create)
				tr.Get("/", h.Team.List)
				tr.Post("/{id}/invite", h.Team.Invite)
			})

			priv.Route("/tasks", func(tr chi.Router) {
				tr.Post("/", h.Task.Create)
				tr.Get("/", h.Task.List)
				tr.Put("/{id}", h.Task.Update)
				tr.Get("/{id}/history", h.Task.History)
			})

			priv.Route("/admin", func(ar chi.Router) {
				ar.Get("/tasks/orphaned-assignees", h.Admin.OrphanedAssignees)
				ar.Get("/tasks/top-creators", h.Admin.TopCreators)
				ar.Get("/teams/stats", h.Admin.TeamStats)
			})
		})
	})

	return r
}
