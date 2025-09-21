package internal

import (
	"context"
	"fmt"
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
	UpdateChangesetReviewState(ctx context.Context, changesetID uint, reviewState ChangesetReviewState) error
	DeleteChangeset(ctx context.Context, changesetID uint) error
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
	err := l.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
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
	err := c.tx.Do(ctx, AdminBranch, "create changeset", func(ctx context.Context) error {
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

	err = c.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		err = c.tx.CreateBranch(ctx, req.Name)
		if err != nil {
			return InternalErrE("failed to create changeset branch", err)
		}
		return nil
	})

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
	err := e.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
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

type DeleteChangeset struct {
	changesetRepo ChangesetRepo
	planRepo      PlanRepo
	applyRepo     ApplyRepo
	planStore     PlanStore
	logStore      LogStore
	tx            TransactionManager
}

func NewDeleteChangeset(changesetRepo ChangesetRepo, planRepo PlanRepo, applyRepo ApplyRepo, planStore PlanStore, logStore LogStore, tx TransactionManager) *DeleteChangeset {
	return &DeleteChangeset{
		changesetRepo: changesetRepo,
		planRepo:      planRepo,
		applyRepo:     applyRepo,
		planStore:     planStore,
		logStore:      logStore,
		tx:            tx,
	}
}

type DeleteChangesetRequest struct {
	ChangesetName string `json:"changeset_name"`
}

type DeleteChangesetResponse struct {
	ID uint `json:"id"`
}

func (d *DeleteChangeset) Exec(ctx context.Context, req DeleteChangesetRequest) (*DeleteChangesetResponse, error) {
	if req.ChangesetName == "" {
		return nil, UserErr("changeset name is required")
	}

	var changeset *Changeset
	err := d.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		changeset, err = d.changesetRepo.GetChangesetByName(ctx, req.ChangesetName)
		if err != nil {
			return InternalErrE("failed to get changeset", err)
		}
		if changeset == nil {
			return UserErr("changeset not found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	plans, err := d.planRepo.ListPlansByChangeset(ctx, changeset.ID)
	if err != nil {
		return nil, InternalErrE("failed to list plans for changeset", err)
	}

	for _, plan := range plans {
		err = d.planStore.DeletePlan(ctx, plan.ID)
		if err != nil {
			return nil, InternalErrE("failed to delete plan from store", err)
		}

		err = d.logStore.DeleteLog(ctx, "plan", plan.ID)
		if err != nil {
			return nil, InternalErrE("failed to delete plan log", err)
		}
	}

	applies, err := d.applyRepo.ListAppliesByChangeset(ctx, changeset.ID)
	if err != nil {
		return nil, InternalErrE("failed to list applies for changeset", err)
	}

	for _, apply := range applies {
		err = d.logStore.DeleteLog(ctx, "apply", apply.ID)
		if err != nil {
			return nil, InternalErrE("failed to delete apply log", err)
		}
	}

	err = d.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		err = d.tx.DeleteBranch(ctx, req.ChangesetName)
		if err != nil {
			return InternalErrE("failed to delete changeset branch", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = d.tx.Do(ctx, AdminBranch, "delete changeset", func(ctx context.Context) error {
		err = d.changesetRepo.DeleteChangeset(ctx, changeset.ID)
		if err != nil {
			return InternalErrE("failed to delete changeset", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &DeleteChangesetResponse{
		ID: changeset.ID,
	}, nil
}
