package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Router struct {
	routes map[string]func(map[string]string) Page

	currentRoute *Route
	routeHistory []Route

	size     Size
	focussed bool
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]func(map[string]string) Page),
	}
}

func (r *Router) Init() tea.Cmd {
	if r.currentRoute != nil {
		return r.currentRoute.Init()
	}
	return nil
}

func (r *Router) Resize(size Size) {
	r.size = size
	r.updateCurrentRoute()
}

func (r *Router) Focus() {
	r.focussed = true
	r.updateCurrentRoute()
}

func (r *Router) Blur() {
	r.focussed = false
	r.updateCurrentRoute()
}

func (r *Router) Update(msg tea.Msg) (*Router, tea.Cmd) {
	switch m := msg.(type) {
	case routeOpenedMsg:
		r.currentRoute.page = m.page
		r.updateCurrentRoute()
		if m.msg != nil {
			return r, func() tea.Msg { return m.msg }
		}
		return r, nil
	case errorMsg:
		r.currentRoute.err = m.err
		r.updateCurrentRoute()
		return r, nil
	}

	if r.currentRoute != nil && r.currentRoute.IsLoading() {
		var cmd tea.Cmd
		*r.currentRoute, cmd = r.currentRoute.Update(msg)
		return r, cmd
	}

	switch m := msg.(type) {
	case openPageRequestedMsg:
		if r.currentRoute != nil && r.currentRoute.path != m.path {
			r.routeHistory = append(r.routeHistory, *r.currentRoute)
		}
		route := NewRoute(m.path)
		r.currentRoute = &route
		r.updateCurrentRoute()
		return r, route.Init()
	case goBackRequestedMsg:
		if len(r.routeHistory) > 0 {
			previousRoute := r.routeHistory[len(r.routeHistory)-1]
			r.routeHistory = r.routeHistory[:len(r.routeHistory)-1]
			r.currentRoute = &previousRoute
			r.updateCurrentRoute()
		}
		return r, nil
	case tea.KeyMsg:
		switch m.String() {
		case "r":
			return r, r.refresh()
		case "esc":
			return r, r.goBack()
		default:
			if r.currentRoute != nil && r.currentRoute.page != nil {
				keyBindings := r.currentRoute.page.KeyBindings()
				for _, binding := range keyBindings {
					if binding.Key == m.String() {
						return r, r.Open(binding.Command)
					}
				}
			}
		}
	}

	if r.currentRoute == nil {
		return r, nil
	}
	var cmd tea.Cmd
	*r.currentRoute, cmd = r.currentRoute.Update(msg)
	return r, cmd
}

func (r *Router) updateCurrentRoute() {
	if r.currentRoute == nil {
		return
	}
	r.currentRoute.Resize(r.size)
	if r.focussed {
		r.currentRoute.Focus()
	} else {
		r.currentRoute.Blur()
	}
}

func (r *Router) executeCommand(command string) tea.Cmd {
	return func() tea.Msg {
		if command == "" {
			return nil
		}

		switch command {
		case "refresh", "r":
			return r.refresh()
		case "back", "b":
			return r.goBack()
		default:
			cmd := r.Open(command)
			if cmd != nil {
				return cmd()
			}
			return nil
		}
	}
}

func (r *Router) refresh() tea.Cmd {
	return func() tea.Msg {
		if r.currentRoute == nil {
			return nil
		}
		cmd := r.Open(r.currentRoute.path)
		if cmd != nil {
			return cmd()
		}
		return nil
	}
}

func (r *Router) goBack() tea.Cmd {
	return func() tea.Msg {
		return goBackRequestedMsg{}
	}
}

func (r *Router) View() string {
	if r.currentRoute == nil {
		return ""
	}
	return r.currentRoute.View()
}

func (r *Router) Register(path string, page func(map[string]string) Page) {
	r.routes[path] = page
}

