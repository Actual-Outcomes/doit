-- +goose Up
CREATE TABLE retries (
    id          VARCHAR(255) PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenant(id),
    project_id  UUID REFERENCES project(id),
    issue_id    VARCHAR(255) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    attempt     INT NOT NULL DEFAULT 1,
    status      VARCHAR(32) NOT NULL DEFAULT 'failed',
    error       TEXT NOT NULL DEFAULT '',
    agent       VARCHAR(255),
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at    TIMESTAMPTZ,
    created_by  VARCHAR(255)
);

CREATE INDEX idx_retries_issue ON retries(issue_id);
CREATE INDEX idx_retries_tenant ON retries(tenant_id);

-- +goose Down
DROP TABLE IF EXISTS retries;
