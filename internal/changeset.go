package internal

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type ChangesetState string

const (
	ChangesetStateDraft      ChangesetState = "Draft"
	ChangesetStateValidating ChangesetState = "Validating"
	ChangesetStateValid      ChangesetState = "Valid"
	ChangesetStateInvalid    ChangesetState = "Invalid"
	ChangesetStateApplying   ChangesetState = "Applying"
	ChangesetStateApplied    ChangesetState = "Applied"
	ChangesetStateFailed     ChangesetState = "Failed"
)

type Changeset struct {
	ID    uint           `gorm:"primarykey"`
	Name  string         `gorm:"index"`
	State ChangesetState `gorm:"default:Draft"`
}

type ChangesetRepo interface {
	GetChangeset(ctx context.Context, changesetID uint) (*Changeset, error)
	GetChangesetByName(ctx context.Context, name string) (*Changeset, error)
	ListChangesets(ctx context.Context) ([]Changeset, error)
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

	branchExists, err := c.tx.BranchExists(ctx, req.Name)
	if err != nil {
		return nil, InternalErrE("failed to check if branch exists", err)
	}
	if branchExists {
		return nil, UserErr(fmt.Sprintf("changeset with name '%s' already exists", req.Name))
	}

	var response *CreateChangesetResponse
	err = c.tx.Do(ctx, req.Name, "create changeset", func(ctx context.Context) error {
		changeset := &Changeset{
			Name:  req.Name,
			State: ChangesetStateDraft,
		}

		err := c.changesetRepo.CreateChangeset(ctx, changeset)
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
		return nil, err
	}

	return response, nil
}

type EnsureChangeset struct {
	changesetRepo ChangesetRepo
	tx            TransactionManager
}

func NewEnsureChangeset(changesetRepo ChangesetRepo, tx TransactionManager) *EnsureChangeset {
	return &EnsureChangeset{
		changesetRepo: changesetRepo,
		tx:            tx,
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

	var response *EnsureChangesetResponse
	err := e.tx.Do(ctx, req.Name, "ensure changeset", func(ctx context.Context) error {
		existingChangeset, err := e.changesetRepo.GetChangesetByName(ctx, req.Name)
		if err == nil {
			response = &EnsureChangesetResponse{
				ID:    existingChangeset.ID,
				Name:  existingChangeset.Name,
				State: existingChangeset.State,
			}
			return nil
		}

		changeset := &Changeset{
			Name:  req.Name,
			State: ChangesetStateDraft,
		}

		err = e.changesetRepo.CreateChangeset(ctx, changeset)
		if err != nil {
			return InternalErrE("failed to create changeset", err)
		}

		response = &EnsureChangesetResponse{
			ID:    changeset.ID,
			Name:  changeset.Name,
			State: changeset.State,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
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
		return nil, err
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

		return nil
	})

	if err != nil {
		return nil, err
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
