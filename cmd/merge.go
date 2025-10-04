package cmd

import (
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/merge"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Manage merges",
	Long:  `Manage merges`,
}

var mergeGetCmd = &cobra.Command{
	Use:   "get [merge-id]",
	Short: "Get a specific merge",
	Long:  `Get details for a specific merge by ID`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return err
		}
		httpClient := client.NewClient(config)
		detailData := merge.NewDetailData(httpClient, changeset, args[0])

		waitForCompletion, err := cmd.Flags().GetBool("wait-for-completion")
		if err != nil {
			return err
		}

		return waitForTaskCompletion(
			ctx,
			waitForCompletion,
			detailData,
			func(resp internal.GetMergeResponse) bool {
				return internal.IsTaskCompleted(resp.State)
			},
		)
	},
}

var mergeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all merges",
	Long:  `List all merges in the system`,
	RunE: func(cmd *cobra.Command, args []string) error {
		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return err
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.NewClient(config)
		tableData := merge.NewTableData(httpClient, changeset)
		return renderTableData(tableData)
	},
}

func init() {
	mergeGetCmd.Flags().Bool("wait-for-completion", false, "Wait for the merge to reach a terminal state before returning")
	mergeGetCmd.Flags().String("changeset", "", "Changeset name (required)")
	_ = mergeGetCmd.MarkFlagRequired("changeset")
	mergeListCmd.Flags().String("changeset", "", "Changeset name (optional)")
	mergeCmd.AddCommand(mergeGetCmd)
	mergeCmd.AddCommand(mergeListCmd)
}
