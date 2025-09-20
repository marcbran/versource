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

func (r *GormComponentRepo) HasComponent(ctx context.Context, componentID uint) (bool, error) {
	db := getTxOrDb(ctx, r.db)
	var count int64
	err := db.WithContext(ctx).Model(&internal.Component{}).Where("id = ?", componentID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check component existence: %w", err)
	}
	return count > 0, nil
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

func (r *GormComponentDiffRepo) ListComponentDiffs(ctx context.Context, changeset string) ([]internal.ComponentDiff, error) {
	db := getTxOrDb(ctx, r.db)

	query := `
		WITH ranked AS (
		  SELECT
			d.to_id,
			d.to_module_version_id,
			d.to_name,
			d.to_variables,
			d.to_status,
			d.to_commit,
			d.to_commit_date,
			d.from_id,
			d.from_module_version_id,
			d.from_name,
			d.from_variables,
			d.from_status,
			d.from_commit,
			d.from_commit_date,
			d.diff_type,
			p.id as plan_id,
			p.component_id as plan_component_id,
			p.changeset_id as plan_changeset_id,
			p.merge_base as plan_merge_base,
			p.head as plan_head,
			p.state as plan_state,
			p.add as plan_add,
			p.change as plan_change,
			p.destroy as plan_destroy,
			ROW_NUMBER() OVER (PARTITION BY d.to_id ORDER BY d.to_commit_date DESC) AS rn
		  FROM dolt_diff_components d
		  JOIN dolt_log(?, "--not", "main", "--tables", "components") l
			ON d.to_commit = l.commit_hash
		  LEFT JOIN plans p
			ON d.to_id = p.component_id AND d.to_commit = p.head
		  ORDER BY d.to_id
		)
		SELECT *
		FROM ranked
		WHERE rn = 1;
	`

	var rawDiffs []rawDiff
	err := db.WithContext(ctx).Raw(query, changeset).Scan(&rawDiffs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list component diffs: %w", err)
	}

	diffs := make([]internal.ComponentDiff, len(rawDiffs))
	for i, raw := range rawDiffs {
		diffs[i] = convertRawDiffToComponentDiff(raw)
	}

	return diffs, nil
}

func (r *GormComponentDiffRepo) GetComponentDiff(ctx context.Context, componentID uint, changeset string) (*internal.ComponentDiff, error) {
	db := getTxOrDb(ctx, r.db)

	query := `
		SELECT
			d.to_id,
			d.to_module_version_id,
			d.to_name,
			d.to_variables,
			d.to_status,
			d.to_commit,
			d.to_commit_date,
			d.from_id,
			d.from_module_version_id,
			d.from_name,
			d.from_variables,
			d.from_status,
			d.from_commit,
			d.from_commit_date,
			d.diff_type,
			p.id as plan_id,
			p.component_id as plan_component_id,
			p.changeset_id as plan_changeset_id,
			p.merge_base as plan_merge_base,
			p.head as plan_head,
			p.state as plan_state,
			p.add as plan_add,
			p.change as plan_change,
			p.destroy as plan_destroy
		FROM dolt_diff_components d
		JOIN dolt_log(?, "--not", "main", "--tables", "components") l
			ON d.to_commit = l.commit_hash
		LEFT JOIN plans p
			ON d.to_id = p.component_id AND d.to_commit = p.head
		WHERE d.to_id = ?
		ORDER BY d.to_commit_date DESC
		LIMIT 1;
	`

	var singleRawDiff rawDiff
	err := db.WithContext(ctx).Raw(query, changeset, componentID).Scan(&singleRawDiff).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get component diff: %w", err)
	}

	diff := convertRawDiffToComponentDiff(singleRawDiff)
	return &diff, nil
}

