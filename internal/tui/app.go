package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcbran/versource/internal/http/client"
)

type App struct {
	client *client.Client
	router *Router

	loading bool
	err     error

	input     textinput.Model
	showInput bool

	size Size
}

func (a *App) contentSize() Size {
	contentWidth := a.size.Width - 4
	contentHeight := a.size.Height - 2
	if a.showInput {
		contentHeight -= 3
	}
	return Size{
		Width:  contentWidth,
		Height: contentHeight,
	}
}

type Size struct {
	Width  int
	Height int
}

var rootKeyBindings = KeyBindings{
	{Key: "m", Help: "View modules", Command: "modules"},
	{Key: "g", Help: "View changesets", Command: "changesets"},
	{Key: "c", Help: "View components", Command: "components"},
	{Key: "p", Help: "View plans", Command: "plans"},
	{Key: "a", Help: "View applies", Command: "applies"},
}

func NewApp(client *client.Client) *App {
	input := textinput.New()
	input.CharLimit = 100

	app := &App{
		client: client,
		router: NewRouter(),

		loading: true,

		input: input,
	}

	app.router.Register("modules", NewModulesPage(client))
	app.router.Register("modules/{moduleID}", NewModuleDetailPage(client))
	app.router.Register("modules/{moduleID}/moduleversions", NewModuleVersionsForModulePage(client))
	app.router.Register("moduleversions", NewModuleVersionsPage(client))
	app.router.Register("changesets", NewChangesetsPage(client))
	app.router.Register("changesets/{changesetName}/components", NewChangesetComponentsPage(client))
	app.router.Register("changesets/{changesetName}/components/diffs", NewComponentDiffsPage(client))
	app.router.Register("changesets/{changesetName}/plans", NewChangesetPlansPage(client))
	app.router.Register("changesets/{changesetName}/applies", NewChangesetAppliesPage(client))
	app.router.Register("components", NewComponentsPage(client))
	app.router.Register("plans", NewPlansPage(client))
	app.router.Register("plans/{planID}/logs", NewPlanLogsPage(client))
	app.router.Register("applies", NewAppliesPage(client))

	return app
}

func (a *App) Init() tea.Cmd {
	return a.router.Open("modules")
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
				a.input.Blur()
				a.showInput = false
				a.input.SetValue("")
				a.router.Focus()
				a.router.Resize(a.contentSize())
				return a, nil
			case "enter":
				command := a.input.Value()
				a.input.Blur()
				a.showInput = false
				a.input.SetValue("")
				a.router.Focus()
				a.router.Resize(a.contentSize())
				return a, a.router.executeCommand(command)
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
		a.router.Resize(a.contentSize())
	case routeOpenedMsg:
		a.input.Blur()
		a.router.Resize(a.contentSize())
		a.router.Focus()
	case dataLoadedMsg:
		a.loading = false
		a.err = nil
	case errorMsg:
		a.loading = false
		a.err = msg.err
	case tea.KeyMsg:
		switch msg.String() {
		case ":":
			a.showInput = true
			a.input.Focus()
			a.router.Blur()
			a.router.Resize(a.contentSize())
			return a, textinput.Blink
		}
	}

	a.router, cmd = a.router.Update(msg)
	return a, cmd
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

	content := a.router.View()

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
