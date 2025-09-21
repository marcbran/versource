package internal

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Rebase struct {
	ID          uint      `gorm:"primarykey"`
	Changeset   Changeset `gorm:"foreignKey:ChangesetID"`
	ChangesetID uint
	MergeBase   string    `gorm:"column:merge_base"`
	Head        string    `gorm:"column:head"`
	State       TaskState `gorm:"default:Queued"`
}

type RebaseRepo interface {
	GetRebase(ctx context.Context, rebaseID uint) (*Rebase, error)
	GetQueuedRebases(ctx context.Context) ([]uint, error)
	GetQueuedRebasesByChangeset(ctx context.Context, changesetID uint) ([]uint, error)
	ListRebases(ctx context.Context) ([]Rebase, error)
	CreateRebase(ctx context.Context, rebase *Rebase) error
	UpdateRebaseState(ctx context.Context, rebaseID uint, state TaskState) error
}

type GetRebase struct {
	rebaseRepo RebaseRepo
	tx         TransactionManager
}

func NewGetRebase(rebaseRepo RebaseRepo, tx TransactionManager) *GetRebase {
	return &GetRebase{
		rebaseRepo: rebaseRepo,
		tx:         tx,
	}
}

type GetRebaseRequest struct {
	RebaseID      uint   `json:"rebase_id"`
	ChangesetName string `json:"changeset_name"`
}

type GetRebaseResponse struct {
	ID          uint      `json:"id"`
	ChangesetID uint      `json:"changeset_id"`
	MergeBase   string    `json:"merge_base"`
	Head        string    `json:"head"`
	State       TaskState `json:"state"`
}

func (g *GetRebase) Exec(ctx context.Context, req GetRebaseRequest) (*GetRebaseResponse, error) {
	if req.ChangesetName == "" {
		return nil, UserErr("changeset name is required")
	}

	var rebase *Rebase
	err := g.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		rebase, err = g.rebaseRepo.GetRebase(ctx, req.RebaseID)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to get rebase", err)
	}

	if rebase == nil {
		return nil, UserErr("rebase not found")
	}

	return &GetRebaseResponse{
		ID:          rebase.ID,
		ChangesetID: rebase.ChangesetID,
		MergeBase:   rebase.MergeBase,
		Head:        rebase.Head,
		State:       rebase.State,
	}, nil
}

type ListRebases struct {
	rebaseRepo RebaseRepo
	tx         TransactionManager
}

func NewListRebases(rebaseRepo RebaseRepo, tx TransactionManager) *ListRebases {
	return &ListRebases{
		rebaseRepo: rebaseRepo,
		tx:         tx,
	}
}

type ListRebasesRequest struct {
	ChangesetName string `json:"changeset_name"`
}

type ListRebasesResponse struct {
	Rebases []Rebase `json:"rebases"`
}

func (l *ListRebases) Exec(ctx context.Context, req ListRebasesRequest) (*ListRebasesResponse, error) {
	if req.ChangesetName == "" {
		return nil, UserErr("changeset name is required")
	}

	var rebases []Rebase
	err := l.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		rebases, err = l.rebaseRepo.ListRebases(ctx)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list rebases", err)
	}

	return &ListRebasesResponse{
		Rebases: rebases,
	}, nil
}

type CreateRebase struct {
	changesetRepo ChangesetRepo
	rebaseRepo    RebaseRepo
	tx            TransactionManager
	rebaseWorker  *RebaseWorker
}

func NewCreateRebase(changesetRepo ChangesetRepo, rebaseRepo RebaseRepo, tx TransactionManager, rebaseWorker *RebaseWorker) *CreateRebase {
	return &CreateRebase{
		changesetRepo: changesetRepo,
		rebaseRepo:    rebaseRepo,
		tx:            tx,
		rebaseWorker:  rebaseWorker,
	}
}

type CreateRebaseRequest struct {
	ChangesetName string `json:"changeset_name"`
}

type CreateRebaseResponse struct {
	ID          uint      `json:"id"`
	ChangesetID uint      `json:"changeset_id"`
	MergeBase   string    `json:"merge_base"`
	Head        string    `json:"head"`
	State       TaskState `json:"state"`
}

