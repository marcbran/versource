package resource

import (
	"context"
	"fmt"

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

func (p *TableData) LoadData() ([]internal.Resource, error) {
	ctx := context.Background()
	resp, err := p.facade.ListResources(ctx, internal.ListResourcesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Resources, nil
}

func (p *TableData) ResolveData(data []internal.Resource) ([]table.Column, []table.Row, []internal.Resource) {
	columns := []table.Column{
		{Title: "UUID", Width: 8},
		{Title: "Provider", Width: 4},
		{Title: "Alias", Width: 4},
		{Title: "Type", Width: 6},
		{Title: "Namespace", Width: 4},
		{Title: "Name", Width: 8},
	}

	var rows []table.Row
	var elems []internal.Resource
	for _, resource := range data {
		alias := ""
		if resource.ProviderAlias != nil {
			alias = *resource.ProviderAlias
		}
		namespace := ""
		if resource.Namespace != nil {
			namespace = *resource.Namespace
		}

		rows = append(rows, table.Row{
			resource.UUID,
			resource.Provider,
			alias,
			resource.ResourceType,
			namespace,
			resource.Name,
		})
		elems = append(elems, resource)
	}

	return columns, rows, elems
}

func (p *TableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *TableData) ElemKeyBindings(elem internal.Resource) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View resource detail", Command: fmt.Sprintf("resources/%s", elem.UUID)},
	}
}
