package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
)

type App struct {
	client *client.Client
	router *Router

	loading bool
	err     error

	input     textinput.Model
	showInput bool

	size Rect

	currentView string
	viewHistory []string

	table   table.Model
	columns []table.Column
	rows    []table.Row
	rowIds  []string
}

func (a *App) contentSize() Rect {
	contentWidth := a.size.Width - 4
	contentHeight := a.size.Height - 2
	if a.showInput {
		contentHeight -= 3
	}
	return Rect{
		Width:  contentWidth,
		Height: contentHeight,
	}
}

func (a *App) cursorView() string {
	if len(a.rowIds) == 0 || a.table.Cursor() < 0 || a.table.Cursor() >= len(a.rowIds) {
		return ""
	}
	selectedId := a.rowIds[a.table.Cursor()]
	return fmt.Sprintf("%s/%s", a.currentView, selectedId)
}

type Rect struct {
	Width  int
	Height int
}

func NewApp(client *client.Client) *App {
	input := textinput.New()
	input.CharLimit = 100

	app := &App{
		client: client,
		router: NewRouter(),

		input: input,

		currentView: "modules",
		viewHistory: []string{},

		table: table.New(),
	}

	app.router.Register("modules", NewModulesPage(client))
	app.router.Register("modules/{moduleID}", NewModulePage(client))
	app.router.Register("modules/{moduleID}/moduleversions", NewModuleVersionsForModulePage(client))
	app.router.Register("moduleversions", NewModuleVersionsPage(client))
	app.router.Register("changesets", NewChangesetsPage(client))
	app.router.Register("changesets/{changesetName}", NewChangesetPage(client))
	app.router.Register("changesets/{changesetName}/components", NewChangesetComponentsPage(client))
	app.router.Register("changesets/{changesetName}/plans", NewChangesetPlansPage(client))
	app.router.Register("changesets/{changesetName}/applies", NewChangesetAppliesPage(client))
	app.router.Register("components", NewComponentsPage(client))
	app.router.Register("plans", NewPlansPage(client))
	app.router.Register("applies", NewAppliesPage(client))

	return app
}

func (a *App) Init() tea.Cmd {
	return a.refresh()
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
				a.table = newTable(a.columns, a.rows, a.contentSize())
				return a, nil
			case "enter":
				command := a.input.Value()
				a.showInput = false
				a.input.SetValue("")
				a.table = newTable(a.columns, a.rows, a.contentSize())
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
		a.table = newTable(a.columns, a.rows, a.contentSize())
	case tea.KeyMsg:
		switch msg.String() {
		case ":":
			a.showInput = true
			a.input.Focus()
			a.table = newTable(a.columns, a.rows, a.contentSize())
			return a, textinput.Blink
		case "r":
			return a, a.refresh()
		case "esc":
			return a, a.goBack()
		case "j", "down":
			if a.table.Cursor() < len(a.table.Rows())-1 {
				a.table.SetCursor(a.table.Cursor() + 1)
			}
		case "k", "up":
			if a.table.Cursor() > 0 {
				a.table.SetCursor(a.table.Cursor() - 1)
			}
		default:
			cmd := a.router.OpenLink(a.cursorView(), msg.String())
			if cmd != nil {
				return a, cmd
			}
			cmd = a.router.OpenLink(a.currentView, msg.String())
			if cmd != nil {
				return a, cmd
			}
			return a, nil
		}
	case dataLoadedMsg:
		a.loading = false
		a.err = nil
		if a.currentView != msg.view {
			a.viewHistory = append(a.viewHistory, a.currentView)
		}
		a.currentView = msg.view
		a.columns, a.rows, a.rowIds = getTable(msg.data)
		a.table = newTable(a.columns, a.rows, a.contentSize())
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

func (a *App) executeCommand(command string) tea.Cmd {
	return func() tea.Msg {
		if command == "" {
			return nil
		}

		switch command {
		case "refresh", "r":
			return a.refresh()
		case "back", "b":
			return a.goBack()
		default:
			cmd := a.router.Open(command)
			if cmd != nil {
				return cmd()
			}
			return nil
		}
	}
}

func (a *App) refresh() tea.Cmd {
	return func() tea.Msg {
		cmd := a.router.Open(a.currentView)
		if cmd != nil {
			return cmd()
		}
		return nil
	}
}

func (a *App) goBack() tea.Cmd {
	return func() tea.Msg {
		if len(a.viewHistory) > 0 {
			previousView := a.viewHistory[len(a.viewHistory)-1]
			a.viewHistory = a.viewHistory[:len(a.viewHistory)-1]
			a.currentView = previousView
			return a.refresh()()
		}
		return nil
	}
}

type dataLoadedMsg struct {
	view string
	data any
}

type errorMsg struct {
	err error
}

func (a *App) View() string {
	if a.loading {
		return "Loading...\nPress 'r' to refresh, 'esc' to go back, 'ctrl+c' to quit, ':' to enter commands"
	}

	if a.err != nil {
		return fmt.Sprintf("Error: %v\nPress 'r' to refresh, 'esc' to go back, 'ctrl+c' to quit, ':' to enter commands", a.err)
	}

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
