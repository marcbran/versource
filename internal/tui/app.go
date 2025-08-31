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
	router      *Router
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

func NewApp(client *http.Client) *App {
	ti := textinput.New()
	ti.CharLimit = 100

	app := &App{
		client:      client,
		router:      NewRouter(),
		currentView: "modules",
		viewHistory: []string{},
		table:       table.New(),
		input:       ti,
	}

	app.router.Register("modules", &ModulesPage{app: app})
	app.router.Register("modules/{moduleID}", &ModulePage{app: app})
	app.router.Register("modules/{moduleID}/moduleversions", &ModuleVersionsForModulePage{app: app})
	app.router.Register("moduleversions", &ModuleVersionsPage{app: app})
	app.router.Register("changesets", &ChangesetsPage{app: app})
	app.router.Register("changesets/{changesetName}", &ChangesetPage{app: app})
	app.router.Register("changesets/{changesetName}/components", &ChangesetComponentsPage{app: app})
	app.router.Register("changesets/{changesetName}/plans", &ChangesetPlansPage{app: app})
	app.router.Register("changesets/{changesetName}/applies", &ChangesetAppliesPage{app: app})
	app.router.Register("components", &ComponentsPage{app: app})
	app.router.Register("plans", &PlansPage{app: app})
	app.router.Register("applies", &AppliesPage{app: app})

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
		{Title: "Review", Width: 2},
	}

	var rows []table.Row
	var ids []string
	for _, changeset := range changesets {
		rows = append(rows, table.Row{
			strconv.FormatUint(uint64(changeset.ID), 10),
			changeset.Name,
			string(changeset.State),
			string(changeset.ReviewState),
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
		page, params := a.router.Match(a.currentView)
		if page != nil {
			return page.Open(params)()
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

type ModulesPage struct {
	app *App
}

func (p *ModulesPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.app.client.ListModules(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "modules", data: resp.Modules}
	}
}

func (p *ModulesPage) Links(params map[string]string) map[string]string {
	return map[string]string{}
}

type ModulePage struct {
	app *App
}

func (p *ModulePage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		return dataLoadedMsg{view: fmt.Sprintf("modules/%s", params["moduleID"]), data: nil}
	}
}

func (p *ModulePage) Links(params map[string]string) map[string]string {
	return map[string]string{
		"enter": fmt.Sprintf("modules/%s/moduleversions", params["moduleID"]),
		"c":     fmt.Sprintf("components?module-id=%s", params["moduleID"]),
	}
}

type ModuleVersionsPage struct {
	app *App
}

func (p *ModuleVersionsPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.app.client.ListModuleVersions(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "moduleversions", data: resp.ModuleVersions}
	}
}

func (p *ModuleVersionsPage) Links(params map[string]string) map[string]string {
	return map[string]string{
		"c": fmt.Sprintf("components?module-version-id=%s", params["moduleVersionID"]),
	}
}

type ModuleVersionsForModulePage struct {
	app *App
}

func (p *ModuleVersionsForModulePage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		moduleID, exists := params["moduleID"]
		if !exists {
			return errorMsg{err: fmt.Errorf("moduleID parameter required")}
		}

		moduleIDUint, err := strconv.ParseUint(moduleID, 10, 32)
		if err != nil {
			return errorMsg{err: err}
		}
		resp, err := p.app.client.ListModuleVersionsForModule(ctx, uint(moduleIDUint))
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: fmt.Sprintf("modules/%s/moduleversions", moduleID), data: resp.ModuleVersions}
	}
}

func (p *ModuleVersionsForModulePage) Links(params map[string]string) map[string]string {
	return map[string]string{
		"m": "modules",
	}
}

type ChangesetsPage struct {
	app *App
}

func (p *ChangesetsPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.app.client.ListChangesets(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "changesets", data: resp.Changesets}
	}
}

func (p *ChangesetsPage) Links(params map[string]string) map[string]string {
	return map[string]string{}
}

type ChangesetPage struct {
	app *App
}

func (p *ChangesetPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		return dataLoadedMsg{view: fmt.Sprintf("changesets/%s", params["changesetName"]), data: nil}
	}
}

func (p *ChangesetPage) Links(params map[string]string) map[string]string {
	changesetName := params["changesetName"]
	return map[string]string{
		"enter": fmt.Sprintf("changesets/%s/components", changesetName),
		"c":     fmt.Sprintf("changesets/%s/components", changesetName),
		"p":     fmt.Sprintf("changesets/%s/plans", changesetName),
		"a":     fmt.Sprintf("changesets/%s/applies", changesetName),
	}
}

