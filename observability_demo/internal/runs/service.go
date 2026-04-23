package runs

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"encroach-demo/internal/obs"
	"encroach-demo/internal/tiles"
)

var tracer = otel.Tracer("runs")

type Run struct {
	ID          int64      `json:"id"`
	RunnerID    int64      `json:"runner_id"`
	RunnerName  string     `json:"runner_name,omitempty"`
	TilesClaimed int       `json:"tiles_claimed"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type RunRequest struct {
	RunnerID int64      `json:"runner_id"`
	Tiles    [][2]int   `json:"tiles"` // [[x,y], [x,y], ...]
}

type Service struct {
	db      *sql.DB
	tileSvc *tiles.Service
}

func NewService(db *sql.DB, tileSvc *tiles.Service) *Service {
	return &Service{db: db, tileSvc: tileSvc}
}

func (s *Service) SubmitRun(ctx context.Context, req RunRequest) (*Run, error) {
	ctx, span := tracer.Start(ctx, "runs.SubmitRun")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("runner.id", req.RunnerID),
		attribute.Int("tiles.submitted", len(req.Tiles)),
	)

	var run Run
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO runs (runner_id) VALUES ($1) RETURNING id, runner_id, tiles_claimed, created_at`,
		req.RunnerID,
	).Scan(&run.ID, &run.RunnerID, &run.TilesClaimed, &run.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert run: %w", err)
	}

	slog.InfoContext(ctx, "run submitted",
		"run_id", run.ID,
		"runner_id", req.RunnerID,
		"tiles", len(req.Tiles),
		"trace_id", span.SpanContext().TraceID().String(),
	)

	return &run, nil
}

// ProcessRun claims tiles and marks the run as processed.
func (s *Service) ProcessRun(ctx context.Context, runID int64, runnerID int64, coords [][2]int) error {
	ctx, span := tracer.Start(ctx, "runs.ProcessRun")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("run.id", runID),
		attribute.Int64("runner.id", runnerID),
		attribute.Int("tiles.count", len(coords)),
	)

	claimed, err := s.tileSvc.ClaimTiles(ctx, runnerID, runID, coords)
	if err != nil {
		return fmt.Errorf("claim tiles: %w", err)
	}

	now := time.Now()
	_, err = s.db.ExecContext(ctx,
		`UPDATE runs SET tiles_claimed = $1, processed_at = $2 WHERE id = $3`,
		claimed, now, runID,
	)
	if err != nil {
		return fmt.Errorf("update run: %w", err)
	}

	obs.RunsProcessed.Inc()

	slog.InfoContext(ctx, "run processed",
		"run_id", runID,
		"runner_id", runnerID,
		"tiles_claimed", claimed,
		"trace_id", span.SpanContext().TraceID().String(),
	)

	return nil
}

type LeaderboardEntry struct {
	RunnerID   int64  `json:"runner_id"`
	RunnerName string `json:"runner_name"`
	TileCount  int    `json:"tile_count"`
	RunCount   int    `json:"run_count"`
}

func (s *Service) Leaderboard(ctx context.Context) ([]LeaderboardEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT tc.runner_id, r.name, COUNT(DISTINCT (tc.x, tc.y)) AS tiles, COUNT(DISTINCT tc.run_id) AS runs
		 FROM tile_claims tc
		 JOIN runners r ON r.id = tc.runner_id
		 GROUP BY tc.runner_id, r.name
		 ORDER BY tiles DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.RunnerID, &e.RunnerName, &e.TileCount, &e.RunCount); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (s *Service) ListRuns(ctx context.Context) ([]Run, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT r.id, r.runner_id, rn.name, r.tiles_claimed, r.processed_at, r.created_at
		 FROM runs r JOIN runners rn ON rn.id = r.runner_id
		 ORDER BY r.created_at DESC LIMIT 50`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		var r Run
		if err := rows.Scan(&r.ID, &r.RunnerID, &r.RunnerName, &r.TilesClaimed, &r.ProcessedAt, &r.CreatedAt); err != nil {
			continue
		}
		runs = append(runs, r)
	}
	return runs, nil
}
