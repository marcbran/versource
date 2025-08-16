package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

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
	Long:  `Create a new module with source, version, and variables`,
	RunE: func(cmd *cobra.Command, args []string) error {
		source, err := cmd.Flags().GetString("source")
		if err != nil {
			return fmt.Errorf("failed to get source flag: %w", err)
		}

		version, err := cmd.Flags().GetString("version")
		if err != nil {
			return fmt.Errorf("failed to get version flag: %w", err)
		}

		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return fmt.Errorf("failed to get changeset flag: %w", err)
		}

		variablesStr, err := cmd.Flags().GetString("variables")
		if err != nil {
			return fmt.Errorf("failed to get variables flag: %w", err)
		}

		if source == "" {
			return fmt.Errorf("source is required")
		}
		if changeset == "" {
			return fmt.Errorf("changeset is required")
		}

		var variables map[string]any
		if variablesStr != "" {
			err := json.Unmarshal([]byte(variablesStr), &variables)
			if err != nil {
				return fmt.Errorf("invalid variables JSON: %w", err)
			}
		}

		config, err := LoadConfig()
		if err != nil {
			return err
		}

		client := http.NewClient(config)

		req := internal.CreateModuleRequest{
			Source:    source,
			Version:   version,
			Changeset: changeset,
			Variables: variables,
		}

		module, err := client.CreateModule(cmd.Context(), req)
		if err != nil {
			return err
		}

		fmt.Printf("Module created successfully with ID: %d\n", module.ID)
		return nil
	},
}

func init() {
	moduleCreateCmd.Flags().String("source", "", "Module source")
	moduleCreateCmd.Flags().String("version", "", "Module version")
	moduleCreateCmd.Flags().String("changeset", "", "Module changeset")
	moduleCreateCmd.Flags().String("variables", "{}", "Module variables as JSON string")
	moduleCreateCmd.MarkFlagRequired("source")
	moduleCreateCmd.MarkFlagRequired("changeset")

	moduleCmd.AddCommand(moduleCreateCmd)
	moduleCmd.AddCommand(moduleUpdateCmd)
}

var moduleUpdateCmd = &cobra.Command{
	Use:   "update [module-id]",
	Short: "Update a module",
	Long:  `Update a module by patching individual fields`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		moduleIDStr := args[0]
		moduleID, err := strconv.ParseUint(moduleIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid module ID: %w", err)
		}

		source, err := cmd.Flags().GetString("source")
		if err != nil {
			return fmt.Errorf("failed to get source flag: %w", err)
		}

		version, err := cmd.Flags().GetString("version")
		if err != nil {
			return fmt.Errorf("failed to get version flag: %w", err)
		}

		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return fmt.Errorf("failed to get changeset flag: %w", err)
		}

		variablesStr, err := cmd.Flags().GetString("variables")
		if err != nil {
			return fmt.Errorf("failed to get variables flag: %w", err)
		}

		if changeset == "" {
			return fmt.Errorf("changeset is required")
		}
		if source == "" && version == "" && variablesStr == "" {
			return fmt.Errorf("at least one field must be provided to update")
		}

		config, err := LoadConfig()
		if err != nil {
			return err
		}

		client := http.NewClient(config)

		req := internal.UpdateModuleRequest{
			ModuleID:  uint(moduleID),
			Changeset: changeset,
		}

		if source != "" {
			req.Source = &source
		}
		if version != "" {
			req.Version = &version
		}
		if variablesStr != "" {
			var variables map[string]any
			err := json.Unmarshal([]byte(variablesStr), &variables)
			if err != nil {
				return fmt.Errorf("invalid variables JSON: %w", err)
			}
			req.Variables = &variables
		}

		module, err := client.UpdateModule(cmd.Context(), req)
		if err != nil {
			return err
		}

		fmt.Printf("Module updated successfully with ID: %d\n", module.ID)
		return nil
	},
}

func init() {
	moduleUpdateCmd.Flags().String("changeset", "", "Changeset name")
	moduleUpdateCmd.Flags().String("source", "", "Module source")
	moduleUpdateCmd.Flags().String("version", "", "Module version")
	moduleUpdateCmd.Flags().String("variables", "", "Module variables as JSON string")
	moduleUpdateCmd.MarkFlagRequired("changeset")
}
