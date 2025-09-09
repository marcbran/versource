package module

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui"
	"github.com/marcbran/versource/internal/tui/platform"
)

type VersionsTableData struct {
	client *client.Client
}

func NewVersionsTable(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.ModuleVersion](&VersionsTableData{client: client})
	}
}

func (p *VersionsTableData) LoadData() ([]internal.ModuleVersion, error) {
	ctx := context.Background()
	resp, err := p.client.ListModuleVersions(ctx)
	if err != nil {
		return nil, err
	}
	return resp.ModuleVersions, nil
}

func (p *VersionsTableData) ResolveData(data []internal.ModuleVersion) ([]table.Column, []table.Row, []internal.ModuleVersion) {

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

func (p *VersionsTableData) KeyBindings(elem internal.ModuleVersion) platform.KeyBindings {
	return tui.KeyBindings.
		With("enter", "View module version detail", fmt.Sprintf("moduleversions/%d", elem.ID)).
		With("c", "View components", fmt.Sprintf("components?module-version-id=%d", elem.ID))
}
