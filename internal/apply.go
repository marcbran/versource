package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	log "github.com/sirupsen/logrus"
	"gorm.io/datatypes"
)

type TaskState string

const (
	TaskStateQueued    TaskState = "Queued"
	TaskStateStarted   TaskState = "Started"
	TaskStateAborted   TaskState = "Aborted"
	TaskStateCompleted TaskState = "Completed"
	TaskStateFailed    TaskState = "Failed"
	TaskStateCancelled TaskState = "Cancelled"
)

type Apply struct {
	ID          uint      `gorm:"primarykey"`
	Plan        Plan      `gorm:"foreignKey:PlanID"`
	PlanID      uint      `gorm:"uniqueIndex"`
	Changeset   Changeset `gorm:"foreignKey:ChangesetID"`
	ChangesetID uint
	State       string `gorm:"default:Queued"`
}

type ApplyRepo interface {
	CreateApply(ctx context.Context, apply *Apply) error
	GetApply(ctx context.Context, applyID uint) (*Apply, error)
	UpdateApplyState(ctx context.Context, applyID uint, state TaskState) error
	GetQueuedApplies(ctx context.Context) ([]uint, error)
	GetQueuedAppliesByChangeset(ctx context.Context, changesetID uint) ([]uint, error)
	ListApplies(ctx context.Context) ([]Apply, error)
}

type State struct {
	ID          uint           `gorm:"primarykey"`
	Component   Component      `gorm:"foreignKey:ComponentID"`
	ComponentID uint           `gorm:"uniqueIndex"`
	Output      datatypes.JSON `gorm:"type:jsonb"`
}

type StateRepo interface {
	UpsertState(ctx context.Context, state *State) error
}

type Resource struct {
	ID           uint  `gorm:"primarykey"`
	State        State `gorm:"foreignKey:StateID"`
	StateID      uint
	Address      string
	Mode         ResourceMode
	ProviderName string
	Count        *int
	ForEach      *string
	Type         string
	Attributes   datatypes.JSON `gorm:"type:jsonb"`
}

type ResourceMode string

const (
	DataResourceMode    ResourceMode = "data"
	ManagedResourceMode ResourceMode = "managed"
)

type ResourceRepo interface {
	UpsertResources(ctx context.Context, resources []Resource) error
}

type RunApply struct {
	config       *Config
	applyRepo    ApplyRepo
	stateRepo    StateRepo
	resourceRepo ResourceRepo
	planStore    PlanStore
	tx           TransactionManager
}

func NewRunApply(config *Config, applyRepo ApplyRepo, stateRepo StateRepo, resourceRepo ResourceRepo, planStore PlanStore, tx TransactionManager) *RunApply {
	return &RunApply{
		config:       config,
		applyRepo:    applyRepo,
		stateRepo:    stateRepo,
		resourceRepo: resourceRepo,
		planStore:    planStore,
		tx:           tx,
	}
}

