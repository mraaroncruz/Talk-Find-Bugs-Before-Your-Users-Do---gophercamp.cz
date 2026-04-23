package config

import (
	"os"
	"strconv"
	"strings"
)

type ObsLevel int

const (
	ObsNone    ObsLevel = 0
	ObsMetrics ObsLevel = 1 << iota
	ObsTraces
	ObsLogs
	ObsFull = ObsMetrics | ObsTraces | ObsLogs
)

func (l ObsLevel) Has(flag ObsLevel) bool { return l&flag != 0 }

func (l ObsLevel) String() string {
	switch l {
	case ObsFull:
		return "full"
	case ObsMetrics:
		return "metrics"
	case ObsTraces:
		return "traces"
	case ObsLogs:
		return "logs"
	default:
		return "none"
	}
}

func parseObsLevel(s string) ObsLevel {
	switch strings.ToLower(s) {
	case "metrics":
		return ObsMetrics
	case "traces":
		return ObsTraces
	case "logs":
		return ObsLogs
	case "full":
		return ObsFull
	default:
		return ObsNone
	}
}

type Config struct {
	Port               string
	DatabaseURL        string
	RedisURL           string
	ObservabilityLevel ObsLevel
	TileStaleCache     bool
	OTLPEndpoint       string
	OTLPHeaders        map[string]string
	Workers            int
}

func Load() *Config {
	return &Config{
		Port:               envOr("PORT", "3000"),
		DatabaseURL:        envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/encroach_demo?sslmode=disable"),
		RedisURL:           envOr("REDIS_URL", "redis://localhost:6379/0"),
		ObservabilityLevel: parseObsLevel(envOr("OBSERVABILITY_LEVEL", "none")),
		TileStaleCache:     envOr("TILE_STALE_CACHE", "false") == "true",
		OTLPEndpoint:       envOr("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:5080"),
		OTLPHeaders:        parseHeaders(envOr("OTEL_EXPORTER_OTLP_HEADERS", "")),
		Workers:            envInt("WORKERS", 4),
	}
}

func parseHeaders(s string) map[string]string {
	headers := map[string]string{}
	if s == "" {
		return headers
	}
	for _, pair := range strings.Split(s, ",") {
		k, v, ok := strings.Cut(pair, "=")
		if ok {
			headers[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return headers
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
