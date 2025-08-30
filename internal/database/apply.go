package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/marcbran/versource/internal"
)

type GormApplyRepo struct {
	db *gorm.DB
}

func NewGormApplyRepo(db *gorm.DB) *GormApplyRepo {
	return &GormApplyRepo{db: db}
}

func (r *GormApplyRepo) GetApply(ctx context.Context, applyID uint) (*internal.Apply, error) {
	db := getTxOrDb(ctx, r.db)
	var apply internal.Apply
	err := db.WithContext(ctx).
		Preload("Plan.Component.ModuleVersion.Module").
		Preload("Plan.Component.ModuleVersion").
		Preload("Plan.Component").
		Preload("Plan.Changeset").
		Preload("Changeset").
		Where("id = ?", applyID).First(&apply).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get apply: %w", err)
	}
	return &apply, nil
}

func (r *GormApplyRepo) GetQueuedApplies(ctx context.Context) ([]uint, error) {
	db := getTxOrDb(ctx, r.db)
	var applies []internal.Apply
	err := db.WithContext(ctx).Where("state = ?", internal.TaskStateQueued).Find(&applies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queued applies: %w", err)
	}

	applyIDs := make([]uint, len(applies))
	for i, apply := range applies {
		applyIDs[i] = apply.ID
	}
	return applyIDs, nil
}

func (r *GormApplyRepo) GetQueuedAppliesByChangeset(ctx context.Context, changesetID uint) ([]uint, error) {
	db := getTxOrDb(ctx, r.db)
	var applies []internal.Apply
	err := db.WithContext(ctx).Where("state = ? AND changeset_id = ?", internal.TaskStateQueued, changesetID).Find(&applies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queued applies for changeset: %w", err)
	}

	applyIDs := make([]uint, len(applies))
	for i, apply := range applies {
		applyIDs[i] = apply.ID
	}
	return applyIDs, nil
}

func (r *GormApplyRepo) ListApplies(ctx context.Context) ([]internal.Apply, error) {
	db := getTxOrDb(ctx, r.db)
	var applies []internal.Apply
	err := db.WithContext(ctx).
		Preload("Plan.Component.ModuleVersion.Module").
		Preload("Plan.Component.ModuleVersion").
		Preload("Plan.Changeset").
		Preload("Changeset").
		Find(&applies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list applies: %w", err)
	}
	return applies, nil
}

func (r *GormApplyRepo) CreateApply(ctx context.Context, apply *internal.Apply) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(apply).Error
	if err != nil {
		return fmt.Errorf("failed to create apply: %w", err)
	}
	return nil
}

func (r *GormApplyRepo) UpdateApplyState(ctx context.Context, applyID uint, state internal.TaskState) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&internal.Apply{}).Where("id = ?", applyID).Update("state", state).Error
	if err != nil {
		return fmt.Errorf("failed to update apply state: %w", err)
	}
	return nil
}
