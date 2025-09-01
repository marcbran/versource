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
	client *client.Client
}

func NewAppliesPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return &AppliesPage{client: client}
	}
}

func (p *AppliesPage) Open() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.client.ListApplies(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "applies", data: resp.Applies}
	}
}

func (p *AppliesPage) Links() map[string]string {
	return map[string]string{
		"c": "components",
		"p": "plans",
	}
}

type ChangesetAppliesPage struct {
	client        *client.Client
	changesetName string
}

func NewChangesetAppliesPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return &ChangesetAppliesPage{
			client:        client,
			changesetName: params["changesetName"],
		}
	}
}

func (p *ChangesetAppliesPage) Open() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		resp, err := p.client.ListApplies(ctx)
		if err != nil {
			return errorMsg{err: err}
		}

		view := fmt.Sprintf("changesets/%s/applies", p.changesetName)
		return dataLoadedMsg{view: view, data: resp.Applies}
	}
}

func (p *ChangesetAppliesPage) Links() map[string]string {
	return map[string]string{
		"c": fmt.Sprintf("changesets/%s/components", p.changesetName),
		"p": fmt.Sprintf("changesets/%s/plans", p.changesetName),
	}
}
