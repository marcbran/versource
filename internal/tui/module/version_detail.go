package module

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type VersionDetailData struct {
	facade          versource.Facade
	moduleVersionID string
}

type VersionDetailViewModel struct {
	ID      uint   `yaml:"id"`
	Version string `yaml:"version"`
	Module  struct {
		ID           uint   `yaml:"id"`
		Name         string `yaml:"name"`
		Source       string `yaml:"source"`
		ExecutorType string `yaml:"executorType"`
	} `yaml:"module"`
}

func NewVersionDetail(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewViewDataViewport(NewVersionDetailData(facade, params["moduleVersionID"]))
	}
}

func NewVersionDetailData(facade versource.Facade, moduleVersionID string) *VersionDetailData {
	return &VersionDetailData{facade: facade, moduleVersionID: moduleVersionID}
}

func (p *VersionDetailData) LoadData() (*versource.GetModuleVersionResponse, error) {
	ctx := context.Background()

	moduleVersionIDUint, err := strconv.ParseUint(p.moduleVersionID, 10, 32)
	if err != nil {
		return nil, err
	}

	moduleVersionResp, err := p.facade.GetModuleVersion(ctx, versource.GetModuleVersionRequest{ModuleVersionID: uint(moduleVersionIDUint)})
	if err != nil {
		return nil, err
	}

	return moduleVersionResp, nil
}

func (p *VersionDetailData) ResolveData(data versource.GetModuleVersionResponse) VersionDetailViewModel {
	return VersionDetailViewModel{
		ID:      data.ModuleVersion.ID,
		Version: data.ModuleVersion.Version,
		Module: struct {
			ID           uint   `yaml:"id"`
			Name         string `yaml:"name"`
			Source       string `yaml:"source"`
			ExecutorType string `yaml:"executorType"`
		}{
			ID:           data.ModuleVersion.Module.ID,
			Name:         data.ModuleVersion.Module.Name,
			Source:       data.ModuleVersion.Module.Source,
			ExecutorType: data.ModuleVersion.Module.ExecutorType,
		},
	}
}

func (p *VersionDetailData) KeyBindings(elem versource.GetModuleVersionResponse) platform.KeyBindings {
	moduleVersionIDUint, err := strconv.ParseUint(p.moduleVersionID, 10, 32)
	if err != nil {
		return platform.KeyBindings{}
	}

	return platform.KeyBindings{
		{Key: "m", Help: "View module", Command: fmt.Sprintf("modules/%d", moduleVersionIDUint)},
		{Key: "c", Help: "View components", Command: fmt.Sprintf("components?module-version-id=%d", moduleVersionIDUint)},
	}
}