func (r *GormComponentDiffRepo) HasComponentConflicts(ctx context.Context, changesetName string) (bool, error) {
	if !internal.IsValidBranch(changesetName) {
		return false, fmt.Errorf("invalid branch name: %s", changesetName)
	}

	db := getTxOrDb(ctx, r.db)

	query := fmt.Sprintf(`SELECT count(1) FROM dolt_diff("main...%s", "components") b JOIN dolt_diff("%s...main", "components") m ON b.to_id = m.to_id`, changesetName, changesetName)

	var count int64
	err := db.WithContext(ctx).Raw(query).Scan(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check component conflicts: %w", err)
	}

	return count > 0, nil
}

type rawDiff struct {
	ToID                *uint          `json:"to_id"`
	ToModuleVersionID   *uint          `json:"to_module_version_id"`
	ToName              *string        `json:"to_name"`
	ToVariables         datatypes.JSON `json:"to_variables"`
	ToStatus            *string        `json:"to_status"`
	ToCommit            string         `json:"to_commit"`
	ToCommitDate        string         `json:"to_commit_date"`
	FromID              *uint          `json:"from_id"`
	FromModuleVersionID *uint          `json:"from_module_version_id"`
	FromName            *string        `json:"from_name"`
	FromVariables       datatypes.JSON `json:"from_variables"`
	FromStatus          *string        `json:"from_status"`
	FromCommit          string         `json:"from_commit"`
	FromCommitDate      string         `json:"from_commit_date"`
	DiffType            string         `json:"diff_type"`
	PlanID              *uint          `json:"plan_id"`
	PlanComponentID     *uint          `json:"plan_component_id"`
	PlanChangesetID     *uint          `json:"plan_changeset_id"`
	PlanMergeBase       *string        `json:"plan_merge_base"`
	PlanHead            *string        `json:"plan_head"`
	PlanState           *string        `json:"plan_state"`
	PlanAdd             *int           `json:"plan_add"`
	PlanChange          *int           `json:"plan_change"`
	PlanDestroy         *int           `json:"plan_destroy"`
}

func convertRawDiffToComponentDiff(raw rawDiff) internal.ComponentDiff {
	var fromComponent, toComponent *internal.Component

	if raw.FromID != nil {
		fromComponent = &internal.Component{}
		fromComponent.ID = *raw.FromID
		if raw.FromModuleVersionID != nil {
			fromComponent.ModuleVersionID = *raw.FromModuleVersionID
		}
		if raw.FromName != nil {
			fromComponent.Name = *raw.FromName
		}
		fromComponent.Variables = raw.FromVariables
		if raw.FromStatus != nil {
			fromComponent.Status = internal.ComponentStatus(*raw.FromStatus)
		}
	}

	if raw.ToID != nil {
		toComponent = &internal.Component{}
		toComponent.ID = *raw.ToID
		if raw.ToModuleVersionID != nil {
			toComponent.ModuleVersionID = *raw.ToModuleVersionID
		}
		if raw.ToName != nil {
			toComponent.Name = *raw.ToName
		}
		toComponent.Variables = raw.ToVariables
		if raw.ToStatus != nil {
			toComponent.Status = internal.ComponentStatus(*raw.ToStatus)
		}
	}

	diffType := internal.DiffType(raw.DiffType)
	if raw.ToStatus != nil && *raw.ToStatus == string(internal.ComponentStatusDeleted) {
		diffType = internal.DiffTypeDeleted
	} else if raw.DiffType == "added" {
		diffType = internal.DiffTypeCreated
	} else if raw.DiffType == "modified" {
		diffType = internal.DiffTypeModified
	}

	var plan *internal.Plan
	if raw.PlanID != nil {
		plan = &internal.Plan{
			ID:          *raw.PlanID,
			ComponentID: *raw.PlanComponentID,
			ChangesetID: *raw.PlanChangesetID,
			MergeBase:   *raw.PlanMergeBase,
			Head:        *raw.PlanHead,
			State:       internal.TaskState(*raw.PlanState),
			Add:         raw.PlanAdd,
			Change:      raw.PlanChange,
			Destroy:     raw.PlanDestroy,
		}
	}

	return internal.ComponentDiff{
		FromComponent: fromComponent,
		ToComponent:   toComponent,
		DiffType:      diffType,
		Plan:          plan,
	}
}
