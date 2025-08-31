package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/marcbran/versource/internal"
)

type GormPlanRepo struct {
	db *gorm.DB
}

func NewGormPlanRepo(db *gorm.DB) *GormPlanRepo {
	return &GormPlanRepo{db: db}
}

func (r *GormPlanRepo) GetPlan(ctx context.Context, planID uint) (*internal.Plan, error) {
	db := getTxOrDb(ctx, r.db)
	var plan internal.Plan
	err := db.WithContext(ctx).
		Preload("Component.ModuleVersion.Module").
		Preload("Component.ModuleVersion").
		Preload("Changeset").
		Where("id = ?", planID).First(&plan).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	return &plan, nil
}

func (r *GormPlanRepo) GetQueuedPlans(ctx context.Context) ([]internal.RunPlanRequest, error) {
	db := getTxOrDb(ctx, r.db)
	var plans []internal.Plan
	err := db.WithContext(ctx).Preload("Component.ModuleVersion.Module").Preload("Changeset").Where("state = ?", internal.TaskStateQueued).Find(&plans).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queued plans: %w", err)
	}

	requests := make([]internal.RunPlanRequest, len(plans))
	for i, plan := range plans {
		requests[i] = internal.RunPlanRequest{
			PlanID: plan.ID,
			Branch: internal.MainBranch, // TODO: get branch for plan
		}
	}
	return requests, nil
}

func (r *GormPlanRepo) ListPlans(ctx context.Context) ([]internal.Plan, error) {
	db := getTxOrDb(ctx, r.db)
	var plans []internal.Plan
	err := db.WithContext(ctx).
		Preload("Component.ModuleVersion.Module").
		Preload("Component.ModuleVersion").
		Preload("Changeset").
		Find(&plans).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}
	return plans, nil
}

func (r *GormPlanRepo) CreatePlan(ctx context.Context, plan *internal.Plan) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(plan).Error
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}
	return nil
}

func (r *GormPlanRepo) UpdatePlanState(ctx context.Context, planID uint, state internal.TaskState) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&internal.Plan{}).Where("id = ?", planID).Update("state", state).Error
	if err != nil {
		return fmt.Errorf("failed to update plan state: %w", err)
	}
	return nil
}
