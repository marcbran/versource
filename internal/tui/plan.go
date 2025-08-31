package tui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal"
)

func getPlansTable(plans []internal.Plan) ([]table.Column, []table.Row, []string) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Component", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	var ids []string
	for _, plan := range plans {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(plan.ID), 10),
			strconv.FormatUint(uint64(plan.ComponentID), 10),
			plan.Changeset.Name,
			plan.State,
		})
		ids = append(ids, strconv.FormatUint(uint64(plan.ID), 10))
	}

	return columns, rows, ids
}

type PlansPage struct {
	app *App
}

func (p *PlansPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.app.client.ListPlans(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "plans", data: resp.Plans}
	}
}

func (p *PlansPage) Links(params map[string]string) map[string]string {
	return map[string]string{}
}

type ChangesetPlansPage struct {
	app *App
}

func (p *ChangesetPlansPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		changesetName := params["changesetName"]

		resp, err := p.app.client.ListPlans(ctx)
		if err != nil {
			return errorMsg{err: err}
		}

		view := fmt.Sprintf("changesets/%s/plans", changesetName)
		return dataLoadedMsg{view: view, data: resp.Plans}
	}
}

func (p *ChangesetPlansPage) Links(params map[string]string) map[string]string {
	changesetName := params["changesetName"]
	return map[string]string{
		"c": fmt.Sprintf("changesets/%s/components", changesetName),
		"a": fmt.Sprintf("changesets/%s/applies", changesetName),
	}
}
