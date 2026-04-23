package api

import (
	"encoding/json"
	"net/http"

	"encroach-demo/internal/runs"
	"encroach-demo/internal/tiles"
)

type Handlers struct {
	tiles *tiles.Service
	runs  *runs.Service
	pool  *runs.WorkerPool
}

func NewHandlers(tileSvc *tiles.Service, runSvc *runs.Service, pool *runs.WorkerPool) *Handlers {
	return &Handlers{tiles: tileSvc, runs: runSvc, pool: pool}
}

func (h *Handlers) GetTerritory(w http.ResponseWriter, r *http.Request) {
	territory, err := h.tiles.GetTerritory(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, territory)
}

func (h *Handlers) SubmitRun(w http.ResponseWriter, r *http.Request) {
	var req runs.RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	run, err := h.runs.SubmitRun(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.pool.Submit(runs.Job{
		RunID:    run.ID,
		RunnerID: run.RunnerID,
		Tiles:    req.Tiles,
	})

	w.WriteHeader(http.StatusAccepted)
	writeJSON(w, run)
}

func (h *Handlers) ListRuns(w http.ResponseWriter, r *http.Request) {
	list, err := h.runs.ListRuns(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, list)
}

func (h *Handlers) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	entries, err := h.runs.Leaderboard(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []runs.LeaderboardEntry{}
	}
	writeJSON(w, entries)
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
