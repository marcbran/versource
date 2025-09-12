package internal

import (
	"context"

	"gorm.io/datatypes"
)

type Module struct {
	ID           uint   `gorm:"primarykey"`
	Name         string `gorm:"uniqueIndex;not null"`
	Source       string
	ExecutorType string `gorm:"not null;default:'terraform-module'"`
}

type ModuleVersion struct {
	ID        uint   `gorm:"primarykey"`
	Module    Module `gorm:"foreignKey:ModuleID"`
	ModuleID  uint
	Version   string
	Variables datatypes.JSON `gorm:"type:jsonb"`
	Outputs   datatypes.JSON `gorm:"type:jsonb"`
}

type ModuleRepo interface {
	GetModule(ctx context.Context, moduleID uint) (*Module, error)
	GetModuleByName(ctx context.Context, name string) (*Module, error)
	GetModuleBySource(ctx context.Context, source string) (*Module, error)
	ListModules(ctx context.Context) ([]Module, error)
	CreateModule(ctx context.Context, module *Module) error
	DeleteModule(ctx context.Context, moduleID uint) error
}

type ModuleVersionRepo interface {
	GetModuleVersion(ctx context.Context, moduleVersionID uint) (*ModuleVersion, error)
	GetLatestModuleVersion(ctx context.Context, moduleID uint) (*ModuleVersion, error)
	ListModuleVersions(ctx context.Context) ([]ModuleVersion, error)
	ListModuleVersionsForModule(ctx context.Context, moduleID uint) ([]ModuleVersion, error)
	CreateModuleVersion(ctx context.Context, moduleVersion *ModuleVersion) error
}

type ListModules struct {
	moduleRepo ModuleRepo
	tx         TransactionManager
}

func NewListModules(moduleRepo ModuleRepo, tx TransactionManager) *ListModules {
	return &ListModules{
		moduleRepo: moduleRepo,
		tx:         tx,
	}
}

type ListModulesRequest struct{}

type ListModulesResponse struct {
	Modules []Module `json:"modules"`
}

func (l *ListModules) Exec(ctx context.Context, req ListModulesRequest) (*ListModulesResponse, error) {
	var modules []Module
	err := l.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		modules, err = l.moduleRepo.ListModules(ctx)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list modules", err)
	}

	return &ListModulesResponse{
		Modules: modules,
	}, nil
}

type GetModule struct {
	moduleRepo        ModuleRepo
	moduleVersionRepo ModuleVersionRepo
	tx                TransactionManager
}

func NewGetModule(moduleRepo ModuleRepo, moduleVersionRepo ModuleVersionRepo, tx TransactionManager) *GetModule {
	return &GetModule{
		moduleRepo:        moduleRepo,
		moduleVersionRepo: moduleVersionRepo,
		tx:                tx,
	}
}

type GetModuleRequest struct {
	ModuleID uint `json:"module_id"`
}

type GetModuleResponse struct {
	Module        Module         `json:"module"`
	LatestVersion *ModuleVersion `json:"latest_version,omitempty"`
}

func (g *GetModule) Exec(ctx context.Context, req GetModuleRequest) (*GetModuleResponse, error) {
	var module *Module
	var latestVersion *ModuleVersion
	err := g.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		module, err = g.moduleRepo.GetModule(ctx, req.ModuleID)
		if err != nil {
			return err
		}

		latestVersion, err = g.moduleVersionRepo.GetLatestModuleVersion(ctx, req.ModuleID)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to get module", err)
	}

	if module == nil {
		return nil, UserErr("module not found")
	}

	return &GetModuleResponse{
		Module:        *module,
		LatestVersion: latestVersion,
	}, nil
}

type GetModuleVersion struct {
	moduleVersionRepo ModuleVersionRepo
	tx                TransactionManager
}

func NewGetModuleVersion(moduleVersionRepo ModuleVersionRepo, tx TransactionManager) *GetModuleVersion {
	return &GetModuleVersion{
		moduleVersionRepo: moduleVersionRepo,
		tx:                tx,
	}
}

type GetModuleVersionRequest struct {
	ModuleVersionID uint `json:"module_version_id"`
}

type GetModuleVersionResponse struct {
	ModuleVersion ModuleVersion `json:"module_version"`
}

