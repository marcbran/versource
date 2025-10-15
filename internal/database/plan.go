package database

import (
	"context"
	"fmt"

	"github.com/marcbran/versource/pkg/versource"
	"gorm.io/gorm"

	"github.com/marcbran/versource/internal"
)

type GormPlanRepo struct {
	db *gorm.DB
}

func NewGormPlanRepo(db *gorm.DB) *GormPlanRepo {
	return &GormPlanRepo{db: db}
}

func (r *GormPlanRepo) GetPlan(ctx context.Context, planID uint) (*versource.Plan, error) {
	db := getTxOrDb(ctx, r.db)
	var plan versource.Plan
	err := db.WithContext(ctx).
		Preload("Changeset").
		Where("id = ?", planID).First(&plan).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	return &plan, nil
}

func (r *GormPlanRepo) GetQueuedPlans(ctx context.Context) ([]uint, error) {
	db := getTxOrDb(ctx, r.db)
	var plans []versource.Plan
	err := db.WithContext(ctx).Preload("Changeset").Where("state = ?", versource.TaskStateQueued).Find(&plans).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queued plans: %w", err)
	}

	planIDs := make([]uint, len(plans))
	for i, plan := range plans {
		planIDs[i] = plan.ID
	}
	return planIDs, nil
}

func (r *GormPlanRepo) ListPlans(ctx context.Context) ([]versource.Plan, error) {
	db := getTxOrDb(ctx, r.db)
	var plans []versource.Plan
	err := db.WithContext(ctx).
		Preload("Changeset").
		Find(&plans).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}
	return plans, nil
}

func (r *GormPlanRepo) ListPlansByChangeset(ctx context.Context, changesetID uint) ([]versource.Plan, error) {
	db := getTxOrDb(ctx, r.db)
	var plans []versource.Plan
	err := db.WithContext(ctx).
		Preload("Changeset").
		Where("changeset_id = ?", changesetID).
		Find(&plans).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list plans by changeset: %w", err)
	}
	return plans, nil
}

func (r *GormPlanRepo) ListPlansByChangesetName(ctx context.Context, changesetName string) ([]versource.Plan, error) {
	db := getTxOrDb(ctx, r.db)
	var plans []versource.Plan
	err := db.WithContext(ctx).
		Preload("Changeset").
		Joins("JOIN changesets ON plans.changeset_id = changesets.id").
		Where("changesets.name = ?", changesetName).
		Find(&plans).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list plans by changeset name: %w", err)
	}
	return plans, nil
}

func (r *GormPlanRepo) CreatePlan(ctx context.Context, plan *versource.Plan) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(plan).Error
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}
	return nil
}

func (r *GormPlanRepo) UpdatePlanState(ctx context.Context, planID uint, state versource.TaskState) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&versource.Plan{}).Where("id = ?", planID).Update("state", state).Error
	if err != nil {
		return fmt.Errorf("failed to update plan state: %w", err)
	}
	return nil
}

func (r *GormPlanRepo) UpdatePlanResourceCounts(ctx context.Context, planID uint, counts internal.PlanResourceCounts) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&versource.Plan{}).Where("id = ?", planID).Updates(map[string]any{
		"add":     counts.AddCount,
		"change":  counts.ChangeCount,
		"destroy": counts.DestroyCount,
	}).Error
	if err != nil {
		return fmt.Errorf("failed to update plan resource counts: %w", err)
	}
	return nil
}

func (r *GormPlanRepo) DeletePlan(ctx context.Context, planID uint) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Delete(&versource.Plan{}, planID).Error
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}
	return nil
}
