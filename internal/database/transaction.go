package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/marcbran/versource/internal"
	"gorm.io/gorm"
)

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

type txKey struct{}

func withTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func getTxOrDb(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db
}

type branchKey struct{}

func withBranch(ctx context.Context, branch string) context.Context {
	return context.WithValue(ctx, branchKey{}, branch)
}

func getBranch(ctx context.Context) string {
	branch, ok := ctx.Value(branchKey{}).(string)
	if !ok {
		return ""
	}
	return branch
}

func (tm *GormTransactionManager) Do(ctx context.Context, branch, message string, fn func(ctx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := ensureBranch(tx, branch)
		if err != nil {
			return fmt.Errorf("failed to ensure branch %s: %w", branch, err)
		}

		err = fn(withTx(withBranch(ctx, branch), tx))
		if err != nil {
			return err
		}

		err = commitChanges(tx, message)
		if err != nil {
			return fmt.Errorf("failed to commit changes: %w", err)
		}

		return nil
	})
}

func (tm *GormTransactionManager) Checkout(ctx context.Context, branch string, fn func(ctx context.Context) error) error {
	err := ensureBranch(tm.db, branch)
	if err != nil {
		return fmt.Errorf("failed to ensure branch %s: %w", branch, err)
	}

	return fn(withBranch(ctx, branch))
}

func ensureBranch(tx *gorm.DB, branch string) error {
	if !internal.IsValidBranch(branch) {
		return fmt.Errorf("invalid branch: %s", branch)
	}

	var activeBranch sql.NullString
	err := tx.Raw("SELECT active_branch()").Scan(&activeBranch).Error
	if err != nil {
		return fmt.Errorf("failed to get active branch: %w", err)
	}

	if activeBranch.Valid && activeBranch.String == branch {
		return nil
	}

	var count int64
	err = tx.Raw("SELECT COUNT(*) FROM dolt_branches WHERE name = ?", branch).Scan(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if count == 0 {
		err = tx.Exec("CALL DOLT_BRANCH(?, ?)", branch, internal.MainBranch).Error
		if err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branch, err)
		}
	}

	err = tx.Exec("CALL DOLT_CHECKOUT(?)", branch).Error
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w", branch, err)
	}

	return nil
}

func commitChanges(tx *gorm.DB, message string) error {
	var count int64
	err := tx.Raw("SELECT COUNT(*) FROM dolt_diff WHERE commit_hash = 'WORKING'").Scan(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check for changes: %w", err)
	}

	if count == 0 {
		return nil
	}

	err = tx.Exec("CALL DOLT_ADD('.')").Error
	if err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	err = tx.Exec("CALL DOLT_COMMIT('-m', ?)", message).Error
	if err != nil {
		return fmt.Errorf("failed to commit with message '%s': %w", message, err)
	}

	return nil
}

func (tm *GormTransactionManager) HasBranch(ctx context.Context, branch string) (bool, error) {
	tx := getTxOrDb(ctx, tm.db)

	var count int64
	err := tx.Raw("SELECT COUNT(*) FROM dolt_branches WHERE name = ?", branch).Scan(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if branch exists: %w", err)
	}

	return count > 0, nil
}

func (tm *GormTransactionManager) CreateBranch(ctx context.Context, branch string) error {
	tx := getTxOrDb(ctx, tm.db)
	parent := getBranch(ctx)
	if parent == branch {
		return fmt.Errorf("cannot create branch with same name as checked out branch %s", branch)
	}

	err := tx.Exec("CALL DOLT_BRANCH(?, ?)", branch, parent).Error
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %w", branch, err)
	}

	return nil
}

func (tm *GormTransactionManager) MergeBranch(ctx context.Context, branch string) error {
	tx := getTxOrDb(ctx, tm.db)
	parent := getBranch(ctx)
	if parent == branch {
		return fmt.Errorf("cannot merge currently checked out branch %s into itself", branch)
	}

	err := tx.Exec("CALL DOLT_MERGE(?, '--no-ff')", branch).Error
	if err != nil {
		return fmt.Errorf("failed to merge branch %s: %w", branch, err)
	}

	return nil
}

