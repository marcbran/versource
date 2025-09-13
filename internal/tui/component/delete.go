package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

type DeleteData struct {
	client        *client.Client
	componentID   string
	changesetName string
}

func NewDelete(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewConfirmationPage(NewDeleteData(client, params["componentID"], params["changesetName"]))
	}
}

func NewDeleteData(client *client.Client, componentID, changesetName string) *DeleteData {
	return &DeleteData{
		client:        client,
		componentID:   componentID,
		changesetName: changesetName,
	}
}

func (d *DeleteData) GetConfirmationDialog() platform.ConfirmationDialog {
	return platform.ConfirmationDialog{
		Title:       "Delete Component",
		Message:     fmt.Sprintf("Are you sure you want to delete component %s? This will set its status to Deleted and reset it to the merge base state.", d.componentID),
		ConfirmText: "Delete",
		CancelText:  "Cancel",
	}
}

func (d *DeleteData) OnConfirm(ctx context.Context) (string, error) {
	componentID, err := strconv.ParseUint(d.componentID, 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid component ID: %w", err)
	}

	req := internal.DeleteComponentRequest{
		ComponentID: uint(componentID),
		Changeset:   d.changesetName,
	}

	_, err = d.client.DeleteComponent(ctx, req)
	if err != nil {
		return "", err
	}

	return "components", nil
}
