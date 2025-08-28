package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
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

func NewTerraformStackFromComponent(component *Component, workDir string) (TerraformStack, error) {
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

func NewTerraformFromComponent(ctx context.Context, component *Component, workDir string) (*tfexec.Terraform, func(), error) {
	terraformStack, err := NewTerraformStackFromComponent(component, workDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert component to terraform stack: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "versource-*")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	modulesDir := filepath.Join(workDir, "modules")
	if _, err := os.Stat(modulesDir); err == nil {
		absModulesDir, err := filepath.Abs(modulesDir)
		if err != nil {
			os.RemoveAll(tempDir)
			return nil, nil, fmt.Errorf("failed to get absolute path for modules directory: %w", err)
		}
		modulesLink := filepath.Join(tempDir, "modules")
		err = os.Symlink(absModulesDir, modulesLink)
		if err != nil {
			os.RemoveAll(tempDir)
			return nil, nil, fmt.Errorf("failed to create modules symlink: %w", err)
		}
	}

	mainJSONPath := filepath.Join(tempDir, "main.tf.json")
	jsonData, err := json.MarshalIndent(terraformStack, "", "  ")
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, nil, fmt.Errorf("failed to marshal stack config: %w", err)
	}

	err = os.WriteFile(mainJSONPath, jsonData, 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, nil, fmt.Errorf("failed to write stack config: %w", err)
	}

	tf, err := tfexec.NewTerraform(tempDir, "terraform")
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, nil, fmt.Errorf("failed to create terraform instance: %w", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tf, cleanup, nil
}
