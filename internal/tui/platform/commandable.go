package platform

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcbran/versource/internal/http/client"
)

type Commandable struct {
	client *client.Client
	router *Router

	input     textinput.Model
	showInput bool

	size Size
}

func (c *Commandable) contentSize() Size {
	contentWidth := c.size.Width - 4
	contentHeight := c.size.Height - 2
	if c.showInput {
		contentHeight -= 3
	}
	return Size{
		Width:  contentWidth,
		Height: contentHeight,
	}
}

func NewCommandable(
	router *Router,
	client *client.Client,
) *Commandable {
	input := textinput.New()
	input.CharLimit = 100

	return &Commandable{
		router: router,
		client: client,
		input:  input,
	}
}

func (c *Commandable) Init() tea.Cmd {
	return c.router.Open("modules")
}

func (c *Commandable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return c, tea.Quit
		}
	}

	if c.showInput {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				c.input.Blur()
				c.showInput = false
				c.input.SetValue("")
				c.router.Focus()
				c.router.Resize(c.contentSize())
				return c, nil
			case "enter":
				command := c.input.Value()
				c.input.Blur()
				c.showInput = false
				c.input.SetValue("")
				c.router.Focus()
				c.router.Resize(c.contentSize())
				return c, c.router.ExecuteCommand(command)
			}
		}
		c.input, cmd = c.input.Update(msg)
		return c, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.size.Width = msg.Width
		c.size.Height = msg.Height
		c.input.Width = msg.Width - 7
		c.router.Resize(c.contentSize())
	case RouteOpenedMsg:
		c.input.Blur()
		c.router.Resize(c.contentSize())
		c.router.Focus()
	case tea.KeyMsg:
		switch msg.String() {
		case ":":
			c.showInput = true
			c.input.Focus()
			c.router.Blur()
			c.router.Resize(c.contentSize())
			return c, textinput.Blink
		}
	}

	c.router, cmd = c.router.Update(msg)
	return c, cmd
}

func (c *Commandable) View() string {
	content := c.router.View()

	if c.showInput {
		inputView := c.input.View()
		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1).
			Render(inputView)

		content = lipgloss.JoinVertical(lipgloss.Left, inputBox, content)
	}

	return content
}
