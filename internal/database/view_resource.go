package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/marcbran/versource/internal"
)

type GormViewResourceRepo struct {
	db *gorm.DB
}

func NewGormViewResourceRepo(db *gorm.DB) *GormViewResourceRepo {
	return &GormViewResourceRepo{db: db}
}

func (r *GormViewResourceRepo) GetViewResource(ctx context.Context, viewResourceID uint) (*internal.ViewResource, error) {
	db := getTxOrDb(ctx, r.db)
	var viewResource internal.ViewResource
	err := db.WithContext(ctx).Where("id = ?", viewResourceID).First(&viewResource).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get view resource: %w", err)
	}
	return &viewResource, nil
}

func (r *GormViewResourceRepo) GetViewResourceByName(ctx context.Context, name string) (*internal.ViewResource, error) {
	db := getTxOrDb(ctx, r.db)
	var viewResource internal.ViewResource
	err := db.WithContext(ctx).Where("name = ?", name).First(&viewResource).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get view resource by name: %w", err)
	}
	return &viewResource, nil
}

func (r *GormViewResourceRepo) ListViewResources(ctx context.Context) ([]internal.ViewResource, error) {
	db := getTxOrDb(ctx, r.db)
	var viewResources []internal.ViewResource
	err := db.WithContext(ctx).Find(&viewResources).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list view resources: %w", err)
	}
	return viewResources, nil
}

func (r *GormViewResourceRepo) CreateViewResource(ctx context.Context, viewResource *internal.ViewResource) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(viewResource).Error
	if err != nil {
		return fmt.Errorf("failed to create view resource: %w", err)
	}
	return nil
}

func (r *GormViewResourceRepo) UpdateViewResource(ctx context.Context, viewResource *internal.ViewResource) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&internal.ViewResource{}).Where("id = ?", viewResource.ID).Updates(viewResource).Error
	if err != nil {
		return fmt.Errorf("failed to update view resource: %w", err)
	}
	return nil
}

func (r *GormViewResourceRepo) DeleteViewResource(ctx context.Context, viewResourceID uint) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Where("id = ?", viewResourceID).Delete(&internal.ViewResource{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete view resource: %w", err)
	}
	return nil
}
