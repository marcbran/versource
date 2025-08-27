package cmd

import (
	"fmt"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http"
	"github.com/spf13/cobra"
)

var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Manage modules",
	Long:  `Manage modules`,
}

var moduleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new module",
	Long:  `Create a new module with source and version`,
	RunE: func(cmd *cobra.Command, args []string) error {
		source, err := cmd.Flags().GetString("source")
		if err != nil {
			return fmt.Errorf("failed to get source flag: %w", err)
		}

		version, err := cmd.Flags().GetString("version")
		if err != nil {
			return fmt.Errorf("failed to get version flag: %w", err)
		}

		if source == "" {
			return fmt.Errorf("source is required")
		}

		config, err := LoadConfig()
		if err != nil {
			return err
		}

		client := http.NewClient(config)

		req := internal.CreateModuleRequest{
			Source:  source,
			Version: version,
		}

		module, err := client.CreateModule(cmd.Context(), req)
		if err != nil {
			return err
		}

		fmt.Printf("Module created successfully with ID: %d, Version ID: %d\n", module.ID, module.VersionID)
		return nil
	},
}

func init() {
	moduleCreateCmd.Flags().String("source", "", "Module source")
	moduleCreateCmd.Flags().String("version", "", "Module version (optional for some source types)")
	moduleCreateCmd.MarkFlagRequired("source")

	moduleCmd.AddCommand(moduleCreateCmd)
}