func (g *GetModuleVersion) Exec(ctx context.Context, req GetModuleVersionRequest) (*GetModuleVersionResponse, error) {
	var moduleVersion *ModuleVersion
	err := g.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		moduleVersion, err = g.moduleVersionRepo.GetModuleVersion(ctx, req.ModuleVersionID)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to get module version", err)
	}

	if moduleVersion == nil {
		return nil, UserErr("module version not found")
	}

	return &GetModuleVersionResponse{
		ModuleVersion: *moduleVersion,
	}, nil
}

type ListModuleVersions struct {
	moduleVersionRepo ModuleVersionRepo
	tx                TransactionManager
}

func NewListModuleVersions(moduleVersionRepo ModuleVersionRepo, tx TransactionManager) *ListModuleVersions {
	return &ListModuleVersions{
		moduleVersionRepo: moduleVersionRepo,
		tx:                tx,
	}
}

type ListModuleVersionsRequest struct{}

type ListModuleVersionsResponse struct {
	ModuleVersions []ModuleVersion `json:"module_versions"`
}

func (l *ListModuleVersions) Exec(ctx context.Context, req ListModuleVersionsRequest) (*ListModuleVersionsResponse, error) {
	var moduleVersions []ModuleVersion
	err := l.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		moduleVersions, err = l.moduleVersionRepo.ListModuleVersions(ctx)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list module versions", err)
	}

	return &ListModuleVersionsResponse{
		ModuleVersions: moduleVersions,
	}, nil
}

type ListModuleVersionsForModule struct {
	moduleVersionRepo ModuleVersionRepo
	tx                TransactionManager
}

func NewListModuleVersionsForModule(moduleVersionRepo ModuleVersionRepo, tx TransactionManager) *ListModuleVersionsForModule {
	return &ListModuleVersionsForModule{
		moduleVersionRepo: moduleVersionRepo,
		tx:                tx,
	}
}

type ListModuleVersionsForModuleRequest struct {
	ModuleID uint `json:"module_id"`
}

type ListModuleVersionsForModuleResponse struct {
	ModuleVersions []ModuleVersion `json:"module_versions"`
}

func (l *ListModuleVersionsForModule) Exec(ctx context.Context, req ListModuleVersionsForModuleRequest) (*ListModuleVersionsForModuleResponse, error) {
	var moduleVersions []ModuleVersion
	err := l.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		moduleVersions, err = l.moduleVersionRepo.ListModuleVersionsForModule(ctx, req.ModuleID)
		return err
	})
	if err != nil {
		return nil, InternalErrE("failed to list module versions for module", err)
	}

	return &ListModuleVersionsForModuleResponse{
		ModuleVersions: moduleVersions,
	}, nil
}

type CreateModule struct {
	moduleRepo        ModuleRepo
	moduleVersionRepo ModuleVersionRepo
	tx                TransactionManager
}

func NewCreateModule(moduleRepo ModuleRepo, moduleVersionRepo ModuleVersionRepo, tx TransactionManager) *CreateModule {
	return &CreateModule{
		moduleRepo:        moduleRepo,
		moduleVersionRepo: moduleVersionRepo,
		tx:                tx,
	}
}

type CreateModuleRequest struct {
	Name         string `json:"name" yaml:"name"`
	Source       string `json:"source" yaml:"source"`
	Version      string `json:"version" yaml:"version"`
	ExecutorType string `json:"executor_type,omitempty" yaml:"executorType,omitempty"`
}

type CreateModuleResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	VersionID uint   `json:"version_id"`
	Version   string `json:"version"`
}

