package cmd

import (
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/apply"
	"github.com/marcbran/versource/pkg/versource"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Manage applies",
	Long:  `Manage applies`,
}

var applyGetCmd = &cobra.Command{
	Use:   "get [apply-id]",
	Short: "Get a specific apply",
	Long:  `Get details for a specific apply by ID`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.New(config)
		detailData := apply.NewDetailData(httpClient, args[0])

		waitForCompletion, err := cmd.Flags().GetBool("wait-for-completion")
		if err != nil {
			return err
		}

		return waitForTaskCompletion(
			ctx,
			waitForCompletion,
			detailData,
			func(resp versource.GetApplyResponse) bool {
				return versource.IsTaskCompleted(resp.Apply.State)
			},
		)
	},
}

var applyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applies",
	Long:  `List all applies in the system`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.New(config)
		tableData := apply.NewTableData(httpClient)

		waitForCompletion, err := cmd.Flags().GetBool("wait-for-completion")
		if err != nil {
			return err
		}

		return waitForTableCompletion(
			ctx,
			waitForCompletion,
			tableData,
			func(applies []versource.Apply) bool {
				for _, apply := range applies {
					if !versource.IsTaskCompleted(apply.State) {
						return false
					}
				}
				return true
			},
		)
	},
}

func init() {
	applyGetCmd.Flags().Bool("wait-for-completion", false, "Wait for the apply to reach a terminal state before returning")
	applyListCmd.Flags().Bool("wait-for-completion", false, "Wait for all applies to reach terminal states before returning")
	applyCmd.AddCommand(applyGetCmd)
	applyCmd.AddCommand(applyListCmd)
}
