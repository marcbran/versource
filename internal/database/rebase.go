package database

import (
	"context"
	"fmt"

	"github.com/marcbran/versource/pkg/versource"
	"gorm.io/gorm"
)

type GormRebaseRepo struct {
	db *gorm.DB
}

func NewGormRebaseRepo(db *gorm.DB) *GormRebaseRepo {
	return &GormRebaseRepo{db: db}
}

func (r *GormRebaseRepo) GetRebase(ctx context.Context, rebaseID uint) (*versource.Rebase, error) {
	db := getTxOrDb(ctx, r.db)
	var rebase versource.Rebase
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
	var rebases []versource.Rebase
	err := db.WithContext(ctx).Where("state = ?", versource.TaskStateQueued).Find(&rebases).Error
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
	var rebases []versource.Rebase
	err := db.WithContext(ctx).Where("state = ? AND changeset_id = ?", versource.TaskStateQueued, changesetID).Find(&rebases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queued rebases by changeset: %w", err)
	}

	rebaseIDs := make([]uint, len(rebases))
	for i, rebase := range rebases {
		rebaseIDs[i] = rebase.ID
	}
	return rebaseIDs, nil
}

func (r *GormRebaseRepo) ListRebases(ctx context.Context) ([]versource.Rebase, error) {
	db := getTxOrDb(ctx, r.db)
	var rebases []versource.Rebase
	err := db.WithContext(ctx).
		Preload("Changeset").
		Order("id DESC").
		Find(&rebases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list rebases: %w", err)
	}
	return rebases, nil
}

func (r *GormRebaseRepo) ListRebasesByChangesetName(ctx context.Context, changesetName string) ([]versource.Rebase, error) {
	db := getTxOrDb(ctx, r.db)
	var rebases []versource.Rebase
	err := db.WithContext(ctx).
		Preload("Changeset").
		Joins("JOIN changesets ON rebases.changeset_id = changesets.id").
		Where("changesets.name = ?", changesetName).
		Order("rebases.id DESC").
		Find(&rebases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list rebases by changeset name: %w", err)
	}
	return rebases, nil
}

func (r *GormRebaseRepo) CreateRebase(ctx context.Context, rebase *versource.Rebase) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(rebase).Error
	if err != nil {
		return fmt.Errorf("failed to create rebase: %w", err)
	}
	return nil
}

func (r *GormRebaseRepo) UpdateRebaseState(ctx context.Context, rebaseID uint, state versource.TaskState) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&versource.Rebase{}).Where("id = ?", rebaseID).Update("state", state).Error
	if err != nil {
		return fmt.Errorf("failed to update rebase state: %w", err)
	}
	return nil
}
