package tui

import (
	"context"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
)

type AppliesTableData struct {
	client *client.Client
}

func NewAppliesPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable(&AppliesTableData{client: client})
	}
}

func (p *AppliesTableData) LoadData() ([]internal.Apply, error) {
	ctx := context.Background()
	resp, err := p.client.ListApplies(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Applies, nil
}

func (p *AppliesTableData) ResolveData(data []internal.Apply) ([]table.Column, []table.Row, []internal.Apply) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Plan", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	var elems []internal.Apply
	for _, apply := range data {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(apply.ID), 10),
			strconv.FormatUint(uint64(apply.PlanID), 10),
			apply.Changeset.Name,
			apply.State,
		})
		elems = append(elems, apply)
	}

	return columns, rows, elems
}

func (p *AppliesTableData) KeyBindings(elem internal.Apply) KeyBindings {
	return rootKeyBindings
}

type ChangesetAppliesTableData struct {
	client        *client.Client
	changesetName string
}

func NewChangesetAppliesPage(client *client.Client) func(params map[string]string) Page {
	return func(params map[string]string) Page {
		return NewDataTable(&ChangesetAppliesTableData{
			client:        client,
			changesetName: params["changesetName"],
		})
	}
}

func (p *ChangesetAppliesTableData) LoadData() ([]internal.Apply, error) {
	ctx := context.Background()

	resp, err := p.client.ListApplies(ctx)
	if err != nil {
		return nil, err
	}

	return resp.Applies, nil
}

func (p *ChangesetAppliesTableData) ResolveData(data []internal.Apply) ([]table.Column, []table.Row, []internal.Apply) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Plan", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	var elems []internal.Apply
	for _, apply := range data {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(apply.ID), 10),
			strconv.FormatUint(uint64(apply.PlanID), 10),
			apply.Changeset.Name,
			apply.State,
		})
		elems = append(elems, apply)
	}

	return columns, rows, elems
}

func (p *ChangesetAppliesTableData) KeyBindings(elem internal.Apply) KeyBindings {
	return changesetKeyBindings(p.changesetName)
}
