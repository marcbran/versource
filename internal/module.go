package internal

import (
	"context"
	"net/url"
	"strings"
)

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
	GetModule(ctx context.Context, moduleID uint) (*Module, error)
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
}

func NewListModules(moduleRepo ModuleRepo) *ListModules {
	return &ListModules{
		moduleRepo: moduleRepo,
	}
}

type ListModulesRequest struct{}

type ListModulesResponse struct {
	Modules []Module `json:"modules"`
}

func (l *ListModules) Exec(ctx context.Context, req ListModulesRequest) (*ListModulesResponse, error) {
	modules, err := l.moduleRepo.ListModules(ctx)
	if err != nil {
		return nil, InternalErrE("failed to list modules", err)
	}

	return &ListModulesResponse{
		Modules: modules,
	}, nil
}

type ListModuleVersions struct {
	moduleVersionRepo ModuleVersionRepo
}

func NewListModuleVersions(moduleVersionRepo ModuleVersionRepo) *ListModuleVersions {
	return &ListModuleVersions{
		moduleVersionRepo: moduleVersionRepo,
	}
}

type ListModuleVersionsRequest struct{}

type ListModuleVersionsResponse struct {
	ModuleVersions []ModuleVersion `json:"module_versions"`
}

func (l *ListModuleVersions) Exec(ctx context.Context, req ListModuleVersionsRequest) (*ListModuleVersionsResponse, error) {
	moduleVersions, err := l.moduleVersionRepo.ListModuleVersions(ctx)
	if err != nil {
		return nil, InternalErrE("failed to list module versions", err)
	}

	return &ListModuleVersionsResponse{
		ModuleVersions: moduleVersions,
	}, nil
}

type ListModuleVersionsForModule struct {
	moduleVersionRepo ModuleVersionRepo
}

func NewListModuleVersionsForModule(moduleVersionRepo ModuleVersionRepo) *ListModuleVersionsForModule {
	return &ListModuleVersionsForModule{
		moduleVersionRepo: moduleVersionRepo,
	}
}

type ListModuleVersionsForModuleRequest struct {
	ModuleID uint `json:"module_id"`
}

type ListModuleVersionsForModuleResponse struct {
	ModuleVersions []ModuleVersion `json:"module_versions"`
}

func (l *ListModuleVersionsForModule) Exec(ctx context.Context, req ListModuleVersionsForModuleRequest) (*ListModuleVersionsForModuleResponse, error) {
	moduleVersions, err := l.moduleVersionRepo.ListModuleVersionsForModule(ctx, req.ModuleID)
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
	module, moduleVersion, err := createModuleWithVersion(req)
	if err != nil {
		return nil, err
	}

	var response *CreateModuleResponse
	err = c.tx.Do(ctx, "main", "create module", func(ctx context.Context) error {
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
	err := u.tx.Do(ctx, "main", "update module", func(ctx context.Context) error {
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
	err := d.tx.Do(ctx, "main", "delete module", func(ctx context.Context) error {
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

func createModuleWithVersion(request CreateModuleRequest) (*Module, *ModuleVersion, error) {
	if request.Source == "" {
		return nil, nil, UserErr("source is required")
	}

	if strings.HasPrefix(request.Source, "./") || strings.HasPrefix(request.Source, "../") {
		if request.Version != "" {
			return nil, nil, UserErr("local paths do not support version parameter")
		}
	} else if strings.HasPrefix(request.Source, "github.com/") || strings.HasPrefix(request.Source, "bitbucket.org/") || strings.HasPrefix(request.Source, "git::") || strings.HasPrefix(request.Source, "hg::") {
		if request.Version != "" {
			return nil, nil, UserErr("git/mercurial sources do not support version parameter, use ref parameter in source string")
		}
	} else if strings.HasPrefix(request.Source, "s3::") {
		if request.Version != "" {
			return nil, nil, UserErr("S3 sources do not support version parameter, use versionId parameter in source string")
		}
	} else if strings.HasPrefix(request.Source, "gcs::") {
		if request.Version != "" {
			return nil, nil, UserErr("GCS sources do not support version parameter, use generation parameter in source string")
		}
	} else if !strings.Contains(request.Source, "::") && !strings.Contains(request.Source, "://") {
		if request.Version == "" {
			return nil, nil, UserErr("terraform registry sources require version parameter")
		}
	}

	extractedVersion, err := extractVersionFromSource(request.Source)
	if err != nil {
		return nil, nil, err
	}

	cleanSource := request.Source
	if extractedVersion != "" {
		cleanSource = stripQueryParameters(request.Source)
	}

	module := &Module{
		Source: cleanSource,
	}

	version := request.Version
	if extractedVersion != "" {
		version = extractedVersion
	}

	moduleVersion := &ModuleVersion{
		Version: version,
	}

	return module, moduleVersion, nil
}

func stripQueryParameters(source string) string {
	if strings.HasPrefix(source, "s3::") {
		urlPart := strings.TrimPrefix(source, "s3::")
		u, err := url.Parse(urlPart)
		if err != nil {
			return source
		}
		u.RawQuery = ""
		return "s3::" + u.String()
	}

	if strings.HasPrefix(source, "gcs::") {
		urlPart := strings.TrimPrefix(source, "gcs::")
		u, err := url.Parse(urlPart)
		if err != nil {
			return source
		}
		u.RawQuery = ""
		return "gcs::" + u.String()
	}

	u, err := url.Parse(source)
	if err != nil {
		return source
	}
	u.RawQuery = ""
	return u.String()
}

func extractVersionFromSource(source string) (string, error) {
	if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") {
		return "", nil
	}

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return "", UserErr("HTTP/HTTPS sources are not supported")
	}

	if strings.HasPrefix(source, "s3::") {
		if !strings.Contains(source, "?versionId=") {
			return "", UserErr("S3 sources require versionId parameter in source string")
		}
		u, err := url.Parse(strings.TrimPrefix(source, "s3::"))
		if err != nil {
			return "", UserErr("invalid S3 source URL")
		}
		return u.Query().Get("versionId"), nil
	}

	if strings.HasPrefix(source, "gcs::") {
		if !strings.Contains(source, "?generation=") {
			return "", UserErr("GCS sources require generation parameter in source string")
		}
		u, err := url.Parse(strings.TrimPrefix(source, "gcs::"))
		if err != nil {
			return "", UserErr("invalid GCS source URL")
		}
		return u.Query().Get("generation"), nil
	}

	if strings.HasPrefix(source, "github.com/") || strings.HasPrefix(source, "bitbucket.org/") || strings.HasPrefix(source, "git::") || strings.HasPrefix(source, "hg::") {
		if !strings.Contains(source, "?ref=") {
			return "", UserErr("git/mercurial sources require ref parameter in source string")
		}
		u, err := url.Parse(source)
		if err != nil {
			return "", UserErr("invalid git/mercurial source URL")
		}
		return u.Query().Get("ref"), nil
	}

	if !strings.Contains(source, "::") && !strings.Contains(source, "://") {
		return "", nil
	}

	return "", UserErr("unknown module source type")
}
