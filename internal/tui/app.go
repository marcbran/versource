package tui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http"
)

type App struct {
	client     *http.Client
	currentTab int
	tabs       []string
	table      table.Model
	loading    bool
	err        error
}

func NewApp(client *http.Client) *App {
	t := table.New()
	t.SetStyles(table.Styles{
		Header:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("240")),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")),
	})

	return &App{
		client: client,
		tabs:   []string{"Modules", "Components", "Plans", "Applies", "Changesets"},
		table:  t,
	}
}

func (a *App) Init() tea.Cmd {
	return a.loadData()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "tab":
			a.currentTab = (a.currentTab + 1) % len(a.tabs)
			return a, a.loadData()
		case "shift+tab":
			a.currentTab = (a.currentTab - 1 + len(a.tabs)) % len(a.tabs)
			return a, a.loadData()
		case "r":
			return a, a.loadData()
		}
	case dataLoadedMsg:
		a.loading = false
		a.err = nil
		columns, rows := a.getTable(msg.data)
		a.updateTable(columns, rows)
	case errorMsg:
		a.loading = false
		a.err = msg.err
	}

	a.table, cmd = a.table.Update(msg)
	return a, cmd
}

func (a *App) getTable(data any) ([]table.Column, []table.Row) {
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
		{Title: "ID", Width: 5},
		{Title: "Source", Width: 50},
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
		{Title: "ID", Width: 5},
		{Title: "Module", Width: 40},
		{Title: "Version", Width: 15},
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
		{Title: "ID", Width: 5},
		{Title: "Component", Width: 10},
		{Title: "Changeset", Width: 15},
		{Title: "State", Width: 12},
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
		{Title: "ID", Width: 5},
		{Title: "Plan", Width: 5},
		{Title: "Changeset", Width: 15},
		{Title: "State", Width: 12},
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
		{Title: "ID", Width: 5},
		{Title: "Name", Width: 20},
		{Title: "State", Width: 12},
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

func (a *App) updateTable(columns []table.Column, rows []table.Row) {
	if len(rows) == 0 {
		placeholderRow := make(table.Row, len(columns))
		for i := range placeholderRow {
			placeholderRow[i] = ""
		}
		placeholderRow[0] = "No data"
		rows = append(rows, placeholderRow)
	}

	a.table.SetColumns(columns)
	a.table.SetRows(rows)
	a.table.SetStyles(table.Styles{
		Header:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("240")),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")),
	})
}

func (a *App) loadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		switch a.tabs[a.currentTab] {
		case "Modules":
			resp, err := a.client.ListModules(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{dataType: "modules", data: resp.Modules}
		case "Components":
			resp, err := a.client.ListComponents(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{dataType: "components", data: resp.Components}
		case "Plans":
			resp, err := a.client.ListPlans(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{dataType: "plans", data: resp.Plans}
		case "Applies":
			resp, err := a.client.ListApplies(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{dataType: "applies", data: resp.Applies}
		case "Changesets":
			resp, err := a.client.ListChangesets(ctx)
			if err != nil {
				return errorMsg{err: err}
			}
			return dataLoadedMsg{dataType: "changesets", data: resp.Changesets}
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
		return "Loading...\nPress 'q' to quit, 'r' to refresh, 'tab' to switch tabs"
	}

	if a.err != nil {
		return fmt.Sprintf("Error: %v\nPress 'r' to retry, 'q' to quit", a.err)
	}

	s := fmt.Sprintf("Versource TUI - %s\n\n", a.tabs[a.currentTab])
	s += "Controls: tab/shift+tab to switch tabs, r to refresh, q to quit\n\n"
	s += a.table.View()
	return s
}
