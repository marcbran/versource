package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http"
	"github.com/spf13/cobra"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Manage components",
	Long:  `Manage components`,
}

var componentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new component",
	Long:  `Create a new component with source, version, and variables`,
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

		req := internal.CreateComponentRequest{
			Source:    source,
			Version:   version,
			Changeset: changeset,
			Variables: variables,
		}

		component, err := client.CreateComponent(cmd.Context(), req)
		if err != nil {
			return err
		}

		fmt.Printf("Component created successfully with ID: %d\n", component.ID)
		return nil
	},
}

func init() {
	componentCreateCmd.Flags().String("source", "", "Component source")
	componentCreateCmd.Flags().String("version", "", "Component version")
	componentCreateCmd.Flags().String("changeset", "", "Component changeset")
	componentCreateCmd.Flags().String("variables", "{}", "Component variables as JSON string")
	componentCreateCmd.MarkFlagRequired("source")
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

		req := internal.UpdateComponentRequest{
			ComponentID: uint(componentID),
			Changeset:   changeset,
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

		component, err := client.UpdateComponent(cmd.Context(), req)
		if err != nil {
			return err
		}

		fmt.Printf("Component updated successfully with ID: %d\n", component.ID)
		return nil
	},
}

func init() {
	componentUpdateCmd.Flags().String("changeset", "", "Changeset name")
	componentUpdateCmd.Flags().String("source", "", "Component source")
	componentUpdateCmd.Flags().String("version", "", "Component version")
	componentUpdateCmd.Flags().String("variables", "", "Component variables as JSON string")
	componentUpdateCmd.MarkFlagRequired("changeset")
}
