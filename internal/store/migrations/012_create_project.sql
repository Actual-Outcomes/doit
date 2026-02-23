-- +goose Up

CREATE TABLE project (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenant(id),
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, slug)
);

ALTER TABLE issues ADD COLUMN project_id UUID REFERENCES project(id);
CREATE INDEX idx_issues_project ON issues(project_id);

-- +goose Down

DROP INDEX IF EXISTS idx_issues_project;
ALTER TABLE issues DROP COLUMN IF EXISTS project_id;
DROP TABLE IF EXISTS project;
