package tui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
)

type ComponentsTableData struct {
	client          *client.Client
	moduleID        string
	moduleVersionID string
}

func NewComponentsPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable(&ComponentsTableData{
			client:          client,
			moduleID:        params["module-id"],
			moduleVersionID: params["module-version-id"],
		})
	}
}

func (p *ComponentsTableData) LoadData() tea.Cmd {
	return func() tea.Msg {
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
			return errorMsg{err: err}
		}

		return dataLoadedMsg{data: resp.Components}
	}
}

func (p *ComponentsTableData) ResolveData(data any) ([]table.Column, []table.Row, []any) {
	components, ok := data.([]internal.Component)
	if !ok {
		return nil, nil, nil
	}

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Module", Width: 7},
		{Title: "Version", Width: 2},
	}

	var rows []table.Row
	var elems []any
	for _, component := range components {
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
			source,
			version,
		})
		elems = append(elems, component)
	}

	return columns, rows, elems
}

func (p *ComponentsTableData) Links(elem any) map[string]string {
	return map[string]string{}
}

type ChangesetComponentsTableData struct {
	client        *client.Client
	changesetName string
}

func NewChangesetComponentsPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable(&ChangesetComponentsTableData{
			client:        client,
			changesetName: params["changesetName"],
		})
	}
}

func (p *ChangesetComponentsTableData) LoadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		req := internal.ListComponentsRequest{
			Changeset: &p.changesetName,
		}

		resp, err := p.client.ListComponents(ctx, req)
		if err != nil {
			return errorMsg{err: err}
		}

		return dataLoadedMsg{data: resp.Components}
	}
}

func (p *ChangesetComponentsTableData) ResolveData(data any) ([]table.Column, []table.Row, []any) {
	components, ok := data.([]internal.Component)
	if !ok {
		return nil, nil, nil
	}

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Module", Width: 7},
		{Title: "Version", Width: 2},
	}

	var rows []table.Row
	var elems []any
	for _, component := range components {
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
			source,
			version,
		})
		elems = append(elems, component)
	}

	return columns, rows, elems
}

func (p *ChangesetComponentsTableData) Links(elem any) map[string]string {
	return map[string]string{
		"d": fmt.Sprintf("changesets/%s/components/diffs", p.changesetName),
		"p": fmt.Sprintf("changesets/%s/plans", p.changesetName),
		"a": fmt.Sprintf("changesets/%s/applies", p.changesetName),
	}
}

type ComponentDiffsTableData struct {
	client    *client.Client
	changeset string
}

func NewComponentDiffsPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable(&ComponentDiffsTableData{
			client:    client,
			changeset: params["changesetName"],
		})
	}
}

func (p *ComponentDiffsTableData) LoadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		req := internal.ListComponentDiffsRequest{
			Changeset: p.changeset,
		}
		resp, err := p.client.ListComponentDiffs(ctx, req)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: resp.Diffs}
	}
}

func (p *ComponentDiffsTableData) ResolveData(data any) ([]table.Column, []table.Row, []any) {
	diffs, ok := data.([]internal.ComponentDiff)
	if !ok {
		return nil, nil, nil
	}

	columns := []table.Column{
		{Title: "Type", Width: 8},
		{Title: "From ID", Width: 8},
		{Title: "To ID", Width: 8},
		{Title: "From Module Version", Width: 15},
		{Title: "To Module Version", Width: 15},
		{Title: "From Variables", Width: 20},
		{Title: "To Variables", Width: 20},
	}

	var rows []table.Row
	var elems []any
	for _, diff := range diffs {
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
			fromModuleVersion,
			toModuleVersion,
			fromVariables,
			toVariables,
		})
		elems = append(elems, diff)
	}

	return columns, rows, elems
}

func (p *ComponentDiffsTableData) Links(elem any) map[string]string {
	return map[string]string{
		"b": fmt.Sprintf("changesets/%s/components", p.changeset),
		"c": fmt.Sprintf("changesets/%s/components", p.changeset),
		"p": fmt.Sprintf("changesets/%s/plans", p.changeset),
		"a": fmt.Sprintf("changesets/%s/applies", p.changeset),
	}
}
