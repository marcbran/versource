package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type DataViewport struct {
	viewport viewport.Model
	content  string
	size     Size
	data     ViewportData
}

func NewDataViewport(data ViewportData) *DataViewport {
	vp := viewport.New(0, 0)

	return &DataViewport{
		viewport: vp,
		data:     data,
	}
}

func (v *DataViewport) Init() tea.Cmd {
	return v.data.LoadData()
}

func (v *DataViewport) Resize(size Size) {
	v.viewport.Width = size.Width
	v.viewport.Height = size.Height
	v.size = size
}

func (v *DataViewport) Update(msg tea.Msg) (Page, tea.Cmd) {
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
		v.content = v.data.ResolveData(msg.data)
		if v.content == "" {
			v.content = "No data available"
		}
		v.viewport.SetContent(v.content)
	}

	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v *DataViewport) View() string {
	return v.viewport.View()
}

func (v *DataViewport) KeyBindings() KeyBindings {
	return v.data.KeyBindings(nil)
}

func (v *DataViewport) Focus() {
}

func (v *DataViewport) Blur() {
}

type ViewportData interface {
	LoadData() tea.Cmd
	ResolveData(data any) string
	KeyBindings(elem any) KeyBindings
}
