-- +goose Up

CREATE TABLE config (
    key   VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE metadata (
    key   VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS metadata;
DROP TABLE IF EXISTS config;
