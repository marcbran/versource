package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/marcbran/versource/internal"
	"gorm.io/datatypes"
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
	err := db.WithContext(ctx).Preload("ModuleVersion.Module").Where("id = ?", componentID).First(&component).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get component: %w", err)
	}
	return &component, nil
}

func (r *GormComponentRepo) GetComponentAtCommit(ctx context.Context, componentID uint, commit string) (*internal.Component, error) {
	db := getTxOrDb(ctx, r.db)
	var component internal.Component
	query := fmt.Sprintf("SELECT * FROM components AS OF '%s' WHERE id = ?", commit)
	err := db.WithContext(ctx).Raw(query, componentID).Scan(&component).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get component at commit: %w", err)
	}

	var moduleVersion internal.ModuleVersion
	moduleVersionQuery := fmt.Sprintf("SELECT * FROM module_versions AS OF '%s' WHERE id = ?", commit)
	err = db.WithContext(ctx).Raw(moduleVersionQuery, component.ModuleVersionID).Scan(&moduleVersion).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get module version at commit: %w", err)
	}

	var module internal.Module
	moduleQuery := fmt.Sprintf("SELECT * FROM modules AS OF '%s' WHERE id = ?", commit)
	err = db.WithContext(ctx).Raw(moduleQuery, moduleVersion.ModuleID).Scan(&module).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get module at commit: %w", err)
	}

	moduleVersion.Module = module
	component.ModuleVersion = moduleVersion

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

type GormComponentDiffRepo struct {
	db *gorm.DB
}

func NewGormComponentDiffRepo(db *gorm.DB) *GormComponentDiffRepo {
	return &GormComponentDiffRepo{db: db}
}

func (r *GormComponentDiffRepo) ListComponentDiffs(ctx context.Context, fromCommit, toCommit string) ([]internal.ComponentDiff, error) {
	db := getTxOrDb(ctx, r.db)

	query := fmt.Sprintf(`
		SELECT
			to_id,
			to_module_version_id,
			to_name,
			to_variables,
			to_commit,
			to_commit_date,
			from_id,
			from_module_version_id,
			from_name,
			from_variables,
			from_commit,
			from_commit_date,
			diff_type
		FROM dolt_diff('%s', '%s', 'components')
	`, fromCommit, toCommit)

	type rawDiff struct {
		ToID                *uint          `json:"to_id"`
		ToModuleVersionID   *uint          `json:"to_module_version_id"`
		ToName              *string        `json:"to_name"`
		ToVariables         datatypes.JSON `json:"to_variables"`
		ToCommit            string         `json:"to_commit"`
		ToCommitDate        string         `json:"to_commit_date"`
		FromID              *uint          `json:"from_id"`
		FromModuleVersionID *uint          `json:"from_module_version_id"`
		FromName            *string        `json:"from_name"`
		FromVariables       datatypes.JSON `json:"from_variables"`
		FromCommit          string         `json:"from_commit"`
		FromCommitDate      string         `json:"from_commit_date"`
		DiffType            string         `json:"diff_type"`
	}

	var rawDiffs []rawDiff
	err := db.WithContext(ctx).Raw(query).Scan(&rawDiffs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list component diffs: %w", err)
	}

	diffs := make([]internal.ComponentDiff, len(rawDiffs))
	for i, raw := range rawDiffs {
		var fromComponent, toComponent internal.Component

		if raw.FromID != nil {
			fromComponent.ID = *raw.FromID
		}
		if raw.FromModuleVersionID != nil {
			fromComponent.ModuleVersionID = *raw.FromModuleVersionID
		}
		if raw.FromName != nil {
			fromComponent.Name = *raw.FromName
		}
		fromComponent.Variables = raw.FromVariables

		if raw.ToID != nil {
			toComponent.ID = *raw.ToID
		}
		if raw.ToModuleVersionID != nil {
			toComponent.ModuleVersionID = *raw.ToModuleVersionID
		}
		if raw.ToName != nil {
			toComponent.Name = *raw.ToName
		}
		toComponent.Variables = raw.ToVariables

		diffs[i] = internal.ComponentDiff{
			FromComponent: fromComponent,
			ToComponent:   toComponent,
			DiffType:      internal.DiffType(raw.DiffType),
		}
	}

	return diffs, nil
}
