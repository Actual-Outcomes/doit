-- +goose Up

CREATE TABLE issues (
    id              VARCHAR(255) PRIMARY KEY,
    content_hash    VARCHAR(64),
    title           VARCHAR(500) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    design          TEXT NOT NULL DEFAULT '',
    acceptance_criteria TEXT NOT NULL DEFAULT '',
    notes           TEXT NOT NULL DEFAULT '',
    spec_id         VARCHAR(1024),
    status          VARCHAR(32) NOT NULL DEFAULT 'open',
    priority        INT NOT NULL DEFAULT 2,
    issue_type      VARCHAR(32) NOT NULL DEFAULT 'task',
    assignee        VARCHAR(255),
    owner           VARCHAR(255),
    estimated_minutes INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      VARCHAR(255),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at       TIMESTAMPTZ,
    due_at          TIMESTAMPTZ,
    defer_until     TIMESTAMPTZ,
    close_reason    TEXT,
    closed_by_session VARCHAR(255),
    external_ref    VARCHAR(255),
    source_system   VARCHAR(255),
    source_repo     VARCHAR(512),
    metadata        JSONB,
    compaction_level INT NOT NULL DEFAULT 0,
    compacted_at    TIMESTAMPTZ,
    compacted_at_commit VARCHAR(64),
    original_size   INT NOT NULL DEFAULT 0,
    sender          VARCHAR(255),
    ephemeral       BOOLEAN NOT NULL DEFAULT FALSE,
    mol_type        VARCHAR(32),
    work_type       VARCHAR(32) DEFAULT 'mutex',
    crystallizes    BOOLEAN NOT NULL DEFAULT FALSE,
    wisp_type       VARCHAR(32),
    pinned          BOOLEAN NOT NULL DEFAULT FALSE,
    is_template     BOOLEAN NOT NULL DEFAULT FALSE,
    quality_score   DOUBLE PRECISION,
    event_kind      VARCHAR(32),
    actor           VARCHAR(255),
    target          VARCHAR(255),
    payload         TEXT,
    await_type      VARCHAR(32),
    await_id        VARCHAR(255),
    timeout_ns      BIGINT,
    agent_state     VARCHAR(32),
    last_activity   TIMESTAMPTZ,
    role_type       VARCHAR(32),
    rig             VARCHAR(255),
    hook_bead       VARCHAR(255),
    role_bead       VARCHAR(255)
);

CREATE INDEX idx_issues_status ON issues (status);
CREATE INDEX idx_issues_priority ON issues (priority);
CREATE INDEX idx_issues_issue_type ON issues (issue_type);
CREATE INDEX idx_issues_assignee ON issues (assignee);
CREATE INDEX idx_issues_created_at ON issues (created_at);
CREATE INDEX idx_issues_updated_at ON issues (updated_at);
CREATE INDEX idx_issues_content_hash ON issues (content_hash);
CREATE INDEX idx_issues_ephemeral ON issues (ephemeral) WHERE ephemeral = TRUE;

-- +goose Down
DROP TABLE IF EXISTS issues;
