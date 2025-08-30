package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http"
)

type App struct {
	client      *http.Client
	currentView string
	table       table.Model
	columns     []table.Column
	rows        []table.Row
	size        Rect
	loading     bool
	err         error
	input       textinput.Model
	showInput   bool
}

type Rect struct {
	Width  int
	Height int
}

func NewApp(client *http.Client) *App {
	ti := textinput.New()
	ti.Placeholder = "Enter command..."
	ti.CharLimit = 100

	return &App{
		client:      client,
		currentView: "modules",
		table:       table.New(),
		input:       ti,
	}
}

func (a *App) Init() tea.Cmd {
	return a.loadData()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if a.showInput {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "ctrl+c":
				return a, tea.Quit
			case "esc":
				a.showInput = false
				a.input.SetValue("")
				a.table = createTable(a.columns, a.rows, a.size, a.showInput)
				return a, nil
			case "enter":
				command := a.input.Value()
				a.showInput = false
				a.input.SetValue("")
				a.table = createTable(a.columns, a.rows, a.size, a.showInput)
				return a, a.executeCommand(command)
			}
		}
		a.input, cmd = a.input.Update(msg)
		return a, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.size.Width = msg.Width
		a.size.Height = msg.Height
		a.input.Width = msg.Width - 7
		a.table = createTable(a.columns, a.rows, a.size, a.showInput)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case ":":
			a.showInput = true
			a.input.Focus()
			a.table = createTable(a.columns, a.rows, a.size, a.showInput)
			return a, textinput.Blink
		case "r":
			return a, a.loadData()
		case "j", "down":
			if a.table.Cursor() < len(a.table.Rows())-1 {
				a.table.SetCursor(a.table.Cursor() + 1)
			}
		case "k", "up":
			if a.table.Cursor() > 0 {
				a.table.SetCursor(a.table.Cursor() - 1)
			}
		}
	case dataLoadedMsg:
		a.loading = false
		a.err = nil
		a.columns, a.rows = getTable(msg.data)
		a.table = createTable(a.columns, a.rows, a.size, a.showInput)
	case errorMsg:
		a.loading = false
		a.err = msg.err
	}

	a.table, cmd = a.table.Update(msg)
	return a, cmd
}

func getTable(data any) ([]table.Column, []table.Row) {
	switch d := data.(type) {
	case []internal.Module:
		return getModulesTable(d)
	case []internal.Component:
		return getComponentsTable(d)
	case []internal.Plan:
		return getPlansTable(d)
	case []internal.Apply:
		return getAppliesTable(d)
	case []internal.Changeset:
		return getChangesetsTable(d)
	default:
		return []table.Column{}, []table.Row{}
	}
}

func getModulesTable(modules []internal.Module) ([]table.Column, []table.Row) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Source", Width: 9},
	}

	var rows []table.Row
	for _, module := range modules {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(module.ID), 10),
			module.Source,
		})
	}

	return columns, rows
}

func getComponentsTable(components []internal.Component) ([]table.Column, []table.Row) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Module", Width: 7},
		{Title: "Version", Width: 2},
	}

	var rows []table.Row
	for _, component := range components {
		source := ""
		version := ""
		if component.ModuleVersion.Module.Source != "" {
			source = component.ModuleVersion.Module.Source
		}
		if component.ModuleVersion.Version != "" {
			version = component.ModuleVersion.Version
		}
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(component.ID), 10),
			source,
			version,
		})
	}

	return columns, rows
}

func getPlansTable(plans []internal.Plan) ([]table.Column, []table.Row) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Component", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	for _, plan := range plans {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(plan.ID), 10),
			strconv.FormatUint(uint64(plan.ComponentID), 10),
			plan.Changeset.Name,
			plan.State,
		})
	}

	return columns, rows
}

func getAppliesTable(applies []internal.Apply) ([]table.Column, []table.Row) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Plan", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	for _, apply := range applies {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(apply.ID), 10),
			strconv.FormatUint(uint64(apply.PlanID), 10),
			apply.Changeset.Name,
			apply.State,
		})
	}

	return columns, rows
}

