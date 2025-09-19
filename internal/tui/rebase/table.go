package rebase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type TableData struct {
	facade        internal.Facade
	changesetName string
}

func NewTable(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable(NewTableData(facade, params["changesetName"]))
	}
}

func NewTableData(facade internal.Facade, changesetName string) *TableData {
	return &TableData{
		facade:        facade,
		changesetName: changesetName,
	}
}

func (p *TableData) LoadData() ([]internal.Rebase, error) {
	ctx := context.Background()
	req := internal.ListRebasesRequest{
		ChangesetName: p.changesetName,
	}
	resp, err := p.facade.ListRebases(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Rebases, nil
}

func (p *TableData) ResolveData(data []internal.Rebase) ([]table.Column, []table.Row, []internal.Rebase) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
		{Title: "Rebase Base", Width: 8},
		{Title: "Head", Width: 8},
	}

	var rows []table.Row
	var elems []internal.Rebase
	for _, rebase := range data {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(rebase.ID), 10),
			rebase.Changeset.Name,
			string(rebase.State),
			rebase.RebaseBase,
			rebase.Head,
		})
		elems = append(elems, rebase)
	}

	return columns, rows, elems
}

func (p *TableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *TableData) ElemKeyBindings(elem internal.Rebase) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View rebase detail", Command: fmt.Sprintf("rebases/%d", elem.ID)},
	}
}
