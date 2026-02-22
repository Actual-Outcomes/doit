-- +goose Up

CREATE TABLE events (
    id         BIGSERIAL PRIMARY KEY,
    issue_id   VARCHAR(255) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    event_type VARCHAR(32) NOT NULL,
    actor      VARCHAR(255) NOT NULL,
    old_value  TEXT,
    new_value  TEXT,
    comment    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_issue ON events (issue_id);
CREATE INDEX idx_events_type ON events (event_type);

-- +goose Down
DROP TABLE IF EXISTS events;
