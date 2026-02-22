-- +goose Up

CREATE TABLE dependencies (
    issue_id      VARCHAR(255) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    depends_on_id VARCHAR(255) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    type          VARCHAR(32) NOT NULL DEFAULT 'blocks',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by    VARCHAR(255),
    metadata      JSONB,
    thread_id     VARCHAR(255),
    PRIMARY KEY (issue_id, depends_on_id)
);

CREATE INDEX idx_deps_depends_on ON dependencies (depends_on_id);
CREATE INDEX idx_deps_type ON dependencies (type);
CREATE INDEX idx_deps_thread ON dependencies (thread_id) WHERE thread_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS dependencies;
