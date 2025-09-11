package internal

import (
	"context"
	"fmt"
	"io"
	"time"

	log "github.com/sirupsen/logrus"
)

type Plan struct {
	ID          uint      `gorm:"primarykey"`
	Component   Component `gorm:"foreignKey:ComponentID"`
	ComponentID uint
	Changeset   Changeset `gorm:"foreignKey:ChangesetID"`
	ChangesetID uint
	MergeBase   string `gorm:"column:merge_base"`
	Head        string `gorm:"column:head"`
	State       string `gorm:"default:Queued"`
	Add         *int   `gorm:"column:add"`
	Change      *int   `gorm:"column:change"`
	Destroy     *int   `gorm:"column:destroy"`
}

type PlanRepo interface {
	GetPlan(ctx context.Context, planID uint) (*Plan, error)
	GetQueuedPlans(ctx context.Context) ([]RunPlanRequest, error)
	ListPlans(ctx context.Context) ([]Plan, error)
	CreatePlan(ctx context.Context, plan *Plan) error
	UpdatePlanState(ctx context.Context, planID uint, state TaskState) error
	UpdatePlanResourceCounts(ctx context.Context, planID uint, counts PlanResourceCounts) error
}

type PlanStore interface {
	StorePlan(ctx context.Context, planID uint, planPath PlanPath) error
	LoadPlan(ctx context.Context, planID uint) (PlanPath, error)
}

type ListPlans struct {
	planRepo PlanRepo
	tx       TransactionManager
}

func NewListPlans(planRepo PlanRepo, tx TransactionManager) *ListPlans {
	return &ListPlans{
		planRepo: planRepo,
		tx:       tx,
	}
}

type ListPlansRequest struct {
	Changeset *string `json:"changeset"`
}

type ListPlansResponse struct {
	Plans []Plan `json:"plans"`
}

