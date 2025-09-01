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
	client *client.Client
}

func NewChangesetsPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return &ChangesetsPage{client: client}
	}
}

func (p *ChangesetsPage) Open() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.client.ListChangesets(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "changesets", data: resp.Changesets}
	}
}

func (p *ChangesetsPage) Links() map[string]string {
	return map[string]string{}
}

type ChangesetPage struct {
	client        *client.Client
	changesetName string
}

func NewChangesetPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return &ChangesetPage{
			client:        client,
			changesetName: params["changesetName"],
		}
	}
}

func (p *ChangesetPage) Open() tea.Cmd {
	return func() tea.Msg {
		view := fmt.Sprintf("changesets/%s", p.changesetName)
		return dataLoadedMsg{view: view, data: nil}
	}
}

func (p *ChangesetPage) Links() map[string]string {
	return map[string]string{
		"c": fmt.Sprintf("changesets/%s/components", p.changesetName),
		"p": fmt.Sprintf("changesets/%s/plans", p.changesetName),
		"a": fmt.Sprintf("changesets/%s/applies", p.changesetName),
	}
}
