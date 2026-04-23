# Encroach Observability Demo

Demo app for the Gophercamp CZ 2026 talk *"Finding Bugs Before Your Users Do."*

A simplified **Encroach** (territory-running game) backend used to demonstrate how metrics, traces, and structured logs turn a silent production bug into a ten-minute fix.

## Intended shape

```
observability_demo/
├── cmd/server/              # main.go — chi router, middleware wiring
├── internal/
│   ├── tiles/               # tile ownership service (where the bug lives)
│   ├── runs/                # run ingest + worker pool
│   ├── mapapi/              # mock upstream map-tile client (304/ETag behavior)
│   └── obs/                 # prometheus, otel, slog setup (toggled by env)
├── frontend/                # Vite app — map + ownership overlay
├── docker-compose.yml       # app + Postgres + Redis + Prometheus + Grafana + Jaeger + OTel collector
├── grafana/                 # pre-built dashboard JSON + datasource provisioning
├── prometheus/              # scrape config + alert rules
├── bin/demo                 # mode driver (broken, metrics, traces, full, fixed)
└── HOWTO.md                 # attendee-facing walkthrough
```

## Demo modes (planned)

| Mode | Bug | Observability |
|------|-----|---------------|
| `broken` | on | none — the "dark" debugging experience |
| `metrics` | on | Prometheus only |
| `traces` | on | OTel/Jaeger only |
| `full` | on | metrics + traces + structured logs |
| `fixed` | bypassed | full — shows the fix |
