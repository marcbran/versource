package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DataTable struct {
	table   table.Model
	columns []table.Column
	rows    []table.Row
	elems   []any

	size    Rect
	data    TableData
	focused bool
}

func NewDataTable(data TableData) *DataTable {
	return &DataTable{
		table: table.New(),
		data:  data,
	}
}

func (t DataTable) Init() tea.Cmd {
	return t.data.LoadData()
}

func (t *DataTable) Resize(size Rect) {
	t.table.SetWidth(size.Width)
	t.table.SetHeight(size.Height)
	t.size = size
}

func (t *DataTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dataLoadedMsg:
		t.columns, t.rows, t.elems = t.data.ResolveData(msg.data)
		t.table = newTable(t.columns, t.rows, t.size)
		if t.focused {
			t.table.Focus()
		}
	}
	var cmd tea.Cmd
	t.table, cmd = t.table.Update(msg)
	return t, cmd
}

func (t DataTable) View() string {
	return t.table.View()
}

func (t DataTable) Links() map[string]string {
	return t.data.Links(t.elems[t.table.Cursor()])
}

func (t *DataTable) Focus() {
	t.focused = true
	t.table.Focus()
}

func (t *DataTable) Blur() {
	t.focused = false
	t.table.Blur()
}

type TableData interface {
	LoadData() tea.Cmd
	ResolveData(data any) ([]table.Column, []table.Row, []any)
	Links(elem any) map[string]string
}

func newTable(columns []table.Column, rows []table.Row, size Rect) table.Model {
	if len(rows) == 0 {
		placeholderRow := make(table.Row, len(columns))
		for i := range placeholderRow {
			placeholderRow[i] = ""
		}
		if len(columns) > 0 {
			placeholderRow[0] = "No data"
		}
		rows = append(rows, placeholderRow)
	}

	adjustedColumns := adjustColumnWidths(columns, size.Width)

	t := table.New(
		table.WithColumns(adjustedColumns),
		table.WithRows(rows),
		table.WithWidth(size.Width),
		table.WithHeight(size.Height),
	)
	t.SetStyles(table.Styles{
		Header:   lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("8")),
		Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("4")),
	})
	return t
}

func adjustColumnWidths(columns []table.Column, totalWidth int) []table.Column {
	if totalWidth <= 0 {
		return columns
	}

	totalWeight := 0
	for _, col := range columns {
		totalWeight += col.Width
	}

	if totalWeight == 0 {
		return columns
	}

	adjusted := make([]table.Column, len(columns))
	allocatedWidth := 0
	for i, col := range columns {
		adjusted[i] = col
		if totalWeight > 0 {
			adjusted[i].Width = max(1, (col.Width*totalWidth)/totalWeight)
		}
		allocatedWidth += adjusted[i].Width
	}

	if len(adjusted) > 0 && allocatedWidth < totalWidth {
		adjusted[len(adjusted)-1].Width += totalWidth - allocatedWidth
	}

	return adjusted
}
