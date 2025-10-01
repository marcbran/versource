package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type RestoreData struct {
	facade        internal.Facade
	componentID   string
	changesetName string
}

func NewRestore(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewConfirmationPage(NewRestoreData(
			facade,
			params["componentID"],
			params["changesetName"],
		))
	}
}

func NewRestoreData(facade internal.Facade, componentID, changesetName string) *RestoreData {
	return &RestoreData{
		facade:        facade,
		componentID:   componentID,
		changesetName: changesetName,
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
		ComponentID:   uint(componentID),
		ChangesetName: r.changesetName,
	}

	_, err = r.facade.RestoreComponent(ctx, req)
	if err != nil {
		return "", err
	}

	return "components", nil
}
