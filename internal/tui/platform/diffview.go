package platform

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Diff struct {
	Left  string
	Right string
}

type DiffView[T any] struct {
	leftViewport  viewport.Model
	rightViewport viewport.Model
	elem          T
	leftContent   string
	rightContent  string
	showDiff      bool

	size Size
	data DiffData[T]
}

func NewDiffView[T any](data DiffData[T]) *DiffView[T] {
	leftVp := viewport.New(0, 0)
	rightVp := viewport.New(0, 0)

	return &DiffView[T]{
		leftViewport:  leftVp,
		rightViewport: rightVp,
		data:          data,
		showDiff:      true,
	}
}

func (d *DiffView[T]) Init() tea.Cmd {
	return func() tea.Msg {
		data, err := d.data.LoadData()
		if err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: *data}
	}
}

func (d *DiffView[T]) Resize(size Size) {
	width := size.Width / 2
	height := size.Height

	d.leftViewport.Width = width
	d.leftViewport.Height = height
	d.rightViewport.Width = width
	d.rightViewport.Height = height
	d.size = size
}

func (d *DiffView[T]) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd
	var leftCmd, rightCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			d.leftViewport.ScrollDown(1)
			d.rightViewport.ScrollDown(1)
		case "k", "up":
			d.leftViewport.ScrollUp(1)
			d.rightViewport.ScrollUp(1)
		case "g":
			d.leftViewport.GotoTop()
			d.rightViewport.GotoTop()
		case "G":
			d.leftViewport.GotoBottom()
			d.rightViewport.GotoBottom()
		case "ctrl+d":
			scrollAmount := d.leftViewport.Height / 2
			d.leftViewport.ScrollDown(scrollAmount)
			d.rightViewport.ScrollDown(scrollAmount)
		case "ctrl+u":
			scrollAmount := d.leftViewport.Height / 2
			d.leftViewport.ScrollUp(scrollAmount)
			d.rightViewport.ScrollUp(scrollAmount)
		case "d":
			d.showDiff = !d.showDiff
		}
	case dataLoadedMsg:
		if data, ok := msg.data.(T); ok {
			d.elem = data
			diff := d.data.ResolveData(data)
			d.leftContent = diff.Left
			d.rightContent = diff.Right
			d.leftViewport.SetContent(d.leftContent)
			d.rightViewport.SetContent(d.rightContent)
		}
	}

	d.leftViewport, leftCmd = d.leftViewport.Update(msg)
	d.rightViewport, rightCmd = d.rightViewport.Update(msg)

	return d, tea.Batch(leftCmd, rightCmd, cmd)
}

func (d *DiffView[T]) View() string {
	if d.leftContent == "" && d.rightContent == "" {
		return "No data available"
	}

	width := d.size.Width / 2

	var leftContent, rightContent string
	if d.showDiff {
		leftDiff, rightDiff := simpleDiff(d.leftContent, d.rightContent, width)
		leftContent = leftDiff
		rightContent = rightDiff
	} else {
		leftContent = d.leftContent
		rightContent = d.rightContent
	}

	d.leftViewport.SetContent(leftContent)
	d.rightViewport.SetContent(rightContent)

	leftView := lipgloss.NewStyle().
		Width(width).
		Height(d.size.Height).
		Render(d.leftViewport.View())

	rightView := lipgloss.NewStyle().
		Width(width).
		Height(d.size.Height).
		Render(d.rightViewport.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
}

func (d *DiffView[T]) KeyBindings() KeyBindings {
	return d.data.KeyBindings(d.elem)
}

func (d *DiffView[T]) ExcludeFromHistory() bool {
	return false
}

func (d *DiffView[T]) Focus() {
}

func (d *DiffView[T]) Blur() {
}

type DiffData[T any] interface {
	LoadData() (*T, error)
	ResolveData(data T) Diff
	KeyBindings(elem T) KeyBindings
}

func simpleDiff(left, right string, width int) (string, string) {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")

	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	var leftResult, rightResult []string

	for i := 0; i < maxLines; i++ {
		leftLine := ""
		rightLine := ""

		if i < len(leftLines) {
			leftLine = leftLines[i]
		}
		if i < len(rightLines) {
			rightLine = rightLines[i]
		}

		leftStyle := lipgloss.NewStyle()
		rightStyle := lipgloss.NewStyle()

		if leftLine != rightLine {
			if leftLine == "" {
				leftStyle = leftStyle.Foreground(lipgloss.Color("8")).Background(lipgloss.Color("52"))
			} else {
				leftStyle = leftStyle.Foreground(lipgloss.Color("9")).Background(lipgloss.Color("52"))
			}
			if rightLine == "" {
				rightStyle = rightStyle.Foreground(lipgloss.Color("8")).Background(lipgloss.Color("22"))
			} else {
				rightStyle = rightStyle.Foreground(lipgloss.Color("10")).Background(lipgloss.Color("22"))
			}
		} else if leftLine != "" {
			leftStyle = leftStyle.Foreground(lipgloss.Color("7"))
			rightStyle = rightStyle.Foreground(lipgloss.Color("7"))
		}

		if leftLine != rightLine {
			paddedLeftLine := leftLine
			paddedRightLine := rightLine

			if len(leftLine) < width {
				paddedLeftLine = leftLine + strings.Repeat(" ", width-len(leftLine))
			}
			if len(rightLine) < width {
				paddedRightLine = rightLine + strings.Repeat(" ", width-len(rightLine))
			}

			leftResult = append(leftResult, leftStyle.Render(paddedLeftLine))
			rightResult = append(rightResult, rightStyle.Render(paddedRightLine))
		} else {
			leftResult = append(leftResult, leftStyle.Render(leftLine))
			rightResult = append(rightResult, rightStyle.Render(rightLine))
		}
	}

	return strings.Join(leftResult, "\n"), strings.Join(rightResult, "\n")
}
