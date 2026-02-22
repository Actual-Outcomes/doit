-- +goose Up

ALTER TABLE issues ADD COLUMN tenant_id UUID REFERENCES tenant(id);
CREATE INDEX idx_issues_tenant ON issues(tenant_id);

-- +goose Down

DROP INDEX IF EXISTS idx_issues_tenant;
ALTER TABLE issues DROP COLUMN IF EXISTS tenant_id;