func (a *RunApply) Exec(ctx context.Context, applyID uint) error {
	var apply *Apply

	err := a.tx.Do(ctx, "main", "start apply", func(ctx context.Context) error {
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

	component := &apply.Plan.Component
	workDir := a.config.Terraform.WorkDir
	tf, cleanup, err := NewTerraformFromComponent(ctx, component, workDir)
	if err != nil {
		return fmt.Errorf("failed to create terraform from component: %w", err)
	}
	defer cleanup()

	log.Info("Created dynamic component config in temp directory")

	err = tf.Init(ctx)
	if err != nil {
		stateErr := a.tx.Do(ctx, "main", "fail apply", func(ctx context.Context) error {
			return a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to initialize terraform: %w, and failed to update apply state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to initialize terraform: %w", err)
	}

	planPath, err := a.planStore.LoadPlan(ctx, apply.PlanID)
	if err != nil {
		stateErr := a.tx.Do(ctx, "main", "fail apply", func(ctx context.Context) error {
			return a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to load plan: %w, and failed to update apply state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to load plan: %w", err)
	}

	log.WithField("plan_path", planPath).Info("Loaded plan")

	err = tf.Apply(ctx, tfexec.DirOrPlan(planPath))
	if err != nil {
		stateErr := a.tx.Do(ctx, "main", "fail apply", func(ctx context.Context) error {
			return a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to apply terraform: %w, and failed to update apply state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to apply terraform: %w", err)
	}

	log.Info("Terraform apply completed successfully")

	tfState, err := tf.Show(ctx)
	if err != nil {
		return fmt.Errorf("failed to get terraform state: %w", err)
	}

	output := make(map[string]any)
	for name, out := range tfState.Values.Outputs {
		if out == nil {
			continue
		}
		if out.Sensitive {
			continue
		}
		output[name] = out.Value
	}
	jsonOutput, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	state := State{
		ComponentID: component.ID,
		Output:      datatypes.JSON(jsonOutput),
	}

	var resources []Resource
	if tfState.Values != nil && tfState.Values.RootModule != nil {
		resources, err = extractResources(tfState.Values.RootModule)
		if err != nil {
			return fmt.Errorf("failed to extract resources: %w", err)
		}
	}

	err = a.tx.Do(ctx, "main", "complete apply", func(ctx context.Context) error {
		err := a.stateRepo.UpsertState(ctx, &state)
		if err != nil {
			stateErr := a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateFailed)
			if stateErr != nil {
				return fmt.Errorf("failed to upsert state: %w, and failed to update apply state: %w", err, stateErr)
			}
			return fmt.Errorf("failed to upsert state: %w", err)
		}

		for i := range resources {
			resources[i].StateID = state.ID
		}

		if len(resources) > 0 {
			err = a.resourceRepo.UpsertResources(ctx, resources)
			if err != nil {
				stateErr := a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateFailed)
				if stateErr != nil {
					return fmt.Errorf("failed to upsert resources: %w, and failed to update apply state: %w", err, stateErr)
				}
				return fmt.Errorf("failed to upsert resources: %w", err)
			}
		}

		err = a.applyRepo.UpdateApplyState(ctx, applyID, TaskStateCompleted)
		if err != nil {
			return fmt.Errorf("failed to update apply state: %w", err)
		}

		log.WithField("state_id", state.ID).Info("Saved component output")

		return nil
	})

	return err
}

type ListApplies struct {
	applyRepo ApplyRepo
}

func NewListApplies(applyRepo ApplyRepo) *ListApplies {
	return &ListApplies{
		applyRepo: applyRepo,
	}
}

type ListAppliesRequest struct{}

type ListAppliesResponse struct {
	Applies []Apply `json:"applies"`
}

func (l *ListApplies) Exec(ctx context.Context, req ListAppliesRequest) (*ListAppliesResponse, error) {
	applies, err := l.applyRepo.ListApplies(ctx)
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
	applyChan chan uint
}

func NewApplyWorker(runApply *RunApply, applyRepo ApplyRepo) *ApplyWorker {
	return &ApplyWorker{
		runApply:  runApply,
		applyRepo: applyRepo,
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
	applyIDs, err := aw.applyRepo.GetQueuedApplies(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to get queued applies")
		return
	}

	for _, applyID := range applyIDs {
		aw.runApplyInBackground(ctx, applyID)
	}
}

func extractResources(module *tfjson.StateModule) ([]Resource, error) {
	var resources []Resource

	for _, tfResource := range module.Resources {
		var count *int
		var forEach *string
		switch index := tfResource.Index.(type) {
		case int:
			count = &index
		case string:
			forEach = &index
		}
		jsonAttributes, err := json.Marshal(tfResource.AttributeValues)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal attributes: %w", err)
		}
		resource := Resource{
			Address:      tfResource.Address,
			Mode:         ResourceMode(tfResource.Mode),
			ProviderName: tfResource.ProviderName,
			Count:        count,
			ForEach:      forEach,
			Type:         tfResource.Type,
			Attributes:   datatypes.JSON(jsonAttributes),
		}
		resources = append(resources, resource)
	}

	for _, childModule := range module.ChildModules {
		childResources, err := extractResources(childModule)
		if err != nil {
			return nil, fmt.Errorf("failed to extract resources: %w", err)
		}
		resources = append(resources, childResources...)
	}

	return resources, nil
}
