package database

import (
	"context"
	"errors"
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

type GormPlanRepo struct {
	db *gorm.DB
}

func NewGormPlanRepo(db *gorm.DB) *GormPlanRepo {
	return &GormPlanRepo{db: db}
}

func (r *GormPlanRepo) CreatePlan(ctx context.Context, plan *internal.Plan) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(plan).Error
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}
	return nil
}

func (r *GormPlanRepo) GetPlan(ctx context.Context, planID uint) (*internal.Plan, error) {
	db := getTxOrDb(ctx, r.db)
	var plan internal.Plan
	err := db.WithContext(ctx).Preload("Component").Preload("Changeset").Where("id = ?", planID).First(&plan).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	return &plan, nil
}

func (r *GormPlanRepo) UpdatePlanState(ctx context.Context, planID uint, state internal.TaskState) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&internal.Plan{}).Where("id = ?", planID).Update("state", state).Error
	if err != nil {
		return fmt.Errorf("failed to update plan state: %w", err)
	}
	return nil
}

func (r *GormPlanRepo) GetQueuedPlans(ctx context.Context) ([]internal.RunPlanRequest, error) {
	db := getTxOrDb(ctx, r.db)
	var plans []internal.Plan
	err := db.WithContext(ctx).Preload("Component").Preload("Changeset").Where("state = ?", internal.TaskStateQueued).Find(&plans).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queued plans: %w", err)
	}

	requests := make([]internal.RunPlanRequest, len(plans))
	for i, plan := range plans {
		requests[i] = internal.RunPlanRequest{
			PlanID: plan.ID,
			Branch: "main", // TODO: get branch for plan
		}
	}
	return requests, nil
}

type GormApplyRepo struct {
	db *gorm.DB
}

func NewGormApplyRepo(db *gorm.DB) *GormApplyRepo {
	return &GormApplyRepo{db: db}
}

func (r *GormApplyRepo) CreateApply(ctx context.Context, apply *internal.Apply) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(apply).Error
	if err != nil {
		return fmt.Errorf("failed to create apply: %w", err)
	}
	return nil
}

func (r *GormApplyRepo) GetApply(ctx context.Context, applyID uint) (*internal.Apply, error) {
	db := getTxOrDb(ctx, r.db)
	var apply internal.Apply
	err := db.WithContext(ctx).Preload("Plan.Component").Preload("Plan.Changeset").Preload("Changeset").Where("id = ?", applyID).First(&apply).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get apply: %w", err)
	}
	return &apply, nil
}

func (r *GormApplyRepo) UpdateApplyState(ctx context.Context, applyID uint, state internal.TaskState) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&internal.Apply{}).Where("id = ?", applyID).Update("state", state).Error
	if err != nil {
		return fmt.Errorf("failed to update apply state: %w", err)
	}
	return nil
}

func (r *GormApplyRepo) GetQueuedApplies(ctx context.Context) ([]uint, error) {
	db := getTxOrDb(ctx, r.db)
	var applies []internal.Apply
	err := db.WithContext(ctx).Where("state = ?", internal.TaskStateQueued).Find(&applies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queued applies: %w", err)
	}

	applyIDs := make([]uint, len(applies))
	for i, apply := range applies {
		applyIDs[i] = apply.ID
	}
	return applyIDs, nil
}

func (r *GormApplyRepo) GetQueuedAppliesByChangeset(ctx context.Context, changesetID uint) ([]uint, error) {
	db := getTxOrDb(ctx, r.db)
	var applies []internal.Apply
	err := db.WithContext(ctx).Where("state = ? AND changeset_id = ?", internal.TaskStateQueued, changesetID).Find(&applies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get queued applies for changeset: %w", err)
	}

	applyIDs := make([]uint, len(applies))
	for i, apply := range applies {
		applyIDs[i] = apply.ID
	}
	return applyIDs, nil
}

type GormChangesetRepo struct {
	db *gorm.DB
}

func NewGormChangesetRepo(db *gorm.DB) *GormChangesetRepo {
	return &GormChangesetRepo{db: db}
}

func (r *GormChangesetRepo) CreateChangeset(ctx context.Context, changeset *internal.Changeset) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Create(changeset).Error
	if err != nil {
		return fmt.Errorf("failed to create changeset: %w", err)
	}
	return nil
}

func (r *GormChangesetRepo) GetChangeset(ctx context.Context, changesetID uint) (*internal.Changeset, error) {
	db := getTxOrDb(ctx, r.db)
	var changeset internal.Changeset
	err := db.WithContext(ctx).Where("id = ?", changesetID).First(&changeset).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get changeset: %w", err)
	}
	return &changeset, nil
}

func (r *GormChangesetRepo) GetChangesetByName(ctx context.Context, name string) (*internal.Changeset, error) {
	db := getTxOrDb(ctx, r.db)
	var changeset internal.Changeset
	err := db.WithContext(ctx).Where("name = ?", name).First(&changeset).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get changeset by name: %w", err)
	}
	return &changeset, nil
}

func (r *GormChangesetRepo) UpdateChangesetState(ctx context.Context, changesetID uint, state internal.ChangesetState) error {
	db := getTxOrDb(ctx, r.db)
	err := db.WithContext(ctx).Model(&internal.Changeset{}).Where("id = ?", changesetID).Update("state", state).Error
	if err != nil {
		return fmt.Errorf("failed to update changeset state: %w", err)
	}
	return nil
}

func (r *GormChangesetRepo) ListChangesets(ctx context.Context) ([]internal.Changeset, error) {
	db := getTxOrDb(ctx, r.db)
	var changesets []internal.Changeset
	err := db.WithContext(ctx).Find(&changesets).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list changesets: %w", err)
	}
	return changesets, nil
}
