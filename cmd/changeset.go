package cmd

import (
	"fmt"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/changeset"
	"github.com/spf13/cobra"
)

var changesetCmd = &cobra.Command{
	Use:   "changeset",
	Short: "Manage changesets",
	Long:  `Manage changesets`,
}

var changesetCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new changeset",
	Long:  `Create a new changeset with a name`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return fmt.Errorf("failed to get name flag: %w", err)
		}

		if name == "" {
			return fmt.Errorf("name is required")
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.CreateChangesetRequest{
			Name: name,
		}

		changeset, err := client.CreateChangeset(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(changeset, "Changeset created successfully with ID: %d\n", changeset.ID)
	},
}

var changesetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all changesets",
	Long:  `List all changesets in the system`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.NewClient(config)
		tableData := changeset.NewTableData(httpClient)
		return renderTableData(tableData)
	},
}

var changesetMergeCmd = &cobra.Command{
	Use:   "merge [changeset-name]",
	Short: "Merge a changeset",
	Long:  `Merge a changeset branch into main`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		changesetName := args[0]
		if changesetName == "" {
			return fmt.Errorf("changeset name is required")
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.CreateMergeRequest{
			ChangesetName: changesetName,
		}

		merge, err := client.CreateMerge(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(merge, "Merge operation created for changeset %s (ID: %d)\n", changesetName, merge.ID)
	},
}

var changesetRebaseCmd = &cobra.Command{
	Use:   "rebase [changeset-name]",
	Short: "Rebase a changeset",
	Long:  `Rebase a changeset branch onto main`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		changesetName := args[0]
		if changesetName == "" {
			return fmt.Errorf("changeset name is required")
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.CreateRebaseRequest{
			ChangesetName: changesetName,
		}

		rebase, err := client.CreateRebase(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(rebase, "Rebase operation created for changeset %s (ID: %d)\n", changesetName, rebase.ID)
	},
}

var changesetDeleteCmd = &cobra.Command{
	Use:   "delete [changeset-name]",
	Short: "Delete a changeset",
	Long:  `Delete a changeset and all its associated data including plans, applies, and logs`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		changesetName := args[0]
		if changesetName == "" {
			return fmt.Errorf("changeset name is required")
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.DeleteChangesetRequest{
			ChangesetName: changesetName,
		}

		_, err = client.DeleteChangeset(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(nil, "Changeset %s deleted successfully\n", changesetName)
	},
}

func init() {
	changesetCreateCmd.Flags().String("name", "", "Changeset name")
	changesetCreateCmd.MarkFlagRequired("name")

	changesetCmd.AddCommand(changesetCreateCmd)
	changesetCmd.AddCommand(changesetListCmd)
	changesetCmd.AddCommand(changesetMergeCmd)
	changesetCmd.AddCommand(changesetRebaseCmd)
	changesetCmd.AddCommand(changesetDeleteCmd)
}
