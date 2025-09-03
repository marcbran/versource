package tfexec

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

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
	terraformModule, err := buildTerraformModule(component)
	if err != nil {
		return nil, err
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

func buildTerraformModule(component *internal.Component) (TerraformModule, error) {
	source := component.ModuleVersion.Module.Source
	version := component.ModuleVersion.Version

	if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") {
		if version != "" {
			return TerraformModule{}, fmt.Errorf("local paths do not support version parameter")
		}
	} else if strings.HasPrefix(source, "github.com/") || strings.HasPrefix(source, "bitbucket.org/") || strings.HasPrefix(source, "git::") || strings.HasPrefix(source, "hg::") {
		if version != "" {
			if strings.Contains(source, "?") {
				source = source + "&ref=" + version
			} else {
				source = source + "?ref=" + version
			}
		}
	} else if strings.HasPrefix(source, "s3::") {
		if version != "" {
			if strings.Contains(source, "?") {
				source = source + "&versionId=" + version
			} else {
				source = source + "?versionId=" + version
			}
		}
	} else if strings.HasPrefix(source, "gcs::") {
		if version != "" {
			if strings.Contains(source, "?") {
				source = source + "&generation=" + version
			} else {
				source = source + "?generation=" + version
			}
		}
	} else if !strings.Contains(source, "::") && !strings.Contains(source, "://") {
		if version == "" {
			return TerraformModule{}, fmt.Errorf("terraform registry sources require version parameter")
		}
	}

	terraformModule := TerraformModule{
		Source: source,
	}

	if version != "" && !strings.HasPrefix(source, "./") && !strings.HasPrefix(source, "../") {
		if strings.HasPrefix(source, "github.com/") || strings.HasPrefix(source, "bitbucket.org/") || strings.HasPrefix(source, "git::") || strings.HasPrefix(source, "hg::") || strings.HasPrefix(source, "s3::") || strings.HasPrefix(source, "gcs::") {
			terraformModule.Version = ""
		} else {
			terraformModule.Version = version
		}
	}

	if component.Variables != nil {
		var variables map[string]any
		err := json.Unmarshal(component.Variables, &variables)
		if err != nil {
			return TerraformModule{}, fmt.Errorf("failed to unmarshal component variables: %w", err)
		}
		terraformModule.Variables = variables
	}

	return terraformModule, nil
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
