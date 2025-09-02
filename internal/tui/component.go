package tui

import (
	"context"
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
	return map[string]string{}
}
