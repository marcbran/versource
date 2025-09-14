package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type ChangesetDiffsTableData struct {
	facade        internal.Facade
	changesetName string
}

func NewChangesetDiffsTable(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable(NewChangesetDiffsTableData(facade, params["changesetName"]))
	}
}

func NewChangesetDiffsTableData(facade internal.Facade, changesetName string) *ChangesetDiffsTableData {
	return &ChangesetDiffsTableData{
		facade:        facade,
		changesetName: changesetName,
	}
}

func (p *ChangesetDiffsTableData) LoadData() ([]internal.ComponentDiff, error) {
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

func (p *ChangesetDiffsTableData) ResolveData(data []internal.ComponentDiff) ([]table.Column, []table.Row, []internal.ComponentDiff) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 10},
		{Title: "Change Type", Width: 8},
		{Title: "Plan", Width: 8},
		{Title: "Add", Width: 3},
		{Title: "Change", Width: 5},
		{Title: "Destroy", Width: 6},
	}

	var rows []table.Row
	var elems []internal.ComponentDiff
	for _, diff := range data {
		toID := "N/A"
		if diff.ToComponent != nil && diff.ToComponent.ID != 0 {
			toID = strconv.FormatUint(uint64(diff.ToComponent.ID), 10)
		}

		toName := "N/A"
		if diff.ToComponent != nil {
			toName = diff.ToComponent.Name
		}

		planState := "None"
		add := "?"
		change := "?"
		destroy := "?"

		if diff.Plan != nil {
			planState = diff.Plan.State
			if planState == "Failed" {
				add = "-"
				change = "-"
				destroy = "-"
			} else {
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
		}

		rows = append(rows, table.Row{
			toID,
			toName,
			string(diff.DiffType),
			planState,
			add,
			change,
			destroy,
		})
		elems = append(elems, diff)
	}

	return columns, rows, elems
}

func (p *ChangesetDiffsTableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *ChangesetDiffsTableData) ElemKeyBindings(elem internal.ComponentDiff) platform.KeyBindings {
	if elem.ToComponent == nil {
		return platform.KeyBindings{}
	}

	return platform.KeyBindings{
		{Key: "enter", Help: "View diff detail", Command: fmt.Sprintf("changesets/%s/diffs/%d", p.changesetName, elem.ToComponent.ID)},
	}
}
