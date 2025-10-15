package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/marcbran/versource/pkg/versource"
	"github.com/pressly/goose/v3"
)

//go:embed main/*.sql
var embedMainMigrations embed.FS

//go:embed admin/*.sql
var embedAdminMigrations embed.FS

func Migrate(ctx context.Context, config *versource.DatabaseConfig) error {
	err := goose.SetDialect("mysql")
	if err != nil {
		return err
	}

	err = migrateBranch(ctx, config, "admin", embedAdminMigrations, "admin")
	if err != nil {
		return err
	}

	err = migrateBranch(ctx, config, "main", embedMainMigrations, "main")
	if err != nil {
		return err
	}

	return nil
}

func migrateBranch(ctx context.Context, config *versource.DatabaseConfig, branchName string, embedFS embed.FS, migrationPath string) error {
	db, err := newDb(config)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping %s database: %w", branchName, err)
	}

	err = ensureBranch(db, branchName)
	if err != nil {
		return fmt.Errorf("failed to ensure %s branch: %w", branchName, err)
	}

	goose.SetBaseFS(embedFS)
	err = goose.UpContext(ctx, db, migrationPath)
	if err != nil {
		return fmt.Errorf("failed to apply %s migrations: %w", branchName, err)
	}

	err = commit(db, branchName)
	if err != nil {
		return err
	}

	return nil
}

func newDb(config *versource.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	return db, nil
}

func ensureBranch(db *sql.DB, branch string) error {
	var activeBranch sql.NullString
	err := db.QueryRow("SELECT active_branch()").Scan(&activeBranch)
	if err != nil {
		return fmt.Errorf("failed to get active branch: %w", err)
	}

	if activeBranch.Valid && activeBranch.String == branch {
		return nil
	}

	var count int64
	err = db.QueryRow("SELECT COUNT(*) FROM dolt_branches WHERE name = ?", branch).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if count == 0 {
		_, err = db.Exec("CALL DOLT_BRANCH(?, ?)", branch, "main")
		if err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branch, err)
		}
	}

	_, err = db.Exec("CALL DOLT_CHECKOUT(?)", branch)
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branch, err)
	}

	return nil
}

func commit(db *sql.DB, branchName string) error {
	var count int64
	err := db.QueryRow("SELECT COUNT(*) FROM dolt_diff WHERE commit_hash = 'WORKING'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for %s changes: %w", branchName, err)
	}

	if count == 0 {
		return nil
	}

	_, err = db.Exec("CALL DOLT_ADD('.')")
	if err != nil {
		return fmt.Errorf("failed to add %s changes: %w", branchName, err)
	}

	_, err = db.Exec(fmt.Sprintf("CALL DOLT_COMMIT('-m', 'migrate: apply %s branch migrations')", branchName))
	if err != nil {
		return fmt.Errorf("failed to commit %s changes: %w", branchName, err)
	}

	return nil
}
