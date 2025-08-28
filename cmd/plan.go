package cmd

import (
	"fmt"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Create a new plan for a component",
	Long:  `Create a new plan for a specific component by its ID with branch name`,
	RunE: func(cmd *cobra.Command, args []string) error {
		componentID, err := cmd.Flags().GetUint("component-id")
		if err != nil {
			return fmt.Errorf("failed to get component-id flag: %w", err)
		}

		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return fmt.Errorf("failed to get changeset flag: %w", err)
		}

		if componentID == 0 {
			return fmt.Errorf("component-id is required")
		}
		if changeset == "" {
			return fmt.Errorf("changeset is required")
		}

		config, err := LoadConfig()
		if err != nil {
			return err
		}

		client := http.NewClient(config)

		req := internal.CreatePlanRequest{
			ComponentID: componentID,
			Changeset:   changeset,
		}

		plan, err := client.CreatePlan(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(plan, "Plan created successfully with ID: %d\n", plan.ID)
	},
}

func init() {
	planCmd.Flags().Uint("component-id", 0, "Component ID")
	planCmd.Flags().String("changeset", "", "Changeset name for the plan")
	planCmd.MarkFlagRequired("component-id")
	planCmd.MarkFlagRequired("changeset")
}
