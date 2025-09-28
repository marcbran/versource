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

type GormComponentChangeRepo struct {
	db *gorm.DB
}

func NewGormComponentChangeRepo(db *gorm.DB) *GormComponentChangeRepo {
	return &GormComponentChangeRepo{db: db}
}

func (r *GormComponentChangeRepo) ListComponentChanges(ctx context.Context) ([]internal.ComponentChange, error) {
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
				m.to_id as from_id,
				m.to_module_version_id as from_module_version_id,
				m.to_name as from_name,
				m.to_variables as from_variables,
				m.to_status as from_status,
				m.to_commit as from_commit,
				m.to_commit_date as from_commit_date,
				p.id as plan_id,
				p.component_id as plan_component_id,
				p.changeset_id as plan_changeset_id,
				p.from as plan_from,
				p.to as plan_to,
				p.state as plan_state,
				p.add as plan_add,
				p.change as plan_change,
				p.destroy as plan_destroy,
				ROW_NUMBER() OVER (
					PARTITION BY d.to_id
					ORDER BY d.to_commit_date DESC, m.to_commit_date DESC
				) AS rn
			FROM dolt_diff_components d
			LEFT JOIN dolt_diff_components AS OF main m
				ON d.to_id = m.to_id
			LEFT JOIN plans AS OF "admin" p
				ON d.to_id = p.component_id AND d.to_commit = p.to
				AND (m.to_commit IS NULL OR m.to_commit = p.from)
		)
		SELECT *
		FROM ranked
		WHERE rn = 1;
	`

	var rawDiffs []rawDiff
	err := db.WithContext(ctx).Raw(query).Scan(&rawDiffs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list component changes: %w", err)
	}

	changes := make([]internal.ComponentChange, len(rawDiffs))
	for i, raw := range rawDiffs {
		changes[i] = convertRawDiffToComponentChange(raw)
	}

	return changes, nil
}

func (r *GormComponentChangeRepo) GetComponentChange(ctx context.Context, componentID uint) (*internal.ComponentChange, error) {
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
			m.to_id as from_id,
			m.to_module_version_id as from_module_version_id,
			m.to_name as from_name,
			m.to_variables as from_variables,
			m.to_status as from_status,
			m.to_commit as from_commit,
			m.to_commit_date as from_commit_date,
			p.id as plan_id,
			p.component_id as plan_component_id,
			p.changeset_id as plan_changeset_id,
			p.from as plan_from,
			p.to as plan_to,
			p.state as plan_state,
			p.add as plan_add,
			p.change as plan_change,
			p.destroy as plan_destroy
		FROM dolt_diff_components d
		LEFT JOIN dolt_diff_components AS OF main AS m
			ON d.to_id = m.to_id
		LEFT JOIN plans AS OF "admin" AS p
			ON d.to_id = p.component_id AND d.to_commit = p.to
			AND (m.to_commit IS NULL OR m.to_commit = p.from)
		WHERE d.to_id = ?
		ORDER BY d.to_commit_date DESC, m.to_commit_date DESC
		LIMIT 1;
	`

	var singleRawDiff rawDiff
	err := db.WithContext(ctx).Raw(query, componentID).Scan(&singleRawDiff).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get component change: %w", err)
	}

	change := convertRawDiffToComponentChange(singleRawDiff)
	return &change, nil
}

func (r *GormComponentChangeRepo) HasComponentConflicts(ctx context.Context, changesetName string) (bool, error) {
	if !internal.IsValidBranch(changesetName) {
		return false, fmt.Errorf("invalid branch name: %s", changesetName)
	}

	db := getTxOrDb(ctx, r.db)

	query := fmt.Sprintf(`
		SELECT count(1)
		FROM dolt_diff("main...%s", "components") b
		JOIN dolt_diff("%s...main", "components") m
		ON b.to_id = m.to_id
		OR b.to_name = m.to_name`, changesetName, changesetName)

	var count int64
	err := db.WithContext(ctx).Raw(query).Scan(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check component conflicts: %w", err)
	}

	return count > 0, nil
}

type rawDiff struct {
	ToID                *uint          `json:"toId"`
	ToModuleVersionID   *uint          `json:"toModuleVersionId"`
	ToName              *string        `json:"toName"`
	ToVariables         datatypes.JSON `json:"toVariables"`
	ToStatus            *string        `json:"toStatus"`
	ToCommit            string         `json:"toCommit"`
	ToCommitDate        string         `json:"toCommitDate"`
	FromID              *uint          `json:"fromId"`
	FromModuleVersionID *uint          `json:"fromModuleVersionId"`
	FromName            *string        `json:"fromName"`
	FromVariables       datatypes.JSON `json:"fromVariables"`
	FromStatus          *string        `json:"fromStatus"`
	FromCommit          string         `json:"fromCommit"`
	FromCommitDate      string         `json:"fromCommitDate"`
	PlanID              *uint          `json:"planId"`
	PlanComponentID     *uint          `json:"planComponentId"`
	PlanChangesetID     *uint          `json:"planChangesetId"`
	PlanFrom            *string        `json:"planFrom"`
	PlanTo              *string        `json:"planTo"`
	PlanState           *string        `json:"planState"`
	PlanAdd             *int           `json:"planAdd"`
	PlanChange          *int           `json:"planChange"`
	PlanDestroy         *int           `json:"planDestroy"`
}

func convertRawDiffToComponentChange(raw rawDiff) internal.ComponentChange {
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

	changeType := internal.ChangeTypeModified
	if raw.FromID == nil {
		changeType = internal.ChangeTypeCreated
	} else if raw.ToStatus != nil && *raw.ToStatus == string(internal.ComponentStatusDeleted) {
		changeType = internal.ChangeTypeDeleted
	}

	var plan *internal.Plan
	if raw.PlanID != nil {
		plan = &internal.Plan{
			ID:          *raw.PlanID,
			ComponentID: *raw.PlanComponentID,
			ChangesetID: *raw.PlanChangesetID,
			From:        *raw.PlanFrom,
			To:          *raw.PlanTo,
			State:       internal.TaskState(*raw.PlanState),
			Add:         raw.PlanAdd,
			Change:      raw.PlanChange,
			Destroy:     raw.PlanDestroy,
		}
	}

	return internal.ComponentChange{
		FromComponent: fromComponent,
		ToComponent:   toComponent,
		ChangeType:    changeType,
		Plan:          plan,
		FromCommit:    raw.FromCommit,
		ToCommit:      raw.ToCommit,
	}
}