func (tm *GormTransactionManager) RebaseBranch(ctx context.Context, onto string) error {
	tx := getTxOrDb(ctx, tm.db)
	branch := getBranch(ctx)
	if branch == onto {
		return fmt.Errorf("cannot rebase currently checked out branch %s onto itself", branch)
	}

	var count int64
	err := tx.WithContext(ctx).Raw("SELECT COUNT(*) FROM DOLT_LOG(?, '--not', ?)", branch, onto).Scan(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check commits to rebase: %w", err)
	}

	if count == 0 {
		return nil
	}

	err = tx.WithContext(ctx).Exec("SET @@autocommit = 0").Error
	if err != nil {
		return fmt.Errorf("failed to start rebase of %s onto %s: %w", branch, onto, err)
	}

	err = tx.WithContext(ctx).Exec("CALL DOLT_REBASE('-i', ?)", onto).Error
	if err != nil {
		return fmt.Errorf("failed to start rebase of %s onto %s: %w", branch, onto, err)
	}

	err = tx.WithContext(ctx).Exec("CALL DOLT_REBASE('--continue')").Error
	if err != nil {
		err = resolveAllConflicts(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to resolve conflicts during rebase: %w", err)
		}
		err = tx.WithContext(ctx).Exec("CALL DOLT_REBASE('--continue')").Error
		if err != nil {
			return fmt.Errorf("failed to continue rebase of %s onto %s after resolving conflicts: %w", branch, onto, err)
		}
	}

	return nil
}

func (tm *GormTransactionManager) DeleteBranch(ctx context.Context, branch string) error {
	tx := getTxOrDb(ctx, tm.db)
	parent := getBranch(ctx)
	if parent == branch {
		return fmt.Errorf("cannot delete currently checked out branch %s", branch)
	}

	err := tx.Exec("CALL DOLT_BRANCH('--force', '--delete', ?)", branch).Error
	if err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", branch, err)
	}

	return nil
}

func (tm *GormTransactionManager) GetMergeBase(ctx context.Context, source, branch string) (string, error) {
	tx := getTxOrDb(ctx, tm.db)

	var mergeBase string
	err := tx.Raw("SELECT DOLT_MERGE_BASE(?, ?)", branch, source).Scan(&mergeBase).Error
	if err != nil {
		return "", fmt.Errorf("failed to get merge base: %w", err)
	}

	return mergeBase, nil
}

func (tm *GormTransactionManager) GetHead(ctx context.Context) (string, error) {
	tx := getTxOrDb(ctx, tm.db)

	var head string
	err := tx.Raw("SELECT DOLT_HASHOF('HEAD')").Scan(&head).Error
	if err != nil {
		return "", fmt.Errorf("failed to get head: %w", err)
	}

	return head, nil
}

func (tm *GormTransactionManager) GetBranchHead(ctx context.Context, branch string) (string, error) {
	tx := getTxOrDb(ctx, tm.db)

	var head string
	err := tx.Raw("SELECT DOLT_HASHOF(?)", branch).Scan(&head).Error
	if err != nil {
		return "", fmt.Errorf("failed to get head of branch %s: %w", branch, err)
	}

	return head, nil
}

func (tm *GormTransactionManager) HasCommitsAfter(ctx context.Context, branch, commit string) (bool, error) {
	tx := getTxOrDb(ctx, tm.db)

	var count int64
	err := tx.Raw("SELECT COUNT(*) FROM dolt_log(?, '--not', ?)", branch, commit).Scan(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check commits after %s: %w", commit, err)
	}

	return count > 0, nil
}

func resolveAllConflicts(ctx context.Context, tx *gorm.DB) error {
	var tables []string
	err := tx.WithContext(ctx).Raw("SELECT `table` FROM dolt_conflicts").Scan(&tables).Error
	if err != nil {
		return fmt.Errorf("failed to query dolt_conflicts: %w", err)
	}

	for _, table := range tables {
		err := tx.WithContext(ctx).Exec("CALL DOLT_CONFLICTS_RESOLVE('--theirs', ?)", table).Error
		if err != nil {
			return fmt.Errorf("failed to resolve conflicts for table %s: %w", table, err)
		}
	}

	if len(tables) > 0 {
		err := tx.WithContext(ctx).Exec("CALL DOLT_ADD('.')").Error
		if err != nil {
			return fmt.Errorf("failed to stage resolved changes: %w", err)
		}
	}

	return nil
}
