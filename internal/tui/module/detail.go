package module

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
	"gopkg.in/yaml.v3"
)

type DetailData struct {
	client   *client.Client
	moduleID string
}

type DetailViewModel struct {
	ID            uint   `yaml:"id"`
	Source        string `yaml:"source"`
	ExecutorType  string `yaml:"executorType"`
	LatestVersion *struct {
		ID      uint   `yaml:"id"`
		Version string `yaml:"version"`
	} `yaml:"latestVersion,omitempty"`
}

func NewDetail(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataViewport(&DetailData{client: client, moduleID: params["moduleID"]})
	}
}

func (p *DetailData) LoadData() (*internal.GetModuleResponse, error) {
	ctx := context.Background()

	moduleIDUint, err := strconv.ParseUint(p.moduleID, 10, 32)
	if err != nil {
		return nil, err
	}

	moduleResp, err := p.client.GetModule(ctx, uint(moduleIDUint))
	if err != nil {
		return nil, err
	}

	return moduleResp, nil
}

func (p *DetailData) ResolveData(data internal.GetModuleResponse) string {
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

	viewModel := DetailViewModel{
		ID:            data.Module.ID,
		Source:        data.Module.Source,
		ExecutorType:  data.Module.ExecutorType,
		LatestVersion: latestVersion,
	}

	yamlData, err := yaml.Marshal(viewModel)
	if err != nil {
		return fmt.Sprintf("Error marshaling to YAML: %v", err)
	}

	return string(yamlData)
}

func (p *DetailData) KeyBindings(elem internal.GetModuleResponse) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "v", Help: "View all versions", Command: fmt.Sprintf("modules/%s/moduleversions", p.moduleID)},
		{Key: "c", Help: "View components", Command: fmt.Sprintf("components?module-id=%s", p.moduleID)},
	}
}
