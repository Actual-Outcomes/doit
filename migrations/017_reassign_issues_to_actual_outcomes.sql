-- +goose Up

-- Reassign all issues to the Actual Outcomes tenant.
-- Migration 016 backfilled tenant_id from project, but issues with NULL project_id
-- (created before projects existed) fell to the "oldest tenant" fallback which may
-- differ from the active Actual Outcomes tenant.
UPDATE issues SET tenant_id = (SELECT id FROM tenant WHERE slug = 'actual-outcomes');

-- +goose Down
-- No rollback — tenant assignment was already wrong before this fix.
