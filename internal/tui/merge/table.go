package merge

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

func (p *TableData) LoadData() ([]versource.Merge, error) {
	ctx := context.Background()
	req := versource.ListMergesRequest{
		ChangesetName: p.changesetName,
	}
	resp, err := p.facade.ListMerges(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Merges, nil
}

func (p *TableData) ResolveData(data []versource.Merge) ([]table.Column, []table.Row, []versource.Merge) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
		{Title: "Merge Base", Width: 8},
		{Title: "Head", Width: 8},
	}

	var rows []table.Row
	var elems []versource.Merge
	for _, merge := range data {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(merge.ID), 10),
			merge.Changeset.Name,
			string(merge.State),
			merge.MergeBase,
			merge.Head,
		})
		elems = append(elems, merge)
	}

	return columns, rows, elems
}

func (p *TableData) KeyBindings() platform.KeyBindings {
	return platform.KeyBindings{}
}

func (p *TableData) ElemKeyBindings(elem versource.Merge) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "enter", Help: "View merge detail", Command: fmt.Sprintf("changesets/%s/merges/%d", p.changesetName, elem.ID)},
	}
}
