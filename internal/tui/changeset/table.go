package changeset

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type TableData struct {
	facade versource.Facade
}

func NewTable(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable(NewTableData(facade))
	}
}

func NewTableData(facade versource.Facade) *TableData {
	return &TableData{facade: facade}
}

func (p *TableData) LoadData() ([]versource.Changeset, error) {
	ctx := context.Background()
	resp, err := p.facade.ListChangesets(ctx, versource.ListChangesetsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Changesets, nil
}

func (p *TableData) ResolveData(data []versource.Changeset) ([]table.Column, []table.Row, []versource.Changeset) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 7},
		{Title: "State", Width: 2},
		{Title: "Review", Width: 2},
	}

	var rows []table.Row
	var elems []versource.Changeset
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

func (p *TableData) ElemKeyBindings(elem versource.Changeset) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View changes", Command: fmt.Sprintf("changesets/%s/changes", elem.Name)},
		{Key: "M", Help: "Merge changeset", Command: fmt.Sprintf("changesets/%s/merge", elem.Name)},
		{Key: "R", Help: "Rebase changeset", Command: fmt.Sprintf("changesets/%s/rebase", elem.Name)},
		{Key: "D", Help: "Delete changeset", Command: fmt.Sprintf("changesets/%s/delete", elem.Name)},
	}
}
