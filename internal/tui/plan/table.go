package plan

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

type TableData struct {
	client *client.Client
}

func NewTable(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.Plan](&TableData{client: client})
	}
}

func (p *TableData) LoadData() ([]internal.Plan, error) {
	ctx := context.Background()
	resp, err := p.client.ListPlans(ctx, internal.ListPlansRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Plans, nil
}

func (p *TableData) ResolveData(data []internal.Plan) ([]table.Column, []table.Row, []internal.Plan) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Component", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	var elems []internal.Plan
	for _, plan := range data {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(plan.ID), 10),
			strconv.FormatUint(uint64(plan.ComponentID), 10),
			plan.Changeset.Name,
			plan.State,
		})
		elems = append(elems, plan)
	}

	return columns, rows, elems
}

func (p *TableData) KeyBindings(elem internal.Plan) platform.KeyBindings {
	return tui.KeyBindings.
		With("l", "View logs", fmt.Sprintf("plans/%d/logs", elem.ID))
}
