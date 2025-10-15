package module

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type DetailData struct {
	facade   versource.Facade
	moduleID string
}

type DetailViewModel struct {
	ID            uint   `yaml:"id"`
	Name          string `yaml:"name"`
	Source        string `yaml:"source"`
	ExecutorType  string `yaml:"executorType"`
	LatestVersion *struct {
		ID      uint   `yaml:"id"`
		Version string `yaml:"version"`
	} `yaml:"latestVersion,omitempty"`
}

func NewDetail(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewViewDataViewport(NewDetailData(facade, params["moduleID"]))
	}
}

func NewDetailData(facade versource.Facade, moduleID string) *DetailData {
	return &DetailData{facade: facade, moduleID: moduleID}
}

func (p *DetailData) LoadData() (*versource.GetModuleResponse, error) {
	ctx := context.Background()

	moduleIDUint, err := strconv.ParseUint(p.moduleID, 10, 32)
	if err != nil {
		return nil, err
	}

	moduleResp, err := p.facade.GetModule(ctx, versource.GetModuleRequest{ModuleID: uint(moduleIDUint)})
	if err != nil {
		return nil, err
	}

	return moduleResp, nil
}

func (p *DetailData) ResolveData(data versource.GetModuleResponse) DetailViewModel {
	var latestVersion *struct {
		ID      uint   `yaml:"id"`
		Version string `yaml:"version"`
	}
	if data.LatestVersion != nil {
		latestVersion = &struct {
			ID      uint   `yaml:"id"`
			Version string `yaml:"version"`
		}{
			ID:      data.LatestVersion.ID,
			Version: data.LatestVersion.Version,
		}
	}

	return DetailViewModel{
		ID:            data.Module.ID,
		Name:          data.Module.Name,
		Source:        data.Module.Source,
		ExecutorType:  data.Module.ExecutorType,
		LatestVersion: latestVersion,
	}
}

func (p *DetailData) KeyBindings(elem versource.GetModuleResponse) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "v", Help: "View all versions", Command: fmt.Sprintf("modules/%s/moduleversions", p.moduleID)},
		{Key: "c", Help: "View components", Command: fmt.Sprintf("components?module-id=%s", p.moduleID)},
	}
}
