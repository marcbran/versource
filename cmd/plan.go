package cmd

import (
	"time"

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
		httpClient := client.NewClient(config)
		detailData := plan.NewDetailData(httpClient, args[0])
		planResp, err := detailData.LoadData()
		if err != nil {
			return err
		}

		waitForCompletion, _ := cmd.Flags().GetBool("wait-for-completion")
		if !waitForCompletion || internal.IsTaskCompleted(planResp.State) {
			return renderValue(planResp, func() string {
				return detailData.ResolveData(*planResp)
			})
		}

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				planResp, err = detailData.LoadData()
				if err != nil {
					return err
				}

				if !internal.IsTaskCompleted(planResp.State) {
					continue
				}

				return renderValue(planResp, func() string {
					return detailData.ResolveData(*planResp)
				})
			}
		}
	},
}

func init() {
	planGetCmd.Flags().Bool("wait-for-completion", false, "Wait for the plan to reach a terminal state before returning")
	planCmd.AddCommand(planGetCmd)
}
