package internal

import "context"

type Module struct {
	ID     uint `gorm:"primarykey"`
	Source string
}

type ModuleVersion struct {
	ID       uint   `gorm:"primarykey"`
	Module   Module `gorm:"foreignKey:ModuleID"`
	ModuleID uint
	Version  string
}

type ModuleRepo interface {
	CreateModule(ctx context.Context, module *Module) error
	GetModule(ctx context.Context, moduleID uint) (*Module, error)
	GetModuleBySource(ctx context.Context, source string) (*Module, error)
}

type ModuleVersionRepo interface {
	CreateModuleVersion(ctx context.Context, moduleVersion *ModuleVersion) error
	GetModuleVersion(ctx context.Context, moduleVersionID uint) (*ModuleVersion, error)
	GetModuleVersions(ctx context.Context, moduleID uint) ([]ModuleVersion, error)
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
	Source  string `json:"source"`
	Version string `json:"version"`
}

type CreateModuleResponse struct {
	ID        uint   `json:"id"`
	Source    string `json:"source"`
	VersionID uint   `json:"version_id"`
	Version   string `json:"version"`
}

func (c *CreateModule) Exec(ctx context.Context, req CreateModuleRequest) (*CreateModuleResponse, error) {
	if req.Source == "" {
		return nil, UserErr("source is required")
	}
	if req.Version == "" {
		return nil, UserErr("version is required")
	}

	var response *CreateModuleResponse
	err := c.tx.Do(ctx, "main", "create module", func(ctx context.Context) error {
		module := &Module{
			Source: req.Source,
		}

		err := c.moduleRepo.CreateModule(ctx, module)
		if err != nil {
			return InternalErrE("failed to create module", err)
		}

		moduleVersion := &ModuleVersion{
			ModuleID: module.ID,
			Version:  req.Version,
		}

		err = c.moduleVersionRepo.CreateModuleVersion(ctx, moduleVersion)
		if err != nil {
			return InternalErrE("failed to create module version", err)
		}

		response = &CreateModuleResponse{
			ID:        module.ID,
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