func (c *CreateRebase) Exec(ctx context.Context, req CreateRebaseRequest) (*CreateRebaseResponse, error) {
	if req.ChangesetName == "" {
		return nil, UserErr("changeset name is required")
	}

	var mergeBase string
	var head string

	err := c.tx.Checkout(ctx, req.ChangesetName, func(ctx context.Context) error {
		var err error
		mergeBase, err = c.tx.GetMergeBase(ctx, MainBranch, req.ChangesetName)
		if err != nil {
			return InternalErrE("failed to get merge base", err)
		}

		head, err = c.tx.GetHead(ctx)
		if err != nil {
			return InternalErrE("failed to get head", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	var response *CreateRebaseResponse
	err = c.tx.Do(ctx, AdminBranch, "create rebase", func(ctx context.Context) error {
		changeset, err := c.changesetRepo.GetChangesetByName(ctx, req.ChangesetName)
		if err != nil {
			return UserErrE("changeset not found", err)
		}

		rebase := &Rebase{
			ChangesetID: changeset.ID,
			Changeset:   *changeset,
			MergeBase:   mergeBase,
			Head:        head,
		}

		err = c.rebaseRepo.CreateRebase(ctx, rebase)
		if err != nil {
			return InternalErrE("failed to create rebase", err)
		}

		response = &CreateRebaseResponse{
			ID:          rebase.ID,
			ChangesetID: rebase.ChangesetID,
			MergeBase:   rebase.MergeBase,
			Head:        rebase.Head,
			State:       rebase.State,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if c.rebaseWorker != nil {
		c.rebaseWorker.QueueRebase(response.ID)
	}

	return response, nil
}

type RebaseWorker struct {
	runRebase  *RunRebase
	rebaseRepo RebaseRepo
	rebaseChan chan uint
}

func NewRebaseWorker(runRebase *RunRebase, rebaseRepo RebaseRepo) *RebaseWorker {
	return &RebaseWorker{
		runRebase:  runRebase,
		rebaseRepo: rebaseRepo,
		rebaseChan: make(chan uint, 100),
	}
}

func (rw *RebaseWorker) Start(ctx context.Context) {
	go rw.processRebases(ctx)
}

func (rw *RebaseWorker) QueueRebase(rebaseID uint) {
	select {
	case rw.rebaseChan <- rebaseID:
		log.WithField("rebase_id", rebaseID).Debug("Queued rebase for processing")
	default:
		log.WithField("rebase_id", rebaseID).Warn("Rebase channel full, rebase will be picked up by polling")
	}
}

func (rw *RebaseWorker) processRebases(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case rebaseID := <-rw.rebaseChan:
			rw.runRebaseInBackground(ctx, rebaseID)
		case <-ticker.C:
			rw.processQueuedRebases(ctx)
		}
	}
}

func (rw *RebaseWorker) runRebaseInBackground(ctx context.Context, rebaseID uint) {
	go func() {
		workerCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
		defer cancel()

		err := rw.runRebase.Exec(workerCtx, rebaseID)
		if err != nil {
			log.WithError(err).WithField("rebase_id", rebaseID).Error("Failed to run rebase")
		} else {
			log.WithField("rebase_id", rebaseID).Info("Rebase completed")
		}
	}()
}

func (rw *RebaseWorker) processQueuedRebases(ctx context.Context) {
	rebaseIDs, err := rw.rebaseRepo.GetQueuedRebases(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to get queued rebases")
		return
	}

	for _, rebaseID := range rebaseIDs {
		rw.runRebaseInBackground(ctx, rebaseID)
	}
}

type RunRebase struct {
	config        *Config
	rebaseRepo    RebaseRepo
	changesetRepo ChangesetRepo
	tx            TransactionManager
}

func NewRunRebase(config *Config, rebaseRepo RebaseRepo, changesetRepo ChangesetRepo, tx TransactionManager) *RunRebase {
	return &RunRebase{
		config:        config,
		rebaseRepo:    rebaseRepo,
		changesetRepo: changesetRepo,
		tx:            tx,
	}
}

func (r *RunRebase) Exec(ctx context.Context, rebaseID uint) error {
	var rebase *Rebase

	err := r.tx.Do(ctx, AdminBranch, "start rebase", func(ctx context.Context) error {
		var err error
		rebase, err = r.rebaseRepo.GetRebase(ctx, rebaseID)
		if err != nil {
			return fmt.Errorf("failed to get rebase: %w", err)
		}

		if rebase.ID != rebaseID {
			return fmt.Errorf("rebase ID mismatch")
		}

		err = r.rebaseRepo.UpdateRebaseState(ctx, rebaseID, TaskStateStarted)
		if err != nil {
			return fmt.Errorf("failed to update rebase state: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Info("Starting rebase operation")

	err = r.tx.Checkout(ctx, rebase.Changeset.Name, func(ctx context.Context) error {
		err = r.tx.RebaseBranch(ctx, MainBranch)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		stateErr := r.tx.Do(ctx, AdminBranch, "fail rebase", func(ctx context.Context) error {
			return r.rebaseRepo.UpdateRebaseState(ctx, rebaseID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("rebase failed: %w, and failed to update rebase state: %w", err, stateErr)
		}
		return fmt.Errorf("rebase failed: %w", err)
	}

	err = r.tx.Do(ctx, AdminBranch, "complete rebase", func(ctx context.Context) error {
		err = r.rebaseRepo.UpdateRebaseState(ctx, rebaseID, TaskStateSucceeded)
		if err != nil {
			return fmt.Errorf("failed to update rebase state: %w", err)
		}

		log.WithField("rebase_id", rebaseID).WithField("changeset_id", rebase.ChangesetID).Info("Rebase completed")

		return nil
	})
	if err != nil {
		stateErr := r.tx.Do(ctx, AdminBranch, "fail rebase completion", func(ctx context.Context) error {
			return r.rebaseRepo.UpdateRebaseState(ctx, rebaseID, TaskStateFailed)
		})
		if stateErr != nil {
			return fmt.Errorf("rebase completion failed: %w, and failed to update rebase state: %w", err, stateErr)
		}
		return fmt.Errorf("rebase completion failed: %w", err)
	}

	return nil
}
