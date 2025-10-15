package tfexec

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/pkg/versource"
	"gorm.io/datatypes"
)

type Executor struct {
	component *versource.Component
	tf        *tfexec.Terraform
}

func NewExecutor(component *versource.Component, workdir string, logs io.Writer) (internal.Executor, error) {
	tf, err := tfexec.NewTerraform(workdir, "terraform")
	if err != nil {
		return nil, fmt.Errorf("failed to create terraform instance: %w", err)
	}

	if logs != nil {
		tf.SetStdout(logs)
		tf.SetStderr(logs)
	}

	return &Executor{
		component: component,
		tf:        tf,
	}, nil
}

func (t *Executor) Init(ctx context.Context) error {
	return t.tf.Init(ctx)
}

func (t *Executor) Plan(ctx context.Context) (internal.PlanPath, internal.PlanResourceCounts, error) {
	tempFile, err := os.CreateTemp("", "plan-*.tfplan")
	if err != nil {
		return "", internal.PlanResourceCounts{}, fmt.Errorf("failed to create temp plan file: %w", err)
	}
	defer tempFile.Close()

	planPath := tempFile.Name()

	var planOptions []tfexec.PlanOption
	planOptions = append(planOptions, tfexec.Out(planPath))

	if t.component.Status == versource.ComponentStatusDeleted {
		planOptions = append(planOptions, tfexec.Destroy(true))
	}

	_, err = t.tf.Plan(ctx, planOptions...)
	if err != nil {
		os.Remove(planPath)
		return "", internal.PlanResourceCounts{}, fmt.Errorf("failed to plan terraform: %w", err)
	}

	resourceCounts, err := extractResourceCountsFromPlan(ctx, t.tf, planPath)
	if err != nil {
		os.Remove(planPath)
		return "", internal.PlanResourceCounts{}, fmt.Errorf("failed to extract resource counts: %w", err)
	}

	return internal.PlanPath(planPath), resourceCounts, nil
}

func (t *Executor) Apply(ctx context.Context, planPath internal.PlanPath) (versource.State, []versource.StateResource, error) {
	err := t.tf.Apply(ctx, tfexec.DirOrPlan(string(planPath)))
	if err != nil {
		return versource.State{}, nil, fmt.Errorf("failed to apply terraform: %w", err)
	}

	tfState, err := t.tf.Show(ctx)
	if err != nil {
		return versource.State{}, nil, fmt.Errorf("failed to get terraform state: %w", err)
	}

	state, err := extractState(tfState)
	if err != nil {
		return versource.State{}, nil, fmt.Errorf("failed to extract state: %w", err)
	}

	var stateResources []versource.StateResource
	if tfState.Values != nil && tfState.Values.RootModule != nil {
		stateResources, err = extractResources(tfState.Values.RootModule)
		if err != nil {
			return versource.State{}, nil, fmt.Errorf("failed to extract resources: %w", err)
		}
	}

	return state, stateResources, nil
}

func (t *Executor) Close() error {
	return nil
}

func extractState(tfState *tfjson.State) (versource.State, error) {
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
		return versource.State{}, fmt.Errorf("failed to marshal output: %w", err)
	}

	state := versource.State{
		Output: datatypes.JSON(jsonOutput),
	}

	return state, nil
}

func extractResources(module *tfjson.StateModule) ([]versource.StateResource, error) {
	var stateResources []versource.StateResource

	for _, tfResource := range module.Resources {
		var count *int
		var forEach *string
		switch index := tfResource.Index.(type) {
		case int:
			count = &index
		case string:
			forEach = &index
		}

		providerInfo := parseProvider(tfResource.ProviderName)

		jsonAttributes, err := json.Marshal(tfResource.AttributeValues)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal attributes: %w", err)
		}

		resourceType := extractResourceType(tfResource.Type, providerInfo.Name)

		resource := versource.Resource{
			Provider:      providerInfo.Name,
			ProviderAlias: providerInfo.Alias,
			ResourceType:  resourceType,
			Namespace:     nil,
			Name:          tfResource.Name,
			Attributes:    datatypes.JSON(jsonAttributes),
		}

		stateResource := versource.StateResource{
			Address:      tfResource.Address,
			Mode:         versource.ResourceMode(tfResource.Mode),
			ProviderName: tfResource.ProviderName,
			Count:        count,
			ForEach:      forEach,
			Type:         tfResource.Type,
			Resource:     resource,
		}
		stateResources = append(stateResources, stateResource)
	}

	for _, childModule := range module.ChildModules {
		childStateResources, err := extractResources(childModule)
		if err != nil {
			return nil, fmt.Errorf("failed to extract resources: %w", err)
		}
		stateResources = append(stateResources, childStateResources...)
	}

	return stateResources, nil
}

type ProviderInfo struct {
	Name  string
	Alias *string
}

func parseProvider(providerName string) ProviderInfo {
	re := regexp.MustCompile(`provider\["([^"]+)"\](?:\.(.+))?`)
	matches := re.FindStringSubmatch(providerName)

	if len(matches) < 2 {
		return ProviderInfo{Name: providerName}
	}

	providerSource := matches[1]
	alias := matches[2]

	parts := strings.Split(providerSource, "/")
	providerNameFromSource := parts[len(parts)-1]

	result := ProviderInfo{Name: providerNameFromSource}
	if alias != "" {
		result.Alias = &alias
	}

	return result
}

func extractResourceType(tfType, providerName string) string {
	prefix := providerName + "_"
	if strings.HasPrefix(tfType, prefix) {
		return strings.TrimPrefix(tfType, prefix)
	}

	return tfType
}

func extractResourceCountsFromPlan(ctx context.Context, tf *tfexec.Terraform, planPath string) (internal.PlanResourceCounts, error) {
	counts := internal.PlanResourceCounts{}

	plan, err := tf.ShowPlanFile(ctx, planPath)
	if err != nil {
		return counts, fmt.Errorf("failed to show plan file: %w", err)
	}

	if plan.ResourceChanges != nil {
		for _, change := range plan.ResourceChanges {
			if change.Change != nil && change.Change.Actions != nil {
				if change.Change.Actions.Create() {
					counts.AddCount++
				}
				if change.Change.Actions.Update() {
					counts.ChangeCount++
				}
				if change.Change.Actions.Delete() {
					counts.DestroyCount++
				}
			}
		}
	}

	return counts, nil
}
