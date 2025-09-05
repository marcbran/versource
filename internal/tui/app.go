package tui

import (
	"fmt"
	"strings"

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

	size Rect

	currentRoute *Route
	routeHistory []Route
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

type IndexPage struct{}

func (p *IndexPage) Init() tea.Cmd {
	return nil
}

func (p *IndexPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return p, nil
}

func (p *IndexPage) View() string {
	return ""
}

func (p *IndexPage) Resize(size Rect) {}

func (p *IndexPage) Links() map[string]string {
	return map[string]string{
		"m": "modules",
		"c": "changesets",
		"p": "plans",
		"a": "applies",
	}
}

func (p *IndexPage) Focus() {
}

func (p *IndexPage) Blur() {
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

		loading: true,

		input: input,

		currentRoute: &Route{Path: "", Page: &IndexPage{}},
	}

	app.router.Register("modules", NewModulesPage(client))
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
				a.currentRoute.Page.Focus()
				a.currentRoute.Page.Resize(a.contentSize())
				return a, nil
			case "enter":
				command := a.input.Value()
				a.input.Blur()
				a.showInput = false
				a.input.SetValue("")
				a.currentRoute.Page.Focus()
				a.currentRoute.Page.Resize(a.contentSize())
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
		a.currentRoute.Page.Resize(a.contentSize())
	case routeOpenedMsg:
		if a.currentRoute.Path != msg.route.Path {
			a.routeHistory = append(a.routeHistory, *a.currentRoute)
		}
		a.currentRoute = &msg.route
		a.currentRoute.Page.Resize(a.contentSize())
		a.input.Blur()
		a.currentRoute.Page.Focus()
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
			a.currentRoute.Page.Blur()
			a.currentRoute.Page.Resize(a.contentSize())
			return a, textinput.Blink
		case "r":
			return a, a.refresh()
		case "esc":
			return a, a.goBack()
		default:
			if link, ok := a.currentRoute.Page.Links()[msg.String()]; ok {
				return a, a.router.Open(link)
			}
		}
	}

	page, cmd := a.currentRoute.Page.Update(msg)
	if p, ok := page.(Page); ok {
		a.currentRoute.Page = p
	}
	return a, cmd
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
		cmd := a.router.Open(a.currentRoute.Path)
		if cmd != nil {
			return cmd()
		}
		return nil
	}
}

func (a *App) goBack() tea.Cmd {
	return func() tea.Msg {
		if len(a.routeHistory) > 0 {
			previousRoute := a.routeHistory[len(a.routeHistory)-1]
			a.routeHistory = a.routeHistory[:len(a.routeHistory)-1]
			a.currentRoute = &previousRoute
			return a.refresh()()
		}
		return nil
	}
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

	content := titledBox(a.currentRoute.Path, a.currentRoute.Page.View())

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
