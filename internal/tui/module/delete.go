package module

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

type DeleteModuleData struct {
	client   *client.Client
	moduleID string
}

func NewDeleteModule(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewConfirmationPage(&DeleteModuleData{client: client, moduleID: params["moduleID"]})
	}
}

func (d *DeleteModuleData) GetConfirmationDialog() platform.ConfirmationDialog {
	return platform.ConfirmationDialog{
		Title:       "Delete Module",
		Message:     fmt.Sprintf("Are you sure you want to delete module %s?\n\nThis action cannot be undone.", d.moduleID),
		ConfirmText: "delete",
		CancelText:  "cancel",
	}
}

func (d *DeleteModuleData) OnConfirm(ctx context.Context) (string, error) {
	moduleID, err := strconv.ParseUint(d.moduleID, 10, 32)
	if err != nil {
		return "", err
	}
	_, err = d.client.DeleteModule(ctx, uint(moduleID))
	if err != nil {
		return "", err
	}
	return "modules", nil
}
