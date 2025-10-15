package internal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/marcbran/versource/pkg/versource"
	"gorm.io/datatypes"
)

type ComponentRepo interface {
	GetComponent(ctx context.Context, componentID uint) (*versource.Component, error)
	GetComponentAtCommit(ctx context.Context, componentID uint, commit string) (*versource.Component, error)
	GetLastCommitOfComponent(ctx context.Context, componentID uint) (string, error)
	HasComponent(ctx context.Context, componentID uint) (bool, error)
	ListComponents(ctx context.Context) ([]versource.Component, error)
	ListComponentsByModule(ctx context.Context, moduleID uint) ([]versource.Component, error)
	ListComponentsByModuleVersion(ctx context.Context, moduleVersionID uint) ([]versource.Component, error)
	CreateComponent(ctx context.Context, component *versource.Component) error
	UpdateComponent(ctx context.Context, component *versource.Component) error
}

type ComponentChangeRepo interface {
	ListComponentChanges(ctx context.Context) ([]versource.ComponentChange, error)
	GetComponentChange(ctx context.Context, componentID uint) (*versource.ComponentChange, error)
	HasComponentConflicts(ctx context.Context, changesetName string) (bool, error)
}

type GetComponent struct {
	componentRepo ComponentRepo
	tx            TransactionManager
}

func NewGetComponent(componentRepo ComponentRepo, tx TransactionManager) *GetComponent {
	return &GetComponent{
		componentRepo: componentRepo,
		tx:            tx,
	}
}

