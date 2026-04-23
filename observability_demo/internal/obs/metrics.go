package obs

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TileCacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tile_cache_hits_total",
		Help: "Number of tile cache hits",
	})

	TileCacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tile_cache_misses_total",
		Help: "Number of tile cache misses",
	})

	TileCacheAgeHours = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tile_cache_age_hours",
		Help: "Age of the current tile cache entry in hours",
	})

	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by method, path, and status",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	RunsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "runs_processed_total",
		Help: "Total runs processed by workers",
	})
)

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(ww, r)

		path := chi.RouteContext(r.Context()).RoutePattern()
		if path == "" {
			path = r.URL.Path
		}

		HTTPRequestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(ww.status)).Inc()
		HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
