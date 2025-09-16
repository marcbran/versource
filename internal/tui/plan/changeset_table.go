package plan

import (
	"context"
	"fmt"
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

func (p *ChangesetTableData) LoadData() ([]internal.Plan, error) {
	ctx := context.Background()

	resp, err := p.facade.ListPlans(ctx, internal.ListPlansRequest{Changeset: &p.changesetName})
	if err != nil {
		return nil, err
	}

	return resp.Plans, nil
}

func (p *ChangesetTableData) ResolveData(data []internal.Plan) ([]table.Column, []table.Row, []internal.Plan) {

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Component", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
		{Title: "Add", Width: 1},
		{Title: "Change", Width: 1},
		{Title: "Destroy", Width: 1},
	}

	var rows []table.Row
	var elems []internal.Plan
	for _, plan := range data {
		addStr := "-"
		if plan.Add != nil {
			addStr = strconv.Itoa(*plan.Add)
		}
		changeStr := "-"
		if plan.Change != nil {
			changeStr = strconv.Itoa(*plan.Change)
		}
		destroyStr := "-"
		if plan.Destroy != nil {
			destroyStr = strconv.Itoa(*plan.Destroy)
		}

		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(plan.ID), 10),
			plan.Component.Name,
			plan.Changeset.Name,
			string(plan.State),
			addStr,
			changeStr,
			destroyStr,
		})
		elems = append(elems, plan)
	}

	return columns, rows, elems
}

func (p *ChangesetTableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *ChangesetTableData) ElemKeyBindings(elem internal.Plan) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "l", Help: "View logs", Command: fmt.Sprintf("plans/%d/logs", elem.ID)},
	}
}
