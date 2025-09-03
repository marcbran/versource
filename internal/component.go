package internal

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
)

type Component struct {
	ID              uint          `gorm:"primarykey"`
	Name            string        `gorm:"not null;default:'';uniqueIndex"`
	ModuleVersion   ModuleVersion `gorm:"foreignKey:ModuleVersionID"`
	ModuleVersionID uint
	Variables       datatypes.JSON `gorm:"type:jsonb"`
}

type ComponentDiff struct {
	FromComponent Component
	ToComponent   Component
	DiffType      DiffType
}

type DiffType string

const (
	DiffTypeAdded    DiffType = "added"
	DiffTypeRemoved  DiffType = "removed"
	DiffTypeModified DiffType = "modified"
)

type ComponentRepo interface {
	GetComponent(ctx context.Context, componentID uint) (*Component, error)
	ListComponents(ctx context.Context) ([]Component, error)
	ListComponentsByModule(ctx context.Context, moduleID uint) ([]Component, error)
	ListComponentsByModuleVersion(ctx context.Context, moduleVersionID uint) ([]Component, error)
	CreateComponent(ctx context.Context, component *Component) error
	UpdateComponent(ctx context.Context, component *Component) error
}

type ComponentDiffRepo interface {
	ListComponentDiffs(ctx context.Context, fromCommit, toCommit string) ([]ComponentDiff, error)
}

type ListComponents struct {
	componentRepo ComponentRepo
	tx            TransactionManager
}

func NewListComponents(componentRepo ComponentRepo, tx TransactionManager) *ListComponents {
	return &ListComponents{
		componentRepo: componentRepo,
		tx:            tx,
	}
}

type ListComponentsRequest struct {
	ModuleID        *uint   `json:"module_id,omitempty"`
	ModuleVersionID *uint   `json:"module_version_id,omitempty"`
	Changeset       *string `json:"changeset,omitempty"`
}

type ListComponentsResponse struct {
	Components []Component `json:"components"`
}

