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

type ChangesetDiffDetailViewModel struct {
	DiffType string `yaml:"diffType"`
	From     *struct {
		ID     uint   `yaml:"id,omitempty"`
		Name   string `yaml:"name,omitempty"`
		Status string `yaml:"status,omitempty"`
		Module *struct {
			ID      uint   `yaml:"id"`
			Source  string `yaml:"source"`
			Version *struct {
				ID      uint   `yaml:"id"`
				Version string `yaml:"version"`
			} `yaml:"version,omitempty"`
		} `yaml:"module,omitempty"`
		Variables map[string]any `yaml:"variables,omitempty"`
	} `yaml:"from,omitempty"`
	To *struct {
		ID     uint   `yaml:"id,omitempty"`
		Name   string `yaml:"name,omitempty"`
		Status string `yaml:"status,omitempty"`
		Module *struct {
			ID      uint   `yaml:"id"`
			Source  string `yaml:"source"`
			Version *struct {
				ID      uint   `yaml:"id"`
				Version string `yaml:"version"`
			} `yaml:"version,omitempty"`
		} `yaml:"module,omitempty"`
		Variables map[string]any `yaml:"variables,omitempty"`
	} `yaml:"to,omitempty"`
	Plan *struct {
		ID      uint   `yaml:"id"`
		State   string `yaml:"state"`
		Add     *int   `yaml:"add,omitempty"`
		Change  *int   `yaml:"change,omitempty"`
		Destroy *int   `yaml:"destroy,omitempty"`
	} `yaml:"plan,omitempty"`
}

func NewChangesetDiffDetail(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataViewport(NewChangesetDiffDetailData(
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

func (p *ChangesetDiffDetailData) ResolveData(data internal.GetComponentDiffResponse) string {
	viewModel := ChangesetDiffDetailViewModel{
		DiffType: string(data.Diff.DiffType),
	}

	if data.Diff.FromComponent != nil {
		fromComponent := data.Diff.FromComponent
		var fromModule *struct {
			ID      uint   `yaml:"id"`
			Source  string `yaml:"source"`
			Version *struct {
				ID      uint   `yaml:"id"`
				Version string `yaml:"version"`
			} `yaml:"version,omitempty"`
		}
		if fromComponent.ModuleVersion.Module.ID != 0 {
			var version *struct {
				ID      uint   `yaml:"id"`
				Version string `yaml:"version"`
			}
			if fromComponent.ModuleVersion.ID != 0 {
				version = &struct {
					ID      uint   `yaml:"id"`
					Version string `yaml:"version"`
				}{
					ID:      fromComponent.ModuleVersion.ID,
					Version: fromComponent.ModuleVersion.Version,
				}
			}

			fromModule = &struct {
				ID      uint   `yaml:"id"`
				Source  string `yaml:"source"`
				Version *struct {
					ID      uint   `yaml:"id"`
					Version string `yaml:"version"`
				} `yaml:"version,omitempty"`
			}{
				ID:      fromComponent.ModuleVersion.Module.ID,
				Source:  fromComponent.ModuleVersion.Module.Source,
				Version: version,
			}
		}

		var fromVariables map[string]any
		if fromComponent.Variables != nil {
			err := json.Unmarshal(fromComponent.Variables, &fromVariables)
			if err != nil {
				fromVariables = nil
			}
		}

		viewModel.From = &struct {
			ID     uint   `yaml:"id,omitempty"`
			Name   string `yaml:"name,omitempty"`
			Status string `yaml:"status,omitempty"`
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
			ID:        fromComponent.ID,
			Name:      fromComponent.Name,
			Status:    string(fromComponent.Status),
			Module:    fromModule,
			Variables: fromVariables,
		}
	}

	if data.Diff.ToComponent != nil {
		toComponent := data.Diff.ToComponent
		var toModule *struct {
			ID      uint   `yaml:"id"`
			Source  string `yaml:"source"`
			Version *struct {
				ID      uint   `yaml:"id"`
				Version string `yaml:"version"`
			} `yaml:"version,omitempty"`
		}
		if toComponent.ModuleVersion.Module.ID != 0 {
			var version *struct {
				ID      uint   `yaml:"id"`
				Version string `yaml:"version"`
			}
			if toComponent.ModuleVersion.ID != 0 {
				version = &struct {
					ID      uint   `yaml:"id"`
					Version string `yaml:"version"`
				}{
					ID:      toComponent.ModuleVersion.ID,
					Version: toComponent.ModuleVersion.Version,
				}
			}

			toModule = &struct {
				ID      uint   `yaml:"id"`
				Source  string `yaml:"source"`
				Version *struct {
					ID      uint   `yaml:"id"`
					Version string `yaml:"version"`
				} `yaml:"version,omitempty"`
			}{
				ID:      toComponent.ModuleVersion.Module.ID,
				Source:  toComponent.ModuleVersion.Module.Source,
				Version: version,
			}
		}

		var toVariables map[string]any
		if toComponent.Variables != nil {
			err := json.Unmarshal(toComponent.Variables, &toVariables)
			if err != nil {
				toVariables = nil
			}
		}

		viewModel.To = &struct {
			ID     uint   `yaml:"id,omitempty"`
			Name   string `yaml:"name,omitempty"`
			Status string `yaml:"status,omitempty"`
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
			ID:        toComponent.ID,
			Name:      toComponent.Name,
			Status:    string(toComponent.Status),
			Module:    toModule,
			Variables: toVariables,
		}
	}

	if data.Diff.Plan != nil {
		viewModel.Plan = &struct {
			ID      uint   `yaml:"id"`
			State   string `yaml:"state"`
			Add     *int   `yaml:"add,omitempty"`
			Change  *int   `yaml:"change,omitempty"`
			Destroy *int   `yaml:"destroy,omitempty"`
		}{
			ID:      data.Diff.Plan.ID,
			State:   data.Diff.Plan.State,
			Add:     data.Diff.Plan.Add,
			Change:  data.Diff.Plan.Change,
			Destroy: data.Diff.Plan.Destroy,
		}
	}

	yamlData, err := yaml.Marshal(viewModel)
	if err != nil {
		return fmt.Sprintf("Error marshaling to YAML: %v", err)
	}

	return string(yamlData)
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
