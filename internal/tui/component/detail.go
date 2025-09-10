package component

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui"
	"github.com/marcbran/versource/internal/tui/platform"
	"gopkg.in/yaml.v3"
)

type DetailData struct {
	client      *client.Client
	componentID string
	changeset   *string
}

type DetailViewModel struct {
	ID     uint   `yaml:"id"`
	Name   string `yaml:"name"`
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

func NewDetail(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		var changeset *string
		if changesetParam := params["changeset"]; changesetParam != "" {
			changeset = &changesetParam
		}
		return platform.NewDataViewport(&DetailData{
			client:      client,
			componentID: params["componentID"],
			changeset:   changeset,
		})
	}
}

func (p *DetailData) LoadData() (*internal.GetComponentResponse, error) {
	ctx := context.Background()

	componentIDUint, err := strconv.ParseUint(p.componentID, 10, 32)
	if err != nil {
		return nil, err
	}

	componentResp, err := p.client.GetComponent(ctx, uint(componentIDUint), p.changeset)
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
	return tui.KeyBindings.
		With("m", "View module", fmt.Sprintf("modules/%d", elem.Component.ModuleVersion.Module.ID)).
		With("v", "View module versions", fmt.Sprintf("modules/%d/moduleversions", elem.Component.ModuleVersion.Module.ID))
}
