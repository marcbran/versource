package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/marcbran/versource/pkg/versource"
	"gorm.io/gorm"
)

type GormStateRepo struct {
	db *gorm.DB
}

func NewGormStateRepo(db *gorm.DB) *GormStateRepo {
	return &GormStateRepo{db: db}
}

func (r *GormStateRepo) UpsertState(ctx context.Context, state *versource.State) error {
	db := getTxOrDb(ctx, r.db)
	var existingState versource.State
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

func (r *GormStateResourceRepo) ListStateResourcesByStateID(ctx context.Context, stateID uint) ([]versource.StateResource, error) {
	db := getTxOrDb(ctx, r.db)
	var resources []versource.StateResource
	err := db.WithContext(ctx).Where("state_id = ?", stateID).Find(&resources).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list state resources: %w", err)
	}
	return resources, nil
}

func (r *GormStateResourceRepo) InsertStateResources(ctx context.Context, resources []versource.StateResource) error {
	if len(resources) == 0 {
		return nil
	}

	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(&resources).Error
	if err != nil {
		return fmt.Errorf("failed to insert state resources: %w", err)
	}
	return nil
}

func (r *GormStateResourceRepo) UpdateStateResources(ctx context.Context, resources []versource.StateResource) error {
	if len(resources) == 0 {
		return nil
	}

	db := getTxOrDb(ctx, r.db)
	for _, resource := range resources {
		err := db.WithContext(ctx).Model(&versource.StateResource{}).Where("id = ?", resource.ID).Updates(&resource).Error
		if err != nil {
			return fmt.Errorf("failed to update state resource %d: %w", resource.ID, err)
		}
	}
	return nil
}

func (r *GormStateResourceRepo) DeleteStateResources(ctx context.Context, stateResourceIDs []uint) error {
	if len(stateResourceIDs) == 0 {
		return nil
	}

	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Where("id IN ?", stateResourceIDs).Delete(&versource.StateResource{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete state resources: %w", err)
	}
	return nil
}
