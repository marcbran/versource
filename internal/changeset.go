package internal

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type ChangesetState string

const (
	ChangesetStateOpen   ChangesetState = "Open"
	ChangesetStateClosed ChangesetState = "Closed"
	ChangesetStateMerged ChangesetState = "Merged"
)

type ChangesetReviewState string

const (
	ChangesetReviewStateDraft    ChangesetReviewState = "Draft"
	ChangesetReviewStatePending  ChangesetReviewState = "Pending"
	ChangesetReviewStateApproved ChangesetReviewState = "Approved"
	ChangesetReviewStateRejected ChangesetReviewState = "Rejected"
)

type Changeset struct {
	ID          uint                 `gorm:"primarykey"`
	Name        string               `gorm:"index"`
	State       ChangesetState       `gorm:"default:Open"`
	ReviewState ChangesetReviewState `gorm:"default:Draft"`
}

type ChangesetRepo interface {
	GetChangeset(ctx context.Context, changesetID uint) (*Changeset, error)
	GetChangesetByName(ctx context.Context, name string) (*Changeset, error)
	GetOpenChangesetByName(ctx context.Context, name string) (*Changeset, error)
	ListChangesets(ctx context.Context) ([]Changeset, error)
	HasOpenChangesetWithName(ctx context.Context, name string) (bool, error)
	HasChangesetWithName(ctx context.Context, name string) (bool, error)
	CreateChangeset(ctx context.Context, changeset *Changeset) error
	UpdateChangesetState(ctx context.Context, changesetID uint, state ChangesetState) error
}

type ListChangesets struct {
	changesetRepo ChangesetRepo
	tx            TransactionManager
}

func NewListChangesets(changesetRepo ChangesetRepo, tx TransactionManager) *ListChangesets {
	return &ListChangesets{
		changesetRepo: changesetRepo,
		tx:            tx,
	}
}

type ListChangesetsRequest struct{}

type ListChangesetsResponse struct {
	Changesets []Changeset `json:"changesets"`
}

func (l *ListChangesets) Exec(ctx context.Context, req ListChangesetsRequest) (*ListChangesetsResponse, error) {
	var changesets []Changeset
	err := l.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		changesets, err = l.changesetRepo.ListChangesets(ctx)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list changesets", err)
	}

	return &ListChangesetsResponse{
		Changesets: changesets,
	}, nil
}

type CreateChangeset struct {
	changesetRepo ChangesetRepo
	tx            TransactionManager
}

func NewCreateChangeset(changesetRepo ChangesetRepo, tx TransactionManager) *CreateChangeset {
	return &CreateChangeset{
		changesetRepo: changesetRepo,
		tx:            tx,
	}
}

type CreateChangesetRequest struct {
	Name string `json:"name"`
}

type CreateChangesetResponse struct {
	ID    uint           `json:"id"`
	Name  string         `json:"name"`
	State ChangesetState `json:"state"`
}

