package internal

import (
	"context"
	"fmt"
	"io"
	"time"

	log "github.com/sirupsen/logrus"
)

type Plan struct {
	ID          uint      `gorm:"primarykey" json:"id" yaml:"id"`
	Changeset   Changeset `gorm:"foreignKey:ChangesetID" json:"changeset" yaml:"changeset"`
	ChangesetID uint      `json:"changesetId" yaml:"changesetId"`
	ComponentID uint      `json:"componentId" yaml:"componentId"`
	From        string    `gorm:"column:from" json:"from" yaml:"from"`
	To          string    `gorm:"column:to" json:"to" yaml:"to"`
	State       TaskState `gorm:"default:Queued" json:"state" yaml:"state"`
	Add         *int      `gorm:"column:add" json:"add" yaml:"add"`
	Change      *int      `gorm:"column:change" json:"change" yaml:"change"`
	Destroy     *int      `gorm:"column:destroy" json:"destroy" yaml:"destroy"`
}

type PlanRepo interface {
	GetPlan(ctx context.Context, planID uint) (*Plan, error)
	GetQueuedPlans(ctx context.Context) ([]uint, error)
	ListPlans(ctx context.Context) ([]Plan, error)
	ListPlansByChangeset(ctx context.Context, changesetID uint) ([]Plan, error)
	ListPlansByChangesetName(ctx context.Context, changesetName string) ([]Plan, error)
	CreatePlan(ctx context.Context, plan *Plan) error
	UpdatePlanState(ctx context.Context, planID uint, state TaskState) error
	UpdatePlanResourceCounts(ctx context.Context, planID uint, counts PlanResourceCounts) error
	DeletePlan(ctx context.Context, planID uint) error
}

type PlanStore interface {
	StorePlan(ctx context.Context, planID uint, planPath PlanPath) error
	LoadPlan(ctx context.Context, planID uint) (PlanPath, error)
	DeletePlan(ctx context.Context, planID uint) error
}

type GetPlan struct {
	planRepo      PlanRepo
	componentRepo ComponentRepo
	tx            TransactionManager
}

func NewGetPlan(planRepo PlanRepo, componentRepo ComponentRepo, tx TransactionManager) *GetPlan {
	return &GetPlan{
		planRepo:      planRepo,
		componentRepo: componentRepo,
		tx:            tx,
	}
}

type GetPlanRequest struct {
	ChangesetName *string `json:"changesetName" yaml:"changesetName"`
	PlanID        uint    `json:"planId" yaml:"planId"`
}

type GetPlanResponse struct {
	ID          uint      `json:"id" yaml:"id"`
	ComponentID uint      `json:"componentId" yaml:"componentId"`
	ChangesetID uint      `json:"changesetId" yaml:"changesetId"`
	From        string    `json:"from" yaml:"from"`
	To          string    `json:"to" yaml:"to"`
	State       TaskState `json:"state" yaml:"state"`
	Add         *int      `json:"add" yaml:"add"`
	Change      *int      `json:"change" yaml:"change"`
	Destroy     *int      `json:"destroy" yaml:"destroy"`
	Component   Component `json:"component" yaml:"component"`
	Changeset   Changeset `json:"changeset" yaml:"changeset"`
}

func (g *GetPlan) Exec(ctx context.Context, req GetPlanRequest) (*GetPlanResponse, error) {
	if req.PlanID == 0 {
		return nil, UserErr("plan ID is required")
	}

	var plan *Plan
	err := g.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		plan, err = g.planRepo.GetPlan(ctx, req.PlanID)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list plans", err)
	}

	var component *Component
	branch := plan.Changeset.Name
	if plan.Changeset.State == ChangesetStateMerged {
		branch = MainBranch
	}

	err = g.tx.Checkout(ctx, branch, func(ctx context.Context) error {
		var err error
		component, err = g.componentRepo.GetComponentAtCommit(ctx, plan.ComponentID, plan.To)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to get component at commit", err)
	}

	return &GetPlanResponse{
		ID:          plan.ID,
		ComponentID: plan.ComponentID,
		ChangesetID: plan.ChangesetID,
		From:        plan.From,
		To:          plan.To,
		State:       plan.State,
		Add:         plan.Add,
		Change:      plan.Change,
		Destroy:     plan.Destroy,
		Component:   *component,
		Changeset:   plan.Changeset,
	}, nil
}

type GetPlanLog struct {
	logStore LogStore
	tx       TransactionManager
}

func NewGetPlanLog(logStore LogStore, tx TransactionManager) *GetPlanLog {
	return &GetPlanLog{
		logStore: logStore,
		tx:       tx,
	}
}

type GetPlanLogRequest struct {
	ChangesetName *string `json:"changesetName" yaml:"changesetName"`
	PlanID        uint    `json:"planId" yaml:"planId"`
}

