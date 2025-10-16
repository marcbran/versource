package cmd

import (
	"fmt"

	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/changeset"
	"github.com/marcbran/versource/internal/tui/component"
	"github.com/marcbran/versource/pkg/versource"
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

		client := client.New(config)

		req := versource.CreateChangesetRequest{
			Name: name,
		}

		changeset, err := client.CreateChangeset(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(changeset, "Changeset created successfully with ID: %d\n", changeset.Changeset.ID)
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
		httpClient := client.New(config)
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

		client := client.New(config)

		req := versource.CreateMergeRequest{
			ChangesetName: changesetName,
		}

		merge, err := client.CreateMerge(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(merge, "Merge operation created for changeset %s (ID: %d)\n", changesetName, merge.Merge.ID)
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

		client := client.New(config)

		req := versource.CreateRebaseRequest{
			ChangesetName: changesetName,
		}

		rebase, err := client.CreateRebase(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(rebase, "Rebase operation created for changeset %s (ID: %d)\n", changesetName, rebase.Rebase.ID)
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

		client := client.New(config)

		req := versource.DeleteChangesetRequest{
			ChangesetName: changesetName,
		}

		_, err = client.DeleteChangeset(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(nil, "Changeset %s deleted successfully\n", changesetName)
	},
}

var changesetChangeCmd = &cobra.Command{
	Use:   "change",
	Short: "Manage changeset changes",
	Long:  `Manage changeset changes`,
}

var changesetChangeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List changes in a changeset",
	Long:  `List all component changes in a specific changeset`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		changesetName, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return fmt.Errorf("failed to get changeset flag: %w", err)
		}
		if changesetName == "" {
			return fmt.Errorf("changeset is required")
		}
		httpClient := client.New(config)
		tableData := component.NewChangesetChangesTableData(httpClient, changesetName)

		waitForCompletion, err := cmd.Flags().GetBool("wait-for-completion")
		if err != nil {
			return fmt.Errorf("failed to get wait-for-completion flag: %w", err)
		}

		return waitForTableCompletion(
			ctx,
			waitForCompletion,
			tableData,
			allPlansCompleted,
		)
	},
}

func allPlansCompleted(changes []versource.ComponentChange) bool {
	for _, change := range changes {
		if change.Plan == nil {
			return false
		}
		if !versource.IsTaskCompleted(change.Plan.State) {
			return false
		}
	}
	return true
}

func init() {
	changesetCreateCmd.Flags().String("name", "", "Changeset name")
	_ = changesetCreateCmd.MarkFlagRequired("name")

	changesetChangeListCmd.Flags().String("changeset", "", "Changeset name")
	_ = changesetChangeListCmd.MarkFlagRequired("changeset")
	changesetChangeListCmd.Flags().Bool("wait-for-completion", false, "Wait until all plans in the changeset are completed")

	changesetChangeCmd.AddCommand(changesetChangeListCmd)

	changesetCmd.AddCommand(changesetCreateCmd)
	changesetCmd.AddCommand(changesetListCmd)
	changesetCmd.AddCommand(changesetChangeCmd)
	changesetCmd.AddCommand(changesetMergeCmd)
	changesetCmd.AddCommand(changesetRebaseCmd)
	changesetCmd.AddCommand(changesetDeleteCmd)
}