func (g *GetComponent) Exec(ctx context.Context, req versource.GetComponentRequest) (*versource.GetComponentResponse, error) {
	var component *versource.Component
	var err error

	if req.ChangesetName != nil {
		err = g.tx.Checkout(ctx, *req.ChangesetName, func(ctx context.Context) error {
			component, err = g.componentRepo.GetComponent(ctx, req.ComponentID)
			return err
		})
	} else {
		component, err = g.componentRepo.GetComponent(ctx, req.ComponentID)
	}

	if err != nil {
		return nil, versource.InternalErrE("failed to get component", err)
	}

	return &versource.GetComponentResponse{
		Component: *component,
	}, nil
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

func (l *ListComponents) Exec(ctx context.Context, req versource.ListComponentsRequest) (*versource.ListComponentsResponse, error) {
	var components []versource.Component

	branch := MainBranch
	if req.ChangesetName != nil {
		branch = *req.ChangesetName
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
		return nil, versource.InternalErrE("failed to list components", err)
	}

	return &versource.ListComponentsResponse{
		Components: components,
	}, nil
}

type GetComponentChange struct {
	componentChangeRepo ComponentChangeRepo
	tx                  TransactionManager
}

func NewGetComponentChange(componentChangeRepo ComponentChangeRepo, tx TransactionManager) *GetComponentChange {
	return &GetComponentChange{
		componentChangeRepo: componentChangeRepo,
		tx:                  tx,
	}
}

func (g *GetComponentChange) Exec(ctx context.Context, req versource.GetComponentChangeRequest) (*versource.GetComponentChangeResponse, error) {
	if req.ChangesetName == "" {
		return nil, versource.UserErr("changeset is required")
	}

	var change *versource.ComponentChange
	err := g.tx.Checkout(ctx, req.ChangesetName, func(ctx context.Context) error {
		var err error
		change, err = g.componentChangeRepo.GetComponentChange(ctx, req.ComponentID)
		return err
	})
	if err != nil {
		return nil, versource.InternalErrE("failed to get component change", err)
	}

	return &versource.GetComponentChangeResponse{
		Change: *change,
	}, nil
}

type ListComponentChanges struct {
	componentChangeRepo ComponentChangeRepo
	tx                  TransactionManager
}

func NewListComponentChanges(componentChangeRepo ComponentChangeRepo, tx TransactionManager) *ListComponentChanges {
	return &ListComponentChanges{
		componentChangeRepo: componentChangeRepo,
		tx:                  tx,
	}
}

func (l *ListComponentChanges) Exec(ctx context.Context, req versource.ListComponentChangesRequest) (*versource.ListComponentChangesResponse, error) {
	if req.ChangesetName == "" {
		return nil, versource.UserErr("changeset is required")
	}

	var changes []versource.ComponentChange
	err := l.tx.Checkout(ctx, req.ChangesetName, func(ctx context.Context) error {
		var err error
		changes, err = l.componentChangeRepo.ListComponentChanges(ctx)
		return err
	})
	if err != nil {
		return nil, versource.InternalErrE("failed to list component changes", err)
	}

	return &versource.ListComponentChangesResponse{
		Changes: changes,
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

func (c *CreateComponent) Exec(ctx context.Context, req versource.CreateComponentRequest) (*versource.CreateComponentResponse, error) {
	if req.ChangesetName == "" {
		return nil, versource.UserErr("changeset is required")
	}

	ensureChangesetReq := versource.EnsureChangesetRequest{
		Name: req.ChangesetName,
	}

	_, err := c.ensureChangeset.Exec(ctx, ensureChangesetReq)
	if err != nil {
		return nil, versource.InternalErrE("failed to ensure changeset", err)
	}

	var response *versource.CreateComponentResponse
	err = c.tx.Do(ctx, req.ChangesetName, "create component", func(ctx context.Context) error {
		latestVersion, err := c.moduleVersionRepo.GetLatestModuleVersion(ctx, req.ModuleID)
		if err != nil {
			return versource.InternalErrE("failed to get latest module version", err)
		}
		if latestVersion == nil {
			return versource.UserErr("module has no versions")
		}

		variablesJSON, err := json.Marshal(req.Variables)
		if err != nil {
			return versource.UserErrE("invalid variables format", err)
		}

		component := &versource.Component{
			Name:            req.Name,
			ModuleVersionID: latestVersion.ID,
			Variables:       datatypes.JSON(variablesJSON),
			Status:          versource.ComponentStatusReady,
		}

		err = c.componentRepo.CreateComponent(ctx, component)
		if err != nil {
			return versource.InternalErrE("failed to create component", err)
		}

		var variables map[string]any
		err = json.Unmarshal(component.Variables, &variables)
		if err != nil {
			return versource.InternalErrE("failed to unmarshal variables", err)
		}

		response = &versource.CreateComponentResponse{
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

	planReq := versource.CreatePlanRequest{
		ComponentID:   response.ID,
		ChangesetName: req.ChangesetName,
	}

	planResp, err := c.createPlan.Exec(ctx, planReq)
	if err != nil {
		return nil, versource.InternalErrE("failed to create plan after component creation", err)
	}

	response.PlanID = planResp.ID

	return response, nil
}

type UpdateComponent struct {
	componentRepo     ComponentRepo
	moduleVersionRepo ModuleVersionRepo
	changesetRepo     ChangesetRepo
	ensureChangeset   *EnsureChangeset
	createPlan        *CreatePlan
	tx                TransactionManager
}

func NewUpdateComponent(componentRepo ComponentRepo, moduleVersionRepo ModuleVersionRepo, changesetRepo ChangesetRepo, ensureChangeset *EnsureChangeset, createPlan *CreatePlan, tx TransactionManager) *UpdateComponent {
	return &UpdateComponent{
		componentRepo:     componentRepo,
		moduleVersionRepo: moduleVersionRepo,
		changesetRepo:     changesetRepo,
		ensureChangeset:   ensureChangeset,
		createPlan:        createPlan,
		tx:                tx,
	}
}

func (u *UpdateComponent) Exec(ctx context.Context, req versource.UpdateComponentRequest) (*versource.UpdateComponentResponse, error) {
	if req.ChangesetName == "" {
		return nil, versource.UserErr("changeset is required")
	}

	var hasChangeset bool
	err := u.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		hasChangeset, err = u.changesetRepo.HasChangesetWithName(ctx, req.ChangesetName)
		if err != nil {
			return versource.InternalErrE("failed to check changeset existence", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var branch string
	if hasChangeset {
		branch = req.ChangesetName
	} else {
		branch = MainBranch
	}

	err = u.tx.Checkout(ctx, branch, func(ctx context.Context) error {
		exists, err := u.componentRepo.HasComponent(ctx, req.ComponentID)
		if err != nil {
			return versource.InternalErrE("failed to check component existence", err)
		}
		if !exists {
			return versource.UserErr("component not found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ensureChangesetReq := versource.EnsureChangesetRequest{
		Name: req.ChangesetName,
	}

	_, err = u.ensureChangeset.Exec(ctx, ensureChangesetReq)
	if err != nil {
		return nil, versource.InternalErrE("failed to ensure changeset", err)
	}

	var response *versource.UpdateComponentResponse
	err = u.tx.Do(ctx, req.ChangesetName, "update component", func(ctx context.Context) error {
		component, err := u.componentRepo.GetComponent(ctx, req.ComponentID)
		if err != nil {
			return versource.UserErrE("component not found", err)
		}

		if component.Status == versource.ComponentStatusDeleted {
			return versource.UserErr("component is deleted")
		}

		if req.ModuleID != nil {
			latestVersion, err := u.moduleVersionRepo.GetLatestModuleVersion(ctx, *req.ModuleID)
			if err != nil {
				return versource.InternalErrE("failed to get latest module version", err)
			}
			if latestVersion == nil {
				return versource.UserErr("module has no versions")
			}
			component.ModuleVersionID = latestVersion.ID
		}
		if req.Variables != nil {
			variablesJSON, err := json.Marshal(*req.Variables)
			if err != nil {
				return versource.UserErrE("invalid variables format", err)
			}
			component.Variables = datatypes.JSON(variablesJSON)
		}

		err = u.componentRepo.UpdateComponent(ctx, component)
		if err != nil {
			return versource.InternalErrE("failed to update component", err)
		}

		var variables map[string]any
		err = json.Unmarshal(component.Variables, &variables)
		if err != nil {
			return versource.InternalErrE("failed to unmarshal variables", err)
		}

		response = &versource.UpdateComponentResponse{
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

	planReq := versource.CreatePlanRequest{
		ComponentID:   response.ID,
		ChangesetName: req.ChangesetName,
	}

	planResp, err := u.createPlan.Exec(ctx, planReq)
	if err != nil {
		return nil, versource.InternalErrE("failed to create plan after component creation", err)
	}

	response.PlanID = planResp.ID

	return response, nil
}

type DeleteComponent struct {
	componentRepo       ComponentRepo
	componentChangeRepo ComponentChangeRepo
	changesetRepo       ChangesetRepo
	ensureChangeset     *EnsureChangeset
	createPlan          *CreatePlan
	tx                  TransactionManager
}

func NewDeleteComponent(componentRepo ComponentRepo, componentChangeRepo ComponentChangeRepo, changesetRepo ChangesetRepo, ensureChangeset *EnsureChangeset, createPlan *CreatePlan, tx TransactionManager) *DeleteComponent {
	return &DeleteComponent{
		componentRepo:       componentRepo,
		componentChangeRepo: componentChangeRepo,
		changesetRepo:       changesetRepo,
		ensureChangeset:     ensureChangeset,
		createPlan:          createPlan,
		tx:                  tx,
	}
}

func (d *DeleteComponent) Exec(ctx context.Context, req versource.DeleteComponentRequest) (*versource.DeleteComponentResponse, error) {
	if req.ChangesetName == "" {
		return nil, versource.UserErr("changeset is required")
	}

	var hasChangeset bool
	err := d.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		hasChangeset, err = d.changesetRepo.HasChangesetWithName(ctx, req.ChangesetName)
		if err != nil {
			return versource.InternalErrE("failed to check changeset existence", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var branch string
	if hasChangeset {
		branch = req.ChangesetName
	} else {
		branch = MainBranch
	}

	err = d.tx.Checkout(ctx, branch, func(ctx context.Context) error {
		exists, err := d.componentRepo.HasComponent(ctx, req.ComponentID)
		if err != nil {
			return versource.InternalErrE("failed to check component existence", err)
		}
		if !exists {
			return versource.UserErr("component not found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ensureChangesetReq := versource.EnsureChangesetRequest{
		Name: req.ChangesetName,
	}

	_, err = d.ensureChangeset.Exec(ctx, ensureChangesetReq)
	if err != nil {
		return nil, versource.InternalErrE("failed to ensure changeset", err)
	}

	var response *versource.DeleteComponentResponse
	err = d.tx.Do(ctx, req.ChangesetName, "delete component", func(ctx context.Context) error {
		component, err := d.componentRepo.GetComponent(ctx, req.ComponentID)
		if err != nil {
			return versource.UserErrE("component not found", err)
		}

		if component.Status == versource.ComponentStatusDeleted {
			return versource.UserErr("component is already deleted")
		}

		componentChange, err := d.componentChangeRepo.GetComponentChange(ctx, req.ComponentID)
		if err != nil {
			return versource.InternalErrE("failed to get component change", err)
		}

		if componentChange.FromComponent != nil {
			component.ModuleVersionID = componentChange.FromComponent.ModuleVersionID
			if componentChange.FromComponent.Variables == nil {
				component.Variables = datatypes.JSON("{}")
			} else {
				component.Variables = componentChange.FromComponent.Variables
			}
		}

		component.Status = versource.ComponentStatusDeleted

		err = d.componentRepo.UpdateComponent(ctx, component)
		if err != nil {
			return versource.InternalErrE("failed to update component", err)
		}

		var variables map[string]any
		err = json.Unmarshal(component.Variables, &variables)
		if err != nil {
			return versource.InternalErrE("failed to unmarshal variables", err)
		}

		response = &versource.DeleteComponentResponse{
			ID:        component.ID,
			Name:      component.Name,
			Source:    component.ModuleVersion.Module.Source,
			Version:   component.ModuleVersion.Version,
			Variables: variables,
			Status:    component.Status,
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete component: %w", err)
	}

	planReq := versource.CreatePlanRequest{
		ComponentID:   response.ID,
		ChangesetName: req.ChangesetName,
	}

	planResp, err := d.createPlan.Exec(ctx, planReq)
	if err != nil {
		return nil, versource.InternalErrE("failed to create plan after component deletion", err)
	}

	response.PlanID = planResp.ID

	return response, nil
}

type RestoreComponent struct {
	componentRepo       ComponentRepo
	componentChangeRepo ComponentChangeRepo
	changesetRepo       ChangesetRepo
	ensureChangeset     *EnsureChangeset
	createPlan          *CreatePlan
	tx                  TransactionManager
}

func NewRestoreComponent(componentRepo ComponentRepo, componentChangeRepo ComponentChangeRepo, changesetRepo ChangesetRepo, ensureChangeset *EnsureChangeset, createPlan *CreatePlan, tx TransactionManager) *RestoreComponent {
	return &RestoreComponent{
		componentRepo:       componentRepo,
		componentChangeRepo: componentChangeRepo,
		changesetRepo:       changesetRepo,
		ensureChangeset:     ensureChangeset,
		createPlan:          createPlan,
		tx:                  tx,
	}
}

func (r *RestoreComponent) Exec(ctx context.Context, req versource.RestoreComponentRequest) (*versource.RestoreComponentResponse, error) {
	if req.ChangesetName == "" {
		return nil, versource.UserErr("changeset is required")
	}

	var hasChangeset bool
	err := r.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		hasChangeset, err = r.changesetRepo.HasChangesetWithName(ctx, req.ChangesetName)
		if err != nil {
			return versource.InternalErrE("failed to check changeset existence", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var branch string
	if hasChangeset {
		branch = req.ChangesetName
	} else {
		branch = MainBranch
	}

	err = r.tx.Checkout(ctx, branch, func(ctx context.Context) error {
		exists, err := r.componentRepo.HasComponent(ctx, req.ComponentID)
		if err != nil {
			return versource.InternalErrE("failed to check component existence", err)
		}
		if !exists {
			return versource.UserErr("component not found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ensureChangesetReq := versource.EnsureChangesetRequest{
		Name: req.ChangesetName,
	}

	_, err = r.ensureChangeset.Exec(ctx, ensureChangesetReq)
	if err != nil {
		return nil, versource.InternalErrE("failed to ensure changeset", err)
	}

	var response *versource.RestoreComponentResponse
	err = r.tx.Do(ctx, req.ChangesetName, "restore component", func(ctx context.Context) error {
		component, err := r.componentRepo.GetComponent(ctx, req.ComponentID)
		if err != nil {
			return versource.UserErrE("component not found", err)
		}

		if component.Status != versource.ComponentStatusDeleted {
			return versource.UserErr("component is not deleted")
		}

		componentChange, err := r.componentChangeRepo.GetComponentChange(ctx, req.ComponentID)
		if err != nil {
			return versource.InternalErrE("failed to get component change", err)
		}

		if componentChange.FromComponent != nil {
			component.ModuleVersionID = componentChange.FromComponent.ModuleVersionID
			if componentChange.FromComponent.Variables == nil {
				component.Variables = datatypes.JSON("{}")
			} else {
				component.Variables = componentChange.FromComponent.Variables
			}
		}

		component.Status = versource.ComponentStatusReady

		err = r.componentRepo.UpdateComponent(ctx, component)
		if err != nil {
			return versource.InternalErrE("failed to update component", err)
		}

		var variables map[string]any
		err = json.Unmarshal(component.Variables, &variables)
		if err != nil {
			return versource.InternalErrE("failed to unmarshal variables", err)
		}

		response = &versource.RestoreComponentResponse{
			ID:        component.ID,
			Name:      component.Name,
			Source:    component.ModuleVersion.Module.Source,
			Version:   component.ModuleVersion.Version,
			Variables: variables,
			Status:    component.Status,
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to restore component: %w", err)
	}

	planReq := versource.CreatePlanRequest{
		ComponentID:   response.ID,
		ChangesetName: req.ChangesetName,
	}

	planResp, err := r.createPlan.Exec(ctx, planReq)
	if err != nil {
		return nil, versource.InternalErrE("failed to create plan after component restoration", err)
	}

	response.PlanID = planResp.ID

	return response, nil
}
