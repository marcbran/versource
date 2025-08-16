package internal

import (
	"context"
	"encoding/json"

	"gorm.io/datatypes"
)

type Module struct {
	ID        uint `gorm:"primarykey"`
	Source    string
	Version   string
	Variables datatypes.JSON `gorm:"type:jsonb"`
}

type ModuleRepo interface {
	GetModule(ctx context.Context, moduleID uint) (*Module, error)
	CreateModule(ctx context.Context, module *Module) error
	UpdateModule(ctx context.Context, module *Module) error
}

type CreateModule struct {
	moduleRepo      ModuleRepo
	ensureChangeset *EnsureChangeset
	createPlan      *CreatePlan
	tx              TransactionManager
}

func NewCreateModule(moduleRepo ModuleRepo, ensureChangeset *EnsureChangeset, createPlan *CreatePlan, tx TransactionManager) *CreateModule {
	return &CreateModule{
		moduleRepo:      moduleRepo,
		ensureChangeset: ensureChangeset,
		createPlan:      createPlan,
		tx:              tx,
	}
}

type CreateModuleRequest struct {
	Changeset string         `json:"changeset"`
	Source    string         `json:"source"`
	Version   string         `json:"version"`
	Variables map[string]any `json:"variables"`
}

type CreateModuleResponse struct {
	ID        uint           `json:"id"`
	Source    string         `json:"source"`
	Version   string         `json:"version"`
	Variables map[string]any `json:"variables"`
	PlanID    uint           `json:"plan_id"`
}

func (c *CreateModule) Exec(ctx context.Context, req CreateModuleRequest) (*CreateModuleResponse, error) {
	if req.Source == "" {
		return nil, UserErr("source is required")
	}
	if req.Changeset == "" {
		return nil, UserErr("changeset is required")
	}

	var response *CreateModuleResponse
	err := c.tx.Do(ctx, req.Changeset, "create module", func(ctx context.Context) error {
		variablesJSON, err := json.Marshal(req.Variables)
		if err != nil {
			return UserErrE("invalid variables format", err)
		}

		module := &Module{
			Source:    req.Source,
			Version:   req.Version,
			Variables: datatypes.JSON(variablesJSON),
		}

		err = c.moduleRepo.CreateModule(ctx, module)
		if err != nil {
			return InternalErrE("failed to create module", err)
		}

		var variables map[string]any
		err = json.Unmarshal(module.Variables, &variables)
		if err != nil {
			return InternalErrE("failed to unmarshal variables", err)
		}

		response = &CreateModuleResponse{
			ID:        module.ID,
			Source:    module.Source,
			Version:   module.Version,
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
		ModuleID:  response.ID,
		Changeset: req.Changeset,
	}

	planResp, err := c.createPlan.Exec(ctx, planReq)
	if err != nil {
		return nil, InternalErrE("failed to create plan after module creation", err)
	}

	response.PlanID = planResp.ID

	return response, nil
}

type UpdateModule struct {
	moduleRepo      ModuleRepo
	ensureChangeset *EnsureChangeset
	tx              TransactionManager
}

func NewUpdateModule(moduleRepo ModuleRepo, ensureChangeset *EnsureChangeset, tx TransactionManager) *UpdateModule {
	return &UpdateModule{
		moduleRepo:      moduleRepo,
		ensureChangeset: ensureChangeset,
		tx:              tx,
	}
}

type UpdateModuleRequest struct {
	ModuleID  uint            `json:"module_id"`
	Changeset string          `json:"changeset"`
	Source    *string         `json:"source,omitempty"`
	Version   *string         `json:"version,omitempty"`
	Variables *map[string]any `json:"variables,omitempty"`
}

type UpdateModuleResponse struct {
	ID        uint           `json:"id"`
	Source    string         `json:"source"`
	Version   string         `json:"version"`
	Variables map[string]any `json:"variables"`
}

func (u *UpdateModule) Exec(ctx context.Context, req UpdateModuleRequest) (*UpdateModuleResponse, error) {
	if req.Changeset == "" {
		return nil, UserErr("changeset is required")
	}

	var response *UpdateModuleResponse
	err := u.tx.Do(ctx, req.Changeset, "update module", func(ctx context.Context) error {
		module, err := u.moduleRepo.GetModule(ctx, req.ModuleID)
		if err != nil {
			return UserErrE("module not found", err)
		}

		if req.Source != nil {
			module.Source = *req.Source
		}
		if req.Version != nil {
			module.Version = *req.Version
		}
		if req.Variables != nil {
			variablesJSON, err := json.Marshal(*req.Variables)
			if err != nil {
				return UserErrE("invalid variables format", err)
			}
			module.Variables = datatypes.JSON(variablesJSON)
		}

		err = u.moduleRepo.UpdateModule(ctx, module)
		if err != nil {
			return InternalErrE("failed to update module", err)
		}

		var variables map[string]any
		err = json.Unmarshal(module.Variables, &variables)
		if err != nil {
			return InternalErrE("failed to unmarshal variables", err)
		}

		response = &UpdateModuleResponse{
			ID:        module.ID,
			Source:    module.Source,
			Version:   module.Version,
			Variables: variables,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}
