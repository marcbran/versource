package database

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/marcbran/versource/internal"
)

type GormStateRepo struct {
	db *gorm.DB
}

func NewGormStateRepo(db *gorm.DB) *GormStateRepo {
	return &GormStateRepo{db: db}
}

func (r *GormStateRepo) UpsertState(ctx context.Context, state *internal.State) error {
	db := getTxOrDb(ctx, r.db)
	var existingState internal.State
	err := db.WithContext(ctx).Where("component_id = ?", state.ComponentID).First(&existingState).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get existing component output: %w", err)
	}
	if existingState.ID == 0 {
		err := db.WithContext(ctx).Create(state).Error
		if err != nil {
			return fmt.Errorf("failed to create state: %w", err)
		}
	} else {
		state.ID = existingState.ID
		err := db.WithContext(ctx).Model(&existingState).Updates(state).Error
		if err != nil {
			return fmt.Errorf("failed to update existing component output: %w", err)
		}
	}
	return nil
}

type GormStateResourceRepo struct {
	db *gorm.DB
}

func NewGormStateResourceRepo(db *gorm.DB) *GormStateResourceRepo {
	return &GormStateResourceRepo{db: db}
}

func (r *GormStateResourceRepo) UpsertStateResources(ctx context.Context, resources []internal.StateResource) error {
	db := getTxOrDb(ctx, r.db)

	for _, resource := range resources {
		var existingResource internal.StateResource
		err := db.WithContext(ctx).Where("state_id = ? AND address = ?", resource.StateID, resource.Address).First(&existingResource).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get existing state resource: %w", err)
		}

		if existingResource.ID == 0 {
			err := db.WithContext(ctx).Create(&resource).Error
			if err != nil {
				return fmt.Errorf("failed to create state resource: %w", err)
			}
		} else {
			resource.ID = existingResource.ID
			err := db.WithContext(ctx).Model(&existingResource).Updates(&resource).Error
			if err != nil {
				return fmt.Errorf("failed to update existing state resource: %w", err)
			}
		}
	}

	return nil
}
