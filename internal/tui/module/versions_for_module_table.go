package module

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

type VersionsForModuleTableData struct {
	client   *client.Client
	moduleID string
}

func NewVersionsForModuleTable(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.ModuleVersion](&VersionsForModuleTableData{client: client, moduleID: params["moduleID"]})
	}
}

func (p *VersionsForModuleTableData) LoadData() ([]internal.ModuleVersion, error) {
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

func (p *VersionsForModuleTableData) ResolveData(data []internal.ModuleVersion) ([]table.Column, []table.Row, []internal.ModuleVersion) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Module", Width: 7},
		{Title: "Version", Width: 2},
	}

	var rows []table.Row
	var elems []internal.ModuleVersion
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

func (p *VersionsForModuleTableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *VersionsForModuleTableData) ElemKeyBindings(elem internal.ModuleVersion) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View module version detail", Command: fmt.Sprintf("moduleversions/%d", elem.ID)},
		{Key: "c", Help: "View components", Command: fmt.Sprintf("components?module-version-id=%d", elem.ID)},
	}
}
