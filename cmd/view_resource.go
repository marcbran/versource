package cmd

import (
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/spf13/cobra"
)

var viewResourceCmd = &cobra.Command{
	Use:   "view-resource",
	Short: "Manage view resources",
	Long:  `Manage view resources that are created from other resources using a view`,
}

var viewResourceGetCmd = &cobra.Command{
	Use:   "get [view-resource-id]",
	Short: "Get a specific view resource",
	Long:  `Get details for a specific view resource by ID`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		viewResourceIDStr := args[0]
		viewResourceID, err := strconv.ParseUint(viewResourceIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid view resource ID: %w", err)
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.GetViewResourceRequest{
			ViewResourceID: uint(viewResourceID),
		}

		viewResource, err := client.GetViewResource(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(viewResource, "View resource retrieved successfully\n")
	},
}

var viewResourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all view resources",
	Long:  `List all view resources in the system`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.ListViewResourcesRequest{}

		viewResources, err := client.ListViewResources(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(viewResources, "View resources listed successfully\n")
	},
}

var viewResourceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new view resource",
	Long:  `Create a new view resource with name and query`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return fmt.Errorf("failed to get name flag: %w", err)
		}

		query, err := cmd.Flags().GetString("query")
		if err != nil {
			return fmt.Errorf("failed to get query flag: %w", err)
		}

		if name == "" {
			return fmt.Errorf("name is required")
		}

		if query == "" {
			return fmt.Errorf("query is required")
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.CreateViewResourceRequest{
			Name:  name,
			Query: query,
		}

		viewResource, err := client.CreateViewResource(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(viewResource, "View resource created successfully with ID: %d\n", viewResource.ID)
	},
}

var viewResourceUpdateCmd = &cobra.Command{
	Use:   "update [view-resource-id]",
	Short: "Update a view resource",
	Long:  `Update a view resource's query`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		viewResourceIDStr := args[0]
		viewResourceID, err := strconv.ParseUint(viewResourceIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid view resource ID: %w", err)
		}

		query, err := cmd.Flags().GetString("query")
		if err != nil {
			return fmt.Errorf("failed to get query flag: %w", err)
		}

		if query == "" {
			return fmt.Errorf("query is required")
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.UpdateViewResourceRequest{
			ViewResourceID: uint(viewResourceID),
			Query:          &query,
		}

		viewResource, err := client.UpdateViewResource(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(viewResource, "View resource updated successfully\n")
	},
}

var viewResourceDeleteCmd = &cobra.Command{
	Use:   "delete [view-resource-id]",
	Short: "Delete a view resource",
	Long:  `Delete a view resource`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		viewResourceIDStr := args[0]
		viewResourceID, err := strconv.ParseUint(viewResourceIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid view resource ID: %w", err)
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.DeleteViewResourceRequest{
			ViewResourceID: uint(viewResourceID),
		}

		_, err = client.DeleteViewResource(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(nil, "View resource deleted successfully\n")
	},
}

func init() {
	viewResourceCreateCmd.Flags().String("name", "", "View resource name")
	viewResourceCreateCmd.Flags().String("query", "", "View resource query")
	_ = viewResourceCreateCmd.MarkFlagRequired("name")
	_ = viewResourceCreateCmd.MarkFlagRequired("query")

	viewResourceUpdateCmd.Flags().String("query", "", "View resource query")
	_ = viewResourceUpdateCmd.MarkFlagRequired("query")

	viewResourceCmd.AddCommand(viewResourceGetCmd)
	viewResourceCmd.AddCommand(viewResourceListCmd)
	viewResourceCmd.AddCommand(viewResourceCreateCmd)
	viewResourceCmd.AddCommand(viewResourceUpdateCmd)
	viewResourceCmd.AddCommand(viewResourceDeleteCmd)
}
