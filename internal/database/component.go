package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/marcbran/versource/internal"
)

type GormComponentRepo struct {
	db *gorm.DB
}

func NewGormComponentRepo(db *gorm.DB) *GormComponentRepo {
	return &GormComponentRepo{db: db}
}

func (r *GormComponentRepo) GetComponent(ctx context.Context, componentID uint) (*internal.Component, error) {
	db := getTxOrDb(ctx, r.db)
	var component internal.Component
	err := db.WithContext(ctx).Where("id = ?", componentID).First(&component).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get component: %w", err)
	}
	return &component, nil
}

func (r *GormComponentRepo) ListComponents(ctx context.Context) ([]internal.Component, error) {
	db := getTxOrDb(ctx, r.db)
	var components []internal.Component
	err := db.WithContext(ctx).Preload("ModuleVersion.Module").Find(&components).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list components: %w", err)
	}
	return components, nil
}

func (r *GormComponentRepo) ListComponentsByModule(ctx context.Context, moduleID uint) ([]internal.Component, error) {
	db := getTxOrDb(ctx, r.db)
	var components []internal.Component
	err := db.WithContext(ctx).
		Preload("ModuleVersion.Module").
		Joins("JOIN module_versions ON components.module_version_id = module_versions.id").
		Where("module_versions.module_id = ?", moduleID).
		Find(&components).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list components by module: %w", err)
	}
	return components, nil
}

func (r *GormComponentRepo) ListComponentsByModuleVersion(ctx context.Context, moduleVersionID uint) ([]internal.Component, error) {
	db := getTxOrDb(ctx, r.db)
	var components []internal.Component
	err := db.WithContext(ctx).
		Preload("ModuleVersion.Module").
		Where("module_version_id = ?", moduleVersionID).
		Find(&components).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list components by module version: %w", err)
	}
	return components, nil
}

func (r *GormComponentRepo) CreateComponent(ctx context.Context, component *internal.Component) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(component).Error
	if err != nil {
		return fmt.Errorf("failed to create component: %w", err)
	}
	return nil
}

func (r *GormComponentRepo) UpdateComponent(ctx context.Context, component *internal.Component) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Save(component).Error
	if err != nil {
		return fmt.Errorf("failed to update component: %w", err)
	}
	return nil
}
