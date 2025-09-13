package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

type RestoreData struct {
	client      *client.Client
	componentID string
	changeset   string
}

func NewRestore(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewConfirmationPage(NewRestoreData(client, params["componentID"], params["changeset"]))
	}
}

func NewRestoreData(client *client.Client, componentID, changeset string) *RestoreData {
	return &RestoreData{
		client:      client,
		componentID: componentID,
		changeset:   changeset,
	}
}

func (r *RestoreData) GetConfirmationDialog() platform.ConfirmationDialog {
	return platform.ConfirmationDialog{
		Title:       "Restore Component",
		Message:     fmt.Sprintf("Are you sure you want to restore component %s? This will set its status to Ready and restore it from the merge base state.", r.componentID),
		ConfirmText: "Restore",
		CancelText:  "Cancel",
	}
}

func (r *RestoreData) OnConfirm(ctx context.Context) (string, error) {
	componentID, err := strconv.ParseUint(r.componentID, 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid component ID: %w", err)
	}

	req := internal.RestoreComponentRequest{
		ComponentID: uint(componentID),
		Changeset:   r.changeset,
	}

	_, err = r.client.RestoreComponent(ctx, req)
	if err != nil {
		return "", err
	}

	return "components", nil
}
