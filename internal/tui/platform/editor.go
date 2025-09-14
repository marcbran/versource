package platform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

type EditorData[T any] interface {
	GetInitialValue() T
	SaveData(ctx context.Context, data T) (string, error)
}

type Editor[T any] struct {
	data EditorData[T]

	tempFile       string
	initialContent string
	currentContent string
	errorMsg       string
	showError      bool

	size    Size
	focused bool
}

func NewEditor[T any](data EditorData[T]) *Editor[T] {
	return &Editor[T]{
		data: data,
	}
}

func (e *Editor[T]) Init() tea.Cmd {
	return e.initializeContent()
}

func (e *Editor[T]) Resize(size Size) {
	e.size = size
}

func (e *Editor[T]) Focus() {
	e.focused = true
}

func (e *Editor[T]) Blur() {
	e.focused = false
}

func (e *Editor[T]) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch m := msg.(type) {
	case contentInitializedMsg:
		e.initialContent = m.content
		e.currentContent = m.content
		return e, e.prepareFile()
	case filePreparedMsg:
		e.tempFile = m.tempFile
		return e, e.openEditor()
	case editorClosedMsg:
		return e, e.readFile()
	case fileReadMsg:
		e.currentContent = m.content
		return e, e.processFile()
	case fileCancelledMsg:
		return e, func() tea.Msg { return goBackRequestedMsg{} }
	case fileProcessedMsg[T]:
		return e, e.saveData(m.data)
	case dataSavedMsg:
		return e, e.navigateToSuccess(m.redirectURL)
	case editorErrorMsg:
		e.errorMsg = m.err
		e.showError = true
		return e, nil
	case tea.KeyMsg:
		switch m.String() {
		case "enter":
			e.showError = false
			return e, e.openEditor()
		case "esc":
			return e, func() tea.Msg { return goBackRequestedMsg{} }
		}
	}
	return e, nil
}

func (e *Editor[T]) View() string {
	if e.showError {
		return e.renderError()
	}
	return e.renderReady()
}

func (e *Editor[T]) KeyBindings() KeyBindings {
	return KeyBindings{}
}

func (e Editor[T]) initializeContent() tea.Cmd {
	return func() tea.Msg {
		if e.initialContent != "" {
			return contentInitializedMsg{content: e.initialContent}
		}
		initialValue := e.data.GetInitialValue()
		initialContent, err := yaml.Marshal(initialValue)
		if err != nil {
			return errorMsg{err: err}
		}
		return contentInitializedMsg{content: string(initialContent)}
	}
}

func (e Editor[T]) prepareFile() tea.Cmd {
	return func() tea.Msg {
		tempDir := os.TempDir()
		tempFile, err := os.CreateTemp(tempDir, "versource-editor-*.yaml")
		if err != nil {
			return errorMsg{err: err}
		}
		defer tempFile.Close()

		_, err = tempFile.Write([]byte(e.initialContent))
		if err != nil {
			return errorMsg{err: err}
		}
		return filePreparedMsg{tempFile: tempFile.Name()}
	}
}

func (e Editor[T]) openEditor() tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, e.tempFile)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return errorMsg{err: err}
		}
		return editorClosedMsg{}
	})
}

func (e Editor[T]) readFile() tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(e.tempFile)
		if err != nil {
			return errorMsg{err: err}
		}
		return fileReadMsg{content: string(content)}
	}
}

func (e Editor[T]) processFile() tea.Cmd {
	return func() tea.Msg {
		if e.currentContent == e.initialContent || strings.TrimSpace(e.currentContent) == "" {
			return fileCancelledMsg{}
		}
		var data T
		err := yaml.Unmarshal([]byte(e.currentContent), &data)
		if err != nil {
			return editorErrorMsg{err: fmt.Sprintf("Invalid YAML: %v", err)}
		}
		return fileProcessedMsg[T]{data: data}
	}
}

func (e Editor[T]) saveData(data T) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		redirectURL, err := e.data.SaveData(ctx, data)
		if err != nil {
			return editorErrorMsg{err: fmt.Sprintf("Save failed: %v", err)}
		}
		return dataSavedMsg{redirectURL: redirectURL}
	}
}

func (e Editor[T]) navigateToSuccess(redirectURL string) tea.Cmd {
	return func() tea.Msg {
		return openPageRequestedMsg{path: redirectURL}
	}
}

func (e Editor[T]) renderError() string {
	errorText := fmt.Sprintf("Error: %s", e.errorMsg)
	helpText := "Press Enter to return to editor, Esc to abort"

	content := lipgloss.JoinVertical(lipgloss.Center, errorText, "", helpText)

	centeredContent := lipgloss.NewStyle().
		Width(e.size.Width).
		Height(e.size.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)

	return centeredContent
}

func (e Editor[T]) renderReady() string {
	readyText := "Press Enter to open editor, Esc to cancel"

	centeredContent := lipgloss.NewStyle().
		Width(e.size.Width).
		Height(e.size.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(readyText)

	return centeredContent
}

type contentInitializedMsg struct {
	content string
}
type filePreparedMsg struct {
	tempFile string
}
type editorClosedMsg struct{}
type fileReadMsg struct {
	content string
}
type fileCancelledMsg struct{}
type fileProcessedMsg[T any] struct {
	data T
}
type dataSavedMsg struct {
	redirectURL string
}
type editorErrorMsg struct {
	err string
}
