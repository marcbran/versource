package internal

import (
	"context"

	"github.com/marcbran/versource/pkg/versource"
)

type ModuleRepo interface {
	GetModule(ctx context.Context, moduleID uint) (*versource.Module, error)
	GetModuleByName(ctx context.Context, name string) (*versource.Module, error)
	GetModuleBySource(ctx context.Context, source string) (*versource.Module, error)
	ListModules(ctx context.Context) ([]versource.Module, error)
	CreateModule(ctx context.Context, module *versource.Module) error
	DeleteModule(ctx context.Context, moduleID uint) error
}

type ModuleVersionRepo interface {
	GetModuleVersion(ctx context.Context, moduleVersionID uint) (*versource.ModuleVersion, error)
	GetLatestModuleVersion(ctx context.Context, moduleID uint) (*versource.ModuleVersion, error)
	ListModuleVersions(ctx context.Context) ([]versource.ModuleVersion, error)
	ListModuleVersionsForModule(ctx context.Context, moduleID uint) ([]versource.ModuleVersion, error)
	CreateModuleVersion(ctx context.Context, moduleVersion *versource.ModuleVersion) error
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

func (g *GetModule) Exec(ctx context.Context, req versource.GetModuleRequest) (*versource.GetModuleResponse, error) {
	var module *versource.Module
	var latestVersion *versource.ModuleVersion
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
		return nil, versource.InternalErrE("failed to get module", err)
	}

	if module == nil {
		return nil, versource.UserErr("module not found")
	}

	return &versource.GetModuleResponse{
		Module:        *module,
		LatestVersion: latestVersion,
	}, nil
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

func (l *ListModules) Exec(ctx context.Context, req versource.ListModulesRequest) (*versource.ListModulesResponse, error) {
	var modules []versource.Module
	err := l.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		modules, err = l.moduleRepo.ListModules(ctx)
		return err
	})
	if err != nil {
		return nil, versource.InternalErrE("failed to list modules", err)
	}

	return &versource.ListModulesResponse{
		Modules: modules,
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

func (c *CreateModule) Exec(ctx context.Context, req versource.CreateModuleRequest) (*versource.CreateModuleResponse, error) {
	if req.Name == "" {
		return nil, versource.UserErr("name is required")
	}

	if req.Source == "" {
		return nil, versource.UserErr("source is required")
	}

	if req.ExecutorType == "" {
		return nil, versource.UserErr("executor type is required")
	}

	module := &versource.Module{
		Name:         req.Name,
		Source:       req.Source,
		ExecutorType: req.ExecutorType,
	}

	moduleVersion := &versource.ModuleVersion{
		Version: req.Version,
	}

	var response *versource.CreateModuleResponse
	err := c.tx.Do(ctx, MainBranch, "create module", func(ctx context.Context) error {
		err := c.moduleRepo.CreateModule(ctx, module)
		if err != nil {
			return versource.InternalErrE("failed to create module", err)
		}

		moduleVersion.ModuleID = module.ID

		err = c.moduleVersionRepo.CreateModuleVersion(ctx, moduleVersion)
		if err != nil {
			return versource.InternalErrE("failed to create module version", err)
		}

		response = &versource.CreateModuleResponse{
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

func (u *UpdateModule) Exec(ctx context.Context, req versource.UpdateModuleRequest) (*versource.UpdateModuleResponse, error) {
	if req.Version == "" {
		return nil, versource.UserErr("version is required")
	}

	var response *versource.UpdateModuleResponse
	err := u.tx.Do(ctx, MainBranch, "update module", func(ctx context.Context) error {
		module, err := u.moduleRepo.GetModule(ctx, req.ModuleID)
		if err != nil {
			return versource.InternalErrE("failed to get module", err)
		}
		if module == nil {
			return versource.UserErr("module not found")
		}

		currentVersion, err := u.moduleVersionRepo.GetLatestModuleVersion(ctx, req.ModuleID)
		if err != nil {
			return versource.InternalErrE("failed to get current module version", err)
		}
		if currentVersion == nil {
			return versource.UserErr("module has no versions")
		}

		if currentVersion.Version == "" {
			return versource.UserErr("cannot update module with empty version")
		}

		moduleVersion := &versource.ModuleVersion{
			ModuleID: req.ModuleID,
			Version:  req.Version,
		}

		err = u.moduleVersionRepo.CreateModuleVersion(ctx, moduleVersion)
		if err != nil {
			return versource.InternalErrE("failed to create module version", err)
		}

		response = &versource.UpdateModuleResponse{
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

func (d *DeleteModule) Exec(ctx context.Context, req versource.DeleteModuleRequest) (*versource.DeleteModuleResponse, error) {
	if req.ModuleID == 0 {
		return nil, versource.UserErr("module_id is required")
	}

	var response *versource.DeleteModuleResponse
	err := d.tx.Do(ctx, MainBranch, "delete module", func(ctx context.Context) error {
		module, err := d.moduleRepo.GetModule(ctx, req.ModuleID)
		if err != nil {
			return versource.InternalErrE("failed to get module", err)
		}
		if module == nil {
			return versource.UserErr("module not found")
		}

		components, err := d.componentRepo.ListComponentsByModule(ctx, req.ModuleID)
		if err != nil {
			return versource.InternalErrE("failed to check module references", err)
		}

		if len(components) > 0 {
			return versource.UserErr("cannot delete module that is referenced by components")
		}

		err = d.moduleRepo.DeleteModule(ctx, req.ModuleID)
		if err != nil {
			return versource.InternalErrE("failed to delete module", err)
		}

		response = &versource.DeleteModuleResponse{
			ModuleID: req.ModuleID,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
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

func (g *GetModuleVersion) Exec(ctx context.Context, req versource.GetModuleVersionRequest) (*versource.GetModuleVersionResponse, error) {
	var moduleVersion *versource.ModuleVersion
	err := g.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		moduleVersion, err = g.moduleVersionRepo.GetModuleVersion(ctx, req.ModuleVersionID)
		return err
	})
	if err != nil {
		return nil, versource.InternalErrE("failed to get module version", err)
	}

	if moduleVersion == nil {
		return nil, versource.UserErr("module version not found")
	}

	return &versource.GetModuleVersionResponse{
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

func (l *ListModuleVersions) Exec(ctx context.Context, req versource.ListModuleVersionsRequest) (*versource.ListModuleVersionsResponse, error) {
	var moduleVersions []versource.ModuleVersion
	err := l.tx.Checkout(ctx, MainBranch, func(ctx context.Context) error {
		var err error
		if req.ModuleID != nil {
			moduleVersions, err = l.moduleVersionRepo.ListModuleVersionsForModule(ctx, *req.ModuleID)
		} else {
			moduleVersions, err = l.moduleVersionRepo.ListModuleVersions(ctx)
		}
		return err
	})
	if err != nil {
		return nil, versource.InternalErrE("failed to list module versions", err)
	}

	return &versource.ListModuleVersionsResponse{
		ModuleVersions: moduleVersions,
	}, nil
}
