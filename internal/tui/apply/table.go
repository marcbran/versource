package apply

import (
	"context"
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
		return platform.NewDataTable(&TableData{client: client})
	}
}

func (p *TableData) LoadData() ([]internal.Apply, error) {
	ctx := context.Background()
	resp, err := p.client.ListApplies(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Applies, nil
}

func (p *TableData) ResolveData(data []internal.Apply) ([]table.Column, []table.Row, []internal.Apply) {
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

func (p *TableData) KeyBindings(elem internal.Apply) platform.KeyBindings {
	return tui.KeyBindings
}
