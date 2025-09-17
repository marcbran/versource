package cmd

import (
	"context"
	"fmt"
	"strconv"
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
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.NewClient(config)
		detailData := plan.NewDetailData(httpClient, args[0])
		return renderViewpointData(detailData)
	},
}

var planWaitCmd = &cobra.Command{
	Use:   "wait [plan-id]",
	Short: "Wait for a plan to complete",
	Long:  `Wait for a plan to reach a terminal state (Succeeded, Failed, or Cancelled)`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		planID, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid plan ID: %s", args[0])
		}

		httpClient := client.NewClient(config)
		ctx := context.Background()

		planResp, err := httpClient.GetPlan(ctx, internal.GetPlanRequest{PlanID: uint(planID)})
		if err != nil {
			return err
		}

		if internal.IsTaskCompleted(planResp.State) {
			return nil
		}

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				planResp, err := httpClient.GetPlan(ctx, internal.GetPlanRequest{PlanID: uint(planID)})
				if err != nil {
					return err
				}

				if internal.IsTaskCompleted(planResp.State) {
					return nil
				}
			}
		}
	},
}

func init() {
	planCmd.AddCommand(planGetCmd)
	planCmd.AddCommand(planWaitCmd)
}
