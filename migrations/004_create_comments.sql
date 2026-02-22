-- +goose Up

CREATE TABLE comments (
    id        BIGSERIAL PRIMARY KEY,
    issue_id  VARCHAR(255) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    author    VARCHAR(255) NOT NULL,
    text      TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_issue ON comments (issue_id);

-- +goose Down
DROP TABLE IF EXISTS comments;
