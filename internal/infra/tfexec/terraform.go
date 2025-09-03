package tfexec

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/marcbran/versource/internal"
)

type TerraformModule struct {
	Source    string         `json:"source"`
	Version   string         `json:"version,omitempty"`
	Variables map[string]any `json:"variables,omitempty"`
}

type TerraformOutput struct {
	Value any `json:"value"`
}

type TerraformBackend struct {
	Local TerraformBackendLocal `json:"local,omitempty"`
}

type TerraformBackendLocal struct {
	Path string `json:"path"`
}

type TerraformStack []any

func NewTerraformStackFromComponent(component *internal.Component, workDir string) (TerraformStack, error) {
	terraformModule := TerraformModule{
		Source: component.ModuleVersion.Module.Source,
	}

	if component.ModuleVersion.Version != "" {
		terraformModule.Version = component.ModuleVersion.Version
	}

	if component.Variables != nil {
		var variables map[string]any
		err := json.Unmarshal(component.Variables, &variables)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal component variables: %w", err)
		}
		terraformModule.Variables = variables
	}

	statePath, err := filepath.Abs(filepath.Join(workDir, "states", fmt.Sprintf("%d.tfstate", component.ID)))
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for state file: %w", err)
	}

	terraformStack := NewTerraformStack().
		AddModule("component", terraformModule).
		AddOutput("output", TerraformOutput{Value: "${module.component}"}).
		AddBackend("backend", TerraformBackend{
			Local: TerraformBackendLocal{
				Path: statePath,
			},
		})

	return terraformStack, nil
}

func NewTerraformStack() TerraformStack {
	return make(TerraformStack, 0)
}

type TerraformModuleContainer struct {
	Module map[string]TerraformModule `json:"module"`
}

func (tc TerraformStack) AddModule(name string, module TerraformModule) TerraformStack {
	container := TerraformModuleContainer{
		Module: map[string]TerraformModule{name: module},
	}
	return append(tc, container)
}

type TerraformOutputContainer struct {
	Output map[string]TerraformOutput `json:"output"`
}

func (tc TerraformStack) AddOutput(name string, output TerraformOutput) TerraformStack {
	container := TerraformOutputContainer{
		Output: map[string]TerraformOutput{name: output},
	}
	return append(tc, container)
}

type TerraformTerraformContainer struct {
	Terraform TerraformBackendContainer `json:"terraform"`
}

type TerraformBackendContainer struct {
	Backend TerraformBackend `json:"backend"`
}

func (tc TerraformStack) AddBackend(name string, backend TerraformBackend) TerraformStack {
	backendConfig := TerraformBackendContainer{
		Backend: backend,
	}
	container := TerraformTerraformContainer{
		Terraform: backendConfig,
	}
	return append(tc, container)
}
