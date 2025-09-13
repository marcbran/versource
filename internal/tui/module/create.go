package module

import (
	"context"
	"fmt"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type CreateModuleData struct {
	facade internal.Facade
}

func NewCreateModule(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewEditor(&CreateModuleData{facade: facade})
	}
}

func (c *CreateModuleData) GetInitialValue() internal.CreateModuleRequest {
	return internal.CreateModuleRequest{
		Name:         "",
		Source:       "",
		Version:      "",
		ExecutorType: "terraform-jsonnet",
	}
}

func (c *CreateModuleData) SaveData(ctx context.Context, data internal.CreateModuleRequest) (string, error) {
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

	return fmt.Sprintf("modules/%d", resp.ID), nil
}
