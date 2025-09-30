package apply

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type TableData struct {
	facade internal.Facade
}

func NewTable(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable(NewTableData(facade))
	}
}

func NewTableData(facade internal.Facade) *TableData {
	return &TableData{facade: facade}
}

func (p *TableData) LoadData() ([]internal.Apply, error) {
	ctx := context.Background()
	resp, err := p.facade.ListApplies(ctx, internal.ListAppliesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Applies, nil
}

func (p *TableData) ResolveData(data []internal.Apply) ([]table.Column, []table.Row, []internal.Apply) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Plan", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	var elems []internal.Apply
	for _, apply := range data {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(apply.ID), 10),
			strconv.FormatUint(uint64(apply.PlanID), 10),
			apply.Changeset.Name,
			string(apply.State),
		})
		elems = append(elems, apply)
	}

	return columns, rows, elems
}

func (p *TableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *TableData) ElemKeyBindings(elem internal.Apply) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View apply details", Command: fmt.Sprintf("applies/%d", elem.ID)},
		{Key: "l", Help: "View logs", Command: fmt.Sprintf("applies/%d/logs", elem.ID)},
	}
}
