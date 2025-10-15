package changeset

import (
	"context"
	"fmt"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type MergeChangesetData struct {
	facade        versource.Facade
	changesetName string
}

func NewMergeChangeset(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewConfirmationPage(&MergeChangesetData{facade: facade, changesetName: params["changesetName"]})
	}
}

func (m *MergeChangesetData) GetConfirmationDialog() platform.ConfirmationDialog {
	return platform.ConfirmationDialog{
		Title:       "Merge Changeset",
		Message:     fmt.Sprintf("Are you sure you want to merge changeset '%s'?\n\nThis will merge all changes from the changeset into the main branch.", m.changesetName),
		ConfirmText: "merge",
		CancelText:  "cancel",
	}
}

func (m *MergeChangesetData) OnConfirm(ctx context.Context) (string, error) {
	_, err := m.facade.CreateMerge(ctx, versource.CreateMergeRequest{ChangesetName: m.changesetName})
	if err != nil {
		return "", err
	}
	return "changesets", nil
}
