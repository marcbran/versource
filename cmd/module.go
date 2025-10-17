package cmd

import (
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/module"
	"github.com/marcbran/versource/pkg/versource"
	"github.com/spf13/cobra"
)

var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Manage modules",
	Long:  `Manage modules`,
}

var moduleGetCmd = &cobra.Command{
	Use:   "get [module-id]",
	Short: "Get a specific module",
	Long:  `Get details for a specific module by ID`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.New(config)
		detailData := module.NewDetailData(httpClient, args[0])
		return renderViewportViewData(detailData)
	},
}

var moduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all modules",
	Long:  `List all modules in the system`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.New(config)
		tableData := module.NewTableData(httpClient)
		return renderTableData(tableData)
	},
}

var moduleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new module",
	Long:  `Create a new module with source and version`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return fmt.Errorf("failed to get name flag: %w", err)
		}

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

		if name == "" {
			return fmt.Errorf("name is required")
		}

		if source == "" {
			return fmt.Errorf("source is required")
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.New(config)

		req := versource.CreateModuleRequest{
			Name:         name,
			Source:       source,
			Version:      version,
			ExecutorType: executorType,
		}

		module, err := client.CreateModule(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(module, "Module created successfully with ID: %d, Version ID: %d\n", module.Module.ID, module.LatestVersion.ID)
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

		client := client.New(config)

		req := versource.UpdateModuleRequest{
			ModuleID: uint(moduleID),
			Version:  version,
		}

		module, err := client.UpdateModule(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(module, "Module updated successfully with Version ID: %d\n", module.LatestVersion.ID)
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

		client := client.New(config)

		_, err = client.DeleteModule(cmd.Context(), versource.DeleteModuleRequest{ModuleID: uint(moduleID)})
		if err != nil {
			return err
		}

		return formatOutput(nil, "Module deleted successfully\n")
	},
}

var moduleVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage module versions",
	Long:  `Manage module versions`,
}

var moduleVersionGetCmd = &cobra.Command{
	Use:   "get [module-version-id]",
	Short: "Get a specific module version",
	Long:  `Get details for a specific module version by ID`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.New(config)
		detailData := module.NewVersionDetailData(httpClient, args[0])
		return renderViewportViewData(detailData)
	},
}

var moduleVersionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List module versions",
	Long:  `List all module versions or versions for a specific module`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		var moduleID *string
		moduleIDStr, err := cmd.Flags().GetString("module-id")
		if err == nil && moduleIDStr != "" {
			moduleID = &moduleIDStr
		}

		httpClient := client.New(config)
		tableData := module.NewVersionsTableData(httpClient, moduleID)
		return renderTableData(tableData)
	},
}

func init() {
	moduleCreateCmd.Flags().String("name", "", "Module name")
	moduleCreateCmd.Flags().String("source", "", "Module source")
	moduleCreateCmd.Flags().String("version", "", "Module version (optional for some source types)")
	moduleCreateCmd.Flags().String("executor", "terraform-jsonnet", "Executor type (terraform-module, terraform-jsonnet)")
	_ = moduleCreateCmd.MarkFlagRequired("name")
	_ = moduleCreateCmd.MarkFlagRequired("source")

	moduleUpdateCmd.Flags().String("version", "", "Module version")
	_ = moduleUpdateCmd.MarkFlagRequired("version")

	moduleVersionListCmd.Flags().String("module-id", "", "Filter versions by module ID")

	moduleVersionCmd.AddCommand(moduleVersionGetCmd)
	moduleVersionCmd.AddCommand(moduleVersionListCmd)

	moduleCmd.AddCommand(moduleGetCmd)
	moduleCmd.AddCommand(moduleListCmd)
	moduleCmd.AddCommand(moduleCreateCmd)
	moduleCmd.AddCommand(moduleUpdateCmd)
	moduleCmd.AddCommand(moduleDeleteCmd)
	moduleCmd.AddCommand(moduleVersionCmd)
}
