package internal

import (
	"context"
	"fmt"
	"io"
	"time"

	log "github.com/sirupsen/logrus"
)

type TaskState string

const (
	TaskStateQueued    TaskState = "Queued"
	TaskStateStarted   TaskState = "Started"
	TaskStateAborted   TaskState = "Aborted"
	TaskStateSucceeded TaskState = "Succeeded"
	TaskStateFailed    TaskState = "Failed"
	TaskStateCancelled TaskState = "Cancelled"
)

func IsTaskCompleted(task TaskState) bool {
	return task == TaskStateSucceeded || task == TaskStateFailed || task == TaskStateCancelled
}

type Apply struct {
	ID          uint      `gorm:"primarykey" json:"id" yaml:"id"`
	Plan        Plan      `gorm:"foreignKey:PlanID" json:"plan" yaml:"plan"`
	PlanID      uint      `gorm:"uniqueIndex" json:"planId" yaml:"planId"`
	Changeset   Changeset `gorm:"foreignKey:ChangesetID" json:"changeset" yaml:"changeset"`
	ChangesetID uint      `json:"changesetId" yaml:"changesetId"`
	State       TaskState `gorm:"default:Queued" json:"state" yaml:"state"`
}

type ApplyRepo interface {
	GetApply(ctx context.Context, applyID uint) (*Apply, error)
	GetQueuedApplies(ctx context.Context) ([]uint, error)
	GetQueuedAppliesByChangeset(ctx context.Context, changesetID uint) ([]uint, error)
	ListApplies(ctx context.Context) ([]Apply, error)
	ListAppliesByChangeset(ctx context.Context, changesetID uint) ([]Apply, error)
	CreateApply(ctx context.Context, apply *Apply) error
	UpdateApplyState(ctx context.Context, applyID uint, state TaskState) error
}

type GetApply struct {
	applyRepo     ApplyRepo
	componentRepo ComponentRepo
	tx            TransactionManager
}

func NewGetApply(applyRepo ApplyRepo, componentRepo ComponentRepo, tx TransactionManager) *GetApply {
	return &GetApply{
		applyRepo:     applyRepo,
		componentRepo: componentRepo,
		tx:            tx,
	}
}

type GetApplyRequest struct {
	ApplyID uint `json:"applyId" yaml:"applyId"`
}

type GetApplyResponse struct {
	ID          uint      `json:"id" yaml:"id"`
	PlanID      uint      `json:"planId" yaml:"planId"`
	ChangesetID uint      `json:"changesetId" yaml:"changesetId"`
	State       TaskState `json:"state" yaml:"state"`
	Plan        struct {
		ID          uint      `json:"id" yaml:"id"`
		State       TaskState `json:"state" yaml:"state"`
		From        string    `json:"from" yaml:"from"`
		To          string    `json:"to" yaml:"to"`
		Add         *int      `json:"add,omitempty" yaml:"add,omitempty"`
		Change      *int      `json:"change,omitempty" yaml:"change,omitempty"`
		Destroy     *int      `json:"destroy,omitempty" yaml:"destroy,omitempty"`
		ComponentID uint      `json:"componentId" yaml:"componentId"`
		Component   struct {
			ID   uint   `json:"id" yaml:"id"`
			Name string `json:"name" yaml:"name"`
		} `json:"component" yaml:"component"`
		ChangesetID uint `json:"changesetId" yaml:"changesetId"`
		Changeset   struct {
			ID   uint   `json:"id" yaml:"id"`
			Name string `json:"name" yaml:"name"`
		} `json:"changeset" yaml:"changeset"`
	} `json:"plan" yaml:"plan"`
	Changeset struct {
		ID   uint   `json:"id" yaml:"id"`
		Name string `json:"name" yaml:"name"`
	} `json:"changeset" yaml:"changeset"`
}

