package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LoadView struct {
	spinner spinner.Model
	size    Size
}

func NewLoadView() LoadView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))

	return LoadView{
		spinner: s,
	}
}

func (l LoadView) Init() tea.Cmd {
	return l.spinner.Tick
}

func (l *LoadView) Resize(size Size) {
	l.size = size
}

func (l LoadView) Update(msg tea.Msg) (LoadView, tea.Cmd) {
	var cmd tea.Cmd
	l.spinner, cmd = l.spinner.Update(msg)
	return l, cmd
}

func (l LoadView) View() string {
	spinnerView := l.spinner.View()
	loadingText := "Loading..."

	content := lipgloss.JoinHorizontal(lipgloss.Center, spinnerView, " ", loadingText)

	return lipgloss.NewStyle().
		Width(l.size.Width).
		Height(l.size.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}
