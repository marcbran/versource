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
	return e.openEditor()
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
	case editorOpenedMsg:
		return e, e.readFile()
	case fileReadMsg:
		return e, e.processFile(m.content)
	case fileCancelledMsg:
		return e, func() tea.Msg { return goBackRequestedMsg{} }
	case fileProcessedMsg:
		if m.success {
			if data, ok := m.data.(T); ok {
				return e, e.saveData(data)
			} else {
				return e, func() tea.Msg { return errorMsg{err: fmt.Errorf("type assertion failed")} }
			}
		} else {
			e.errorMsg = m.err
			e.showError = true
			return e, nil
		}
	case dataSavedMsg:
		if m.success {
			return e, e.navigateToSuccess(m.redirectURL)
		} else {
			e.errorMsg = m.err
			e.showError = true
			return e, nil
		}
	case tea.KeyMsg:
		if e.showError {
			switch m.String() {
			case "enter":
				e.showError = false
				return e, e.openEditor()
			case "esc":
				return e, func() tea.Msg { return goBackRequestedMsg{} }
			}
		} else {
			switch m.String() {
			case "enter":
				return e, e.openEditor()
			case "esc":
				return e, func() tea.Msg { return goBackRequestedMsg{} }
			}
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

func (e *Editor[T]) openEditor() tea.Cmd {
	tempFile, err := e.createTempFile()
	if err != nil {
		return func() tea.Msg { return errorMsg{err: err} }
	}
	e.tempFile = tempFile

	contentToWrite := e.currentContent
	if contentToWrite == "" {
		initialValue := e.data.GetInitialValue()
		yamlBytes, err := yaml.Marshal(initialValue)
		if err != nil {
			return func() tea.Msg { return errorMsg{err: err} }
		}
		contentToWrite = string(yamlBytes)
	}

	err = os.WriteFile(tempFile, []byte(contentToWrite), 0644)
	if err != nil {
		return func() tea.Msg { return errorMsg{err: err} }
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, tempFile)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return errorMsg{err: err}
		}
		return editorOpenedMsg{}
	})
}

func (e *Editor[T]) readFile() tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(e.tempFile)
		if err != nil {
			return errorMsg{err: err}
		}

		e.currentContent = string(content)
		e.cleanup()

		return fileReadMsg{content: e.currentContent}
	}
}

func (e *Editor[T]) processFile(content string) tea.Cmd {
	return func() tea.Msg {
		initialValue := e.data.GetInitialValue()
		initialYaml, err := yaml.Marshal(initialValue)
		if err != nil {
			return fileProcessedMsg{success: false, err: fmt.Sprintf("Failed to marshal initial value: %v", err)}
		}

		if content == string(initialYaml) || strings.TrimSpace(content) == "" {
			return fileCancelledMsg{}
		}

		var data T
		err = yaml.Unmarshal([]byte(content), &data)
		if err != nil {
			return fileProcessedMsg{success: false, err: fmt.Sprintf("Invalid YAML: %v", err)}
		}

		return fileProcessedMsg{success: true, data: data}
	}
}

func (e *Editor[T]) saveData(data T) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		redirectURL, err := e.data.SaveData(ctx, data)
		if err != nil {
			return dataSavedMsg{success: false, err: fmt.Sprintf("Save failed: %v", err)}
		}

		return dataSavedMsg{success: true, redirectURL: redirectURL}
	}
}

func (e *Editor[T]) navigateToSuccess(redirectURL string) tea.Cmd {
	return func() tea.Msg {
		return openPageRequestedMsg{path: redirectURL}
	}
}

func (e *Editor[T]) createTempFile() (string, error) {
	tmpDir := os.TempDir()
	file, err := os.CreateTemp(tmpDir, "versource-editor-*.yaml")
	if err != nil {
		return "", err
	}
	defer file.Close()
	return file.Name(), nil
}

func (e *Editor[T]) cleanup() {
	if e.tempFile != "" {
		os.Remove(e.tempFile)
	}
}

func (e *Editor[T]) renderError() string {
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

func (e *Editor[T]) renderReady() string {
	readyText := "Press Enter to open editor, Esc to cancel"

	centeredContent := lipgloss.NewStyle().
		Width(e.size.Width).
		Height(e.size.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(readyText)

	return centeredContent
}

type editorOpenedMsg struct{}
type fileReadMsg struct {
	content string
}
type fileCancelledMsg struct{}
type fileProcessedMsg struct {
	success bool
	data    any
	err     string
}
type dataSavedMsg struct {
	success     bool
	redirectURL string
	err         string
}