func (g *GetApply) Exec(ctx context.Context, req GetApplyRequest) (*GetApplyResponse, error) {
	if req.ApplyID == 0 {
		return nil, UserErr("apply ID is required")
	}

	var apply *Apply
	err := g.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		apply, err = g.applyRepo.GetApply(ctx, req.ApplyID)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to get apply", err)
	}

	var component *Component
	branch := apply.Plan.Changeset.Name
	if apply.Plan.Changeset.State == ChangesetStateMerged {
		branch = MainBranch
	}

	err = g.tx.Checkout(ctx, branch, func(ctx context.Context) error {
		var err error
		component, err = g.componentRepo.GetComponentAtCommit(ctx, apply.Plan.ComponentID, apply.Plan.To)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to get component at commit", err)
	}

	response := &GetApplyResponse{
		ID:          apply.ID,
		PlanID:      apply.PlanID,
		ChangesetID: apply.ChangesetID,
		State:       apply.State,
		Plan: struct {
			ID          uint      `json:"id" yaml:"id"`
			State       TaskState `json:"state" yaml:"state"`
			From        string    `json:"from" yaml:"from"`
			To          string    `json:"to" yaml:"to"`
			Add         *int      `json:"add,omitempty" yaml:"add,omitempty"`
			Change      *int      `json:"change,omitempty" yaml:"change,omitempty"`
			Destroy     *int      `json:"destroy,omitempty" yaml:"destroy,omitempty"`
			ComponentID uint      `json:"componentId" yaml:"componentId"`
			Component   struct {
				ID   uint   `json:"id" yaml:"id"`
				Name string `json:"name" yaml:"name"`
			} `json:"component" yaml:"component"`
			ChangesetID uint `json:"changesetId" yaml:"changesetId"`
			Changeset   struct {
				ID   uint   `json:"id" yaml:"id"`
				Name string `json:"name" yaml:"name"`
			} `json:"changeset" yaml:"changeset"`
		}{
			ID:          apply.Plan.ID,
			State:       apply.Plan.State,
			From:        apply.Plan.From,
			To:          apply.Plan.To,
			Add:         apply.Plan.Add,
			Change:      apply.Plan.Change,
			Destroy:     apply.Plan.Destroy,
			ComponentID: apply.Plan.ComponentID,
			Component: struct {
				ID   uint   `json:"id" yaml:"id"`
				Name string `json:"name" yaml:"name"`
			}{
				ID:   component.ID,
				Name: component.Name,
			},
			ChangesetID: apply.Plan.ChangesetID,
			Changeset: struct {
				ID   uint   `json:"id" yaml:"id"`
				Name string `json:"name" yaml:"name"`
			}{
				ID:   apply.Plan.Changeset.ID,
				Name: apply.Plan.Changeset.Name,
			},
		},
		Changeset: struct {
			ID   uint   `json:"id" yaml:"id"`
			Name string `json:"name" yaml:"name"`
		}{
			ID:   apply.Changeset.ID,
			Name: apply.Changeset.Name,
		},
	}

	return response, nil
}

type GetApplyLog struct {
	logStore LogStore
}

func NewGetApplyLog(logStore LogStore) *GetApplyLog {
	return &GetApplyLog{
		logStore: logStore,
	}
}

type GetApplyLogRequest struct {
	ApplyID uint `json:"applyId" yaml:"applyId"`
}

type GetApplyLogResponse struct {
	Content io.ReadCloser `json:"content" yaml:"content"`
}

func (g *GetApplyLog) Exec(ctx context.Context, req GetApplyLogRequest) (*GetApplyLogResponse, error) {
	if req.ApplyID == 0 {
		return nil, UserErr("apply ID is required")
	}

	reader, err := g.logStore.LoadLog(ctx, "apply", req.ApplyID)
	if err != nil {
		return nil, InternalErrE("failed to load apply log", err)
	}

	return &GetApplyLogResponse{
		Content: reader,
	}, nil
}

