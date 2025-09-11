package module

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
		return platform.NewDataTable[internal.Module](&TableData{client: client})
	}
}

func (p *TableData) LoadData() ([]internal.Module, error) {
	ctx := context.Background()
	resp, err := p.client.ListModules(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Modules, nil
}

func (p *TableData) ResolveData(data []internal.Module) ([]table.Column, []table.Row, []internal.Module) {

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Source", Width: 9},
	}

	var rows []table.Row
	var elems []internal.Module
	for _, module := range data {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(module.ID), 10),
			module.Source,
		})
		elems = append(elems, module)
	}

	return columns, rows, elems
}

func (p *TableData) KeyBindings(elem internal.Module) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View module detail", Command: fmt.Sprintf("modules/%d", elem.ID)},
		{Key: "v", Help: "View module versions", Command: fmt.Sprintf("modules/%d/moduleversions", elem.ID)},
		{Key: "c", Help: "View components", Command: fmt.Sprintf("components?module-id=%d", elem.ID)},
		{Key: "C", Help: "Create module", Command: "modules/create"},
	}
}
