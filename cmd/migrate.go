package cmd

import (
	"github.com/marcbran/versource/internal/database/migrations"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run database migrations to set up the database schema`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		return migrations.Migrate(cmd.Context(), config.Database)
	},
}
