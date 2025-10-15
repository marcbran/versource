package module

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type DeleteModuleData struct {
	facade   versource.Facade
	moduleID string
}

func NewDeleteModule(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewConfirmationPage(&DeleteModuleData{facade: facade, moduleID: params["moduleID"]})
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
	_, err = d.facade.DeleteModule(ctx, versource.DeleteModuleRequest{ModuleID: uint(moduleID)})
	if err != nil {
		return "", err
	}
	return "modules", nil
}
