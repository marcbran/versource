package changeset

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

type TableData struct {
	client *client.Client
}

func NewTable(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable(NewTableData(client))
	}
}

func NewTableData(client *client.Client) *TableData {
	return &TableData{client: client}
}

func (p *TableData) LoadData() ([]internal.Changeset, error) {
	ctx := context.Background()
	resp, err := p.client.ListChangesets(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Changesets, nil
}

func (p *TableData) ResolveData(data []internal.Changeset) ([]table.Column, []table.Row, []internal.Changeset) {

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

func (p *TableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *TableData) ElemKeyBindings(elem internal.Changeset) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View component diffs", Command: fmt.Sprintf("changesets/%s/components", elem.Name)},
	}
}
