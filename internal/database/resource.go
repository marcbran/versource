package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/marcbran/versource/internal"
)

type GormResourceRepo struct {
	db *gorm.DB
}

func NewGormResourceRepo(db *gorm.DB) *GormResourceRepo {
	return &GormResourceRepo{db: db}
}

func (r *GormResourceRepo) InsertResources(ctx context.Context, resources []internal.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(&resources).Error
	if err != nil {
		return fmt.Errorf("failed to insert resources: %w", err)
	}
	return nil
}

func (r *GormResourceRepo) UpdateResources(ctx context.Context, resources []internal.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	db := getTxOrDb(ctx, r.db)
	for _, resource := range resources {
		err := db.WithContext(ctx).Model(&internal.Resource{}).Where("uuid = ?", resource.UUID).Updates(&resource).Error
		if err != nil {
			return fmt.Errorf("failed to update resource %s: %w", resource.UUID, err)
		}
	}
	return nil
}

func (r *GormResourceRepo) DeleteResources(ctx context.Context, resourceUUIDs []string) error {
	if len(resourceUUIDs) == 0 {
		return nil
	}

	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Where("uuid IN ?", resourceUUIDs).Delete(&internal.Resource{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete resources: %w", err)
	}
	return nil
}

func (r *GormResourceRepo) ListResources(ctx context.Context) ([]internal.Resource, error) {
	db := getTxOrDb(ctx, r.db)
	var resources []internal.Resource
	err := db.WithContext(ctx).Find(&resources).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}
	return resources, nil
}
