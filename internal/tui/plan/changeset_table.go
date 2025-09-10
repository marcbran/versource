package plan

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

type ChangesetTableData struct {
	client        *client.Client
	changesetName string
}

func NewChangesetTable(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataTable[internal.Plan](&ChangesetTableData{
			client:        client,
			changesetName: params["changesetName"],
		})
	}
}

func (p *ChangesetTableData) LoadData() ([]internal.Plan, error) {
	ctx := context.Background()

	resp, err := p.client.ListPlans(ctx, internal.ListPlansRequest{Changeset: &p.changesetName})
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

func (p *ChangesetTableData) KeyBindings(elem internal.Plan) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "l", Help: "View logs", Command: fmt.Sprintf("plans/%d/logs", elem.ID)},
	}
}
