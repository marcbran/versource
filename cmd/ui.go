package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/apply"
	"github.com/marcbran/versource/internal/tui/changeset"
	"github.com/marcbran/versource/internal/tui/component"
	"github.com/marcbran/versource/internal/tui/module"
	"github.com/marcbran/versource/internal/tui/plan"
	"github.com/marcbran/versource/internal/tui/platform"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Start the terminal user interface",
	Long:  `Start an interactive terminal user interface for managing versource`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			configDir = filepath.Join(homeDir, ".config")
		}

		logDir := filepath.Join(configDir, "versource")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		logFile := filepath.Join(logDir, "tui.log")
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		defer file.Close()

		log.SetOutput(file)
		log.SetLevel(log.InfoLevel)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		router := platform.NewRouter().
			Register("modules", module.NewTable(client)).
			Register("modules/{moduleID}", module.NewDetail(client)).
			Register("modules/{moduleID}/moduleversions", module.NewVersionsForModuleTable(client)).
			Register("moduleversions", module.NewVersionsTable(client)).
			Register("moduleversions/{moduleVersionID}", module.NewVersionDetail(client)).
			Register("components", component.NewTable(client)).
			Register("plans", plan.NewTable(client)).
			Register("plans/{planID}/logs", plan.NewLogs(client)).
			Register("applies", apply.NewTable(client)).
			Register("changesets", changeset.NewTable(client)).
			Register("changesets/{changesetName}/components", component.NewChangesetTable(client)).
			Register("changesets/{changesetName}/components/diffs", component.NewChangesetDiffTable(client)).
			Register("changesets/{changesetName}/plans", plan.NewChangesetTable(client))

		app := platform.NewCommandable(router, client)

		p := tea.NewProgram(app, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
}
