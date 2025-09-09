package tui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

func changesetKeyBindings(changesetName string) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "m", Help: "View modules", Command: fmt.Sprintf("changesets/%s/modules", changesetName)},
		{Key: "c", Help: "View components", Command: fmt.Sprintf("changesets/%s/components", changesetName)},
		{Key: "d", Help: "View component diffs", Command: fmt.Sprintf("changesets/%s/components/diffs", changesetName)},
		{Key: "p", Help: "View plans", Command: fmt.Sprintf("changesets/%s/plans", changesetName)},
		{Key: "a", Help: "View applies", Command: fmt.Sprintf("changesets/%s/applies", changesetName)},
	}
}

type ChangesetsTableData struct {
	client *client.Client
}

func NewChangesetsPage(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.Changeset](&ChangesetsTableData{client: client})
	}
}

func (p *ChangesetsTableData) LoadData() ([]internal.Changeset, error) {
	ctx := context.Background()
	resp, err := p.client.ListChangesets(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Changesets, nil
}

func (p *ChangesetsTableData) ResolveData(data []internal.Changeset) ([]table.Column, []table.Row, []internal.Changeset) {

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 7},
		{Title: "State", Width: 2},
		{Title: "Review", Width: 2},
	}

	var rows []table.Row
	var elems []internal.Changeset
	for _, changeset := range data {
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

func (p *ChangesetsTableData) KeyBindings(elem internal.Changeset) platform.KeyBindings {
	return changesetKeyBindings(elem.Name).
		With("enter", "View component diffs", fmt.Sprintf("changesets/%s/components/diffs", elem.Name))
}
