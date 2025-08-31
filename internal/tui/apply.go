package tui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal"
)

func getAppliesTable(applies []internal.Apply) ([]table.Column, []table.Row, []string) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Plan", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	var ids []string
	for _, apply := range applies {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(apply.ID), 10),
			strconv.FormatUint(uint64(apply.PlanID), 10),
			apply.Changeset.Name,
			apply.State,
		})
		ids = append(ids, strconv.FormatUint(uint64(apply.ID), 10))
	}

	return columns, rows, ids
}

type AppliesPage struct {
	app *App
}

func (p *AppliesPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.app.client.ListApplies(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "applies", data: resp.Applies}
	}
}

func (p *AppliesPage) Links(params map[string]string) map[string]string {
	return map[string]string{}
}

type ChangesetAppliesPage struct {
	app *App
}

func (p *ChangesetAppliesPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		changesetName := params["changesetName"]

		resp, err := p.app.client.ListApplies(ctx)
		if err != nil {
			return errorMsg{err: err}
		}

		view := fmt.Sprintf("changesets/%s/applies", changesetName)
		return dataLoadedMsg{view: view, data: resp.Applies}
	}
}

func (p *ChangesetAppliesPage) Links(params map[string]string) map[string]string {
	changesetName := params["changesetName"]
	return map[string]string{
		"c": fmt.Sprintf("changesets/%s/components", changesetName),
		"p": fmt.Sprintf("changesets/%s/plans", changesetName),
	}
}
