export default function Leaderboard({ entries, runners }) {
  if (!entries.length) return null;

  return (
    <div style={{ marginTop: "1.5rem" }}>
      <h3 style={{ fontSize: "0.875rem", fontWeight: 600, color: "#94a3b8", marginBottom: "0.75rem" }}>
        Leaderboard
      </h3>
      <div style={{ display: "flex", flexDirection: "column", gap: "0.375rem" }}>
        {entries.map((entry, i) => {
          const runner = runners.find((r) => r.id === entry.runner_id);
          const color = runner?.color || "#6366f1";
          return (
            <div
              key={entry.runner_id}
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
              <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
                <span style={{ color: "#475569", fontWeight: 700, fontSize: "0.75rem" }}>
                  #{i + 1}
                </span>
                <span style={{ color, fontWeight: 600 }}>{entry.runner_name}</span>
              </div>
              <div style={{ display: "flex", gap: "1rem", color: "#64748b", fontSize: "0.75rem" }}>
                <span>{entry.tile_count} tiles</span>
                <span>{entry.run_count} runs</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
