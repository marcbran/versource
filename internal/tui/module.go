package tui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal"
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
	app *App
}

func (p *ModulesPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.app.client.ListModules(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "modules", data: resp.Modules}
	}
}

func (p *ModulesPage) Links(params map[string]string) map[string]string {
	return map[string]string{}
}

type ModulePage struct {
	app *App
}

func (p *ModulePage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		return dataLoadedMsg{view: fmt.Sprintf("modules/%s", params["moduleID"]), data: nil}
	}
}

func (p *ModulePage) Links(params map[string]string) map[string]string {
	return map[string]string{
		"enter": fmt.Sprintf("modules/%s/moduleversions", params["moduleID"]),
		"c":     fmt.Sprintf("components?module-id=%s", params["moduleID"]),
	}
}

type ModuleVersionsPage struct {
	app *App
}

func (p *ModuleVersionsPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.app.client.ListModuleVersions(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "moduleversions", data: resp.ModuleVersions}
	}
}

func (p *ModuleVersionsPage) Links(params map[string]string) map[string]string {
	return map[string]string{
		"c": fmt.Sprintf("components?module-version-id=%s", params["moduleVersionID"]),
	}
}

type ModuleVersionsForModulePage struct {
	app *App
}

func (p *ModuleVersionsForModulePage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		moduleID, exists := params["moduleID"]
		if !exists {
			return errorMsg{err: fmt.Errorf("moduleID parameter required")}
		}

		moduleIDUint, err := strconv.ParseUint(moduleID, 10, 32)
		if err != nil {
			return errorMsg{err: err}
		}
		resp, err := p.app.client.ListModuleVersionsForModule(ctx, uint(moduleIDUint))
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: fmt.Sprintf("modules/%s/moduleversions", moduleID), data: resp.ModuleVersions}
	}
}

func (p *ModuleVersionsForModulePage) Links(params map[string]string) map[string]string {
	return map[string]string{
		"m": "modules",
	}
}
