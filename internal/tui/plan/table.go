package plan

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type TableData struct {
	facade        internal.Facade
	changesetName string
}

func NewTable(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable(NewTableData(facade, params["changesetName"]))
	}
}

func NewTableData(facade internal.Facade, changesetName string) *TableData {
	return &TableData{
		facade:        facade,
		changesetName: changesetName,
	}
}

func (p *TableData) LoadData() ([]internal.Plan, error) {
	ctx := context.Background()
	req := internal.ListPlansRequest{}
	if p.changesetName != "" {
		req.Changeset = &p.changesetName
	}
	resp, err := p.facade.ListPlans(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Plans, nil
}

func (p *TableData) ResolveData(data []internal.Plan) ([]table.Column, []table.Row, []internal.Plan) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Component", Width: 4},
		{Title: "Changeset", Width: 4},
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

func (p *TableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *TableData) ElemKeyBindings(elem internal.Plan) platform.KeyBindings {
	detailCommand := fmt.Sprintf("plans/%d", elem.ID)
	logsCommand := fmt.Sprintf("plans/%d/logs", elem.ID)

	if p.changesetName != "" {
		detailCommand = fmt.Sprintf("changesets/%s/plans/%d", p.changesetName, elem.ID)
		logsCommand = fmt.Sprintf("changesets/%s/plans/%d/logs", p.changesetName, elem.ID)
	}

	return platform.KeyBindings{
		{Key: "enter", Help: "View plan detail", Command: detailCommand},
		{Key: "l", Help: "View logs", Command: logsCommand},
	}
}
