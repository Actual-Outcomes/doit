-- +goose Up

CREATE TABLE labels (
    issue_id VARCHAR(255) NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    label    VARCHAR(255) NOT NULL,
    PRIMARY KEY (issue_id, label)
);

CREATE INDEX idx_labels_label ON labels (label);

-- +goose Down
DROP TABLE IF EXISTS labels;
