package resource

import (
	"context"
	"fmt"

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

func (p *TableData) LoadData() ([]versource.Resource, error) {
	ctx := context.Background()
	resp, err := p.facade.ListResources(ctx, versource.ListResourcesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Resources, nil
}

func (p *TableData) ResolveData(data []versource.Resource) ([]table.Column, []table.Row, []versource.Resource) {
	columns := []table.Column{
		{Title: "UUID", Width: 8},
		{Title: "Provider", Width: 4},
		{Title: "Alias", Width: 4},
		{Title: "Type", Width: 6},
		{Title: "Namespace", Width: 4},
		{Title: "Name", Width: 8},
	}

	var rows []table.Row
	var elems []versource.Resource
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

func (p *TableData) ElemKeyBindings(elem versource.Resource) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View resource detail", Command: fmt.Sprintf("resources/%s", elem.UUID)},
	}
}
