package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Route struct {
	Path string
	Page Page
}

type Page interface {
	tea.Model
	Resize(size Rect)
	Links() map[string]string
}

type Router struct {
	routes map[string]func(map[string]string) Page
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]func(map[string]string) Page),
	}
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
