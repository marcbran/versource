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

type ChangesetDiffDetailData struct {
	facade        internal.Facade
	componentID   string
	changesetName string
}

func NewChangesetDiffDetail(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDiffView(NewChangesetDiffDetailData(
			facade,
			params["componentID"],
			params["changesetName"],
		))
	}
}

func NewChangesetDiffDetailData(facade internal.Facade, componentID string, changesetName string) *ChangesetDiffDetailData {
	return &ChangesetDiffDetailData{
		facade:        facade,
		componentID:   componentID,
		changesetName: changesetName,
	}
}

func (p *ChangesetDiffDetailData) LoadData() (*internal.GetComponentDiffResponse, error) {
	ctx := context.Background()

	componentIDUint, err := strconv.ParseUint(p.componentID, 10, 32)
	if err != nil {
		return nil, err
	}

	diffResp, err := p.facade.GetComponentDiff(ctx, internal.GetComponentDiffRequest{
		ComponentID: uint(componentIDUint),
		Changeset:   p.changesetName,
	})
	if err != nil {
		return nil, err
	}

	return diffResp, nil
}

func (p *ChangesetDiffDetailData) ResolveData(data internal.GetComponentDiffResponse) platform.Diff {
	var leftYAML, rightYAML string

	if data.Diff.FromComponent != nil {
		fromYAML, err := p.componentToYAML(data.Diff.FromComponent)
		if err != nil {
			leftYAML = fmt.Sprintf("Error: %v", err)
		} else {
			leftYAML = fromYAML
		}
	} else {
		leftYAML = ""
	}

	if data.Diff.ToComponent != nil {
		toYAML, err := p.componentToYAML(data.Diff.ToComponent)
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

func (p *ChangesetDiffDetailData) componentToYAML(component *internal.Component) (string, error) {
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

func (p *ChangesetDiffDetailData) KeyBindings(elem internal.GetComponentDiffResponse) platform.KeyBindings {
	keyBindings := platform.KeyBindings{}

	if elem.Diff.ToComponent != nil {
		keyBindings = append(keyBindings, platform.KeyBinding{
			Key:     "c",
			Help:    "View component detail",
			Command: fmt.Sprintf("changesets/%s/components/%d", p.changesetName, elem.Diff.ToComponent.ID),
		})
	}

	if elem.Diff.ToComponent != nil && elem.Diff.ToComponent.ModuleVersion.Module.ID != 0 {
		keyBindings = append(keyBindings, platform.KeyBinding{
			Key:     "m",
			Help:    "View module",
			Command: fmt.Sprintf("modules/%d", elem.Diff.ToComponent.ModuleVersion.Module.ID),
		})
	}

	if elem.Diff.Plan != nil {
		keyBindings = append(keyBindings, platform.KeyBinding{
			Key:     "p",
			Help:    "View plan",
			Command: fmt.Sprintf("changesets/%s/plans/%d", p.changesetName, elem.Diff.Plan.ID),
		})
	}

	return keyBindings
}