func getChangesetsTable(changesets []internal.Changeset) ([]table.Column, []table.Row) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 7},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	for _, changeset := range changesets {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(changeset.ID), 10),
			changeset.Name,
			string(changeset.State),
		})
	}

	return columns, rows
}

func createTable(columns []table.Column, rows []table.Row, size Rect, showInput bool) table.Model {
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

	tableHeight := size.Height - 2
	if showInput {
		tableHeight -= 3
	}

	t := table.New(
		table.WithColumns(adjustedColumns),
		table.WithRows(rows),
		table.WithHeight(tableHeight),
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

	borderSpace := 2
	paddingSpace := 2
	availableWidth := totalWidth - borderSpace - paddingSpace
	adjusted := make([]table.Column, len(columns))
	allocatedWidth := 0
	for i, col := range columns {
		adjusted[i] = col
		if totalWeight > 0 {
			adjusted[i].Width = max(1, (col.Width*availableWidth)/totalWeight)
		}
		allocatedWidth += adjusted[i].Width
	}

	if len(adjusted) > 0 && allocatedWidth < availableWidth {
		adjusted[len(adjusted)-1].Width += availableWidth - allocatedWidth
	}

	return adjusted
}

func (a *App) loadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		switch a.currentView {
		case "modules":
			resp, err := a.client.ListModules(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{dataType: "modules", data: resp.Modules}
		case "components":
			resp, err := a.client.ListComponents(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{dataType: "components", data: resp.Components}
		case "plans":
			resp, err := a.client.ListPlans(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{dataType: "plans", data: resp.Plans}
		case "applies":
			resp, err := a.client.ListApplies(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{dataType: "applies", data: resp.Applies}
		case "changesets":
			resp, err := a.client.ListChangesets(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{dataType: "changesets", data: resp.Changesets}
		}

		return nil
	}
}

func (a *App) executeCommand(command string) tea.Cmd {
	return func() tea.Msg {
		parts := strings.Fields(command)
		if len(parts) == 0 {
			return nil
		}

		switch parts[0] {
		case "refresh", "r":
			return a.loadData()()
		case "quit", "q":
			return tea.Quit
		case "modules":
			a.currentView = "modules"
			return a.loadData()()
		case "components":
			a.currentView = "components"
			return a.loadData()()
		case "plans":
			a.currentView = "plans"
			return a.loadData()()
		case "applies":
			a.currentView = "applies"
			return a.loadData()()
		case "changesets":
			a.currentView = "changesets"
			return a.loadData()()
		}

		return nil
	}
}

type dataLoadedMsg struct {
	dataType string
	data     any
}

type errorMsg struct {
	err error
}

func (a *App) View() string {
	if a.loading {
		return "Loading...\nPress 'q' to quit, 'r' to refresh, ':' to enter commands"
	}

	if a.err != nil {
		return fmt.Sprintf("Error: %v\nPress 'r' to retry, 'q' to quit", a.err)
	}

	a.table.SetWidth(0)

	tableView := a.table.View()

	content := titledBox(a.currentView, tableView)

	if a.showInput {
		inputView := a.input.View()
		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1).
			Render(inputView)

		content = lipgloss.JoinVertical(lipgloss.Left, inputBox, content)
	}

	return content
}

func titledBox(title, content string) string {
	contentWidth := lipgloss.Width(content)
	titleWidth := lipgloss.Width(title)
	space := max(0, contentWidth-titleWidth)
	left := space / 2
	right := space - left

	border := lipgloss.RoundedBorder()
	top := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(border.TopLeft+strings.Repeat(border.Top, left)+" ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Render(title) + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(strings.Repeat(border.Top, right)+border.TopRight)

	body := lipgloss.NewStyle().
		Border(border).
		Padding(0, 1).
		BorderForeground(lipgloss.Color("8")).
		BorderTop(false).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, top, body)
}
