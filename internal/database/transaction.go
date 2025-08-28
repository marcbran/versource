package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type txKey struct{}

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

func withTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func getTxOrDb(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db
}

func (tm *GormTransactionManager) Checkout(ctx context.Context, branch string, fn func(ctx context.Context) error) error {
	err := ensureBranch(tm.db, branch)
	if err != nil {
		return fmt.Errorf("failed to ensure branch %s: %w", branch, err)
	}

	return fn(ctx)
}

func (tm *GormTransactionManager) Do(ctx context.Context, branch, message string, fn func(ctx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := ensureBranch(tx, branch)
		if err != nil {
			return fmt.Errorf("failed to ensure branch %s: %w", branch, err)
		}

		err = fn(withTx(ctx, tx))
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

func ensureBranch(tx *gorm.DB, branch string) error {
	var activeBranch string
	err := tx.Raw("SELECT active_branch()").Scan(&activeBranch).Error
	if err != nil {
		return fmt.Errorf("failed to get active branch: %w", err)
	}

	if activeBranch == branch {
		return nil
	}

	var count int64
	err = tx.Raw("SELECT COUNT(*) FROM dolt_branches WHERE name = ?", branch).Scan(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if count == 0 {
		err = tx.Exec("CALL DOLT_BRANCH(?, ?)", branch, "main").Error
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

func (tm *GormTransactionManager) MergeBranch(ctx context.Context, branch string) error {
	tx := getTxOrDb(ctx, tm.db)

	err := tx.Exec("CALL DOLT_MERGE(?, '--no-ff')", branch).Error
	if err != nil {
		return fmt.Errorf("failed to merge branch %s: %w", branch, err)
	}

	return nil
}

func (tm *GormTransactionManager) DeleteBranch(ctx context.Context, branch string) error {
	tx := getTxOrDb(ctx, tm.db)

	err := tx.Exec("CALL DOLT_BRANCH('--force', '--delete', ?)", branch).Error
	if err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", branch, err)
	}

	return nil
}

func (tm *GormTransactionManager) BranchExists(ctx context.Context, branch string) (bool, error) {
	tx := getTxOrDb(ctx, tm.db)

	var count int64
	err := tx.Raw("SELECT COUNT(*) FROM dolt_branches WHERE name = ?", branch).Scan(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if branch exists: %w", err)
	}

	return count > 0, nil
}
