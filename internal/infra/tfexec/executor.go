package tfexec

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/marcbran/versource/internal"
	"gorm.io/datatypes"
)

type TerraformExecutor struct {
	component *internal.Component
	workdir   string
	tf        *tfexec.Terraform
	tempDir   string
}

func NewExecutor(component *internal.Component, workdir string) (internal.Executor, error) {
	terraformStack, err := NewTerraformStackFromComponent(component, workdir)
	if err != nil {
		return nil, fmt.Errorf("failed to convert component to terraform stack: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "versource-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	modulesDir := filepath.Join(workdir, "modules")
	if _, err := os.Stat(modulesDir); err == nil {
		absModulesDir, err := filepath.Abs(modulesDir)
		if err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("failed to get absolute path for modules directory: %w", err)
		}
		modulesLink := filepath.Join(tempDir, "modules")
		err = os.Symlink(absModulesDir, modulesLink)
		if err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("failed to create modules symlink: %w", err)
		}
	}

	mainJSONPath := filepath.Join(tempDir, "main.tf.json")
	jsonData, err := json.MarshalIndent(terraformStack, "", "  ")
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to marshal stack config: %w", err)
	}

	err = os.WriteFile(mainJSONPath, jsonData, 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to write stack config: %w", err)
	}

	tf, err := tfexec.NewTerraform(tempDir, "terraform")
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to create terraform instance: %w", err)
	}

	return &TerraformExecutor{
		component: component,
		workdir:   workdir,
		tf:        tf,
		tempDir:   tempDir,
	}, nil
}

func (t *TerraformExecutor) Init(ctx context.Context, logs io.Writer) error {
	if logs != nil {
		t.tf.SetStdout(logs)
		t.tf.SetStderr(logs)
	}
	return t.tf.Init(ctx)
}

func (t *TerraformExecutor) Plan(ctx context.Context, logs io.Writer) (internal.PlanPath, error) {
	if logs != nil {
		t.tf.SetStdout(logs)
		t.tf.SetStderr(logs)
	}

	tempFile, err := os.CreateTemp("", "plan-*.tfplan")
	if err != nil {
		return "", fmt.Errorf("failed to create temp plan file: %w", err)
	}
	defer tempFile.Close()

	planPath := tempFile.Name()
	_, err = t.tf.Plan(ctx, tfexec.Out(planPath))
	if err != nil {
		os.Remove(planPath)
		return "", fmt.Errorf("failed to plan terraform: %w", err)
	}

	return internal.PlanPath(planPath), nil
}

func (t *TerraformExecutor) Apply(ctx context.Context, planPath internal.PlanPath, logs io.Writer) (internal.State, []internal.Resource, error) {
	if logs != nil {
		t.tf.SetStdout(logs)
		t.tf.SetStderr(logs)
	}

	err := t.tf.Apply(ctx, tfexec.DirOrPlan(string(planPath)))
	if err != nil {
		return internal.State{}, nil, fmt.Errorf("failed to apply terraform: %w", err)
	}

	tfState, err := t.tf.Show(ctx)
	if err != nil {
		return internal.State{}, nil, fmt.Errorf("failed to get terraform state: %w", err)
	}

	state, err := extractState(tfState)
	if err != nil {
		return internal.State{}, nil, fmt.Errorf("failed to extract state: %w", err)
	}

	var resources []internal.Resource
	if tfState.Values != nil && tfState.Values.RootModule != nil {
		resources, err = extractResources(tfState.Values.RootModule)
		if err != nil {
			return internal.State{}, nil, fmt.Errorf("failed to extract resources: %w", err)
		}
	}

	return state, resources, nil
}

func (t *TerraformExecutor) Close() error {
	os.RemoveAll(t.tempDir)
	return nil
}

func extractState(tfState *tfjson.State) (internal.State, error) {
	output := make(map[string]any)
	for name, out := range tfState.Values.Outputs {
		if out == nil {
			continue
		}
		if out.Sensitive {
			continue
		}
		output[name] = out.Value
	}

	jsonOutput, err := json.Marshal(output)
	if err != nil {
		return internal.State{}, fmt.Errorf("failed to marshal output: %w", err)
	}

	state := internal.State{
		Output: datatypes.JSON(jsonOutput),
	}

	return state, nil
}

func extractResources(module *tfjson.StateModule) ([]internal.Resource, error) {
	var resources []internal.Resource

	for _, tfResource := range module.Resources {
		var count *int
		var forEach *string
		switch index := tfResource.Index.(type) {
		case int:
			count = &index
		case string:
			forEach = &index
		}
		jsonAttributes, err := json.Marshal(tfResource.AttributeValues)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal attributes: %w", err)
		}
		resource := internal.Resource{
			Address:      tfResource.Address,
			Mode:         internal.ResourceMode(tfResource.Mode),
			ProviderName: tfResource.ProviderName,
			Count:        count,
			ForEach:      forEach,
			Type:         tfResource.Type,
			Attributes:   datatypes.JSON(jsonAttributes),
		}
		resources = append(resources, resource)
	}

	for _, childModule := range module.ChildModules {
		childResources, err := extractResources(childModule)
		if err != nil {
			return nil, fmt.Errorf("failed to extract resources: %w", err)
		}
		resources = append(resources, childResources...)
	}

	return resources, nil
}
