package tui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"gopkg.in/yaml.v3"
)

type ModulesTableData struct {
	client *client.Client
}

func NewModulesPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable(&ModulesTableData{client: client})
	}
}

func (p *ModulesTableData) LoadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.client.ListModules(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: resp.Modules}
	}
}

func (p *ModulesTableData) ResolveData(data any) ([]table.Column, []table.Row, []any) {
	modules, ok := data.([]internal.Module)
	if !ok {
		return nil, nil, nil
	}

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Source", Width: 9},
	}

	var rows []table.Row
	var elems []any
	for _, module := range modules {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(module.ID), 10),
			module.Source,
		})
		elems = append(elems, module)
	}

	return columns, rows, elems
}

func (p *ModulesTableData) KeyBindings(elem any) KeyBindings {
	if module, ok := elem.(internal.Module); ok {
		return rootKeyBindings.
			With("enter", "View module detail", fmt.Sprintf("modules/%d", module.ID)).
			With("v", "View module versions", fmt.Sprintf("modules/%d/moduleversions", module.ID)).
			With("c", "View components", fmt.Sprintf("components?module-id=%d", module.ID))
	}
	return rootKeyBindings
}

type ModuleVersionsTableData struct {
	client *client.Client
}

func NewModuleVersionsPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable(&ModuleVersionsTableData{client: client})
	}
}

func (p *ModuleVersionsTableData) LoadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.client.ListModuleVersions(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: resp.ModuleVersions}
	}
}

func (p *ModuleVersionsTableData) ResolveData(data any) ([]table.Column, []table.Row, []any) {
	moduleVersions, ok := data.([]internal.ModuleVersion)
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
		elems = append(elems, moduleVersion)
	}

	return columns, rows, elems
}

func (p *ModuleVersionsTableData) KeyBindings(elem any) KeyBindings {
	if moduleVersion, ok := elem.(internal.ModuleVersion); ok {
		return rootKeyBindings.
			With("enter", "View module version detail", fmt.Sprintf("moduleversions/%d", moduleVersion.ID)).
			With("c", "View components", fmt.Sprintf("components?module-version-id=%d", moduleVersion.ID))
	}
	return rootKeyBindings
}

type ModuleVersionsForModuleTableData struct {
	client   *client.Client
	moduleID string
}

func NewModuleVersionsForModulePage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable(&ModuleVersionsForModuleTableData{client: client, moduleID: params["moduleID"]})
	}
}

func (p *ModuleVersionsForModuleTableData) LoadData() tea.Cmd {
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
		return dataLoadedMsg{data: resp.ModuleVersions}
	}
}

func (p *ModuleVersionsForModuleTableData) ResolveData(data any) ([]table.Column, []table.Row, []any) {
	moduleVersions, ok := data.([]internal.ModuleVersion)
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
		elems = append(elems, moduleVersion)
	}

	return columns, rows, elems
}

func (p *ModuleVersionsForModuleTableData) KeyBindings(elem any) KeyBindings {
	if moduleVersion, ok := elem.(internal.ModuleVersion); ok {
		return rootKeyBindings.
			With("enter", "View module version detail", fmt.Sprintf("moduleversions/%d", moduleVersion.ID)).
			With("c", "View components", fmt.Sprintf("components?module-version-id=%d", moduleVersion.ID))
	}
	return rootKeyBindings
}

type ModuleDetailData struct {
	client   *client.Client
	moduleID string
}

type ModuleDetailViewModel struct {
	ID            uint   `yaml:"id"`
	Source        string `yaml:"source"`
	ExecutorType  string `yaml:"executorType"`
	LatestVersion *struct {
		ID      uint   `yaml:"id"`
		Version string `yaml:"version"`
	} `yaml:"latestVersion,omitempty"`
}

func NewModuleDetailPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataViewport(&ModuleDetailData{client: client, moduleID: params["moduleID"]})
	}
}

