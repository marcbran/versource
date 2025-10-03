package database

import (
	"context"
	"errors"
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

func (r *GormResourceRepo) UpsertResources(ctx context.Context, resources []internal.Resource) error {
	db := getTxOrDb(ctx, r.db)

	for _, resource := range resources {
		var existingResource internal.Resource
		err := db.WithContext(ctx).Where("uuid = ?", resource.UUID).First(&existingResource).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get existing resource: %w", err)
		}

		if existingResource.UUID == "" {
			err := db.WithContext(ctx).Create(&resource).Error
			if err != nil {
				return fmt.Errorf("failed to create resource: %w", err)
			}
		} else {
			err := db.WithContext(ctx).Model(&existingResource).Updates(&resource).Error
			if err != nil {
				return fmt.Errorf("failed to update existing resource: %w", err)
			}
		}
	}

	return nil
}
