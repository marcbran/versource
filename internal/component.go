package internal

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
)

type Component struct {
	ID              uint            `gorm:"primarykey" json:"id" yaml:"id"`
	Name            string          `gorm:"not null;default:'';uniqueIndex" json:"name" yaml:"name"`
	ModuleVersion   ModuleVersion   `gorm:"foreignKey:ModuleVersionID" json:"moduleVersion" yaml:"moduleVersion"`
	ModuleVersionID uint            `json:"moduleVersionId" yaml:"moduleVersionId"`
	Variables       datatypes.JSON  `gorm:"type:jsonb" json:"variables" yaml:"variables"`
	Status          ComponentStatus `gorm:"default:Ready" json:"status" yaml:"status"`
}

type ComponentStatus string

const (
	ComponentStatusReady   ComponentStatus = "Ready"
	ComponentStatusDeleted ComponentStatus = "Deleted"
)

type ComponentChange struct {
	FromComponent *Component `json:"fromComponent,omitempty" yaml:"fromComponent,omitempty"`
	ToComponent   *Component `json:"toComponent,omitempty" yaml:"toComponent,omitempty"`
	ChangeType    ChangeType `json:"changeType" yaml:"changeType"`
	Plan          *Plan      `json:"plan,omitempty" yaml:"plan,omitempty"`
	FromCommit    string     `json:"fromCommit,omitempty" yaml:"fromCommit,omitempty"`
	ToCommit      string     `json:"toCommit,omitempty" yaml:"toCommit,omitempty"`
}

type ChangeType string

const (
	ChangeTypeCreated  ChangeType = "Created"
	ChangeTypeDeleted  ChangeType = "Deleted"
	ChangeTypeModified ChangeType = "Modified"
)

type ComponentRepo interface {
	GetComponent(ctx context.Context, componentID uint) (*Component, error)
	GetComponentAtCommit(ctx context.Context, componentID uint, commit string) (*Component, error)
	GetLastCommitOfComponent(ctx context.Context, componentID uint) (string, error)
	HasComponent(ctx context.Context, componentID uint) (bool, error)
	ListComponents(ctx context.Context) ([]Component, error)
	ListComponentsByModule(ctx context.Context, moduleID uint) ([]Component, error)
	ListComponentsByModuleVersion(ctx context.Context, moduleVersionID uint) ([]Component, error)
	CreateComponent(ctx context.Context, component *Component) error
	UpdateComponent(ctx context.Context, component *Component) error
}

