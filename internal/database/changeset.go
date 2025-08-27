package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/marcbran/versource/internal"
)

type GormChangesetRepo struct {
	db *gorm.DB
}

func NewGormChangesetRepo(db *gorm.DB) *GormChangesetRepo {
	return &GormChangesetRepo{db: db}
}

func (r *GormChangesetRepo) CreateChangeset(ctx context.Context, changeset *internal.Changeset) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(changeset).Error
	if err != nil {
		return fmt.Errorf("failed to create changeset: %w", err)
	}
	return nil
}

func (r *GormChangesetRepo) GetChangeset(ctx context.Context, changesetID uint) (*internal.Changeset, error) {
	db := getTxOrDb(ctx, r.db)
	var changeset internal.Changeset
	err := db.WithContext(ctx).Where("id = ?", changesetID).First(&changeset).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get changeset: %w", err)
	}
	return &changeset, nil
}

func (r *GormChangesetRepo) GetChangesetByName(ctx context.Context, name string) (*internal.Changeset, error) {
	db := getTxOrDb(ctx, r.db)
	var changeset internal.Changeset
	err := db.WithContext(ctx).Where("name = ?", name).First(&changeset).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get changeset by name: %w", err)
	}
	return &changeset, nil
}

func (r *GormChangesetRepo) UpdateChangesetState(ctx context.Context, changesetID uint, state internal.ChangesetState) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&internal.Changeset{}).Where("id = ?", changesetID).Update("state", state).Error
	if err != nil {
		return fmt.Errorf("failed to update changeset state: %w", err)
	}
	return nil
}

func (r *GormChangesetRepo) ListChangesets(ctx context.Context) ([]internal.Changeset, error) {
	db := getTxOrDb(ctx, r.db)
	var changesets []internal.Changeset
	err := db.WithContext(ctx).Find(&changesets).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list changesets: %w", err)
	}
	return changesets, nil
}
