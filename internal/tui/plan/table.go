package plan

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type TableData struct {
	facade        versource.Facade
	changesetName string
}

func NewTable(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable(NewTableData(facade, params["changesetName"]))
	}
}

func NewTableData(facade versource.Facade, changesetName string) *TableData {
	return &TableData{
		facade:        facade,
		changesetName: changesetName,
	}
}

func (p *TableData) LoadData() ([]versource.Plan, error) {
	ctx := context.Background()
	req := versource.ListPlansRequest{
		ChangesetName: p.changesetName,
	}
	resp, err := p.facade.ListPlans(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Plans, nil
}

func (p *TableData) ResolveData(data []versource.Plan) ([]table.Column, []table.Row, []versource.Plan) {
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
	var elems []versource.Plan
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
			strconv.FormatUint(uint64(plan.ComponentID), 10),
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

func (p *TableData) ElemKeyBindings(elem versource.Plan) platform.KeyBindings {
	changesetPrefix := ""
	if p.changesetName != "" {
		changesetPrefix = fmt.Sprintf("changesets/%s", p.changesetName)
	}
	keyBindings := platform.KeyBindings{
		{Key: "enter", Help: "View plan detail", Command: fmt.Sprintf("%s/plans/%d", changesetPrefix, elem.ID)},
		{Key: "l", Help: "View logs", Command: fmt.Sprintf("%s/plans/%d/logs", changesetPrefix, elem.ID)},
	}

	if p.changesetName != "" {
		keyBindings = append(keyBindings, platform.KeyBinding{
			Key: "esc", Help: "View changesets", Command: "changesets",
		})
	}

	return keyBindings
}
