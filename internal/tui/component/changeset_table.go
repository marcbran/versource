package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

type ChangesetTableData struct {
	client        *client.Client
	changesetName string
}

func NewChangesetTable(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.Component](&ChangesetTableData{
			client:        client,
			changesetName: params["changesetName"],
		})
	}
}

func (p *ChangesetTableData) LoadData() ([]internal.Component, error) {
	ctx := context.Background()

	req := internal.ListComponentsRequest{
		Changeset: &p.changesetName,
	}

	resp, err := p.client.ListComponents(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Components, nil
}

func (p *ChangesetTableData) ResolveData(data []internal.Component) ([]table.Column, []table.Row, []internal.Component) {

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

func (p *ChangesetTableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *ChangesetTableData) ElemKeyBindings(elem internal.Component) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "P", Help: "Create plan for component", Command: fmt.Sprintf("changesets/%s/components/%d/plans/create", p.changesetName, elem.ID)},
		{Key: "D", Help: "Delete component", Command: fmt.Sprintf("changesets/%s/components/%d/delete", p.changesetName, elem.ID)},
		{Key: "R", Help: "Restore component", Command: fmt.Sprintf("changesets/%s/components/%d/restore", p.changesetName, elem.ID)},
	}
}
