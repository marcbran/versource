package internal

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/marcbran/versource/pkg/versource"
	log "github.com/sirupsen/logrus"
)

type PlanRepo interface {
	GetPlan(ctx context.Context, planID uint) (*versource.Plan, error)
	GetQueuedPlans(ctx context.Context) ([]uint, error)
	ListPlans(ctx context.Context) ([]versource.Plan, error)
	ListPlansByChangeset(ctx context.Context, changesetID uint) ([]versource.Plan, error)
	ListPlansByChangesetName(ctx context.Context, changesetName string) ([]versource.Plan, error)
	CreatePlan(ctx context.Context, plan *versource.Plan) error
	UpdatePlanState(ctx context.Context, planID uint, state versource.TaskState) error
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

func (g *GetPlan) Exec(ctx context.Context, req versource.GetPlanRequest) (*versource.GetPlanResponse, error) {
	if req.PlanID == 0 {
		return nil, versource.UserErr("plan ID is required")
	}

	var plan *versource.Plan
	err := g.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		plan, err = g.planRepo.GetPlan(ctx, req.PlanID)
		return err
	})
	if err != nil {
		return nil, versource.InternalErrE("failed to list plans", err)
	}

	var component *versource.Component
	branch := plan.Changeset.Name
	if plan.Changeset.State == versource.ChangesetStateMerged {
		branch = MainBranch
	}

	err = g.tx.Checkout(ctx, branch, func(ctx context.Context) error {
		var err error
		component, err = g.componentRepo.GetComponentAtCommit(ctx, plan.ComponentID, plan.To)
		return err
	})
	if err != nil {
		return nil, versource.InternalErrE("failed to get component at commit", err)
	}

	return &versource.GetPlanResponse{
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

func (g *GetPlanLog) Exec(ctx context.Context, req versource.GetPlanLogRequest) (*versource.GetPlanLogResponse, error) {
	if req.PlanID == 0 {
		return nil, versource.UserErr("plan ID is required")
	}

	var reader io.ReadCloser
	err := g.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		reader, err = g.logStore.LoadLog(ctx, "plan", req.PlanID)
		return err
	})
	if err != nil {
		return nil, versource.InternalErrE("failed to load plan log", err)
	}

	return &versource.GetPlanLogResponse{
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

func (l *ListPlans) Exec(ctx context.Context, req versource.ListPlansRequest) (*versource.ListPlansResponse, error) {
	var plans []versource.Plan

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
		return nil, versource.InternalErrE("failed to list plans", err)
	}

	return &versource.ListPlansResponse{
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

func (c *CreatePlan) Exec(ctx context.Context, req versource.CreatePlanRequest) (*versource.CreatePlanResponse, error) {
	if req.ChangesetName == "" {
		return nil, versource.UserErr("changeset is required")
	}

	var from string
	var to string

	err := c.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		commit, err := c.componentRepo.GetLastCommitOfComponent(ctx, req.ComponentID)
		if err != nil {
			return versource.InternalErrE("failed to get last commit of component from main", err)
		}

		from = commit
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = c.tx.Checkout(ctx, req.ChangesetName, func(ctx context.Context) error {
		change, err := c.componentChangeRepo.GetComponentChange(ctx, req.ComponentID)
		if err != nil {
			return versource.InternalErrE("failed to get component change from changeset", err)
		}

		if change.ToComponent == nil {
			return versource.UserErr("cannot create plan for component with no changes")
		}

		to = change.ToCommit
		return nil
	})
	if err != nil {
		return nil, err
	}

	var response *versource.CreatePlanResponse
	err = c.tx.Do(ctx, AdminBranch, "create plan", func(ctx context.Context) error {
		changeset, err := c.changesetRepo.GetChangesetByName(ctx, req.ChangesetName)
		if err != nil {
			return versource.UserErrE("changeset not found", err)
		}

		plan := &versource.Plan{
			ComponentID: req.ComponentID,
			ChangesetID: changeset.ID,
			From:        from,
			To:          to,
		}

		err = c.planRepo.CreatePlan(ctx, plan)
		if err != nil {
			return versource.InternalErrE("failed to create plan", err)
		}

		response = &versource.CreatePlanResponse{
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
			stateErr := pw.tx.Do(ctx, AdminBranch, "fail plan", func(ctx context.Context) error {
				return pw.planRepo.UpdatePlanState(ctx, planID, versource.TaskStateFailed)
			})
			if stateErr != nil {
				log.WithError(err).
					WithField("plan_id", planID).
					Error("Failed to fail plan")
			}
			log.WithError(err).
				WithField("plan_id", planID).
				Error("Failed to run plan")
		} else {
			log.WithField("plan_id", planID).
				Info("Plan completed successfully")
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
	config        *versource.Config
	planRepo      PlanRepo
	planStore     PlanStore
	logStore      LogStore
	tx            TransactionManager
	newExecutor   NewExecutor
	componentRepo ComponentRepo
}

func NewRunPlan(config *versource.Config, planRepo PlanRepo, planStore PlanStore, logStore LogStore, tx TransactionManager, newExecutor NewExecutor, componentRepo ComponentRepo) *RunPlan {
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
	var plan *versource.Plan

	err := r.tx.Do(ctx, AdminBranch, "start plan", func(ctx context.Context) error {
		var err error
		plan, err = r.planRepo.GetPlan(ctx, planID)
		if err != nil {
			return fmt.Errorf("failed to get plan: %w", err)
		}

		if plan.ID != planID {
			return fmt.Errorf("plan ID mismatch")
		}

		err = r.planRepo.UpdatePlanState(ctx, planID, versource.TaskStateStarted)
		if err != nil {
			return fmt.Errorf("failed to update plan state: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	var component *versource.Component
	err = r.tx.Checkout(ctx, plan.Changeset.Name, func(ctx context.Context) error {
		var err error
		component, err = r.componentRepo.GetComponentAtCommit(ctx, plan.ComponentID, plan.To)
		return err
	})
	if err != nil {
		return err
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
		return err
	}

	planPath, resourceCounts, err := executor.Plan(ctx)
	if err != nil {
		return err
	}

	err = r.planStore.StorePlan(ctx, planID, planPath)
	if err != nil {
		return err
	}

	err = r.tx.Do(ctx, AdminBranch, "succeed plan", func(ctx context.Context) error {
		updateErr := r.planRepo.UpdatePlanResourceCounts(ctx, planID, resourceCounts)
		if updateErr != nil {
			return fmt.Errorf("failed to update plan resource counts: %w", updateErr)
		}

		err := r.planRepo.UpdatePlanState(ctx, planID, versource.TaskStateSucceeded)
		if err != nil {
			return fmt.Errorf("failed to update plan state: %w", err)
		}

		log.WithField("plan_id", planID).Info("Plan completed successfully")

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
