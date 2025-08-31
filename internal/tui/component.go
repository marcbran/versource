package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal"
)

func getComponentsTable(components []internal.Component) ([]table.Column, []table.Row, []string) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Module", Width: 7},
		{Title: "Version", Width: 2},
	}

	var rows []table.Row
	var ids []string
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
		ids = append(ids, strconv.FormatUint(uint64(component.ID), 10))
	}

	return columns, rows, ids
}

type ComponentsPage struct {
	app *App
}

func (p *ComponentsPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		req := internal.ListComponentsRequest{}

		if moduleIDStr, ok := params["module-id"]; ok {
			if moduleID, err := strconv.ParseUint(moduleIDStr, 10, 32); err == nil {
				moduleIDUint := uint(moduleID)
				req.ModuleID = &moduleIDUint
			}
		}

		if moduleVersionIDStr, ok := params["module-version-id"]; ok {
			if moduleVersionID, err := strconv.ParseUint(moduleVersionIDStr, 10, 32); err == nil {
				moduleVersionIDUint := uint(moduleVersionID)
				req.ModuleVersionID = &moduleVersionIDUint
			}
		}

		resp, err := p.app.client.ListComponents(ctx, req)
		if err != nil {
			return errorMsg{err: err}
		}

		view := "components"
		if len(params) > 0 {
			queryParts := make([]string, 0)
			if moduleIDStr, ok := params["module-id"]; ok {
				queryParts = append(queryParts, fmt.Sprintf("module-id=%s", moduleIDStr))
			}
			if moduleVersionIDStr, ok := params["module-version-id"]; ok {
				queryParts = append(queryParts, fmt.Sprintf("module-version-id=%s", moduleVersionIDStr))
			}
			if len(queryParts) > 0 {
				view = fmt.Sprintf("components?%s", strings.Join(queryParts, "&"))
			}
		}

		return dataLoadedMsg{view: view, data: resp.Components}
	}
}

func (p *ComponentsPage) Links(params map[string]string) map[string]string {
	return map[string]string{
		"m": "modules",
		"v": "moduleversions",
	}
}

type ChangesetComponentsPage struct {
	app *App
}

func (p *ChangesetComponentsPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		changesetName := params["changesetName"]

		req := internal.ListComponentsRequest{
			Changeset: &changesetName,
		}

		resp, err := p.app.client.ListComponents(ctx, req)
		if err != nil {
			return errorMsg{err: err}
		}

		view := fmt.Sprintf("changesets/%s/components", changesetName)
		return dataLoadedMsg{view: view, data: resp.Components}
	}
}

func (p *ChangesetComponentsPage) Links(params map[string]string) map[string]string {
	changesetName := params["changesetName"]
	return map[string]string{
		"p": fmt.Sprintf("changesets/%s/plans", changesetName),
		"a": fmt.Sprintf("changesets/%s/applies", changesetName),
	}
}
