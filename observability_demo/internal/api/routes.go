package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"encroach-demo/internal/config"
	"encroach-demo/internal/obs"
	"encroach-demo/internal/runs"
	"encroach-demo/internal/tiles"
)

func NewRouter(cfg *config.Config, tileSvc *tiles.Service, runSvc *runs.Service, pool *runs.WorkerPool, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	if cfg.ObservabilityLevel.Has(config.ObsMetrics) {
		r.Use(obs.PrometheusMiddleware)
		r.Handle("/metrics", promhttp.Handler())
	}

	h := NewHandlers(tileSvc, runSvc, pool)

	r.Get("/health", h.HealthCheck)
	r.Route("/api", func(r chi.Router) {
		r.Get("/territory", h.GetTerritory)
		r.Post("/runs", h.SubmitRun)
		r.Get("/runs", h.ListRuns)
		r.Get("/leaderboard", h.GetLeaderboard)
	})

	// Serve frontend — falls back to a simple status page if frontend isn't built
	fs := http.Dir("frontend/dist")
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fs.Open("index.html"); err != nil {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<!DOCTYPE html>
<html><head><title>Encroach Demo</title>
<style>body{font-family:system-ui;max-width:600px;margin:4rem auto;padding:0 1rem}
a{color:#2563eb}pre{background:#f1f5f9;padding:1rem;border-radius:6px;overflow-x:auto}</style>
</head><body>
<h1>Encroach Demo</h1>
<p>API is running. Frontend not yet built.</p>
<h3>Endpoints</h3>
<pre>GET  <a href="/api/territory">/api/territory</a>  — tile ownership grid
POST /api/runs          — submit a run
GET  <a href="/api/runs">/api/runs</a>       — recent runs
GET  <a href="/api/leaderboard">/api/leaderboard</a> — tile leaderboard
GET  <a href="/health">/health</a>         — health check</pre>
</body></html>`))
			return
		}
		http.FileServer(fs).ServeHTTP(w, r)
	})
	r.Handle("/assets/*", http.FileServer(fs))

	var handler http.Handler = r

	if cfg.ObservabilityLevel.Has(config.ObsTraces) {
		handler = otelhttp.NewHandler(handler, "encroach")
	}

	return handler
}
