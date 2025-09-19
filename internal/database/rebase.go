package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/marcbran/versource/internal"
)

type GormRebaseRepo struct {
	db *gorm.DB
}

func NewGormRebaseRepo(db *gorm.DB) *GormRebaseRepo {
	return &GormRebaseRepo{db: db}
}

func (r *GormRebaseRepo) GetRebase(ctx context.Context, rebaseID uint) (*internal.Rebase, error) {
	db := getTxOrDb(ctx, r.db)
	var rebase internal.Rebase
	err := db.WithContext(ctx).
		Preload("Changeset").
		Where("id = ?", rebaseID).First(&rebase).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get rebase: %w", err)
	}
	return &rebase, nil
}

func (r *GormRebaseRepo) GetQueuedRebases(ctx context.Context) ([]uint, error) {
	db := getTxOrDb(ctx, r.db)
	var rebases []internal.Rebase
	err := db.WithContext(ctx).Where("state = ?", internal.TaskStateQueued).Find(&rebases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queued rebases: %w", err)
	}

	rebaseIDs := make([]uint, len(rebases))
	for i, rebase := range rebases {
		rebaseIDs[i] = rebase.ID
	}
	return rebaseIDs, nil
}

func (r *GormRebaseRepo) GetQueuedRebasesByChangeset(ctx context.Context, changesetID uint) ([]uint, error) {
	db := getTxOrDb(ctx, r.db)
	var rebases []internal.Rebase
	err := db.WithContext(ctx).Where("state = ? AND changeset_id = ?", internal.TaskStateQueued, changesetID).Find(&rebases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queued rebases by changeset: %w", err)
	}

	rebaseIDs := make([]uint, len(rebases))
	for i, rebase := range rebases {
		rebaseIDs[i] = rebase.ID
	}
	return rebaseIDs, nil
}

func (r *GormRebaseRepo) ListRebases(ctx context.Context) ([]internal.Rebase, error) {
	db := getTxOrDb(ctx, r.db)
	var rebases []internal.Rebase
	err := db.WithContext(ctx).
		Preload("Changeset").
		Order("id DESC").
		Find(&rebases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list rebases: %w", err)
	}
	return rebases, nil
}

func (r *GormRebaseRepo) CreateRebase(ctx context.Context, rebase *internal.Rebase) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(rebase).Error
	if err != nil {
		return fmt.Errorf("failed to create rebase: %w", err)
	}
	return nil
}

func (r *GormRebaseRepo) UpdateRebaseState(ctx context.Context, rebaseID uint, state internal.TaskState) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&internal.Rebase{}).Where("id = ?", rebaseID).Update("state", state).Error
	if err != nil {
		return fmt.Errorf("failed to update rebase state: %w", err)
	}
	return nil
}