func (p *ModuleDetailData) LoadData() (*internal.GetModuleResponse, error) {
	ctx := context.Background()

	moduleIDUint, err := strconv.ParseUint(p.moduleID, 10, 32)
	if err != nil {
		return nil, err
	}

	moduleResp, err := p.client.GetModule(ctx, uint(moduleIDUint))
	if err != nil {
		return nil, err
	}

	return moduleResp, nil
}

func (p *ModuleDetailData) ResolveData(data internal.GetModuleResponse) string {
	var latestVersion *struct {
		ID      uint   `yaml:"id"`
		Version string `yaml:"version"`
	}
	if data.LatestVersion != nil {
		latestVersion = &struct {
			ID      uint   `yaml:"id"`
			Version string `yaml:"version"`
		}{
			ID:      data.LatestVersion.ID,
			Version: data.LatestVersion.Version,
		}
	}

	viewModel := ModuleDetailViewModel{
		ID:            data.Module.ID,
		Source:        data.Module.Source,
		ExecutorType:  data.Module.ExecutorType,
		LatestVersion: latestVersion,
	}

	yamlData, err := yaml.Marshal(viewModel)
	if err != nil {
		return fmt.Sprintf("Error marshaling to YAML: %v", err)
	}

	return string(yamlData)
}

func (p *ModuleDetailData) KeyBindings(elem internal.GetModuleResponse) KeyBindings {
	return rootKeyBindings.
		With("v", "View all versions", fmt.Sprintf("modules/%s/moduleversions", p.moduleID)).
		With("c", "View components", fmt.Sprintf("components?module-id=%s", p.moduleID))
}

type ModuleVersionDetailData struct {
	client          *client.Client
	moduleVersionID string
}

type ModuleVersionDetailViewModel struct {
	ID      uint   `yaml:"id"`
	Version string `yaml:"version"`
	Module  struct {
		ID           uint   `yaml:"id"`
		Source       string `yaml:"source"`
		ExecutorType string `yaml:"executorType"`
	} `yaml:"module"`
}

func NewModuleVersionDetailPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataViewport(&ModuleVersionDetailData{client: client, moduleVersionID: params["moduleVersionID"]})
	}
}

func (p *ModuleVersionDetailData) LoadData() (*internal.GetModuleVersionResponse, error) {
	ctx := context.Background()

	moduleVersionIDUint, err := strconv.ParseUint(p.moduleVersionID, 10, 32)
	if err != nil {
		return nil, err
	}

	moduleVersionResp, err := p.client.GetModuleVersion(ctx, uint(moduleVersionIDUint))
	if err != nil {
		return nil, err
	}

	return moduleVersionResp, nil
}

func (p *ModuleVersionDetailData) ResolveData(data internal.GetModuleVersionResponse) string {
	viewModel := ModuleVersionDetailViewModel{
		ID:      data.ModuleVersion.ID,
		Version: data.ModuleVersion.Version,
		Module: struct {
			ID           uint   `yaml:"id"`
			Source       string `yaml:"source"`
			ExecutorType string `yaml:"executorType"`
		}{
			ID:           data.ModuleVersion.Module.ID,
			Source:       data.ModuleVersion.Module.Source,
			ExecutorType: data.ModuleVersion.Module.ExecutorType,
		},
	}

	yamlData, err := yaml.Marshal(viewModel)
	if err != nil {
		return fmt.Sprintf("Error marshaling to YAML: %v", err)
	}

	return string(yamlData)
}

func (p *ModuleVersionDetailData) KeyBindings(elem internal.GetModuleVersionResponse) KeyBindings {
	moduleVersionIDUint, err := strconv.ParseUint(p.moduleVersionID, 10, 32)
	if err != nil {
		return rootKeyBindings
	}

	return rootKeyBindings.
		With("m", "View module", fmt.Sprintf("modules/%d", moduleVersionIDUint)).
		With("c", "View components", fmt.Sprintf("components?module-version-id=%d", moduleVersionIDUint))
}