func (l *ListComponents) Exec(ctx context.Context, req ListComponentsRequest) (*ListComponentsResponse, error) {
	var components []Component

	branch := MainBranch
	if req.Changeset != nil {
		branch = *req.Changeset
	}

	err := l.tx.Checkout(ctx, branch, func(ctx context.Context) error {
		var err error

		if req.ModuleVersionID != nil {
			components, err = l.componentRepo.ListComponentsByModuleVersion(ctx, *req.ModuleVersionID)
		} else if req.ModuleID != nil {
			components, err = l.componentRepo.ListComponentsByModule(ctx, *req.ModuleID)
		} else {
			components, err = l.componentRepo.ListComponents(ctx)
		}

		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list components", err)
	}

	return &ListComponentsResponse{
		Components: components,
	}, nil
}

type CreateComponent struct {
	componentRepo     ComponentRepo
	moduleRepo        ModuleRepo
	moduleVersionRepo ModuleVersionRepo
	ensureChangeset   *EnsureChangeset
	createPlan        *CreatePlan
	tx                TransactionManager
}

func NewCreateComponent(componentRepo ComponentRepo, moduleRepo ModuleRepo, moduleVersionRepo ModuleVersionRepo, ensureChangeset *EnsureChangeset, createPlan *CreatePlan, tx TransactionManager) *CreateComponent {
	return &CreateComponent{
		componentRepo:     componentRepo,
		moduleRepo:        moduleRepo,
		moduleVersionRepo: moduleVersionRepo,
		ensureChangeset:   ensureChangeset,
		createPlan:        createPlan,
		tx:                tx,
	}
}

type CreateComponentRequest struct {
	Changeset string         `json:"changeset"`
	ModuleID  uint           `json:"module_id"`
	Name      string         `json:"name"`
	Variables map[string]any `json:"variables"`
}

type CreateComponentResponse struct {
	ID        uint           `json:"id"`
	Name      string         `json:"name"`
	Source    string         `json:"source"`
	Version   string         `json:"version"`
	Variables map[string]any `json:"variables"`
	PlanID    uint           `json:"plan_id"`
}

func (c *CreateComponent) Exec(ctx context.Context, req CreateComponentRequest) (*CreateComponentResponse, error) {
	if req.Changeset == "" {
		return nil, UserErr("changeset is required")
	}

	ensureChangesetReq := EnsureChangesetRequest{
		Name: req.Changeset,
	}

	_, err := c.ensureChangeset.Exec(ctx, ensureChangesetReq)
	if err != nil {
		return nil, InternalErrE("failed to ensure changeset", err)
	}

	var response *CreateComponentResponse
	err = c.tx.Do(ctx, req.Changeset, "create component", func(ctx context.Context) error {
		latestVersion, err := c.moduleVersionRepo.GetLatestModuleVersion(ctx, req.ModuleID)
		if err != nil {
			return InternalErrE("failed to get latest module version", err)
		}
		if latestVersion == nil {
			return UserErr("module has no versions")
		}

		variablesJSON, err := json.Marshal(req.Variables)
		if err != nil {
			return UserErrE("invalid variables format", err)
		}

		component := &Component{
			Name:            req.Name,
			ModuleVersionID: latestVersion.ID,
			Variables:       datatypes.JSON(variablesJSON),
		}

		err = c.componentRepo.CreateComponent(ctx, component)
		if err != nil {
			return InternalErrE("failed to create component", err)
		}

		var variables map[string]any
		err = json.Unmarshal(component.Variables, &variables)
		if err != nil {
			return InternalErrE("failed to unmarshal variables", err)
		}

		response = &CreateComponentResponse{
			ID:        component.ID,
			Name:      component.Name,
			Source:    latestVersion.Module.Source,
			Version:   latestVersion.Version,
			Variables: variables,
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create component: %w", err)
	}

	planReq := CreatePlanRequest{
		ComponentID: response.ID,
		Changeset:   req.Changeset,
	}

	planResp, err := c.createPlan.Exec(ctx, planReq)
	if err != nil {
		return nil, InternalErrE("failed to create plan after component creation", err)
	}

	response.PlanID = planResp.ID

	return response, nil
}

type UpdateComponent struct {
	componentRepo     ComponentRepo
	moduleVersionRepo ModuleVersionRepo
	ensureChangeset   *EnsureChangeset
	tx                TransactionManager
}

func NewUpdateComponent(componentRepo ComponentRepo, moduleVersionRepo ModuleVersionRepo, ensureChangeset *EnsureChangeset, tx TransactionManager) *UpdateComponent {
	return &UpdateComponent{
		componentRepo:     componentRepo,
		moduleVersionRepo: moduleVersionRepo,
		ensureChangeset:   ensureChangeset,
		tx:                tx,
	}
}

type UpdateComponentRequest struct {
	ComponentID uint            `json:"component_id"`
	Changeset   string          `json:"changeset"`
	ModuleID    *uint           `json:"module_id,omitempty"`
	Variables   *map[string]any `json:"variables,omitempty"`
}

type UpdateComponentResponse struct {
	ID        uint           `json:"id"`
	Name      string         `json:"name"`
	Source    string         `json:"source"`
	Version   string         `json:"version"`
	Variables map[string]any `json:"variables"`
}

func (u *UpdateComponent) Exec(ctx context.Context, req UpdateComponentRequest) (*UpdateComponentResponse, error) {
	if req.Changeset == "" {
		return nil, UserErr("changeset is required")
	}

	var response *UpdateComponentResponse
	err := u.tx.Do(ctx, req.Changeset, "update component", func(ctx context.Context) error {
		component, err := u.componentRepo.GetComponent(ctx, req.ComponentID)
		if err != nil {
			return UserErrE("component not found", err)
		}

		if req.ModuleID != nil {
			latestVersion, err := u.moduleVersionRepo.GetLatestModuleVersion(ctx, *req.ModuleID)
			if err != nil {
				return InternalErrE("failed to get latest module version", err)
			}
			if latestVersion == nil {
				return UserErr("module has no versions")
			}
			component.ModuleVersionID = latestVersion.ID
		}
		if req.Variables != nil {
			variablesJSON, err := json.Marshal(*req.Variables)
			if err != nil {
				return UserErrE("invalid variables format", err)
			}
			component.Variables = datatypes.JSON(variablesJSON)
		}

		err = u.componentRepo.UpdateComponent(ctx, component)
		if err != nil {
			return InternalErrE("failed to update component", err)
		}

		var variables map[string]any
		err = json.Unmarshal(component.Variables, &variables)
		if err != nil {
			return InternalErrE("failed to unmarshal variables", err)
		}

		response = &UpdateComponentResponse{
			ID:        component.ID,
			Name:      component.Name,
			Source:    component.ModuleVersion.Module.Source,
			Version:   component.ModuleVersion.Version,
			Variables: variables,
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update component: %w", err)
	}

	return response, nil
}

type ListComponentDiffs struct {
	componentDiffRepo ComponentDiffRepo
	tx                TransactionManager
}

func NewListComponentDiffs(componentDiffRepo ComponentDiffRepo, tx TransactionManager) *ListComponentDiffs {
	return &ListComponentDiffs{
		componentDiffRepo: componentDiffRepo,
		tx:                tx,
	}
}

type ListComponentDiffsRequest struct {
	Changeset string `json:"changeset"`
}

type ListComponentDiffsResponse struct {
	Diffs []ComponentDiff `json:"diffs"`
}

func (l *ListComponentDiffs) Exec(ctx context.Context, req ListComponentDiffsRequest) (*ListComponentDiffsResponse, error) {
	if req.Changeset == "" {
		return nil, UserErr("changeset is required")
	}

	var diffs []ComponentDiff
	err := l.tx.Checkout(ctx, req.Changeset, func(ctx context.Context) error {
		mergeBase, err := l.tx.GetMergeBase(ctx, MainBranch, req.Changeset)
		if err != nil {
			return InternalErrE("failed to get merge base", err)
		}

		head, err := l.tx.GetHead(ctx)
		if err != nil {
			return InternalErrE("failed to get head", err)
		}

		if !IsValidCommitHash(mergeBase) {
			return InternalErrE("invalid merge base commit hash", fmt.Errorf("merge base '%s' is not a valid commit hash", mergeBase))
		}

		if !IsValidCommitHash(head) {
			return InternalErrE("invalid head commit hash", fmt.Errorf("head '%s' is not a valid commit hash", head))
		}

		diffs, err = l.componentDiffRepo.ListComponentDiffs(ctx, mergeBase, head)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list component diffs", err)
	}

	return &ListComponentDiffsResponse{
		Diffs: diffs,
	}, nil
}
