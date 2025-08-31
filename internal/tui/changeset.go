package tui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal"
)

func getChangesetsTable(changesets []internal.Changeset) ([]table.Column, []table.Row, []string) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 7},
		{Title: "State", Width: 2},
		{Title: "Review", Width: 2},
	}

	var rows []table.Row
	var ids []string
	for _, changeset := range changesets {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(changeset.ID), 10),
			changeset.Name,
			string(changeset.State),
			string(changeset.ReviewState),
		})
		ids = append(ids, strconv.FormatUint(uint64(changeset.ID), 10))
	}

	return columns, rows, ids
}

type ChangesetsPage struct {
	app *App
}

func (p *ChangesetsPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.app.client.ListChangesets(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "changesets", data: resp.Changesets}
	}
}

func (p *ChangesetsPage) Links(params map[string]string) map[string]string {
	return map[string]string{}
}

type ChangesetPage struct {
	app *App
}

func (p *ChangesetPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		return dataLoadedMsg{view: fmt.Sprintf("changesets/%s", params["changesetName"]), data: nil}
	}
}

func (p *ChangesetPage) Links(params map[string]string) map[string]string {
	changesetName := params["changesetName"]
	return map[string]string{
		"enter": fmt.Sprintf("changesets/%s/components", changesetName),
		"c":     fmt.Sprintf("changesets/%s/components", changesetName),
		"p":     fmt.Sprintf("changesets/%s/plans", changesetName),
		"a":     fmt.Sprintf("changesets/%s/applies", changesetName),
	}
}
