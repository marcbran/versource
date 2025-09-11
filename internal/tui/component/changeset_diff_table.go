package component

import (
	"context"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

type DiffTableData struct {
	client    *client.Client
	changeset string
}

func NewChangesetDiffTable(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.ComponentDiff](&DiffTableData{
			client:    client,
			changeset: params["changesetName"],
		})
	}
}

func (p *DiffTableData) LoadData() ([]internal.ComponentDiff, error) {
	ctx := context.Background()
	req := internal.ListComponentDiffsRequest{
		Changeset: p.changeset,
	}
	resp, err := p.client.ListComponentDiffs(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Diffs, nil
}

func (p *DiffTableData) ResolveData(data []internal.ComponentDiff) ([]table.Column, []table.Row, []internal.ComponentDiff) {

	columns := []table.Column{
		{Title: "Type", Width: 8},
		{Title: "From ID", Width: 8},
		{Title: "To ID", Width: 8},
		{Title: "From Name", Width: 10},
		{Title: "To Name", Width: 10},
		{Title: "From Module Version", Width: 15},
		{Title: "To Module Version", Width: 15},
		{Title: "From Variables", Width: 20},
		{Title: "To Variables", Width: 20},
	}

	var rows []table.Row
	var elems []internal.ComponentDiff
	for _, diff := range data {
		fromID := "N/A"
		if diff.FromComponent.ID != 0 {
			fromID = strconv.FormatUint(uint64(diff.FromComponent.ID), 10)
		}

		toID := "N/A"
		if diff.ToComponent.ID != 0 {
			toID = strconv.FormatUint(uint64(diff.ToComponent.ID), 10)
		}

		fromModuleVersion := "N/A"
		if diff.FromComponent.ModuleVersionID != 0 {
			fromModuleVersion = strconv.FormatUint(uint64(diff.FromComponent.ModuleVersionID), 10)
		}

		toModuleVersion := "N/A"
		if diff.ToComponent.ModuleVersionID != 0 {
			toModuleVersion = strconv.FormatUint(uint64(diff.ToComponent.ModuleVersionID), 10)
		}

		fromVariables := "{}"
		if diff.FromComponent.Variables != nil {
			fromVariables = string(diff.FromComponent.Variables)
		}

		toVariables := "{}"
		if diff.ToComponent.Variables != nil {
			toVariables = string(diff.ToComponent.Variables)
		}

		rows = append(rows, table.Row{
			string(diff.DiffType),
			fromID,
			toID,
			diff.FromComponent.Name,
			diff.ToComponent.Name,
			fromModuleVersion,
			toModuleVersion,
			fromVariables,
			toVariables,
		})
		elems = append(elems, diff)
	}

	return columns, rows, elems
}

func (p *DiffTableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *DiffTableData) ElemKeyBindings(elem internal.ComponentDiff) platform.KeyBindings {
	return platform.KeyBindings{}
}
