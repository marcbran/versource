package module

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

func (p *TableData) LoadData() ([]versource.Module, error) {
	ctx := context.Background()
	resp, err := p.facade.ListModules(ctx, versource.ListModulesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Modules, nil
}

func (p *TableData) ResolveData(data []versource.Module) ([]table.Column, []table.Row, []versource.Module) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 5},
		{Title: "Source", Width: 15},
	}

	var rows []table.Row
	var elems []versource.Module
	for _, module := range data {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(module.ID), 10),
			module.Name,
			module.Source,
		})
		elems = append(elems, module)
	}

	return columns, rows, elems
}

func (p *TableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "C", Help: "Create module", Command: "modules/create"},
	}
}

func (p *TableData) ElemKeyBindings(elem versource.Module) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View module detail", Command: fmt.Sprintf("modules/%d", elem.ID)},
		{Key: "v", Help: "View module versions", Command: fmt.Sprintf("modules/%d/moduleversions", elem.ID)},
		{Key: "D", Help: "Delete module", Command: fmt.Sprintf("modules/%d/delete", elem.ID)},
	}
}
