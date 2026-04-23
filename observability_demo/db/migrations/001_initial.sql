CREATE TABLE IF NOT EXISTS runners (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS runs (
    id BIGSERIAL PRIMARY KEY,
    runner_id BIGINT NOT NULL REFERENCES runners(id),
    tiles_claimed INT NOT NULL DEFAULT 0,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tile_claims (
    id BIGSERIAL PRIMARY KEY,
    runner_id BIGINT NOT NULL REFERENCES runners(id),
    run_id BIGINT NOT NULL REFERENCES runs(id),
    x INT NOT NULL,
    y INT NOT NULL,
    claimed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (x, y)
);

CREATE INDEX idx_tile_claims_coords ON tile_claims (x, y);
CREATE INDEX idx_runs_runner ON runs (runner_id);

-- Seed data: a few runners
INSERT INTO runners (name) VALUES ('Marcus'), ('Priya'), ('Demo Runner')
ON CONFLICT DO NOTHING;
