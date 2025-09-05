package tui

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	log "github.com/sirupsen/logrus"
)

type PlansTableData struct {
	client *client.Client
}

func NewPlansPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable(&PlansTableData{client: client})
	}
}

func (p *PlansTableData) LoadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.client.ListPlans(ctx, internal.ListPlansRequest{})
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: resp.Plans}
	}
}

func (p *PlansTableData) ResolveData(data any) ([]table.Column, []table.Row, []any) {
	plans, ok := data.([]internal.Plan)
	if !ok {
		return nil, nil, nil
	}

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Component", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	var elems []any
	for _, plan := range plans {
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

func (p *PlansTableData) KeyBindings(elem any) KeyBindings {
	if plan, ok := elem.(internal.Plan); ok {
		return rootKeyBindings.
			With("l", "View logs", fmt.Sprintf("plans/%d/logs", plan.ID))
	}
	return rootKeyBindings
}

type ChangesetPlansTableData struct {
	client        *client.Client
	changesetName string
}

func NewChangesetPlansPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable(&ChangesetPlansTableData{
			client:        client,
			changesetName: params["changesetName"],
		})
	}
}

func (p *ChangesetPlansTableData) LoadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		resp, err := p.client.ListPlans(ctx, internal.ListPlansRequest{Changeset: &p.changesetName})
		if err != nil {
			return errorMsg{err: err}
		}

		return dataLoadedMsg{data: resp.Plans}
	}
}

func (p *ChangesetPlansTableData) ResolveData(data any) ([]table.Column, []table.Row, []any) {
	plans, ok := data.([]internal.Plan)
	if !ok {
		return nil, nil, nil
	}

	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Component", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	var elems []any
	for _, plan := range plans {
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

func (p *ChangesetPlansTableData) KeyBindings(elem any) KeyBindings {
	if plan, ok := elem.(internal.Plan); ok {
		return changesetKeyBindings(p.changesetName).
			With("l", "View logs", fmt.Sprintf("plans/%d/logs", plan.ID))
	}
	return changesetKeyBindings(p.changesetName)
}

type PlanLogsPageData struct {
	client *client.Client
	planID uint
}

func NewPlanLogsPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		planID, _ := strconv.ParseUint(params["planID"], 10, 32)
		return NewDataViewport(&PlanLogsPageData{
			client: client,
			planID: uint(planID),
		})
	}
}

func (p *PlanLogsPageData) LoadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.client.GetPlanLog(ctx, p.planID)
		log.WithField("resp", resp).Infof("resp: %+v", resp)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: resp}
	}
}

func (p *PlanLogsPageData) ResolveData(data any) string {
	logResp, ok := data.(*internal.GetPlanLogResponse)
	if !ok {
		return "Failed to load log data"
	}

	content, err := io.ReadAll(logResp.Content)
	if err != nil {
		return "Failed to read log content"
	}
	defer logResp.Content.Close()

	return string(content)
}

func (p *PlanLogsPageData) KeyBindings(elem any) KeyBindings {
	return rootKeyBindings
}
