package tui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
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
	client *client.Client
}

func NewPlansPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return &PlansPage{client: client}
	}
}

func (p *PlansPage) Open() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.client.ListPlans(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "plans", data: resp.Plans}
	}
}

func (p *PlansPage) Links() map[string]string {
	return map[string]string{
		"c": "components",
		"a": "applies",
	}
}

type ChangesetPlansPage struct {
	client        *client.Client
	changesetName string
}

func NewChangesetPlansPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return &ChangesetPlansPage{
			client:        client,
			changesetName: params["changesetName"],
		}
	}
}

func (p *ChangesetPlansPage) Open() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		resp, err := p.client.ListPlans(ctx)
		if err != nil {
			return errorMsg{err: err}
		}

		view := fmt.Sprintf("changesets/%s/plans", p.changesetName)
		return dataLoadedMsg{view: view, data: resp.Plans}
	}
}

func (p *ChangesetPlansPage) Links() map[string]string {
	return map[string]string{
		"c": fmt.Sprintf("changesets/%s/components", p.changesetName),
		"a": fmt.Sprintf("changesets/%s/applies", p.changesetName),
	}
}