func (l *ListPlans) Exec(ctx context.Context, req ListPlansRequest) (*ListPlansResponse, error) {
	var plans []Plan

	branch := MainBranch
	if req.Changeset != nil {
		branch = *req.Changeset
	}

	err := l.tx.Checkout(ctx, branch, func(ctx context.Context) error {
		var err error
		plans, err = l.planRepo.ListPlans(ctx)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list plans", err)
	}

	return &ListPlansResponse{
		Plans: plans,
	}, nil
}

type CreatePlan struct {
	componentRepo ComponentRepo
	planRepo      PlanRepo
	changesetRepo ChangesetRepo
	tx            TransactionManager
	planWorker    *PlanWorker
}

func NewCreatePlan(componentRepo ComponentRepo, planRepo PlanRepo, changesetRepo ChangesetRepo, tx TransactionManager, planWorker *PlanWorker) *CreatePlan {
	return &CreatePlan{
		componentRepo: componentRepo,
		planRepo:      planRepo,
		changesetRepo: changesetRepo,
		tx:            tx,
		planWorker:    planWorker,
	}
}

type CreatePlanRequest struct {
	ComponentID uint   `json:"component_id"`
	Changeset   string `json:"changeset"`
}

type CreatePlanResponse struct {
	ID          uint   `json:"id"`
	ComponentID uint   `json:"component_id"`
	ChangesetID uint   `json:"changeset_id"`
	MergeBase   string `json:"merge_base"`
	Head        string `json:"head"`
	State       string `json:"state"`
}

func (c *CreatePlan) Exec(ctx context.Context, req CreatePlanRequest) (*CreatePlanResponse, error) {
	if req.Changeset == "" {
		return nil, UserErr("changeset is required")
	}

	var response *CreatePlanResponse
	err := c.tx.Do(ctx, req.Changeset, "create plan", func(ctx context.Context) error {
		changeset, err := c.changesetRepo.GetChangesetByName(ctx, req.Changeset)
		if err != nil {
			return UserErrE("changeset not found", err)
		}

		component, err := c.componentRepo.GetComponent(ctx, req.ComponentID)
		if err != nil {
			return UserErrE("component not found", err)
		}

		mergeBase, err := c.tx.GetMergeBase(ctx, MainBranch, req.Changeset)
		if err != nil {
			return InternalErrE("failed to get merge base", err)
		}

		head, err := c.tx.GetHead(ctx)
		if err != nil {
			return InternalErrE("failed to get head", err)
		}

		plan := &Plan{
			ComponentID: req.ComponentID,
			Component:   *component,
			ChangesetID: changeset.ID,
			MergeBase:   mergeBase,
			Head:        head,
		}

		err = c.planRepo.CreatePlan(ctx, plan)
		if err != nil {
			return InternalErrE("failed to create plan", err)
		}

		response = &CreatePlanResponse{
			ID:          plan.ID,
			ComponentID: plan.ComponentID,
			ChangesetID: plan.ChangesetID,
			MergeBase:   plan.MergeBase,
			Head:        plan.Head,
			State:       plan.State,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if c.planWorker != nil {
		c.planWorker.QueuePlan(response.ID, req.Changeset)
	}

	return response, nil
}

type RunPlan struct {
	config      *Config
	planRepo    PlanRepo
	planStore   PlanStore
	logStore    LogStore
	applyRepo   ApplyRepo
	tx          TransactionManager
	newExecutor NewExecutor
}

func NewRunPlan(config *Config, planRepo PlanRepo, planStore PlanStore, logStore LogStore, applyRepo ApplyRepo, tx TransactionManager, newExecutor NewExecutor) *RunPlan {
	return &RunPlan{
		config:      config,
		planRepo:    planRepo,
		planStore:   planStore,
		logStore:    logStore,
		applyRepo:   applyRepo,
		tx:          tx,
		newExecutor: newExecutor,
	}
}

type RunPlanRequest struct {
	PlanID uint   `json:"plan_id"`
	Branch string `json:"branch"`
}

func (r *RunPlan) Exec(ctx context.Context, req RunPlanRequest) error {
	var plan *Plan

	err := r.tx.Do(ctx, req.Branch, "start plan", func(ctx context.Context) error {
		var err error
		plan, err = r.planRepo.GetPlan(ctx, req.PlanID)
		if err != nil {
			return fmt.Errorf("failed to get plan: %w", err)
		}

		if plan.ID != req.PlanID {
			return fmt.Errorf("plan ID mismatch")
		}

		err = r.planRepo.UpdatePlanState(ctx, req.PlanID, TaskStateStarted)
		if err != nil {
			return fmt.Errorf("failed to update plan state: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	component := &plan.Component
	workDir := r.config.Terraform.WorkDir

	logWriter, err := r.logStore.NewLogWriter("plan", req.PlanID)
	if err != nil {
		return fmt.Errorf("failed to create log writer: %w", err)
	}
	defer logWriter.Close()

	executor, err := r.newExecutor(component, workDir, logWriter)
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}
	defer executor.Close()

	log.Info("Created dynamic component config in temp directory")

	err = executor.Init(ctx)
	if err != nil {
		stateErr := r.tx.Do(ctx, req.Branch, "fail plan", func(ctx context.Context) error {
			return r.planRepo.UpdatePlanState(ctx, req.PlanID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to initialize executor: %w, and failed to update plan state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to initialize executor: %w", err)
	}

	planPath, resourceCounts, err := executor.Plan(ctx)
	if err != nil {
		stateErr := r.tx.Do(ctx, req.Branch, "fail plan", func(ctx context.Context) error {
			return r.planRepo.UpdatePlanState(ctx, req.PlanID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to plan executor: %w, and failed to update plan state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to plan executor: %w", err)
	}

	err = r.planStore.StorePlan(ctx, req.PlanID, planPath)
	if err != nil {
		stateErr := r.tx.Do(ctx, req.Branch, "fail plan", func(ctx context.Context) error {
			return r.planRepo.UpdatePlanState(ctx, req.PlanID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to store plan: %w, and failed to update plan state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to store plan: %w", err)
	}

	err = r.tx.Do(ctx, req.Branch, "complete plan", func(ctx context.Context) error {
		updateErr := r.planRepo.UpdatePlanResourceCounts(ctx, req.PlanID, resourceCounts)
		if updateErr != nil {
			return fmt.Errorf("failed to update plan resource counts: %w", updateErr)
		}
		apply := &Apply{
			PlanID:      req.PlanID,
			ChangesetID: plan.ChangesetID,
		}

		err := r.applyRepo.CreateApply(ctx, apply)
		if err != nil {
			stateErr := r.planRepo.UpdatePlanState(ctx, req.PlanID, TaskStateFailed)
			if stateErr != nil {
				return fmt.Errorf("failed to create apply: %w, and failed to update plan state: %w", err, stateErr)
			}
			return fmt.Errorf("failed to create apply: %w", err)
		}

		err = r.planRepo.UpdatePlanState(ctx, req.PlanID, TaskStateCompleted)
		if err != nil {
			return fmt.Errorf("failed to update plan state: %w", err)
		}

		log.WithField("apply_id", apply.ID).WithField("plan_id", req.PlanID).Info("Created apply for plan")

		return nil
	})

	return err
}

type GetPlanLog struct {
	logStore LogStore
}

func NewGetPlanLog(logStore LogStore) *GetPlanLog {
	return &GetPlanLog{
		logStore: logStore,
	}
}

type GetPlanLogRequest struct {
	PlanID uint `json:"plan_id"`
}

type GetPlanLogResponse struct {
	Content io.ReadCloser
}

func (g *GetPlanLog) Exec(ctx context.Context, req GetPlanLogRequest) (*GetPlanLogResponse, error) {
	if req.PlanID == 0 {
		return nil, UserErr("plan ID is required")
	}

	reader, err := g.logStore.LoadLog(ctx, "plan", req.PlanID)
	if err != nil {
		return nil, InternalErrE("failed to load plan log", err)
	}

	return &GetPlanLogResponse{
		Content: reader,
	}, nil
}

type PlanWorker struct {
	runPlan  *RunPlan
	planRepo PlanRepo
	planChan chan RunPlanRequest
}

func NewPlanWorker(runPlan *RunPlan, planRepo PlanRepo) *PlanWorker {
	return &PlanWorker{
		runPlan:  runPlan,
		planRepo: planRepo,
		planChan: make(chan RunPlanRequest, 100),
	}
}

func (pw *PlanWorker) Start(ctx context.Context) {
	go pw.processPlans(ctx)
}

func (pw *PlanWorker) QueuePlan(planID uint, branch string) {
	req := RunPlanRequest{
		PlanID: planID,
		Branch: branch,
	}
	select {
	case pw.planChan <- req:
		log.WithField("plan_id", planID).WithField("branch", branch).Debug("Queued plan for processing")
	default:
		log.WithField("plan_id", planID).WithField("branch", branch).Warn("Plan channel full, plan will be picked up by polling")
	}
}

func (pw *PlanWorker) processPlans(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case req := <-pw.planChan:
			pw.runPlanInBackground(ctx, req)
		case <-ticker.C:
			pw.processQueuedPlans(ctx)
		}
	}
}

func (pw *PlanWorker) runPlanInBackground(ctx context.Context, req RunPlanRequest) {
	go func() {
		workerCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
		defer cancel()

		err := pw.runPlan.Exec(workerCtx, req)
		if err != nil {
			log.WithError(err).WithField("plan_id", req.PlanID).Error("Failed to run plan")
		} else {
			log.WithField("plan_id", req.PlanID).Info("Plan completed successfully")
		}
	}()
}

func (pw *PlanWorker) processQueuedPlans(ctx context.Context) {
	requests, err := pw.planRepo.GetQueuedPlans(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to get queued plans")
		return
	}

	for _, req := range requests {
		pw.runPlanInBackground(ctx, req)
	}
}
