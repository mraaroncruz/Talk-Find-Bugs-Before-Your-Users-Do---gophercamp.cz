export default function RunLog({ runs, runners }) {
  if (!runs.length) {
    return (
      <div>
        <h3 style={{ fontSize: "0.875rem", fontWeight: 600, color: "#94a3b8", marginBottom: "0.75rem" }}>
          Recent Runs
        </h3>
        <p style={{ fontSize: "0.8125rem", color: "#475569" }}>
          No runs yet. Submit one!
        </p>
      </div>
    );
  }

  return (
    <div>
      <h3 style={{ fontSize: "0.875rem", fontWeight: 600, color: "#94a3b8", marginBottom: "0.75rem" }}>
        Recent Runs
      </h3>
      <div style={{ display: "flex", flexDirection: "column", gap: "0.5rem" }}>
        {runs.slice(0, 15).map((run) => {
          const runner = runners.find((r) => r.id === run.runner_id);
          const color = runner?.color || "#6366f1";
          const name = run.runner_name || runner?.name || "Unknown";
          const time = new Date(run.created_at).toLocaleTimeString();
          const processed = !!run.processed_at;

          return (
            <div
              key={run.id}
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                padding: "0.5rem 0.75rem",
                background: "#1e293b",
                borderRadius: "6px",
                borderLeft: `3px solid ${color}`,
                fontSize: "0.8125rem",
              }}
            >
              <div>
                <span style={{ color, fontWeight: 600 }}>{name}</span>
                <span style={{ color: "#64748b", marginLeft: "0.5rem" }}>
                  {run.tiles_claimed} tiles
                </span>
              </div>
              <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
                <span style={{ color: "#475569", fontSize: "0.75rem" }}>{time}</span>
                <span
                  style={{
                    width: "8px",
                    height: "8px",
                    borderRadius: "50%",
                    background: processed ? "#22c55e" : "#eab308",
                  }}
                  title={processed ? "Processed" : "Pending"}
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