type ComponentChangeRepo interface {
	ListComponentChanges(ctx context.Context) ([]ComponentChange, error)
	GetComponentChange(ctx context.Context, componentID uint) (*ComponentChange, error)
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

type GetComponentRequest struct {
	ComponentID   uint    `json:"componentId" yaml:"componentId"`
	ChangesetName *string `json:"changesetName,omitempty" yaml:"changesetName,omitempty"`
}

type GetComponentResponse struct {
	Component Component `json:"component" yaml:"component"`
}

func (g *GetComponent) Exec(ctx context.Context, req GetComponentRequest) (*GetComponentResponse, error) {
	var component *Component
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
		return nil, InternalErrE("failed to get component", err)
	}

	return &GetComponentResponse{
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

type ListComponentsRequest struct {
	ModuleID        *uint   `json:"moduleId,omitempty" yaml:"moduleId,omitempty"`
	ModuleVersionID *uint   `json:"moduleVersionId,omitempty" yaml:"moduleVersionId,omitempty"`
	ChangesetName   *string `json:"changesetName,omitempty" yaml:"changesetName,omitempty"`
}

type ListComponentsResponse struct {
	Components []Component `json:"components" yaml:"components"`
}

func (l *ListComponents) Exec(ctx context.Context, req ListComponentsRequest) (*ListComponentsResponse, error) {
	var components []Component

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
		return nil, InternalErrE("failed to list components", err)
	}

	return &ListComponentsResponse{
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

type GetComponentChangeRequest struct {
	ComponentID   uint   `json:"componentId" yaml:"componentId"`
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type GetComponentChangeResponse struct {
	Change ComponentChange `json:"change" yaml:"change"`
}

func (g *GetComponentChange) Exec(ctx context.Context, req GetComponentChangeRequest) (*GetComponentChangeResponse, error) {
	if req.ChangesetName == "" {
		return nil, UserErr("changeset is required")
	}

	var change *ComponentChange
	err := g.tx.Checkout(ctx, req.ChangesetName, func(ctx context.Context) error {
		var err error
		change, err = g.componentChangeRepo.GetComponentChange(ctx, req.ComponentID)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to get component change", err)
	}

	return &GetComponentChangeResponse{
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

type ListComponentChangesRequest struct {
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type ListComponentChangesResponse struct {
	Changes []ComponentChange `json:"changes" yaml:"changes"`
}

func (l *ListComponentChanges) Exec(ctx context.Context, req ListComponentChangesRequest) (*ListComponentChangesResponse, error) {
	if req.ChangesetName == "" {
		return nil, UserErr("changeset is required")
	}

	var changes []ComponentChange
	err := l.tx.Checkout(ctx, req.ChangesetName, func(ctx context.Context) error {
		var err error
		changes, err = l.componentChangeRepo.ListComponentChanges(ctx)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list component changes", err)
	}

	return &ListComponentChangesResponse{
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

type CreateComponentRequest struct {
	ChangesetName string         `json:"changesetName" yaml:"changesetName"`
	ModuleID      uint           `json:"moduleId" yaml:"moduleId"`
	Name          string         `json:"name" yaml:"name"`
	Variables     map[string]any `json:"variables" yaml:"variables"`
}

type CreateComponentResponse struct {
	ID        uint           `json:"id" yaml:"id"`
	Name      string         `json:"name" yaml:"name"`
	Source    string         `json:"source" yaml:"source"`
	Version   string         `json:"version" yaml:"version"`
	Variables map[string]any `json:"variables" yaml:"variables"`
	PlanID    uint           `json:"planId" yaml:"planId"`
}

func (c *CreateComponent) Exec(ctx context.Context, req CreateComponentRequest) (*CreateComponentResponse, error) {
	if req.ChangesetName == "" {
		return nil, UserErr("changeset is required")
	}

	ensureChangesetReq := EnsureChangesetRequest{
		Name: req.ChangesetName,
	}

	_, err := c.ensureChangeset.Exec(ctx, ensureChangesetReq)
	if err != nil {
		return nil, InternalErrE("failed to ensure changeset", err)
	}

	var response *CreateComponentResponse
	err = c.tx.Do(ctx, req.ChangesetName, "create component", func(ctx context.Context) error {
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
			Status:          ComponentStatusReady,
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
		ComponentID:   response.ID,
		ChangesetName: req.ChangesetName,
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

type UpdateComponentRequest struct {
	ComponentID   uint            `json:"componentId" yaml:"componentId"`
	ChangesetName string          `json:"changesetName" yaml:"changesetName"`
	ModuleID      *uint           `json:"moduleId,omitempty" yaml:"moduleId,omitempty"`
	Variables     *map[string]any `json:"variables,omitempty" yaml:"variables,omitempty"`
}

type UpdateComponentResponse struct {
	ID        uint           `json:"id" yaml:"id"`
	Name      string         `json:"name" yaml:"name"`
	Source    string         `json:"source" yaml:"source"`
	Version   string         `json:"version" yaml:"version"`
	Variables map[string]any `json:"variables" yaml:"variables"`
	PlanID    uint           `json:"planId" yaml:"planId"`
}

func (u *UpdateComponent) Exec(ctx context.Context, req UpdateComponentRequest) (*UpdateComponentResponse, error) {
	if req.ChangesetName == "" {
		return nil, UserErr("changeset is required")
	}

	var hasChangeset bool
	err := u.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		hasChangeset, err = u.changesetRepo.HasChangesetWithName(ctx, req.ChangesetName)
		if err != nil {
			return InternalErrE("failed to check changeset existence", err)
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
			return InternalErrE("failed to check component existence", err)
		}
		if !exists {
			return UserErr("component not found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ensureChangesetReq := EnsureChangesetRequest{
		Name: req.ChangesetName,
	}

	_, err = u.ensureChangeset.Exec(ctx, ensureChangesetReq)
	if err != nil {
		return nil, InternalErrE("failed to ensure changeset", err)
	}

	var response *UpdateComponentResponse
	err = u.tx.Do(ctx, req.ChangesetName, "update component", func(ctx context.Context) error {
		component, err := u.componentRepo.GetComponent(ctx, req.ComponentID)
		if err != nil {
			return UserErrE("component not found", err)
		}

		if component.Status == ComponentStatusDeleted {
			return UserErr("component is deleted")
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

	planReq := CreatePlanRequest{
		ComponentID:   response.ID,
		ChangesetName: req.ChangesetName,
	}

	planResp, err := u.createPlan.Exec(ctx, planReq)
	if err != nil {
		return nil, InternalErrE("failed to create plan after component creation", err)
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

type DeleteComponentRequest struct {
	ComponentID   uint   `json:"componentId" yaml:"componentId"`
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type DeleteComponentResponse struct {
	ID        uint            `json:"id" yaml:"id"`
	Name      string          `json:"name" yaml:"name"`
	Source    string          `json:"source" yaml:"source"`
	Version   string          `json:"version" yaml:"version"`
	Variables map[string]any  `json:"variables" yaml:"variables"`
	Status    ComponentStatus `json:"status" yaml:"status"`
	PlanID    uint            `json:"planId" yaml:"planId"`
}

func (d *DeleteComponent) Exec(ctx context.Context, req DeleteComponentRequest) (*DeleteComponentResponse, error) {
	if req.ChangesetName == "" {
		return nil, UserErr("changeset is required")
	}

	var hasChangeset bool
	err := d.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		hasChangeset, err = d.changesetRepo.HasChangesetWithName(ctx, req.ChangesetName)
		if err != nil {
			return InternalErrE("failed to check changeset existence", err)
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
			return InternalErrE("failed to check component existence", err)
		}
		if !exists {
			return UserErr("component not found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ensureChangesetReq := EnsureChangesetRequest{
		Name: req.ChangesetName,
	}

	_, err = d.ensureChangeset.Exec(ctx, ensureChangesetReq)
	if err != nil {
		return nil, InternalErrE("failed to ensure changeset", err)
	}

	var response *DeleteComponentResponse
	err = d.tx.Do(ctx, req.ChangesetName, "delete component", func(ctx context.Context) error {
		component, err := d.componentRepo.GetComponent(ctx, req.ComponentID)
		if err != nil {
			return UserErrE("component not found", err)
		}

		if component.Status == ComponentStatusDeleted {
			return UserErr("component is already deleted")
		}

		componentChange, err := d.componentChangeRepo.GetComponentChange(ctx, req.ComponentID)
		if err != nil {
			return InternalErrE("failed to get component change", err)
		}

		if componentChange.FromComponent != nil {
			component.ModuleVersionID = componentChange.FromComponent.ModuleVersionID
			if componentChange.FromComponent.Variables == nil {
				component.Variables = datatypes.JSON("{}")
			} else {
				component.Variables = componentChange.FromComponent.Variables
			}
		}

		component.Status = ComponentStatusDeleted

		err = d.componentRepo.UpdateComponent(ctx, component)
		if err != nil {
			return InternalErrE("failed to update component", err)
		}

		var variables map[string]any
		err = json.Unmarshal(component.Variables, &variables)
		if err != nil {
			return InternalErrE("failed to unmarshal variables", err)
		}

		response = &DeleteComponentResponse{
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

	planReq := CreatePlanRequest{
		ComponentID:   response.ID,
		ChangesetName: req.ChangesetName,
	}

	planResp, err := d.createPlan.Exec(ctx, planReq)
	if err != nil {
		return nil, InternalErrE("failed to create plan after component deletion", err)
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

type RestoreComponentRequest struct {
	ComponentID   uint   `json:"componentId" yaml:"componentId"`
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type RestoreComponentResponse struct {
	ID        uint            `json:"id" yaml:"id"`
	Name      string          `json:"name" yaml:"name"`
	Source    string          `json:"source" yaml:"source"`
	Version   string          `json:"version" yaml:"version"`
	Variables map[string]any  `json:"variables" yaml:"variables"`
	Status    ComponentStatus `json:"status" yaml:"status"`
	PlanID    uint            `json:"planId" yaml:"planId"`
}

func (r *RestoreComponent) Exec(ctx context.Context, req RestoreComponentRequest) (*RestoreComponentResponse, error) {
	if req.ChangesetName == "" {
		return nil, UserErr("changeset is required")
	}

	var hasChangeset bool
	err := r.tx.Checkout(ctx, AdminBranch, func(ctx context.Context) error {
		var err error
		hasChangeset, err = r.changesetRepo.HasChangesetWithName(ctx, req.ChangesetName)
		if err != nil {
			return InternalErrE("failed to check changeset existence", err)
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
			return InternalErrE("failed to check component existence", err)
		}
		if !exists {
			return UserErr("component not found")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ensureChangesetReq := EnsureChangesetRequest{
		Name: req.ChangesetName,
	}

	_, err = r.ensureChangeset.Exec(ctx, ensureChangesetReq)
	if err != nil {
		return nil, InternalErrE("failed to ensure changeset", err)
	}

	var response *RestoreComponentResponse
	err = r.tx.Do(ctx, req.ChangesetName, "restore component", func(ctx context.Context) error {
		component, err := r.componentRepo.GetComponent(ctx, req.ComponentID)
		if err != nil {
			return UserErrE("component not found", err)
		}

		if component.Status != ComponentStatusDeleted {
			return UserErr("component is not deleted")
		}

		componentChange, err := r.componentChangeRepo.GetComponentChange(ctx, req.ComponentID)
		if err != nil {
			return InternalErrE("failed to get component change", err)
		}

		if componentChange.FromComponent != nil {
			component.ModuleVersionID = componentChange.FromComponent.ModuleVersionID
			if componentChange.FromComponent.Variables == nil {
				component.Variables = datatypes.JSON("{}")
			} else {
				component.Variables = componentChange.FromComponent.Variables
			}
		}

		component.Status = ComponentStatusReady

		err = r.componentRepo.UpdateComponent(ctx, component)
		if err != nil {
			return InternalErrE("failed to update component", err)
		}

		var variables map[string]any
		err = json.Unmarshal(component.Variables, &variables)
		if err != nil {
			return InternalErrE("failed to unmarshal variables", err)
		}

		response = &RestoreComponentResponse{
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

	planReq := CreatePlanRequest{
		ComponentID:   response.ID,
		ChangesetName: req.ChangesetName,
	}

	planResp, err := r.createPlan.Exec(ctx, planReq)
	if err != nil {
		return nil, InternalErrE("failed to create plan after component restoration", err)
	}

	response.PlanID = planResp.ID

	return response, nil
}
