-- +goose Up
CREATE TABLE lessons (
    id          VARCHAR(255) PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenant(id),
    project_id  UUID REFERENCES project(id),
    issue_id    VARCHAR(255) REFERENCES issues(id) ON DELETE SET NULL,
    title       VARCHAR(500) NOT NULL,
    mistake     TEXT NOT NULL,
    correction  TEXT NOT NULL,
    expert      VARCHAR(255),
    components  TEXT[] NOT NULL DEFAULT '{}',
    severity    INT NOT NULL DEFAULT 2,
    status      VARCHAR(32) NOT NULL DEFAULT 'open',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  VARCHAR(255),
    resolved_at TIMESTAMPTZ,
    resolved_by VARCHAR(255)
);
CREATE INDEX idx_lessons_tenant ON lessons(tenant_id);
CREATE INDEX idx_lessons_project ON lessons(project_id);
CREATE INDEX idx_lessons_status ON lessons(status);
CREATE INDEX idx_lessons_severity ON lessons(severity);
CREATE INDEX idx_lessons_issue ON lessons(issue_id);
CREATE INDEX idx_lessons_components ON lessons USING GIN(components);

-- +goose Down
DROP TABLE IF EXISTS lessons;
