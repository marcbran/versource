package cmd

import (
	"github.com/marcbran/versource/internal/database"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Start the HTTP server",
	Long:  `Start the HTTP server to handle API requests`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		return database.Migrate(cmd.Context(), config.Database)
	},
}
