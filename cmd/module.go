package cmd

import (
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
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

		executorType, err := cmd.Flags().GetString("executor")
		if err != nil {
			return fmt.Errorf("failed to get executor flag: %w", err)
		}

		if source == "" {
			return fmt.Errorf("source is required")
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.CreateModuleRequest{
			Source:       source,
			Version:      version,
			ExecutorType: executorType,
		}

		module, err := client.CreateModule(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(module, "Module created successfully with ID: %d, Version ID: %d\n", module.ID, module.VersionID)
	},
}

var moduleUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a module with a new version",
	Long:  `Update a module by creating a new version`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("module ID is required")
		}

		moduleIDStr := args[0]
		moduleID, err := strconv.ParseUint(moduleIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid module ID: %w", err)
		}

		version, err := cmd.Flags().GetString("version")
		if err != nil {
			return fmt.Errorf("failed to get version flag: %w", err)
		}

		if version == "" {
			return fmt.Errorf("version is required")
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.UpdateModuleRequest{
			Version: version,
		}

		module, err := client.UpdateModule(cmd.Context(), uint(moduleID), req)
		if err != nil {
			return err
		}

		return formatOutput(module, "Module updated successfully with Version ID: %d\n", module.VersionID)
	},
}

var moduleDeleteCmd = &cobra.Command{
	Use:   "delete [module-id]",
	Short: "Delete a module",
	Long:  `Delete a module (only if not referenced by any components)`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		moduleIDStr := args[0]
		moduleID, err := strconv.ParseUint(moduleIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid module ID: %w", err)
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		_, err = client.DeleteModule(cmd.Context(), uint(moduleID))
		if err != nil {
			return err
		}

		return formatOutput(nil, "Module deleted successfully\n")
	},
}

func init() {
	moduleCreateCmd.Flags().String("source", "", "Module source")
	moduleCreateCmd.Flags().String("version", "", "Module version (optional for some source types)")
	moduleCreateCmd.Flags().String("executor", "terraform-module", "Executor type (terraform-module, terraform-jsonnet)")
	moduleCreateCmd.MarkFlagRequired("source")
	moduleCreateCmd.MarkFlagRequired("executor")

	moduleUpdateCmd.Flags().String("version", "", "Module version")
	moduleUpdateCmd.MarkFlagRequired("version")

	moduleCmd.AddCommand(moduleCreateCmd)
	moduleCmd.AddCommand(moduleUpdateCmd)
	moduleCmd.AddCommand(moduleDeleteCmd)
}
