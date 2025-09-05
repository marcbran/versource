package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Route struct {
	Path string
	Page Page
}

type Resizer interface {
	Resize(size Size)
}

type Focuser interface {
	Focus()
	Blur()
}

type Page interface {
	Init() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View() string
	Resizer
	Focuser
	Links() map[string]string
}

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
	r.focussed = true
	r.updateCurrentRoute()
}

func (r *Router) refresh() tea.Cmd {
	return func() tea.Msg {
		cmd := r.Open(r.currentRoute.Path)
		if cmd != nil {
			return cmd()
		}
		return nil
	}
}

func (r *Router) goBack() tea.Cmd {
	return func() tea.Msg {
		if len(r.routeHistory) > 0 {
			previousRoute := r.routeHistory[len(r.routeHistory)-1]
			r.routeHistory = r.routeHistory[:len(r.routeHistory)-1]
			r.currentRoute = &previousRoute
			return r.refresh()()
		}
		return nil
	}
}

func (r *Router) Update(msg tea.Msg) (*Router, tea.Cmd) {
	switch msg := msg.(type) {
	case routeOpenedMsg:
		if r.currentRoute != nil && r.currentRoute.Path != msg.route.Path {
			r.routeHistory = append(r.routeHistory, *r.currentRoute)
		}
		r.currentRoute = &msg.route
		r.updateCurrentRoute()
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			return r, r.refresh()
		case "esc":
			return r, r.goBack()
		default:
			if link, ok := r.currentRoute.Page.Links()[msg.String()]; ok {
				return r, r.Open(link)
			}
		}
	}
	if r.currentRoute == nil {
		return r, nil
	}
	var cmd tea.Cmd
	r.currentRoute.Page, cmd = r.currentRoute.Page.Update(msg)
	return r, cmd
}

func (r *Router) updateCurrentRoute() {
	if r.currentRoute == nil {
		return
	}
	r.currentRoute.Page.Resize(r.size)
	if r.focussed {
		r.currentRoute.Page.Focus()
	} else {
		r.currentRoute.Page.Blur()
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

func (r *Router) View() string {
	return titledBox(r.currentRoute.Path, r.currentRoute.Page.View())
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
			return routeOpenedMsg{route: Route{Path: path, Page: page}}
		},
		page.Init(),
	)
}

type routeOpenedMsg struct {
	route Route
}

type dataLoadedMsg struct {
	data any
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
