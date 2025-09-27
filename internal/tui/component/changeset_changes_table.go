package component

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type ChangesetChangesTableData struct {
	facade        internal.Facade
	changesetName string
}

func NewChangesetChangesTable(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable(NewChangesetChangesTableData(facade, params["changesetName"]))
	}
}

func NewChangesetChangesTableData(facade internal.Facade, changesetName string) *ChangesetChangesTableData {
	return &ChangesetChangesTableData{
		facade:        facade,
		changesetName: changesetName,
	}
}

func (p *ChangesetChangesTableData) LoadData() ([]internal.ComponentChange, error) {
	ctx := context.Background()
	req := internal.ListComponentChangesRequest{
		Changeset: p.changesetName,
	}
	resp, err := p.facade.ListComponentChanges(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Changes, nil
}

func (p *ChangesetChangesTableData) ResolveData(data []internal.ComponentChange) ([]table.Column, []table.Row, []internal.ComponentChange) {
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
	var elems []internal.ComponentChange
	for _, change := range data {
		toID := "N/A"
		if change.ToComponent != nil && change.ToComponent.ID != 0 {
			toID = strconv.FormatUint(uint64(change.ToComponent.ID), 10)
		}

		toName := "N/A"
		if change.ToComponent != nil {
			toName = change.ToComponent.Name
		}

		planState := "None"
		add := "?"
		changeCount := "?"
		destroy := "?"

		if change.Plan != nil {
			planState = string(change.Plan.State)
			if planState == "Failed" {
				add = "-"
				changeCount = "-"
				destroy = "-"
			} else {
				if change.Plan.Add != nil {
					add = strconv.Itoa(*change.Plan.Add)
				} else {
					add = "0"
				}
				if change.Plan.Change != nil {
					changeCount = strconv.Itoa(*change.Plan.Change)
				} else {
					changeCount = "0"
				}
				if change.Plan.Destroy != nil {
					destroy = strconv.Itoa(*change.Plan.Destroy)
				} else {
					destroy = "0"
				}
			}
		}

		rows = append(rows, table.Row{
			toID,
			toName,
			string(change.ChangeType),
			planState,
			add,
			changeCount,
			destroy,
		})
		elems = append(elems, change)
	}

	return columns, rows, elems
}

func (p *ChangesetChangesTableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *ChangesetChangesTableData) ElemKeyBindings(elem internal.ComponentChange) platform.KeyBindings {
	if elem.ToComponent == nil {
		return platform.KeyBindings{}
	}

	return platform.KeyBindings{
		{Key: "enter", Help: "View change detail", Command: fmt.Sprintf("changesets/%s/changes/%d", p.changesetName, elem.ToComponent.ID)},
	}
}