func (c *CreateChangeset) Exec(ctx context.Context, req CreateChangesetRequest) (*CreateChangesetResponse, error) {
	if req.Name == "" {
		return nil, UserErr("name is required")
	}

	var response *CreateChangesetResponse
	err := c.tx.Do(ctx, MainBranch, "create changeset", func(ctx context.Context) error {
		hasChangesets, err := c.changesetRepo.HasChangesetWithName(ctx, req.Name)
		if err != nil {
			return InternalErrE("failed to check for changesets", err)
		}
		if hasChangesets {
			return UserErr("cannot create changeset: changeset with this name already exists")
		}

		changeset := &Changeset{
			Name:  req.Name,
			State: ChangesetStateOpen,
		}

		err = c.changesetRepo.CreateChangeset(ctx, changeset)
		if err != nil {
			return InternalErrE("failed to create changeset", err)
		}

		response = &CreateChangesetResponse{
			ID:    changeset.ID,
			Name:  changeset.Name,
			State: changeset.State,
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create changeset: %w", err)
	}

	err = c.tx.CreateBranch(ctx, req.Name)
	if err != nil {
		return nil, InternalErrE("failed to create changeset branch", err)
	}

	return response, nil
}

type EnsureChangeset struct {
	changesetRepo   ChangesetRepo
	createChangeset *CreateChangeset
	tx              TransactionManager
}

func NewEnsureChangeset(changesetRepo ChangesetRepo, tx TransactionManager) *EnsureChangeset {
	return &EnsureChangeset{
		changesetRepo:   changesetRepo,
		createChangeset: NewCreateChangeset(changesetRepo, tx),
		tx:              tx,
	}
}

type EnsureChangesetRequest struct {
	Name string `json:"name"`
}

type EnsureChangesetResponse struct {
	ID    uint           `json:"id"`
	Name  string         `json:"name"`
	State ChangesetState `json:"state"`
}

func (e *EnsureChangeset) Exec(ctx context.Context, req EnsureChangesetRequest) (*EnsureChangesetResponse, error) {
	if req.Name == "" {
		return nil, UserErr("name is required")
	}

	var existingChangeset *Changeset
	err := e.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		existingChangeset, err = e.changesetRepo.GetChangesetByName(ctx, req.Name)
		if err != nil {
			return InternalErrE("failed to get changeset", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get changeset: %w", err)
	}
	if existingChangeset != nil {
		return &EnsureChangesetResponse{
			ID:    existingChangeset.ID,
			Name:  existingChangeset.Name,
			State: existingChangeset.State,
		}, nil
	}

	createResp, err := e.createChangeset.Exec(ctx, CreateChangesetRequest(req))
	if err != nil {
		return nil, fmt.Errorf("failed to create changeset: %w", err)
	}

	return &EnsureChangesetResponse{
		ID:    createResp.ID,
		Name:  createResp.Name,
		State: createResp.State,
	}, nil
}

type MergeChangeset struct {
	changesetRepo ChangesetRepo
	applyRepo     ApplyRepo
	applyWorker   *ApplyWorker
	tx            TransactionManager
}

func NewMergeChangeset(changesetRepo ChangesetRepo, applyRepo ApplyRepo, applyWorker *ApplyWorker, tx TransactionManager) *MergeChangeset {
	return &MergeChangeset{
		changesetRepo: changesetRepo,
		applyRepo:     applyRepo,
		applyWorker:   applyWorker,
		tx:            tx,
	}
}

type MergeChangesetRequest struct {
	ChangesetName string `json:"changeset_name"`
}

type MergeChangesetResponse struct {
	ID    uint           `json:"id"`
	Name  string         `json:"name"`
	State ChangesetState `json:"state"`
}

func (m *MergeChangeset) Exec(ctx context.Context, req MergeChangesetRequest) (*MergeChangesetResponse, error) {
	if req.ChangesetName == "" {
		return nil, UserErr("changeset_name is required")
	}

	var response *MergeChangesetResponse
	err := m.tx.Checkout(ctx, req.ChangesetName, func(ctx context.Context) error {
		changeset, err := m.changesetRepo.GetChangesetByName(ctx, req.ChangesetName)
		if err != nil {
			return UserErrE("changeset not found", err)
		}

		response = &MergeChangesetResponse{
			ID:    changeset.ID,
			Name:  changeset.Name,
			State: changeset.State,
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to merge changeset: %w", err)
	}

	err = m.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		err := m.tx.MergeBranch(ctx, req.ChangesetName)
		if err != nil {
			return InternalErrE("failed to merge changeset branch", err)
		}

		err = m.tx.DeleteBranch(ctx, req.ChangesetName)
		if err != nil {
			return InternalErrE("failed to delete changeset branch", err)
		}

		err = m.changesetRepo.UpdateChangesetState(ctx, response.ID, ChangesetStateMerged)
		if err != nil {
			return InternalErrE("failed to update changeset state to merged", err)
		}

		response.State = ChangesetStateMerged

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to merge changeset: %w", err)
	}

	if m.applyWorker != nil {
		applyIDs, err := m.applyRepo.GetQueuedAppliesByChangeset(ctx, response.ID)
		if err != nil {
			log.WithError(err).WithField("changeset_id", response.ID).Error("Failed to get queued applies for changeset")
			return response, nil
		}
		for _, applyID := range applyIDs {
			m.applyWorker.QueueApply(applyID)
		}
		log.WithField("changeset_id", response.ID).WithField("apply_count", len(applyIDs)).Info("Enqueued applies for merged changeset")
	}

	return response, nil
}
