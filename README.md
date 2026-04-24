# Finding Bugs Before Your Users Do

**Gophercamp CZ 2026 — Brno, Czech Republic — April 24, 2026**

Talk by [Aaron Cruz](https://encroach.app) · [@mraaroncruz](https://twitter.com/mraaroncruz)

---

## What this repo contains

- **`slides/`** — the talk slides (`presentation.html`, open in any browser)
- **`observability_demo/`** — the Go demo app used during the talk, with full instructions in [`observability_demo/HOWTO.md`](observability_demo/HOWTO.md)

## The demo app

The demo is a simplified version of [Encroach](https://encroach.app) — a territory-running game where you claim your city one run at a time. Runners submit GPS tracks, the app computes which grid tiles they covered, and those tiles show up on a shared map.

The app ships with a deliberately broken mode and a full observability stack (Prometheus, Grafana, OpenTelemetry, OpenObserve) that you can run locally with one command:

```bash
cd observability_demo
bin/demo up      # start Postgres, Redis, Prometheus, Grafana, OpenObserve
bin/demo full    # start the app with full observability + the bug active
bin/demo seed    # generate some runs to populate the dashboards
```

See [`observability_demo/HOWTO.md`](observability_demo/HOWTO.md) for the full walkthrough.

## Requirements

- Go 1.21+
- Docker (for the infrastructure stack)
- Node.js (for the frontend, built automatically by `bin/demo`)

## Links

- [Encroach](https://encroach.app) — the real app
- [Gophercamp CZ](https://www.gophercamp.cz) — the conference
