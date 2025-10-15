package cmd

import (
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/rebase"
	"github.com/marcbran/versource/pkg/versource"
	"github.com/spf13/cobra"
)

var rebaseCmd = &cobra.Command{
	Use:   "rebase",
	Short: "Manage rebases",
	Long:  `Manage rebases`,
}

var rebaseGetCmd = &cobra.Command{
	Use:   "get [rebase-id]",
	Short: "Get a specific rebase",
	Long:  `Get details for a specific rebase by ID`,
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
		detailData := rebase.NewDetailData(httpClient, changeset, args[0])

		waitForCompletion, err := cmd.Flags().GetBool("wait-for-completion")
		if err != nil {
			return err
		}

		return waitForTaskCompletion(
			ctx,
			waitForCompletion,
			detailData,
			func(resp versource.GetRebaseResponse) bool {
				return versource.IsTaskCompleted(resp.State)
			},
		)
	},
}

var rebaseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all rebases",
	Long:  `List all rebases in the system`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return err
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.NewClient(config)
		tableData := rebase.NewTableData(httpClient, changeset)

		waitForCompletion, err := cmd.Flags().GetBool("wait-for-completion")
		if err != nil {
			return err
		}

		return waitForTableCompletion(
			ctx,
			waitForCompletion,
			tableData,
			func(rebases []versource.Rebase) bool {
				for _, rebase := range rebases {
					if !versource.IsTaskCompleted(rebase.State) {
						return false
					}
				}
				return true
			},
		)
	},
}

func init() {
	rebaseGetCmd.Flags().Bool("wait-for-completion", false, "Wait for the rebase to reach a terminal state before returning")
	rebaseGetCmd.Flags().String("changeset", "", "Changeset name (required)")
	_ = rebaseGetCmd.MarkFlagRequired("changeset")
	rebaseListCmd.Flags().String("changeset", "", "Changeset name (optional)")
	rebaseListCmd.Flags().Bool("wait-for-completion", false, "Wait for all rebases to reach terminal states before returning")
	rebaseCmd.AddCommand(rebaseGetCmd)
	rebaseCmd.AddCommand(rebaseListCmd)
}
