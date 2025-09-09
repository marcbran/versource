package tui

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
)

type PlansTableData struct {
	client *client.Client
}

func NewPlansPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable[internal.Plan](&PlansTableData{client: client})
	}
}

func (p *PlansTableData) LoadData() ([]internal.Plan, error) {
	ctx := context.Background()
	resp, err := p.client.ListPlans(ctx, internal.ListPlansRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Plans, nil
}

func (p *PlansTableData) ResolveData(data []internal.Plan) ([]table.Column, []table.Row, []internal.Plan) {

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

func (p *PlansTableData) KeyBindings(elem internal.Plan) KeyBindings {
	return rootKeyBindings.
		With("l", "View logs", fmt.Sprintf("plans/%d/logs", elem.ID))
}

type ChangesetPlansTableData struct {
	client        *client.Client
	changesetName string
}

func NewChangesetPlansPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable[internal.Plan](&ChangesetPlansTableData{
			client:        client,
			changesetName: params["changesetName"],
		})
	}
}

func (p *ChangesetPlansTableData) LoadData() ([]internal.Plan, error) {
	ctx := context.Background()

	resp, err := p.client.ListPlans(ctx, internal.ListPlansRequest{Changeset: &p.changesetName})
	if err != nil {
		return nil, err
	}

	return resp.Plans, nil
}

func (p *ChangesetPlansTableData) ResolveData(data []internal.Plan) ([]table.Column, []table.Row, []internal.Plan) {

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

func (p *ChangesetPlansTableData) KeyBindings(elem internal.Plan) KeyBindings {
	return changesetKeyBindings(p.changesetName).
		With("l", "View logs", fmt.Sprintf("plans/%d/logs", elem.ID))
}

type PlanLogsPageData struct {
	client *client.Client
	planID string
}

func NewPlanLogsPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataViewport(&PlanLogsPageData{client: client, planID: params["planID"]})
	}
}

func (p *PlanLogsPageData) LoadData() (*internal.GetPlanLogResponse, error) {
	ctx := context.Background()

	planIDUint, err := strconv.ParseUint(p.planID, 10, 32)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.GetPlanLog(ctx, uint(planIDUint))
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *PlanLogsPageData) ResolveData(data internal.GetPlanLogResponse) string {
	content, err := io.ReadAll(data.Content)
	if err != nil {
		return "Failed to read log content"
	}
	defer data.Content.Close()

	return string(content)
}

func (p *PlanLogsPageData) KeyBindings(elem internal.GetPlanLogResponse) KeyBindings {
	return rootKeyBindings
}
