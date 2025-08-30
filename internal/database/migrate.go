package database

import (
	"context"
	"embed"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/marcbran/versource/internal"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Migrate(ctx context.Context, config *internal.DatabaseConfig) error {
	db, err := NewDb(config)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	goose.SetBaseFS(embedMigrations)

	err = goose.SetDialect("mysql")
	if err != nil {
		return err
	}

	err = goose.UpContext(ctx, db, "migrations")
	if err != nil {
		return err
	}

	var count int64
	err = db.QueryRow("SELECT COUNT(*) FROM dolt_diff WHERE commit_hash = 'WORKING'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for changes: %w", err)
	}

	if count == 0 {
		return nil
	}

	_, err = db.Exec("CALL DOLT_ADD('.')")
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	_, err = db.Exec("CALL DOLT_COMMIT('-m', 'migrate: apply database migrations')")
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}