type ComponentsPage struct {
	app *App
}

func (p *ComponentsPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		req := internal.ListComponentsRequest{}

		if moduleIDStr, ok := params["module-id"]; ok {
			if moduleID, err := strconv.ParseUint(moduleIDStr, 10, 32); err == nil {
				moduleIDUint := uint(moduleID)
				req.ModuleID = &moduleIDUint
			}
		}

		if moduleVersionIDStr, ok := params["module-version-id"]; ok {
			if moduleVersionID, err := strconv.ParseUint(moduleVersionIDStr, 10, 32); err == nil {
				moduleVersionIDUint := uint(moduleVersionID)
				req.ModuleVersionID = &moduleVersionIDUint
			}
		}

		resp, err := p.app.client.ListComponents(ctx, req)
		if err != nil {
			return errorMsg{err: err}
		}

		view := "components"
		if len(params) > 0 {
			queryParts := make([]string, 0)
			if moduleIDStr, ok := params["module-id"]; ok {
				queryParts = append(queryParts, fmt.Sprintf("module-id=%s", moduleIDStr))
			}
			if moduleVersionIDStr, ok := params["module-version-id"]; ok {
				queryParts = append(queryParts, fmt.Sprintf("module-version-id=%s", moduleVersionIDStr))
			}
			if len(queryParts) > 0 {
				view = fmt.Sprintf("components?%s", strings.Join(queryParts, "&"))
			}
		}

		return dataLoadedMsg{view: view, data: resp.Components}
	}
}

func (p *ComponentsPage) Links(params map[string]string) map[string]string {
	return map[string]string{
		"m": "modules",
		"v": "moduleversions",
	}
}

type PlansPage struct {
	app *App
}

func (p *PlansPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.app.client.ListPlans(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "plans", data: resp.Plans}
	}
}

func (p *PlansPage) Links(params map[string]string) map[string]string {
	return map[string]string{}
}

type AppliesPage struct {
	app *App
}

func (p *AppliesPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := p.app.client.ListApplies(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{view: "applies", data: resp.Applies}
	}
}

func (p *AppliesPage) Links(params map[string]string) map[string]string {
	return map[string]string{}
}

type ChangesetComponentsPage struct {
	app *App
}

func (p *ChangesetComponentsPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		changesetName := params["changesetName"]

		req := internal.ListComponentsRequest{}

		resp, err := p.app.client.ListComponents(ctx, req)
		if err != nil {
			return errorMsg{err: err}
		}

		view := fmt.Sprintf("changesets/%s/components", changesetName)
		return dataLoadedMsg{view: view, data: resp.Components}
	}
}

func (p *ChangesetComponentsPage) Links(params map[string]string) map[string]string {
	changesetName := params["changesetName"]
	return map[string]string{
		"p": fmt.Sprintf("changesets/%s/plans", changesetName),
		"a": fmt.Sprintf("changesets/%s/applies", changesetName),
	}
}

type ChangesetPlansPage struct {
	app *App
}

func (p *ChangesetPlansPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		changesetName := params["changesetName"]

		resp, err := p.app.client.ListPlans(ctx)
		if err != nil {
			return errorMsg{err: err}
		}

		view := fmt.Sprintf("changesets/%s/plans", changesetName)
		return dataLoadedMsg{view: view, data: resp.Plans}
	}
}

func (p *ChangesetPlansPage) Links(params map[string]string) map[string]string {
	changesetName := params["changesetName"]
	return map[string]string{
		"c": fmt.Sprintf("changesets/%s/components", changesetName),
		"a": fmt.Sprintf("changesets/%s/applies", changesetName),
	}
}

type ChangesetAppliesPage struct {
	app *App
}

func (p *ChangesetAppliesPage) Open(params map[string]string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		changesetName := params["changesetName"]

		resp, err := p.app.client.ListApplies(ctx)
		if err != nil {
			return errorMsg{err: err}
		}

		view := fmt.Sprintf("changesets/%s/applies", changesetName)
		return dataLoadedMsg{view: view, data: resp.Applies}
	}
}

func (p *ChangesetAppliesPage) Links(params map[string]string) map[string]string {
	changesetName := params["changesetName"]
	return map[string]string{
		"c": fmt.Sprintf("changesets/%s/components", changesetName),
		"p": fmt.Sprintf("changesets/%s/plans", changesetName),
	}
}
