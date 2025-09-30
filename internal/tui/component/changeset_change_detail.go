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

type ChangesetChangeDetailData struct {
	facade        internal.Facade
	componentID   string
	changesetName string
}

func NewChangesetChangeDetail(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDiffView(NewChangesetChangeDetailData(
			facade,
			params["componentID"],
			params["changesetName"],
		))
	}
}

func NewChangesetChangeDetailData(facade internal.Facade, componentID string, changesetName string) *ChangesetChangeDetailData {
	return &ChangesetChangeDetailData{
		facade:        facade,
		componentID:   componentID,
		changesetName: changesetName,
	}
}

func (p *ChangesetChangeDetailData) LoadData() (*internal.GetComponentChangeResponse, error) {
	ctx := context.Background()

	componentIDUint, err := strconv.ParseUint(p.componentID, 10, 32)
	if err != nil {
		return nil, err
	}

	changeResp, err := p.facade.GetComponentChange(ctx, internal.GetComponentChangeRequest{
		ComponentID: uint(componentIDUint),
		Changeset:   p.changesetName,
	})
	if err != nil {
		return nil, err
	}

	return changeResp, nil
}

func (p *ChangesetChangeDetailData) ResolveData(data internal.GetComponentChangeResponse) platform.Diff {
	var leftYAML, rightYAML string

	if data.Change.FromComponent != nil {
		fromYAML, err := p.componentToYAML(data.Change.FromComponent)
		if err != nil {
			leftYAML = fmt.Sprintf("Error: %v", err)
		} else {
			leftYAML = fromYAML
		}
	} else {
		leftYAML = ""
	}

	if data.Change.ToComponent != nil {
		toYAML, err := p.componentToYAML(data.Change.ToComponent)
		if err != nil {
			rightYAML = fmt.Sprintf("Error: %v", err)
		} else {
			rightYAML = toYAML
		}
	} else {
		rightYAML = ""
	}

	return platform.Diff{
		Left:  leftYAML,
		Right: rightYAML,
	}
}

func (p *ChangesetChangeDetailData) componentToYAML(component *internal.Component) (string, error) {
	viewModel := struct {
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
	}{
		ID:     component.ID,
		Name:   component.Name,
		Status: string(component.Status),
	}

	if component.ModuleVersion.Module.ID != 0 {
		var version *struct {
			ID      uint   `yaml:"id"`
			Version string `yaml:"version"`
		}
		if component.ModuleVersion.ID != 0 {
			version = &struct {
				ID      uint   `yaml:"id"`
				Version string `yaml:"version"`
			}{
				ID:      component.ModuleVersion.ID,
				Version: component.ModuleVersion.Version,
			}
		}

		viewModel.Module = &struct {
			ID      uint   `yaml:"id"`
			Source  string `yaml:"source"`
			Version *struct {
				ID      uint   `yaml:"id"`
				Version string `yaml:"version"`
			} `yaml:"version,omitempty"`
		}{
			ID:      component.ModuleVersion.Module.ID,
			Source:  component.ModuleVersion.Module.Source,
			Version: version,
		}
	}

	if component.Variables != nil {
		err := json.Unmarshal(component.Variables, &viewModel.Variables)
		if err != nil {
			viewModel.Variables = nil
		}
	}

	yamlData, err := yaml.Marshal(viewModel)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

func (p *ChangesetChangeDetailData) KeyBindings(elem internal.GetComponentChangeResponse) platform.KeyBindings {
	keyBindings := platform.KeyBindings{
		{Key: "esc", Help: "View changes", Command: fmt.Sprintf("changesets/%s/changes", p.changesetName)},
	}

	if elem.Change.ToComponent != nil {
		keyBindings = append(keyBindings, platform.KeyBinding{
			Key:     "c",
			Help:    "View component detail",
			Command: fmt.Sprintf("changesets/%s/components/%d", p.changesetName, elem.Change.ToComponent.ID),
		})
	}

	if elem.Change.ToComponent != nil && elem.Change.ToComponent.ModuleVersion.Module.ID != 0 {
		keyBindings = append(keyBindings, platform.KeyBinding{
			Key:     "m",
			Help:    "View module",
			Command: fmt.Sprintf("modules/%d", elem.Change.ToComponent.ModuleVersion.Module.ID),
		})
	}

	if elem.Change.Plan != nil {
		keyBindings = append(keyBindings, platform.KeyBinding{
			Key:     "p",
			Help:    "View plan",
			Command: fmt.Sprintf("changesets/%s/plans/%d", p.changesetName, elem.Change.Plan.ID),
		})
	}

	return keyBindings
}
