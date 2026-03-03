-- +goose Up

-- Backfill tenant_id from each issue's project
UPDATE issues
SET tenant_id = p.tenant_id
FROM project p
WHERE issues.project_id = p.id
  AND issues.tenant_id IS NULL;

-- Fallback for orphan issues (NULL project_id): assign to oldest tenant
UPDATE issues
SET tenant_id = (SELECT id FROM tenant ORDER BY created_at ASC LIMIT 1)
WHERE tenant_id IS NULL;

-- Now enforce NOT NULL
ALTER TABLE issues ALTER COLUMN tenant_id SET NOT NULL;

-- +goose Down

ALTER TABLE issues ALTER COLUMN tenant_id DROP NOT NULL;
