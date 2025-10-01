package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type CreateComponentData struct {
	facade        internal.Facade
	moduleID      string
	changesetName string
}

func NewCreateComponent(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewEditor(&CreateComponentData{
			facade:        facade,
			moduleID:      params["module-id"],
			changesetName: params["changesetName"],
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
		ChangesetName: c.changesetName,
		ModuleID:      moduleID,
		Name:          "",
		Variables:     make(map[string]any),
	}
}

func (c *CreateComponentData) SaveData(ctx context.Context, data internal.CreateComponentRequest) (string, error) {
	if data.ChangesetName == "" {
		return "", fmt.Errorf("changeset is required")
	}

	if data.ModuleID == 0 {
		return "", fmt.Errorf("moduleId is required")
	}

	if data.Name == "" {
		return "", fmt.Errorf("name is required")
	}

	_, err := c.facade.CreateComponent(ctx, data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("changesets/%s/changes", data.ChangesetName), nil
}
