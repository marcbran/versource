package database

import (
	"context"
	"fmt"

	"github.com/marcbran/versource/pkg/versource"
	"gorm.io/gorm"
)

type GormModuleRepo struct {
	db *gorm.DB
}

func NewGormModuleRepo(db *gorm.DB) *GormModuleRepo {
	return &GormModuleRepo{db: db}
}

func (r *GormModuleRepo) GetModule(ctx context.Context, moduleID uint) (*versource.Module, error) {
	db := getTxOrDb(ctx, r.db)
	var module versource.Module
	err := db.WithContext(ctx).Where("id = ?", moduleID).First(&module).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %w", err)
	}
	return &module, nil
}

func (r *GormModuleRepo) GetModuleByName(ctx context.Context, name string) (*versource.Module, error) {
	db := getTxOrDb(ctx, r.db)
	var module versource.Module
	err := db.WithContext(ctx).Where("name = ?", name).First(&module).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get module by name: %w", err)
	}
	return &module, nil
}

func (r *GormModuleRepo) GetModuleBySource(ctx context.Context, source string) (*versource.Module, error) {
	db := getTxOrDb(ctx, r.db)
	var module versource.Module
	err := db.WithContext(ctx).Where("source = ?", source).First(&module).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get module by source: %w", err)
	}
	return &module, nil
}

func (r *GormModuleRepo) ListModules(ctx context.Context) ([]versource.Module, error) {
	db := getTxOrDb(ctx, r.db)
	var modules []versource.Module
	err := db.WithContext(ctx).Find(&modules).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}
	return modules, nil
}

func (r *GormModuleRepo) CreateModule(ctx context.Context, module *versource.Module) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(module).Error
	if err != nil {
		return fmt.Errorf("failed to create module: %w", err)
	}
	return nil
}

func (r *GormModuleRepo) DeleteModule(ctx context.Context, moduleID uint) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Where("id = ?", moduleID).Delete(&versource.Module{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete module: %w", err)
	}
	return nil
}

type GormModuleVersionRepo struct {
	db *gorm.DB
}

func NewGormModuleVersionRepo(db *gorm.DB) *GormModuleVersionRepo {
	return &GormModuleVersionRepo{db: db}
}

func (r *GormModuleVersionRepo) GetModuleVersion(ctx context.Context, moduleVersionID uint) (*versource.ModuleVersion, error) {
	db := getTxOrDb(ctx, r.db)
	var moduleVersion versource.ModuleVersion
	err := db.WithContext(ctx).Preload("Module").Where("id = ?", moduleVersionID).First(&moduleVersion).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get module version: %w", err)
	}
	return &moduleVersion, nil
}

func (r *GormModuleVersionRepo) GetLatestModuleVersion(ctx context.Context, moduleID uint) (*versource.ModuleVersion, error) {
	db := getTxOrDb(ctx, r.db)
	var moduleVersion versource.ModuleVersion
	err := db.WithContext(ctx).Preload("Module").Where("module_id = ?", moduleID).Order("id DESC").First(&moduleVersion).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest module version: %w", err)
	}
	return &moduleVersion, nil
}

func (r *GormModuleVersionRepo) ListModuleVersions(ctx context.Context) ([]versource.ModuleVersion, error) {
	db := getTxOrDb(ctx, r.db)
	var moduleVersions []versource.ModuleVersion
	err := db.WithContext(ctx).Preload("Module").Find(&moduleVersions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list module versions: %w", err)
	}
	return moduleVersions, nil
}

func (r *GormModuleVersionRepo) ListModuleVersionsForModule(ctx context.Context, moduleID uint) ([]versource.ModuleVersion, error) {
	db := getTxOrDb(ctx, r.db)
	var moduleVersions []versource.ModuleVersion
	err := db.WithContext(ctx).Preload("Module").Where("module_id = ?", moduleID).Find(&moduleVersions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list module versions for module: %w", err)
	}
	return moduleVersions, nil
}

func (r *GormModuleVersionRepo) CreateModuleVersion(ctx context.Context, moduleVersion *versource.ModuleVersion) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(moduleVersion).Error
	if err != nil {
		return fmt.Errorf("failed to create module version: %w", err)
	}
	return nil
}
