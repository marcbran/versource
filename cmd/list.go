package cmd

import (
	"encoding/json"
	"github.com/marcbran/versource/internal"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all resources of a certain view",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir := os.Getenv("VERSOURCE_CONFIG_HOME")
		if configDir == "" {
			configDir = path.Join(os.Getenv("XDG_CONFIG_HOME"), "versource")
		}
		dataDir := os.Getenv("VERSOURCE_DATA_HOME")
		if dataDir == "" {
			dataDir = path.Join(os.Getenv("XDG_DATA_HOME"), "versource")
		}
		logFile, err := logFile(dataDir)
		if err != nil {
			return err
		}
		defer logFile.Close()

		var query string
		if len(args) > 0 {
			query = args[0]
		}
		resource := os.Getenv("resource")
		list, err := internal.List(cmd.Context(), configDir, dataDir, query, resource)
		if err != nil {
			return err
		}

		b, err := json.Marshal(list)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(b)
		if err != nil {
			return err
		}
		return nil
	},
}