type ListApplies struct {
	applyRepo ApplyRepo
	tx        TransactionManager
}

func NewListApplies(applyRepo ApplyRepo, tx TransactionManager) *ListApplies {
	return &ListApplies{
		applyRepo: applyRepo,
		tx:        tx,
	}
}

type ListAppliesRequest struct{}

type ListAppliesResponse struct {
	Applies []Apply `json:"applies" yaml:"applies"`
}

func (l *ListApplies) Exec(ctx context.Context, req ListAppliesRequest) (*ListAppliesResponse, error) {
	var applies []Apply
	err := l.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		applies, err = l.applyRepo.ListApplies(ctx)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list applies", err)
	}

	return &ListAppliesResponse{
		Applies: applies,
	}, nil
}

type ApplyWorker struct {
	runApply  *RunApply
	applyRepo ApplyRepo
	tx        TransactionManager
	applyChan chan uint
}

func NewApplyWorker(runApply *RunApply, applyRepo ApplyRepo, tx TransactionManager) *ApplyWorker {
	return &ApplyWorker{
		runApply:  runApply,
		applyRepo: applyRepo,
		tx:        tx,
		applyChan: make(chan uint, 100),
	}
}

func (aw *ApplyWorker) Start(ctx context.Context) {
	go aw.processApplies(ctx)
}

func (aw *ApplyWorker) QueueApply(applyID uint) {
	select {
	case aw.applyChan <- applyID:
		log.WithField("apply_id", applyID).Debug("Queued apply for processing")
	default:
		log.WithField("apply_id", applyID).Warn("Apply channel full, apply will be picked up by polling")
	}
}

func (aw *ApplyWorker) processApplies(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case applyID := <-aw.applyChan:
			aw.runApplyInBackground(ctx, applyID)
		case <-ticker.C:
			aw.processQueuedApplies(ctx)
		}
	}
}

func (aw *ApplyWorker) runApplyInBackground(ctx context.Context, applyID uint) {
	go func() {
		workerCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
		defer cancel()

		err := aw.runApply.Exec(workerCtx, applyID)
		if err != nil {
			log.WithError(err).WithField("apply_id", applyID).Error("Failed to run apply")
		} else {
			log.WithField("apply_id", applyID).Info("Apply completed successfully")
		}
	}()
}

func (aw *ApplyWorker) processQueuedApplies(ctx context.Context) {
	var applyIDs []uint
	err := aw.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		applyIDs, err = aw.applyRepo.GetQueuedApplies(ctx)
		return err
	})
	if err != nil {
		log.WithError(err).Error("Failed to get queued applies")
		return
	}

	for _, applyID := range applyIDs {
		aw.runApplyInBackground(ctx, applyID)
	}
}

type RunApply struct {
	config            *Config
	applyRepo         ApplyRepo
	stateRepo         StateRepo
	stateResourceRepo StateResourceRepo
	resourceRepo      ResourceRepo
	planStore         PlanStore
	logStore          LogStore
	tx                TransactionManager
	newExecutor       NewExecutor
	componentRepo     ComponentRepo
}

func NewRunApply(config *Config, applyRepo ApplyRepo, stateRepo StateRepo, stateResourceRepo StateResourceRepo, resourceRepo ResourceRepo, planStore PlanStore, logStore LogStore, tx TransactionManager, newExecutor NewExecutor, componentRepo ComponentRepo) *RunApply {
	return &RunApply{
		config:            config,
		applyRepo:         applyRepo,
		stateRepo:         stateRepo,
		stateResourceRepo: stateResourceRepo,
		resourceRepo:      resourceRepo,
		planStore:         planStore,
		logStore:          logStore,
		tx:                tx,
		newExecutor:       newExecutor,
		componentRepo:     componentRepo,
	}
}

