package internal

import (
	"context"
	"encoding/json"

	"gorm.io/datatypes"
)

type Component struct {
	ID              uint          `gorm:"primarykey"`
	ModuleVersion   ModuleVersion `gorm:"foreignKey:ModuleVersionID"`
	ModuleVersionID uint
	Variables       datatypes.JSON `gorm:"type:jsonb"`
}

type ComponentRepo interface {
	GetComponent(ctx context.Context, componentID uint) (*Component, error)
	CreateComponent(ctx context.Context, component *Component) error
	UpdateComponent(ctx context.Context, component *Component) error
	ListComponents(ctx context.Context) ([]Component, error)
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
	Variables map[string]any `json:"variables"`
}

type CreateComponentResponse struct {
	ID        uint           `json:"id"`
	Source    string         `json:"source"`
	Version   string         `json:"version"`
	Variables map[string]any `json:"variables"`
	PlanID    uint           `json:"plan_id"`
}

func (c *CreateComponent) Exec(ctx context.Context, req CreateComponentRequest) (*CreateComponentResponse, error) {
	if req.Changeset == "" {
		return nil, UserErr("changeset is required")
	}

	var response *CreateComponentResponse
	err := c.tx.Do(ctx, req.Changeset, "create component", func(ctx context.Context) error {
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
			Source:    latestVersion.Module.Source,
			Version:   latestVersion.Version,
			Variables: variables,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	ensureChangesetReq := EnsureChangesetRequest{
		Name: req.Changeset,
	}

	_, err = c.ensureChangeset.Exec(ctx, ensureChangesetReq)
	if err != nil {
		return nil, InternalErrE("failed to ensure changeset", err)
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
			Source:    component.ModuleVersion.Module.Source,
			Version:   component.ModuleVersion.Version,
			Variables: variables,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

type ListComponentsRequest struct{}

type ListComponentsResponse struct {
	Components []Component `json:"components"`
}

type ListComponents struct {
	componentRepo ComponentRepo
}

func NewListComponents(componentRepo ComponentRepo) *ListComponents {
	return &ListComponents{
		componentRepo: componentRepo,
	}
}

func (l *ListComponents) Exec(ctx context.Context, req ListComponentsRequest) (*ListComponentsResponse, error) {
	components, err := l.componentRepo.ListComponents(ctx)
	if err != nil {
		return nil, InternalErrE("failed to list components", err)
	}

	return &ListComponentsResponse{
		Components: components,
	}, nil
}
