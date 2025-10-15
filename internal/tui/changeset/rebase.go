package changeset

import (
	"context"
	"fmt"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type RebaseChangesetData struct {
	facade        versource.Facade
	changesetName string
}

func NewRebaseChangeset(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewConfirmationPage(&RebaseChangesetData{facade: facade, changesetName: params["changesetName"]})
	}
}

func (r *RebaseChangesetData) GetConfirmationDialog() platform.ConfirmationDialog {
	return platform.ConfirmationDialog{
		Title:       "Rebase Changeset",
		Message:     fmt.Sprintf("Are you sure you want to rebase changeset '%s'?\n\nThis will rebase all commits from the changeset onto the tip of main.", r.changesetName),
		ConfirmText: "rebase",
		CancelText:  "cancel",
	}
}

func (r *RebaseChangesetData) OnConfirm(ctx context.Context) (string, error) {
	_, err := r.facade.CreateRebase(ctx, versource.CreateRebaseRequest{ChangesetName: r.changesetName})
	if err != nil {
		return "", err
	}
	return "changesets", nil
}
