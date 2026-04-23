# How to Use the Encroach Observability Demo

Demo app for the Gophercamp CZ 2026 talk: **"Finding Bugs Before Your Users Do"**

## Quick Start

```bash
# 1. Start infrastructure
bin/demo up

# 2. Run the demo
bin/demo full

# 3. In another terminal, seed some data
bin/demo seed
```

## The Demo Script

Use `bin/demo` to run the app in different modes:

```bash
bin/demo              # Show available modes
bin/demo broken       # Bug enabled, no observability (the "dark" debugging experience)
bin/demo metrics      # Bug enabled + Prometheus metrics
bin/demo traces       # Bug enabled + OpenTelemetry traces
bin/demo full         # Bug enabled + full observability stack
bin/demo fixed        # Bug bypassed + full observability (shows the fix)
```

## The Story

**The Scenario**: A runner posts in support: "I ran 8 miles and claimed 14 blocks. My map shows zero new territory."

**The Bug**: The tile-ownership cache is ETag-validated against the upstream map-tile provider. The provider returns 304 Not Modified (map data unchanged), so we keep serving cached ownership data — which doesn't include recent claims. No errors anywhere.

**The Investigation**:
1. `bin/demo broken` — try to debug with just logs (you can't)
2. `bin/demo metrics` — Grafana shows 100% cache hit rate (suspicious!)
3. `bin/demo traces` — OpenObserve shows traces ending at Redis, no upstream call
4. `bin/demo full` — structured logs reveal cache is 47 hours stale, same ETag for days

## URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| App | http://localhost:3000 | — |
| Grafana | http://localhost:3001 | admin / admin |
| OpenObserve | http://localhost:5080 | root@example.com / Complexpass#123 |
| Prometheus | http://localhost:9090 | — |

## App Features

| Action | What It Does |
|--------|--------------|
| GET /api/territory | Returns 20×20 grid with tile ownership |
| POST /api/runs | Submit a run (tiles to claim) |
| GET /api/runs | List recent runs |
| GET /api/leaderboard | Territory leaderboard |
| GET /metrics | Prometheus metrics endpoint |

## Environment Variables

The demo script sets these for you, but for reference:

| Variable | Values | Effect |
|----------|--------|--------|
| `OBSERVABILITY_LEVEL` | `none`, `metrics`, `traces`, `logs`, `full` | What observability is enabled |
| `TILE_STALE_CACHE` | `true`/`false` | Enable the stale cache bug |
| `TILE_MOCK_MODE` | `true`/`false` | Use mock tile provider (always true for demos) |
| `TILE_FORCE_REFRESH` | `true`/`false` | Bypass cache entirely (the "fix") |

## Demo Walkthrough

### Act 1: The Problem

**Run: `bin/demo broken`**

1. Open http://localhost:3000/api/territory — note the ownership state
2. In another terminal: `bin/demo seed` — submit runs
3. Check territory again — ownership should update but some tiles are stale
4. Check server logs — no errors, all runs "processed successfully"
5. "The system says everything is fine, but tiles aren't updating."

**Stop the server (Ctrl+C) before moving to the next step.**

### Act 2: The Investigation

#### Step 2a: Add Metrics

**Run: `bin/demo metrics`**

1. Seed some data: `bin/demo seed`
2. Open Grafana at http://localhost:3001
3. Find the "Encroach — Tile Service" dashboard
4. Notice: cache hit rate is 100%, tile provider calls at zero
5. **First clue**: We're never refreshing tile data

#### Step 2b: Add Traces

**Run: `bin/demo traces`**

1. Seed data: `bin/demo seed`
2. Open OpenObserve at http://localhost:5080
3. Go to Traces, search for service "encroach"
4. Find a `tiles.GetTerritory` trace
5. See the waterfall: handler → cache.get → done. No `mapapi.Fetch` span (or it shows 304).
6. **Second clue**: Trace ends at Redis. Cache is 47 hours old but still "valid."

#### Step 2c: Full Observability

**Run: `bin/demo full`**

1. Seed data: `bin/demo seed`
2. Server logs are now structured JSON
3. Look for `"msg":"tile provider response"` — you'll see `etag` == `etag_sent`
4. The ETag hasn't changed in 47 hours → provider keeps returning 304
5. **The bug is clear.**

### Act 3: The Fix

**Run: `bin/demo fixed`**

1. Seed data: `bin/demo seed`
2. Check Grafana — cache hit rate drops, tile provider calls resume
3. Check OpenObserve traces — `mapapi.Fetch` spans appear with status 200
4. Territory endpoint returns fresh ownership data
