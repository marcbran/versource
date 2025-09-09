package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type DataViewport[T any] struct {
	viewport viewport.Model
	elem     T
	content  string

	size Size
	data ViewportData[T]
}

func NewDataViewport[T any](data ViewportData[T]) *DataViewport[T] {
	vp := viewport.New(0, 0)

	return &DataViewport[T]{
		viewport: vp,
		data:     data,
	}
}

func (v *DataViewport[T]) Init() tea.Cmd {
	return func() tea.Msg {
		data, err := v.data.LoadData()
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: *data}
	}
}

func (v *DataViewport[T]) Resize(size Size) {
	v.viewport.Width = size.Width
	v.viewport.Height = size.Height
	v.size = size
}

func (v *DataViewport[T]) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			v.viewport.ScrollDown(1)
		case "k", "up":
			v.viewport.ScrollUp(1)
		case "g":
			v.viewport.GotoTop()
		case "G":
			v.viewport.GotoBottom()
		case "ctrl+d":
			v.viewport.ScrollDown(v.viewport.Height / 2)
		case "ctrl+u":
			v.viewport.ScrollUp(v.viewport.Height / 2)
		}
	case dataLoadedMsg:
		if data, ok := msg.data.(T); ok {
			v.elem = data
			v.content = v.data.ResolveData(data)
			if v.content == "" {
				v.content = "No data available"
			}
			v.viewport.SetContent(v.content)
		}
	}

	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v *DataViewport[T]) View() string {
	return v.viewport.View()
}

func (v *DataViewport[T]) KeyBindings() KeyBindings {
	return v.data.KeyBindings(v.elem)
}

func (v *DataViewport[T]) Focus() {
}

func (v *DataViewport[T]) Blur() {
}

type ViewportData[T any] interface {
	LoadData() (*T, error)
	ResolveData(data T) string
	KeyBindings(elem T) KeyBindings
}
