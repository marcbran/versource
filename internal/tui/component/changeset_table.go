package component

import (
	"context"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type ChangesetTableData struct {
	facade        internal.Facade
	changesetName string
}

func NewChangesetTable(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable(NewChangesetTableData(facade, params["changesetName"]))
	}
}

func NewChangesetTableData(facade internal.Facade, changesetName string) *ChangesetTableData {
	return &ChangesetTableData{
		facade:        facade,
		changesetName: changesetName,
	}
}

func (p *ChangesetTableData) LoadData() ([]internal.ComponentDiff, error) {
	ctx := context.Background()
	req := internal.ListComponentDiffsRequest{
		Changeset: p.changesetName,
	}
	resp, err := p.facade.ListComponentDiffs(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Diffs, nil
}

func (p *ChangesetTableData) ResolveData(data []internal.ComponentDiff) ([]table.Column, []table.Row, []internal.ComponentDiff) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 10},
		{Title: "Type", Width: 8},
		{Title: "Add", Width: 3},
		{Title: "Change", Width: 5},
		{Title: "Destroy", Width: 6},
	}

	var rows []table.Row
	var elems []internal.ComponentDiff
	for _, diff := range data {
		toID := "N/A"
		if diff.ToComponent.ID != 0 {
			toID = strconv.FormatUint(uint64(diff.ToComponent.ID), 10)
		}

		add := "?"
		change := "?"
		destroy := "?"

		if diff.Plan != nil {
			if diff.Plan.Add != nil {
				add = strconv.Itoa(*diff.Plan.Add)
			} else {
				add = "0"
			}
			if diff.Plan.Change != nil {
				change = strconv.Itoa(*diff.Plan.Change)
			} else {
				change = "0"
			}
			if diff.Plan.Destroy != nil {
				destroy = strconv.Itoa(*diff.Plan.Destroy)
			} else {
				destroy = "0"
			}
		}

		rows = append(rows, table.Row{
			toID,
			diff.ToComponent.Name,
			string(diff.DiffType),
			add,
			change,
			destroy,
		})
		elems = append(elems, diff)
	}

	return columns, rows, elems
}

func (p *ChangesetTableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *ChangesetTableData) ElemKeyBindings(elem internal.ComponentDiff) platform.KeyBindings {
	return platform.KeyBindings{}
}
