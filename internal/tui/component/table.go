package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type TableData struct {
	facade          internal.Facade
	moduleID        string
	moduleVersionID string
	changesetName   string
}

func NewTable(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		var moduleId string
		if moduleIdParam, ok := params["module-id"]; ok {
			moduleId = moduleIdParam
		}
		var moduleVersionId string
		if moduleVersionIdParam, ok := params["module-version-id"]; ok {
			moduleVersionId = moduleVersionIdParam
		}
		var changesetName string
		if changesetNameParam, ok := params["changesetName"]; ok {
			changesetName = changesetNameParam
		}
		return platform.NewDataTable(NewTableData(facade, moduleId, moduleVersionId, changesetName))
	}
}

func NewTableData(facade internal.Facade, moduleID, moduleVersionID, changesetName string) *TableData {
	return &TableData{
		facade:          facade,
		moduleID:        moduleID,
		moduleVersionID: moduleVersionID,
		changesetName:   changesetName,
	}
}

func (p *TableData) LoadData() ([]internal.Component, error) {
	ctx := context.Background()

	req := internal.ListComponentsRequest{}

	if p.moduleID != "" {
		moduleID, err := strconv.ParseUint(p.moduleID, 10, 32)
		if err == nil {
			moduleIDUint := uint(moduleID)
			req.ModuleID = &moduleIDUint
		}
	}

	if p.moduleVersionID != "" {
		moduleVersionID, err := strconv.ParseUint(p.moduleVersionID, 10, 32)
		if err == nil {
			moduleVersionIDUint := uint(moduleVersionID)
			req.ModuleVersionID = &moduleVersionIDUint
		}
	}

	if p.changesetName != "" {
		req.ChangesetName = &p.changesetName
	}

	resp, err := p.facade.ListComponents(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Components, nil
}

func (p *TableData) ResolveData(data []internal.Component) ([]table.Column, []table.Row, []internal.Component) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 3},
		{Title: "Module", Width: 6},
		{Title: "Version", Width: 2},
		{Title: "Status", Width: 1},
	}

	var rows []table.Row
	var elems []internal.Component
	for _, component := range data {
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
			component.Name,
			source,
			version,
			string(component.Status),
		})
		elems = append(elems, component)
	}

	return columns, rows, elems
}

func (p *TableData) KeyBindings() platform.KeyBindings {
	command := "components/create"
	if p.moduleID != "" {
		command = fmt.Sprintf("components/create?module-id=%s", p.moduleID)
	}
	return platform.KeyBindings{
		{Key: "C", Help: "Create component", Command: command},
	}
}

func (p *TableData) ElemKeyBindings(elem internal.Component) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View component detail", Command: fmt.Sprintf("components/%d", elem.ID)},
		{Key: "E", Help: "Edit component", Command: fmt.Sprintf("components/%d/edit", elem.ID)},
		{Key: "D", Help: "Delete component", Command: fmt.Sprintf("components/%d/delete", elem.ID)},
	}
}
