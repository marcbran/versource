package module

import (
	"context"
	"fmt"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type CreateModuleData struct {
	facade versource.Facade
}

func NewCreateModule(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewEditor(&CreateModuleData{facade: facade})
	}
}

func (c *CreateModuleData) GetInitialValue() (versource.CreateModuleRequest, error) {
	return versource.CreateModuleRequest{
		Name:         "",
		Source:       "",
		Version:      "",
		ExecutorType: "terraform-jsonnet",
	}, nil
}

func (c *CreateModuleData) SaveData(ctx context.Context, data versource.CreateModuleRequest) (string, error) {
	if data.Name == "" {
		return "", fmt.Errorf("name is required")
	}

	if data.Source == "" {
		return "", fmt.Errorf("source is required")
	}

	if data.Version == "" {
		return "", fmt.Errorf("version is required")
	}

	if data.ExecutorType == "" {
		data.ExecutorType = "terraform-jsonnet"
	}

	resp, err := c.facade.CreateModule(ctx, data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("modules/%d", resp.Module.ID), nil
}
