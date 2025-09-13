package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type CreateComponentData struct {
	facade   internal.Facade
	moduleID string
}

func NewCreateComponent(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewEditor(&CreateComponentData{
			facade:   facade,
			moduleID: params["module-id"],
		})
	}
}

func (c *CreateComponentData) GetInitialValue() internal.CreateComponentRequest {
	moduleID := uint(0)
	if c.moduleID != "" {
		id, err := strconv.ParseUint(c.moduleID, 10, 32)
		if err == nil {
			moduleID = uint(id)
		}
	}

	return internal.CreateComponentRequest{
		Changeset: "",
		ModuleID:  moduleID,
		Name:      "",
		Variables: make(map[string]any),
	}
}

func (c *CreateComponentData) SaveData(ctx context.Context, data internal.CreateComponentRequest) (string, error) {
	if data.Changeset == "" {
		return "", fmt.Errorf("changeset is required")
	}

	if data.ModuleID == 0 {
		return "", fmt.Errorf("moduleId is required")
	}

	if data.Name == "" {
		return "", fmt.Errorf("name is required")
	}

	resp, err := c.facade.CreateComponent(ctx, data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("changesets/%s/components/%d", data.Changeset, resp.ID), nil
}
