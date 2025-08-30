package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/marcbran/versource/internal"
)

type GormModuleRepo struct {
	db *gorm.DB
}

func NewGormModuleRepo(db *gorm.DB) *GormModuleRepo {
	return &GormModuleRepo{db: db}
}

func (r *GormModuleRepo) CreateModule(ctx context.Context, module *internal.Module) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(module).Error
	if err != nil {
		return fmt.Errorf("failed to create module: %w", err)
	}
	return nil
}

func (r *GormModuleRepo) GetModule(ctx context.Context, moduleID uint) (*internal.Module, error) {
	db := getTxOrDb(ctx, r.db)
	var module internal.Module
	err := db.WithContext(ctx).Where("id = ?", moduleID).First(&module).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %w", err)
	}
	return &module, nil
}

func (r *GormModuleRepo) GetModuleBySource(ctx context.Context, source string) (*internal.Module, error) {
	db := getTxOrDb(ctx, r.db)
	var module internal.Module
	err := db.WithContext(ctx).Where("source = ?", source).First(&module).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get module by source: %w", err)
	}
	return &module, nil
}

func (r *GormModuleRepo) ListModules(ctx context.Context) ([]internal.Module, error) {
	db := getTxOrDb(ctx, r.db)
	var modules []internal.Module
	err := db.WithContext(ctx).Find(&modules).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}
	return modules, nil
}

type GormModuleVersionRepo struct {
	db *gorm.DB
}

func NewGormModuleVersionRepo(db *gorm.DB) *GormModuleVersionRepo {
	return &GormModuleVersionRepo{db: db}
}

func (r *GormModuleVersionRepo) CreateModuleVersion(ctx context.Context, moduleVersion *internal.ModuleVersion) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(moduleVersion).Error
	if err != nil {
		return fmt.Errorf("failed to create module version: %w", err)
	}
	return nil
}

func (r *GormModuleVersionRepo) GetModuleVersion(ctx context.Context, moduleVersionID uint) (*internal.ModuleVersion, error) {
	db := getTxOrDb(ctx, r.db)
	var moduleVersion internal.ModuleVersion
	err := db.WithContext(ctx).Preload("Module").Where("id = ?", moduleVersionID).First(&moduleVersion).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get module version: %w", err)
	}
	return &moduleVersion, nil
}

func (r *GormModuleVersionRepo) ListModuleVersionsForModule(ctx context.Context, moduleID uint) ([]internal.ModuleVersion, error) {
	db := getTxOrDb(ctx, r.db)
	var moduleVersions []internal.ModuleVersion
	err := db.WithContext(ctx).Preload("Module").Where("module_id = ?", moduleID).Find(&moduleVersions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list module versions for module: %w", err)
	}
	return moduleVersions, nil
}

func (r *GormModuleVersionRepo) GetLatestModuleVersion(ctx context.Context, moduleID uint) (*internal.ModuleVersion, error) {
	db := getTxOrDb(ctx, r.db)
	var moduleVersion internal.ModuleVersion
	err := db.WithContext(ctx).Preload("Module").Where("module_id = ?", moduleID).Order("id DESC").First(&moduleVersion).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest module version: %w", err)
	}
	return &moduleVersion, nil
}

func (r *GormModuleVersionRepo) ListModuleVersions(ctx context.Context) ([]internal.ModuleVersion, error) {
	db := getTxOrDb(ctx, r.db)
	var moduleVersions []internal.ModuleVersion
	err := db.WithContext(ctx).Preload("Module").Find(&moduleVersions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list module versions: %w", err)
	}
	return moduleVersions, nil
}