type GetPlanLogResponse struct {
	Content io.ReadCloser `json:"content" yaml:"content"`
}

func (g *GetPlanLog) Exec(ctx context.Context, req GetPlanLogRequest) (*GetPlanLogResponse, error) {
	if req.PlanID == 0 {
		return nil, UserErr("plan ID is required")
	}

	var reader io.ReadCloser
	err := g.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		reader, err = g.logStore.LoadLog(ctx, "plan", req.PlanID)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to load plan log", err)
	}

	return &GetPlanLogResponse{
		Content: reader,
	}, nil
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
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type ListPlansResponse struct {
	Plans []Plan `json:"plans" yaml:"plans"`
}

func (l *ListPlans) Exec(ctx context.Context, req ListPlansRequest) (*ListPlansResponse, error) {
	var plans []Plan

	err := l.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		if req.ChangesetName == "" {
			plans, err = l.planRepo.ListPlans(ctx)
		} else {
			plans, err = l.planRepo.ListPlansByChangesetName(ctx, req.ChangesetName)
		}
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
	componentRepo       ComponentRepo
	componentChangeRepo ComponentChangeRepo
	planRepo            PlanRepo
	changesetRepo       ChangesetRepo
	tx                  TransactionManager
	planWorker          *PlanWorker
}

func NewCreatePlan(componentRepo ComponentRepo, componentChangeRepo ComponentChangeRepo, planRepo PlanRepo, changesetRepo ChangesetRepo, tx TransactionManager, planWorker *PlanWorker) *CreatePlan {
	return &CreatePlan{
		componentRepo:       componentRepo,
		componentChangeRepo: componentChangeRepo,
		planRepo:            planRepo,
		changesetRepo:       changesetRepo,
		tx:                  tx,
		planWorker:          planWorker,
	}
}

type CreatePlanRequest struct {
	ComponentID uint   `json:"componentId" yaml:"componentId"`
	Changeset   string `json:"changeset" yaml:"changeset"`
}

type CreatePlanResponse struct {
	ID          uint      `json:"id" yaml:"id"`
	ComponentID uint      `json:"componentId" yaml:"componentId"`
	ChangesetID uint      `json:"changesetId" yaml:"changesetId"`
	From        string    `json:"from" yaml:"from"`
	To          string    `json:"to" yaml:"to"`
	State       TaskState `json:"state" yaml:"state"`
}

func (c *CreatePlan) Exec(ctx context.Context, req CreatePlanRequest) (*CreatePlanResponse, error) {
	if req.Changeset == "" {
		return nil, UserErr("changeset is required")
	}

	var from string
	var to string

	err := c.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		commit, err := c.componentRepo.GetLastCommitOfComponent(ctx, req.ComponentID)
		if err != nil {
			return InternalErrE("failed to get last commit of component from main", err)
		}

		from = commit
		return nil
	})

	if err != nil {
		return nil, err
	}

	err = c.tx.Checkout(ctx, req.Changeset, func(ctx context.Context) error {
		change, err := c.componentChangeRepo.GetComponentChange(ctx, req.ComponentID)
		if err != nil {
			return InternalErrE("failed to get component change from changeset", err)
		}

		if change.ToComponent == nil {
			return UserErr("cannot create plan for component with no changes")
		}

		to = change.ToCommit
		return nil
	})

	if err != nil {
		return nil, err
	}

	var response *CreatePlanResponse
	err = c.tx.Do(ctx, AdminBranch, "create plan", func(ctx context.Context) error {
		changeset, err := c.changesetRepo.GetChangesetByName(ctx, req.Changeset)
		if err != nil {
			return UserErrE("changeset not found", err)
		}

		plan := &Plan{
			ComponentID: req.ComponentID,
			ChangesetID: changeset.ID,
			From:        from,
			To:          to,
		}

		err = c.planRepo.CreatePlan(ctx, plan)
		if err != nil {
			return InternalErrE("failed to create plan", err)
		}

		response = &CreatePlanResponse{
			ID:          plan.ID,
			ComponentID: plan.ComponentID,
			ChangesetID: plan.ChangesetID,
			From:        plan.From,
			To:          plan.To,
			State:       plan.State,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if c.planWorker != nil {
		c.planWorker.QueuePlan(response.ID)
	}

	return response, nil
}

type PlanWorker struct {
	runPlan  *RunPlan
	planRepo PlanRepo
	tx       TransactionManager
	planChan chan uint
}

func NewPlanWorker(runPlan *RunPlan, planRepo PlanRepo, tx TransactionManager) *PlanWorker {
	return &PlanWorker{
		runPlan:  runPlan,
		planRepo: planRepo,
		tx:       tx,
		planChan: make(chan uint, 100),
	}
}

func (pw *PlanWorker) Start(ctx context.Context) {
	go pw.processPlans(ctx)
}

func (pw *PlanWorker) QueuePlan(planID uint) {
	select {
	case pw.planChan <- planID:
		log.WithField("plan_id", planID).Debug("Queued plan for processing")
	default:
		log.WithField("plan_id", planID).Warn("Plan channel full, plan will be picked up by polling")
	}
}

func (pw *PlanWorker) processPlans(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case planID := <-pw.planChan:
			pw.runPlanInBackground(ctx, planID)
		case <-ticker.C:
			pw.processQueuedPlans(ctx)
		}
	}
}

