package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type TableData struct {
	facade          versource.Facade
	moduleID        string
	moduleVersionID string
	changesetName   string
}

func NewTable(facade versource.Facade) func(params map[string]string) platform.Page {
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

func NewTableData(facade versource.Facade, moduleID, moduleVersionID, changesetName string) *TableData {
	return &TableData{
		facade:          facade,
		moduleID:        moduleID,
		moduleVersionID: moduleVersionID,
		changesetName:   changesetName,
	}
}

func (p *TableData) LoadData() ([]versource.Component, error) {
	ctx := context.Background()

	req := versource.ListComponentsRequest{}

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

func (p *TableData) ResolveData(data []versource.Component) ([]table.Column, []table.Row, []versource.Component) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 3},
		{Title: "Module", Width: 3},
		{Title: "Version", Width: 3},
		{Title: "Status", Width: 1},
	}

	var rows []table.Row
	var elems []versource.Component
	for _, component := range data {
		module := ""
		version := ""
		if component.ModuleVersion.Module.Name != "" {
			module = component.ModuleVersion.Module.Name
		}
		if component.ModuleVersion.Version != "" {
			version = component.ModuleVersion.Version
		}
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(component.ID), 10),
			component.Name,
			module,
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

func (p *TableData) ElemKeyBindings(elem versource.Component) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View component detail", Command: fmt.Sprintf("components/%d", elem.ID)},
		{Key: "E", Help: "Edit component", Command: fmt.Sprintf("components/%d/edit", elem.ID)},
		{Key: "D", Help: "Delete component", Command: fmt.Sprintf("components/%d/delete", elem.ID)},
	}
}
