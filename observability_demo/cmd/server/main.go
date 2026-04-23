package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"encroach-demo/internal/api"
	"encroach-demo/internal/cache"
	"encroach-demo/internal/config"
	"encroach-demo/internal/database"
	"encroach-demo/internal/obs"
	"encroach-demo/internal/runs"
	"encroach-demo/internal/tiles"
)

func main() {
	cfg := config.Load()

	// --- Observability ---
	logger, shutdownLogs := obs.InitLogging(cfg.ObservabilityLevel, cfg.OTLPEndpoint, cfg.OTLPHeaders)
	defer shutdownLogs()

	if cfg.ObservabilityLevel.Has(config.ObsTraces) {
		shutdownTraces := obs.InitTracing(cfg.OTLPEndpoint, cfg.OTLPHeaders)
		defer shutdownTraces()
	}

	// --- Dependencies ---
	db := database.Connect(cfg.DatabaseURL)
	defer db.Close()

	rdb := cache.Connect(cfg.RedisURL)
	defer rdb.Close()

	tileSvc := tiles.NewService(db, rdb, cfg.TileStaleCache)
	runSvc := runs.NewService(db, tileSvc)

	// --- Worker pool ---
	pool := runs.NewWorkerPool(runSvc, cfg.Workers)
	go pool.Start()
	defer pool.Stop()

	// --- HTTP server ---
	handler := api.NewRouter(cfg, tileSvc, runSvc, pool, logger)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		<-sigCh

		slog.Info("shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	logger.Info("server starting",
		"port", cfg.Port,
		"observability", cfg.ObservabilityLevel.String(),
		"stale_cache", cfg.TileStaleCache,
	)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
