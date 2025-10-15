package module

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type VersionsTableData struct {
	facade   versource.Facade
	moduleID *string
}

func NewVersionsTable(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		var moduleID *string
		if moduleIDStr, exists := params["moduleID"]; exists && moduleIDStr != "" {
			moduleID = &moduleIDStr
		}
		return platform.NewDataTable(NewVersionsTableData(facade, moduleID))
	}
}

func NewVersionsTableData(facade versource.Facade, moduleID *string) *VersionsTableData {
	return &VersionsTableData{facade: facade, moduleID: moduleID}
}

func (p *VersionsTableData) LoadData() ([]versource.ModuleVersion, error) {
	ctx := context.Background()

	req := versource.ListModuleVersionsRequest{}
	if p.moduleID != nil {
		moduleIDUint, err := strconv.ParseUint(*p.moduleID, 10, 32)
		if err != nil {
			return nil, err
		}
		moduleID := uint(moduleIDUint)
		req.ModuleID = &moduleID
	}

	resp, err := p.facade.ListModuleVersions(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.ModuleVersions, nil
}

func (p *VersionsTableData) ResolveData(data []versource.ModuleVersion) ([]table.Column, []table.Row, []versource.ModuleVersion) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Module", Width: 7},
		{Title: "Version", Width: 2},
	}

	var rows []table.Row
	var elems []versource.ModuleVersion
	for _, moduleVersion := range data {
		name := ""
		if moduleVersion.Module.Name != "" {
			name = moduleVersion.Module.Name
		}
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(moduleVersion.ID), 10),
			name,
			moduleVersion.Version,
		})
		elems = append(elems, moduleVersion)
	}

	return columns, rows, elems
}

func (p *VersionsTableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *VersionsTableData) ElemKeyBindings(elem versource.ModuleVersion) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View module version detail", Command: fmt.Sprintf("moduleversions/%d", elem.ID)},
		{Key: "c", Help: "View components", Command: fmt.Sprintf("components?module-version-id=%d", elem.ID)},
	}
}
