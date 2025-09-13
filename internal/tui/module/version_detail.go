package module

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
	"gopkg.in/yaml.v3"
)

type VersionDetailData struct {
	facade          internal.Facade
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

func NewVersionDetail(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataViewport(NewVersionDetailData(facade, params["moduleVersionID"]))
	}
}

func NewVersionDetailData(facade internal.Facade, moduleVersionID string) *VersionDetailData {
	return &VersionDetailData{facade: facade, moduleVersionID: moduleVersionID}
}

func (p *VersionDetailData) LoadData() (*internal.GetModuleVersionResponse, error) {
	ctx := context.Background()

	moduleVersionIDUint, err := strconv.ParseUint(p.moduleVersionID, 10, 32)
	if err != nil {
		return nil, err
	}

	moduleVersionResp, err := p.facade.GetModuleVersion(ctx, internal.GetModuleVersionRequest{ModuleVersionID: uint(moduleVersionIDUint)})
	if err != nil {
		return nil, err
	}

	return moduleVersionResp, nil
}

func (p *VersionDetailData) ResolveData(data internal.GetModuleVersionResponse) string {
	viewModel := VersionDetailViewModel{
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

	yamlData, err := yaml.Marshal(viewModel)
	if err != nil {
		return fmt.Sprintf("Error marshaling to YAML: %v", err)
	}

	return string(yamlData)
}

func (p *VersionDetailData) KeyBindings(elem internal.GetModuleVersionResponse) platform.KeyBindings {
	moduleVersionIDUint, err := strconv.ParseUint(p.moduleVersionID, 10, 32)
	if err != nil {
		return platform.KeyBindings{}
	}

	return platform.KeyBindings{
		{Key: "m", Help: "View module", Command: fmt.Sprintf("modules/%d", moduleVersionIDUint)},
		{Key: "c", Help: "View components", Command: fmt.Sprintf("components?module-version-id=%d", moduleVersionIDUint)},
	}
}
