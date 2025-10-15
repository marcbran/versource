package apply

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

func (p *TableData) LoadData() ([]versource.Apply, error) {
	ctx := context.Background()
	resp, err := p.facade.ListApplies(ctx, versource.ListAppliesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Applies, nil
}

func (p *TableData) ResolveData(data []versource.Apply) ([]table.Column, []table.Row, []versource.Apply) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Plan", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	var elems []versource.Apply
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

func (p *TableData) ElemKeyBindings(elem versource.Apply) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View apply details", Command: fmt.Sprintf("applies/%d", elem.ID)},
		{Key: "l", Help: "View logs", Command: fmt.Sprintf("applies/%d/logs", elem.ID)},
	}
}
