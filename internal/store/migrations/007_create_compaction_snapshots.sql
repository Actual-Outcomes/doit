-- +goose Up

CREATE TABLE compaction_snapshots (
    id        BIGSERIAL PRIMARY KEY,
    issue_id  VARCHAR(255) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    level     INT NOT NULL,
    summary   TEXT NOT NULL,
    original  TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_compaction_issue ON compaction_snapshots (issue_id);

-- +goose Down
DROP TABLE IF EXISTS compaction_snapshots;
