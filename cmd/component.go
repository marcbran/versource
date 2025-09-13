package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/component"
	"github.com/spf13/cobra"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Manage components",
	Long:  `Manage components`,
}

var componentGetCmd = &cobra.Command{
	Use:   "get [component-id]",
	Short: "Get a specific component",
	Long:  `Get details for a specific component by ID`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.NewClient(config)
		detailData := component.NewDetailData(httpClient, args[0], nil)
		return renderViewpointData(detailData)
	},
}

var componentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all components",
	Long:  `List all components in the system`,
	RunE: func(cmd *cobra.Command, args []string) error {
		moduleIDStr, err := cmd.Flags().GetString("module-id")
		if err != nil {
			return fmt.Errorf("failed to get module-id flag: %w", err)
		}

		moduleVersionIDStr, err := cmd.Flags().GetString("module-version-id")
		if err != nil {
			return fmt.Errorf("failed to get module-version-id flag: %w", err)
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.NewClient(config)
		tableData := component.NewTableData(httpClient, moduleIDStr, moduleVersionIDStr)
		return renderTableData(tableData)
	},
}

var componentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new component",
	Long:  `Create a new component with name, module ID and variables (uses latest module version)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return fmt.Errorf("failed to get name flag: %w", err)
		}
		moduleIDStr, err := cmd.Flags().GetString("module-id")
		if err != nil {
			return fmt.Errorf("failed to get module-id flag: %w", err)
		}

		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return fmt.Errorf("failed to get changeset flag: %w", err)
		}

		variableMap, err := cmd.Flags().GetStringToString("variable")
		if err != nil {
			return fmt.Errorf("failed to get variable flags: %w", err)
		}

		if name == "" {
			return fmt.Errorf("name is required")
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

		variables, err := parseVariables(variableMap)
		if err != nil {
			return err
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

		variableMap, err := cmd.Flags().GetStringToString("variable")
		if err != nil {
			return fmt.Errorf("failed to get variable flags: %w", err)
		}

		if changeset == "" {
			return fmt.Errorf("changeset is required")
		}
		if moduleIDStr == "" && len(variableMap) == 0 {
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
		if len(variableMap) > 0 {
			variables, err := parseVariables(variableMap)
			if err != nil {
				return err
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

var componentDeleteCmd = &cobra.Command{
	Use:   "delete [component-id]",
	Short: "Delete a component",
	Long:  `Delete a component by setting its status to Deleted and resetting to merge base state`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		componentIDStr := args[0]
		componentID, err := strconv.ParseUint(componentIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid component ID: %w", err)
		}

		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return fmt.Errorf("failed to get changeset flag: %w", err)
		}

		if changeset == "" {
			return fmt.Errorf("changeset is required")
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.DeleteComponentRequest{
			ComponentID: uint(componentID),
			Changeset:   changeset,
		}

		component, err := client.DeleteComponent(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(component, "Component deleted successfully with ID: %d\n", component.ID)
	},
}

func parseVariables(variableMap map[string]string) (map[string]any, error) {
	variables := make(map[string]any)

	for key, valueStr := range variableMap {
		var value any
		if valueStr == "" {
			value = ""
		} else if valueStr[0] == '{' || valueStr[0] == '[' {
			err := json.Unmarshal([]byte(valueStr), &value)
			if err != nil {
				return nil, fmt.Errorf("invalid JSON in variable '%s': %w", key, err)
			}
		} else if valueStr == "true" {
			value = true
		} else if valueStr == "false" {
			value = false
		} else if valueStr == "null" {
			value = nil
		} else if num, err := strconv.ParseFloat(valueStr, 64); err == nil {
			value = num
		} else {
			value = valueStr
		}

		variables[key] = value
	}

	return variables, nil
}

func init() {
	componentListCmd.Flags().String("module-id", "", "Filter components by module ID")
	componentListCmd.Flags().String("module-version-id", "", "Filter components by module version ID")

	componentCreateCmd.Flags().String("name", "", "Component name")
	componentCreateCmd.Flags().String("module-id", "", "Module ID (will use latest version)")
	componentCreateCmd.Flags().String("changeset", "", "Component changeset")
	componentCreateCmd.Flags().StringToString("variable", nil, "Component variable in key=value format (can be used multiple times)")
	componentCreateCmd.MarkFlagRequired("name")
	componentCreateCmd.MarkFlagRequired("module-id")
	componentCreateCmd.MarkFlagRequired("changeset")

	componentUpdateCmd.Flags().String("changeset", "", "Changeset name")
	componentUpdateCmd.Flags().String("module-id", "", "Module ID (will use latest version)")
	componentUpdateCmd.Flags().StringToString("variable", nil, "Component variable in key=value format (can be used multiple times)")
	componentUpdateCmd.MarkFlagRequired("changeset")

	componentDeleteCmd.Flags().String("changeset", "", "Changeset name")
	componentDeleteCmd.MarkFlagRequired("changeset")

	componentCmd.AddCommand(componentGetCmd)
	componentCmd.AddCommand(componentListCmd)
	componentCmd.AddCommand(componentCreateCmd)
	componentCmd.AddCommand(componentUpdateCmd)
	componentCmd.AddCommand(componentDeleteCmd)
}
