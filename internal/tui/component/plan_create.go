package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

type CreatePlanData struct {
	client        *client.Client
	changesetName string
	componentID   string
}

func NewCreatePlan(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewConfirmationPage(&CreatePlanData{
			client:        client,
			changesetName: params["changesetName"],
			componentID:   params["componentID"],
		})
	}
}

func (c *CreatePlanData) GetConfirmationDialog() platform.ConfirmationDialog {
	return platform.ConfirmationDialog{
		Title:       "Create Plan",
		Message:     fmt.Sprintf("Are you sure you want to create a new plan for component ID %s in changeset '%s'?", c.componentID, c.changesetName),
		ConfirmText: "create",
		CancelText:  "cancel",
	}
}

func (c *CreatePlanData) OnConfirm(ctx context.Context) (string, error) {
	componentIDUint, err := strconv.ParseUint(c.componentID, 10, 32)
	if err != nil {
		return "", err
	}

	req := internal.CreatePlanRequest{
		ComponentID: uint(componentIDUint),
		Changeset:   c.changesetName,
	}

	_, err = c.client.CreatePlan(ctx, req)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("changesets/%s/plans", c.changesetName), nil
}
