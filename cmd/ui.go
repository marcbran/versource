package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui"
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

		err := os.MkdirAll(logDir, 0o755)
		if err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		logFile := filepath.Join(logDir, "tui.log")
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
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

		client := client.New(config)

		return tui.RunApp(client)
	},
}