func (r *Router) Match(path string) Page {
	for routePath, page := range r.routes {
		if params := matchPath(routePath, path); params != nil {
			return page(params)
		}
	}
	return nil
}

func (r *Router) Open(path string) tea.Cmd {
	page := r.Match(path)
	if page == nil {
		return nil
	}
	return tea.Sequence(
		func() tea.Msg {
			return openPageRequestedMsg{path: path}
		},
		func() tea.Msg {
			return routeOpenedMsg{path: path, page: page, msg: page.Init()()}
		},
	)
}

type openPageRequestedMsg struct {
	path string
}

type routeOpenedMsg struct {
	path string
	page Page
	msg  tea.Msg
}

type goBackRequestedMsg struct{}

type dataLoadedMsg struct {
	data any
}

type errorMsg struct {
	err error
}

func matchPath(routePath, actualPath string) map[string]string {
	actualPathParts := strings.SplitN(actualPath, "?", 2)

	routeParts := strings.Split(routePath, "/")
	actualParts := strings.Split(actualPathParts[0], "/")

	if len(routeParts) != len(actualParts) {
		return nil
	}

	params := make(map[string]string)

	for i, routePart := range routeParts {
		actualPart := actualParts[i]

		if strings.HasPrefix(routePart, "{") && strings.HasSuffix(routePart, "}") {
			paramName := routePart[1 : len(routePart)-1]
			params[paramName] = actualPart
		} else if routePart != actualPart {
			return nil
		}
	}

	if len(actualPathParts) > 1 {
		actualQuery := actualPathParts[1]
		actualQueryParams := parseQueryString(actualQuery)

		for key, value := range actualQueryParams {
			params[key] = value
		}
	}

	return params
}

func parseQueryString(query string) map[string]string {
	params := make(map[string]string)
	if query == "" {
		return params
	}

	pairs := strings.Split(query, "&")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			params[kv[0]] = kv[1]
		}
	}
	return params
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

type Route struct {
	path string
	page Page
	err  error

	size     Size
	loadView LoadView
}

func NewRoute(path string) Route {
	return Route{
		path:     path,
		loadView: NewLoadView(),
	}
}

func (r Route) Init() tea.Cmd {
	return r.loadView.Init()
}

func (r Route) IsLoading() bool {
	return r.page == nil && r.err == nil
}

func (r *Route) Resize(size Size) {
	r.size = size
	r.loadView.Resize(size)
	if r.page != nil {
		r.page.Resize(size)
	}
}

func (r *Route) Focus() {
	if r.page != nil {
		r.page.Focus()
	}
}

func (r *Route) Blur() {
	if r.page != nil {
		r.page.Blur()
	}
}

func (r Route) Update(msg tea.Msg) (Route, tea.Cmd) {
	if r.IsLoading() {
		var cmd tea.Cmd
		r.loadView, cmd = r.loadView.Update(msg)
		return r, cmd
	}
	var cmd tea.Cmd
	r.page, cmd = r.page.Update(msg)
	return r, cmd
}

func (r Route) View() string {
	if r.err != nil {
		errorText := fmt.Sprintf("Error: %v", r.err)

		centeredContent := lipgloss.NewStyle().
			Width(r.size.Width).
			Height(r.size.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(errorText)

		return titledBox("Error", centeredContent)
	}

	if r.page == nil {
		return titledBox(r.path, r.loadView.View())
	}

	return titledBox(r.path, r.page.View())
}

type Page interface {
	Init() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View() string
	Resizer
	Focuser
	KeyBindings() KeyBindings
}

type KeyBindings []KeyBinding

func (k KeyBindings) With(key, help, command string) KeyBindings {
	return append(k, KeyBinding{Key: key, Help: help, Command: command})
}

type KeyBinding struct {
	Key     string
	Help    string
	Command string
}

type Resizer interface {
	Resize(size Size)
}

type Focuser interface {
	Focus()
	Blur()
}
