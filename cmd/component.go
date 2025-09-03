package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/spf13/cobra"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Manage components",
	Long:  `Manage components`,
}

var componentCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new component",
	Long:  `Create a new component with name, module ID and variables (uses latest module version)`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if name == "" {
			return fmt.Errorf("name is required")
		}

		moduleIDStr, err := cmd.Flags().GetString("module-id")
		if err != nil {
			return fmt.Errorf("failed to get module-id flag: %w", err)
		}

		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return fmt.Errorf("failed to get changeset flag: %w", err)
		}

		variablesStr, err := cmd.Flags().GetString("variables")
		if err != nil {
			return fmt.Errorf("failed to get variables flag: %w", err)
		}

		if moduleIDStr == "" {
			return fmt.Errorf("module-id is required")
		}
		if changeset == "" {
			return fmt.Errorf("changeset is required")
		}

		moduleID, err := strconv.ParseUint(moduleIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid module ID: %w", err)
		}

		var variables map[string]any
		if variablesStr != "" {
			err := json.Unmarshal([]byte(variablesStr), &variables)
			if err != nil {
				return fmt.Errorf("invalid variables JSON: %w", err)
			}
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.CreateComponentRequest{
			ModuleID:  uint(moduleID),
			Changeset: changeset,
			Name:      name,
			Variables: variables,
		}

		component, err := client.CreateComponent(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(component, "Component created successfully with ID: %d\n", component.ID)
	},
}

func init() {
	componentCreateCmd.Flags().String("module-id", "", "Module ID (will use latest version)")
	componentCreateCmd.Flags().String("changeset", "", "Component changeset")
	componentCreateCmd.Flags().String("variables", "{}", "Component variables as JSON string")
	componentCreateCmd.MarkFlagRequired("module-id")
	componentCreateCmd.MarkFlagRequired("changeset")

	componentCmd.AddCommand(componentCreateCmd)
	componentCmd.AddCommand(componentUpdateCmd)
}

var componentUpdateCmd = &cobra.Command{
	Use:   "update [component-id]",
	Short: "Update a component",
	Long:  `Update a component by patching individual fields`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		componentIDStr := args[0]
		componentID, err := strconv.ParseUint(componentIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid component ID: %w", err)
		}

		moduleIDStr, err := cmd.Flags().GetString("module-id")
		if err != nil {
			return fmt.Errorf("failed to get module-id flag: %w", err)
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
		if moduleIDStr == "" && variablesStr == "" {
			return fmt.Errorf("at least one field must be provided to update")
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.UpdateComponentRequest{
			ComponentID: uint(componentID),
			Changeset:   changeset,
		}

		if moduleIDStr != "" {
			moduleID, err := strconv.ParseUint(moduleIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid module ID: %w", err)
			}
			moduleIDUint := uint(moduleID)
			req.ModuleID = &moduleIDUint
		}
		if variablesStr != "" {
			var variables map[string]any
			err := json.Unmarshal([]byte(variablesStr), &variables)
			if err != nil {
				return fmt.Errorf("invalid variables JSON: %w", err)
			}
			req.Variables = &variables
		}

		component, err := client.UpdateComponent(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(component, "Component updated successfully with ID: %d\n", component.ID)
	},
}

func init() {
	componentUpdateCmd.Flags().String("changeset", "", "Changeset name")
	componentUpdateCmd.Flags().String("module-id", "", "Module ID (will use latest version)")
	componentUpdateCmd.Flags().String("variables", "", "Component variables as JSON string")
	componentUpdateCmd.MarkFlagRequired("changeset")
}
