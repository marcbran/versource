package tui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
	"gopkg.in/yaml.v3"
)

type ModulesTableData struct {
	client *client.Client
}

func NewModulesPage(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.Module](&ModulesTableData{client: client})
	}
}

func (p *ModulesTableData) LoadData() ([]internal.Module, error) {
	ctx := context.Background()
	resp, err := p.client.ListModules(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Modules, nil
}

func (p *ModulesTableData) ResolveData(data []internal.Module) ([]table.Column, []table.Row, []internal.Module) {

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Source", Width: 9},
	}

	var rows []table.Row
	var elems []internal.Module
	for _, module := range data {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(module.ID), 10),
			module.Source,
		})
		elems = append(elems, module)
	}

	return columns, rows, elems
}

func (p *ModulesTableData) KeyBindings(elem internal.Module) platform.KeyBindings {
	return KeyBindings.
		With("enter", "View module detail", fmt.Sprintf("modules/%d", elem.ID)).
		With("v", "View module versions", fmt.Sprintf("modules/%d/moduleversions", elem.ID)).
		With("c", "View components", fmt.Sprintf("components?module-id=%d", elem.ID))
}

type ModuleVersionsTableData struct {
	client *client.Client
}

func NewModuleVersionsPage(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.ModuleVersion](&ModuleVersionsTableData{client: client})
	}
}

func (p *ModuleVersionsTableData) LoadData() ([]internal.ModuleVersion, error) {
	ctx := context.Background()
	resp, err := p.client.ListModuleVersions(ctx)
	if err != nil {
		return nil, err
	}
	return resp.ModuleVersions, nil
}

func (p *ModuleVersionsTableData) ResolveData(data []internal.ModuleVersion) ([]table.Column, []table.Row, []internal.ModuleVersion) {

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Module", Width: 7},
		{Title: "Version", Width: 2},
	}

	var rows []table.Row
	var elems []internal.ModuleVersion
	for _, moduleVersion := range data {
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

func (p *ModuleVersionsTableData) KeyBindings(elem internal.ModuleVersion) platform.KeyBindings {
	return KeyBindings.
		With("enter", "View module version detail", fmt.Sprintf("moduleversions/%d", elem.ID)).
		With("c", "View components", fmt.Sprintf("components?module-version-id=%d", elem.ID))
}

type ModuleVersionsForModuleTableData struct {
	client   *client.Client
	moduleID string
}

func NewModuleVersionsForModulePage(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.ModuleVersion](&ModuleVersionsForModuleTableData{client: client, moduleID: params["moduleID"]})
	}
}

func (p *ModuleVersionsForModuleTableData) LoadData() ([]internal.ModuleVersion, error) {
	ctx := context.Background()

	moduleIDUint, err := strconv.ParseUint(p.moduleID, 10, 32)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.ListModuleVersionsForModule(ctx, uint(moduleIDUint))
	if err != nil {
		return nil, err
	}
	return resp.ModuleVersions, nil
}

func (p *ModuleVersionsForModuleTableData) ResolveData(data []internal.ModuleVersion) ([]table.Column, []table.Row, []internal.ModuleVersion) {

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Module", Width: 7},
		{Title: "Version", Width: 2},
	}

	var rows []table.Row
	var elems []internal.ModuleVersion
	for _, moduleVersion := range data {
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

func (p *ModuleVersionsForModuleTableData) KeyBindings(elem internal.ModuleVersion) platform.KeyBindings {
	return KeyBindings.
		With("enter", "View module version detail", fmt.Sprintf("moduleversions/%d", elem.ID)).
		With("c", "View components", fmt.Sprintf("components?module-version-id=%d", elem.ID))
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

func NewModuleDetailPage(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataViewport(&ModuleDetailData{client: client, moduleID: params["moduleID"]})
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

func (p *ModuleDetailData) KeyBindings(elem internal.GetModuleResponse) platform.KeyBindings {
	return KeyBindings.
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

func NewModuleVersionDetailPage(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataViewport(&ModuleVersionDetailData{client: client, moduleVersionID: params["moduleVersionID"]})
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

func (p *ModuleVersionDetailData) KeyBindings(elem internal.GetModuleVersionResponse) platform.KeyBindings {
	moduleVersionIDUint, err := strconv.ParseUint(p.moduleVersionID, 10, 32)
	if err != nil {
		return KeyBindings
	}

	return KeyBindings.
		With("m", "View module", fmt.Sprintf("modules/%d", moduleVersionIDUint)).
		With("c", "View components", fmt.Sprintf("components?module-version-id=%d", moduleVersionIDUint))
}
