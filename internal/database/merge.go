package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/marcbran/versource/internal"
)

type GormMergeRepo struct {
	db *gorm.DB
}

func NewGormMergeRepo(db *gorm.DB) *GormMergeRepo {
	return &GormMergeRepo{db: db}
}

func (r *GormMergeRepo) GetMerge(ctx context.Context, mergeID uint) (*internal.Merge, error) {
	db := getTxOrDb(ctx, r.db)
	var merge internal.Merge
	err := db.WithContext(ctx).
		Preload("Changeset").
		Where("id = ?", mergeID).First(&merge).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get merge: %w", err)
	}
	return &merge, nil
}

func (r *GormMergeRepo) GetQueuedMerges(ctx context.Context) ([]uint, error) {
	db := getTxOrDb(ctx, r.db)
	var merges []internal.Merge
	err := db.WithContext(ctx).Where("state = ?", internal.TaskStateQueued).Find(&merges).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queued merges: %w", err)
	}

	mergeIDs := make([]uint, len(merges))
	for i, merge := range merges {
		mergeIDs[i] = merge.ID
	}
	return mergeIDs, nil
}

func (r *GormMergeRepo) GetQueuedMergesByChangeset(ctx context.Context, changesetID uint) ([]uint, error) {
	db := getTxOrDb(ctx, r.db)
	var merges []internal.Merge
	err := db.WithContext(ctx).Where("state = ? AND changeset_id = ?", internal.TaskStateQueued, changesetID).Find(&merges).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queued merges by changeset: %w", err)
	}

	mergeIDs := make([]uint, len(merges))
	for i, merge := range merges {
		mergeIDs[i] = merge.ID
	}
	return mergeIDs, nil
}

func (r *GormMergeRepo) ListMerges(ctx context.Context) ([]internal.Merge, error) {
	db := getTxOrDb(ctx, r.db)
	var merges []internal.Merge
	err := db.WithContext(ctx).
		Preload("Changeset").
		Order("id DESC").
		Find(&merges).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list merges: %w", err)
	}
	return merges, nil
}

func (r *GormMergeRepo) ListMergesByChangesetName(ctx context.Context, changesetName string) ([]internal.Merge, error) {
	db := getTxOrDb(ctx, r.db)
	var merges []internal.Merge
	err := db.WithContext(ctx).
		Preload("Changeset").
		Joins("JOIN changesets ON merges.changeset_id = changesets.id").
		Where("changesets.name = ?", changesetName).
		Order("merges.id DESC").
		Find(&merges).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list merges by changeset name: %w", err)
	}
	return merges, nil
}

func (r *GormMergeRepo) CreateMerge(ctx context.Context, merge *internal.Merge) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(merge).Error
	if err != nil {
		return fmt.Errorf("failed to create merge: %w", err)
	}
	return nil
}

func (r *GormMergeRepo) UpdateMergeState(ctx context.Context, mergeID uint, state internal.TaskState) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&internal.Merge{}).Where("id = ?", mergeID).Update("state", state).Error
	if err != nil {
		return fmt.Errorf("failed to update merge state: %w", err)
	}
	return nil
}
