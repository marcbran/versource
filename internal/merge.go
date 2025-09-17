package internal

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Merge struct {
	ID          uint      `gorm:"primarykey"`
	Changeset   Changeset `gorm:"foreignKey:ChangesetID"`
	ChangesetID uint
	MergeBase   string    `gorm:"column:merge_base"`
	Head        string    `gorm:"column:head"`
	State       TaskState `gorm:"default:Queued"`
}

type MergeRepo interface {
	GetMerge(ctx context.Context, mergeID uint) (*Merge, error)
	GetQueuedMerges(ctx context.Context) ([]uint, error)
	GetQueuedMergesByChangeset(ctx context.Context, changesetID uint) ([]uint, error)
	ListMerges(ctx context.Context) ([]Merge, error)
	CreateMerge(ctx context.Context, merge *Merge) error
	UpdateMergeState(ctx context.Context, mergeID uint, state TaskState) error
}

type ListMerges struct {
	mergeRepo MergeRepo
	tx        TransactionManager
}

func NewListMerges(mergeRepo MergeRepo, tx TransactionManager) *ListMerges {
	return &ListMerges{
		mergeRepo: mergeRepo,
		tx:        tx,
	}
}

type ListMergesRequest struct{}

type ListMergesResponse struct {
	Merges []Merge `json:"merges"`
}

func (l *ListMerges) Exec(ctx context.Context, req ListMergesRequest) (*ListMergesResponse, error) {
	var merges []Merge
	err := l.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		merges, err = l.mergeRepo.ListMerges(ctx)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list merges", err)
	}

	return &ListMergesResponse{
		Merges: merges,
	}, nil
}

type CreateMerge struct {
	changesetRepo ChangesetRepo
	mergeRepo     MergeRepo
	tx            TransactionManager
	mergeWorker   *MergeWorker
}

func NewCreateMerge(changesetRepo ChangesetRepo, mergeRepo MergeRepo, tx TransactionManager, mergeWorker *MergeWorker) *CreateMerge {
	return &CreateMerge{
		changesetRepo: changesetRepo,
		mergeRepo:     mergeRepo,
		tx:            tx,
		mergeWorker:   mergeWorker,
	}
}

type CreateMergeRequest struct {
	ChangesetName string `json:"changeset_name"`
}

type CreateMergeResponse struct {
	ID          uint      `json:"id"`
	ChangesetID uint      `json:"changeset_id"`
	MergeBase   string    `json:"merge_base"`
	Head        string    `json:"head"`
	State       TaskState `json:"state"`
}

