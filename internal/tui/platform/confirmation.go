package platform

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfirmationPage struct {
	message     string
	confirmText string
	cancelText  string
	data        ConfirmationData
	size        Size
	focused     bool
}

type ConfirmationData interface {
	GetConfirmationDialog() ConfirmationDialog
	OnConfirm(ctx context.Context) (string, error)
}

type ConfirmationDialog struct {
	Title       string
	Message     string
	ConfirmText string
	CancelText  string
}

func NewConfirmationPage(data ConfirmationData) *ConfirmationPage {
	dialog := data.GetConfirmationDialog()
	return &ConfirmationPage{
		message:     dialog.Message,
		confirmText: dialog.ConfirmText,
		cancelText:  dialog.CancelText,
		data:        data,
	}
}

func (c *ConfirmationPage) Init() tea.Cmd {
	return nil
}

func (c *ConfirmationPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return c, c.onConfirm()
		case "esc":
			return c, func() tea.Msg { return goBackRequestedMsg{} }
		}
	}
	return c, nil
}

func (c *ConfirmationPage) onConfirm() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		redirectPath, err := c.data.OnConfirm(ctx)
		if err != nil {
			return errorMsg{err: err}
		}
		return openPageRequestedMsg{path: redirectPath}
	}
}

func (c *ConfirmationPage) View() string {
	helpText := fmt.Sprintf("Press Enter to %s, Esc to %s", c.confirmText, c.cancelText)

	content := lipgloss.JoinVertical(lipgloss.Center, c.message, "", helpText)

	centeredContent := lipgloss.NewStyle().
		Width(c.size.Width).
		Height(c.size.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)

	return centeredContent
}

func (c *ConfirmationPage) Resize(size Size) {
	c.size = size
}

func (c *ConfirmationPage) Focus() {
	c.focused = true
}

func (c *ConfirmationPage) Blur() {
	c.focused = false
}

func (c *ConfirmationPage) KeyBindings() KeyBindings {
	return KeyBindings{}
}

func (c *ConfirmationPage) ExcludeFromHistory() bool {
	return true
}
