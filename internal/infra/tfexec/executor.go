package tfexec

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/marcbran/versource/internal"
	"gorm.io/datatypes"
)

type Executor struct {
	tf *tfexec.Terraform
}

func NewExecutor(component *internal.Component, workdir string, logs io.Writer) (internal.Executor, error) {
	tf, err := tfexec.NewTerraform(workdir, "terraform")
	if err != nil {
		return nil, fmt.Errorf("failed to create terraform instance: %w", err)
	}

	if logs != nil {
		tf.SetStdout(logs)
		tf.SetStderr(logs)
	}

	return &Executor{
		tf: tf,
	}, nil
}

func (t *Executor) Init(ctx context.Context) error {
	return t.tf.Init(ctx)
}

func (t *Executor) Plan(ctx context.Context) (internal.PlanPath, error) {
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

func (t *Executor) Apply(ctx context.Context, planPath internal.PlanPath) (internal.State, []internal.Resource, error) {
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

func (t *Executor) Close() error {
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
