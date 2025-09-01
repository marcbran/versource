package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
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
	client          *client.Client
	moduleID        string
	moduleVersionID string
}

func NewComponentsPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return &ComponentsPage{
			client:          client,
			moduleID:        params["module-id"],
			moduleVersionID: params["module-version-id"],
		}
	}
}

func (p *ComponentsPage) Open() tea.Cmd {
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

		view := "components"
		if p.moduleID != "" || p.moduleVersionID != "" {
			queryParts := make([]string, 0)
			if p.moduleID != "" {
				queryParts = append(queryParts, fmt.Sprintf("module-id=%s", p.moduleID))
			}
			if p.moduleVersionID != "" {
				queryParts = append(queryParts, fmt.Sprintf("module-version-id=%s", p.moduleVersionID))
			}
			if len(queryParts) > 0 {
				view = fmt.Sprintf("components?%s", strings.Join(queryParts, "&"))
			}
		}

		return dataLoadedMsg{view: view, data: resp.Components}
	}
}

func (p *ComponentsPage) Links() map[string]string {
	return map[string]string{
		"m": "modules",
		"v": "moduleversions",
	}
}

type ChangesetComponentsPage struct {
	client        *client.Client
	changesetName string
}

func NewChangesetComponentsPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return &ChangesetComponentsPage{
			client:        client,
			changesetName: params["changesetName"],
		}
	}
}

func (p *ChangesetComponentsPage) Open() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		req := internal.ListComponentsRequest{
			Changeset: &p.changesetName,
		}

		resp, err := p.client.ListComponents(ctx, req)
		if err != nil {
			return errorMsg{err: err}
		}

		view := fmt.Sprintf("changesets/%s/components", p.changesetName)
		return dataLoadedMsg{view: view, data: resp.Components}
	}
}

func (p *ChangesetComponentsPage) Links() map[string]string {
	return map[string]string{
		"p": fmt.Sprintf("changesets/%s/plans", p.changesetName),
		"a": fmt.Sprintf("changesets/%s/applies", p.changesetName),
	}
}
