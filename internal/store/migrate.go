package store

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// RunMigrations applies all pending database migrations.
func RunMigrations(databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("opening db for migrations: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting goose dialect: %w", err)
	}

	current, err := goose.GetDBVersion(db)
	if err != nil {
		current = 0
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	after, err := goose.GetDBVersion(db)
	if err != nil {
		after = 0
	}

	if after > current {
		slog.Info("migrations applied", "from", current, "to", after)
	} else {
		slog.Info("migrations up to date", "version", after)
	}

	return nil
}
