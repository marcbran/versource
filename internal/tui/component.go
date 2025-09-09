package tui

import (
	"context"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

type ComponentsTableData struct {
	client          *client.Client
	moduleID        string
	moduleVersionID string
}

func NewComponentsPage(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.Component](&ComponentsTableData{
			client:          client,
			moduleID:        params["module-id"],
			moduleVersionID: params["module-version-id"],
		})
	}
}

func (p *ComponentsTableData) LoadData() ([]internal.Component, error) {
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

func (p *ComponentsTableData) ResolveData(data []internal.Component) ([]table.Column, []table.Row, []internal.Component) {

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

func (p *ComponentsTableData) KeyBindings(elem internal.Component) platform.KeyBindings {
	return KeyBindings
}

type ChangesetComponentsTableData struct {
	client        *client.Client
	changesetName string
}

func NewChangesetComponentsPage(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.Component](&ChangesetComponentsTableData{
			client:        client,
			changesetName: params["changesetName"],
		})
	}
}

func (p *ChangesetComponentsTableData) LoadData() ([]internal.Component, error) {
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

func (p *ChangesetComponentsTableData) ResolveData(data []internal.Component) ([]table.Column, []table.Row, []internal.Component) {

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

func (p *ChangesetComponentsTableData) KeyBindings(elem internal.Component) platform.KeyBindings {
	return changesetKeyBindings(p.changesetName)
}

type ComponentDiffsTableData struct {
	client    *client.Client
	changeset string
}

func NewComponentDiffsPage(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.ComponentDiff](&ComponentDiffsTableData{
			client:    client,
			changeset: params["changesetName"],
		})
	}
}

func (p *ComponentDiffsTableData) LoadData() ([]internal.ComponentDiff, error) {
	ctx := context.Background()
	req := internal.ListComponentDiffsRequest{
		Changeset: p.changeset,
	}
	resp, err := p.client.ListComponentDiffs(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Diffs, nil
}

func (p *ComponentDiffsTableData) ResolveData(data []internal.ComponentDiff) ([]table.Column, []table.Row, []internal.ComponentDiff) {

	columns := []table.Column{
		{Title: "Type", Width: 8},
		{Title: "From ID", Width: 8},
		{Title: "To ID", Width: 8},
		{Title: "From Name", Width: 10},
		{Title: "To Name", Width: 10},
		{Title: "From Module Version", Width: 15},
		{Title: "To Module Version", Width: 15},
		{Title: "From Variables", Width: 20},
		{Title: "To Variables", Width: 20},
	}

	var rows []table.Row
	var elems []internal.ComponentDiff
	for _, diff := range data {
		fromID := "N/A"
		if diff.FromComponent.ID != 0 {
			fromID = strconv.FormatUint(uint64(diff.FromComponent.ID), 10)
		}

		toID := "N/A"
		if diff.ToComponent.ID != 0 {
			toID = strconv.FormatUint(uint64(diff.ToComponent.ID), 10)
		}

		fromModuleVersion := "N/A"
		if diff.FromComponent.ModuleVersionID != 0 {
			fromModuleVersion = strconv.FormatUint(uint64(diff.FromComponent.ModuleVersionID), 10)
		}

		toModuleVersion := "N/A"
		if diff.ToComponent.ModuleVersionID != 0 {
			toModuleVersion = strconv.FormatUint(uint64(diff.ToComponent.ModuleVersionID), 10)
		}

		fromVariables := "{}"
		if diff.FromComponent.Variables != nil {
			fromVariables = string(diff.FromComponent.Variables)
		}

		toVariables := "{}"
		if diff.ToComponent.Variables != nil {
			toVariables = string(diff.ToComponent.Variables)
		}

		rows = append(rows, table.Row{
			string(diff.DiffType),
			fromID,
			toID,
			diff.FromComponent.Name,
			diff.ToComponent.Name,
			fromModuleVersion,
			toModuleVersion,
			fromVariables,
			toVariables,
		})
		elems = append(elems, diff)
	}

	return columns, rows, elems
}

func (p *ComponentDiffsTableData) KeyBindings(elem internal.ComponentDiff) platform.KeyBindings {
	return changesetKeyBindings(p.changeset)
}
