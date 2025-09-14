package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type ChangesetTableData struct {
	facade        internal.Facade
	changesetName string
}

func NewChangesetTable(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable(NewChangesetTableData(facade, params["changesetName"]))
	}
}

func NewChangesetTableData(facade internal.Facade, changesetName string) *ChangesetTableData {
	return &ChangesetTableData{
		facade:        facade,
		changesetName: changesetName,
	}
}

func (p *ChangesetTableData) LoadData() ([]internal.Component, error) {
	ctx := context.Background()
	req := internal.ListComponentsRequest{
		Changeset: &p.changesetName,
	}
	resp, err := p.facade.ListComponents(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Components, nil
}

func (p *ChangesetTableData) ResolveData(data []internal.Component) ([]table.Column, []table.Row, []internal.Component) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 3},
		{Title: "Module", Width: 6},
		{Title: "Version", Width: 2},
		{Title: "Status", Width: 1},
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
			string(component.Status),
		})
		elems = append(elems, component)
	}

	return columns, rows, elems
}

func (p *ChangesetTableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "C", Help: "Create component", Command: fmt.Sprintf("changesets/%s/components/create", p.changesetName)},
	}
}

func (p *ChangesetTableData) ElemKeyBindings(elem internal.Component) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View component detail", Command: fmt.Sprintf("changesets/%s/components/%d", p.changesetName, elem.ID)},
		{Key: "E", Help: "Edit component", Command: fmt.Sprintf("changesets/%s/components/%d/edit", p.changesetName, elem.ID)},
		{Key: "D", Help: "Delete component", Command: fmt.Sprintf("changesets/%s/components/%d/delete", p.changesetName, elem.ID)},
		{Key: "S", Help: "Restore component", Command: fmt.Sprintf("changesets/%s/components/%d/restore", p.changesetName, elem.ID)},
		{Key: "P", Help: "Create plan", Command: fmt.Sprintf("changesets/%s/components/%d/plans/create", p.changesetName, elem.ID)},
	}
}
