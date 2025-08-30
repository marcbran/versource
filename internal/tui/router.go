package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Page interface {
	Open(params map[string]string) tea.Cmd
	Links(params map[string]string) map[string]string
}

type Router struct {
	routes map[string]Page
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]Page),
	}
}

func (r *Router) Register(path string, page Page) {
	r.routes[path] = page
}

func (r *Router) Match(path string) (Page, map[string]string) {
	for routePath, page := range r.routes {
		if params := matchPath(routePath, path); params != nil {
			return page, params
		}
	}
	return nil, nil
}

func (r *Router) Open(path string) tea.Cmd {
	page, params := r.Match(path)
	if page == nil {
		return nil
	}
	return page.Open(params)
}

func (r *Router) Links(path string) map[string]string {
	page, params := r.Match(path)
	if page == nil {
		return nil
	}
	return page.Links(params)
}

func (r *Router) OpenLink(view string, link string) tea.Cmd {
	links := r.Links(view)
	if targetPath, ok := links[link]; ok {
		return r.Open(targetPath)
	}
	return nil
}

func matchPath(routePath, actualPath string) map[string]string {
	routeParts := strings.Split(routePath, "/")
	actualParts := strings.Split(actualPath, "/")

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

	return params
}
