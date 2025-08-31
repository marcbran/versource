package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui"
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Start the terminal user interface",
	Long:  `Start an interactive terminal user interface for managing versource`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)
		app := tui.NewApp(client)

		p := tea.NewProgram(app, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run TUI: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
}
