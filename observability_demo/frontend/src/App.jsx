import { useState, useEffect, useCallback } from "react";
import Grid from "./Grid.jsx";
import RunLog from "./RunLog.jsx";
import Leaderboard from "./Leaderboard.jsx";

const RUNNERS = [
  { id: 1, name: "Marcus", color: "#22c55e" },
  { id: 2, name: "Priya", color: "#3b82f6" },
  { id: 3, name: "Demo Runner", color: "#f59e0b" },
];

export default function App() {
  const [territory, setTerritory] = useState([]);
  const [runs, setRuns] = useState([]);
  const [activeRunner, setActiveRunner] = useState(RUNNERS[2]);
  const [submitting, setSubmitting] = useState(false);
  const [lastRefresh, setLastRefresh] = useState(null);
  const [leaderboard, setLeaderboard] = useState([]);

  const fetchTerritory = useCallback(async () => {
    try {
      const res = await fetch("/api/territory");
      const data = await res.json();
      setTerritory(data);
      setLastRefresh(new Date());
    } catch (e) {
      console.error("Failed to fetch territory:", e);
    }
  }, []);

  const fetchRuns = useCallback(async () => {
    try {
      const res = await fetch("/api/runs");
      const data = await res.json();
      setRuns(data || []);
    } catch (e) {
      console.error("Failed to fetch runs:", e);
    }
  }, []);

  const fetchLeaderboard = useCallback(async () => {
    try {
      const res = await fetch("/api/leaderboard");
      const data = await res.json();
      setLeaderboard(data || []);
    } catch (e) {
      console.error("Failed to fetch leaderboard:", e);
    }
  }, []);

  useEffect(() => {
    fetchTerritory();
    fetchRuns();
    fetchLeaderboard();
    const interval = setInterval(() => {
      fetchTerritory();
      fetchRuns();
      fetchLeaderboard();
    }, 5000);
    return () => clearInterval(interval);
  }, [fetchTerritory, fetchRuns, fetchLeaderboard]);

  const submitRun = async () => {
    setSubmitting(true);

    // Generate a random cluster of tiles to claim
    const startX = Math.floor(Math.random() * 15);
    const startY = Math.floor(Math.random() * 15);
    const tiles = [];
    for (let dx = 0; dx < 5; dx++) {
      for (let dy = 0; dy < 3; dy++) {
        tiles.push([startX + dx, startY + dy]);
      }
    }

    try {
      await fetch("/api/runs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          runner_id: activeRunner.id,
          tiles,
        }),
      });

      // Brief delay then refresh — gives the worker time to process
      setTimeout(() => {
        fetchTerritory();
        fetchRuns();
        fetchLeaderboard();
        setSubmitting(false);
      }, 800);
    } catch (e) {
      console.error("Failed to submit run:", e);
      setSubmitting(false);
    }
  };

  const stats = {
    total: territory.length,
    claimed: territory.filter((t) => t.owner_id).length,
    mine: territory.filter((t) => t.owner_id === activeRunner.id).length,
  };

  return (
    <div>
      <header style={{ marginBottom: "2rem" }}>
        <h1
          style={{
            fontSize: "1.75rem",
            fontWeight: 800,
            letterSpacing: "-0.02em",
          }}
        >
          Encroach
          <span
            style={{
              fontSize: "0.875rem",
              fontWeight: 400,
              color: "#94a3b8",
              marginLeft: "0.75rem",
            }}
          >
            Claim your city, one run at a time
          </span>
        </h1>
      </header>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 320px",
          gap: "2rem",
          alignItems: "start",
        }}
      >
        <div>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              marginBottom: "1rem",
            }}
          >
            <h2
              style={{ fontSize: "1rem", fontWeight: 600, color: "#94a3b8" }}
            >
              Territory Map
            </h2>
            <div
              style={{ display: "flex", gap: "0.5rem", alignItems: "center" }}
            >
              {RUNNERS.map((r) => (
                <button
                  key={r.id}
                  onClick={() => setActiveRunner(r)}
                  style={{
                    padding: "0.25rem 0.75rem",
                    borderRadius: "9999px",
                    border:
                      activeRunner.id === r.id
                        ? `2px solid ${r.color}`
                        : "2px solid #334155",
                    background:
                      activeRunner.id === r.id ? r.color + "22" : "transparent",
                    color: activeRunner.id === r.id ? r.color : "#94a3b8",
                    cursor: "pointer",
                    fontSize: "0.8125rem",
                    fontWeight: 500,
                  }}
                >
                  {r.name}
                </button>
              ))}
            </div>
          </div>

          <Grid territory={territory} runners={RUNNERS} activeRunner={activeRunner} />

          <div
            style={{
              display: "flex",
              gap: "1.5rem",
              marginTop: "1rem",
              fontSize: "0.8125rem",
              color: "#64748b",
            }}
          >
            <span>
              {stats.claimed}/{stats.total} claimed
            </span>
            <span style={{ color: activeRunner.color }}>
              {stats.mine} yours
            </span>
            {lastRefresh && (
              <span>
                Updated {lastRefresh.toLocaleTimeString()}
              </span>
            )}
          </div>
        </div>

        <aside>
          <button
            onClick={submitRun}
            disabled={submitting}
            style={{
              width: "100%",
              padding: "0.75rem",
              borderRadius: "8px",
              border: "none",
              background: submitting ? "#334155" : activeRunner.color,
              color: "#fff",
              fontWeight: 700,
              fontSize: "0.9375rem",
              cursor: submitting ? "wait" : "pointer",
              marginBottom: "1.5rem",
              transition: "background 0.15s",
            }}
          >
            {submitting ? "Processing..." : `Submit Run as ${activeRunner.name}`}
          </button>

          <Leaderboard entries={leaderboard} runners={RUNNERS} />

          <RunLog runs={runs} runners={RUNNERS} />
        </aside>
      </div>
    </div>
  );
}
