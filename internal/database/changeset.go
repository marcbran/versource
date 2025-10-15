package database

import (
	"context"
	"fmt"

	"github.com/marcbran/versource/pkg/versource"
	"gorm.io/gorm"
)

type GormChangesetRepo struct {
	db *gorm.DB
}

func NewGormChangesetRepo(db *gorm.DB) *GormChangesetRepo {
	return &GormChangesetRepo{db: db}
}

func (r *GormChangesetRepo) GetChangeset(ctx context.Context, changesetID uint) (*versource.Changeset, error) {
	db := getTxOrDb(ctx, r.db)
	var changeset versource.Changeset
	err := db.WithContext(ctx).Where("id = ?", changesetID).First(&changeset).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get changeset: %w", err)
	}
	return &changeset, nil
}

func (r *GormChangesetRepo) GetChangesetByName(ctx context.Context, name string) (*versource.Changeset, error) {
	db := getTxOrDb(ctx, r.db)
	var changeset versource.Changeset
	err := db.WithContext(ctx).Where("name = ?", name).First(&changeset).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get changeset by name: %w", err)
	}
	return &changeset, nil
}

func (r *GormChangesetRepo) GetOpenChangesetByName(ctx context.Context, name string) (*versource.Changeset, error) {
	db := getTxOrDb(ctx, r.db)
	var changeset versource.Changeset
	err := db.WithContext(ctx).Where("state = ? AND name = ?", versource.ChangesetStateOpen, name).First(&changeset).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get open changeset by name: %w", err)
	}
	return &changeset, nil
}

func (r *GormChangesetRepo) ListChangesets(ctx context.Context) ([]versource.Changeset, error) {
	db := getTxOrDb(ctx, r.db)
	var changesets []versource.Changeset
	err := db.WithContext(ctx).Find(&changesets).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list changesets: %w", err)
	}
	return changesets, nil
}

func (r *GormChangesetRepo) HasOpenChangesetWithName(ctx context.Context, name string) (bool, error) {
	db := getTxOrDb(ctx, r.db)
	var count int64
	err := db.WithContext(ctx).Model(&versource.Changeset{}).Where("state = ? AND name = ?", versource.ChangesetStateOpen, name).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check for open changesets: %w", err)
	}
	return count > 0, nil
}

func (r *GormChangesetRepo) HasChangesetWithName(ctx context.Context, name string) (bool, error) {
	db := getTxOrDb(ctx, r.db)
	var count int64
	err := db.WithContext(ctx).Model(&versource.Changeset{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check for changesets: %w", err)
	}
	return count > 0, nil
}

func (r *GormChangesetRepo) CreateChangeset(ctx context.Context, changeset *versource.Changeset) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(changeset).Error
	if err != nil {
		return fmt.Errorf("failed to create changeset: %w", err)
	}
	return nil
}

func (r *GormChangesetRepo) UpdateChangesetState(ctx context.Context, changesetID uint, state versource.ChangesetState) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&versource.Changeset{}).Where("id = ?", changesetID).Update("state", state).Error
	if err != nil {
		return fmt.Errorf("failed to update changeset state: %w", err)
	}
	return nil
}

func (r *GormChangesetRepo) UpdateChangesetReviewState(ctx context.Context, changesetID uint, reviewState versource.ChangesetReviewState) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&versource.Changeset{}).Where("id = ?", changesetID).Update("review_state", reviewState).Error
	if err != nil {
		return fmt.Errorf("failed to update changeset review state: %w", err)
	}
	return nil
}

func (r *GormChangesetRepo) DeleteChangeset(ctx context.Context, changesetID uint) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Delete(&versource.Changeset{}, changesetID).Error
	if err != nil {
		return fmt.Errorf("failed to delete changeset: %w", err)
	}
	return nil
}
