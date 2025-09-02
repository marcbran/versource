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

type ChangesetsTableData struct {
	client *client.Client
}

func NewChangesetsPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable(&ChangesetsTableData{client: client})
	}
}

func (p *ChangesetsTableData) LoadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.client.ListChangesets(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: resp.Changesets}
	}
}

func (p *ChangesetsTableData) ResolveData(data any) ([]table.Column, []table.Row, []any) {
	changesets, ok := data.([]internal.Changeset)
	if !ok {
		return nil, nil, nil
	}

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 7},
		{Title: "State", Width: 2},
		{Title: "Review", Width: 2},
	}

	var rows []table.Row
	var elems []any
	for _, changeset := range changesets {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(changeset.ID), 10),
			changeset.Name,
			string(changeset.State),
			string(changeset.ReviewState),
		})
		elems = append(elems, changeset)
	}

	return columns, rows, elems
}

func (p *ChangesetsTableData) Links(elem any) map[string]string {
	changeset, ok := elem.(internal.Changeset)
	if !ok {
		return map[string]string{}
	}
	return map[string]string{
		"enter": fmt.Sprintf("changesets/%s/components/diffs", changeset.Name),
		"c":     fmt.Sprintf("changesets/%s/components", changeset.Name),
		"p":     fmt.Sprintf("changesets/%s/plans", changeset.Name),
		"a":     fmt.Sprintf("changesets/%s/applies", changeset.Name),
	}
}
