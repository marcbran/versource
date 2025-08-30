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
	viewHistory []string
	table       table.Model
	columns     []table.Column
	rows        []table.Row
	rowIds      []string
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
		viewHistory: []string{},
		table:       table.New(),
		input:       ti,
	}
}

func (a *App) Init() tea.Cmd {
	return a.loadData(a.currentView)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		}
	}

	if a.showInput {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
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
		case ":":
			a.showInput = true
			a.input.Focus()
			a.table = createTable(a.columns, a.rows, a.size, a.showInput)
			return a, textinput.Blink
		case "r":
			return a, a.executeCommand("refresh")
		case "j", "down":
			if a.table.Cursor() < len(a.table.Rows())-1 {
				a.table.SetCursor(a.table.Cursor() + 1)
			}
		case "k", "up":
			if a.table.Cursor() > 0 {
				a.table.SetCursor(a.table.Cursor() - 1)
			}
		case "esc":
			return a, a.goBack()
		case "enter":
			if len(a.rowIds) > 0 && a.table.Cursor() >= 0 && a.table.Cursor() < len(a.rowIds) {
				selectedId := a.rowIds[a.table.Cursor()]
				return a, a.executePathCommand(fmt.Sprintf("/%s/%s", a.currentView, selectedId))
			}
		}
	case dataLoadedMsg:
		a.loading = false
		a.err = nil
		if a.currentView != msg.view {
			a.viewHistory = append(a.viewHistory, a.currentView)
		}
		a.currentView = msg.view
		a.columns, a.rows, a.rowIds = getTable(msg.data)
		a.table = createTable(a.columns, a.rows, a.size, a.showInput)
	case errorMsg:
		a.loading = false
		a.err = msg.err
	}

	a.table, cmd = a.table.Update(msg)
	return a, cmd
}

func getTable(data any) ([]table.Column, []table.Row, []string) {
	switch d := data.(type) {
	case []internal.Module:
		return getModulesTable(d)
	case []internal.ModuleVersion:
		return getModuleVersionsTable(d)
	case []internal.Changeset:
		return getChangesetsTable(d)
	case []internal.Component:
		return getComponentsTable(d)
	case []internal.Plan:
		return getPlansTable(d)
	case []internal.Apply:
		return getAppliesTable(d)
	default:
		return []table.Column{}, []table.Row{}, []string{}
	}
}

func getModulesTable(modules []internal.Module) ([]table.Column, []table.Row, []string) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Source", Width: 9},
	}

	var rows []table.Row
	var ids []string
	for _, module := range modules {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(module.ID), 10),
			module.Source,
		})
		ids = append(ids, strconv.FormatUint(uint64(module.ID), 10))
	}

	return columns, rows, ids
}

func getModuleVersionsTable(moduleVersions []internal.ModuleVersion) ([]table.Column, []table.Row, []string) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Module", Width: 7},
		{Title: "Version", Width: 2},
	}

	var rows []table.Row
	var ids []string
	for _, moduleVersion := range moduleVersions {
		source := ""
		if moduleVersion.Module.Source != "" {
			source = moduleVersion.Module.Source
		}
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(moduleVersion.ID), 10),
			source,
			moduleVersion.Version,
		})
		ids = append(ids, strconv.FormatUint(uint64(moduleVersion.ID), 10))
	}

	return columns, rows, ids
}

func getChangesetsTable(changesets []internal.Changeset) ([]table.Column, []table.Row, []string) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Name", Width: 7},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	var ids []string
	for _, changeset := range changesets {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(changeset.ID), 10),
			changeset.Name,
			string(changeset.State),
		})
		ids = append(ids, strconv.FormatUint(uint64(changeset.ID), 10))
	}

	return columns, rows, ids
}

func getComponentsTable(components []internal.Component) ([]table.Column, []table.Row, []string) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Module", Width: 7},
		{Title: "Version", Width: 2},
	}

	var rows []table.Row
	var ids []string
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
		ids = append(ids, strconv.FormatUint(uint64(component.ID), 10))
	}

	return columns, rows, ids
}

func getPlansTable(plans []internal.Plan) ([]table.Column, []table.Row, []string) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Component", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	var ids []string
	for _, plan := range plans {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(plan.ID), 10),
			strconv.FormatUint(uint64(plan.ComponentID), 10),
			plan.Changeset.Name,
			plan.State,
		})
		ids = append(ids, strconv.FormatUint(uint64(plan.ID), 10))
	}

	return columns, rows, ids
}

