package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DataTable[T any] struct {
	table   table.Model
	columns []table.Column
	rows    []table.Row
	elems   []T

	size    Size
	data    TableData[T]
	focused bool
}

func NewDataTable[T any](data TableData[T]) *DataTable[T] {
	return &DataTable[T]{
		table: table.New(),
		data:  data,
	}
}

func (t DataTable[T]) Init() tea.Cmd {
	return func() tea.Msg {
		data, err := t.data.LoadData()
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: data}
	}
}

func (t *DataTable[T]) Resize(size Size) {
	t.table.SetWidth(size.Width)
	t.table.SetHeight(size.Height)
	t.size = size
}

func (t *DataTable[T]) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case dataLoadedMsg:
		if data, ok := msg.data.([]T); ok {
			t.columns, t.rows, t.elems = t.data.ResolveData(data)
			t.table = newTable(t.columns, t.rows, t.size)
			if t.focused {
				t.table.Focus()
			}
		}
	}
	var cmd tea.Cmd
	t.table, cmd = t.table.Update(msg)
	return t, cmd
}

func (t DataTable[T]) View() string {
	return t.table.View()
}

func (t DataTable[T]) KeyBindings() KeyBindings {
	cursor := t.table.Cursor()
	if cursor < 0 || cursor >= len(t.rows) {
		return KeyBindings{}
	}
	return t.data.KeyBindings(t.elems[cursor])
}

func (t *DataTable[T]) Focus() {
	t.focused = true
	t.table.Focus()
}

func (t *DataTable[T]) Blur() {
	t.focused = false
	t.table.Blur()
}

type TableData[T any] interface {
	LoadData() ([]T, error)
	ResolveData(data []T) ([]table.Column, []table.Row, []T)
	KeyBindings(elem T) KeyBindings
}

func newTable(columns []table.Column, rows []table.Row, size Size) table.Model {
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
