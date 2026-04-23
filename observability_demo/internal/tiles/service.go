package tiles

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"encroach-demo/internal/obs"
)

var tracer = otel.Tracer("tiles")

type Tile struct {
	X       int    `json:"x"`
	Y       int    `json:"y"`
	OwnerID *int64 `json:"owner_id,omitempty"`
	Owner   string `json:"owner,omitempty"`
}

type cacheEntry struct {
	Tiles    []Tile    `json:"tiles"`
	CachedAt time.Time `json:"cached_at"`
}

type Service struct {
	db         *sql.DB
	rdb        *redis.Client
	staleCache bool
}

func NewService(db *sql.DB, rdb *redis.Client, staleCache bool) *Service {
	return &Service{
		db:         db,
		rdb:        rdb,
		staleCache: staleCache,
	}
}

const gridSize = 20
const cacheKey = "territory:grid"

func (s *Service) GetTerritory(ctx context.Context) ([]Tile, error) {
	ctx, span := tracer.Start(ctx, "tiles.GetTerritory")
	defer span.End()

	cached, err := s.getFromCache(ctx)
	if err == nil && cached != nil {
		obs.TileCacheHits.Inc()
		ageHours := time.Since(cached.CachedAt).Hours()
		obs.TileCacheAgeHours.Set(ageHours)

		span.SetAttributes(
			attribute.Bool("cache.hit", true),
			attribute.Float64("cache.age_hours", ageHours),
		)

		slog.InfoContext(ctx, "tile cache hit",
			"cache_age_hours", ageHours,
			"trace_id", span.SpanContext().TraceID().String(),
		)

		return cached.Tiles, nil
	}

	obs.TileCacheMisses.Inc()
	span.SetAttributes(attribute.Bool("cache.hit", false))

	tiles, err := s.buildFromDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("build territory: %w", err)
	}

	s.setCache(ctx, tiles)
	return tiles, nil
}

// ClaimTiles records tile claims for a run. Called by the worker.
func (s *Service) ClaimTiles(ctx context.Context, runnerID, runID int64, coords [][2]int) (int, error) {
	ctx, span := tracer.Start(ctx, "tiles.ClaimTiles")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("runner.id", runnerID),
		attribute.Int("tile.count", len(coords)),
	)

	claimed := 0
	for _, c := range coords {
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO tile_claims (runner_id, run_id, x, y)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (x, y) DO UPDATE SET runner_id = $1, run_id = $2, claimed_at = now()`,
			runnerID, runID, c[0], c[1],
		)
		if err != nil {
			slog.ErrorContext(ctx, "claim tile failed", "x", c[0], "y", c[1], "error", err)
			continue
		}
		claimed++
	}

	slog.InfoContext(ctx, "tiles claimed",
		"runner_id", runnerID,
		"run_id", runID,
		"claimed", claimed,
		"total", len(coords),
		"trace_id", span.SpanContext().TraceID().String(),
	)

	// BUG: cache should be invalidated here so the next read reflects new claims.
	// In stale-cache mode this line is skipped — that's the bug.
	if !s.staleCache {
		s.rdb.Del(ctx, cacheKey)
	}

	return claimed, nil
}

func (s *Service) buildFromDB(ctx context.Context) ([]Tile, error) {
	ctx, span := tracer.Start(ctx, "tiles.buildFromDB")
	defer span.End()

	tiles := make([]Tile, 0, gridSize*gridSize)

	for y := range gridSize {
		for x := range gridSize {
			tiles = append(tiles, Tile{X: x, Y: y})
		}
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT tc.x, tc.y, tc.runner_id, r.name
		 FROM tile_claims tc
		 JOIN runners r ON r.id = tc.runner_id
		 WHERE tc.x < $1 AND tc.y < $2`,
		gridSize, gridSize,
	)
	if err != nil {
		return tiles, err
	}
	defer rows.Close()

	claims := map[[2]int]Tile{}
	for rows.Next() {
		var t Tile
		var ownerID int64
		if err := rows.Scan(&t.X, &t.Y, &ownerID, &t.Owner); err != nil {
			continue
		}
		t.OwnerID = &ownerID
		claims[[2]int{t.X, t.Y}] = t
	}

	for i := range tiles {
		if claimed, ok := claims[[2]int{tiles[i].X, tiles[i].Y}]; ok {
			tiles[i] = claimed
		}
	}

	return tiles, nil
}

func (s *Service) getFromCache(ctx context.Context) (*cacheEntry, error) {
	ctx, span := tracer.Start(ctx, "tiles.cache.get")
	defer span.End()

	val, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var entry cacheEntry
	if err := json.Unmarshal([]byte(val), &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (s *Service) setCache(ctx context.Context, tiles []Tile) {
	entry := cacheEntry{
		Tiles:    tiles,
		CachedAt: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	s.rdb.Set(ctx, cacheKey, data, 0)
}
