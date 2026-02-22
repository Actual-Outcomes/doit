-- +goose Up

CREATE TABLE child_counters (
    parent_id  VARCHAR(255) PRIMARY KEY,
    last_child INT NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE IF EXISTS child_counters;
