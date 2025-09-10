package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui"
	"github.com/marcbran/versource/internal/tui/platform"
)

type TableData struct {
	client          *client.Client
	moduleID        string
	moduleVersionID string
}

func NewTable(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.Component](&TableData{
			client:          client,
			moduleID:        params["module-id"],
			moduleVersionID: params["module-version-id"],
		})
	}
}

func (p *TableData) LoadData() ([]internal.Component, error) {
	ctx := context.Background()

	req := internal.ListComponentsRequest{}

	if p.moduleID != "" {
		if moduleID, err := strconv.ParseUint(p.moduleID, 10, 32); err == nil {
			moduleIDUint := uint(moduleID)
			req.ModuleID = &moduleIDUint
		}
	}

	if p.moduleVersionID != "" {
		if moduleVersionID, err := strconv.ParseUint(p.moduleVersionID, 10, 32); err == nil {
			moduleVersionIDUint := uint(moduleVersionID)
			req.ModuleVersionID = &moduleVersionIDUint
		}
	}

	resp, err := p.client.ListComponents(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Components, nil
}

func (p *TableData) ResolveData(data []internal.Component) ([]table.Column, []table.Row, []internal.Component) {

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 3},
		{Title: "Module", Width: 7},
		{Title: "Version", Width: 2},
	}

	var rows []table.Row
	var elems []internal.Component
	for _, component := range data {
		source := ""
		version := ""
		if component.ModuleVersion.Module.Source != "" {
			source = component.ModuleVersion.Module.Source
		}
		if component.ModuleVersion.Version != "" {
			version = component.ModuleVersion.Version
		}
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(component.ID), 10),
			component.Name,
			source,
			version,
		})
		elems = append(elems, component)
	}

	return columns, rows, elems
}

func (p *TableData) KeyBindings(elem internal.Component) platform.KeyBindings {
	return tui.KeyBindings.
		With("enter", "View component detail", fmt.Sprintf("components/%d", elem.ID))
}
