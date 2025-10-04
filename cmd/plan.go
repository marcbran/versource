package cmd

import (
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/plan"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Manage plans",
	Long:  `Manage plans`,
}

var planGetCmd = &cobra.Command{
	Use:   "get [plan-id]",
	Short: "Get a specific plan",
	Long:  `Get details for a specific plan by ID`,
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
		waitForCompletion, err := cmd.Flags().GetBool("wait-for-completion")
		if err != nil {
			return err
		}

		httpClient := client.NewClient(config)
		detailData := plan.NewDetailData(httpClient, changeset, args[0])

		return waitForTaskCompletion(
			ctx,
			waitForCompletion,
			detailData,
			func(resp internal.GetPlanResponse) bool {
				return internal.IsTaskCompleted(resp.State)
			},
		)
	},
}

func init() {
	planGetCmd.Flags().Bool("wait-for-completion", false, "Wait for the plan to reach a terminal state before returning")
	planGetCmd.Flags().String("changeset", "", "Changeset name to get the plan from")
	planCmd.AddCommand(planGetCmd)
}
