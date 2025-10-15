package changeset

import (
	"context"
	"fmt"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type DeleteChangesetData struct {
	facade        versource.Facade
	changesetName string
}

func NewDeleteChangeset(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewConfirmationPage(&DeleteChangesetData{facade: facade, changesetName: params["changesetName"]})
	}
}

func (d *DeleteChangesetData) GetConfirmationDialog() platform.ConfirmationDialog {
	return platform.ConfirmationDialog{
		Title:       "Delete Changeset",
		Message:     fmt.Sprintf("Are you sure you want to delete changeset '%s'?\n\nThis will permanently remove the changeset, its branch, all plans, applies, and logs. This action cannot be undone.", d.changesetName),
		ConfirmText: "delete",
		CancelText:  "cancel",
	}
}

func (d *DeleteChangesetData) OnConfirm(ctx context.Context) (string, error) {
	_, err := d.facade.DeleteChangeset(ctx, versource.DeleteChangesetRequest{ChangesetName: d.changesetName})
	if err != nil {
		return "", err
	}
	return "changesets", nil
}
