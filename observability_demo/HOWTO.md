# How to Use the Encroach Observability Demo

Demo app for **"Finding Bugs Before Your Users Do"** — Gophercamp CZ 2026.

## Quick Start

```bash
# 1. Start infrastructure
bin/demo up

# 2. Run the app
bin/demo full

# 3. In another terminal, seed some data
bin/demo seed
```

## Demo Modes

```bash
bin/demo broken    # Bug on, no observability
bin/demo metrics   # Bug on + Prometheus
bin/demo traces    # Bug on + OpenTelemetry
bin/demo full      # Bug on + metrics + traces + structured logs
bin/demo fixed     # Bug off + full observability (shows the fix)
```

## The Story

A runner posts in #support: *"I ran 8 miles and claimed 14 blocks. My map shows zero new territory."*

**The bug**: Workers write tile claims to Postgres successfully, but the Redis cache is never invalidated afterwards. Every call to `GET /api/territory` hits a warm cache and returns the same snapshot from 47 hours ago. No errors anywhere. The system looks completely healthy.

**The investigation**:
1. `bin/demo metrics` — Grafana shows cache hit rate pinned at 100%, cache age climbing. A healthy cache has misses.
2. `bin/demo traces` — OpenObserve shows every `GetTerritory` trace ending at Redis with no `buildFromDB` span. Claims are going to the database (visible in worker traces), but territory reads never consult it.
3. `bin/demo full` — structured logs show `cache_age_hours: 47` on every single request.

**The fix**: one line — `s.rdb.Del(ctx, cacheKey)` after writing claims in `ClaimTiles`.

## URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| App | http://localhost:3000 | — |
| Grafana | http://localhost:3001 | admin / admin |
| OpenObserve | http://localhost:5080 | root@example.com / Complexpass#123 |
| Prometheus | http://localhost:9090 | — |

## API

| Endpoint | Description |
|----------|-------------|
| `GET /api/territory` | 20×20 tile ownership grid |
| `POST /api/runs` | Submit a run (array of tile coordinates) |
| `GET /api/runs` | List recent runs |
| `GET /api/leaderboard` | Territory leaderboard |
| `GET /metrics` | Prometheus metrics |

## Environment Variables

`bin/demo` sets these for you. For reference:

| Variable | Values | Effect |
|----------|--------|--------|
| `OBSERVABILITY_LEVEL` | `none`, `metrics`, `traces`, `logs`, `full` | Which observability pillars are active |
| `TILE_STALE_CACHE` | `true` / `false` | Enables the bug (skips cache invalidation after claims) |
