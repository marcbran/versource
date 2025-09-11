package component

import (
	"context"
	"fmt"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

type CreateComponentData struct {
	client *client.Client
}

func NewCreateComponent(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewEditor(&CreateComponentData{client: client})
	}
}

func (c *CreateComponentData) GetInitialValue() internal.CreateComponentRequest {
	return internal.CreateComponentRequest{
		Changeset: "",
		ModuleID:  0,
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

	resp, err := c.client.CreateComponent(ctx, data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("changesets/%s/components/%d", data.Changeset, resp.ID), nil
}