func (c *CreateModule) Exec(ctx context.Context, req CreateModuleRequest) (*CreateModuleResponse, error) {
	if req.Name == "" {
		return nil, UserErr("name is required")
	}

	if req.Source == "" {
		return nil, UserErr("source is required")
	}

	if req.ExecutorType == "" {
		return nil, UserErr("executor type is required")
	}

	module := &Module{
		Name:         req.Name,
		Source:       req.Source,
		ExecutorType: req.ExecutorType,
	}

	moduleVersion := &ModuleVersion{
		Version: req.Version,
	}

	var response *CreateModuleResponse
	err := c.tx.Do(ctx, MainBranch, "create module", func(ctx context.Context) error {
		err := c.moduleRepo.CreateModule(ctx, module)
		if err != nil {
			return InternalErrE("failed to create module", err)
		}

		moduleVersion.ModuleID = module.ID

		err = c.moduleVersionRepo.CreateModuleVersion(ctx, moduleVersion)
		if err != nil {
			return InternalErrE("failed to create module version", err)
		}

		response = &CreateModuleResponse{
			ID:        module.ID,
			Name:      module.Name,
			Source:    module.Source,
			VersionID: moduleVersion.ID,
			Version:   moduleVersion.Version,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

type UpdateModule struct {
	tx                TransactionManager
	moduleRepo        ModuleRepo
	moduleVersionRepo ModuleVersionRepo
}

func NewUpdateModule(moduleRepo ModuleRepo, moduleVersionRepo ModuleVersionRepo, tx TransactionManager) *UpdateModule {
	return &UpdateModule{
		moduleRepo:        moduleRepo,
		moduleVersionRepo: moduleVersionRepo,
		tx:                tx,
	}
}

type UpdateModuleRequest struct {
	ModuleID uint   `json:"module_id"`
	Version  string `json:"version"`
}

type UpdateModuleResponse struct {
	ModuleID  uint   `json:"module_id"`
	VersionID uint   `json:"version_id"`
	Version   string `json:"version"`
}

func (u *UpdateModule) Exec(ctx context.Context, req UpdateModuleRequest) (*UpdateModuleResponse, error) {
	if req.Version == "" {
		return nil, UserErr("version is required")
	}

	var response *UpdateModuleResponse
	err := u.tx.Do(ctx, MainBranch, "update module", func(ctx context.Context) error {
		module, err := u.moduleRepo.GetModule(ctx, req.ModuleID)
		if err != nil {
			return InternalErrE("failed to get module", err)
		}
		if module == nil {
			return UserErr("module not found")
		}

		currentVersion, err := u.moduleVersionRepo.GetLatestModuleVersion(ctx, req.ModuleID)
		if err != nil {
			return InternalErrE("failed to get current module version", err)
		}
		if currentVersion == nil {
			return UserErr("module has no versions")
		}

		if currentVersion.Version == "" {
			return UserErr("cannot update module with empty version")
		}

		moduleVersion := &ModuleVersion{
			ModuleID: req.ModuleID,
			Version:  req.Version,
		}

		err = u.moduleVersionRepo.CreateModuleVersion(ctx, moduleVersion)
		if err != nil {
			return InternalErrE("failed to create module version", err)
		}

		response = &UpdateModuleResponse{
			ModuleID:  req.ModuleID,
			VersionID: moduleVersion.ID,
			Version:   moduleVersion.Version,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

type DeleteModule struct {
	moduleRepo    ModuleRepo
	componentRepo ComponentRepo
	tx            TransactionManager
}

func NewDeleteModule(moduleRepo ModuleRepo, componentRepo ComponentRepo, tx TransactionManager) *DeleteModule {
	return &DeleteModule{
		moduleRepo:    moduleRepo,
		componentRepo: componentRepo,
		tx:            tx,
	}
}

type DeleteModuleRequest struct {
	ModuleID uint `json:"module_id"`
}

type DeleteModuleResponse struct {
	ModuleID uint `json:"module_id"`
}

func (d *DeleteModule) Exec(ctx context.Context, req DeleteModuleRequest) (*DeleteModuleResponse, error) {
	if req.ModuleID == 0 {
		return nil, UserErr("module_id is required")
	}

	var response *DeleteModuleResponse
	err := d.tx.Do(ctx, MainBranch, "delete module", func(ctx context.Context) error {
		module, err := d.moduleRepo.GetModule(ctx, req.ModuleID)
		if err != nil {
			return InternalErrE("failed to get module", err)
		}
		if module == nil {
			return UserErr("module not found")
		}

		components, err := d.componentRepo.ListComponentsByModule(ctx, req.ModuleID)
		if err != nil {
			return InternalErrE("failed to check module references", err)
		}

		if len(components) > 0 {
			return UserErr("cannot delete module that is referenced by components")
		}

		err = d.moduleRepo.DeleteModule(ctx, req.ModuleID)
		if err != nil {
			return InternalErrE("failed to delete module", err)
		}

		response = &DeleteModuleResponse{
			ModuleID: req.ModuleID,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}
