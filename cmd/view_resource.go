package cmd

import (
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/pkg/versource"
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

		req := versource.GetViewResourceRequest{
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

		req := versource.ListViewResourcesRequest{}

		viewResources, err := client.ListViewResources(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(viewResources, "View resources listed successfully\n")
	},
}

var viewResourceSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save a view resource",
	Long:  `Save a view resource. If a view resource with the same name exists, it will be updated. Otherwise, a new one will be created.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		req := versource.SaveViewResourceRequest{
			Query: query,
		}

		viewResource, err := client.SaveViewResource(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(viewResource, "View resource saved successfully with ID: %d\n", viewResource.ID)
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

		req := versource.DeleteViewResourceRequest{
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
	viewResourceSaveCmd.Flags().String("query", "", "View resource query")
	_ = viewResourceSaveCmd.MarkFlagRequired("query")

	viewResourceCmd.AddCommand(viewResourceGetCmd)
	viewResourceCmd.AddCommand(viewResourceListCmd)
	viewResourceCmd.AddCommand(viewResourceSaveCmd)
	viewResourceCmd.AddCommand(viewResourceDeleteCmd)
}
