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

func getModulesTable(modules []internal.Module) ([]table.Column, []table.Row, []string) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Source", Width: 9},
	}

	var rows []table.Row
	var ids []string
	for _, module := range modules {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(module.ID), 10),
			module.Source,
		})
		ids = append(ids, strconv.FormatUint(uint64(module.ID), 10))
	}

	return columns, rows, ids
}

func getModuleVersionsTable(moduleVersions []internal.ModuleVersion) ([]table.Column, []table.Row, []string) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Module", Width: 7},
		{Title: "Version", Width: 2},
	}

	var rows []table.Row
	var ids []string
	for _, moduleVersion := range moduleVersions {
		source := ""
		if moduleVersion.Module.Source != "" {
			source = moduleVersion.Module.Source
		}
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(moduleVersion.ID), 10),
			source,
			moduleVersion.Version,
		})
		ids = append(ids, strconv.FormatUint(uint64(moduleVersion.ID), 10))
	}

	return columns, rows, ids
}

type ModulesPage struct {
	client *client.Client
}

func NewModulesPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return &ModulesPage{client: client}
	}
}

func (p *ModulesPage) Open() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.client.ListModules(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "modules", data: resp.Modules}
	}
}

func (p *ModulesPage) Links() map[string]string {
	return map[string]string{}
}

type ModulePage struct {
	client   *client.Client
	moduleID string
}

func NewModulePage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return &ModulePage{client: client, moduleID: params["moduleID"]}
	}
}

func (p *ModulePage) Open() tea.Cmd {
	return func() tea.Msg {
		return dataLoadedMsg{view: fmt.Sprintf("modules/%s", p.moduleID), data: nil}
	}
}

func (p *ModulePage) Links() map[string]string {
	return map[string]string{
		"enter": fmt.Sprintf("modules/%s/moduleversions", p.moduleID),
		"c":     fmt.Sprintf("components?module-id=%s", p.moduleID),
	}
}

type ModuleVersionsPage struct {
	client          *client.Client
	moduleVersionID string
}

func NewModuleVersionsPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return &ModuleVersionsPage{client: client, moduleVersionID: params["moduleVersionID"]}
	}
}

func (p *ModuleVersionsPage) Open() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.client.ListModuleVersions(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "moduleversions", data: resp.ModuleVersions}
	}
}

func (p *ModuleVersionsPage) Links() map[string]string {
	return map[string]string{
		"c": fmt.Sprintf("components?module-version-id=%s", p.moduleVersionID),
	}
}

type ModuleVersionsForModulePage struct {
	client   *client.Client
	moduleID string
}

func NewModuleVersionsForModulePage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return &ModuleVersionsForModulePage{client: client, moduleID: params["moduleID"]}
	}
}

func (p *ModuleVersionsForModulePage) Open() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		moduleIDUint, err := strconv.ParseUint(p.moduleID, 10, 32)
		if err != nil {
			return errorMsg{err: err}
		}
		resp, err := p.client.ListModuleVersionsForModule(ctx, uint(moduleIDUint))
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: fmt.Sprintf("modules/%s/moduleversions", p.moduleID), data: resp.ModuleVersions}
	}
}

func (p *ModuleVersionsForModulePage) Links() map[string]string {
	return map[string]string{
		"m": "modules",
	}
}
