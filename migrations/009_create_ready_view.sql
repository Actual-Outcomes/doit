-- +goose Up

-- ready_issues: issues with status 'open', not ephemeral, not deferred,
-- and not blocked by any open (non-closed) issue via a 'blocks' dependency.
CREATE VIEW ready_issues AS
SELECT i.*
FROM issues i
WHERE i.status = 'open'
  AND i.ephemeral = FALSE
  AND (i.defer_until IS NULL OR i.defer_until <= NOW())
  AND NOT EXISTS (
      SELECT 1
      FROM dependencies d
      JOIN issues blocker ON blocker.id = d.depends_on_id
      WHERE d.issue_id = i.id
        AND d.type = 'blocks'
        AND blocker.status != 'closed'
  );

-- +goose Down
DROP VIEW IF EXISTS ready_issues;
