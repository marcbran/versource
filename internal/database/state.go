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