func (pw *PlanWorker) runPlanInBackground(ctx context.Context, planID uint) {
	go func() {
		workerCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
		defer cancel()

		err := pw.runPlan.Exec(workerCtx, planID)
		if err != nil {
			log.WithError(err).WithField("plan_id", planID).Error("Failed to run plan")
		} else {
			log.WithField("plan_id", planID).Info("Plan completed successfully")
		}
	}()
}

func (pw *PlanWorker) processQueuedPlans(ctx context.Context) {
	var planIDs []uint
	err := pw.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		planIDs, err = pw.planRepo.GetQueuedPlans(ctx)
		return err
	})
	if err != nil {
		log.WithError(err).Error("Failed to get queued plans")
		return
	}

	for _, planID := range planIDs {
		pw.runPlanInBackground(ctx, planID)
	}
}

type RunPlan struct {
	config        *Config
	planRepo      PlanRepo
	planStore     PlanStore
	logStore      LogStore
	tx            TransactionManager
	newExecutor   NewExecutor
	componentRepo ComponentRepo
}

func NewRunPlan(config *Config, planRepo PlanRepo, planStore PlanStore, logStore LogStore, tx TransactionManager, newExecutor NewExecutor, componentRepo ComponentRepo) *RunPlan {
	return &RunPlan{
		config:        config,
		planRepo:      planRepo,
		planStore:     planStore,
		logStore:      logStore,
		tx:            tx,
		newExecutor:   newExecutor,
		componentRepo: componentRepo,
	}
}

func (r *RunPlan) Exec(ctx context.Context, planID uint) error {
	var plan *Plan

	err := r.tx.Do(ctx, AdminBranch, "start plan", func(ctx context.Context) error {
		var err error
		plan, err = r.planRepo.GetPlan(ctx, planID)
		if err != nil {
			return fmt.Errorf("failed to get plan: %w", err)
		}

		if plan.ID != planID {
			return fmt.Errorf("plan ID mismatch")
		}

		err = r.planRepo.UpdatePlanState(ctx, planID, TaskStateStarted)
		if err != nil {
			return fmt.Errorf("failed to update plan state: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	var component *Component
	err = r.tx.Checkout(ctx, plan.Changeset.Name, func(ctx context.Context) error {
		var err error
		component, err = r.componentRepo.GetComponentAtCommit(ctx, plan.ComponentID, plan.To)
		return err
	})
	if err != nil {
		stateErr := r.tx.Do(ctx, AdminBranch, "fail plan", func(ctx context.Context) error {
			return r.planRepo.UpdatePlanState(ctx, planID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to get component at commit: %w, and failed to update plan state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to get component at commit: %w", err)
	}

	workDir := r.config.Terraform.WorkDir

	logWriter, err := r.logStore.NewLogWriter("plan", planID)
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
		stateErr := r.tx.Do(ctx, AdminBranch, "fail plan", func(ctx context.Context) error {
			return r.planRepo.UpdatePlanState(ctx, planID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to initialize executor: %w, and failed to update plan state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to initialize executor: %w", err)
	}

	planPath, resourceCounts, err := executor.Plan(ctx)
	if err != nil {
		stateErr := r.tx.Do(ctx, AdminBranch, "fail plan", func(ctx context.Context) error {
			return r.planRepo.UpdatePlanState(ctx, planID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to plan executor: %w, and failed to update plan state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to plan executor: %w", err)
	}

	err = r.planStore.StorePlan(ctx, planID, planPath)
	if err != nil {
		stateErr := r.tx.Do(ctx, AdminBranch, "fail plan", func(ctx context.Context) error {
			return r.planRepo.UpdatePlanState(ctx, planID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to store plan: %w, and failed to update plan state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to store plan: %w", err)
	}

	err = r.tx.Do(ctx, AdminBranch, "complete plan", func(ctx context.Context) error {
		updateErr := r.planRepo.UpdatePlanResourceCounts(ctx, planID, resourceCounts)
		if updateErr != nil {
			return fmt.Errorf("failed to update plan resource counts: %w", updateErr)
		}

		err := r.planRepo.UpdatePlanState(ctx, planID, TaskStateSucceeded)
		if err != nil {
			return fmt.Errorf("failed to update plan state: %w", err)
		}

		log.WithField("plan_id", planID).Info("Plan completed successfully")

		return nil
	})

	return err
}