func getAppliesTable(applies []internal.Apply) ([]table.Column, []table.Row, []string) {
	columns := []table.Column{
		{Title: "ID", Width: 1},
		{Title: "Plan", Width: 1},
		{Title: "Changeset", Width: 6},
		{Title: "State", Width: 2},
	}

	var rows []table.Row
	var ids []string
	for _, apply := range applies {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(apply.ID), 10),
			strconv.FormatUint(uint64(apply.PlanID), 10),
			apply.Changeset.Name,
			apply.State,
		})
		ids = append(ids, strconv.FormatUint(uint64(apply.ID), 10))
	}

	return columns, rows, ids
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

func (a *App) executePathCommand(path string) tea.Cmd {
	return func() tea.Msg {
		if path == "" {
			return nil
		}

		switch {
		case strings.HasPrefix(path, "/modules/"):
			parts := strings.Split(path, "/")
			if len(parts) == 3 {
				moduleID := parts[2]
				return a.loadData(fmt.Sprintf("modules/%s/moduleversions", moduleID))()
			}
		case strings.HasPrefix(path, "/moduleversions/"):
		case strings.HasPrefix(path, "/changesets/"):
		case strings.HasPrefix(path, "/components/"):
		case strings.HasPrefix(path, "/plans/"):
		case strings.HasPrefix(path, "/applies/"):
		}

		return nil
	}
}

func (a *App) executeCommand(command string) tea.Cmd {
	return func() tea.Msg {
		if command == "" {
			return nil
		}

		switch command {
		case "refresh", "r":
			return a.loadData(a.currentView)()
		case "back", "b":
			return a.goBack()
		case "modules":
			return a.loadData("modules")()
		case "moduleversions":
			return a.loadData("moduleversions")()
		case "components":
			return a.loadData("components")()
		case "plans":
			return a.loadData("plans")()
		case "applies":
			return a.loadData("applies")()
		case "changesets":
			return a.loadData("changesets")()
		default:
			if !strings.HasPrefix(command, "modules/") || !strings.HasSuffix(command, "/moduleversions") {
				return nil
			}
			parts := strings.Split(command, "/")
			if len(parts) != 3 {
				return nil
			}
			_, err := strconv.ParseUint(parts[1], 10, 32)
			if err != nil {
				return nil
			}
			return a.loadData(command)()
		}
	}
}

func (a *App) goBack() tea.Cmd {
	return func() tea.Msg {
		if len(a.viewHistory) > 0 {
			previousView := a.viewHistory[len(a.viewHistory)-1]
			a.viewHistory = a.viewHistory[:len(a.viewHistory)-1]
			return a.loadData(previousView)()
		}
		return nil
	}
}

func (a *App) loadData(view string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		switch view {
		case "modules":
			resp, err := a.client.ListModules(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{view: view, dataType: "modules", data: resp.Modules}
		case "moduleversions":
			resp, err := a.client.ListModuleVersions(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{view: view, dataType: "moduleversions", data: resp.ModuleVersions}
		case "components":
			resp, err := a.client.ListComponents(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{view: view, dataType: "components", data: resp.Components}
		case "plans":
			resp, err := a.client.ListPlans(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{view: view, dataType: "plans", data: resp.Plans}
		case "applies":
			resp, err := a.client.ListApplies(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{view: view, dataType: "applies", data: resp.Applies}
		case "changesets":
			resp, err := a.client.ListChangesets(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{view: view, dataType: "changesets", data: resp.Changesets}
		default:
			if !strings.HasPrefix(view, "modules/") || !strings.HasSuffix(view, "/moduleversions") {
				return nil
			}
			parts := strings.Split(view, "/")
			if len(parts) != 3 {
				return nil
			}
			moduleID, err := strconv.ParseUint(parts[1], 10, 32)
			if err != nil {
				return nil
			}
			resp, err := a.client.ListModuleVersionsForModule(ctx, uint(moduleID))
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{view: view, dataType: "moduleversions", data: resp.ModuleVersions}
		}
	}
}

type dataLoadedMsg struct {
	view     string
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