func (c *CreateMerge) Exec(ctx context.Context, req CreateMergeRequest) (*CreateMergeResponse, error) {
	if req.ChangesetName == "" {
		return nil, UserErr("changeset name is required")
	}

	var response *CreateMergeResponse
	err := c.tx.Do(ctx, MainBranch, "create merge", func(ctx context.Context) error {
		changeset, err := c.changesetRepo.GetChangesetByName(ctx, req.ChangesetName)
		if err != nil {
			return UserErrE("changeset not found", err)
		}

		mergeBase, err := c.tx.GetMergeBase(ctx, MainBranch, changeset.Name)
		if err != nil {
			return InternalErrE("failed to get merge base", err)
		}

		head, err := c.tx.GetBranchHead(ctx, changeset.Name)
		if err != nil {
			return InternalErrE("failed to get head of changeset branch", err)
		}

		merge := &Merge{
			ChangesetID: changeset.ID,
			Changeset:   *changeset,
			MergeBase:   mergeBase,
			Head:        head,
		}

		err = c.mergeRepo.CreateMerge(ctx, merge)
		if err != nil {
			return InternalErrE("failed to create merge", err)
		}

		response = &CreateMergeResponse{
			ID:          merge.ID,
			ChangesetID: merge.ChangesetID,
			MergeBase:   merge.MergeBase,
			Head:        merge.Head,
			State:       merge.State,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if c.mergeWorker != nil {
		c.mergeWorker.QueueMerge(response.ID)
	}

	return response, nil
}

type MergeWorker struct {
	runMerge  *RunMerge
	mergeRepo MergeRepo
	mergeChan chan uint
}

func NewMergeWorker(runMerge *RunMerge, mergeRepo MergeRepo) *MergeWorker {
	return &MergeWorker{
		runMerge:  runMerge,
		mergeRepo: mergeRepo,
		mergeChan: make(chan uint, 100),
	}
}

func (mw *MergeWorker) Start(ctx context.Context) {
	go mw.processMerges(ctx)
}

func (mw *MergeWorker) QueueMerge(mergeID uint) {
	select {
	case mw.mergeChan <- mergeID:
		log.WithField("merge_id", mergeID).Debug("Queued merge for processing")
	default:
		log.WithField("merge_id", mergeID).Warn("Merge channel full, merge will be picked up by polling")
	}
}

func (mw *MergeWorker) processMerges(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case mergeID := <-mw.mergeChan:
			mw.runMergeInBackground(ctx, mergeID)
		case <-ticker.C:
			mw.processQueuedMerges(ctx)
		}
	}
}

func (mw *MergeWorker) runMergeInBackground(ctx context.Context, mergeID uint) {
	go func() {
		workerCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
		defer cancel()

		err := mw.runMerge.Exec(workerCtx, mergeID)
		if err != nil {
			log.WithError(err).WithField("merge_id", mergeID).Error("Failed to run merge")
		} else {
			log.WithField("merge_id", mergeID).Info("Merge completed successfully")
		}
	}()
}

func (mw *MergeWorker) processQueuedMerges(ctx context.Context) {
	mergeIDs, err := mw.mergeRepo.GetQueuedMerges(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to get queued merges")
		return
	}

	for _, mergeID := range mergeIDs {
		mw.runMergeInBackground(ctx, mergeID)
	}
}

type RunMerge struct {
	config             *Config
	mergeRepo          MergeRepo
	changesetRepo      ChangesetRepo
	planRepo           PlanRepo
	planStore          PlanStore
	logStore           LogStore
	tx                 TransactionManager
	listComponentDiffs *ListComponentDiffs
}

func NewRunMerge(config *Config, mergeRepo MergeRepo, changesetRepo ChangesetRepo, planRepo PlanRepo, planStore PlanStore, logStore LogStore, tx TransactionManager, listComponentDiffs *ListComponentDiffs) *RunMerge {
	return &RunMerge{
		config:             config,
		mergeRepo:          mergeRepo,
		changesetRepo:      changesetRepo,
		planRepo:           planRepo,
		planStore:          planStore,
		logStore:           logStore,
		tx:                 tx,
		listComponentDiffs: listComponentDiffs,
	}
}

func (r *RunMerge) Exec(ctx context.Context, mergeID uint) error {
	var merge *Merge

	err := r.tx.Do(ctx, MainBranch, "start merge", func(ctx context.Context) error {
		var err error
		merge, err = r.mergeRepo.GetMerge(ctx, mergeID)
		if err != nil {
			return fmt.Errorf("failed to get merge: %w", err)
		}

		if merge.ID != mergeID {
			return fmt.Errorf("merge ID mismatch")
		}

		err = r.mergeRepo.UpdateMergeState(ctx, mergeID, TaskStateStarted)
		if err != nil {
			return fmt.Errorf("failed to update merge state: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Info("Starting merge preparation")

	var canMerge bool
	err = r.tx.Do(ctx, merge.Changeset.Name, "prepare merge", func(ctx context.Context) error {
		var err error
		canMerge, err = r.prepareMerge(ctx, merge)
		return err
	})
	if err != nil {
		stateErr := r.tx.Do(ctx, MainBranch, "fail merge preparation", func(ctx context.Context) error {
			return r.mergeRepo.UpdateMergeState(ctx, mergeID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("merge preparation failed: %w, and failed to update merge state: %w", err, stateErr)
		}
		return fmt.Errorf("merge preparation failed: %w", err)
	}

	if !canMerge {
		log.Info("Merge validation failed, marking changeset as rejected")
		err = r.tx.Do(ctx, MainBranch, "reject changeset", func(ctx context.Context) error {
			err = r.mergeRepo.UpdateMergeState(ctx, mergeID, TaskStateFailed)
			if err != nil {
				return fmt.Errorf("failed to update merge state: %w", err)
			}
			err = r.changesetRepo.UpdateChangesetReviewState(ctx, merge.ChangesetID, ChangesetReviewStateRejected)
			if err != nil {
				return fmt.Errorf("failed to update changeset review state: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to reject changeset: %w", err)
		}
		return nil
	}

	log.Info("Merge preparation completed, starting merge operation")

	err = r.tx.MergeBranch(ctx, merge.Changeset.Name)
	if err != nil {
		stateErr := r.tx.Do(ctx, MainBranch, "fail merge", func(ctx context.Context) error {
			return r.mergeRepo.UpdateMergeState(ctx, mergeID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("failed to merge changeset: %w, and failed to update merge state: %w", err, stateErr)
		}
		return fmt.Errorf("failed to merge changeset: %w", err)
	}

	log.Info("Changeset merge completed successfully")

	err = r.tx.Do(ctx, MainBranch, "complete merge", func(ctx context.Context) error {
		err = r.changesetRepo.UpdateChangesetState(ctx, merge.ChangesetID, ChangesetStateMerged)
		if err != nil {
			stateErr := r.tx.Do(ctx, MainBranch, "fail changeset merge", func(ctx context.Context) error {
				return r.mergeRepo.UpdateMergeState(ctx, mergeID, TaskStateFailed)
			})
			if stateErr != nil {
				return fmt.Errorf("failed to update changeset state: %w, and failed to update merge state: %w", err, stateErr)
			}
			return fmt.Errorf("failed to update changeset state: %w", err)
		}
		err = r.mergeRepo.UpdateMergeState(ctx, mergeID, TaskStateSucceeded)
		if err != nil {
			return fmt.Errorf("failed to update merge state: %w", err)
		}

		log.WithField("merge_id", mergeID).WithField("changeset_id", merge.ChangesetID).Info("Merge completed and changeset marked as merged")

		return nil
	})

	return err
}

func (r *RunMerge) prepareMerge(ctx context.Context, merge *Merge) (bool, error) {
	changesetName := merge.Changeset.Name

	diffsResp, err := r.listComponentDiffs.Exec(ctx, ListComponentDiffsRequest{
		Changeset: changesetName,
	})
	if err != nil {
		return false, fmt.Errorf("failed to list component diffs: %w", err)
	}

	canMerge, err := r.validateMerge(ctx, merge, diffsResp.Diffs)
	if err != nil {
		return false, err
	}
	if !canMerge {
		return false, nil
	}

	err = r.cleanupBeforeMerge(ctx, merge, diffsResp.Diffs)
	if err != nil {
		return false, fmt.Errorf("failed to cleanup orphaned plans: %w", err)
	}

	return true, nil
}

func (r *RunMerge) validateMerge(ctx context.Context, merge *Merge, diffs []ComponentDiff) (bool, error) {
	changesetName := merge.Changeset.Name

	hasCommitsAfter, err := r.tx.HasCommitsAfter(ctx, changesetName, merge.Head)
	if err != nil {
		return false, fmt.Errorf("failed to check commits after head: %w", err)
	}
	if hasCommitsAfter {
		return false, nil
	}

	currentMergeBase, err := r.tx.GetMergeBase(ctx, MainBranch, changesetName)
	if err != nil {
		return false, fmt.Errorf("failed to get current merge base: %w", err)
	}
	if currentMergeBase != merge.MergeBase {
		return false, nil
	}

	for _, diff := range diffs {
		if diff.Plan == nil {
			return false, nil
		}
		if diff.Plan.State != TaskStateSucceeded {
			return false, nil
		}
	}

	return true, nil
}

func (r *RunMerge) cleanupBeforeMerge(ctx context.Context, merge *Merge, diffs []ComponentDiff) error {
	validPlanIDs := make(map[uint]bool)
	for _, diff := range diffs {
		if diff.Plan != nil {
			validPlanIDs[diff.Plan.ID] = true
		}
	}

	allPlans, err := r.planRepo.ListPlansByChangeset(ctx, merge.ChangesetID)
	if err != nil {
		return fmt.Errorf("failed to list plans for changeset: %w", err)
	}

	for _, plan := range allPlans {
		if !validPlanIDs[plan.ID] {
			err = r.planRepo.DeletePlan(ctx, plan.ID)
			if err != nil {
				return fmt.Errorf("failed to delete orphaned plan %d from repo: %w", plan.ID, err)
			}

			err = r.planStore.DeletePlan(ctx, plan.ID)
			if err != nil {
				return fmt.Errorf("failed to delete orphaned plan %d from store: %w", plan.ID, err)
			}

			err = r.logStore.DeleteLog(ctx, "plan", plan.ID)
			if err != nil {
				return fmt.Errorf("failed to delete orphaned plan %d logs: %w", plan.ID, err)
			}
		}
	}

	return nil
}
