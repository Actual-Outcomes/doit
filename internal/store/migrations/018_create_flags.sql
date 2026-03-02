-- +goose Up
CREATE TABLE flags (
    id          VARCHAR(255) PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenant(id),
    project_id  UUID REFERENCES project(id),
    issue_id    VARCHAR(255) REFERENCES issues(id) ON DELETE SET NULL,
    type        VARCHAR(64) NOT NULL,
    severity    INT NOT NULL DEFAULT 2,
    summary     TEXT NOT NULL,
    context     JSONB NOT NULL DEFAULT '{}',
    status      VARCHAR(32) NOT NULL DEFAULT 'open',
    resolution  TEXT,
    resolved_by VARCHAR(255),
    resolved_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  VARCHAR(255)
);
CREATE INDEX idx_flags_tenant ON flags(tenant_id);
CREATE INDEX idx_flags_project ON flags(project_id);
CREATE INDEX idx_flags_issue ON flags(issue_id);
CREATE INDEX idx_flags_status ON flags(status);
CREATE INDEX idx_flags_severity ON flags(severity);
CREATE INDEX idx_flags_type ON flags(type);

-- +goose Down
DROP TABLE IF EXISTS flags;
