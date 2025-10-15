package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type DeleteData struct {
	facade        versource.Facade
	componentID   string
	changesetName string
}

func NewDelete(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewConfirmationPage(NewDeleteData(facade, params["componentID"], params["changesetName"]))
	}
}

func NewDeleteData(facade versource.Facade, componentID, changesetName string) *DeleteData {
	return &DeleteData{
		facade:        facade,
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

	req := versource.DeleteComponentRequest{
		ComponentID:   uint(componentID),
		ChangesetName: d.changesetName,
	}

	_, err = d.facade.DeleteComponent(ctx, req)
	if err != nil {
		return "", err
	}

	return "components", nil
}
