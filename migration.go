package taskifyv2

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed db/migration/*.sql
var migrationFiles embed.FS

func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrationFiles)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "db/migration"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