func (a *RunApply) Exec(ctx context.Context, applyID uint) error {
	var apply *Apply

	err := a.tx.Do(ctx, AdminBranch, "start apply", func(ctx context.Context) error {
		var err error
		apply, err = a.applyRepo.GetApply(ctx, applyID)
		if err != nil {
			return fmt.Errorf("failed to get apply: %w", err)
		}

		if apply.ID != applyID {
			return fmt.Errorf("apply ID mismatch")
		}

		err = a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateStarted)
		if err != nil {
			return fmt.Errorf("failed to update apply state: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	logWriter, err := a.logStore.NewLogWriter("apply", applyID)
	if err != nil {
		return fmt.Errorf("failed to create log writer: %w", err)
	}
	defer logWriter.Close()

	var component *Component
	err = a.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		component, err = a.componentRepo.GetComponentAtCommit(ctx, apply.Plan.ComponentID, apply.Plan.To)
		return err
	})
	if err != nil {
		stateErr := a.tx.Do(ctx, AdminBranch, "fail apply", func(ctx context.Context) error {
			return a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to get component at commit: %w, and failed to update apply state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to get component at commit: %w", err)
	}

	workDir := a.config.Terraform.WorkDir
	executor, err := a.newExecutor(component, workDir, logWriter)
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}
	defer executor.Close()

	log.Info("Created dynamic component config in temp directory")

	err = executor.Init(ctx)
	if err != nil {
		stateErr := a.tx.Do(ctx, AdminBranch, "fail apply", func(ctx context.Context) error {
			return a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to initialize terraform: %w, and failed to update apply state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to initialize terraform: %w", err)
	}

	planPath, err := a.planStore.LoadPlan(ctx, apply.PlanID)
	if err != nil {
		stateErr := a.tx.Do(ctx, AdminBranch, "fail apply", func(ctx context.Context) error {
			return a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to load plan: %w, and failed to update apply state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to load plan: %w", err)
	}

	log.WithField("plan_path", planPath).Info("Loaded plan")

	state, stateResources, err := executor.Apply(ctx, planPath)
	if err != nil {
		stateErr := a.tx.Do(ctx, AdminBranch, "fail apply", func(ctx context.Context) error {
			return a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to apply terraform: %w, and failed to update apply state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to apply terraform: %w", err)
	}

	log.Info("Terraform apply completed successfully")

	err = a.tx.Do(ctx, AdminBranch, "complete apply", func(ctx context.Context) error {
		state.ComponentID = component.ID

		err := a.stateRepo.UpsertState(ctx, &state)
		if err != nil {
			stateErr := a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateFailed)
			if stateErr != nil {
				return fmt.Errorf("failed to upsert state: %w, and failed to update apply state: %w", err, stateErr)
			}
			return fmt.Errorf("failed to upsert state: %w", err)
		}

		if len(stateResources) > 0 {
			var resources []Resource
			for i := range stateResources {
				stateResources[i].StateID = state.ID
				resources = append(resources, stateResources[i].Resource)
			}

			err = a.resourceRepo.UpsertResources(ctx, resources)
			if err != nil {
				stateErr := a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateFailed)
				if stateErr != nil {
					return fmt.Errorf("failed to upsert resources: %w, and failed to update apply state: %w", err, stateErr)
				}
				return fmt.Errorf("failed to upsert resources: %w", err)
			}

			for i := range stateResources {
				stateResources[i].ResourceID = resources[i].ID
			}

			err = a.stateResourceRepo.UpsertStateResources(ctx, stateResources)
			if err != nil {
				stateErr := a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateFailed)
				if stateErr != nil {
					return fmt.Errorf("failed to upsert state resources: %w, and failed to update apply state: %w", err, stateErr)
				}
				return fmt.Errorf("failed to upsert state resources: %w", err)
			}
		}

		err = a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateSucceeded)
		if err != nil {
			return fmt.Errorf("failed to update apply state: %w", err)
		}

		log.WithField("state_id", state.ID).Info("Saved component output")

		return nil
	})

	return err
}
