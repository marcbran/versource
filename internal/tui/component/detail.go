package component

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
	"gopkg.in/yaml.v3"
)

type DetailData struct {
	facade        internal.Facade
	componentID   string
	changesetName string
}

type DetailViewModel struct {
	ID     uint   `yaml:"id"`
	Name   string `yaml:"name"`
	Status string `yaml:"status"`
	Module *struct {
		ID      uint   `yaml:"id"`
		Source  string `yaml:"source"`
		Version *struct {
			ID      uint   `yaml:"id"`
			Version string `yaml:"version"`
		} `yaml:"version,omitempty"`
	} `yaml:"module,omitempty"`
	Variables map[string]any `yaml:"variables,omitempty"`
}

func NewDetail(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataViewport(NewDetailData(
			facade,
			params["componentID"],
			params["changesetName"],
		))
	}
}

func NewDetailData(facade internal.Facade, componentID string, changesetName string) *DetailData {
	return &DetailData{
		facade:        facade,
		componentID:   componentID,
		changesetName: changesetName,
	}
}

func (p *DetailData) LoadData() (*internal.GetComponentResponse, error) {
	ctx := context.Background()

	componentIDUint, err := strconv.ParseUint(p.componentID, 10, 32)
	if err != nil {
		return nil, err
	}

	componentResp, err := p.facade.GetComponent(ctx, internal.GetComponentRequest{ComponentID: uint(componentIDUint), ChangesetName: &p.changesetName})
	if err != nil {
		return nil, err
	}

	return componentResp, nil
}

func (p *DetailData) ResolveData(data internal.GetComponentResponse) string {
	var module *struct {
		ID      uint   `yaml:"id"`
		Source  string `yaml:"source"`
		Version *struct {
			ID      uint   `yaml:"id"`
			Version string `yaml:"version"`
		} `yaml:"version,omitempty"`
	}
	if data.Component.ModuleVersion.Module.ID != 0 {
		var version *struct {
			ID      uint   `yaml:"id"`
			Version string `yaml:"version"`
		}
		if data.Component.ModuleVersion.ID != 0 {
			version = &struct {
				ID      uint   `yaml:"id"`
				Version string `yaml:"version"`
			}{
				ID:      data.Component.ModuleVersion.ID,
				Version: data.Component.ModuleVersion.Version,
			}
		}

		module = &struct {
			ID      uint   `yaml:"id"`
			Source  string `yaml:"source"`
			Version *struct {
				ID      uint   `yaml:"id"`
				Version string `yaml:"version"`
			} `yaml:"version,omitempty"`
		}{
			ID:      data.Component.ModuleVersion.Module.ID,
			Source:  data.Component.ModuleVersion.Module.Source,
			Version: version,
		}
	}

	var variables map[string]any
	if data.Component.Variables != nil {
		err := json.Unmarshal(data.Component.Variables, &variables)
		if err != nil {
			variables = nil
		}
	}

	viewModel := DetailViewModel{
		ID:        data.Component.ID,
		Name:      data.Component.Name,
		Status:    string(data.Component.Status),
		Module:    module,
		Variables: variables,
	}

	yamlData, err := yaml.Marshal(viewModel)
	if err != nil {
		return fmt.Sprintf("Error marshaling to YAML: %v", err)
	}

	return string(yamlData)
}

func (p *DetailData) KeyBindings(elem internal.GetComponentResponse) platform.KeyBindings {
	changesetPrefix := ""
	if p.changesetName != "" {
		changesetPrefix = fmt.Sprintf("changesets/%s", p.changesetName)
	}
	return platform.KeyBindings{
		{Key: "m", Help: "View module", Command: fmt.Sprintf("modules/%d", elem.Component.ModuleVersion.Module.ID)},
		{Key: "v", Help: "View module versions", Command: fmt.Sprintf("modules/%d/moduleversions", elem.Component.ModuleVersion.Module.ID)},
		{Key: "esc", Help: "View changes", Command: fmt.Sprintf("%s/%s/components", changesetPrefix, p.changesetName)},
		{Key: "E", Help: "Edit component", Command: fmt.Sprintf("%s/%s/components/%d/edit", changesetPrefix, p.changesetName, elem.Component.ID)},
		{Key: "D", Help: "Delete component", Command: fmt.Sprintf("%s/%s/components/%d/delete", changesetPrefix, p.changesetName, elem.Component.ID)},
	}
}
